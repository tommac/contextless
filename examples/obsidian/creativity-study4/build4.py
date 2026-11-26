#!/usr/bin/env python3
"""Round-4 two-phase packet builder.

Phase 1 (before any reader runs): question + labeled packets, runs 0-7.
Phase 2 (after out0-3.json exist): combined packets, runs 8-11; each combined
packet carries that repeat's question-only draft plus the labeled notes.
"""
import json
from pathlib import Path
import sys

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE.parent / "studytool"))
import studytool  # noqa: E402

PROTOCOL = json.loads((HERE / "protocol.json").read_text())
SELECTION = json.loads((HERE / "selection.json").read_text())


def phase1(outdir):
    phase1_protocol = HERE / "protocol-phase1.json"
    phase1_protocol.write_text(json.dumps(
        {**PROTOCOL, "conditions": ["question", "labeled"]}, indent=2))
    return studytool.build(phase1_protocol, HERE / "selection.json", outdir)


def labeled_cards():
    _, prose = studytool.vault_prose()
    cards = {}
    for key, picks in SELECTION.items():
        cards[key] = [{"title": Path(p["path"]).stem,
                       "text": studytool.excerpt(prose[p["path"]]),
                       "relationship": p["label"]} for p in picks]
    return cards


def phase2(outdir):
    outdir = Path(outdir)
    cards = labeled_cards()
    arm_index = PROTOCOL["conditions"].index("combined")
    for rep in range(PROTOCOL["repeats"]):
        drafts = json.loads((outdir / f"out{rep}.json").read_text())
        packets = {}
        for key, problem in PROTOCOL["problems"].items():
            packets[key] = {"problem": problem, "draft": drafts[key], "material": cards[key]}
        keys = list(packets)
        shift = (arm_index + rep) % len(keys)
        order = keys[shift:] + keys[:shift]
        if rep % 2:
            order.reverse()
        batch = {k: packets[k] for k in order}
        number = arm_index * PROTOCOL["repeats"] + rep
        (outdir / f"input{number}.json").write_text(json.dumps(
            {"instructions": PROTOCOL["combined_reader"], "problems": batch},
            ensure_ascii=False, indent=2))
        print(f"input{number}.json built (drafts from out{rep}.json)")


if __name__ == "__main__":
    audit = phase1(sys.argv[1]) if sys.argv[2] == "1" else phase2(sys.argv[1])
    if audit:
        print(json.dumps(audit, indent=2))
