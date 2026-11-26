"""Build three text-only evidence packets for a selected eight-case LoCoMo pilot."""

import argparse
from contextlib import closing
import json
from pathlib import Path
import re
import sqlite3

from adapter import DATA_SHA, checked

CASES = (
    "conv-50:175", "conv-42:222", "conv-48:152", "conv-49:173",
    "conv-43:209", "conv-26:107", "conv-26:106", "conv-26:139",
)
STOP = set("a an at by can did do does for from how in is it of on or our than that the this to was were what when with".split())


def render(obj):
    return json.dumps(obj, ensure_ascii=False, indent=1)


def packet_size(packet):
    # Include the enclosing JSON array and all question metadata for each independent case.
    return len(render([packet]).encode("utf-8"))


def get_windows(sample, item, number):
    conversation = sample["conversation"]
    turns = []
    session_bounds = {}
    for session in sorted((k for k in conversation if re.fullmatch(r"session_\d+", k)),
                          key=lambda k: int(k.split("_")[1])):
        start = len(turns)
        turns.extend({k: t[k] for k in ("dia_id", "speaker", "text")} for t in conversation[session])
        stop = len(turns)
        for row in range(start, stop):
            session_bounds[row] = (start, stop, conversation[session + "_date_time"])
    gold, = item["evidence"]
    anchor, = [i for i,t in enumerate(turns) if t["dia_id"] == gold]
    start, stop, date = session_bounds[anchor]
    question = item["question"]

    def packet(rows):
        return {"id": number, "question": question, "date": date,
                "passages": [turns[r] for r in sorted(rows)]}

    nearby = packet(range(max(start, anchor - 1), min(stop, anchor + 2)))
    limit = packet_size(nearby)
    words = [w for w in re.findall(r"[a-z0-9]+", question.lower()) if w not in STOP]
    query = " OR ".join('"' + w + '"' for w in words)
    chosen = [anchor]
    with closing(sqlite3.connect(":memory:")) as db:
        db.execute("create virtual table turns using fts5(body, tokenize='porter unicode61')")
        db.executemany("insert into turns(rowid,body) values(?,?)",
                       ((i + 1, t["speaker"] + " " + t["text"]) for i,t in enumerate(turns)))
        if query:
            ranked = db.execute("select rowid from turns where turns match ? order by bm25(turns), rowid", (query,))
            for (rowid,) in ranked:
                candidate = rowid - 1
                if candidate != anchor and packet_size(packet(chosen + [candidate])) <= limit:
                    chosen.append(candidate)
                    if len(chosen) == 3:
                        break
    return packet([anchor]), nearby, packet(chosen)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("data", type=Path)
    parser.add_argument("output", type=Path, help="directory for reader packets containing public dialogue text")
    args = parser.parse_args()
    samples = {s["sample_id"]:s for s in json.loads(checked(args.data, DATA_SHA))}
    args.output.mkdir(parents=True, exist_ok=True)
    arms = {"anchor": [], "neighbors": [], "search": []}
    sizes = []
    for n, key in enumerate(CASES, 1):
        dialog, index = key.split(":")
        item = samples[dialog]["qa"][int(index)]
        packets = get_windows(samples[dialog], item, n)
        row = {"id": n}
        for (name, group), packet in zip(arms.items(), packets):
            group.append(packet)
            row[name] = packet_size(packet)
        sizes.append(row)
    for name, group in arms.items():
        (args.output / (name + ".json")).write_text(render(group), encoding="utf-8")
    print(json.dumps({"utf8_bytes_per_case_as_singleton_array": sizes}, indent=2))


if __name__ == "__main__":
    main()
