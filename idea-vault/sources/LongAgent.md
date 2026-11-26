---
type: source
status: external
area: plans
needs: none
year: 2024
url: https://arxiv.org/html/2402.11550v2
---
A Fudan system (February 2024) in which a leader splits a query over dozens to hundreds of members that each read one fixed-size chunk. When members' answers disagree, the leader has them exchange chunks pairwise and answer again, and members that had hallucinated mostly corrected themselves. It beat GPT-4 with full context on needle tasks, and training members to refuse raised correct refusals. The results are the paper's own, on a fine-tuned 7B model.

## Evidence
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "the leader identifies members whose answers disagree; those members exchange chunks pairwise and answer again"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "single-document 81.53% vs 62.00%; multi-document 55.33% vs 50.37%"
- `research_notes/Parallel AI agents data structures/mapreduce_and_shared_memory.md`: "Unit: fixed-size chunk per member (4k window), dozens to hundreds of members sharing one model."

## Relations
- supports:: [[Disagreeing units are rerun with each other's evidence pinned]]
- supports:: [[No information is a result not a failure]]
- supports:: [[Multi-agent gains are narrow and cost more tokens]]
