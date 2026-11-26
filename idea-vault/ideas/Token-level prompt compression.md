---
type: idea
status: rejected
area: compression
needs: none
---
Shortening a prompt by deleting tokens from it is rejected for this design. LLMLingua scored well below the uncompressed prompt on AppWorld, and in a study of coding runs the share of patches that applied fell from 27 of 40 to 15 of 40. It also breaks exact quote checks, which this project relies on to verify what a model cites.

## Evidence
- `reports/Cutting edge context memory research.md`: "LLMLingua 39.3 against 56.0 uncompressed on AppWorld"
- `reports/Cutting edge context memory research.md`: "patch applicability fell from 27/40 to 15/40"
- `reports/Cutting edge context memory research.md`: "Reject. It also breaks exact quote checks"

## Relations
- part-of:: [[Lossy compression loses constraints and exact text]]
- contradicts:: [[Quotes are pinned to the paragraphs that hold them]]
- contradicts:: [[A record goes in whole or is listed by id]]
- tested-by:: [[LLMLingua scored 39.3 on AppWorld against 56.0 uncompressed]]
- source:: [[LLMLingua]]
- source:: [[Token Reduction Is Not Cost Reduction]]
- tested-by:: [[Compression cut cost 59 percent and patch applicability from 27 to 15 of 40]]
