---
type: idea
status: proposed
area: selection
needs: none
---
Fill the ranked part of a packet by giving slots to named aspects of the question instead of taking candidates in score order: PM-2 seats over aspects that code can define, namely sessions with matches, the distinct values recorded for the queried key, and rare terms that recur in the first-pass results. Using feedback terms as aspects for allocation, and not as new query terms, would keep the coverage benefit of feedback without the drift. It would need complete-evidence coverage and mean recall reported at a fixed token budget against score order, a fixed session floor and MMR. The research note found no application of xQuAD or PM-2 to context selection, and prior art stops at EdgeMem's fixed session floor.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Explicit-aspect slot allocation in the pack. PM-2 seats over deterministic aspects (sessions with matches, distinct values of the queried key, recurring rare terms from the first pass)."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Report complete-evidence coverage and mean recall at a fixed token budget against score order, a fixed session floor and MMR. Prior art stops at EdgeMem's fixed floor."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "That keeps the coverage benefit of feedback without the drift measured in the first section, and no source was found that tests it."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Not found (EdgeMem has a fixed session floor)"

## Relations
- builds-on:: [[Ranked atoms are sent as excerpts under a token budget]]
- builds-on:: [[The value lane finds stated values near the question's words]]
- builds-on:: [[Whole-conversation questions are the weak spot]]
- alternative-to:: [[Budgeted selection needs a monotone submodular objective]]
- analog-of:: [[Result diversification]]
- tested-by:: [[PM-2 scored 0.4546 against xQuAD 0.4074 and MMR 0.3349]]
- source:: [[PM-2]]
