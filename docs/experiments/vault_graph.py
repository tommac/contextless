#!/usr/bin/env python3
"""Check the idea vault and analyse its link graph.

usage: vault_graph.py <repo> [--write]

The vault (<repo>/idea-vault) holds one note per idea, result, source or
outside-field concept, with typed relations and quotes from the files each
note was drawn from. This script trusts none of it:

  1. Check. Every quote must occur in the file it names; every relation must
     use a known key and point at a note that exists; front matter values
     must be valid; note names must be unique across folders.
  2. Analyse. Ideas nobody has tested; ideas a measured result contradicts;
     and "missing links": pairs of ideas in different areas that are not
     linked but share many neighbours, scored with the Adamic-Adar index
     (shared neighbours, each weighted by 1 / log of its degree). This is
     Swanson's ABC model of literature-based discovery (A relates to B, B
     relates to C, nobody has related A to C) applied to the project's own
     ideas: candidates for combinations worth thinking about, not findings.

With --write it writes the generated notes 00 to 03 into the vault.
No model, no network.
"""
import math
import os
import re
import sys
from collections import Counter, defaultdict

TYPES = {"idea": "ideas", "result": "results", "source": "sources", "field": "fields"}
STATUS = {
    "idea": {"built", "designed", "proposed", "measured-kept", "measured-dropped", "rejected", "external"},
    "result": {"measured", "preprint", "vendor-reported", "disputed", "claimed"},
    "source": {"external"},
    "field": {"external"},
}
AREAS = {"storage", "packets", "selection", "recall", "index", "caching", "plans", "review", "evaluation",
         "structure", "compression", "cost"}
RELS = {"supports", "contradicts", "builds-on", "alternative-to", "tested-by", "source", "part-of", "analog-of"}
BAD = set('/\\:*?"<>|#^[]')
GENERATED = ("00 ", "01 ", "02 ", "03 ")

EV = re.compile(r'^- `([^`]+)`: "(.*)"\s*$')
REL = re.compile(r"^- ([a-z-]+):: \[\[([^\]|]+)(?:\|[^\]]*)?\]\]\s*$")


def norm(s):
    return " ".join(s.split())


class Note:
    def __init__(self, path, folder, name):
        self.path, self.folder, self.name = path, folder, name
        self.meta, self.evidence, self.rels, self.problems = {}, [], [], []
        self.text = ""


def load(vault):
    notes = []
    for folder in sorted(TYPES.values()):
        d = os.path.join(vault, folder)
        if not os.path.isdir(d):
            continue
        for f in sorted(os.listdir(d)):
            if not f.endswith(".md"):
                continue
            n = Note(os.path.join(d, f), folder, f[:-3])
            with open(n.path, encoding="utf-8") as source:
                raw = source.read()
            body = raw
            if raw.startswith("---\n") and "\n---\n" in raw[4:]:
                head, body = raw[4:].split("\n---\n", 1)
                for line in head.splitlines():
                    if ":" in line:
                        k, v = line.split(":", 1)
                        n.meta[k.strip()] = v.split("#")[0].strip() if k.strip() != "url" else v.strip()
            else:
                n.problems.append("no front matter")
            n.text = body
            section = ""
            for line in body.splitlines():
                if line.startswith("## "):
                    section = line[3:].strip().lower()
                    continue
                if section == "evidence" and line.startswith("- "):
                    m = EV.match(line)
                    if m:
                        n.evidence.append((m.group(1), m.group(2)))
                    else:
                        n.problems.append("evidence line not in the form - `file`: \"quote\"")
                if section == "relations" and line.startswith("- "):
                    m = REL.match(line)
                    if m:
                        n.rels.append((m.group(1), m.group(2).strip()))
                    else:
                        n.problems.append("relation line not in the form - key:: [[Note]]: " + line[:60])
            notes.append(n)
    return notes


def check(repo, notes):
    by_name = defaultdict(list)
    for n in notes:
        by_name[n.name].append(n)
    files = {}
    quotes = found = 0
    for n in notes:
        t = n.meta.get("type")
        if TYPES.get(t) != n.folder:
            n.problems.append(f"type {t!r} does not match folder {n.folder}")
        elif n.meta.get("status") not in STATUS[t]:
            n.problems.append(f"status {n.meta.get('status')!r} is not valid for a {t}")
        if n.meta.get("area") not in AREAS:
            n.problems.append(f"area {n.meta.get('area')!r}")
        if set(n.name) & BAD or len(n.name) > 80:
            n.problems.append("note name has a forbidden character or is over 80 characters")
        if len(by_name[n.name]) > 1:
            n.problems.append("another note has this name")
        if not n.evidence:
            n.problems.append("no evidence line")
        for path, quote in n.evidence:
            quotes += 1
            full = os.path.join(repo, path)
            if full not in files:
                files[full] = norm(open(full, encoding="utf-8").read()) if os.path.isfile(full) else None
            if files[full] is None:
                n.problems.append(f"evidence file not found: {path}")
            elif norm(quote) not in files[full]:
                n.problems.append(f"quote not in {path}: {quote[:50]!r}")
            else:
                found += 1
        outside = re.sub(r'"[^"\n]*"', "", n.text)
        if "\u2014" in outside or "\u2014" in n.name:
            n.problems.append("em-dash outside a quote")
        for key, target in n.rels:
            if key not in RELS:
                n.problems.append(f"unknown relation {key!r}")
            if target not in by_name:
                n.problems.append(f"link to a note that does not exist: {target}")
    return by_name, quotes, found


def graph(notes, by_name):
    nbr = defaultdict(set)
    incoming = defaultdict(list)
    for n in notes:
        for key, target in n.rels:
            if target in by_name and key in RELS and target != n.name:
                nbr[n.name].add(target)
                nbr[target].add(n.name)
                incoming[target].append((key, n.name))
    return nbr, incoming


def analyse(notes, by_name):
    nbr, incoming = graph(notes, by_name)
    meta = {n.name: n.meta for n in notes}
    rels = {n.name: n.rels for n in notes}
    ideas = [n for n in notes if n.meta.get("type") == "idea"]

    def tested(n):
        if any(k == "tested-by" for k, _ in n.rels):
            return True
        return any(k in ("supports", "contradicts") and meta.get(src, {}).get("type") == "result"
                   for k, src in incoming[n.name])

    untested = [n for n in ideas if n.meta.get("status") in ("built", "designed", "proposed") and not tested(n)]
    against = []
    for n in ideas:
        cons = [src for k, src in incoming[n.name]
                if k == "contradicts" and meta.get(src, {}).get("type") == "result"]
        if cons and n.meta.get("status") in ("built", "designed", "proposed", "measured-kept"):
            against.append((n, cons))

    # missing links: unlinked pairs of ideas in different areas, by Adamic-Adar
    names = [n.name for n in ideas]
    pairs = []
    for i, a in enumerate(names):
        for c in names[i + 1:]:
            if c in nbr[a] or meta[a].get("area") == meta[c].get("area"):
                continue
            shared = nbr[a] & nbr[c]
            if len(shared) < 2:
                continue
            score = sum(1 / math.log(1 + len(nbr[b])) for b in shared)
            # worth a look only if something is still to be done on one side
            open_side = {meta[a].get("status"), meta[c].get("status")} & {"designed", "proposed", "external"}
            pairs.append((score, a, c, sorted(shared, key=lambda b: len(nbr[b])), bool(open_side)))
    pairs.sort(key=lambda p: (-p[0], p[1], p[2]))
    hubs = sorted(((len(v), k) for k, v in nbr.items()), reverse=True)
    lonely = [n for n in notes if n.meta.get("type") in ("field", "source") and
              not any(meta.get(x, {}).get("type") == "idea" and meta[x].get("status") != "external" for x in nbr[n.name])]
    return dict(nbr=nbr, untested=untested, against=against, pairs=pairs, hubs=hubs, lonely=lonely, rels=rels)


def link(name):
    return f"[[{name}]]"


def write(vault, notes, a):
    head = "---\ntype: generated\n---\nGenerated by `docs/experiments/vault_graph.py` from the notes and their relations. Do not edit: run the script again.\n\n"
    by_area = defaultdict(list)
    for n in notes:
        by_area[n.meta.get("area", "?")].append(n)
    out = [head + "# Map of the vault\n"]
    counts = Counter((n.meta.get("type"), n.meta.get("status")) for n in notes)
    out.append("Notes: " + ", ".join(f"{k[0]} {k[1]} {v}" for k, v in sorted(counts.items(), key=lambda kv: (str(kv[0][0]), str(kv[0][1])))) + ".\n")
    out.append("Start with [[01 Untested ideas]], [[02 Ideas the evidence is against]] and [[03 Missing links]].\n")
    for area in sorted(by_area):
        out.append(f"\n## {area}\n")
        for t in ("idea", "result", "field", "source"):
            ns = sorted((n for n in by_area[area] if n.meta.get("type") == t), key=lambda n: n.name)
            if ns:
                out.append(f"\n{t.capitalize()}s:\n")
                out += [f"- {link(n.name)} ({n.meta.get('status')})" for n in ns]
    open(os.path.join(vault, "00 Map of the vault.md"), "w", encoding="utf-8").write("\n".join(out) + "\n")

    out = [head + "# Untested ideas\n\nIdeas that are built, designed or proposed and that no result note tests.\n"]
    for n in sorted(a["untested"], key=lambda n: (n.meta.get("status"), n.meta.get("area"), n.name)):
        out.append(f"- {link(n.name)} ({n.meta.get('status')}, {n.meta.get('area')}, needs {n.meta.get('needs')})")
    open(os.path.join(vault, "01 Untested ideas.md"), "w", encoding="utf-8").write("\n".join(out) + "\n")

    out = [head + "# Ideas the evidence is against\n\nIdeas still built, designed or proposed that a result note contradicts.\n"]
    for n, cons in sorted(a["against"], key=lambda x: x[0].name):
        out.append(f"- {link(n.name)} ({n.meta.get('status')}): contradicted by " + ", ".join(link(c) for c in cons))
    open(os.path.join(vault, "02 Ideas the evidence is against.md"), "w", encoding="utf-8").write("\n".join(out) + "\n")

    out = [head + "# Missing links\n\nPairs of ideas in different areas that no note links, though they share neighbours. "
           "The score is the Adamic-Adar index: each shared neighbour counts 1 / log(1 + its number of links), so a rare shared "
           "neighbour says more than a hub. These are candidates for combinations to think about, not findings.\n"]
    for score, x, c, shared, open_side in a["pairs"][:60]:
        mark = "" if open_side else " (both already built or measured)"
        out.append(f"- {score:.2f}: {link(x)} and {link(c)}{mark}, through " + ", ".join(link(s) for s in shared[:5]))
    open(os.path.join(vault, "03 Missing links.md"), "w", encoding="utf-8").write("\n".join(out) + "\n")


def main():
    repo = sys.argv[1]
    vault = os.path.join(repo, "idea-vault")
    notes = load(vault)
    by_name, quotes, found = check(repo, notes)
    bad = [n for n in notes if n.problems]
    links = sum(len(n.rels) for n in notes)
    print(f"{len(notes)} notes, {links} relations; {found} of {quotes} quotes found in the file they cite; {len(bad)} notes with problems")
    kinds = Counter()
    for n in bad:
        for p in n.problems:
            kinds[p.split(":")[0][:45]] += 1
    for k, v in kinds.most_common(12):
        print(f"  {v:4} {k}")
    if "--problems" in sys.argv:
        for n in bad:
            for p in n.problems:
                print(f"{n.folder}/{n.name}: {p}")
    a = analyse(notes, by_name)
    print(f"untested ideas {len(a['untested'])}; ideas a result contradicts {len(a['against'])}; candidate missing links {len(a['pairs'])}; "
          f"outside notes not linked to a project idea {len(a['lonely'])}")
    print("most linked:", "; ".join(f"{k} ({d})" for d, k in a["hubs"][:8]))
    if "--write" in sys.argv:
        write(vault, notes, a)
        print("wrote 00 to 03 in", vault)
    if "--pairs" in sys.argv:
        for score, x, c, shared, open_side in a["pairs"][:40]:
            print(f"{score:.2f} | {x} <-> {c} | via {', '.join(shared[:4])}")


if __name__ == "__main__":
    main()
