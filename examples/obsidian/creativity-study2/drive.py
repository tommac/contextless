#!/usr/bin/env python3
"""Drive the round-2 creativity study sessions through cx workspaces.

Usage:
  python3 drive.py readers          # 8 fresh reader sessions -> outN.json
  python3 drive.py reviewers        # 3 fresh reviewer sessions -> ratingsN.json

Every session is a brand-new cx workspace, so nothing carries across runs.
"""
import json
import re
import subprocess
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
STUDY = ROOT / "examples/obsidian/creativity-study2"
RUN = STUDY / "run"
CFG = STUDY / "study-config.yaml"
CX = ROOT / "bin/cx"

REVIEWER_PROMPT = """You are scoring anonymized design answers. Each packet below groups candidate answers for one design problem. Score every design on three integer scales from 1 to 5.

Originality: 1 generic familiar pattern; 3 a meaningful adaptation; 5 a surprising, intelligible mechanism distinctive even among these designs.
Feasibility: 1 breaks constraints; 3 plausible with gaps; 5 a workable concrete plan within stated resources.
Discriminating test: 1 uninformative or absent; 3 names a comparison with a measurable result; 5 isolates the mechanism with a fair simple baseline and an outcome that could disfavor it.

Judge each design independently, only against the problem it answers. Return only one JSON object mapping each design id to an array of three integers: [originality, feasibility, test].

"""


def extract_json(text):
    text = text.strip()
    if text.startswith("```"):
        text = re.sub(r"^```[a-zA-Z]*\s*", "", text)
        text = re.sub(r"\s*```\s*$", "", text)
    start, end = text.find("{"), text.rfind("}")
    if start == -1 or end <= start:
        raise ValueError("no JSON object in reply")
    slice_ = text[start:end + 1]
    try:
        return json.loads(slice_)
    except json.JSONDecodeError:
        # Some replies escape structural newlines as literal \n outside strings.
        relaxed = slice_.replace("\\n", "\n").replace("\\t", "\t")
        return json.loads(relaxed, strict=False)


def new_workspace(name):
    subprocess.run([str(CX), "new", name], capture_output=True, text=True)


def say(name, payload):
    proc = subprocess.run([str(CX), "-config", str(CFG), "say", name, "-"],
                          input=payload, capture_output=True, text=True, timeout=1800)
    if proc.returncode != 0:
        raise RuntimeError(f"cx say failed for {name}: {proc.stderr[-800:]}")
    return proc.stdout


def readers():
    for run in range(8):
        out = RUN / f"out{run}.json"
        if out.exists():
            print(f"skip out{run}.json (exists)", flush=True)
            continue
        name = f"cr2-r{run}"
        raw = RUN / f"raw{run}.txt"
        if raw.exists():
            reply = raw.read_text()
            print(f"reusing raw{run}.txt", flush=True)
        else:
            payload = (RUN / f"input{run}.json").read_text()
            started = time.time()
            reply = None
            for suffix in ("", "b", "c"):
                candidate = name + suffix
                new_workspace(candidate)
                try:
                    reply = say(candidate, payload)
                    break
                except RuntimeError as exc:
                    print(f"run {run} attempt {candidate} failed: {exc}", flush=True)
                    time.sleep(20)
            if reply is None:
                raise RuntimeError(f"run {run}: all attempts failed")
            raw.write_text(reply)
            print(f"run {run} answered in {time.time()-started:.0f}s", flush=True)
        answers = extract_json(reply)
        if set(answers) != {"A", "B", "C", "D"}:
            raise ValueError(f"run {run}: bad keys {sorted(answers)}")
        out.write_text(json.dumps(answers, ensure_ascii=False, indent=2))
        words = {k: len(v.split()) for k, v in answers.items()}
        print(f"out{run}.json ok words={words}", flush=True)


def reviewers():
    for judge in range(3):
        out = RUN / f"ratings{judge}.json"
        if out.exists():
            print(f"skip ratings{judge}.json (exists)", flush=True)
            continue
        name = f"cr2-j{judge}"
        payload = REVIEWER_PROMPT + (RUN / f"review{judge}.json").read_text()
        new_workspace(name)
        started = time.time()
        reply = say(name, payload)
        (RUN / f"rawreview{judge}.txt").write_text(reply)
        ratings = extract_json(reply)
        ratings = {str(k): [int(x) for x in v] for k, v in ratings.items()}
        out.write_text(json.dumps(ratings, indent=2) + "\n")
        print(f"ratings{judge}.json ok in {time.time()-started:.0f}s n={len(ratings)}", flush=True)


if __name__ == "__main__":
    mode = sys.argv[1]
    RUN.mkdir(exist_ok=True)
    {"readers": readers, "reviewers": reviewers}[mode]()
