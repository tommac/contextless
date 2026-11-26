# Round 2: fresh problems, same protocol

**In this round the highest-scoring condition gave the reader three deliberately chosen notes without labels.** It averaged 2.92/5 for originality. The labeled variant followed at 2.54, ordinary search at 2.42, and the question alone last at 1.83. Round 1's label advantage did not replicate, but the broader pattern did: both curated conditions outscored both uncurated ones, and the question alone still produced the most original answer on one of the four problems.

Absolute scores are not comparable across rounds. Round 1 used GPT-6.2 Astra readers and reviewers; this round used muse-spark-1.3 at xhigh for every session, because the codex lane was quota-limited and the claude lane auth-limited on the day. Comparisons within the round are unaffected: all conditions shared one reader, one instruction set and one output limit.

## Scores

Four fresh problems, four conditions and two generations per condition produced 32 answers on November 24, 2026. Three reviewers independently saw each answer. These are mean ratings on five-point scales. Each problem contributes equally.

| Reader input | Originality | Feasibility | Test distinguishes the idea from a simpler alternative |
| --- | ---: | ---: | ---: |
| Question only | 1.83 | **4.71** | 3.04 |
| Three ordinary search results | 2.42 | 3.58 | 3.25 |
| Three deliberately selected notes | **2.92** | 3.13 | 3.21 |
| Same notes with relationship labels | 2.54 | 3.92 | **3.83** |

| Design problem | Question only | Ordinary search | Selected | Selected and labeled |
| --- | ---: | ---: | ---: | ---: |
| Repair cafe skill-building | **3.00** | 2.33 | 2.67 | 2.67 |
| Two street musicians, one crowd | 2.17 | 2.17 | **3.17** | 1.83 |
| Unattended community radio | 1.17 | **2.50** | **2.50** | 2.33 |
| Saying goodbye to a bridge | 1.00 | 2.67 | **3.33** | **3.33** |

The second table shows originality only. Selected beat labeled on two problems, tied on one and lost on one: round 1's ordering reversed. Labels still earned the best average for the discriminating-test score and the best feasibility among the sourced conditions. The clearest curated win came on the street-musician problem: the top design (4.00 originality) split the night into clock-synced 20-minute supersteps with a joint five-minute silence at each barrier, so wanderers redistribute and neither end can lock the crowd in, with no signaling required. The question-only win came on the repair cafe: chairs grouped into "fault families" (dead, intermittent, jammed) so visitors open same-fault devices together, tested by swapping brands the next Saturday.

## What each person or model saw

A Kimi curator chose and labeled three vault notes per problem before generation. The timed interval was 93.2 seconds total and includes inspecting the 526-title catalog, reading twelve shortlisted notes' opening prose, selecting, and writing the labels and selections; no note overlapped the round-1 selections. Round 1's Codex curator took 23.9 seconds for the same job, so curation cost varies by curator. The four local BM25 queries took 4.9 milliseconds together after 0.157 seconds to index the vault. [Selections](selection.json) · [Timing, input sizes and packet hashes](selection-run.json)

All eight reader sessions used muse-spark-1.3 at xhigh reasoning effort, each in a fresh isolated workspace, pinned to that single provider with cross-provider review disabled. Each session saw the same instructions and answered all four problems for one condition. Problem order rotated across conditions and reversed for the second generation. Answers had a 110-word limit; the 32 answers ranged from 71 to 103 words and none needed truncation. Source conditions used the same 650-byte excerpt cap as round 1. One reader session's first attempt died in a provider crash and was rerun in a new fresh workspace; one completed reply was re-parsed from its saved raw text after a harness bug. Neither affects session isolation. [All answers](answers.json) · [Protocol](protocol.json)

Three fresh sessions of the same model and effort rated all 32 answers with independently shuffled problem and answer orders, under the round-1 rubric. Reviewers agreed on the originality order 149 times, disagreed 19 times, and tied 168 times across all within-problem design pairs and reviewer pairs, noticeably more disagreement than round 1's 142/3/191, which is itself a reader-model effect worth noting when comparing rounds. [Full scores](scores.json) · [Reviewer one](ratings0.json) · [Reviewer two](ratings1.json) · [Reviewer three](ratings2.json)

## Use the result

Across both rounds, the two curated conditions occupy the top two originality slots; ordinary search never beats them on average, and it ranked last in round 1. Whether to add labels is now genuinely uncertain: +0.33 originality in round 1, −0.38 here, with labels still improving test quality in both rounds (4.58 and 3.83, each the condition maximum). A reasonable reading is that selection is the load-bearing step and labels pay when the design hinges on combining mechanisms across fields, as in round 1's exhibit problem. Keep a question-only draft in contention for open briefs: it won one problem in each round and remains the cheapest input.

## Reproduce

From the checkout, the commands below rebuild the eight reader packets from the current vault and reproduce the scores from the saved ratings. `drive.py` recreates the live sessions through fresh cx workspaces pinned by `study-config.yaml`; rerun it only for new independent generations.

```sh
study=examples/obsidian/creativity-study2
python3 "$study/packets.py" "$study/run"
python3 "$study/drive.py" readers     # needs the pinned provider lane
python3 "$study/score.py" blind "$study/run"
python3 "$study/drive.py" reviewers
python3 "$study/score.py" score "$study/run"
```
