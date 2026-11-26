"""Replay the public Go 1.26 API additions against questions saved at Go 1.25."""

import argparse
from collections import defaultdict
from dataclasses import dataclass
import hashlib
import json
from pathlib import Path
import re


DIGESTS = {
    24: "5d717afb14c44034ec8e6a3d12e1a1abd9786ad64aecf161b85ccad976c440dd",
    25: "8a0ef361fd1ad415078f440cbe155dc7a1c7e6a2ad83f97f93527f2c0530755d",
    26: "3b3fd1a5bfbc6b3ab3811dd4a0fb3930d29590c76d7203051a35e4b45360198b",
}


@dataclass(frozen=True)
class Record:
    version: int
    line: int
    package: str
    declaration: str
    original: str


def release(directory, version):
    path = directory / f"go1.{version}.txt"
    raw = path.read_bytes()
    if hashlib.sha256(raw).hexdigest() != DIGESTS[version]:
        raise ValueError(f"unexpected contents of {path}")
    records = []
    for line, original in enumerate(raw.decode("utf-8").splitlines(), start=1):
        match = re.fullmatch(r"pkg ([^,]+), (.+) #\d+", original)
        if not match:
            raise ValueError(f"unrecognized API declaration on {path}:{line}")
        records.append(Record(version, line, *match.groups(), original))
    return records


@dataclass(frozen=True)
class Question:
    package: str
    declaration: str  # Empty for a count of distinct declarations in the package.
    through: int      # Zero for a question that tracks the latest release.


def replay_all(records, questions):
    """Independent reference: evaluate each predicate over the specified releases."""
    answers = {}
    for q in questions:
        matches = {r.declaration for r in records
                   if r.package == q.package and (not q.through or r.version <= q.through)
                   and (not q.declaration or r.declaration == q.declaration)}
        answers[q] = len(matches)
    return answers


class WatchedAnswers:
    def __init__(self, prior, old_records):
        self.by_predicate = defaultdict(list)
        for q in prior:
            self.by_predicate[q.package, q.declaration].append(q)
        self.known = {(r.package, r.declaration) for r in old_records}
        self.answers = prior.copy()

    def advance(self, new_records):
        """Apply chronologically appended release records, retaining their source locations."""
        witnesses = defaultdict(list)
        for r in new_records:
            signature = (r.package, r.declaration)
            if signature in self.known:
                continue
            self.known.add(signature)
            for key in (signature, (r.package, "")):
                for q in self.by_predicate.get(key, ()):
                    if not q.through or r.version <= q.through:
                        self.answers[q] += 1
                        witnesses[q].append(r)
        return witnesses


def run(directory, list_changes=False):
    old = release(directory, 24) + release(directory, 25)
    delta = release(directory, 26)
    archive = old + delta
    # A retrospective workload covers every actual declaration and package.
    # Only the workload builder looks ahead to choose which questions to enroll.
    keys = {(r.package, r.declaration) for r in archive} | {(r.package, "") for r in archive}
    questions = [Question(package, decl, cutoff) for package, decl in sorted(keys)
                 for cutoff in (0, 25)]
    prior = replay_all(old, questions)
    watch = WatchedAnswers(prior, old)
    witnesses = watch.advance(delta)
    predicted = watch.answers
    gold = replay_all(archive, questions)
    changed = {q for q in questions if gold[q] != prior[q]}
    if predicted != gold or set(witnesses) != changed:
        raise AssertionError("updated values or the set of affected questions disagreed with full replay")
    repeats = watch.advance(delta)
    if watch.answers != gold or repeats:
        raise AssertionError("replaying the same release changed an answer")

    packages = {r.package for r in delta}
    coarse = {q for q in questions if q.package in packages and not q.through}

    def score(reopened):
        return {"reopened": len(reopened), "changed_missed": len(changed - reopened),
                "unchanged_reopened": len(reopened - changed)}

    peek = Question("bytes", "method (*Buffer) Peek(int) ([]uint8, error)", 0)
    pinned = Question(peek.package, peek.declaration, 25)
    witness, = witnesses[peek]
    report = {
        "saved_questions": len(questions), "old_records": len(old), "new_records": len(delta),
        "answers_changed": len(changed), "answers_unchanged": len(questions) - len(changed),
        "changed_presence_answers": sum(bool(q.declaration) for q in changed),
        "changed_package_counts": sum(not q.declaration for q in changed),
        "strategies": {"all": score(set(questions)), "changed_cited_bytes": score(set()),
                       "current_answers_in_changed_packages": score(coarse),
                       "matching_saved_predicates": score(set(witnesses))},
        "bytes.Buffer.Peek": {"previous": bool(prior[peek]), "current": bool(predicted[peek]),
                              "as_of_1.25": bool(predicted[pinned]),
                              "source": f"go1.{witness.version}.txt:{witness.line}",
                              "original": witness.original},
    }
    if list_changes:
        report["changes"] = [
            {"package": q.package, "declaration": q.declaration or "(package count)",
             "previous": prior[q], "current": predicted[q],
             "evidence": [{"source": f"go1.{r.version}.txt:{r.line}", "original": r.original}
                          for r in witnesses[q]]}
            for q in questions if q in changed
        ]
    return report


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("api", type=Path, help="directory containing the three pinned Go API ledgers")
    parser.add_argument("--list-changes", action="store_true", help="include the evidence for every revised answer")
    args = parser.parse_args()
    print(json.dumps(run(args.api, args.list_changes), indent=2))
