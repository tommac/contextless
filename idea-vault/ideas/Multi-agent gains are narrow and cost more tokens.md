---
type: idea
status: external
area: plans
needs: model-calls
---
The measured gains of parallel agents sit in a narrow shape: decomposable analysis, independent tool calls, and inputs longer than the model's usable window. Outside it the mean effect was slightly negative, token cost ran 1.6 to 6.2 times a single agent, and a single agent usually ties or wins when budgets are matched. Much of the reported gain is extra tokens spread over separate context windows, and the margin shrinks as base models improve. None of this has been measured in ContextlessAI: no local evidence exists yet that plans help.

## Evidence
- `reports/Parallel AI agents data structures.md`: "Most of the celebrated multi-agent gains turn out to be more tokens in more context windows, which a single well-budgeted call often matches, so the honest default is one agent"
- `reports/Parallel AI agents data structures.md`: "The benefit is the workflow and the budget, less the separation into agents."
- `reports/Parallel AI agents data structures.md`: "No local evidence exists yet that plans help; the external mean is -3.5%"
- `research_notes/Parallel AI agents data structures/evidence_and_failures.md`: "MetaGPT on HumanEval 67.0% single to 87.7% multi with ChatGPT, against 90.2% to 93.2% with Gemini-2.0-Flash"

## Relations
- tested-by:: [[Multi-agent coordination averaged minus 3.5 percent over 260 configurations]]
- tested-by:: [[Centralized agents gained 80.8 percent on decomposable financial analysis]]
- tested-by:: [[Multi-agent token cost ran 1.6 to 6.2 times a single agent]]
- tested-by:: [[A single agent scored 0.418 against 0.369 to 0.388 at equal thinking budget]]
- tested-by:: [[Token usage alone explained 80 percent of BrowseComp variance]]
- source:: [[Towards a Science of Scaling Agent Systems]]
- source:: [[Anthropic multi-agent research system]]
