#!/usr/bin/env python3
"""Paired analysis: combined revision vs its own question-only draft.

Every combined repeat revises the same repeat's question draft, so each
(problem, repeat) pair yields one originality/feasibility/test delta.
"""
import json
from collections import defaultdict
from pathlib import Path
from statistics import mean

HERE = Path(__file__).resolve().parent
METRICS = ("originality", "feasibility", "test")


def main():
    run = HERE / "run"
    rows = json.loads((run / "answers.json").read_text())
    ratings = [json.loads((run / f"ratings{j}.json").read_text()) for j in range(3)]
    score = {}
    for row in rows:
        vals = [list(map(mean, zip(*[r[row["id"]] for r in ratings])))]
        score[(row["condition"], row["problem"], row["repeat"])] = vals[0]

    pairs = []
    for problem in "ABCD":
        for rep in (1, 2, 3, 4):
            q = score[("question", problem, rep)]
            c = score[("combined", problem, rep)]
            pairs.append({"problem": problem, "repeat": rep,
                          "draft": [round(v, 2) for v in q],
                          "revision": [round(v, 2) for v in c],
                          "delta": [round(c[i] - q[i], 2) for i in range(3)]})
    by_metric = defaultdict(list)
    for p in pairs:
        for i, m in enumerate(METRICS):
            by_metric[m].append(p["delta"][i])
    summary = {
        "pairs": pairs,
        "mean_delta": {m: round(mean(v), 3) for m, v in by_metric.items()},
        "revisions_improved": {m: sum(1 for v in by_metric[m] if v > 0) for m in METRICS},
        "revisions_worsened": {m: sum(1 for v in by_metric[m] if v < 0) for m in METRICS},
        "n_pairs": len(pairs),
    }
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
