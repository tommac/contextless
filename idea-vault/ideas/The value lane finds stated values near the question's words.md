---
type: idea
status: measured-kept
area: selection
needs: none
---
The value lane ranks text by whether a sentence states a value, such as a number, date, name or identifier, next to the words of the question. It was added for the config-value blind spot that LANTERN's failure analysis named. It is fused with the exact and full-text lanes, and with it 13 of 14 planted facts were in the packet verbatim. For paragraph-sized atoms only its best hits are used.

## Evidence
- `docs/ARCHITECTURE.md`: "**value**: the turn states a value (a number, date, time, amount, code identifier, path or name) within four words of a query word in the same sentence"
- `docs/ROADMAP.md`: "LANTERN's RRF; its config-value blind spot"
- `docs/ARCHITECTURE.md`: "LANTERN (Cisco, 2026) and its failure analysis"

## Relations
- part-of:: [[Fused ranking over exact full-text and value lanes]]
- builds-on:: [[Recovery probes measure selection without a model]]
- alternative-to:: [[Full text alone matched all lanes on real probes]]
- source:: [[LANTERN]]
- tested-by:: [[Fused lanes with a value lane put 13 of 14 planted facts verbatim]]
