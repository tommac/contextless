---
type: result
status: preprint
area: compression
needs: none
measured-by: others
---
As an agent-history baseline in the ACON paper, with gpt-4.1 as the agent, LLMLingua scored 39.3 on AppWorld against 56.0 with no compression, though it held up on 8-objective question answering. The measurement was made by the authors of a competing method. The same paper shows that compression lowers peak tokens while steps can rise, from 16.14 to 22.82 on AppWorld with its own method.

## Evidence
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "As an agent-history baseline in ACON, LLMLingua scored 39.3 on AppWorld against 56.0 uncompressed, though it held up on 8-objective QA (EM 0.363 against 0.366)"
- `research_notes/Cutting edge context memory research/compression_and_caching.md`: "Peak tokens fall, but steps can rise: AppWorld went from 16.14 steps uncompressed to 22.82 with ACON UTCO history compression"

## Relations
- supports:: [[Lossy compression loses constraints and exact text]]
- supports:: [[Turns are stored verbatim and never rewritten]]
- supports:: [[Measure the bill not the tokens]]
- source:: [[LLMLingua]]
