# Do selected notes make designs more original?

**In this pilot, the highest-scoring condition gave the reader three deliberately chosen notes and a short relationship label for each.** It averaged 3.96/5 for originality. The question alone scored 3.50. The advantage depended on the design problem. The question alone produced the most original answers on two of the four problems.

[Round 2](../creativity-study2/README.md) repeated the comparison on fresh problems with a different reader model, and [round 3](../creativity-study3/README.md) pitted the two curated conditions head-to-head. The pooled verdict: curation beats search and the question alone on originality; labels add a small, problem-dependent originality edge and a consistent gain in feasibility and test quality.

## Scores

Four fresh prompts, four conditions and two generations per condition produced 32 answers on November 21, 2026. Three reviewers independently saw each answer. These are mean ratings on five-point scales. Each problem contributes equally.

| Reader input | Originality | Feasibility | Test distinguishes the idea from a simpler alternative |
| --- | ---: | ---: | ---: |
| Question only | 3.50 | 4.58 | 3.46 |
| Three ordinary search results | 3.38 | 4.79 | 4.21 |
| Three deliberately selected notes | 3.63 | 4.46 | 4.54 |
| Same notes with relationship labels | **3.96** | **4.83** | **4.58** |

| Design problem | Question only | Ordinary search | Selected | Selected and labeled |
| --- | ---: | ---: | ---: | ---: |
| Combine abandoned prototypes | 2.83 | 3.33 | 3.33 | **3.83** |
| Two builders share one postcard each | 3.33 | 3.50 | 4.33 | **5.00** |
| Instrument discovery for beginners | **3.83** | 3.50 | 3.67 | 3.50 |
| Revisit a design over six sessions | **4.00** | 3.17 | 3.17 | 3.50 |

The second table shows originality only. The labeled condition beat the unlabeled condition on three problems. It beat the question alone on two and lost on two. In the exhibit problem, both labeled designs received 5/5 for originality from every reviewer. One proposed a flashing firefly and a clicking cricket that run independently and synchronize when they see one another's beacons. Cover and uncover the detectors without changing their rates to test whether synchronization comes from interaction. On the six-session design problem, a question-only answer explored different mechanisms by reversing assumptions one at a time. A communal lamp could circulate if ownership changed, or ask for company if the direction of the signal changed.

## What each person or model saw

A Codex curator chose and labeled three vault notes per problem before generation. The timed interval was 23.89 seconds total (5.97 seconds per problem averaged over one batch). It includes inspecting the catalog and shortlisted prose, selecting notes, and writing the labels and selections. Human selection time was unmeasured. Question-only involved no note selection. The four local BM25 queries took 8.58 milliseconds together after 0.283 seconds to read and index the 526-note vault. Those are single wall-clock observations on this machine. [Selections](selection.json) · [Timing, input sizes and packet hashes](selection-run.json)

All eight reader sessions used GPT-6.2 Astra at ultra reasoning effort. Each saw the same instructions and answered all four problems for one condition. Problem order rotated across conditions and reversed for the second generation. Answers had a 110-word limit. The 32 answers ranged from 97 to 110 words, and none needed truncation. For source conditions, all three note excerpts used the same maximum of 650 UTF-8 bytes per note. Short notes used fewer bytes. The final two conditions have identical titles and excerpt text. The labels add words and include guidance, so this comparison does not isolate visual graph layout or semantics from extra prompt text. Readers received serialized notes and labels. [Prior scoring and selection protocol](protocol.json) · [All answers](answers.json)

Three fresh sessions of the same model and effort rated all 32 answers with independently shuffled problem and answer orders. Reviewers received anonymous answer IDs, problems and the same scoring rubric. They had no access to source packets, condition keys or one another's ratings. Across all pairs of designs within a problem and all pairs of reviewers, reviewers agreed on the originality order 142 times, disagreed 3 times, and at least one reviewer tied the designs 191 times. Agreement within one model family cannot establish human preference. Four problems and two generations per condition give a direction for another test, with considerable sensitivity to the problems and the curator. [Full scores](scores.json) · [Reviewer one](ratings0.json) · [Reviewer two](ratings1.json) · [Reviewer three](ratings2.json)

## Use the result

For questions where distinct mechanisms have to work together, select a few notes from different fields and label each with a concrete question. In the exhibit trial the labels asked what can run independently, what could take the place of direct signaling, and where mechanical agreement could mislead. For an open-ended design brief, a question-only draft is a cheap independent alternative. The two modes were evaluated separately here; [round 4](../creativity-study4/README.md) tested the combination and found that revising a draft with labeled notes raised originality in 12 of 16 pairs (never lowered it) but kept the draft's weaker tests.

Open **One postcard science exhibit** in the project vault for the editable winning setup. From the repository root, preview a short creative brief with the command below. Append `--question-only` for an independent draft. The live adapter identifies each file by path. The archived study reader saw its title. [Editable Canvas](<../../../idea-vault/One postcard science exhibit.canvas>) · [Blind review board with all 144 designs from rounds 1-4](<../../../idea-vault/Creativity blind review.canvas>)

```sh
python3 examples/obsidian/brief.py --creative 'idea-vault/One postcard science exhibit.canvas'
```

## Reproduce

From the checkout, the commands below rebuild the eight reader packets from the current vault and reproduce the scores from the saved ratings. Check the packet hashes against the saved selection run if the vault changes. For new independent generations, return one JSON object mapping problem letters to answer strings in `outN.json` alongside each corresponding `inputN.json`. Running `score.py blind "$work"` on those files produces separate anonymous packets for three reviewers.

```sh
study=examples/obsidian/creativity-study
work=$(mktemp -d)
python3 "$study/packets.py" "$work"
python3 "$study/score.py" score "$study"
```
