---
type: idea
status: built
area: review
needs: none
---
An answer assembled from plan units, or every answer when review is set to always, goes to a reviewer on a provider that wrote none of it, with the same packet plus the answer. Each problem the reviewer reports must quote the answer exactly or name something missing, and any other is discarded. On a revise verdict the writer's role gets one revision run with the surviving problems as a supplied record. It is one round: a check, not a loop. The rule comes from BookWriter's judge independence and MultiAIWorkspace's maker and checker.

## Evidence
- `docs/ARCHITECTURE.md`: "A reviewer on a provider that wrote none of the answer gets the same packet plus the answer. Problems must quote the answer verbatim or name something missing; others are discarded."
- `docs/ARCHITECTURE.md`: "One round: a check, not a loop."
- `docs/ARCHITECTURE.md`: "Problems must quote the answer verbatim or name something missing; others are discarded."
- `docs/ARCHITECTURE.md`: "BookWriter judge independence, MAW maker/checker"

## Relations
- builds-on:: [[LLM reviewers are lenient or noisy]]
- builds-on:: [[Units return typed records with checked claims]]
- builds-on:: [[Roles map to ordered provider profiles kept in config]]
- analog-of:: [[Maker-checker separation]]
- source:: [[BookWriter]]
- source:: [[MultiAIWorkspace]]
