"""Shared creativity-study toolkit: packet building, blinding, scoring, calibration.

Consolidates the per-round copies in creativity-study{,2,3} into one tested
module and adds AI-only reviewer calibration: seeded designs with known
defects are mixed into the blind review, and each reviewer's catch rate is
measured from its ratings. No human scoring required.

Subcommands:
  build     protocol.json selection.json outdir   -> inputN.json + audit json
  blind     protocol.json workdir [--seeds f]     -> answers.json, review{0..2}.json
  score     protocol.json workdir                 -> means, agreement, calibration
"""
import argparse
from contextlib import closing
import json
from pathlib import Path
import random
import re
import sqlite3
from collections import defaultdict
from itertools import combinations
from statistics import mean
from time import perf_counter
import sys

ROOT = Path(__file__).resolve().parents[3]
sys.path.insert(0, str(ROOT / "docs/experiments"))
from link_retrieval import STOP  # noqa: E402
from vault_graph import load  # noqa: E402

METRICS = ("originality", "feasibility", "test")
SEED_BASE = 900  # seeded calibration designs take ids 901+, clear of 101-132


def excerpt(prose, cap=650):
    if len(prose.encode()) <= cap:
        return prose
    return prose.encode()[:cap].decode("utf-8", "ignore").rsplit(" ", 1)[0]


def vault_prose():
    vault = ROOT / "idea-vault"
    notes = load(str(vault))
    return vault, {Path(n.path).relative_to(vault).as_posix():
                   re.split(r"^## (?:Evidence|Relations)\s*$", n.text, flags=re.M)[0].strip()
                   for n in notes}


def build(protocol_path, selection_path, output):
    protocol = json.loads(Path(protocol_path).read_text())
    selections = json.loads(Path(selection_path).read_text()) if selection_path else None
    start = perf_counter()
    vault, prose = vault_prose()
    with closing(sqlite3.connect(":memory:")) as db:
        db.execute("CREATE VIRTUAL TABLE docs USING fts5(path UNINDEXED, title, body, tokenize='porter unicode61')")
        db.executemany("INSERT INTO docs VALUES (?,?,?)",
                       [(p, Path(p).stem, body) for p, body in prose.items()])
        indexed = perf_counter()
        search = {}
        for key, question in protocol["problems"].items():
            words = [w for w in re.findall(r"[a-z0-9]+", question.lower()) if w not in STOP]
            match = " OR ".join('"' + w + '"' for w in words)
            search[key] = [row[0] for row in db.execute(
                "SELECT path FROM docs WHERE docs MATCH ? ORDER BY rank,path LIMIT 3", (match,))]
        searched = perf_counter()
    audit = {"vault_size": len(prose), "index_seconds": indexed - start,
             "searches_seconds": searched - indexed, "search": search, "source_bytes": {}}
    arms = [a for a in protocol["conditions"]
            if selections is not None or a not in ("selected", "labeled")]
    output = Path(output)
    output.mkdir(parents=True, exist_ok=True)
    for arm in arms:
        packets = {}
        for key, problem in protocol["problems"].items():
            if arm == "search":
                picked = [{"path": p} for p in search[key]]
            elif arm in ("selected", "labeled"):
                picked = selections[key]
            else:
                picked = []
            cards = []
            for item in picked:
                card = {"title": Path(item["path"]).stem, "text": excerpt(prose[item["path"]])}
                if arm == "labeled":
                    card["relationship"] = item["label"]
                cards.append(card)
            packets[key] = {"problem": problem}
            if cards:
                packets[key]["material"] = cards
            audit["source_bytes"].setdefault(arm, {})[key] = sum(len(c["text"].encode()) for c in cards)
        for rep in range(protocol["repeats"]):
            keys = list(packets)
            i = protocol["conditions"].index(arm)
            shift = (i + rep) % len(keys)
            order = keys[shift:] + keys[:shift]
            if rep % 2:
                order.reverse()
            batch = {k: packets[k] for k in order}
            number = i * protocol["repeats"] + rep
            (output / f"input{number}.json").write_text(json.dumps(
                {"instructions": protocol["reader"], "problems": batch}, ensure_ascii=False, indent=2))
    return audit


def run_numbers(protocol):
    for index, arm in enumerate(protocol["conditions"]):
        for rep in range(protocol["repeats"]):
            yield arm, rep, index * protocol["repeats"] + rep


def blind(protocol, workdir, seeds=None, shuffle_seed=32391):
    """Anonymize outN.json into answers.json + per-reviewer packets.

    seeds: optional list of {problem, defect, answer}; each gets an id in the
    901+ range and is recorded in seeds.json. Seeds never enter score means.
    """
    workdir = Path(workdir)
    ids = list(range(101, 900))
    random.Random(shuffle_seed).shuffle(ids)
    results = []
    for arm, rep, run in run_numbers(protocol):
        answers = json.loads((workdir / f"out{run}.json").read_text())
        if set(answers) != set(protocol["problems"]):
            raise ValueError(f"incomplete run {run}")
        for problem in sorted(answers):
            full = answers[problem]
            words = full.split()
            if not words:
                raise ValueError("empty design")
            clipped = " ".join(words[:110]) if len(words) > 110 else full
            results.append({"id": str(ids.pop()), "condition": arm, "repeat": rep + 1,
                            "problem": problem, "original_words": len(words), "answer": clipped,
                            **({"unclipped": full} if len(words) > 110 else {})})
    seed_rows = []
    if seeds:
        for i, seed in enumerate(seeds):
            if seed["problem"] not in protocol["problems"]:
                raise ValueError(f"seed for unknown problem {seed['problem']}")
            row = {"id": str(SEED_BASE + 1 + i), "condition": "seed", "repeat": 0,
                   "problem": seed["problem"], "defect": seed["defect"], "answer": seed["answer"]}
            seed_rows.append(row)
        (workdir / "seeds.json").write_text(json.dumps(
            {r["id"]: {"problem": r["problem"], "defect": r["defect"]} for r in seed_rows}, indent=2) + "\n")
    (workdir / "answers.json").write_text(json.dumps(results, ensure_ascii=False, indent=2) + "\n")
    everything = results + [{**r, "original_words": len(r["answer"].split())} for r in seed_rows]
    for judge in range(3):
        rng = random.Random(61920 + judge)
        packets = []
        problems = list(protocol["problems"])
        rng.shuffle(problems)
        for key in problems:
            designs = [{"id": row["id"], "answer": row["answer"]} for row in everything if row["problem"] == key]
            rng.shuffle(designs)
            packets.append({"problem": protocol["problems"][key], "designs": designs})
        (workdir / f"review{judge}.json").write_text(json.dumps(packets, ensure_ascii=False, indent=2) + "\n")


def summarize(protocol, workdir):
    workdir = Path(workdir)
    rows = json.loads((workdir / "answers.json").read_text())
    ids = {r["id"] for r in rows}
    seeds = {}
    if (workdir / "seeds.json").exists():
        seeds = json.loads((workdir / "seeds.json").read_text())
        ids |= set(seeds)
    by_arm, by_problem = defaultdict(list), defaultdict(lambda: defaultdict(list))
    scores = []
    for judge in range(3):
        points = json.loads((workdir / f"ratings{judge}.json").read_text())
        if set(points) != ids:
            raise ValueError(f"missing or extra IDs in judge {judge}: "
                             f"missing {sorted(ids - set(points))[:5]}, extra {sorted(set(points) - ids)[:5]}")
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
        for problem in protocol["problems"]:
            members = [r["id"] for r in rows if r["problem"] == problem]
            for a, b in combinations(members, 2):
                product = (j[a][0] - j[b][0]) * (k[a][0] - k[b][0])
                if product == 0:
                    tied += 1
                else:
                    comparable += 1
                    agreed += product > 0
    out = {"means": {k: means(v) for k, v in by_arm.items()},
           "by_problem": {p: {k: means(v) for k, v in arms.items()} for p, arms in by_problem.items()},
           "originality_order": {"agree": agreed, "disagree": comparable - agreed, "either_tied": tied},
           "word_cap_applied": sum(r["original_words"] > 110 for r in rows)}
    if seeds:
        out["calibration"] = calibrate(protocol, workdir, ratings=scores)
    return out


def calibrate(protocol, workdir, ratings=None):
    """Catch rates and leniency from seeded designs with known defects.

    A catch is a score of at most 2 on the metric the seed was built to fail:
    generic -> originality, breaks-constraint -> feasibility, weak-test -> test.
    """
    workdir = Path(workdir)
    seeds = json.loads((workdir / "seeds.json").read_text())
    defect_metric = {"generic": 0, "breaks-constraint": 1, "weak-test": 2}
    if ratings is None:
        ratings = [json.loads((workdir / f"ratings{j}.json").read_text()) for j in range(3)]
    real_ids = {r["id"] for r in json.loads((workdir / "answers.json").read_text())}
    per_judge = []
    for points in ratings:
        catches = total = 0
        seed_scores, real_scores = [], []
        for sid, meta in seeds.items():
            rating = points[sid]
            total += 1
            catches += rating[defect_metric[meta["defect"]]] <= 2
            seed_scores.append(mean(rating))
        for rid in real_ids:
            real_scores.append(mean(points[rid]))
        per_judge.append({"catch_rate": round(catches / total, 3), "catches": catches, "seeds": total,
                          "seed_mean_score": round(mean(seed_scores), 3),
                          "real_mean_score": round(mean(real_scores), 3)})
    return {"defect_metric": {k: METRICS[v] for k, v in defect_metric.items()},
            "per_judge": per_judge,
            "mean_catch_rate": round(mean(j["catch_rate"] for j in per_judge), 3)}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("mode", choices=["build", "blind", "score", "calibrate"])
    parser.add_argument("protocol", type=Path)
    parser.add_argument("workdir", type=Path)
    parser.add_argument("--selection", type=Path, default=None)
    parser.add_argument("--seeds", type=Path, default=None)
    args = parser.parse_args()
    protocol = json.loads(args.protocol.read_text())
    if args.mode == "build":
        print(json.dumps(build(args.protocol, args.selection, args.workdir), indent=2))
    elif args.mode == "blind":
        seeds = json.loads(args.seeds.read_text()) if args.seeds else None
        blind(protocol, args.workdir, seeds=seeds)
    else:
        fn = summarize if args.mode == "score" else calibrate
        print(json.dumps(fn(protocol, args.workdir), indent=2))
