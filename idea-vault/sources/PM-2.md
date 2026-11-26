---
type: source
status: external
area: selection
needs: none
year: 2012
url: https://ciir-publications.cs.umass.edu/getpdf.php?id=1050
---
PM-2 (Dang and Croft, SIGIR 2012) is a search result diversification method that treats result slots as seats in an election. For each slot it picks the aspect whose weight is largest relative to the seats it already holds, then the document that best serves that aspect while also serving others, and credits each aspect with a share of the seat. On the TREC Web track it beat xQuAD and MMR, and a term-level follow-up reports that both xQuAD and PM-2 can diversify with terms as aspects. The research note proposes PM-2 seats over aspects that code can define for a conversation.

## Evidence
- `research_notes/Cutting edge context memory research/database_patterns.md`: "treats result slots as seats in an election"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "then picks the document that best serves that aspect while also serving others, and credits each aspect with a share of the seat"
- `research_notes/Cutting edge context memory research/database_patterns.md`: "A term-level follow-up by Dang and Croft reports that xQuAD and PM-2 can diversify with terms as aspects."

## Relations
- part-of:: [[Result diversification]]
- alternative-to:: [[Lin and Bilmes 2011]]
- tested-by:: [[PM-2 scored 0.4546 against xQuAD 0.4074 and MMR 0.3349]]
