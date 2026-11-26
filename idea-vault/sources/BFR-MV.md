---
type: source
status: external
area: evaluation
needs: none
year: 2026
url: https://arxiv.org/html/2610.09348
---
BFR-MV is an October 2026 preprint that the notes call the strongest evaluation protocol they found. It tests paired outcomes exactly, bootstraps at the level of questions, clusters its intervals and corrects for multiple tests, and it kept LongMemEval-S unscored until its protocol was frozen. It also states openly that all ten LoCoMo conversations were visible during development, which the notes hold up as the model disclosure.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "BFR uses exact two-sided McNemar for judged accuracy, a 10,000-sample question-level paired bootstrap for F1, clustered 95% intervals for ablations, Holm correction for type-level tests, and conversation-level five-fold analysis"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "The BFR paper states that all 10 LoCoMo conversations were visible during development, so LoCoMo only shows stability"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "BFR argues top-k relevance does not ensure the retrieved set jointly suffices, and measures set-level sufficiency"

## Relations
- supports:: [[Paired tests with conversation-level resampling]]
- supports:: [[Locked holdout with a hashed analysis plan]]
- supports:: [[Minimum evidence package for a credible memory study]]
- supports:: [[Three layers reached used and correct]]
