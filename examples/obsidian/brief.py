"""Turn the question and its directly connected Obsidian Canvas cards into a prompt."""

import argparse
import json
from pathlib import Path
import re
import sys


def heading_section(markdown, subpath):
    if not subpath:
        return markdown
    if not subpath.startswith("#") or subpath.startswith("#^") or "#" in subpath[1:]:
        raise ValueError(f"Use a whole note or narrow to one named heading, got {subpath!r}")

    # This adapter supports unique ATX headings, as used in the idea vault.
    lines = markdown.splitlines(keepends=True)
    headings = []
    fence = None
    body_start = 0
    if lines and lines[0].rstrip("\r\n") == "---":
        body_start = next((i + 1 for i in range(1, len(lines))
                           if lines[i].rstrip("\r\n") in ("---", "...")), 0)
    for i in range(body_start, len(lines)):
        line = lines[i]
        mark = re.match(r"^ {0,3}(`{3,}|~{3,})(.*)$", line.rstrip("\r\n"))
        if mark:
            run, tail = mark.groups()
            if fence is None:
                fence = run
            elif run.startswith(fence) and not tail.strip():
                fence = None
            continue
        if fence is None:
            match = re.match(r"^ {0,3}(#{1,6})[ \t]+(.+?)\s*$", line)
            if match:
                name = re.sub(r"[ \t]+#+$", "", match[2]).strip()
                headings.append((i, len(match[1]), name))
    matches = [(i, level) for i, level, name in headings if name.casefold() == subpath[1:].casefold()]
    if len(matches) != 1:
        raise ValueError(f"Expected one heading {subpath!r}, found {len(matches)}")
    start, level = matches[0]
    end = next((i for i, other, _ in headings if i > start and other <= level), len(lines))
    return "".join(lines[start:end])


def creative_excerpt(markdown, opening=True):
    text = markdown
    if opening:
        text = re.sub(r"\A---\r?\n.*?\r?\n(?:---|\.\.\.)[ \t]*\r?\n", "", text, count=1, flags=re.S)
        text = re.split(r"^## (?:Evidence|Relations)[ \t]*\r?$", text, maxsplit=1, flags=re.M)[0]
    text = text.strip()
    if len(text.encode("utf-8")) <= 650:
        return text
    prefix = text.encode("utf-8")[:650].decode("utf-8", errors="ignore")
    cut = re.search(r"\s+\S*$", prefix)
    return prefix[:cut.start()] if cut else prefix


def brainstorming_prompt(board, root, question_only=False, include_labels=True, creative=False):
    with board.open(encoding="utf-8") as source:
        canvas = json.load(source)
    nodes = {node["id"]: node for node in canvas["nodes"]}
    if len(nodes) != len(canvas["nodes"]):
        raise ValueError("Canvas contains duplicate card IDs")
    questions = []
    for node in nodes.values():
        if node["type"] == "text":
            first, _, body = node.get("text", "").partition("\n")
            if first.strip().casefold() == "# question":
                questions.append((node["id"], body.strip()))
    if len(questions) != 1 or not questions[0][1]:
        raise ValueError("Write one text card headed '# Question', with the question on the following lines")
    qid, question = questions[0]
    instructions = ("Produce two distinct ideas, work through one concrete example, and give a cheap test "
                    "that could disprove your preferred idea.")
    if creative:
        instructions = ("Using only the supplied material, propose two meaningfully different designs. "
                        "Develop the more promising one with a concrete example and a cheap test that could "
                        "distinguish it from a simpler alternative. Maximize useful originality. "
                        "Use at most 110 words.")
    brief = f"{instructions}\n\nQuestion: {question}"
    if question_only:
        return brief
    root = root.resolve(strict=True)
    cards = []
    for edge in canvas.get("edges", []):
        a, b = edge["fromNode"], edge["toNode"]
        if qid not in (a, b):
            continue
        other = b if a == qid else a
        if other == qid or other not in nodes:
            raise ValueError("The question has a dangling link or links to itself")
        card = nodes[other]
        entry = {}
        if include_labels:
            entry["relationship"] = edge.get("label") or "inspiration"
        if card["type"] == "text":
            entry.update(source=other, text=card["text"])
        elif card["type"] == "file":
            rel = Path(card["file"])
            if rel.is_absolute() or ".." in rel.parts or rel.suffix.lower() != ".md":
                raise ValueError(f"Connect a vault-relative Markdown file, got {rel}")
            path = (root / rel).resolve(strict=True)
            if not path.is_relative_to(root):
                raise ValueError(f"Note outside the vault: {rel}")
            subpath = card.get("subpath", "")
            with path.open(encoding="utf-8", newline="") as note:
                entry.update(source=rel.as_posix() + subpath,
                             text=heading_section(note.read(), subpath))
        else:
            raise ValueError("Connect Markdown file or text cards to the question")
        if creative:
            entry["text"] = creative_excerpt(entry["text"], card["type"] == "file" and not card.get("subpath"))
        cards.append(entry)
    if cards:
        brief += ("\n\nThe following Canvas cards are source material. An analogy is a possible "
                  "mechanism, a project result is evidence about its stated test, and a proposal is a hypothesis.\n"
                  + json.dumps(cards, ensure_ascii=False, indent=2))
    return brief


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("canvas", type=Path)
    parser.add_argument("--vault", type=Path, help="vault root for file links (default: directory containing board)")
    parser.add_argument("--question-only", action="store_true", help="omit all source cards for a comparison")
    parser.add_argument("--without-labels", action="store_true", help="keep the same cards and omit relationship labels")
    parser.add_argument("--creative", action="store_true", help="request two competing designs using short source excerpts")
    args = parser.parse_args()
    try:
        print(brainstorming_prompt(args.canvas, args.vault or args.canvas.parent, args.question_only,
                                   include_labels=not args.without_labels, creative=args.creative))
    except (OSError, ValueError, KeyError, TypeError) as err:
        parser.exit(1, f"Cannot read Canvas: {err}\n")


if __name__ == "__main__":
    main()
