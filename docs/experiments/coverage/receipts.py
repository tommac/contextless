"""Research prototype for one typed predicate over immutable, versioned source events."""

import hashlib
import json
import sqlite3
from dataclasses import asdict, dataclass

ZERO = "0" * 64
MAX_INT = (1 << 63) - 1
PARSER = "ascii-event-v1"
SCHEMA = """
CREATE TABLE epochs(revision INTEGER PRIMARY KEY, end_seq INTEGER NOT NULL, root TEXT NOT NULL);
CREATE TABLE events(seq INTEGER PRIMARY KEY, scope INTEGER, item INTEGER, tick INTEGER, blocked INTEGER, raw BLOB NOT NULL);
CREATE INDEX event_scope ON events(scope,seq);
CREATE TABLE receipts(plan TEXT, rev INTEGER, n INTEGER, unknown INTEGER, PRIMARY KEY(plan,rev));
CREATE TABLE versions(plan TEXT, item INTEGER, first_rev INTEGER, last_rev INTEGER, tick INTEGER, seq INTEGER,
    blocked INTEGER, PRIMARY KEY(plan,item,first_rev));
CREATE UNIQUE INDEX latest ON versions(plan,item) WHERE last_rev IS NULL;
CREATE INDEX pages ON versions(plan,blocked,item,first_rev,last_rev);
"""


def encode(scope, item, tick, blocked):
    return f"scope={scope} item={item} effective={tick} state={'blocked' if blocked else 'clear'}".encode()


def decode(raw):
    try:
        scope, item, tick, state = raw.split()
        fields = (
            int(scope.split(b"=")[1]),
            int(item.split(b"=")[1]),
            int(tick.split(b"=")[1]),
            state == b"state=blocked",
        )
        # Reject unsupported fields, signs, labels and encodings instead of silently losing a correction.
        if min(fields[:3]) < 0 or max(fields[:3]) > MAX_INT or encode(*fields) != raw:
            return None
        return fields
    except (IndexError, ValueError):
        return None


@dataclass(frozen=True)
class Plan:
    scope: int
    as_of: int
    parser: str = PARSER

    def key(self):
        if self.parser != PARSER:
            raise ValueError("unsupported parser revision")
        if any(
            type(v) is not int or not 0 <= v <= MAX_INT
            for v in (self.scope, self.as_of)
        ):
            raise ValueError("scope or cutoff outside supported range")
        data = {"operator": "count-latest-blocked-v1", **asdict(self)}
        return hashlib.sha256(json.dumps(data, sort_keys=True).encode()).hexdigest()


@dataclass(frozen=True)
class Receipt:
    plan: str
    revision: int
    root: str


def create(path):
    db = sqlite3.connect(path)
    db.executescript(SCHEMA)
    with db:
        db.execute("INSERT INTO epochs VALUES(0,0,?)", (ZERO,))
    return db


def append(db, raws):
    """One sealed segment per transaction. Length-prefixing makes the byte stream unambiguous."""
    with db:
        db.execute("BEGIN IMMEDIATE")
        rev, seq, parent = db.execute(
            "SELECT * FROM epochs ORDER BY revision DESC LIMIT 1"
        ).fetchone()
        h = hashlib.sha256(bytes.fromhex(parent))

        def rows():
            nonlocal seq
            for raw in raws:
                seq += 1
                h.update(len(raw).to_bytes(8, "big"))
                h.update(raw)
                yield (seq, *(decode(raw) or (None, None, None, None)), raw)

        db.executemany("INSERT INTO events VALUES(?,?,?,?,?,?)", rows())
        db.execute("INSERT INTO epochs VALUES(?,?,?)", (rev + 1, seq, h.hexdigest()))
    return rev + 1


def snapshot(db, rev):
    row = db.execute(
        "SELECT end_seq,root FROM epochs WHERE revision=?", (rev,)
    ).fetchone()
    if row is None:
        raise ValueError("unknown revision")
    return row


def validate(db, plan, receipt):
    if receipt.plan != plan.key() or snapshot(db, receipt.revision)[1] != receipt.root:
        raise ValueError("different plan or snapshot ancestry")
    row = db.execute(
        "SELECT n,unknown FROM receipts WHERE plan=? AND rev=?",
        (receipt.plan, receipt.revision),
    ).fetchone()
    if row is None:
        raise ValueError("missing receipt state")
    return row


def advance(db, plan, rev, prior=None):
    """Persist only the changed winners. Earlier receipts can still page their historical result."""
    key = plan.key()
    with db:
        db.execute("BEGIN IMMEDIATE")
        end, root = snapshot(db, rev)
        latest = db.execute(
            "SELECT max(rev) FROM receipts WHERE plan=?", (key,)
        ).fetchone()[0]
        if prior is None:
            if latest is not None:
                raise ValueError("existing state requires its latest receipt")
            start, n, unknown = 0, 0, 0
        else:
            n, unknown = validate(db, plan, prior)
            if prior.revision != latest or rev <= latest:
                raise ValueError("resume from the latest receipt to a later revision")
            start = snapshot(db, prior.revision)[0]
        # Unknown scope can affect any item in any scope.
        unknown += db.execute(
            "SELECT count(*) FROM events WHERE scope IS NULL AND seq>? AND seq<=?",
            (start, end),
        ).fetchone()[0]
        # Fold this delta by item, then compare to previous winners. Peak fold memory is
        # proportional to distinct candidate items in this scope and delta.
        changes = {}
        candidates = 0
        for item, tick, seq, state in db.execute(
            "SELECT item,tick,seq,blocked FROM events WHERE scope=? AND seq>? AND seq<=?",
            (plan.scope, start, end),
        ):
            candidates += 1
            if tick <= plan.as_of and (
                item not in changes or (tick, seq) > changes[item][:2]
            ):
                changes[item] = (tick, seq, state)
        updated = 0
        for item, (tick, seq, state) in changes.items():
            old = db.execute(
                "SELECT tick,seq,blocked FROM versions WHERE plan=? AND item=? AND last_rev IS NULL",
                (key, item),
            ).fetchone()
            if old is not None:
                if (tick, seq) <= old[:2]:
                    continue
                n -= old[2]
                db.execute(
                    "UPDATE versions SET last_rev=? WHERE plan=? AND item=? AND last_rev IS NULL",
                    (rev, key, item),
                )
            db.execute(
                "INSERT INTO versions VALUES(?,?,?,NULL,?,?,?)",
                (key, item, rev, tick, seq, state),
            )
            n += state
            updated += 1
        db.execute("INSERT INTO receipts VALUES(?,?,?,?)", (key, rev, n, unknown))
    return Receipt(key, rev, root), {
        "new_rows": end - start,
        "typed_candidates": candidates,
        "changed_winners": updated,
    }


def cursor(receipt, item):
    key = f"{receipt.plan}|{receipt.root}|{receipt.revision}|{item}"
    return f"{item}:{hashlib.sha256(key.encode()).hexdigest()}"


def page(db, plan, receipt, after=None, limit=2):
    item = -1
    if after is not None:
        if (
            not isinstance(after, str)
            or not after.partition(":")[0].isascii()
            or not after.partition(":")[0].isdecimal()
        ):
            raise ValueError("invalid continuation")
        item = int(after.partition(":")[0])
        if item > MAX_INT or cursor(receipt, item) != after:
            raise ValueError("cursor belongs to another snapshot or query")
    if not 1 <= limit <= 64:
        raise ValueError("page size outside supported range")
    n, unknown = validate(db, plan, receipt)
    rows = list(
        db.execute(
            """SELECT v.item,v.seq,e.raw FROM versions v JOIN events e ON e.seq=v.seq
        WHERE v.plan=? AND v.blocked=1 AND v.item>? AND v.first_rev<=? AND (v.last_rev IS NULL OR v.last_rev>?)
        ORDER BY v.item LIMIT ?""",
            (receipt.plan, item, receipt.revision, receipt.revision, limit + 1),
        )
    )
    return {
        "receipt": asdict(receipt),
        "exact_count": None if unknown else n,
        "count_on_parsed_rows": n,
        "unparsed_rows": unknown,
        "after": cursor(receipt, rows[limit - 1][0]) if len(rows) > limit else None,
        "examples": [{"id": seq, "text": raw.decode()} for _, seq, raw in rows[:limit]],
    }


def full_scan(db, plan, rev):
    """Independent raw-byte fold. No materialized state or secondary index."""
    end, _ = snapshot(db, rev)
    winners = {}
    unknown = 0
    for seq, raw in db.execute("SELECT seq,raw FROM events WHERE seq<=?", (end,)):
        ev = decode(raw)
        if ev is None:
            unknown += 1
            continue
        scope, item, tick, state = ev
        if (
            scope == plan.scope
            and tick <= plan.as_of
            and (item not in winners or (tick, seq) > winners[item][:2])
        ):
            winners[item] = (tick, seq, state)
    return {item: seq for item, (_, seq, state) in winners.items() if state}, unknown


def sql_baseline(db, plan, rev):
    end, _ = snapshot(db, rev)
    return db.execute(
        """SELECT coalesce(sum(blocked),0) FROM (
        SELECT blocked,row_number() OVER(PARTITION BY item ORDER BY tick DESC,seq DESC) AS rank
        FROM events WHERE scope=? AND tick<=? AND seq<=?) WHERE rank=1""",
        (plan.scope, plan.as_of, end),
    ).fetchone()[0]
