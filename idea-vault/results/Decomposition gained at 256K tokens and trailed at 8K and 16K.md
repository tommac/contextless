---
type: result
status: vendor-reported
area: plans
needs: none
measured-by: others
---
On the OOLONG benchmark, LCM's gain over raw Opus 4.6 was +18.5 against +8.5 for Claude Code at 256K tokens, while Claude Code was slightly ahead at 8K and 16K. Recursive Language Models are likewise slightly worse than the base model on short inputs. Gains from decomposition depend on length, and it is not free below the point where the window degrades. LCM's figures are a vendor result on one benchmark split.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "By length (gain over raw Opus): 8K +11.2 vs +13.1 and 16K +25.0 vs +26.3 (Claude Code slightly ahead); 256K +18.5 vs +8.5"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "Decomposition is not free below the point where the window degrades."
- `reports/Parallel AI agents data structures.md`: "Measured; length-dependent; LCM is a vendor result on one benchmark split"

## Relations
- supports:: [[Plans that buy nothing are answered in one call]]
- supports:: [[A short history is sent whole]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
- source:: [[LCM Lossless Context Management]]
