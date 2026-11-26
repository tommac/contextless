---
type: source
status: external
area: evaluation
needs: none
year: 2026
url: https://arxiv.org/html/2610.31710v2
---
Compact-Bench is an October 2026 preprint that replays each of 35 SWE-bench repairs twice, once with the full history and once with only a summary, and judges both by the repository's tests. Each case is labelled preserved, damaged, rescued or unsolved, so gains cannot hide losses. It is the report's model for paired replay and its strongest outside case for the verbatim rule.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "paired full-history versus summary-only replay on 35 SWE-bench repairs, judged by repository tests"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "each case is labeled preserved, damaged, rescued or unsolved, so gains cannot hide losses"
- `reports/Cutting edge context memory research.md`: "Strongest external case for the verbatim rule"

## Relations
- supports:: [[Atom-level causal evaluation]]
- supports:: [[Lossy compression loses constraints and exact text]]
- supports:: [[A run that compacts is thrown away]]
- tested-by:: [[Compaction damage ranged from 5 to 64 percent across 14 summarizers]]
