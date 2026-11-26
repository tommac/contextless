---
type: idea
status: built
area: plans
needs: none
---
A role maps to a list of provider profiles in order. A quota or auth failure cools the failing provider down, for the retry-after time when one is given and otherwise on a doubling ladder, and routing skips a provider while it is in cooldown. The call moves to the next provider in the list. The cooldown ladder came from BookWriter, and keeping cooldowns as data came from MultiAIWorkspace.

## Evidence
- `docs/ARCHITECTURE.md`: "A provider in cooldown (quota, rate limit) is skipped."
- `docs/ARCHITECTURE.md`: "The role (`responder` by default) maps to a list of profiles in order. A provider in cooldown (quota, rate limit) is skipped."
- `docs/ARCHITECTURE.md`: "When a run fails on one provider, the next attempt's brief lists what failed, by profile and exit class."

## Relations
- builds-on:: [[Roles map to ordered provider profiles kept in config]]
- builds-on:: [[Provider CLIs run as stateless workers]]
- source:: [[BookWriter]]
- source:: [[MultiAIWorkspace]]
