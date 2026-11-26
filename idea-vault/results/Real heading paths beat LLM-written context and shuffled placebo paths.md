---
type: result
status: measured
area: index
needs: none
measured-by: others
---
In a placebo-controlled test on 951 Wikipedia and 4,303 QASPER queries with dense retrieval, real heading paths beat shuffled placebo paths by 0.010 and 0.016 coverage-nDCG@10, so the content of the path and not the added tokens drives the gain. The same test puts figures on the comparison with model-written context: structure-aligned chunks with real heading paths beat LLM-contextualized fixed windows by 0.022 and 0.012, and at 512 tokens LLM-written prefixes fell below no prefix. The effects are small, the test is a preprint by a group independent of the methods, and it bears only indirectly on model-written atom titles.

## Evidence
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Structure-aligned chunks with real heading paths beat LLM-contextualized fixed windows, +0.022 and +0.012."
- `research_notes/Cutting edge context memory research/selection_and_retrieval.md`: "Real heading paths beat shuffled placebo paths, +0.010 and +0.016, so content rather than added tokens drives the gain."

## Relations
- supports:: [[Deterministic keys on spans with verbatim output]]
- contradicts:: [[Atom titles as APIs]]
- source:: [[Kuehlkamp et al 2026]]
