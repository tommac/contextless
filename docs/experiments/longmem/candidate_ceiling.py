"""Diagnose all-message coverage ceilings for the pinned public LongMemEval-S rankings.

Usage: python3 candidate_ceiling.py DATA.json RANKS.jsonl [--by-type]
       [--check-cases public-cases.json]

One selected atom represents its source message. Labels identify messages, not the
answer-bearing words. The oracle ceiling picks the smallest framed atom from
each required message. The deployable comparison only sees source positions,
framed byte lengths and rank order.
"""

import argparse
from collections import Counter
from itertools import zip_longest
import json
from pathlib import Path

from compare import BUDGET, dataset, packet, size


ARMS = ("BM25", "Contextless")
REASONS = (
    "complete_now",
    "required_message_missing_from_candidates",
    "all_present_minimum_exceeds_budget",
    "feasible_but_current_pack_misses",
)


def round_robin(candidates):
    """Visit distinct source messages in first-hit order, one atom per pass."""
    by_message = {}
    for candidate in candidates:
        by_message.setdefault(candidate[0], []).append(candidate)
    for group in zip_longest(*by_message.values()):
        yield from (item for item in group if item is not None)


def packed_messages(candidates):
    remaining = BUDGET
    opened = set()
    for message, length in candidates:
        if length <= remaining:
            remaining -= length
            opened.add(message)
    return opened


def classify(candidates, required):
    opened = packed_messages(candidates)
    rotated = packed_messages(round_robin(candidates))
    shortest = {}
    for message, length in candidates:
        if message not in shortest or length < shortest[message]:
            shortest[message] = length

    current = required <= opened
    alternative = required <= rotated
    if current:
        reason = REASONS[0]
    elif not required <= shortest.keys():
        reason = REASONS[1]
    elif sum(shortest[key] for key in required) > BUDGET:
        reason = REASONS[2]
    else:
        reason = REASONS[3]

    return reason, current, alternative


def load_ranks(path):
    rows = {}
    with path.open() as f:
        for line in f:
            row = json.loads(line)
            qid = row["QID"]
            if qid in rows:
                raise ValueError(f"duplicate public question {qid} in ranking file")
            for arm in ARMS:
                if len(row[arm]) > 24 or len({h["ID"] for h in row[arm]}) != len(
                    row[arm]
                ):
                    raise ValueError(
                        f"{qid}/{arm}: expected at most 24 distinct ranked atoms"
                    )
            rows[qid] = row
    return rows


def summarize(records):
    result = {}
    for arm in ARMS:
        counts = Counter(case[arm][0] for case in records)
        outcomes = Counter((case[arm][1], case[arm][2]) for case in records)
        result[arm] = {
            "current_pack": {reason: counts[reason] for reason in REASONS},
            "candidate_oracle_under_budget": counts[REASONS[0]] + counts[REASONS[3]],
            "round_robin_complete": outcomes[True, True] + outcomes[False, True],
            "round_robin_vs_current": {
                "won": outcomes[False, True],
                "lost": outcomes[True, False],
                "both_complete": outcomes[True, True],
                "both_incomplete": outcomes[False, False],
            },
        }
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("data", type=Path)
    parser.add_argument("ranks", type=Path)
    parser.add_argument("--by-type", action="store_true")
    parser.add_argument(
        "--check-cases", type=Path, help="validate both original per-case outcomes"
    )
    args = parser.parse_args()
    questions = dataset(args.data)
    ranks = load_ranks(args.ranks)
    if ranks.keys() != questions.keys():
        raise ValueError("incomplete rankings or extra question ids")
    references = {}
    if args.check_cases:
        references = {
            c["question"]: c for c in json.loads(args.check_cases.read_text())
        }
        if references.keys() != questions.keys():
            raise ValueError("reference cases do not match public dataset")

    records = []
    for qid, q in questions.items():
        if qid.endswith("_abs"):
            continue
        required = {
            (i, j)
            for i, session in enumerate(q["haystack_sessions"])
            for j, turn in enumerate(session)
            if turn.get("has_answer")
        }
        if not required:
            raise ValueError(f"missing answer turn labels for {qid}")
        case = {"type": q["question_type"]}
        union_messages = set()
        for arm in ARMS:
            candidates = [
                ((h["Session"], h["Message"]), size(packet(q, h)))
                for h in ranks[qid][arm]
            ]
            union_messages.update(message for message, _ in candidates)
            case[arm] = classify(candidates, required)
            if references:
                ref = references[qid][arm]
                if (
                    ref["required_turns"] != len(required)
                    or len(required & packed_messages(candidates))
                    != ref["covered_turns"]
                ):
                    raise ValueError(
                        f"original pack diverged from reference at {qid}/{arm}"
                    )
        case["missing_from_union"] = not required <= union_messages
        records.append(case)

    report = {
        "questions": len(questions),
        "answerable": len(records),
        "abstention_excluded": len(questions) - len(records),
        "budget_bytes": BUDGET,
        "max_candidates": 24,
        "missing_required_message_even_in_union": sum(
            r["missing_from_union"] for r in records
        ),
        "all_questions": summarize(records),
    }
    if args.by_type:
        report["by_question_type"] = {
            kind: summarize([r for r in records if r["type"] == kind])
            for kind in sorted({r["type"] for r in records})
        }
    print(json.dumps(report, indent=2))


if __name__ == "__main__":
    main()
