---
type: idea
status: proposed
area: review
needs: model-calls
---
The survey proposes measuring the reviewer the way the probes measure retrieval, without trusting it. Known errors are seeded into otherwise good assembled answers (a changed value, a dropped unit claim, a claim with no evidence) to measure catch rate, and the reviewer is run on clean answers to count the problems that survive the quote rule as false positives. A same-provider reviewer is compared on the same items, because whether a different-provider reviewer catches more is unmeasured anywhere. Thirty items give 3.3-point resolution, so small differences are noise.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Seed known errors into otherwise good assembled answers (a changed value, a dropped unit claim, a claim with no evidence) and measure catch rate"
- `reports/Parallel AI agents data structures.md`: "Whether the different-provider rule itself adds catch rate is unmeasured anywhere; keep it, and measure it."
- `reports/Parallel AI agents data structures.md`: "Thirty items give the same 3.3-point resolution the probe sets have, so treat small differences as noise."

## Relations
- builds-on:: [[Independent review on another provider]]
- builds-on:: [[LLM reviewers are lenient or noisy]]
- analog-of:: [[Recovery probes measure selection without a model]]
- part-of:: [[Maker-checker separation]]
