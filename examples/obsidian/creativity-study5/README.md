# Does note curation improve a working policy?

Sixteen public briefs ask a reader to design a policy for a bounded passage cache or read queue. For each brief the reader makes two independent drafts. Each draft receives five independent revision attempts. The main comparison is between revision with labeled notes and free revision of the same draft. An empty run directory can be prepared offline. A new scientific result requires real reader outputs.

| Arm | Revision material |
| --- | --- |
| Free | Starting draft and brief |
| Search | Three BM25 results from the public idea vault |
| Random | Three notes sampled without replacement from a fixed list of outside-field concepts, excluding the curated picks for that brief |
| Neutral | Three deliberately selected notes with neutral captions |
| Labeled | The same three notes with a curator's relationship labels |

The two curated packets use identical note titles and excerpts. Neutral and meaningful captions have equal word counts and equal padded UTF-8 lengths. Actual token usage can differ under the provider tokenizer. “Distant” means membership in the predeclared list of ten concepts outside cache engineering and scheduling. The comparison tests this pool. It does not estimate a continuous semantic-distance effect.

[Protocol](protocol.json) · [48 selected notes, labels, and curation timing](selection.json)

## Run

Run the following commands from the public checkout with Python 3.9+ and an existing build of the current Go CLI. Substitute the verified reader and judge identities and the absolute paths to the study adapters. Keep model, effort, provider settings, and adapter contents fixed for the run.

```sh
study=examples/obsidian/creativity-study5/study.py
work=$(mktemp -d)
python3 "$study" prepare "$work" --reader "$READER_MODEL" --reader-family "$READER_FAMILY" \
  --judge "$JUDGE_MODEL" --judge-family "$JUDGE_FAMILY"
python3 "$study" run "$work" --stage draft --adapter "$READER_ADAPTER"
python3 "$study" run "$work" --stage revise --adapter "$READER_ADAPTER"
python3 "$study" run "$work" --stage judge --adapter "$JUDGE_ADAPTER"
python3 "$study" score "$work" --cx "$PWD/bin/cx"
```

Preparation freezes the complete set of reader materials and commits to one randomly generated evaluation seed per brief by SHA-256. The runner writes the seeds and condition keys in the run's private directory. Keep that directory away from readers, curators, and judges through completion. Saved seeds in the final results reproduce the workload. Each reader call sees four problems for one condition and repeat in a fresh directory. Thirty-two drafts produce 160 final proposals in 48 generation calls. Three independent reviewer sessions per brief produce 48 review calls. Reviewers share one pinned model from a family distinct from the reader. The model family identifies the underlying model lineage independently of its serving route.

An adapter is an executable that accepts exactly one JSON object on stdin with `model`, `family`, `instructions`, an opaque-ID-to-task mapping in `tasks`, and `max_output_tokens`. It creates a fresh, tool-free model session, enforces the output allowance, and returns the provider-resolved identity and actual usage. Its stdout must contain a single JSON object.

```json
{"model":"resolved-id","family":"family","answers":{"task-id":{"policy":{"recency":2,"value":1},"test":{"workload":"reuse","baseline":"lru","prediction":"win"},"explanation":"The policy protects recently useful passages. Test whether it saves more weighted fetch cost than LRU."}},"usage":{"input_tokens":512,"output_tokens":88}}
```

A reviewer returns three integer scores in each answer array. Put `null` in an unavailable token field. Set `--jobs` according to existing provider capacity. Run the same stage command again after an interruption. Successful jobs are bound to their input digest and skipped. Each failed attempt retains stdout, stderr, and elapsed time under `attempts/`. Inspect these local logs before sharing them. No provider account, automatic paid fallback, or network access is needed for offline scoring.

For an existing model-running workflow, use `packets "$work" --stage draft` (or `revise` / `judge`). Give the provider only the request file's contents. Import its response once with `ingest "$work" --stage draft --job JOB --response FILE --elapsed SECONDS`. A terminal transport failure can be retried on the exact packet. Keep a complete, schema-shaped but invalid design as received; the evaluator scores it as a failure. Missing generation jobs prevent an aggregate result.

## Reading the scores

The simulator compares each policy against both established baselines under the same holdout trace. A successful policy must satisfy its resource limit, strictly beat the better baseline, and produce a distinct decision fingerprint on the diagnostic traces. The policy language permits a bounded set of mechanisms. A different fingerprint establishes different decisions from the reference policies on those probes.

`results.json` contains each outcome, the declared test's observed difference, a separate count of confirmed test predictions, invalid designs, usage and timings. It includes the paired labeled-minus-free difference in success rates and a 95% interval from 10,000 resamples of whole briefs. It reports incomplete judge coverage explicitly and computes judge means only after every review is present. Each blind packet mixes ten genuine answers and one weak-test control from the same ID pool. Judges score originality, feasibility, and test quality before seeing execution outcomes. Judge sessions and repeated designs share briefs and should not be treated as independent experiments.

The curation timer in the selection file covers one curator's time to read, choose, and label the cards, including coordination during that interval. It does not estimate human curation time. The run separately records index time, per-brief search and random-selection time, full request bytes, provider usage where available, and adapter wall time. This study allows no additional passage opens. The simulator can be rerun even when reviewers are unavailable.

```sh
python3 -m unittest discover -s examples/obsidian/creativity-study5
```

## Replay the November 27 run

The archived run contains the frozen protocol, public note packets, committed workload seeds, all 96 accepted structured model responses, and the simulator and reviewer scores. Rebuild the CLI and run the following from the repository root to check the packet contents, execute every prototype again, and verify the paired interval and ratings.

```sh
make build
python3 examples/obsidian/creativity-study5/replay.py --cx ./bin/cx
```

For a new live run, the supplied `subscription_adapter.py` supports reader `gpt-6.1-sol` (family `openai`) and judge `claude-opus-5-5` (family `anthropic`) through their existing CLI subscriptions. Both use high reasoning effort. Provider-reported usage applies to a four-task generation batch or one review packet. The scoring harness rejects responses exceeding its output limit. [November 27 findings](results.md)
