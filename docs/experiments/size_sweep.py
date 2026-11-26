import sys, json, statistics
sys.path.insert(0, ".")
import atoms_exp as X
P = sys.argv[1]
wss = [("bw", f"{P}/cxroot/workspaces/bw", f"{P}/probes-bw.jsonl"), ("bwall", f"{P}/cxroot/workspaces/bwall", f"{P}/probes-bw.jsonl"), ("s8", f"{P}/cxsizing/workspaces/s8", f"{P}/probes-s8.jsonl")]
orig = X.atoms_of
for mn, mx in [(80, 800), (200, 2400), (400, 4800), (800, 9600)]:
    X.atoms_of = lambda b, mn=mn, mx=mx: orig(b, mn, mx)
    line = []
    for name, ws, pf in wss:
        turns = X.load_turns(ws); ids = {t["id"] for t in turns}
        probes = [json.loads(l) for l in open(pf) if l.strip()]
        units = X.build_units(turns, "atom")
        costs = []
        for p in probes:
            t = {p["source"]} | {a for a in p.get("also", []) if a in ids}
            c, r = X.evaluate(units, X.rank(units, "porter unicode61", p["question"]), t, X.norm(p["answer"]))
            costs.append(c)
        f = [c for c in costs if c is not None]
        line.append(f"{name}: units {len(units):4d}, at 2k/4k/8k {sum(c<=2000 for c in f):2d}/{sum(c<=4000 for c in f):2d}/{sum(c<=8000 for c in f):2d}, median {int(statistics.median(f)):5d}")
    print(f"atoms {mn//4*1:>4}-{mx//4} tokens (about):  " + " | ".join(line))
