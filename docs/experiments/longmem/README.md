# Public memory comparison

The offline command compares plain atom-level BM25 and ContextlessAI's existing fused atom ranker on all 500 questions from the cleaned LongMemEval-S release. Both use the same paragraph splitter, history order and SQLite FTS index. The ranker receives conversation text and the question. Session identifiers, gold labels and reference answers are available only to the scorer. Calendar metadata is rendered with the returned passages. The release has 470 answerable questions and 30 abstention cases.

## Reproduce the model-free result

From the repository root, with Go 1.26 and Python 3 installed, run the following in a shell. The public dataset is about 265 MiB.

```sh
work=$(mktemp -d)
data="$work/longmemeval_s_cleaned.json"
curl -fL --retry 2 -o "$data" https://huggingface.co/datasets/xiaowu0162/longmemeval-cleaned/resolve/98d7416c24c778c2fee6e6f3006e7a073259d48f/longmemeval_s_cleaned.json
go run ./docs/experiments/longmem -data "$data" > "$work/ranks.jsonl"
python3 docs/experiments/longmem/compare.py "$data" "$work/ranks.jsonl" --case-results "$work/cases.json"
```

Both programs check the dataset SHA-256. The first writes only ranks and source addresses. Each address includes a zero-based session, message, and UTF-8 byte range. For a fast inspection, add `-id 830ce83f` to the Go command. On that knowledge-update question plain BM25 included text from all three annotated messages. The fused ranker reached two. On `55241a1f` the result went the other way. [All 500 outcomes](../../../reports/2026-11-20-experiments/public-cases.json)

Each arm ranks 24 atoms. In rank order the scorer takes whole passages that fit in **8,000 UTF-8 bytes**, including headers. This budget is independent of the private study's token budget. A source message counts as represented when any paragraph of that message is present. The model-free score cannot establish that the specific answer sentence survived or that the reader used it.

## Public answer experiment after the frozen study

Use the same reader model snapshot across all arms, and keep the full-history arm subject to the model's actual context limit. Record overflows and failed cells. The primary compression comparison uses three treatments of **the identical ranked ContextlessAI candidates**.

| Arm | What the reader gets |
|---|---|
| verbatim | Original passages that fit, selected in rank order |
| summary | If the pool overflows, one question-conditioned compression call over the whole candidate set. An over-budget summary is a recorded failure. Otherwise the original passages go through. |
| overview | On overflow the reader sees an ID and short opening for each candidate and requests IDs once. The overview and the selected original passages share the final byte budget. |
| bm25 | Same verbatim packing over an independently generated BM25 ranking |
| full | All timestamped user and assistant messages, without a retrieval budget |
| mem0 | A fresh local memory store ingests the timestamped conversation one session at a time. Search returns up to 24 memories under the explicit 0.1 threshold. |
| graphiti | A distinct graph group ingests dated episodes in order. Search returns up to 24 facts with validity times. |

The comparison isolates final-reader evidence space. Total compute differs by design. The extra summary call sees the entire candidate pool, and the overview reader pays for its selection turn. Report model-resolved identifiers, input, output, cache use, latency, failures and per-arm accuracy. For Mem0 and Graphiti, measure provider-side ingest, retrieval and embedding usage independently before making a cost claim. Their SDK defaults and dependency versions belong in the run manifest. The native adapters have been checked against pinned source. End-to-end execution with those dependencies remains pending.

The live driver requires a provisioned OpenAI API reader. Pick the reader snapshot once and use the same value for every arm. For a one-question integration check, set the shell variable `reader` to that snapshot and run:

```sh
python3 docs/experiments/longmem/compare.py "$data" "$work/ranks.jsonl" \
  --arm overview --id 830ce83f --model "$reader" --out "$work/answers"
```

Omitting `--id` processes all 500 cases and writes a blind `hypotheses.jsonl` with only question IDs and answers. Completed cells are reused on restart. A failed cell stops the batch and is written separately for inspection. Changing the reader, ranking file or runner requires a separate output directory. Treat any retries as additional attempts in the cost and failure report.

The Mem0 arm expects `mem0ai` installed from commit `fff8a6472512d1bc1a4bccab8c645b65a1df7fe5` in its own environment. Graphiti expects `graphiti-core` from commit `9b8a4630092084734b080f2e6f409c2c23bdcc3a` and a disposable Neo4j benchmark database configured with `LME_NEO4J_URI`, `LME_NEO4J_USER` and `LME_NEO4J_PASSWORD`. Native retrieval can make paid model calls during ingest. The driver checks the installed VCS revision before starting either arm. [Mem0 source](https://github.com/mem0ai/mem0/tree/fff8a6472512d1bc1a4bccab8c645b65a1df7fe5) · [Graphiti source](https://github.com/getzep/graphiti/tree/9b8a4630092084734b080f2e6f409c2c23bdcc3a)

For scoring, use the same upstream judge revision and model on every arm. From the pinned upstream repository's `src/evaluation` directory, its command is `python3 evaluate_qa.py gpt-4o HYPOTHESES CLEANED_DATA`. That revision maps the alias to `gpt-4o-2024-08-06`. Verify availability before starting a scored batch. Keep the judge blind to arm names. Report paired wins, losses, ties, abstention accuracy and results by overflow alongside overall scores. Only 11 of the 500 ContextlessAI candidate pools fit the chosen budget, so conclusions about the fit subgroup will be uncertain. [Upstream evaluator](https://github.com/xiaowu0162/LongMemEval/blob/9e0b455f4ef0e2ab8f2e582289761153549043fc/src/evaluation/evaluate_qa.py) · [Dataset release](https://huggingface.co/datasets/xiaowu0162/longmemeval-cleaned/tree/98d7416c24c778c2fee6e6f3006e7a073259d48f)
