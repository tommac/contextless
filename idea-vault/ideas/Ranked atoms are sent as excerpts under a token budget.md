---
type: idea
status: built
area: packets
needs: none
---
The ranked section of a packet is filled with the best atoms until their text and headers reach relevant_tokens, 8,000 by default, in place of six ranked whole turns. Adjacent atoms of one turn go out as one excerpt under a header that names the turn, its paragraph numbers, its recall id and the message it answers. The rest of the turn stays in the catalog and can be recalled by id, so nothing is lost by sending a part.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "ranked atoms sent as excerpts under a token budget (`relevant_tokens`, default 8,000) in place of six ranked turns"
- `docs/ARCHITECTURE.md`: "Adjacent atoms of a turn are sent as one excerpt, headed by the turn, its paragraph numbers, its recall id and the opening of the message it answers"

## Relations
- builds-on:: [[Atoms are paragraph-sized addresses into a turn]]
- builds-on:: [[Packets are sized before the model starts]]
- builds-on:: [[Recall by id or by search replaces memory]]
- alternative-to:: [[A record goes in whole or is listed by id]]
- tested-by:: [[Atoms at 8k found 107 of 120 against 103 for whole turns]]
- tested-by:: [[Atom packets were 6 to 44 percent smaller at the 8k budget]]
