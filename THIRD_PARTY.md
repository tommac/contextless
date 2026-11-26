# Third-party data and code

The Apache License in `LICENSE` covers the code and documents written for
this project. It does not cover the items below, which keep their own terms.

## Benchmark data

No benchmark dataset is committed here. The scripts download each one from
its own repository at a pinned revision. Some result files hold material
derived from them:

| Source | Terms | What this repository holds |
|---|---|---|
| [LoCoMo](https://github.com/snap-research/locomo/tree/3eb6f2c585f5e1699204e3c3bdf7adc5c28cb376) (Maharana et al.) | CC BY-NC 4.0 ([license](https://github.com/snap-research/locomo/blob/3eb6f2c585f5e1699204e3c3bdf7adc5c28cb376/LICENSE.txt)): attribution, non-commercial | `reports/2026-11-20-large-memory/locomo-*`: question ids, categories and per-arm outcomes. No dialogue text. |
| [LongMemEval](https://github.com/xiaowu0162/LongMemEval) (Wu et al.) | See its repository | `reports/2026-11-20-experiments/public-cases.json`: question text and per-arm outcomes for a sample of cases; aggregate counts elsewhere. |
| [ContextBench](https://github.com/RulinShao/ContextBench/tree/fb533c18c78a82d3f1b2047a9c1700735cf829ae) | See its repository | Aggregate counts only. |
| Go API files (`api/go1.*.txt`, golang/go) | BSD 3-Clause | Not committed; `examples/answer_watch` downloads them. |

If you reuse those result files, follow the source's terms, including the
non-commercial condition on anything derived from LoCoMo.

The two long conversations behind the answer tests in the README are private
and are not in this repository.

## Go modules

Fetched by `go build`, not copied into this repository:

| Module | License |
|---|---|
| `modernc.org/sqlite` and the `modernc.org` modules it needs | BSD 3-Clause |
| `gopkg.in/yaml.v3` | MIT and Apache 2.0 |
| `github.com/google/uuid` | BSD 3-Clause |
| `github.com/dustin/go-humanize`, `github.com/mattn/go-isatty`, `github.com/ncruces/go-strftime` | MIT |
| `github.com/hashicorp/golang-lru/v2` | MPL 2.0 |
| `golang.org/x/*`, `github.com/remyoudompheng/bigfft` | BSD 3-Clause |

## Code from the author's other projects

`internal/providers` was copied from BookWriter v2, and `internal/packet` and
`internal/route` re-implement designs from BookWriter v2 and MultiAIWorkspace.
Both are by the same author and are published here under the Apache License.

## Model CLIs

`cx` starts command-line tools from Anthropic, OpenAI, Google, Moonshot, xAI
and others when you enable them. They are not part of this repository, and
your use of each is under its vendor's terms.
