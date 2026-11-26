# Which answers need a second look?

November 20, 2026

An agent that keeps the original archive can notice when new evidence changes an earlier answer. A source citation alone can miss the update. All the old quotes may still be accurate. Two experiments test what information the agent should keep alongside a quote.

## Save the question an answer depends on

Go 1.26 added `bytes.Buffer.Peek`. A developer who previously asked whether that declaration appeared in the release ledger deserves an update. A developer asking specifically about the ledger through Go 1.25 should keep the earlier answer. The Go project publishes the individual declarations and freezes each release file. [Release record](https://github.com/golang/go/blob/go1.26.0/api/go1.26.txt#L1) · [Go API ledger format](https://github.com/golang/go/blob/go1.26.0/api/README)

The replay starts with 334 records from Go 1.24 and 1.25 and appends the 180 records from 1.26. It saves the exact condition behind each answer, either the presence of one declaration or the number of distinct declarations in one package. Every question has a live and a historical version. An inverted index maps each condition back to the questions that depend on it. On append, the program checks the new signature and its package, applies the time cutoff and keeps the original source line for each change. Repeating the same release produced zero further updates.

| How to select answers for review | Answers reopened | Changed answers missed | Unchanged answers reopened |
| --- | ---: | ---: | ---: |
| Recompute everything | 1,128 | 0 | 924 |
| Reopen when previously cited bytes change | 0 | 204 | 0 |
| Reopen current questions in any affected package | 402 | 0 | 198 |
| Reopen when a new record satisfies the saved condition | 204 | 0 | 0 |

Full replay and the saved conditions agreed on all 1,128 answers. The 204 changes comprised 180 declaration-presence answers and 24 package counts. Each updated answer had a new matching source record. The workload is retrospective. It includes every distinct declaration and package in the three releases so it exercises real changes, along with matching historical controls. The experiment measures logical update selection on two exact predicates over append-only inputs. It does not estimate how often users' questions change, LLM accuracy, or the cost of extracting a condition from prose. A historical answer covers the supplied ledger starting at Go 1.24.

Field-level dependencies extracted from agent read traces already have public evaluations. A November preprint found that omitted alternative support could break selective repair. The additional case to test in conversation is an answer based on the *absence* of a matching record, such as “Nobody has approved the revised date for instructors.” Persisting what was checked lets a later approval trigger review even if the approving person uses different words. A free-text implementation still has to identify the event, audience and authority, and handle a second independent approval. [Dependency-aware context graphs](https://arxiv.org/abs/2611.19859) · [PostgreSQL predicate reads](https://www.postgresql.org/docs/current/transaction-iso.html#XACT-SERIALIZABLE)

### Reproduce the release replay

With a Go distribution that includes all three API files, pass its `api` directory to the script. Add `--list-changes` to see every revised answer with its old value, new value and original source lines. You can also fetch the three frozen files from the public release. The replay checks all three SHA-256 digests. Python 3.9 or newer is sufficient.

```sh
api=$(mktemp -d)
for v in 24 25 26; do
  curl -fsSL "https://raw.githubusercontent.com/golang/go/go1.26.0/api/go1.$v.txt" -o "$api/go1.$v.txt" || exit 1
done
python3 examples/answer_watch/replay.py "$api"
```

## Remember who answered

The previous public retrieval experiment excluded LoCoMo's 446 adversarial questions. This audit pairs questions within a dialogue when swapping the two participants' names gives the same wording after normalization. It requires identical, nonempty evidence IDs and a positive reference answer that matches the adversarial lure. Of 278 wording matches, 271 also matched the lure. Four had multiple candidate positives. The remaining 267 pairs span all ten dialogues.

Given the annotated turns, a rule that accepts a fact only when every source speaker is the person named in the question matched both published decisions on 259 of 267 pairs. It abstained on both members of three mixed-speaker pairs. It reversed the expected decisions on five pairs. One of those five has a reversed label in the public benchmark. On the other four the indexed speaker was asking a question. An answer was in the surrounding exchange. These are checks of source attribution using supplied evidence IDs. A person can also say something true about somebody else. [LoCoMo](https://aclanthology.org/2024.acl-long.747/) · [Public label correction](https://github.com/snap-research/locomo/issues/43) · [Prior work on speaker errors](https://arxiv.org/abs/2609.03467)

We tested those four inspected cases and four other-person controls. The controls came from four other dialogues, taking the first qualifying single-turn pair per dialogue after sorting by the SHA-256 of its positive ID. Three fresh Codex assistant sessions each saw the same eight questions with one type of packet, without gold answers or condition labels. The instruction was to answer only from each question's passages, cite the turn IDs, and reply “unknown” if they did not establish the answer for the named person.

| Input to reader | Direct reference-consistent answers on four selected positives | Declined to assign the fact to the other person on four controls | Total serialized bytes¹ |
| --- | ---: | ---: | ---: |
| Annotated turn alone | 0 | 4 | 2,689 |
| Same turn plus up to two ordinary BM25 hits | 1 | 4 | 5,691 |
| Same turn with its predecessor and successor in the session | 4 | 4 | 5,920 |

¹ Sum of eight individually serialized packets, including questions, IDs, speaker labels, anchor dates and JSON delimiters. Search had the same per-case ceiling and maximum turn count as the adjacent window. It searched across the entire dialogue with SQLite FTS5 Porter stemming. Each arm started with the benchmark's annotated turn. Deploying this policy also requires finding the starting turn.

| Answerable case | Detail in the reference | Decisive reply | What the ordinary-search reader did |
| --- | --- | --- | --- |
| conv-48:152 | A bench evokes conversations with her mother | D19:19 | Declined. The retrieved bench passage said it had personal meaning. |
| conv-26:107 | Running helps her de-stress and clear her mind | D7:20 and D7:22 | Declined while naming the benefit in D7:20. The reader distinguished a current benefit from the reason she started. The next reply addressed her motivation. |
| conv-26:106 | She uses the shoes for running | D7:20 | Declined. The anchor asked whether the shoes were for walking or running. |
| conv-26:139 | Transgender poets shared personal stories | D17:19 | Answered and cited the reply. |

The four positives were deliberately chosen after inspecting a failure. Each condition had one reader run. The table describes reader behavior on these examples, with manual comparison to the reference. It cannot predict an overall accuracy gain. The next controlled test should freeze a rule for recognizing unanswered questions, retrieve the starting turn without labels, and count unnecessary neighbor reads as well as recovered answers across fresh conversations.

### Reproduce the conversation inputs

The scripts use the pinned LoCoMo-10 release under CC BY-NC 4.0. They verify the dataset digest. The second command writes packets containing source text to a temporary directory. It makes no model calls. [Release and licensing](../docs/experiments/locomo/README.md)

```sh
work=$(mktemp -d)
curl -fsSL https://raw.githubusercontent.com/snap-research/locomo/3eb6f2c585f5e1699204e3c3bdf7adc5c28cb376/data/locomo10.json -o "$work/data.json"
python3 docs/experiments/locomo/speaker_swap.py "$work/data.json"
python3 docs/experiments/locomo/reply_windows.py "$work/data.json" "$work/packets"
```
