"""Run with python3 from any directory. Both experimental binaries use WAL + FULL on fresh databases."""

import argparse
import json
from pathlib import Path
import sqlite3
import subprocess
import tempfile
import time

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]


def build(scratch, source, name):
    # Both arms use the same explicit commit durability irrespective of SQLite defaults.
    adjusted = scratch / (name + ".go")
    adjusted.write_text(
        source.replace(
            "journal_mode(WAL)&", "journal_mode(WAL)&_pragma=synchronous(FULL)&"
        )
    )
    overlay = scratch / (name + ".json")
    overlay.write_text(
        json.dumps({"Replace": {str(REPO / "internal/store/store.go"): str(adjusted)}})
    )
    binary = scratch / name
    subprocess.run(
        ["go", "build", "-overlay", overlay, "-o", binary, "./cmd/cx"],
        check=True,
        cwd=REPO,
    )
    return binary


def measure(binary, work, transcript, count):
    cfg = work / "config.yaml"
    cfg.write_text("root: " + json.dumps(str(work / "state")) + "\n")

    def cx(*args):
        start = time.perf_counter()
        subprocess.run(
            [binary, "-config", cfg, *args], check=True, stdout=subprocess.DEVNULL
        )
        return time.perf_counter() - start

    cx("new", "bench")
    times = {}
    snapshots = []
    for stage, args in [
        ("import", ["import", "bench", transcript]),
        ("reindex", ["reindex", "bench"]),
    ]:
        times[stage] = round(cx(*args), 3)
        with sqlite3.connect(work / "state/workspaces/bench/.cx/state.db") as db:
            text = db.execute("select id,type,text from fts order by id").fetchall()
            assert len(text) == len({row[0] for row in text}) == count
            ranks = db.execute(
                "select id,bm25(fts) from fts where fts match 'memory OR evidence' order by bm25(fts),id limit 30"
            ).fetchall()
            assert ranks and db.execute("pragma integrity_check").fetchall() == [
                ("ok",)
            ]
            snapshots.append((text, ranks))
    assert snapshots[0] == snapshots[1]
    return times, snapshots[0]


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--copies", type=int, default=1)
    parser.add_argument("--pairs", type=int, default=4)
    args = parser.parse_args()
    assert min(args.copies, args.pairs) > 0
    records = [
        {"role": "user", "content": p.read_text(), "timestamp": "2026-01-01T00:00:00Z"}
        for p in sorted((REPO / "idea-vault").glob("*/*.md"))
    ] * args.copies
    with tempfile.TemporaryDirectory() as tmp:
        scratch = Path(tmp)
        source = REPO / "internal/store/store.go"
        candidate = scratch / "candidate-source.go"
        subprocess.run(
            ["patch", "-s", "-o", candidate, source, HERE / "candidate.patch"],
            check=True,
        )
        binaries = [
            build(scratch, source.read_text(), "scan"),
            build(scratch, candidate.read_text(), "rowid"),
        ]
        transcript = scratch / "transcript.json"
        transcript.write_text(json.dumps(records))
        reference = None
        for trial in range(args.pairs):
            for binary in binaries if trial % 2 == 0 else binaries[::-1]:
                with tempfile.TemporaryDirectory(dir=scratch) as case:
                    times, state = measure(binary, Path(case), transcript, len(records))
                    if reference is None:
                        reference = state
                    assert state == reference
                    print(
                        json.dumps(
                            dict(
                                trial=trial + 1,
                                arm=binary.name,
                                rows=len(records),
                                **times,
                            )
                        ),
                        flush=True,
                    )


if __name__ == "__main__":
    main()
