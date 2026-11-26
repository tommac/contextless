# Contributing

Thank you for looking. This is a research system with one maintainer, so
small, well-evidenced changes are the easiest to take.

## Build and test

```sh
go build ./...
go vet ./...
go test ./... -count=1     # scripted providers, no model calls, no keys
make demo                  # model-free walk through the engine
```

Go 1.26 or later, on macOS or Linux. Every change to the engine needs a test
that fails without it.

## What makes a change easy to accept

- **Say what you measured.** If a change is meant to improve retrieval or
  answers, write the pass rule before the run, report the result that went
  against you as plainly as the one that went for you, and separate "the
  source text reached the packet" from "a model answered better".
- **Keep the invariants.** A turn is stored verbatim and never rewritten.
  Model-written records (cards, notes, facts) point at text; they never
  stand in for it. Anything a model quotes is checked against the source by
  code.
- **No private data.** Do not commit transcripts, probe or question files
  made from private conversations, keys, or paths from your machine.
  Benchmarks are downloaded by script from their own repositories.
- **Docs in plain sentences.** Report numbers with their sample size and
  what kind of evidence they are.

## Idea vault

`idea-vault/` holds one note per idea or result. Each evidence line quotes
the file it cites, and `python3 docs/experiments/vault_graph.py .` checks
that every quote is still there. Prefer quoting `docs/` over the README.

## License

By contributing you agree that your contribution is licensed under the
Apache License, Version 2.0, as in `LICENSE`.
