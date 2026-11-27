# Replay two passage-cache designs

Run `cx demo creativity` for the offline comparison. The original workload scans high-value passages only once and repeatedly revisits three ordinary passages. In the changed workload, three high-value passages each return once after six unrelated reads. Every event retains the same position, cost, and value across the two traces. Only three passage IDs change.

| Policy | Original trace | Changed trace |
| --- | ---: | ---: |
| Two-touch admission | 33 | 33 |
| Admit on first access and protect valuable passages | 9 | 75 |
| LRU | 27 | 27 |
| Full 2Q | 27 | 27 |

The score totals the cost times the value of every saved fetch. Both traces have 63 requests and four resident slots. These are authored demonstration workloads. To try a different pattern, copy [the workload](../../internal/creativity/testdata/replay.json), edit the page sequence or weights, and run `cx demo creativity -workload your-copy.json`. A segment can repeat named pages or scan a specified number of distinct pages. Weights and repeat counts default to one.

## Evaluate a new prototype

The study runner sends one JSON object on standard input for each prototype. For example:

```sh
cx demo creativity -evaluate <<'JSON'
{"family":"cache","scenario":1,"seed":42,"policy":{"admit":2,"history":2,"frequency":2,"recency":1},"test":{"workload":"scan","baseline":"lru","prediction":"win"}}
JSON
```

Each family has eight scenarios covering a recurring working set, scan pollution, bursts, a shifting working set, costly passages, repeated bursts, high-value arrivals, and delayed reuse. Their capacity or per-epoch read credits are 4, 6, 8, 4, 6, 8, 5, and 7. A runner keeps the hidden seeds and resulting traces outside the model packet. This source and its generator are public.

- **Cache.** A hit earns fetch cost times passage value. On a full cache the candidate evicts the resident with the lowest weighted score, resolving ties by oldest last access. `admit` is the number of observed requests required for admission, from one to three. `history` bounds metadata for pending admissions to zero through three times the resident capacity. An entry forgotten before admission starts over. The LRU baseline admits each miss. The 2Q baseline uses an incoming FIFO target of one quarter of capacity and a nonresident outgoing FIFO of one half, with a minimum of one slot each. A read from the outgoing FIFO promotes the key to its protected LRU queue.
- **Read scheduling.** Each epoch has new requests and a fixed read-credit allowance. Serving a pending request on or before its deadline earns its passage value and spends its cost. The candidate repeatedly selects the affordable request with the highest weighted score. Ties use arrival order. The FIFO baseline serves the oldest affordable request. The value-cost baseline greedily maximizes value per credit. Admission and history parameters apply only to caching.
- **Policy weights.** Each integer weight is between −4 and 4 and defaults to zero. Both families use passage value, read cost, age, recency, and frequency. Cache age measures duration of residency. Scheduling age measures waiting time. Cache frequency counts observed requests for a resident. Scheduling frequency and recency use a rolling window of the last three times capacity arrivals. Recency, age, and frequency features saturate at eight.
- **Tests.** Specify `scan`, `burst`, or `reuse`, choose a baseline for that family, and predict `win`, `tie`, or `loss`. The evaluator runs the intervention on a separate reproducible trace. The JSON report distinguishes a correct prediction from a score difference between the designs.
- **Success.** The evaluator records the resource peak, requires a strictly higher score than both baselines on the hidden main trace, and hashes a diagnostic decision sequence on three separate fixed traces. A signature match identifies equivalence on those probes. Distinctness is relative to the two comparison policies and those probes. The policy language itself enforces the resource bound for valid prototypes; the study runner counts invalid or missing submissions as failures.

## Edit the assumption in Obsidian

From the checkout, build the CLI and create the demo in a new directory.

```sh
make build
bin/cx demo creativity -out ./cache-workshop
bin/cx demo creativity -board './cache-workshop/Passage cache.canvas'
```

Open `cache-workshop` as an Obsidian vault. The **Design comparisons** Base shows the two admission designs, their predictions and measured outcomes. The independent four-slot design has one revision. In **Return pattern**, change the assumption and switch the last line from `Trace: after` to `Trace: before`. Run the second command again. It runs both policies on the selected workload and flags the admission decision. You can edit `workload.json` to supply your own observations. The trace selector makes the mapping from the source note to the experiment explicit.

Use the same workflow on another board with a `# Question` card and directly connected notes. Label an edge `constraint: <requirement>` for a requirement. Other labels carry inspiration. The text form of `obsidian brief -creative` bounds each excerpt at 650 UTF-8 bytes. Its JSON form exposes the stable edge keys used by the notebook.

```sh
board='./cache-workshop/Passage cache.canvas'
bin/cx obsidian brief "$board" -creative -json
bin/cx obsidian review "$board" -edge return-pattern
bin/cx obsidian design "$board" ./cache-workshop/admission.json
```

A design JSON contains an ID, its current parent revision (zero for a new design), title, prediction, outcome, and design choices. The example `admission.json` is ready for editing. Each choice declares the exact source edges it relies on through `any_of`. Use alternatives there only when any one of those sources independently supports that choice. Record distinct requirements as distinct choices. A review reports the changed source, its previous passage and its replacement. Another intact support can keep a choice supported. An inspiration change is shown separately. Removing an edge withdraws its support. A text edit requests a check even when the edited wording might have the same meaning.

To attach an executable prototype, add a `prototype` object in the same format as the evaluator request above. Saving runs it and records the measured result in the outcome property. Without a prototype, the supplied outcome is a user observation. Save a new revision after reviewing the evidence. Old revisions and their original passages remain available, and re-exporting a revision preserves annotations.

The complete archived source bodies and design receipts are stored locally in `.cx/creative.sqlite` in the selected vault. Include that database in a consistent vault backup. A note revision records the source SHA-256 and byte range. `bin/cx obsidian source "$board" SHA256` reopens the complete historical source, including after the original has been edited. When the board is in a subdirectory, pass `-vault` with the same vault root on every command. Keep a board's vault-relative path stable to address its saved design history.

## Scale experiment

`bin/cx demo creativity -scale` runs three independently applied edits at 1,000, 100,000 and one million unrelated indexed records. It compares following an explicitly changed edge, running the production keyword watcher on the update text, and reviewing all designs. The 4,096-byte allowance covers all candidate designs per edit. The index receives the edited edge ID. Arbitrary, unlinked prose corrections require separate retrieval.

At one million unrelated records, the tightened constraint sent one 265-byte packet for the affected design. The keyword watcher missed it. Review-all admitted that design and 15 unnecessary designs. All three comparison fixtures were sorted first, favoring review-all. These are locally serialized packets. Timing includes alert persistence for the watcher and a candidate count for review-all, with uncontrolled OS and SQLite caches. [Raw measurements](scale-results.json)

A separate resource probe gives every selector the same scripted repair. When the fixed reservation design reaches the reviewer, its configured slot count is aligned with the physical limit. The probe requests 12 distinct passages twice. After the limit fell from eight slots to four, the old design served 8 of 24 requests. Source-linked review and review-all delivered the repair and served 24 of 24. The keyword watcher served 8 of 24. Both control edits served 24 of 24 throughout. This measures the consequence of delivering a known repair on a fixed workload.
