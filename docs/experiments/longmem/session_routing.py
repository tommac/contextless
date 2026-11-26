"""Fixed, model-free candidate discovery on pinned LongMemEval-S.

Run: python3 session_routing.py DATA.json RANKS.jsonl

Use the original ContextlessAI atoms in lane one. In lane two, use Porter
FTS5 BM25 to select three timestamped sessions, then rank whole messages from
those sessions with the whole-history message index's IDF. Alternate lanes
starting with an atom. To avoid overlapping material, accept a whole message
only if no atom of it has been selected; selecting the whole message suppresses
later atoms of it. Keep at most 24 candidates. The control follows the same
rule without session filtering. All thresholds are fixed across questions.
"""

import argparse
from collections import Counter
from contextlib import closing
import json
from pathlib import Path
import re
import sqlite3
import time

from candidate_ceiling import load_ranks
from compare import BUDGET, dataset, packet, size

STOP = set(
    """the and for that this with you are was but not have has can what how why
when who which will would should could from into about there their then than them they
your our its also just like any all out use using want need please""".split()
)
LIMIT = 24


def query_terms(text):
    words = dict.fromkeys(
        w for w in re.findall(r"\w{3,}", text.lower()) if w not in STOP
    )
    return " OR ".join(f'"{w}"' for w in list(words)[:32])


def secondary(q):
    # Only contents, roles and timestamps go into either index.
    addresses = {}
    turns = []
    sessions = []
    for sid, (date, messages) in enumerate(
        zip(q["haystack_dates"], q["haystack_sessions"])
    ):
        session = [date]
        for mid, msg in enumerate(messages):
            row = len(turns) + 1
            body = msg["content"]
            text = msg["role"] + " " + body
            session.append(text)
            turns.append((row, sid, text))
            addresses[row] = (
                (sid, mid),
                size(f"[t{row:06}#full {date} {msg['role']}] {body}\n"),
            )
        sessions.append((sid + 1, "\n".join(session)))
    query = query_terms(q["question"])
    if not query:
        return [], []
    with closing(sqlite3.connect(":memory:")) as db:
        db.execute(
            "create virtual table sessions using fts5(body, tokenize='porter unicode61')"
        )
        db.execute(
            "create virtual table turns using fts5(sid unindexed, body, tokenize='porter unicode61')"
        )
        db.executemany("insert into sessions(rowid,body) values(?,?)", sessions)
        db.executemany("insert into turns(rowid,sid,body) values(?,?,?)", turns)
        best = [
            r[0] - 1
            for r in db.execute(
                "select rowid from sessions where sessions match ? order by bm25(sessions),rowid limit 3",
                (query,),
            )
        ]
        sql = "select rowid from turns where turns match ?"
        order = " order by bm25(turns),rowid"
        # Read complete ordered lanes; duplicates may disqualify early candidates.
        flat = [addresses[r[0]] for r in db.execute(sql + order, (query,))]
        routed = []
        if best:
            select = " and sid in (" + ",".join("?" for _ in best) + ")"
            routed = [
                addresses[r[0]]
                for r in db.execute(sql + select + order, (query, *best))
            ]
    return flat, routed


def interleave(atoms, whole):
    primary, secondary_lane = iter(atoms), iter(whole)
    opened, full, out = set(), set(), []
    while len(out) < LIMIT:
        previous_length = len(out)
        for message, length in primary:
            if message not in full:
                out.append((message, length))
                opened.add(message)
                break
        if len(out) == LIMIT:
            break
        for message, length in secondary_lane:
            if message not in opened:
                out.append((message, length))
                full.add(message)
                opened.add(message)
                break
        if len(out) == previous_length:
            break
    return out


def score(candidates, required):
    opened, shortest = set(), {}
    left = BUDGET
    for message, length in candidates:
        shortest[message] = min(shortest.get(message, length), length)
        if length <= left:
            left -= length
            opened.add(message)
    present = required <= shortest.keys()
    feasible = present and sum(shortest[m] for m in required) <= BUDGET
    return {
        "complete": required <= opened,
        "present": present,
        "feasible": feasible,
        "bytes": BUDGET - left,
    }


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("data", type=Path)
    p.add_argument("ranks", type=Path)
    args = p.parse_args()
    questions, ranks = dataset(args.data), load_ranks(args.ranks)
    if questions.keys() != ranks.keys():
        raise ValueError("public dataset and ranking case IDs differ")
    stats = {
        arm: Counter()
        for arm in ("atoms", "atoms_plus_flat_messages", "atoms_plus_routed_messages")
    }
    start = time.monotonic()
    n = 0
    for qid, q in questions.items():
        if qid.endswith("_abs"):
            continue
        atoms = [
            ((h["Session"], h["Message"]), size(packet(q, h)))
            for h in ranks[qid]["Contextless"]
        ]
        flat, routed = secondary(q)
        lanes = [atoms, interleave(atoms, flat), interleave(atoms, routed)]
        required = {
            (s, m)
            for s, messages in enumerate(q["haystack_sessions"])
            for m, msg in enumerate(messages)
            if msg.get("has_answer")
        }
        if not required:
            raise ValueError(f"missing labels for {qid}")
        base = score(atoms, required)
        for (arm, counts), lane in zip(stats.items(), lanes):
            result = score(lane, required)
            counts.update(result)
            counts["won"] += result["complete"] and not base["complete"]
            counts["lost"] += not result["complete"] and base["complete"]
            counts["newly_reachable"] += result["present"] and not base["present"]
            counts["newly_unreachable"] += not result["present"] and base["present"]
            counts["previously_unreachable_now_complete"] += (
                result["complete"] and not base["present"]
            )
        n += 1
    print(
        json.dumps(
            {
                "answerable": n,
                "max_candidates": LIMIT,
                "max_bytes": BUDGET,
                "elapsed_seconds": round(time.monotonic() - start, 2),
                "arms": stats,
            },
            indent=2,
        )
    )


if __name__ == "__main__":
    main()
