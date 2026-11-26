---
type: source
status: external
area: plans
needs: none
---
MultiAIWorkspace, called MAW for short in the project's documents, is where the multi-provider side of ContextlessAI comes from. ContextlessAI took its per-provider lanes, with cooldowns kept as data and interleaving across providers, so that one request can be split into units that run in parallel. The repository lane runs a coder unit like a MultiAIWorkspace worker reduced to one unit, and cx stats follows its proposal for per-agent statistics from events already kept. Two of its mechanisms were combined with a BookWriter counterpart: machine-wide lanes with BookWriter's routing table and cooldown ladder, and maker and checker with BookWriter's judge independence.

## Evidence
- `docs/ARCHITECTURE.md`: "It also takes MultiAIWorkspace's idea of per-provider lanes, so one request can be split into units that run in parallel"
- `docs/ARCHITECTURE.md`: "A coder unit naming one runs like a MultiAIWorkspace worker, reduced to one unit"
- `docs/ARCHITECTURE.md`: "BookWriter judge independence, MAW maker/checker"
- `docs/ROADMAP.md`: "MAW proposal 2.5: per-agent statistics from events already kept"

## Relations
- alternative-to:: [[BookWriter]]
