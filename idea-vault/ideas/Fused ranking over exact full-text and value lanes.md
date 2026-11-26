---
type: idea
status: measured-kept
area: selection
needs: none
---
What goes into a packet verbatim is ranked by code, not by a model. Three ranked lanes are fused: exact phrase, full text and a value lane, and the winner of each lane is always kept. The fusion is LANTERN's RRF. On the planted-fact probes it put 13 of 14 source turns in the packet verbatim, up from 11, and a search for the question found 12 of 14.

## Evidence
- `docs/ROADMAP.md`: "Exact, full-text and value lanes, each lane's winner kept: 13 of 14 verbatim, and 12 of 14 found by a search for the question."
- `docs/ARCHITECTURE.md`: "Ranked lanes over the verbatim text are fused with Reciprocal Rank Fusion (k = 60, LANTERN's choice), and each lane's top hit always makes the cut:"
- `docs/ARCHITECTURE.md`: "Among whole turns this is a full lane. Among atoms only its two best hits are fused"
- `docs/ARCHITECTURE.md`: "Fused lanes with RRF, the value lane, recovery probes"

## Relations
- builds-on:: [[Ranked turns outlive the bulk of their thread]]
- builds-on:: [[Recovery probes measure selection without a model]]
- builds-on:: [[Knowledge graphs lose to BM25 as a default store]]
- alternative-to:: [[Hybrid lexical and embedding retrieval beats either alone]]
- alternative-to:: [[Full text alone matched all lanes on real probes]]
- source:: [[LANTERN]]
- tested-by:: [[Fused lanes with a value lane put 13 of 14 planted facts verbatim]]
