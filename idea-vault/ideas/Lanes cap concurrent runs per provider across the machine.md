---
type: idea
status: built
area: plans
needs: none
---
A lane is a cap on how many runs one provider may have going at once. The cap holds across every cx process on the machine, not only inside one conversation. Plan units run across lanes, which is how one request is split into units that run in parallel on several providers. Lanes came from MultiAIWorkspace, together with cooldowns kept as data and interleaving across providers.

## Evidence
- `docs/ARCHITECTURE.md`: "Lanes cap concurrent runs per provider."
- `docs/ARCHITECTURE.md`: "It also takes MultiAIWorkspace's idea of per-provider lanes, so one request can be split into units that run in parallel"
- `docs/ARCHITECTURE.md`: "Per-provider lanes, cooldowns as data, interleaving across providers"

## Relations
- builds-on:: [[Provider CLIs run as stateless workers]]
- builds-on:: [[Roles map to ordered provider profiles kept in config]]
- source:: [[MultiAIWorkspace]]
