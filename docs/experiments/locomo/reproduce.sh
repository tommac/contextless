#!/bin/sh
# Invoke from the ContextlessAI checkout root with the adjacent adapter.py.
set -eu
adapter=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/adapter.py
work=${1:-$(mktemp -d)}
mkdir -p "$work/module/docs/experiments/longmem"
release=3eb6f2c585f5e1699204e3c3bdf7adc5c28cb376
repo=https://raw.githubusercontent.com/snap-research/locomo
curl -fsSL --retry 2 "$repo/$release/data/locomo10.json" -o "$work/locomo10.json"
git archive 04e171c3a8b80160179ceb84042c23eda2d7d7cb go.mod go.sum internal | tar -x -C "$work/module"
longmem=docs/experiments/longmem
python3 "$adapter" prepare "$work/locomo10.json" "$longmem/main.go" "$work/input.json" "$work/module/$longmem/main.go"
work=$(cd "$work" && pwd)
(cd "$work/module" && go run "./$longmem" -data "$work/input.json" > "$work/ranks.jsonl")
python3 "$adapter" score "$work/locomo10.json" "$longmem" "$work/ranks.jsonl" "$work/cases.json" "$work/results.json"
echo "Saved $work/results.json and $work/cases.json"
