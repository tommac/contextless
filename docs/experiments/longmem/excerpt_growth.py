"""Compare evidence packing on pinned LongMemEval-S, with exact UTF-8 source addresses.

python3 excerpt_growth.py DATA.json RANKS.jsonl [--cases FILE]
"""

import argparse
from collections import Counter
from dataclasses import dataclass
import json
import math
from pathlib import Path
import re
import statistics
import sys

from candidate_ceiling import load_ranks, round_robin
from compare import BUDGET, dataset
from session_routing import STOP, interleave, secondary


@dataclass(frozen=True)
class Source:
    addr: tuple
    id: str
    date: str
    role: str
    start: int
    text: bytes

    def header(self, lo, hi):
        ref = self.id
        if (lo, hi) != (0, len(self.text)):
            ref += f"@{self.start + lo}:{self.start + hi}"
        return f"[{ref} {self.date} {self.role}] ".encode()

    def cost(self, lo, hi):
        return len(self.header(lo, hi)) + hi - lo + 1


def whole_pack(pool):
    selected, left = [], BUDGET
    for src in pool:
        cost = src.cost(0, len(src.text))
        if cost <= left:
            left -= cost
            selected.append((src, 0, len(src.text)))
    return selected


def terms(text):
    return {x for x in re.findall(r"\w{3,}", text.casefold()) if x not in STOP}


def boundaries(text):
    # Cuts follow punctuation, a newline, or 32 non-whitespace tokens. Each end
    # is at a UTF-8 boundary. Consecutive spans reconstruct every original byte.
    ends = [0]
    words = 0
    for match in re.finditer(rb"\S+\s*", text):
        words += 1
        chunk = match.group()
        if words == 32 or b"\n" in chunk or chunk.rstrip().endswith((b".", b"!", b"?")):
            ends.append(match.end())
            words = 0
    if ends[-1] != len(text):
        ends.append(len(text))
    return ends


def grow(pool, question, use_terms):
    """Buy one seed per candidate in source-diverse rank order, then expand in rounds."""
    diverse = [src for _, src in round_robin([(s.addr, s) for s in pool])]
    query = terms(question)
    counts = Counter(t for src in diverse for t in terms(src.text.decode()) & query)
    weights = {t: 1 + math.log((1 + len(diverse)) / (1 + n)) for t, n in counts.items()}
    remaining = BUDGET
    chosen = []
    for src in diverse:
        ends = boundaries(src.text)
        if len(ends) < 2:
            continue
        best = 0
        if use_terms:
            def relevance(i):
                words = terms(src.text[ends[i]:ends[i + 1]].decode())
                return sum(weights[t] for t in sorted(words & weights.keys()))
            best = max(range(len(ends) - 1), key=relevance)
        cost = src.cost(ends[best], ends[best + 1])
        if cost <= remaining:
            chosen.append([src, ends, best, best, best + 1])
            remaining -= cost

    while True:
        changed = False
        for entry in chosen:
            src, ends, anchor, lo, hi = entry
            # Alternate outward. Try the other side if the first cannot fit.
            neighbors = []
            if lo:
                neighbors.append((anchor - lo, lo - 1, hi))
            if hi < len(ends) - 1:
                neighbors.append((hi - anchor - 1, lo, hi + 1))
            old = src.cost(ends[lo], ends[hi])
            for _, lower, upper in sorted(neighbors):
                added = src.cost(ends[lower], ends[upper]) - old
                if added <= remaining:
                    remaining -= added
                    entry[3:] = [lower, upper]
                    changed = True
                    break
        if not changed:
            return [(src, ends[lo], ends[hi]) for src, ends, _, lo, hi in chosen]


def reader_input(q):
    return {
        "question": q["question"],
        "haystack_dates": q["haystack_dates"],
        "haystack_sessions": [
            [{"role": msg["role"], "content": msg["content"]} for msg in session]
            for session in q["haystack_sessions"]
        ],
    }


def candidate_pools(q, ranks):
    # Only text, role, date, the question and the saved production ranks reach the selector.
    turns = q["haystack_sessions"]
    dates = q["haystack_dates"]
    def source(s, m, id_, lo, hi):
        msg = turns[s][m]
        raw = msg["content"].encode()
        if not 0 <= lo < hi <= len(raw):
            raise ValueError("invalid ranked source offset")
        return Source((s, m), id_, dates[s], msg["role"], lo, raw[lo:hi])
    atoms = [source(h["Session"], h["Message"], h["ID"], h["Start"], h["End"])
             for h in ranks]
    _, routed = secondary(q)
    seq = 1
    offsets = []
    for messages in turns:
        offsets.append(seq)
        seq += len(messages)
    routed_pool = []
    for (s, m), _ in routed:
        raw = turns[s][m]["content"].encode()
        if raw:
            item = source(s, m, f"t{offsets[s]+m:06}#full", 0, len(raw))
            routed_pool.append((item.addr, item))
        # A full message once skipped due to atom overlap stays ineligible.
        # At most 24 original atoms plus 24 new messages can affect the result.
        if len(routed_pool) >= 48:
            break
    mixed = interleave([(s.addr, s) for s in atoms], routed_pool)
    return {"atoms": atoms, "routed": [src for _, src in mixed]}


def normalize(text):
    return " ".join(text.casefold().split())


def reference_pattern(answer):
    return re.compile(r"(?<!\w)" + re.escape(normalize(str(answer))) + r"(?!\w)")


def select(pool, question, method):
    if method == "whole":
        return whole_pack(pool)
    return grow(pool, question, use_terms=(method == "query_grow"))


def render(selected):
    return b"".join(src.header(lo, hi) + src.text[lo:hi] + b"\n" for src, lo, hi in selected)


METHODS = ("whole", "first_grow", "query_grow")


def evaluate(q, pools):
    gold = {(s, m): msg["content"].encode()
            for s, turns in enumerate(q["haystack_sessions"]) for m, msg in enumerate(turns)
            if msg.get("has_answer")}
    if not gold:
        raise ValueError("no annotated source messages on answerable question")
    match = reference_pattern(q["answer"]).search
    eligible = any(match(normalize(raw.decode())) for raw in gold.values())
    row = {
        "id": q["question_id"],
        "type": q["question_type"],
        "literal_eligible": eligible,
        "answer_in_question": match(normalize(q["question"])) is not None,
        "arms": {},
    }
    for lane, pool in pools.items():
        for method in METHODS:
            selected = select(pool, q["question"], method)
            used = len(render(selected))
            if used > BUDGET:
                raise ValueError("packet overflow")
            covered = set()
            literal = False
            for src, lo, hi in selected:
                s, m = src.addr
                raw = q["haystack_sessions"][s][m]["content"].encode()
                excerpt = src.text[lo:hi]
                if not excerpt or raw[src.start + lo:src.start + hi] != excerpt:
                    raise ValueError("excerpt did not resolve to original bytes")
                if src.addr in gold:
                    covered.add(src.addr)
                    literal |= match(normalize(excerpt.decode())) is not None
            row["arms"][f"{lane}_{method}"] = {
                "all_messages": covered == gold.keys(), "literal": literal, "bytes": used,
            }
    return row


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("data", type=Path)
    parser.add_argument("ranks", type=Path)
    parser.add_argument("--cases", type=Path)
    parser.add_argument("--show", metavar="QUESTION_ID", help="render one exact evidence packet to stdout")
    arms = [f"{lane}_{method}" for lane in ("atoms", "routed") for method in METHODS]
    parser.add_argument("--arm", choices=arms, help="packing method for --show")
    args = parser.parse_args()
    if bool(args.show) != bool(args.arm) or (args.show and args.cases):
        parser.error("use --show and --arm together for a single packet, or --cases for scoring")
    questions, ranks = dataset(args.data), load_ranks(args.ranks)
    if questions.keys() != ranks.keys():
        raise ValueError("dataset and rank keys differ")
    if args.show:
        if args.show not in questions:
            parser.error("unknown public question ID")
        q = reader_input(questions[args.show])
        lane, method = args.arm.split("_", 1)
        pools = candidate_pools(q, ranks[args.show]["Contextless"])
        sys.stdout.buffer.write(render(select(pools[lane], q["question"], method)))
        return

    cases = []
    for id_, q in questions.items():
        if id_.endswith("_abs"):
            continue
        pools = candidate_pools(reader_input(q), ranks[id_]["Contextless"])
        cases.append(evaluate(q, pools))
    summary = {"answerable": len(cases), "abstention_excluded": len(questions) - len(cases),
               "budget_bytes": BUDGET, "literal_eligible": sum(c["literal_eligible"] for c in cases), "arms": {}}
    for arm in cases[0]["arms"]:
        totals = {"median_bytes": statistics.median(c["arms"][arm]["bytes"] for c in cases)}
        eligible = [c for c in cases if c["literal_eligible"]]
        for name, metric, group in [
            ("all_messages", "all_messages", cases),
            ("literal", "literal", eligible),
            ("literal_absent_from_question", "literal", [c for c in eligible if not c["answer_in_question"]]),
        ]:
            bits = [(c["arms"]["atoms_whole"][metric], c["arms"][arm][metric]) for c in group]
            totals[name] = {
                "n": len(group),
                "count": sum(b for _, b in bits),
                "won": bits.count((False, True)),
                "lost": bits.count((True, False)),
            }
        summary["arms"][arm] = totals
    if args.cases:
        args.cases.write_text(json.dumps(cases, separators=(",", ":")) + "\n")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
