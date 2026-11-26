"""Run the two offline controls against the pinned ContextBench log-triage release."""
import argparse
from collections import Counter, defaultdict
from hashlib import sha1
import json
from pathlib import Path
import re
import statistics
from time import perf_counter

# Git blob IDs from the task release at fb533c18c78a82d3f1b2047a9c1700735cf829ae.
BLOBS = (
    ("9377d1d2ab06a693ea6a1e6b1707a77e84b9c209", "2b75915a13e2c8e767afada5eb751b71a1d5c9f7"),
    ("b483cebd5155b3a70708000c84f80c6ce89ffefd", "5da8bfea17bd8477259f98d17008698678a91a56"),
    ("08006821967c6aa9561132bb0f968d41a2b999d3", "7fe0d702b1ae3248bbaf0d1643db14eacf4e7633"),
    ("dcbccc21b8c5ee82343b0dfadce4f58c6b5b35ee", "ccbb3d366f085499659499b7e54ed3b24a7572f3"),
    ("74309c124e53d92fe838e8d2b89fb7ee65c41159", "976571f2d9fe15c0a41be0d9f013826be923c4fa"),
    ("dc2a8509f6a0401e9e8c8e007a30a70d2b27e2a2", "12b0aa9c1836dcb7c122befa151f3912a0f7906c"),
    ("bb2147375894d873414d7ab2391094888958d714", "abe0b0fd396da063eab1d9132d2dbeb5bfa0a762"),
    ("fddcfae371577078f0b460bf9741ad92224617a4", "8d5dff251e54730a833a32537abfb842715cbe6c"),
)

LOG = re.compile(rb'^[\d\-:TZ]+ \[([A-Z]+)\] ([\w-]+) req=([0-9a-f]+) ([^\n]+)$', re.M)
LOOKUP = re.compile(r'QUERY (\d+): What is the message of the log line with req=([0-9a-f]+)\?')
COUNT = re.compile(r'QUERY (\d+): How many \[([A-Z]+)\] log lines are from service "([\w-]+)"\?')


def checked_bytes(path, expected):
    data = path.read_bytes()
    digest = sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
    if digest != expected:
        raise ValueError(f"unexpected Git blob for {path}: {digest}")
    return data


def parse_stream(stream):
    records = [json.loads(line) for line in stream.splitlines()]
    batches, queries = [], []
    for record in records:
        if record["header"].startswith("log batch") and not queries:
            batches.append(record["body"].encode())
        elif record["header"].startswith("query"):
            queries.append(record["body"].splitlines()[0])
        else:
            raise ValueError(f"unexpected record order or header: {record['header']}")
    return b"".join(batches), queries


def build_index(data):
    counts, addresses = Counter(), {}
    for match in LOG.finditer(data):
        level, service, request = match.group(1, 2, 3)
        counts[level, service] += 1
        if request in addresses:
            raise ValueError("duplicate request ID")
        addresses[request] = match.span(4)
    return counts, addresses


def solve(data, index, queries):
    counts, addresses = index
    answers, sliced_bytes = {}, 0
    for query in queries:
        if match := LOOKUP.fullmatch(query):
            start, end = addresses[match[2].encode()]
            answer = data[start:end].decode()
            sliced_bytes += end - start
        elif match := COUNT.fullmatch(query):
            answer = str(counts[match[2].encode(), match[3].encode()])
        else:
            raise ValueError(f"unknown prompt: {query}")
        answers[int(match[1])] = answer
    return answers, sliced_bytes


def reference(data, queries):
    answers, scan_bytes = {}, 0
    for query in queries:
        if match := LOOKUP.fullmatch(query):
            prefix = b" req=" + match[2].encode() + b" "
            start = data.index(prefix) + len(prefix)
            end = data.index(b"\n", start)
            answer = data[start:end].decode()
            scan_bytes += end + 1
        elif match := COUNT.fullmatch(query):
            prefix = b"[" + match[2].encode() + b"] " + match[3].encode() + b" req="
            answer = str(data.count(prefix))
            scan_bytes += len(data)
        else:
            raise ValueError(f"unknown prompt: {query}")
        answers[int(match[1])] = answer
    return answers, scan_bytes


def run(directory):
    rows, cases = [], []
    for seed, (stream_blob, gold_blob) in enumerate(BLOBS, start=42):
        stream = checked_bytes(directory / f"stream-{seed}.jsonl", stream_blob)
        data, questions = parse_stream(stream)
        index = build_index(data)
        indexed, sliced = solve(data, index, questions)
        scanned, logical = reference(data, questions)
        # Neither method sees the evaluator's answers while predicting.
        gold = json.loads(checked_bytes(directory / f"gold-{seed}.json", gold_blob))
        expected = {r["qid"]: str(r["answer"]) for r in gold["queries"]}
        if len(index[1]) != gold["meta"]["lines"] or indexed.keys() != expected.keys() or scanned.keys() != expected.keys():
            raise ValueError(f"incomplete coverage on seed {seed}")
        rows.append(dict(seed=seed, lines=len(index[1]), counter_keys=len(index[0]),
                         raw_bytes=len(data), source_reaccess=sliced, reference_span_bytes=logical,
                         exact_index=sum(indexed[k] == v for k, v in expected.items()),
                         exact_scan=sum(scanned[k] == v for k, v in expected.items()),
                         questions=len(expected), approx_tokens=gold["meta"]["approx_log_tokens"]))
        cases.append((data, questions))

    timings = defaultdict(list)
    for repetition in range(5):
        for mode in (("index", "repeated") if repetition % 2 == 0 else ("repeated", "index")):
            start = perf_counter()
            for data, questions in cases:
                if mode == "index":
                    solve(data, build_index(data), questions)
                else:
                    reference(data, questions)
            timings[mode].append(perf_counter() - start)
    return {"seeds": rows, "median_seconds": {k: statistics.median(v) for k, v in timings.items()}}


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("directory", type=Path, help="directory holding stream-{seed}.jsonl and gold-{seed}.json")
    print(json.dumps(run(parser.parse_args().directory), indent=2))
