---
type: idea
status: proposed
area: evaluation
needs: model-calls
---
The proposal is to test whether an answer really depends on a retrieved atom by changing the packet: remove the atom, supply only that atom, or change the value that carries the answer and see whether the answer follows. A deterministic compiler over immutable text makes these tests clean, because two arms differ only by a known set of atoms. The closest prior art is ContextCite, the paired replay of Compact-Bench, MemTrace and UTILMEM. The report judges a validated value-swap protocol for closed APIs, applied across memory systems, likely new as a protocol and moderately uncertain, because two of the nearest papers were seen only as snippets.

## Evidence
- `reports/Cutting edge context memory research.md`: "**Atom-level causal evaluation** with removal, sufficiency and value swap on a replayable log"
- `reports/Cutting edge context memory research.md`: "A validated value-swap protocol for closed APIs, applied across memory systems"
- `reports/Cutting edge context memory research.md`: "**Likely new as a protocol, moderately uncertain.** The last two were seen only as snippets"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Removing or altering a passage tests whether it was actually used, which correctness and citation checks cannot show."

## Relations
- builds-on:: [[Determinism is the instrument]]
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[Three layers reached used and correct]]
- source:: [[ContextCite]]
- source:: [[Compact-Bench]]
- source:: [[MemTrace]]
