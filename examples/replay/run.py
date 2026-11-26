"""From a checkout run python3 examples/replay/run.py. All workspaces are temporary."""

import json
from pathlib import Path
import subprocess
import tempfile

REPO = Path(__file__).resolve().parents[2]
HISTORY = [
    "Book Cedar Hall for the autumn meeting.",
    "I have Cedar Hall on the draft agenda.",
    "The hall is under repair. Move the meeting to Harbor Room instead.",
    "I have updated the invitation to Harbor Room.",
    "Final correction from the coordinator: meet in Maple Library. The time stays the same.",
]
QUESTION = "Which venue did we settle on, and what two venues did it replace?"


def main():
    with tempfile.TemporaryDirectory() as d:
        tmp = Path(d)
        exe = tmp / "cx"
        subprocess.run(["go", "build", "-o", exe, "./cmd/cx"], cwd=REPO, check=True)
        cfg = tmp / "config.yaml"
        cfg.write_text("root: " + json.dumps(str(tmp / "state")))

        def cx(*args):
            return subprocess.check_output([exe, "-config", cfg, *args], text=True)

        (tmp / "chat.json").write_text(
            json.dumps(
                [
                    dict(
                        role="user" if i % 2 == 0 else "assistant",
                        content=body,
                        timestamp="2026-01-01T00:00:00Z",
                    )
                    for i, body in enumerate(HISTORY)
                ]
            )
        )
        cx("new", "example")
        cx("import", "example", tmp / "chat.json")
        notes = tmp / "notes"
        notes.mkdir()
        changes = [(1, "Cedar Hall"), (3, "Harbor Room"), (5, "Maple Library")]
        sources = [dict(turn=f"t{i:06}", answer=venue) for i, venue in changes]
        (notes / "example-demo.jsonl").write_text(
            "\n".join(
                json.dumps(
                    dict(
                        title=venue,
                        evidence=[dict(turn=f"t{i:06}", quote=HISTORY[i - 1])],
                    )
                )
                for i, venue in changes
            )
        )
        (tmp / "question.jsonl").write_text(
            json.dumps(dict(id="venues", question=QUESTION, sources=sources))
        )
        dump = tmp / "packets"
        cx(
            "zoom",
            f"example={tmp / 'question.jsonl'}",
            "-notes",
            notes,
            "-dump",
            dump,
            "-dump-only",
            "-dump-total",
            "1000",
        )
        (row,) = json.loads((dump / "index.json").read_text())
        catalog = row["packets"]["catalog"]
        print((dump / "p" / catalog / "packet.md").read_text())
        # Two fixed, disclosed requests show exactly what evidence is available after each.
        for run, ids in [
            ("stale", ["t000003#a1"]),
            ("timeline", [f"t{i:06}#a1" for i, _ in changes]),
        ]:
            cx("pull", dump, run, catalog, *ids)
            base = dump / "pulls" / run / catalog
            log = json.loads(base.with_suffix(".json").read_text())
            excerpt = base.with_suffix(".md").read_text()
            assert log["given"] == ids and all(
                HISTORY[int(a[1:7]) - 1] in excerpt for a in ids
            )
            latest = HISTORY[-1] in excerpt
            assert latest == (run == "timeline")
            print(
                f"\n{run}: {log['tokens']} estimated excerpt tokens; catalog {row['tokens']['catalog']} estimated tokens"
            )
            print(excerpt)
            print(
                "Final venue supported by opened evidence:",
                "Maple Library" if latest else "unresolved",
            )


if __name__ == "__main__":
    main()
