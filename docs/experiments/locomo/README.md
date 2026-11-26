## LoCoMo-10 cross-dataset retrieval check

The unchanged excerpt selector was applied to a different public conversational benchmark after its LongMemEval-S run. The original data and annotations are from Maharana et al., *Evaluating Very Long-Term Conversational Memory of LLM Agents* (ACL 2024). LoCoMo combines generated and human-edited conversations. Its release is licensed under CC BY-NC 4.0. The adapter and result files contain no conversation text. [Paper](https://aclanthology.org/2024.acl-long.747/) · [Pinned release](https://github.com/snap-research/locomo/tree/3eb6f2c585f5e1699204e3c3bdf7adc5c28cb376) · [License](https://github.com/snap-research/locomo/blob/3eb6f2c585f5e1699204e3c3bdf7adc5c28cb376/LICENSE.txt)

Across 1,533 questions, the original atom pool covered all annotated messages on 890 (58.06%). Adding the routed lane reached 958 (62.49%), with 84 paired gains and 16 losses. The net gain was positive in each dialogue. Every candidate pool fit the 8,000-byte cap. All three packing policies therefore gave identical coverage, literal-match scores and byte counts within each pool on every case. The primary BM25 and fused lanes produced different candidate sets on 193 questions, yet had identical all-annotated-message outcomes for every question here. The median whole-packet sizes on the BM25 lane were 4,032 and 4,103 bytes, respectively.

| Dialogue | Eligible | Atom pool, whole | Routed pool, whole or growing |
|---|---:|---:|---:|
| conv-26 | 150 | 88 | 91 |
| conv-30 | 81 | 55 | 58 |
| conv-41 | 152 | 90 | 99 |
| conv-42 | 197 | 117 | 121 |
| conv-43 | 178 | 106 | 114 |
| conv-44 | 123 | 62 | 67 |
| conv-47 | 149 | 82 | 96 |
| conv-48 | 191 | 113 | 124 |
| conv-49 | 156 | 85 | 89 |
| conv-50 | 156 | 92 | 99 |
| **All** | **1,533** | **890** | **958** |

The strict-ID slice reached 889 versus 957 of 1,527. Within the 480 questions with literal answers in a labeled turn, packets retained that answer on 332 versus 363 (34 gains and three losses). Excluding answers also present verbatim in the question gave 330 versus 361 of 478. The secondary lane changes both the session filter and the indexed speaker metadata. This experiment measures that complete routing policy; it cannot attribute the gain to either component alone.

The evaluation covers 1,533 of 1,986 released questions. It excludes all 446 category-5 adversarial questions, four without evidence, and three containing at least one unresolvable annotation. Six included questions needed only normalization of list separators, an extra colon after the letter D, or leading zeroes to existing identifiers. A separate result for the 1,527 questions using strict exact ID matches allows comparison without these repairs. Each question ID in the result files is the original sample ID followed by the zero-based QA array index. Evidence IDs resolve within their own conversation.

Numbered sessions keep their original order, and each turn's only searchable body in the primary lane is its original text. The native speaker name becomes the role and the original session date-time string becomes the date. The secondary lane includes the role in message FTS, and uses dates when ranking sessions. Packet headers include the date-time and speaker. This text-only treatment excludes images, captions, URLs, generated observations and summaries. Thus 1,226 of the 5,882 messages have a released caption which the reader did not see. Native turn text is short: median 109 UTF-8 bytes, 95th percentile 257, maximum 454.

Both primary lanes use the production Go paragraph splitter and ranker. The Go adapter only replaces the dataset digest and question-count guards on the existing LongMemEval ranking driver. BM25 and the Contextless fused lane each supply up to 24 candidates. Each lane is separately crossed with the original ranked list or its interleave with session-routed full messages, then each resulting pool is packed with whole items, first-span growth or query-directed span growth. Each frozen arm uses the same maximum of 8,000 UTF-8 bytes with headers and delimiters. Ranks receive the raw question and turns. Only the scorer receives reference answers and annotated turn IDs. The objective is to have any original byte span from every annotated source turn in the packet. Literal reference coverage checks for a normalized, exact contiguous answer in a selected span from an annotated turn, conditional on one existing in an original annotated text. Neither measure establishes answer accuracy or reader utility.

Run `sh path/to/reproduce.sh` from the ContextlessAI checkout with `adapter.py` next to the shell script. Go 1.26 and Python 3.9+ with SQLite FTS5 are required. The runner checks dataset SHA-256 `79fa87e90f04081343b8c8debecb80a9a6842b76a7aa537dc9fdf651ea698ff4`, unmodified selector SHA-256 `028f9324af76918a3626c93e58c1856b2551b1d4ee510b9b17f5536e19971131`, and the hashes of its imports and Go rank driver. The current scoring-only edit (SHA-256 `85c921390dcc4c549713df65ab28dc1b16714588ade5583abd58b4c23e632e36`) was also accepted after comparing all eleven non-scoring Python definitions byte-for-byte. That edit adds a slice of literal cases where the answer is absent from the question. It uses the committed Go ranker at `04e171c3a8b80160179ceb84042c23eda2d7d7cb`. Outputs include aggregate and per-dialogue metrics, all 1,533 scored question IDs, and all 453 exclusions with reasons.
