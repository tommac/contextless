"""Rebuild the four creativity conditions and measure their local selection costs."""
import argparse
from contextlib import closing
import json
from pathlib import Path
import re
import sqlite3
import sys
from time import perf_counter

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
sys.path.insert(0, str(ROOT / "docs/experiments"))
from link_retrieval import STOP
from vault_graph import load


def excerpt(prose):
    if len(prose.encode()) <= 650:
        return prose
    return prose.encode()[:650].decode("utf-8", "ignore").rsplit(" ", 1)[0]


def build(output, selections):
    protocol = json.loads((HERE / "protocol.json").read_text())
    start = perf_counter()
    vault = ROOT / "idea-vault"
    notes = load(str(vault))
    prose = {Path(n.path).relative_to(vault).as_posix():
             re.split(r"^## (?:Evidence|Relations)\s*$", n.text, flags=re.M)[0].strip()
             for n in notes}
    with closing(sqlite3.connect(":memory:")) as db:
        db.execute("CREATE VIRTUAL TABLE docs USING fts5(path UNINDEXED, title, body, tokenize='porter unicode61')")
        db.executemany("INSERT INTO docs VALUES (?,?,?)",
                       [(p, Path(p).stem, body) for p, body in prose.items()])
        indexed = perf_counter()
        search = {}
        for key, question in protocol["problems"].items():
            words = [w for w in re.findall(r"[a-z0-9]+", question.lower()) if w not in STOP]
            match = " OR ".join('"'+w+'"' for w in words)
            search[key] = [row[0] for row in db.execute(
                "SELECT path FROM docs WHERE docs MATCH ? ORDER BY rank,path LIMIT 3", (match,))]
        searched = perf_counter()
    audit = {"vault_size": len(notes), "index_seconds": indexed - start,
             "four_searches_seconds": searched - indexed, "search": search, "source_bytes": {}}
    arms = ["question", "search"]
    if selections is not None:
        arms += ["selected", "labeled"]
    for arm in arms:
        packets = {}
        for key, problem in protocol["problems"].items():
            cards = []
            if arm == "search":
                picked = [{"path": path} for path in search[key]]
            elif arm in ("selected", "labeled"):
                picked = selections[key]
            else:
                picked = []
            for item in picked:
                path = item["path"]
                card = {"title": Path(path).stem, "text": excerpt(prose[path])}
                if arm == "labeled":
                    card["relationship"] = item["label"]
                cards.append(card)
            packets[key] = {"problem": problem}
            if cards:
                packets[key]["material"] = cards
            audit["source_bytes"].setdefault(arm, {})[key] = sum(len(c["text"].encode()) for c in cards)
        for rep in range(2):
            # Latin-square rotation across arms and reverse for the second generation.
            keys = list(packets)
            i = protocol["conditions"].index(arm)
            order = keys[i:] + keys[:i]
            if rep:
                order.reverse()
            batch = {k: packets[k] for k in order}
            number = (i * 5 + rep * 4) % 8
            (output / f"input{number}.json").write_text(json.dumps(
                {"instructions": protocol["reader"], "problems": batch}, ensure_ascii=False, indent=2))
    return audit


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("output", type=Path)
    parser.add_argument("--selection", type=Path, default=HERE / "selection.json")
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    selected = json.loads(args.selection.read_text()) if args.selection.is_file() else None
    print(json.dumps(build(args.output, selected), indent=2))
