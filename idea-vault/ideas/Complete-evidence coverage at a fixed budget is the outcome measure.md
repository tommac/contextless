---
type: idea
status: proposed
area: evaluation
needs: none
---
Score a selector by the share of questions for which every piece of gold evidence is in the packet at a fixed token budget, not by recall-any or mean recall. Most memory papers report recall-any or mean recall; complete-evidence coverage was found in one paper only, where it trailed mean recall by 8.6 points for a BM25-only pipeline. No source reports it for a purely lexical pipeline at paragraph-span granularity, which is this project's setting. It is the measure on which slot allocation, a change ledger and cover density would be judged.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The right outcome measure is complete-evidence coverage at a fixed token budget."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Most memory papers report recall-any or mean recall; complete-evidence coverage was found in one paper only."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No source reports complete-evidence coverage for a purely lexical pipeline at paragraph-span granularity"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The measurable question is whether a deterministic view lookup raises complete-evidence coverage at a fixed token budget over the same lanes without the view."

## Relations
- builds-on:: [[Recovery probes measure selection without a model]]
- builds-on:: [[Budget-matched baselines]]
- builds-on:: [[Completeness by accounting against code-found candidates]]
- builds-on:: [[Whole-conversation questions are the weak spot]]
- part-of:: [[Three layers reached used and correct]]
- tested-by:: [[With BM25 alone complete coverage was 73.19 against 81.83 mean recall]]
- source:: [[Threader]]
