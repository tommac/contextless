"""Audit attribution on naturally paired LoCoMo questions using their annotated turns."""

import argparse
from collections import Counter, defaultdict
import json
from pathlib import Path
import re

from adapter import DATA_SHA, checked


def normalize(value):
    return " ".join(re.findall(r"\w+", str(value).casefold()))


def role_accepts(question, names, roles):
    """Given a question and source speakers, would this rule accept its evidence?"""
    subject = {m.casefold() for m in names.findall(question)}
    return len(subject) == 1 and roles == subject


def audit_dialogue(dialogue):
    conversation = dialogue["conversation"]
    a, b = conversation["speaker_a"], conversation["speaker_b"]
    other = {a.casefold(): b, b.casefold(): a}
    names = re.compile(r"\b(?:" + "|".join(map(re.escape, [a, b])) + r")\b", re.I)
    turns = {}
    for key in conversation:
        if re.fullmatch(r"session_\d+", key):
            for turn in conversation[key]:
                address = turn["dia_id"]
                if address in turns:
                    raise ValueError(f"Duplicate turn ID {address}")
                turns[address] = turn["speaker"].casefold()

    questions = dialogue["qa"]
    by_question = defaultdict(list)
    for index, question in enumerate(questions):
        if question["category"] != 5:
            by_question[normalize(question["question"])].append(index)

    counts = Counter()
    pairs, ambiguous = [], []
    for index, item in enumerate(questions):
        if item["category"] != 5:
            continue
        counts["adversarial_questions"] += 1
        q = item["question"]
        swapped = names.sub(lambda m: other[m[0].casefold()], q)
        key = normalize(swapped)
        found = by_question.get(key, [])
        if key == normalize(q) or not found:
            continue
        counts["matched_question"] += 1

        evidence = sorted(item.get("evidence", []))
        matches = [i for i in found if evidence and sorted(questions[i].get("evidence", [])) == evidence]
        if not matches:
            continue
        counts["matched_evidence"] += 1
        lure = item.get("adversarial_answer")
        if lure is None:
            continue
        matches = [i for i in matches if normalize(questions[i]["answer"]) == normalize(lure)]
        if not matches:
            continue
        counts["matched_lure"] += 1
        suffix = f"{dialogue['sample_id']}:"
        if len(matches) != 1:
            counts["ambiguous_positive"] += 1
            ambiguous.append({"adversarial": suffix + str(index),
                              "positives": [suffix + str(i) for i in matches], "evidence": evidence})
            continue
        if any(turn_id not in turns for turn_id in evidence):
            raise ValueError(f"Unresolved raw evidence for {suffix}{index}")
        pos = matches[0]
        pos_q = questions[pos]["question"]
        roles = {turns[turn_id] for turn_id in evidence}
        pos_accept = role_accepts(pos_q, names, roles)
        neg_accept = role_accepts(q, names, roles)
        counts["unique_pairs"] += 1
        counts["positive_accept"] += pos_accept
        counts["adversarial_decline"] += not neg_accept
        counts["both_correct"] += pos_accept and not neg_accept
        if pos_accept:
            speaker_case = "positive_subject_only"
        elif neg_accept:
            speaker_case = "opposite_subject_only"
        elif len(roles) > 1:
            speaker_case = "mixed_speakers"
        else:
            speaker_case = "undetermined_subject"
        counts[speaker_case] += 1
        pairs.append({"positive": suffix + str(pos), "adversarial": suffix + str(index),
                      "evidence": evidence, "positive_accept": pos_accept,
                      "adversarial_accept": neg_accept, "source_speakers": speaker_case})
    return counts, pairs, ambiguous


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("data", type=Path)
    parser.add_argument("--cases", type=Path, help="save paired question IDs, turn IDs and rule decisions")
    args = parser.parse_args()
    sample = json.loads(checked(args.data, DATA_SHA))
    total = Counter()
    by_dialogue, cases, ambiguous = {}, [], []
    for dialogue in sample:
        counts, pairs, duplicates = audit_dialogue(dialogue)
        total.update(counts)
        by_dialogue[dialogue["sample_id"]] = counts
        cases.extend(pairs)
        ambiguous.extend(duplicates)
    if len({p["positive"] for p in cases}) != len(cases):
        raise ValueError("A positive question occurs in multiple accepted pairs")
    print(json.dumps({
        "dataset_sha256": DATA_SHA, "counts": total, "by_dialogue": by_dialogue,
        "ambiguous": ambiguous, "role_rule_disagreements": [
            p for p in cases if not p["positive_accept"] or p["adversarial_accept"]
        ],
    }, indent=2, sort_keys=True))
    if args.cases:
        args.cases.write_text(json.dumps(cases, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
