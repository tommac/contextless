"""Granularity experiment: whole turns vs paragraph atoms as retrieval units.

For each probe (question, source turn, also-turns, answer), rank units by
SQLite FTS5 bm25 with the same query rule as cx (lowercase words of 3+
characters, stopwords removed, OR). Fill a budget in rank order and report
the tokens needed before a unit holding the answer is included.
Read-only on the workspaces; no model calls.
"""
import json
import os
import re
import sqlite3
import statistics
import sys

TPB = 0.30  # cx DefaultTokensPerByte
BUDGETS = [2000, 4000, 8000, 16000, 32000, 64000]
STOP = set("""the and for that this with you are was but not have has can what how why when who which will
would should could from into about there their then than them they your our its also just like any all out
use using want need please""".split())
WORD = re.compile(r"[^\W]{3,}", re.UNICODE)
FENCE = re.compile(r"^\s*(```|~~~)")


def tok(s):
    return int(len(s.encode()) * TPB) + 1


def norm(s):
    return " ".join(s.lower().split())


def load_turns(ws):
    turns = []
    for fn in sorted(os.listdir(os.path.join(ws, "turns"))):
        if not fn.endswith(".md"):
            continue
        txt = open(os.path.join(ws, "turns", fn), encoding="utf-8").read()
        end = txt.index("\n---\n", 4)
        meta = {}
        for line in txt[4:end].splitlines():
            if ": " in line:
                k, v = line.split(": ", 1)
                meta[k] = v
        turns.append(dict(id=meta["id"], seq=int(meta["seq"]), role=meta["role"],
                          reply_to=meta.get("reply_to", ""), body=txt[end + 5:].strip()))
    return turns


def atoms_of(body, min_chars=200, max_chars=2400):
    blocks, cur, fence = [], [], False
    for line in body.split("\n"):
        if FENCE.match(line):
            fence = not fence
            cur.append(line)
            continue
        if not fence and not line.strip():
            if cur:
                blocks.append("\n".join(cur))
                cur = []
            continue
        cur.append(line)
    if cur:
        blocks.append("\n".join(cur))
    merged, buf = [], ""
    for b in blocks:  # short blocks (headings, one-liners) join the next one
        buf = buf + "\n\n" + b if buf else b
        if len(buf) >= min_chars:
            merged.append(buf)
            buf = ""
    if buf:
        if merged:
            merged[-1] += "\n\n" + buf
        else:
            merged.append(buf)
    out = []
    for m in merged:
        while len(m) > max_chars:
            cut = m.rfind("\n", 0, max_chars)
            if cut < max_chars // 3:
                cut = m.rfind(". ", 0, max_chars) + 1
            if cut < max_chars // 3:
                cut = max_chars
            out.append(m[:cut].strip())
            m = m[cut:].strip()
        if m:
            out.append(m)
    return out


def build_units(turns, kind):
    """Units: dicts with turn, k (atom index), text (sent), index (indexed), cost."""
    by_id = {t["id"]: t for t in turns}
    units = []
    for t in turns:
        if kind == "turn":
            units.append(dict(turn=t["id"], k=0, text=t["body"], index=t["body"], cost=tok(t["body"]) + 15))
            continue
        q = by_id.get(t["reply_to"])
        qline = q["body"][:400] if q else ""
        for k, a in enumerate(atoms_of(t["body"])):
            idx = a
            if kind == "atom+ctx" and qline:
                idx = qline + "\n\n" + a  # indexed with the question it answers; sent as a short header
            hdr = 25 + (tok(qline[:120]) if kind == "atom+ctx" and qline else 0)
            units.append(dict(turn=t["id"], k=k, text=a, index=idx, cost=tok(a) + hdr))
    return units


def rank(units, tokenizer, question):
    db = sqlite3.connect(":memory:")
    db.execute(f"CREATE VIRTUAL TABLE f USING fts5(text, tokenize='{tokenizer}')")
    db.executemany("INSERT INTO f(rowid, text) VALUES (?, ?)", [(i, u["index"]) for i, u in enumerate(units)])
    seen, terms = set(), []
    for w in WORD.findall(question.lower()):
        if w in STOP or w in seen:
            continue
        seen.add(w)
        terms.append('"' + w.replace('"', "") + '"')
    if not terms:
        return []
    rows = db.execute("SELECT rowid FROM f WHERE f MATCH ? ORDER BY bm25(f)", (" OR ".join(terms[:32]),)).fetchall()
    return [r[0] for r in rows]


def evaluate(units, order, targets, ans, neighbors=False):
    """Tokens and rank at which a unit holding the answer enters the packet."""
    pos = {(u["turn"], u["k"]): i for i, u in enumerate(units)}
    included, cost = set(), 0
    for r, i in enumerate(order, 1):
        group = [i]
        if neighbors:
            u = units[i]
            for dk in (-1, 1):
                j = pos.get((u["turn"], u["k"] + dk))
                if j is not None:
                    group.append(j)
        for j in group:
            if j not in included:
                included.add(j)
                cost += units[j]["cost"]
        for j in group:
            u = units[j]
            if u["turn"] in targets:
                nxt = pos.get((u["turn"], u["k"] + 1))
                joined = u["text"] + ("\n\n" + units[nxt]["text"] if nxt is not None else "")
                if ans in norm(u["text"]) or (ans in norm(joined) and nxt in included):
                    return cost, r
    return None, None


def main():
    P = sys.argv[1]
    ws_list = [("bw", f"{P}/cxroot/workspaces/bw", f"{P}/probes-bw.jsonl"),
               ("bwall", f"{P}/cxroot/workspaces/bwall", f"{P}/probes-bw.jsonl"),
               ("s8", f"{P}/cxsizing/workspaces/s8", f"{P}/probes-s8.jsonl")]
    variants = [("turn", "unicode61", False), ("turn", "porter unicode61", False),
                ("atom", "unicode61", False), ("atom", "porter unicode61", False),
                ("atom", "porter unicode61", True), ("atom+ctx", "porter unicode61", False),
                ("atom+ctx", "porter unicode61", True)]
    results = {}
    for name, ws, pf in ws_list:
        turns = load_turns(ws)
        ids = {t["id"] for t in turns}
        probes = [json.loads(l) for l in open(pf) if l.strip() and not l.startswith("#")]
        hist = sum(tok(t["body"]) for t in turns)
        unit_cache = {}
        print(f"\n== {name}: {len(turns)} turns, history about {hist // 1000}k tokens, {len(probes)} probes")
        for kind, tz, nb in variants:
            units = unit_cache.setdefault(kind, build_units(turns, kind))
            costs, ranks, top6 = [], [], []
            for p in probes:
                targets = {p["source"]} | {a for a in p.get("also", []) if a in ids}
                order = rank(units, tz, p["question"])
                c, r = evaluate(units, order, targets, norm(p["answer"]), nb)
                costs.append(c)
                ranks.append(r)
                if kind == "turn":
                    top6.append(sum(units[i]["cost"] for i in order[:6]))
            label = f"{kind:8s} {('porter' if 'porter' in tz else 'plain'):6s} {'+nbr' if nb else '    '}"
            found = [c for c in costs if c is not None]
            rec = " ".join(f"{sum(1 for c in found if c <= b):2d}" for b in BUDGETS)
            med = int(statistics.median(found)) if found else -1
            medr = int(statistics.median([r for r in ranks if r])) if found else -1
            extra = f"  top-6 turns cost median {int(statistics.median(top6))}" if top6 else ""
            print(f"  {label} units={len(units):5d}  answer within {'/'.join(str(b // 1000) + 'k' for b in BUDGETS)}: {rec}"
                  f"  ever {len(found):2d}/{len(probes)}  median tokens-to-answer {med:6d}  median rank {medr:3d}{extra}")
            results[(name, label)] = costs
    json.dump({f"{k[0]}|{k[1]}": v for k, v in results.items()}, open(os.path.join(os.path.dirname(__file__), "results.json"), "w"))


if __name__ == "__main__":
    main()
