"""Blind archived reader outputs for review, or summarize completed reviews."""
import argparse
from collections import defaultdict
from itertools import combinations
import json
from pathlib import Path
import random
from statistics import mean

HERE = Path(__file__).resolve().parent
PROTO = json.loads((HERE / "protocol.json").read_text())
METRICS = ("originality", "feasibility", "test")


def read(path):
    return json.loads(path.read_text())


def write(path, data):
    path.write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n")


def blind(work):
    ids = list(range(101, 133))
    random.Random(32391).shuffle(ids)
    results = []
    for index, arm in enumerate(PROTO["conditions"]):
        for repeat in range(PROTO["repeats"]):
            run = index * PROTO["repeats"] + repeat
            answers = read(work / f"out{run}.json")
            if set(answers) != set(PROTO["problems"]):
                raise ValueError(f"incomplete run {run}")
            for problem in sorted(answers):
                full = answers[problem]
                words = full.split()
                if not words:
                    raise ValueError("empty design")
                clipped = " ".join(words[:110]) if len(words) > 110 else full
                results.append({"id": str(ids.pop()), "condition": arm, "repeat": repeat + 1,
                                "problem": problem, "original_words": len(words), "answer": clipped,
                                **({"unclipped": full} if len(words) > 110 else {})})
    write(work / "answers.json", results)
    for judge in range(3):
        rng = random.Random(61920 + judge)
        packets = []
        problems = list(PROTO["problems"])
        rng.shuffle(problems)
        for key in problems:
            designs = [{"id": row["id"], "answer": row["answer"]} for row in results if row["problem"] == key]
            rng.shuffle(designs)
            packets.append({"problem": PROTO["problems"][key], "designs": designs})
        write(work / f"review{judge}.json", packets)


def summarize(directory):
    rows = read(directory / "answers.json")
    ids = {r["id"] for r in rows}
    by_arm, by_problem = defaultdict(list), defaultdict(lambda: defaultdict(list))
    scores = []
    for judge in range(3):
        points = read(directory / f"ratings{judge}.json")
        if set(points) != ids:
            raise ValueError(f"missing or extra IDs in judge {judge}")
        for row in rows:
            rating = points[row["id"]]
            if len(rating) != 3 or not all(type(v) is int and 1 <= v <= 5 for v in rating):
                raise ValueError(f"bad ratings for {row['id']}")
            by_arm[row["condition"]].append(rating)
            by_problem[row["problem"]][row["condition"]].append(rating)
        scores.append(points)
    def means(vals):
        return dict(zip(METRICS, (round(mean(v), 3) for v in zip(*vals))))
    comparable = agreed = tied = 0
    for j, k in combinations(scores, 2):
        for problem in PROTO["problems"]:
            members = [r["id"] for r in rows if r["problem"] == problem]
            for a, b in combinations(members, 2):
                product = (j[a][0] - j[b][0]) * (k[a][0] - k[b][0])
                if product == 0:
                    tied += 1
                else:
                    comparable += 1
                    agreed += product > 0
    return {"means": {k: means(v) for k, v in by_arm.items()},
            "by_problem": {p: {k: means(v) for k, v in arms.items()} for p, arms in by_problem.items()},
            "originality_order": {"agree": agreed, "disagree": comparable-agreed, "either_tied": tied},
            "word_cap_applied": sum(r["original_words"] > 110 for r in rows)}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=["blind", "score"])
    parser.add_argument("directory", type=Path)
    args = parser.parse_args()
    if args.mode == "blind":
        blind(args.directory)
    else:
        print(json.dumps(summarize(args.directory), indent=2))
