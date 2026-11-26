---
type: result
status: measured
area: selection
needs: none
measured-by: us
---
On 2026-11-15 three Claude Code sessions were imported and 30 probes per conversation were written by a sub-agent that read the turns but not the ranking code. With six ranked whole turns and all lanes the source turn was in the packet verbatim for 27, 26 and 29 of 30; full text only gave 26, 26 and 28, and ten ranked slots 28, 28 and 29. The workspaces were not indexed, and full text alone finds 87 to 93 percent, so most questions share rare words with their answer.

## Evidence
- `docs/ROADMAP.md`: "the source turn was in the packet verbatim for 27, 26 and 29 of 30"
- `docs/ARCHITECTURE.md`: "| one BookWriter session | 140 | about 80k tokens | 26 of 30 | 27 of 30 | 28 of 30 |"
- `docs/ARCHITECTURE.md`: "A model wrote the questions after reading the source, and full text alone finds 87 to 93 percent, so most questions share rare words with their answer"

## Relations
- supports:: [[Recovery probes measure selection without a model]]
- supports:: [[Fused ranking over exact full-text and value lanes]]
- supports:: [[Full text alone matched all lanes on real probes]]
- supports:: [[Held-out probes written from memory without the transcript]]
