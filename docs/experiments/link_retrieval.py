"""Known-item retrieval on analyst-authored questions. Run once from any directory."""

import collections
import math
import operator
from pathlib import Path
import re
import sqlite3

import vault_graph

VAULT = Path(__file__).resolve().parents[2] / "idea-vault"
# Named targets defined before comparing ranks. Other valid notes are unjudged.
PROBES = [
    (
        "How often must the provider reuse a prefix for our packet layout to pay off?",
        "Spine and epoch break even once 14 to 29 percent of calls hit",
    ),
    (
        "How sensitive was our experiment to small differences between paragraph and turn selection?",
        "The atoms comparison could only detect a gap near 10.5 points",
    ),
    (
        "Does word normalization help both small and large retrieval units?",
        "Stemming helped atoms by 6 to 17 percent and hurt whole turns",
    ),
    (
        "What was the effect of returning more context around each search hit?",
        "Recall search found 108 of 150 at one paragraph and 127 at four",
    ),
    (
        "Can a judge give credit for a vague answer that is factually wrong?",
        "A LoCoMo judge accepted 62.81 percent of vague wrong answers",
    ),
    (
        "How frequently do references point to evidence the answer did not rely on?",
        "Up to 57 percent of citations were unfaithful in adversarial tests",
    ),
    (
        "What is the evidence that a fixed team size is optimal?",
        "Three to five agents is a rule of thumb with no measured curve",
    ),
    (
        "What happens to input size when the provider cache stays cold?",
        "With no cache hits the stable layout raises input per message",
    ),
    (
        "Did splitting text into self-contained claims help a strong search system?",
        "Propositions helped weak retrievers and not strong ones",
    ),
    (
        "Were LANTERN retrieval failures mostly missing candidates or ordering mistakes?",
        "LANTERN misses were 361 of 420 coverage and 59 of 420 ranking",
    ),
    (
        "When does average evidence recall conceal missing pieces of a full answer?",
        "With BM25 alone complete coverage was 73.19 against 81.83 mean recall",
    ),
    (
        "How many source passages were reachable with the complete transcript and what did that cost?",
        "Floor whole history held all 120 answers at 297525 tokens",
    ),
]
STOP = set(
    "a an at by can did do does for from how in is it of on or our than that the this to was were what when with".split()
)


def run():
    notes = vault_graph.load(VAULT)
    names = {n.name for n in notes}
    maps = {m: collections.defaultdict(set) for m in ("tested-by", "all")}
    with sqlite3.connect(":memory:") as db:
        db.execute(
            "create virtual table docs using fts5(name,body,tokenize='porter unicode61')"
        )
        for note in notes:
            # A relation target in the searchable text would leak the link into the flat arm.
            prose = note.text.split("## Evidence")[0].split("## Relations")[0]
            db.execute("insert into docs values (?,?)", (note.name, prose))
            for kind, dst in note.rels:
                if dst in names:
                    for mode in ("all", "tested-by"):
                        if mode == "all" or kind == mode:
                            maps[mode][note.name].add(dst)
                            maps[mode][dst].add(note.name)
        scores = collections.defaultdict(list)
        for question, target in PROBES:
            assert target in names, target
            words = [
                w for w in re.findall(r"[a-z0-9]+", question.lower()) if w not in STOP
            ]
            match = " OR ".join('"' + w + '"' for w in words)
            ranked = [
                r[0]
                for r in db.execute(
                    "select name from docs where docs match ? order by bm25(docs), name",
                    (match,),
                )
            ]
            base = {name: 1 / (i + 1) for i, name in enumerate(ranked)}
            positions = []
            for mode in ("flat", "tested-by", "all"):
                weights = collections.Counter(base)
                if mode != "flat":
                    for seed in ranked[:3]:
                        neighbors = maps[mode][seed]
                        for dest in neighbors:
                            weights[dest] += base[seed] / math.sqrt(len(neighbors))
                ordered = sorted(weights, key=lambda n: (-weights[n], n))
                rank = ordered.index(target) + 1 if target in ordered else 0
                scores[mode].append(1 / rank if 0 < rank <= 5 else 0)
                positions.append(str(rank or "miss"))
            print(question, *positions, sep="\t")
        for mode, values in scores.items():
            print(
                mode,
                "hits@5",
                sum(v > 0 for v in values),
                "/",
                len(values),
                "MRR@5",
                round(sum(values) / len(values), 3),
                "improved/tied/worsened",
                [
                    sum(f(a, b) for a, b in zip(values, scores["flat"]))
                    for f in (operator.gt, operator.eq, operator.lt)
                ],
            )


if __name__ == "__main__":
    run()
