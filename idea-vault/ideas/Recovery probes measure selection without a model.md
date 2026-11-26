---
type: idea
status: built
area: evaluation
needs: none
---
A probe is a question with a known source turn. The cx probe command, and a test that plants 14 facts in an 800-turn conversation, report per fact type where each source turn lands: in the packet verbatim, named in it, or found by a search. No model runs. The idea is LANTERN's evaluation without its LLM judge. It was built first, so every later Stage 2b change was kept only if the probes or a test showed it working.

## Evidence
- `docs/ROADMAP.md`: "`cx probe` and `TestProbesRecoverPlantedFacts` report per fact type where a probe's source turn sits in the packet. Measured first, so every change below was kept only if the probes or a test showed it working."
- `docs/ROADMAP.md`: "LANTERN's evaluation, without its LLM judge"
- `docs/ARCHITECTURE.md`: "`cx probe <ws> probes.jsonl` reports per fact type whether the source is in the packet verbatim, named by its own line, only inside a digest range, or missing"
- `docs/ARCHITECTURE.md`: "synthetic 800-turn conversation in `probe_test.go` (about 620k tokens,"

## Relations
- part-of:: [[Determinism is the instrument]]
- builds-on:: [[Each call gets a fresh packet built by code]]
- source:: [[LANTERN]]
- tested-by:: [[Ranked turns at tier 3 raised planted facts verbatim from 3 to 11 of 14]]
