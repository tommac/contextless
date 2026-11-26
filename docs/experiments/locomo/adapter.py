"""Adapt LoCoMo-10 to the frozen excerpt experiment. No model calls."""
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import re
import statistics

DATA_SHA = "79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4"
SELECTOR_SHA = "028f9324af76918a3626c93e58c1856b2551b1d4ee510b9b17f5536e19971131"
SCORING_ONLY_REVISION = "85c921390dcc4c549713df65ab28dc1b16714588ade5583abd58b4c23e632e36"
RANKER_SOURCE_SHA = "591f9c09e8b7ec1173185bc9b03b9f85517b4e406001dce5429382dad18491b7"


FROZEN_IMPORTS = {'session_routing.py': 'c5b7d1293abe2aae9f9c1aa61bdd9c38d4ca71ec9b1d649803dbd94c4d07cbca', 'candidate_ceiling.py': '92735818cc10bc6b0b7aba4725edc3d803e225b32414ce69b00af083087f476d', 'compare.py': '873011c5d4bddefac053bde44acade8d4ecbd10486afba99fb8ea6bd779d7595', 'baselines.py': '433b73ec147c57b263db9bbce587f077e5f1033b125b969fa084812adc9b71f7'}


def checked(path, sha):
    data = path.read_bytes()
    if hashlib.sha256(data).hexdigest() != sha:
        raise ValueError(f"unexpected SHA-256 for {path}")
    return data


def normalize_evidence(evidence, addresses):
    """Tolerate separators and lossless integer spelling. Never infer missing IDs."""
    resolved, strict = set(), True
    for field in evidence:
        if field in addresses:
            resolved.add(addresses[field])
            continue
        strict = False
        for identifier in re.split(r"[;,\s]+", field.strip()):
            match = re.fullmatch(r"D:?(\d+):(\d+)", identifier)
            if not match:
                return None, False
            standard = "D%d:%d" % tuple(map(int, match.groups()))
            if standard not in addresses:
                return None, False
            resolved.add(addresses[standard])
    return resolved, strict


def cases(data):
    seen = set()
    for dialogue in json.loads(checked(data, DATA_SHA)):
        conv = dialogue["conversation"]
        keys = sorted((k for k in conv if re.fullmatch(r"session_\d+", k)),
                      key=lambda k: int(k.split("_")[1]))
        sessions, addresses = [], {}
        for sid, k in enumerate(keys):
            msgs = []
            for mid, m in enumerate(conv[k]):
                if m["dia_id"] in addresses:
                    raise ValueError("duplicate native dialogue ID")
                addresses[m["dia_id"]] = (sid, mid)
                msgs.append({"role": m["speaker"], "content": m["text"]})
            sessions.append(msgs)
        dates = [conv[k + "_date_time"] for k in keys]
        for i, qa in enumerate(dialogue["qa"]):
            key = f"{dialogue['sample_id']}:{i}"
            if key in seen:
                raise ValueError("duplicate question key")
            seen.add(key)
            view = {"question_id": key, "question": qa["question"],
                    "haystack_sessions": sessions, "haystack_dates": dates}
            evidence, strict = normalize_evidence(qa.get("evidence", []), addresses)
            if qa["category"] == 5:
                status = "adversarial"
            elif not qa.get("evidence"):
                status = "no_labels"
            elif evidence is None:
                status = "unresolved"
            else:
                status = "strict" if strict else "reformatted"
            yield key, view, qa, evidence, status


def prepare(args):
    """Build a ranker input with no labels/answers and adjust only its dataset guards."""
    digest = hashlib.sha256()
    n = 0
    with args.inputs.open("wb") as out:
        def emit(chunk):
            digest.update(chunk)
            out.write(chunk)
        emit(b"[")
        for _, view, _, _, status in cases(args.data):
            if status not in ("strict", "reformatted"):
                continue
            if n:
                emit(b",\n")
            public = {k: view[k] for k in ("question_id", "question", "haystack_sessions")}
            emit(json.dumps(public, separators=(",", ":"), ensure_ascii=False).encode())
            n += 1
        emit(b"]")
    original = checked(args.original_go, RANKER_SOURCE_SHA).decode()
    old_digest = "d6f21ea9d60a0d56f34a05b609c79c88a451d2ae03597821ea3d5a9678c3a442"
    if original.count(old_digest) != 1 or original.count("count != 500") != 1:
        raise ValueError("unknown Go adapter guards")
    args.adapted_go.write_text(original.replace(old_digest, digest.hexdigest())
                               .replace("count != 500", f"count != {n}"))
    print(f"Prepared {n} questions, SHA-256 {digest.hexdigest()}")


def summarize(rows):
    if not rows:
        return {"questions": 0}
    result = {"questions": len(rows), "literal_eligible": sum(r["literal_eligible"] for r in rows), "arms": {}}
    result["pools_fitting_budget"] = {k: sum(r["fits"][k] for r in rows) for k in rows[0]["fits"]}
    for name in rows[0]["arms"]:
        arm = {}
        for label in ("all_messages", "literal", "literal_absent_from_question"):
            metric = "all_messages" if label == "all_messages" else "literal"
            selected = [r for r in rows if metric != "literal" or r["literal_eligible"]]
            if label == "literal_absent_from_question":
                selected = [r for r in selected if not r["answer_in_question"]]
            baseline = name.split("_")[0] + "_atoms_whole"
            paired = [(r["arms"][baseline][metric], r["arms"][name][metric]) for r in selected]
            arm[label] = {"denominator": len(paired), "count": sum(b for a, b in paired), "wins_from_atoms_whole": paired.count((False, True)),
                           "losses_from_atoms_whole": paired.count((True, False))}
        arm["median_bytes"] = statistics.median(r["arms"][name]["bytes"] for r in rows)
        result["arms"][name] = arm
    return result


def score(args):
    import sys
    selector = hashlib.sha256((args.selectors / "excerpt_growth.py").read_bytes()).hexdigest()
    if selector not in (SELECTOR_SHA, SCORING_ONLY_REVISION):
        raise ValueError("unexpected excerpt selector revision")
    for filename, digest in FROZEN_IMPORTS.items():
        checked(args.selectors / filename, digest)
    sys.path.insert(0, str(args.selectors.resolve()))
    from excerpt_growth import candidate_pools, evaluate, normalize, reference_pattern
    from compare import BUDGET

    rankings = {r["QID"]: r for r in (json.loads(line) for line in args.ranks.open())}
    rows, excluded, statuses = [], [], Counter()
    for key, view, qa, gold, status in cases(args.data):
        statuses[status] += 1
        if status not in ("strict", "reformatted"):
            excluded.append({"id": key, "reason": status})
            continue
        rank = rankings.pop(key)
        echoed = reference_pattern(qa["answer"]).search(normalize(qa["question"])) is not None
        # Output rows contain no conversations, reference answers or questions.
        row = {"id": key, "dialogue": key.split(":")[0], "category": qa["category"],
               "id_resolution": status, "answer_in_question": echoed, "annotated_turns": len(gold), "fits": {}, "arms": {}}
        scorer = {**view, "answer": qa["answer"], "question_type": qa["category"],
                  "haystack_sessions": [
                      [dict(m, has_answer=(s, i) in gold) for i, m in enumerate(session)]
                      for s, session in enumerate(view["haystack_sessions"])]}
        for lane in ("BM25", "Contextless"):
            pools = candidate_pools(view, rank[lane] or [])
            for label, pool in pools.items():
                row["fits"][f"{lane}_{label}"] = sum(src.cost(0, len(src.text)) for src in pool) <= BUDGET
            result = evaluate(scorer, pools)
            row["literal_eligible"] = result["literal_eligible"]
            row["arms"].update({f"{lane}_{arm}": v for arm, v in result["arms"].items()})
        rows.append(row)
    if rankings:
        raise ValueError("unexpected rankings")
    args.case_output.write_text(json.dumps({"included": rows, "excluded": excluded}, separators=(",", ":")) + "\n")
    summary = {"dataset_sha256": DATA_SHA, "frozen_selector_sha256": SELECTOR_SHA, "executed_selector_sha256": selector, "budget_bytes": 8000,
               "eligibility_counts": dict(statuses), "all": summarize(rows),
               "strict": summarize([r for r in rows if r["id_resolution"] == "strict"]),
               "by_dialogue": {d: summarize([r for r in rows if r["dialogue"] == d])
                               for d in sorted({r["dialogue"] for r in rows})},
               "by_category": {c: summarize([r for r in rows if r["category"] == c])
                               for c in sorted({r["category"] for r in rows})}}
    args.output.write_text(json.dumps(summary, indent=2) + "\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    subs = parser.add_subparsers(dest="command", required=True)
    p = subs.add_parser("prepare")
    for field in ("data", "original_go", "inputs", "adapted_go"):
        p.add_argument(field, type=Path)
    p = subs.add_parser("score")
    for field in ("data", "selectors", "ranks", "case_output", "output"):
        p.add_argument(field, type=Path)
    args = parser.parse_args()
    {"prepare": prepare, "score": score}[args.command](args)
