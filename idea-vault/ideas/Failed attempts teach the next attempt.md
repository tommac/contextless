---
type: idea
status: built
area: plans
needs: none
---
When a run fails on one provider, the failure is carried forward as data. The brief of the next attempt lists what failed, by profile and exit class, so the next provider is told what already went wrong. The idea comes from BookWriter's lastCritiques and MultiAIWorkspace's proposal 2.6.

## Evidence
- `docs/ROADMAP.md`: "When a run fails on one provider, the next provider's brief says what failed and why, as data."
- `docs/ARCHITECTURE.md`: "When a run fails on one provider, the next attempt's brief lists what failed, by profile and exit class."
- `docs/ROADMAP.md`: "BookWriter `lastCritiques`, MAW proposal 2.6"

## Relations
- builds-on:: [[A provider in cooldown is skipped and the call moves on]]
- builds-on:: [[Each call gets a fresh packet built by code]]
- source:: [[BookWriter]]
- source:: [[MultiAIWorkspace]]
