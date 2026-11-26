# Study toolkit: shared harness + AI-only reviewer calibration

One tested module replacing the per-round script copies in
creativity-study{,2,3}, plus a calibration mechanism that measures reviewer
quality without any human scoring.

## Why calibration

The study vault flags two measurement risks: LLM reviewers are lenient or
noisy, and judges otherwise need human calibration. Seeded-error calibration
answers both with code: `seeds.py` builds twelve designs with one known
defect each (a generic pattern, a constraint violation, or an uninformative
test), `blind` mixes them into the anonymous review packet under 9xx ids, and
`score` reports each reviewer's catch rate (a score of at most 2 on the
metric the seed was built to fail) and its leniency (mean score on seeds
versus real answers). A reviewer who cannot catch planted defects should not
be trusted on subtle ones.

## Commands

```sh
python3 studytool.py build  protocol.json selection.json workdir
python3 studytool.py blind  protocol.json workdir [--seeds seeds.json]
python3 studytool.py score  protocol.json workdir
python3 run_reviewers.py    workdir <cx-workspace-prefix> <config.yaml>
python3 -m unittest test_studytool
```

`build` covers any condition list and repeat count (rounds 1-3 are special
cases). `blind` clips at the 110-word ceiling, anonymizes into 1xx ids, and
keeps seeds out of `answers.json` so they never enter condition means.
`score` reproduces the per-round summaries and appends a `calibration`
section when seeds are present.

## First calibration result

Measured on round 3's 32 answers plus the 12 seeds, three fresh
muse-spark-1.3 reviewers. All three caught 12 of 12 planted defects
(catch rate 1.0), scoring seeds 1.5-1.8 on average against 3.5-3.8 for
real answers: the reviewers are harsh on clear defects, not lenient.
Re-rating the same real answers reproduced round 3's direction on all
three metrics (labeled 3.13 vs selected 2.98 on originality), so the
head-to-head conclusion is stable under reviewer resampling. Full data:
`../creativity-study3/calibration/calibration-report.json`.
