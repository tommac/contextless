# Round 4: does a question-only draft plus labeled notes beat either alone?

**Combining works, and the paired design shows why.** Revising a cheap question-only draft with labeled notes raised originality in 12 of 16 draft-revision pairs and lowered it in none, gaining +0.73 on average. The combined condition landed at 3.00 originality, close to reading labeled notes from scratch (3.15) and well ahead of the question alone (2.27). But the revision kept the draft's feasibility (4.04) and did not import the labels' test sharpness: combined tests scored 3.19, against 3.85 for labeled from scratch. Round 1 flagged this combination as untested; the answer is that it is a genuine third mode with a distinct profile, not a free lunch.

## Scores

Same four problems, same twelve labeled notes, same reader model (muse-spark-1.3 xhigh) and rubric as rounds 2-3. Four repeats per condition produced 48 answers on November 24, 2026. Combined repeat r revised the draft from question repeat r, so every combined answer has a paired question-only baseline. Three reviewers scored all answers blind.

| Reader input | Originality | Feasibility | Discriminating test |
| --- | ---: | ---: | ---: |
| Question only | 2.27 | **4.04** | 3.31 |
| Same notes with relationship labels | **3.15** | 3.73 | **3.85** |
| Question draft revised with labeled notes | 3.00 | **4.04** | 3.19 |

Paired draft-to-revision deltas (16 pairs, reviewer-averaged): originality **+0.73** (12 improved, 0 worsened), feasibility **0.00** (3 improved, 3 worsened), test **−0.13** (1 improved, 4 worsened). Revisions add originality without risk but inherit the draft's test structure instead of sharpening it. [Paired analysis](paired-analysis.json)

Reviewer agreement: 330 agreements, 18 disagreements, 444 ties across within-problem pairs. Answers ranged from 65 to 108 words; none needed truncation.

## What each person or model saw

Selections were reused verbatim from round 2, so no new curation time was spent. Question and labeled packets matched earlier rounds exactly. Each combined packet carried one repeat's question-only draft per problem plus the three labeled notes, with instructions to keep original mechanisms, extend whatever the material improves, and hold the same answer shape and 110-word ceiling. All twelve reader sessions and three reviewer sessions were fresh isolated workspaces pinned to muse-spark-1.3 xhigh; three sessions recovered from transient provider crashes on fresh retry workspaces. [Protocol](protocol.json) · [Timing and packet hashes](selection-run.json) · [All answers](answers.json) · [Full scores](scores.json)

## Use the result

The three modes now have distinct, evidence-backed profiles. Labeled notes from scratch give the sharpest tests and the top originality mean. A question-only draft is cheapest and most feasible. Revising that draft with labeled notes is the reliable middle: it buys most of the originality gain with zero observed downside on the draft's feasibility, but if the discriminating test matters (and in this study's design briefs it should), go straight to the labeled notes. This also closes the loop on round 1's recommendation to keep a question-only draft in contention: the draft is not just a competitor, it is a viable first stage, provided the revision does not need a better test.

## Reproduce

```sh
study=examples/obsidian/creativity-study4
python3 "$study/build4.py" "$study/run" 1      # question + labeled packets
python3 "$study/drive.py" readers              # runs 0-7
python3 "$study/build4.py" "$study/run" 2      # combined packets from drafts
python3 "$study/drive.py" readers              # runs 8-11
python3 ../studytool/studytool.py blind protocol.json run
python3 ../studytool/run_reviewers.py run cr4-rev study-config.yaml
python3 ../studytool/studytool.py score protocol.json run
python3 analyze4.py
```
