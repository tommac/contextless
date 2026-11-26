# Brainstorm with an Obsidian Canvas

Open `idea-vault` as a vault in Obsidian and select **Idea workshop**. The board asks how an agent should spend its last read when two passages conflict. It connects the question to two database ideas and one measured retrieval result. The labeled lines tell the reader how you want to use each note.

For creative design, open **One postcard science exhibit** and run
`python3 examples/obsidian/brief.py --creative 'idea-vault/One postcard science exhibit.canvas'`.
This asks for two designs and a test using short excerpts from your chosen notes.
Across three blind rounds ([1](creativity-study/README.md),
[2](creativity-study2/README.md), [3](creativity-study3/README.md)) with 96
scored designs, chosen notes beat ordinary search and the question alone on
originality, and labels on those notes consistently sharpened the proposed
test. The question alone still won individual problems, so for open briefs
keep a `--question-only` draft in contention.

## Try your own question

1. Duplicate the Canvas. Double-click the **Question** card and replace its body. Keep `# Question` on the first line.
2. Drag a Markdown note from the file explorer onto the board. Drag from the edge of the question card to the note. Double-click the line to label it. For example, connect an idea with **Borrow this mechanism** and a failed experiment with **What would this rule out?**
3. For a long note, right-click its card and choose **Narrow to heading**. The reader gets that section, including its subheadings. Use a unique Markdown heading written with `#` characters. You can also connect a text card you wrote on the board.
4. From the repository root, preview the reader's input with Python 3.9 or newer. Use your copy's filename when trying your own board.

```sh
python3 examples/obsidian/brief.py 'idea-vault/Idea workshop.canvas'
```

Each run reads the saved board and its directly connected cards. Relabel a line and the next prompt carries the new instruction alongside the same source. Disconnect a card and its text drops out. Change a source note and the next prompt includes the edit. For a board in a subfolder, pass `--vault idea-vault` so file links resolve from the vault root.

You can paste the preview into an AI chat. On macOS, append `| pbcopy` to the preview command to copy it. To send it through ContextlessAI, use the name of an existing workspace with a configured reader in place of `mychat`. This records the brief as a turn in that workspace.

```sh
brief=$(python3 examples/obsidian/brief.py 'idea-vault/Idea workshop.canvas') &&
  printf '%s\n' "$brief" | bin/cx say mychat -
```

## What changed in one trial

Four fresh assistant contexts received the same question and instructions to propose two ideas, work through one example and give a cheap test.

| Input | How the assistant spent the extra read |
| --- | --- |
| Question only | Seek the decision-maker's reply or trace a quotation to its original. In its example, a vote could confirm Thursday, preserve Tuesday or defer the decision. |
| Question and cards | Seek a decision or a condition under which both dates apply. In its example, “Thursday replaces Tuesday for everyone” resolves the conflict. “Thursday is for instructors” leaves the attendee date open. |
| Same cards with labels removed | Seek the reply on which the two predictions disagree, or open the continuation for a qualification. It suggested that two dates might apply to different offices. |
| Top three ordinary search results | Read the reply to the extension request. The owner may have approved Monday for the appendix while keeping Friday for the main report. |

Use `--question-only` and `--without-labels` to recreate the first and third inputs. All three versions with notes considered when conflicting answers could each be valid. This trial cannot establish a creativity advantage for Canvas or its labels. [Read the four answers and how the search control worked](trial.md).
