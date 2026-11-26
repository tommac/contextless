#!/usr/bin/env python3
"""Send reviewN.json packets to fresh cx sessions and collect ratingsN.json.

Usage: python3 run_reviewers.py <workdir> <workspace_prefix> <config.yaml>
Identical rubric to the study rounds; one fresh isolated workspace per judge.
"""
import json
import re
import subprocess
import sys
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
CX = ROOT / "bin/cx"

PROMPT = """You are scoring anonymized design answers. Each packet below groups candidate answers for one design problem. Score every design on three integer scales from 1 to 5.

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
    slice_ = text[text.find("{"):text.rfind("}") + 1]
    try:
        return json.loads(slice_)
    except json.JSONDecodeError:
        return json.loads(slice_.replace("\\n", "\n").replace("\\t", "\t"), strict=False)


def main():
    workdir, prefix, config = Path(sys.argv[1]), sys.argv[2], Path(sys.argv[3])
    for judge in range(3):
        out = workdir / f"ratings{judge}.json"
        if out.exists():
            print(f"skip {out}", flush=True)
            continue
        payload = PROMPT + (workdir / f"review{judge}.json").read_text()
        reply = None
        for suffix in ("", "b", "c"):
            name = f"{prefix}-j{judge}{suffix}"
            subprocess.run([str(CX), "new", name], capture_output=True, text=True)
            proc = subprocess.run([str(CX), "-config", str(config), "say", name, "-"],
                                  input=payload, capture_output=True, text=True, timeout=1800)
            if proc.returncode == 0:
                reply = proc.stdout
                break
            print(f"judge {judge} attempt {name} failed: {proc.stderr[-300:]}", flush=True)
            time.sleep(20)
        if reply is None:
            raise RuntimeError(f"judge {judge}: all attempts failed")
        (workdir / f"rawreview{judge}.txt").write_text(reply)
        ratings = {str(k): [int(x) for x in v] for k, v in extract_json(reply).items()}
        out.write_text(json.dumps(ratings, indent=2) + "\n")
        print(f"ratings{judge}.json ok n={len(ratings)}", flush=True)


if __name__ == "__main__":
    main()
