---
type: idea
status: proposed
area: index
needs: none
---
Retrieve small and read large is best done here as multi-key indexing: keep the paragraph span as the charged unit, add keys that code can derive (speaker, date, file path, heading path, tool name, extracted identifiers) to the full-text index only, and always return the verbatim span. It is the no-model analogue of fact-augmented keys and contextual BM25, and the evidence supports extra derived keys with verbatim output more firmly than it supports hierarchy. Whether deterministic keys deliver comparable gains is unmeasured.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "keep paragraph spans as the charged unit, and add deterministic keys to BM25 only (speaker, date, file path, heading path, tool name, extracted identifiers)"
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Cheap structural context (real headings, speaker, date) has better controlled evidence than LLM-written context."

## Relations
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[The atom index is stemmed and searches card text]]
- alternative-to:: [[Atom titles as APIs]]
- source:: [[Kuehlkamp et al 2026]]
- alternative-to:: [[Idea notes as verified keys over verbatim paragraphs]]
- tested-by:: [[Real heading paths beat LLM-written context and shuffled placebo paths]]
- tested-by:: [[Fact-augmented keys raised recall 9.4 percent on LongMemEval]]
