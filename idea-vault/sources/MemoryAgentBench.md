---
type: source
status: external
area: evaluation
needs: none
year: 2025
url: https://arxiv.org/html/2507.05257
---
MemoryAgentBench is a third-party benchmark of 2,071 questions over contexts of about 103K to 1.44M tokens, testing accurate retrieval, test-time learning, long-range understanding and selective forgetting. It includes full-context runs on six long-context models, and its venue is not confirmed on the paper page. It is the report's main source for BM25 as a hard floor against graph and memory systems.

## Evidence
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "2,071 questions, about 103K to 1.44M tokens of context"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Accurate retrieval, test-time learning, long-range understanding, selective forgetting"
- `research_notes/Cutting edge context memory research/evaluation_methods.md`: "Mixes reused datasets; venue not confirmed on the paper page"

## Relations
- supports:: [[Knowledge graphs lose to BM25 as a default store]]
- alternative-to:: [[LongMemEval]]
- tested-by:: [[BM25 tied HippoRAG-v2 overall and beat GraphRAG and Zep on retrieval]]
