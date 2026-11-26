---
type: idea
status: rejected
area: plans
needs: none
---
A shared board that all agents may read and write decouples them and allows opportunistic contribution. With LLM agents its measured gain is small and needs extra resolver and cleaner roles, and the strongest blackboard result isolates helpers from each other with one reader, which makes it hub and spoke with self-selection. The survey's verdict is to avoid the open form. For ContextlessAI it says to give units no shared writable surface at all: no blackboard, no shared scratch file, no unit-to-unit mailbox or chat.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Avoid the open form. Shared mutable coordination state failed at Cursor"
- `reports/Parallel AI agents data structures.md`: "Do not give units a shared writable surface: no blackboard, no shared scratch file, no unit-to-unit mailbox or chat."
- `reports/Parallel AI agents data structures.md`: "Measured small gain (81.68 against 80.09 for chain-of-thought) with extra resolver and cleaner roles"

## Relations
- analog-of:: [[Blackboard architecture]]
- analog-of:: [[A shared writable memory for units has no support]]
- alternative-to:: [[One writer per surface]]
- alternative-to:: [[A task list with dependency placeholders resolved by code]]
- tested-by:: [[Full consensus fell from 89.8 percent on MMLU to 29.4 on MATH]]
- tested-by:: [[Twenty lock-coordinated agents ran at the throughput of one to three]]
