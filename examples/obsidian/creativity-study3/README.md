# Round 3: do the relationship labels pay? A head-to-head

**Yes, in this head-to-head they did: modestly for originality, clearly for feasibility and test quality.** With the same four problems, the same twelve notes, the same reader model and the same rubric as round 2, the labeled condition beat the unlabeled selected condition on all three metrics, both in round 3 alone and pooled with round 2. The originality advantage is problem-dependent: labels won three of the four problems and lost the street-musician problem in both rounds.

## Pooled verdict (rounds 2 + 3)

Six generations per condition per problem; 72 anonymous ratings per condition from three reviewers. Means on five-point scales.

| Reader input | Originality | Feasibility | Discriminating test |
| --- | ---: | ---: | ---: |
| Three deliberately selected notes | 2.67 | 3.38 | 3.38 |
| Same notes with relationship labels | **2.81** | **4.00** | **3.99** |

Labeled minus selected, originality by problem: repair cafe **+0.33**, street musicians **−0.28**, community radio **+0.17**, bridge farewell **+0.33**. Labels won three problems, lost one.

## Round 3 alone

Four generations per condition produced 32 answers on November 24, 2026. Three reviewers independently saw each answer, anonymized and shuffled per reviewer.

| Reader input | Originality | Feasibility | Discriminating test |
| --- | ---: | ---: | ---: |
| Three deliberately selected notes | 2.54 | 3.50 | 3.46 |
| Same notes with relationship labels | **2.94** | **4.04** | **4.06** |

Reviewer agreement was weaker than in the four-condition rounds: 93 agreements, 40 disagreements, 203 ties across within-problem pairs (round 2: 149/19/168). With only curated conditions in the pool, designs within a problem resemble one another more, so ties and disagreements both rise: the head-to-head measures a smaller, noisier gap by construction.

## What each person or model saw

The selection was reused verbatim from round 2 (same twelve notes, same 650-byte excerpt cap), so no new curation time was spent (round 2's Kimi curator took 93.2 s). The only difference between the two conditions is the per-note relationship label. All eight reader sessions and all three reviewer sessions ran muse-spark-1.3 at xhigh, each in a fresh isolated workspace, pinned to that single provider. Problem order rotated per generation and reversed on odd ones. Answers ranged from 71 to 98 words; none needed truncation. [Selections](../creativity-study2/selection.json) · [Timing, input sizes and packet hashes](selection-run.json) · [Protocol](protocol.json) · [All answers](answers.json) · [Full scores](scores.json) · [Pooled analysis](pooled-scores.json)

## Use the result

Round 1 gave labels +0.33 originality across four problems on a GPT reader; round 2 gave −0.38 on a muse reader; this head-to-head gives +0.14 pooled with round 2 and a clean sweep within round 3 alone. The label effect on originality is real but small and problem-dependent: coordination-style briefs like the street musicians favored unlabeled notes in both muse rounds. The robust wins are feasibility (+0.63) and test quality (+0.61), where labels led in every round. Practical reading: when you select notes for a design brief, spend the extra minute writing relationship labels; they cost a little input text, rarely hurt originality, and consistently sharpen the proposed test.

## Reproduce

```sh
study=examples/obsidian/creativity-study3
python3 "$study/packets.py" "$study/run"
python3 "$study/drive.py" readers     # needs the pinned provider lane
python3 "$study/score.py" blind "$study/run"
python3 "$study/drive.py" reviewers
python3 "$study/score.py" score "$study/run"
python3 "$study/pool.py"              # pools with round 2's curated conditions
```
