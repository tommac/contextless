# What helped a working prototype?

On November 27, 2026, revisions with deliberately selected, labeled notes beat both reference policies in **11 of 32** cases. Free revision of the same starting drafts did so in **9 of 32**. The paired difference was **+6.25 percentage points**, with a 95% bootstrap interval from **−9.38 to +21.88 points**, resampling the 16 briefs. The experiment leaves the advantage of curation unresolved. The random-note condition produced the most successful policies in this sample.

| Revision material | Successful / 32 | Predicted test outcome confirmed / 32 | Blind originality / 5 | Blind feasibility / 5 | Blind test quality / 5 | Revision input tokens |
|---|---:|---:|---:|---:|---:|---:|
| Draft alone | 9 | 18 | 2.74 | 3.93 | 3.24 | 60,046 |
| Ordinary search | 9 | 14 | 2.71 | 3.92 | 3.49 | 72,101 |
| Random cross-domain notes | 12 | 17 | 2.63 | 3.89 | 3.31 | 71,262 |
| Selected notes, neutral labels | 10 | 20 | 2.60 | 3.88 | 3.36 | 73,076 |
| Selected notes, meaningful labels | 11 | 17 | 2.73 | 3.79 | 3.43 | 73,132 |

All 160 candidates ran, met the simulator resource limits, and produced a behavioral fingerprint different from both reference policies on the diagnostic cases. Those two criteria therefore did little to distinguish the study conditions. Measured wins depended on hidden workload performance. For caching the fingerprint is the hit/miss sequence, and for scheduling it records the simulated service decisions. Broader algorithmic novelty and human preference remain open.

Two cases show why execution matters. On the expensive-passage field guide, one labeled revision scored 603 weighted saved-fetch points and its free revision scored 563. The stronger baseline scored 595. On the tutor's worksheet scan the ordering reversed. Free revision scored 238, labeled revision 228 and the stronger baseline 230. Each comparison starts from the same draft.

## What went into the run

- Eight cache briefs and eight read-scheduling briefs each received two independently generated starting drafts. Five revisions reused each starting draft. GPT-6.1 Sol at high effort produced the drafts and revisions. Each generation packet contained four briefs, a common output ceiling and a 110-word explanation limit. Curation took 56.36 seconds for 48 note selections and labels, including coordination. This was one AI curator. The two curated arms saw equal excerpt text, label word counts and padded byte lengths. Neutral captions sometimes sound artificial.
- Three independently shuffled Claude Opus 5.5 high-effort reviews per brief yielded 480 blinded design ratings across 48 packets. Every packet included one planted weak test, identified in 48 of 48 reviews. These are repeated judgments from one model family. Readers and reviewers had separate sessions, and the reviewer packets contained no conditions or simulator outcomes.
- The protocol committed to one hidden workload seed per brief before generation. The demonstration workload and the public study use separate traces. Tests exercised two families of bounded policies and one reader. Cross-domain inspiration winning the largest raw count is an exploratory observation. The single prespecified comparison was labeled versus free revision.
- The 96 accepted calls used 8 draft packets, 40 revision packets and 48 judge packets. Three earlier draft adapter attempts were discarded after the Codex CLI emitted a context-window metadata warning. No answer from those attempts entered the experiment. Recorded token totals cover accepted calls. Usage for discarded attempts is unavailable. The run manifest captures the generation-time script digest. The scoring and reporting implementation was finalized during the run. The replay checks the saved content and digest of every accepted request.
- The same source ledger is useful independently of the ratings. On the editable demo, changing the observed reuse pattern identifies which admission choice needs review and reopens the exact prior passage. The source selector also has a separately measured scale experiment. [Demo and measurements](../../creativity-demo/README.md)

[Offline replay instructions](README.md#replay-the-november-27-run) · [Recorded packets, prototypes, ratings, usage and seeds](2026-11-27-run.json.gz) · [Frozen study protocol](protocol.json)
