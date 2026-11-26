---
type: idea
status: designed
area: evaluation
needs: model-calls
---
Third on the agenda, and the first test in which a model answers from a packet. The hypothesis is that the current 8k packet is within 8 points of whole history at under 40 percent of its billed input, and beats plain BM25 atoms by enough to justify the extra lanes. With 60 pilot questions it can detect only about 14 points, so it serves to estimate discordance and to decide go or no-go. It is falsified if the fused packet is more than 8 points worse than full context, or if BM25 ties it at equal or lower cost.

## Evidence
- `reports/Cutting edge context memory research.md`: "The current 8k packet is within 8 points of whole history at under 40% of its billed input, and beats plain BM25 atoms enough to justify the extra lanes"
- `reports/Cutting edge context memory research.md`: "At 15% discordance this detects only 14 points, so it is for discordance estimates and go or no-go."
- `reports/Cutting edge context memory research.md`: "Fused is worse than full context by more than 8 points, or BM25 ties fused at equal or lower cost, in which case the extra lanes do not earn their place in packets"

## Relations
- builds-on:: [[R1 Repair the instrument and record the offline floor]]
- builds-on:: [[No model has yet answered from a packet]]
- builds-on:: [[Fused ranking over exact full-text and value lanes]]
- builds-on:: [[Whole history finds every source at ten times the tokens]]
- builds-on:: [[Full text alone matched all lanes on real probes]]
- builds-on:: [[Non-inferiority needs a pre-set margin]]
