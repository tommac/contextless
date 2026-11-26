---
type: field
status: external
area: selection
needs: none
---
Result diversification picks a set of results that together cover the different aspects of a question, instead of taking the top results by score alone. In classic search the methods that name the aspects and give each a share of the slots (xQuAD, PM-2) work, and pairwise novelty (MMR) is the weakest and sometimes below no diversification. In retrieval-augmented generation MMR-style selection yields a few points when its weight is tuned per query, and on conversational memory the only direct data are a null result for MMR, a bundled distinct-session rule and one report of complete-evidence coverage. No work was found applying xQuAD or PM-2 to context selection, or measuring time-spread sampling.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "do not rely on pairwise novelty; name the aspects and allocate slots to them. In the xQuAD study MMR fell below undiversified BM25."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No work was found applying xQuAD or PM-2 to context selection, or measuring time-spread sampling."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "The oracle weight is worth up to 18 F1 over plain retrieval; the best fixed weight differs by dataset"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "Recall of the answer-bearing sentence at a 5% keep rate: Natural Questions 46.28 to 50.43, TriviaQA 63.44 to 65.29, NarrativeQA 18.60 to 20.88."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "EdgeMem mixes the global ranking with representatives from at least 3 distinct sessions, 5 for enumerative questions."

## Relations
- alternative-to:: [[Submodular maximization under a knapsack budget]]
- analog-of:: [[Budgeted selection needs a monotone submodular objective]]
- tested-by:: [[PM-2 scored 0.4546 against xQuAD 0.4074 and MMR 0.3349]]
- tested-by:: [[MMR reranking did not help SmartSearch on conversational memory]]
- tested-by:: [[With BM25 alone complete coverage was 73.19 against 81.83 mean recall]]
- tested-by:: [[Plain MMR beat AdaGReS at a 1024 token budget]]
- source:: [[PM-2]]
