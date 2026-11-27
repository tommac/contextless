"""Replay the recorded prototypes and verify the published ratings without model calls."""
import argparse
import gzip
import json
from pathlib import Path
import subprocess
import sys
import tempfile

import study


def replay(bundle, cx, work):
    with gzip.open(bundle, "rt", encoding="utf-8") as f:
        data = json.load(f)
    for name in ("manifest", "cards", "selection-audit"):
        study.save(work / (name + ".json"), data[name])
    study.save(work / "private/holdouts.json", data["holdouts"])
    for job, response in data["responses"].items():
        if not job.replace("-", "").isalnum():
            raise ValueError("invalid job name")
        study.save(work / "responses" / (job + ".json"), response)
    for i, attempt in enumerate(data["attempts"]):
        job = attempt["job"]
        if job not in data["responses"]:
            raise ValueError("unknown job in attempt log")
        study.save(work / "attempts" / job / str(i) / "meta.json", attempt["metadata"])
    for stage in ("draft", "revise", "judge"):
        for job, (packet, _) in study.jobs_for(work, stage).items():
            if packet != data["requests"][job]:
                raise ValueError("archived packet differs at " + job)
    subprocess.run([sys.executable, str(Path(study.__file__).resolve()), "score",
                    str(work), "--cx", str(cx.resolve())], check=True, stdout=subprocess.DEVNULL)
    result = study.read(work / "results.json")
    expected = data["results"]
    for field in ("arms", "primary", "holdout_commitment", "rows", "review", "usage"):
        if result[field] != expected[field]:
            raise ValueError("recorded result differs at " + field)
    return result


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--cx", type=Path, required=True)
    parser.add_argument("--bundle", type=Path,
                        default=Path(__file__).with_name("2026-11-27-run.json.gz"))
    args = parser.parse_args()
    with tempfile.TemporaryDirectory() as tmp:
        result = replay(args.bundle, args.cx, Path(tmp))
    print("Verified all %d prototypes and %d blind review packets." %
          (len(result["rows"]), result["review"]["jobs_complete"]))
    print(json.dumps(result["primary"], indent=2))
