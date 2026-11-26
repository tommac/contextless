"""Run the synthetic vault experiment against the checkout in the current directory."""
import json
from pathlib import Path
import subprocess
import tempfile


def main():
    here = Path(__file__).resolve().parent
    root = Path.cwd()
    virtual = root / "internal/vault/physical_study_test.go"
    if not (virtual.parent / "vault.go").is_file():
        raise SystemExit("Run this from the repository root")
    with tempfile.TemporaryDirectory() as temp:
        overlay = Path(temp) / "overlay.json"
        overlay.write_text(json.dumps({"Replace": {str(virtual): str(here / "physical_study.go.txt")}}))
        subprocess.run(["go", "test", "-overlay", str(overlay), "-run", "^TestPhysicalStudyActualVault$",
                        "-count=1", "-timeout", "10m", "-v", "./internal/vault"], check=True)


if __name__ == "__main__":
    main()
