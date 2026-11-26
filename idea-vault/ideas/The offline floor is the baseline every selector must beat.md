---
type: idea
status: built
area: evaluation
needs: none
---
cx floor measures every baseline on the same questions with no model: whether the answer's source text is in the packet, how much ranked text comes before it, what the packets cost and how long they take to build. Its arms run from whole history through recent turns only and six whole turns to atoms at five budgets, each compared with the current default by wins, losses and an exact p value. A new selector has to beat this floor on the same questions before a model call is spent on it. The floor places the answer in the packet and does not show that a model uses it.

## Evidence
- `docs/CONTEXT-COMPILER.md`: "`cx floor` measures every baseline on the same questions with no model"
- `docs/CONTEXT-COMPILER.md`: "These probes place the answer in the packet. They do not show a model uses it, and none of this has been run with a model."

## Relations
- builds-on:: [[Recovery probes measure selection without a model]]
- builds-on:: [[Determinism is the instrument]]
- supports:: [[No model has yet answered from a packet]]
- tested-by:: [[Floor atoms at 8k held 107 of 120 answers at 25913 tokens]]
- tested-by:: [[Floor six whole turns held 103 of 120 answers]]
- tested-by:: [[Floor recent turns only held none of 120 answers]]
