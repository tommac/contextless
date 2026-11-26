"""Exercise the typed coverage prototype on an immutable synthetic event stream."""

import argparse
from dataclasses import replace
import json
from pathlib import Path
import tempfile
import sqlite3
import time

from receipts import (
    Plan,
    advance,
    append,
    create,
    encode,
    full_scan,
    page,
    sql_baseline,
)


def timed(fn, *args):
    start = time.perf_counter()
    value = fn(*args)
    return value, round((time.perf_counter() - start) * 1000, 3)


def check_pages(db, plan, receipt, expected):
    actual = {}
    cursor = None
    while True:
        result = page(db, plan, receipt, cursor, limit=64)
        for entry in result["examples"]:
            raw = db.execute(
                "SELECT item,raw FROM events WHERE seq=?", (entry["id"],)
            ).fetchone()
            if raw[0] in actual or raw[1] != entry["text"].encode():
                raise AssertionError("duplicate or mismatched witness")
            actual[raw[0]] = entry["id"]
        cursor = result["after"]
        if cursor is None:
            break
    assert actual == expected


def rejected(fn, *args):
    try:
        fn(*args)
    except ValueError:
        return
    raise AssertionError("invalid continuation accepted")


def known_answers():
    db = create(":memory:")
    plan = Plan(7, 10)
    rev = append(
        db,
        [
            encode(7, 10, 4, 1),
            encode(7, 20, 7, 1),
            encode(7, 20, 6, 0),
            encode(7, 30, 11, 1),
        ],
    )
    before, _ = advance(db, plan, rev)
    check_pages(db, plan, before, {10: 1, 20: 2})
    old_cursor = page(db, plan, before, limit=1)["after"]
    rev = append(db, [encode(7, 20, 7, 0), encode(7, 5, 1, 1)])
    after, _ = advance(db, plan, rev, before)
    check_pages(db, plan, after, {5: 6, 10: 1})
    rejected(page, db, plan, after, old_cursor)
    rev = append(db, [b"scope=7 item=9223372036854775808 effective=1 state=blocked"])
    unknown, _ = advance(db, plan, rev, after)
    result = page(db, plan, unknown)
    assert result["unparsed_rows"] == 1 and result["exact_count"] is None
    assert page(db, plan, before)["exact_count"] == 2
    db.close()


def maintain_sql(db, plan, base, rev):
    """Conventional incremental SQL control for this one plan, retaining just its current winners."""
    start = db.execute(
        "SELECT end_seq FROM epochs WHERE revision=?", (base,)
    ).fetchone()[0]
    end = db.execute("SELECT end_seq FROM epochs WHERE revision=?", (rev,)).fetchone()[
        0
    ]
    with db:
        db.execute(
            """INSERT INTO sql_current(item,tick,seq,blocked)
            SELECT item,tick,seq,blocked FROM events WHERE scope=? AND tick<=? AND seq>? AND seq<=?
            ON CONFLICT(item) DO UPDATE SET tick=excluded.tick,seq=excluded.seq,blocked=excluded.blocked
            WHERE (excluded.tick,excluded.seq)>(sql_current.tick,sql_current.seq)""",
            (plan.scope, plan.as_of, start, end),
        )
    return db.execute("SELECT coalesce(sum(blocked),0) FROM sql_current").fetchone()[0]


def run(total):
    if total < 128:
        raise ValueError("at least 128 records required")
    # Clock ticks are event time; ordinals are ingestion order. An event that
    # arrives late can still win. Distinct scopes make irrelevant appends common.
    known_answers()
    plan = Plan(scope=7, as_of=total)
    tickets = max(64, total // (4 * 64) * 64)
    with tempfile.TemporaryDirectory() as scratch:
        path = Path(scratch) / "events.db"
        db = create(path)
        db.execute(
            "CREATE TABLE sql_current(item INTEGER PRIMARY KEY,tick INTEGER,seq INTEGER,blocked INTEGER)"
        )

        def initial():
            for i in range(total):
                yield encode(i % 64, i % tickets, i, (i // 64 + i // tickets) % 3 == 0)

        steps = [
            ("initial", initial()),
            ("unrelated append", (encode(8, i, total, 1) for i in range(10_000))),
            (
                "corrections",
                [
                    encode(7, 7, total, 1),  # touches an existing item
                    encode(7, 7, total - 1, 0),  # older despite arriving later
                    encode(7, 7, total, 0),  # equal-time correction wins
                    encode(7, 7, total + 1, 1),  # beyond the as-of time
                    encode(
                        7, total + 7, total - 1, 1
                    ),  # historical event, recorded late
                ],
            ),
            (
                "concentrated append",
                (
                    encode(7, 7 + 64 * (i % 64), total, i % 5 == 0)
                    for i in range(10_000)
                ),
            ),
            ("unparsed append", [b"free text: reopen the earlier ticket"]),
        ]
        results = []
        first = prior = None
        for label, events in steps:
            revision, build_ms = timed(append, db, events)
            (receipt, work), refresh_ms = timed(advance, db, plan, revision, prior)
            result, page_ms = timed(page, db, plan, receipt)
            (expected, unknown), scan_ms = timed(full_scan, db, plan, revision)
            sql_count, sql_ms = timed(sql_baseline, db, plan, revision)
            current, current_ms = timed(maintain_sql, db, plan, revision - 1, revision)
            assert (
                current == sql_count == len(expected) == result["count_on_parsed_rows"]
            )
            assert unknown == result["unparsed_rows"]
            assert result["exact_count"] == (None if unknown else len(expected))
            check_pages(db, plan, receipt, expected)
            envelope = len(
                json.dumps(result, ensure_ascii=False, separators=(",", ":")).encode()
            )
            results.append(
                {
                    "step": label,
                    **work,
                    "exact_count": result["exact_count"],
                    "unparsed_rows": unknown,
                    "packet_utf8_bytes": envelope,
                    "build_ms": build_ms,
                    "update_ms": refresh_ms,
                    "page_ms": page_ms,
                    "full_raw_scan_ms": scan_ms,
                    "indexed_sql_count_ms": sql_ms,
                    "incremental_sql_ms": current_ms,
                }
            )
            if first is None:
                first = (receipt, expected)
            prior = receipt
        # Old receipt pages must still resolve exactly after later edits, including
        # an unparseable record. A reopen exercises disk-persisted query state.
        db.close()
        db = sqlite3.connect(path)
        check_pages(db, plan, *first)
        rejected(page, db, plan, replace(prior, root="0" * 64))
        rejected(advance, db, Plan(7, plan.as_of - 1), prior.revision, prior)
        rejected(advance, db, replace(plan, parser="v2"), prior.revision, prior)
        rejected(advance, db, plan, prior.revision, first[0])
        # A changed as-of is a different query: cold build and verify it.
        changed = Plan(7, total + 1)
        changed_receipt, _ = advance(db, changed, prior.revision)
        expected, _ = full_scan(db, changed, prior.revision)
        check_pages(db, changed, changed_receipt, expected)
        raw_bytes = db.execute("SELECT sum(length(raw)) FROM events").fetchone()[0]
        db.close()
        return {
            "initial_events": total,
            "raw_payload_bytes_final": raw_bytes,
            "sqlite_file_bytes_with_both_controls": path.stat().st_size,
            "steps": results,
            "checks": [
                "all positive IDs and source bytes after each stage",
                "old page after reopen",
                "hand-labeled backdated and equal-time corrections",
                "cross-snapshot cursor rejected",
                "out-of-range source integer retained as unknown",
                "wrong root rejected",
                "different cutoff rejected",
                "different decoder rejected",
                "stale resume rejected",
                "new plan cold rebuild",
            ],
        }


if __name__ == "__main__":
    opts = argparse.ArgumentParser(description=__doc__)
    opts.add_argument("--records", type=int, default=1000000)
    print(json.dumps(run(opts.parse_args().records), indent=2))
