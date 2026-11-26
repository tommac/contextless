---
type: result
status: measured
area: cost
needs: none
measured-by: us
---
If every call were served its shared prefix at a tenth of the price, the billed-equivalent input per message on the three real conversations would fall from 14,381 to 10,267, from 18,741 to 11,632 and from 25,483 to 18,312. The shared prefix is measured by replay with no model call; the price is an assumption that every prefix is a hit. So this is a ceiling, not a saving anyone has been billed.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "billed-equivalent input per message would fall from 14,381 to 10,267, 18,741 to 11,632 and 25,483 to 18,312"
- `docs/CONTEXT-COMPILER.md`: "Those savings are a ceiling."

## Relations
- supports:: [[Prompt cache savings are a ceiling until a live test]]
- supports:: [[Catalog spine and epoch window make the prefix repeat]]
- supports:: [[Measure the bill not the tokens]]
