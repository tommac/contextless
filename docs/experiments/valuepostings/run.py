"""Run the value-postings experiment as an overlay in the production packet package."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[2]
PUBLIC_SHA = "d6f21ea9d60a0d56f34a05b609c79c88a451d2ae03597821ea3d5a9678c3a442"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--sizes", default="1000,10000,50000", help="synthetic atom counts, comma-separated")
    parser.add_argument("--public-data", type=Path, help="optionally validate all 500 pinned cleaned LongMemEval-S questions")
    args = parser.parse_args()
    sizes = [int(n) for n in args.sizes.split(",")]
    if not sizes or min(sizes) <= 0:
        parser.error("sizes must be positive")
    env = dict(os.environ, CX_VALUEPOSTINGS_SIZES=",".join(str(n) for n in sizes))
    env.pop("CX_VALUEPOSTINGS_PUBLIC", None)
    if args.public_data:
        digest = hashlib.sha256()
        with args.public_data.open("rb") as data:
            for chunk in iter(lambda: data.read(1 << 20), b""):
                digest.update(chunk)
        if digest.hexdigest() != PUBLIC_SHA:
            parser.error("expected the pinned public LongMemEval-S release")
        env["CX_VALUEPOSTINGS_PUBLIC"] = str(args.public_data.resolve())
    with tempfile.TemporaryDirectory() as tmp:
        overlay = Path(tmp) / "overlay.json"
        virtual = ROOT / "internal/packet/value_postings_experiment_test.go"
        overlay.write_text(json.dumps({"Replace": {str(virtual): str(HERE / "experiment.go.txt")}}))
        subprocess.run(["go", "test", "-overlay", str(overlay), "-run", "^TestValuePostings", "-count=1",
                        "-timeout", "30m", "-v", "./internal/packet"], cwd=ROOT, env=env, check=True)


if __name__ == "__main__":
    main()
