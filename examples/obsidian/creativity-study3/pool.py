#!/usr/bin/env python3
"""Pool round-2 and round-3 ratings for the selected-vs-labeled head-to-head.

Both rounds used the same four problems, the same twelve selected notes, the
same reader instructions, the same model and the same rubric, so their curated
conditions pool directly: 2 + 4 = 6 generations per condition per problem.
"""
import json
from collections import defaultdict
from pathlib import Path
from statistics import mean

HERE = Path(__file__).resolve().parent
ROUND2 = HERE.parent / "creativity-study2" / "run"
METRICS = ("originality", "feasibility", "test")


def load(directory):
    rows = json.loads((directory / "answers.json").read_text())
    ratings = [json.loads((directory / f"ratings{j}.json").read_text()) for j in range(3)]
    out = []
    for row in rows:
        if row["condition"] not in ("selected", "labeled"):
            continue
        for points in ratings:
            out.append({"problem": row["problem"], "condition": row["condition"],
                        "scores": points[row["id"]]})
    return out


def main():
    data = load(ROUND2) + load(HERE / "run")
    by_condition = defaultdict(list)
    by_problem = defaultdict(lambda: defaultdict(list))
    for row in data:
        by_condition[row["condition"]].append(row["scores"])
        by_problem[row["problem"]][row["condition"]].append(row["scores"])

    def means(vals):
        return dict(zip(METRICS, (round(mean(v), 3) for v in zip(*vals))))

    pooled = {k: means(v) for k, v in by_condition.items()}
    problems = {p: {k: means(v) for k, v in arms.items()} for p, arms in by_problem.items()}
    n = {k: len(v) for k, v in by_condition.items()}
    diff = {p: round(arms["labeled"]["originality"] - arms["selected"]["originality"], 3)
            for p, arms in problems.items()}
    summary = {"ratings_per_condition": n, "pooled_means": pooled,
               "per_problem": problems, "labeled_minus_selected_originality": diff,
               "labeled_wins": sum(1 for d in diff.values() if d > 0),
               "selected_wins": sum(1 for d in diff.values() if d < 0),
               "ties": sum(1 for d in diff.values() if d == 0)}
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
