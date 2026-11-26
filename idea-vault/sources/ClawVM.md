---
type: source
status: external
area: packets
needs: none
year: 2026
url: https://arxiv.org/html/2604.10352
---
ClawVM (EuroMLSys 2026) treats an agent's context as virtual memory. Typed pages have stable ids, provenance and minimum-fidelity floors, and each is kept at several resolutions: full, compressed, structured and pointer. The prompt is assembled by code, pinned minimums first and then upgrades by utility per token. It is the closest published template for this project's deterministic packet building. Its evidence is synthetic replays at very small budgets, with no test of answer quality by a live model.

## Evidence
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "Prompt assembly installs pinned minimums, then greedily upgrades by utility per token with deterministic tie-breaks"
- `research_notes/Cutting edge context memory research/memory_data_structures.md`: "there is no end-to-end task-quality evaluation with a live model"

## Relations
- supports:: [[Each call gets a fresh packet built by code]]
- analog-of:: [[Recall as paging]]
- analog-of:: [[Levels of detail like map tiles]]
- supports:: [[Determinism is the instrument]]
