"""Compare two cx binaries on end-to-end imports into fresh workspaces."""
import argparse
import json
from pathlib import Path
import sqlite3
import statistics
import subprocess
import tempfile
import time


def measure(binary, directory, transcript, count):
    root = directory / "state"
    cfg = directory / "config.yaml"
    cfg.write_text("root: " + json.dumps(str(root)) + "\n")
    cmd = [binary, "-config", cfg]
    subprocess.run(cmd + ["new", "bench"], capture_output=True, check=True)
    start = time.perf_counter()
    subprocess.run(cmd + ["import", "bench", transcript], capture_output=True, check=True)
    elapsed = time.perf_counter() - start
    vault = root / "workspaces/bench"
    tree = subprocess.check_output(["git", "-C", vault, "rev-parse", "HEAD:turns"], text=True).strip()
    if len(list((vault / "turns").glob("*.md"))) != count:
        raise ValueError("turn count mismatch")
    with sqlite3.connect(vault / ".cx/state.db") as db:
        rows = db.execute("SELECT count(DISTINCT id) FROM fts WHERE type='turn'").fetchone()[0]
        if rows != count or db.execute("PRAGMA integrity_check").fetchone() != ("ok",):
            raise ValueError("index verification failed")
    return round(elapsed, 4), tree


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("baseline", type=Path)
    p.add_argument("candidate", type=Path)
    p.add_argument("--turns", type=int, default=4096)
    p.add_argument("--pairs", type=int, default=3)
    args = p.parse_args()
    if min(args.turns, args.pairs) < 1:
        p.error("turns and pairs must be positive")
    runs = {"baseline": [], "candidate": []}
    trees = set()
    with tempfile.TemporaryDirectory() as temp:
        scratch = Path(temp)
        transcript = scratch / "chat.jsonl"
        with transcript.open("w") as out:
            for i in range(args.turns):
                msg = {"role": "assistant" if i % 2 else "user", "timestamp": "2026-11-19T14:00:00Z",
                       "content": f"Record {i} for workshop {i % 79}. " + "Check the drawing against the original. " * 5}
                out.write(json.dumps(msg) + "\n")
        for pair in range(args.pairs):
            for arm in list(runs)[::1 if pair % 2 == 0 else -1]:
                with tempfile.TemporaryDirectory(dir=scratch) as run:
                    elapsed, tree = measure(getattr(args, arm).resolve(), Path(run), transcript, args.turns)
                    trees.add(tree)
                    runs[arm].append(elapsed)
    if len(trees) != 1:
        raise ValueError("committed turn filenames or contents differed between runs")
    print(json.dumps({"turns": args.turns, "turns_git_tree": trees.pop(), "seconds": runs,
                      "median_seconds": {k: statistics.median(v) for k, v in runs.items()}}, indent=2))


if __name__ == "__main__":
    main()
