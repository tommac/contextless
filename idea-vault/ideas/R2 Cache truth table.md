---
type: idea
status: designed
area: caching
needs: model-calls
---
Second on the agenda, needing the owner's approval for about 300 short calls. The hypothesis is that a shared start inside one stdin block gets no cache hit at the level of the brief on Claude or GPT-5.6 and later, that per-run directories also prevent reuse, and that a separately delivered block in a stable directory is required. Zero hits in 20 pairs would bound the hit rate just under the lowest break-even point. It is falsified if single-block prompts hit above that bound, or if separate blocks in a stable directory still show no hits.

## Evidence
- `reports/Cutting edge context memory research.md`: "A shared start inside one stdin block gets no brief-level hit on Claude or GPT-5.6 and later, and per-run directories also prevent reuse."
- `reports/Cutting edge context memory research.md`: "Zero hits in 20 gives a 95% upper bound of 13.9%, just below the lowest break-even of 14%."
- `reports/Cutting edge context memory research.md`: "Single-block A+B and A+C hits on Claude or GPT-5.6+ exceed the bound, or separate blocks in a stable directory still show no hits"

## Relations
- builds-on:: [[R1 Repair the instrument and record the offline floor]]
- builds-on:: [[Prompt cache savings are a ceiling until a live test]]
- builds-on:: [[Catalog spine and epoch window make the prefix repeat]]
- builds-on:: [[The brief runs from stable to volatile]]
- builds-on:: [[One stdin brief in a per-run directory predicts cache misses]]
- source:: [[Dont Break the Cache]]
