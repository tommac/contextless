---
type: field
status: external
area: review
needs: none
---
Maker-checker separation is the rule that whoever made a piece of work is not the one who checks it. MultiAIWorkspace calls it maker and checker, and BookWriter calls it judge independence. ContextlessAI applies it across providers: the reviewer can be any provider except the writer's, so an answer is never checked by the provider that wrote it.

## Evidence
- `docs/ARCHITECTURE.md`: "BookWriter judge independence, MAW maker/checker"
- `docs/ARCHITECTURE.md`: "Reviewer never on the writer's provider"
- `docs/ROADMAP.md`: "A reviewer on a provider that wrote none of the answer checks it against the packet; problems must quote the answer verbatim."

## Relations
- source:: [[MultiAIWorkspace]]
- source:: [[BookWriter]]
