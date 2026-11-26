---
type: field
status: external
area: structure
needs: none
---
Database provenance annotates each stored row with a variable and carries polynomials of those variables through queries, so a derived row records how it was derived. In LLM retrieval the measured data are thin: one position paper with preliminary numbers on provenance tokens, one lineage per evidence demand in LineageRAG, lineage used to stop double counting in AutoViewMem, and support sets in Stashbird. For this project the research note proposes recording, for each packed span, which lanes and query terms derived it, and scoring by the number of independent derivations. No measurement of provenance-based ranking or packing was found.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "annotate tuples with variables and propagate polynomials through queries; bag semantics, why-provenance, probabilistic and incomplete databases are instances obtained by evaluating the polynomial in different semirings"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "record for each packed span which lanes and query terms derived it, then score under a counting semiring (number of independent derivations). This is a proposal; nothing measured was found."
- `research_notes/Cutting edge context memory research/database_patterns.md`: "AutoViewMem maps retrieved items to source span ids before scoring so that several view-specific items from one span do not inflate coverage"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "average gains over leading GraphRAG baselines on HotpotQA, 2WikiMultiHopQA and MuSiQue were +3.51 R@5, +5.96 EM, +5.22 F1"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "No measurement of provenance-based ranking or packing was found."

## Relations
- analog-of:: [[The indexer writes cards facts and decisions with checked quotes]]
- analog-of:: [[Quotes are pinned to the paragraphs that hold them]]
- analog-of:: [[Proof-carrying memory]]
- analog-of:: [[Views with lineage instead of summaries]]
- analog-of:: [[The assembler accounts for every claim]]
- tested-by:: [[A provenance token cut acceptance of fabricated facts by about 60 percent]]
- tested-by:: [[Support sets gave 20 of 20 correct final states against 8 of 20]]
