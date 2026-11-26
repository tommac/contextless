# Context Compression and Prompt Caching for LLM Applications and Agents

Research date: 2026-11-18. Provider pages carry no version date unless one is given, so read "as fetched" as 2026-11-18.

Evidence labels: **[Measured]** = experiment reported by the source. **[Documented]** = provider's current docs. **[Source]** = read directly from code. **[Claimed]** = stated without a checkable method, or self-reported by a vendor. **[Search-level]** = seen only in a search excerpt or abstract summary; full text not read. No live API cache tests were run for these notes.

## Prompt compression and summarization: what is lost, and what do measured comparisons show?

### Takeaway
For coding agents, dropping or masking old tool output is about as good as LLM summarization and cheaper, but neither reliably beats a full-context agent on solve rate. The most repeatable harm from lossy compaction is the loss of constraints and byte-exact text, not loss of general gist. "Lossless" systems keep the originals but still depend on the agent choosing to retrieve them.

### Cited Findings

**Observation masking versus summarization (JetBrains Research and TUM, "The Complexity Trap", arXiv 2508.21433, v3; NeurIPS 2025 DL4C workshop per search result)**
- **[Measured]** SWE-agent on SWE-bench Verified, n=500 per model. Solve rate and cost per instance — [paper](https://arxiv.org/html/2508.21433v3):

| Model | Raw | Observation masking | LLM summary |
|---|---|---|---|
| Qwen3-32B | 17.0%, $1.12 | 15.0%, $0.55 | 16.0%, $0.50 |
| Qwen3-32B (thinking) | 23.0%, $0.51 | 24.6%, $0.46 | 24.8%, $0.51 |
| Qwen3-Coder 480B | 53.4%, $1.29 | 54.8%, $0.61 | 53.8%, $0.64 |
| Gemini 2.5 Flash | 32.8%, $0.41 | 35.6%, $0.18 | 36.0%, $0.24 |
| Gemini 2.5 Flash (thinking) | 40.4%, $0.56 | 36.4%, $0.24 | 31.4%, $0.25 |

- **[Measured]** Only two solve-rate changes are significant (paired bootstrap, p<0.05), and both are losses on Gemini 2.5 Flash with thinking: masking −4.0 pp (95% CI −7.8 to −0.2) and summary −9.0 pp (CI −13.0 to −5.2). The apparent gains elsewhere are not significant — [paper](https://arxiv.org/html/2508.21433v3).
- **[Measured]** Summary calls took up to 7.2% of instance cost, and removing them erased most of the efficiency gap to masking. Summary also lengthened runs: about 15% over Raw on Qwen3-Coder, and 52 turns versus 44 for masking on Gemini 2.5 Flash — [paper](https://arxiv.org/html/2508.21433v3).
- **[Measured]** Masking window M=10 worked best and M=20 was worse. On a 50-instance OpenHands probe the window had to be retuned. A hybrid was tested only on n=50, on one model — [paper](https://arxiv.org/html/2508.21433v3).

**ACON (Kang et al., arXiv 2510.00615, v3 dated 2026-06-01; code under github.com/microsoft/acon)**
- **[Measured]** With gpt-4.1 as agent and compressor, history compression — [paper PDF](https://arxiv.org/pdf/2510.00615):

| Benchmark | No compression | FIFO | Retrieval | LLMLingua | Naive prompting | ACON UT | ACON UTCO |
|---|---|---|---|---|---|---|---|
| AppWorld accuracy / peak (k tokens) | 56.0 / 9.93 | 45.8 / 6.73 | 27.4 / 8.39 | 39.3 / 7.50 | 43.5 / 6.93 | 51.2 / 7.17 | 56.5 / 7.33 |
| OfficeBench accuracy / peak | 76.84 / 7.27 | 67.37 / 4.02 | 65.26 / 4.33 | 70.53 / 4.65 | 71.58 / 4.40 | 74.74 / 4.93 | 72.63 / 4.54 |
| 8-objective QA EM / peak | 0.366 / 10.35 | 0.293 / 5.09 | 0.331 / 5.11 | 0.363 / 5.68 | 0.376 / 4.73 | 0.373 / 4.71 | 0.335 / 4.65 |

- **[Measured]** Peak tokens fall, but steps can rise: AppWorld went from 16.14 steps uncompressed to 22.82 with ACON UTCO history compression — [paper PDF](https://arxiv.org/pdf/2510.00615).
- **[Measured]** On AppWorld, estimated API cost per task went from $0.331 to $0.285 (history) or $0.272 (observation), using a Qwen3-14B compressor and **excluding input caching**. Median latency rose from 73.24 s to 87.68 s or 101.92 s in a local Qwen3-14B setup — [paper PDF](https://arxiv.org/pdf/2510.00615).

**Token-level compression (LLMLingua family and successors)**
- **[Claimed by authors, Findings of ACL 2024]** LLMLingua-2 is 3x to 6x faster than prior compressors and cuts end-to-end latency 1.6x to 2.9x at 2x to 5x compression — [abstract](https://arxiv.org/abs/2403.12968).
- **[Measured, by authors of a competing method]** As an agent-history baseline in ACON, LLMLingua scored 39.3 on AppWorld against 56.0 uncompressed, though it held up on 8-objective QA (EM 0.363 against 0.366) — [ACON PDF](https://arxiv.org/pdf/2510.00615).
- **[Measured]** Łajewska et al. (EMNLP Findings 2025) report that controlling compression granularity improved a soft-prompting method by up to 23% in downstream performance, and their improved method preserves 2.7x more entities — [abstract](https://arxiv.org/abs/2503.19114).
- **[Measured, single author at a compression vendor, 2026-03-06]** 5,400 API calls at compression ratio r=0.3: DeepSeek output grew 56x on MBPP (instruction survival about 0.15) but 5x on HumanEval (about 0.72). GPT-4o-mini was comparatively stable. A separate GPU measurement found a 17.4% token reduction gave only a 4.8% reduction in joules — [paper PDF](https://arxiv.org/pdf/2603.23527). The paper itself warns that 74% of DeepSeek MBPP trials hit the 1,024-token output ceiling, so those means are censored.

**Compaction failures (2026 cluster, all preprints)**
- **[Measured]** Governance Decay (Chen, 2026-06-21): across 1,323 episodes and seven model families, constraint violations went from 0% with the policy in context to 30% after compaction, up to 59% for some models. If the constraint survived the summary, violations stayed at 0%; if dropped, they reached 38%. Pinning constraints outside compaction restored 0% on this benchmark — [abstract](https://arxiv.org/abs/2606.22528).
- **[Measured]** Lost in Context Compaction (Wang et al., 2026-07-31, revised 2026-10-11): evaluated compactors retained 17% of injected side constraints on average. An inbox agent deleted email without approval in 30 of 100 sessions after compaction and in none before. A dedicated extractor retained over 90% — [abstract](https://arxiv.org/abs/2608.11242).
- **[Measured]** Benchmarks Are Surprisingly Compaction-Friendly (Salle et al., 2026-10-22): using OpenCode, five strategies, three models. No method beat default self-summarization. Compaction threshold had no measurable effect for the two 1M-context models. Methods discarding more content cost some models up to 20 points. The first summary dropped about a third of ProgramBench anti-cheating rules, and in real SWE-chat sessions process and prohibition instructions were dropped about 6x more often than build specifications. The authors argue benchmarks hide this because goals are fixed and progress lives on disk — [abstract](https://arxiv.org/abs/2610.31778).
- **[Measured]** Byte-exact loss: in a single-shot SEARCH/REPLACE study (Sonnet 4.6, 40 SWE-bench-derived Go tasks), compression cut billable tokens 73% and cost 59%, but patch applicability fell from 27/40 to 15/40, and resolved tasks went from 2/40 to 1/40. The sample is small — [Token Reduction paper](https://arxiv.org/html/2607.12161).
- **[Search-level]** arXiv 2608.11392 reports that a rule surviving compaction in degraded form can look present yet fail to constrain behaviour, so text-presence audits give false assurance — [listing](https://www.alphaxiv.org/abs/2608.11392).

**Lossless or recoverable approaches**
- **[Measured, self-reported by the system's authors]** LCM (Ehrlich and Blackman, Voltropy, submitted 2026-02-14 despite the 2605 ID — [arXiv record](https://arxiv.org/abs/2605.04050)): verbatim immutable store, summary DAG, soft and hard thresholds, and a three-level escalation ending in deterministic truncation to 512 tokens. On OOLONG trec_coarse, Volt averaged 74.8 against 70.3 for Claude Code v2.1.4, both with Opus 4.6. Claude Code led at 8K and 16K; Volt led at larger sizes — [paper](https://arxiv.org/html/2605.04050v1).
- **[Documented by authors]** LCM's "lossless" means retrievable, not used: the paper says it cannot guarantee the agent will retrieve, and `lcm_expand` is limited to sub-agents. Tasks showing benchmark memorization were excluded, and the raw baseline could not be decontaminated the same way — [paper](https://arxiv.org/html/2605.04050v1).
- **[Measured, abstract figures]** ARC, Addressable Recall Compaction (Dang et al., 2026-07-27): append-only ID-addressed tool log with citations in context. Needle-in-a-Haystack 99.40% against 88.12% for the best baseline; LongBench-v2 Hard 29.97% against 28.25% — [abstract](https://arxiv.org/abs/2607.25066). The tests use Qwen3-8B and Qwen3-32B and are not coding-agent benchmarks. Do not confuse it with a different "ARC" (Active and Reflection-driven Context Management, [arXiv 2601.12030 listing](https://fugumt.com/fugumt/paper_check/2601.12030v1_enmode)).

### Inferences
- Default ranking for tool-heavy coding work: keep recent observations verbatim, replace older ones with a pointer or placeholder, summarize rarely, and store constraints outside anything that gets summarized.
- Peak-token reduction is a weak success metric, because ACON and The Complexity Trap both show runs can get longer.
- Benchmark results probably understate compaction risk for real sessions, because constraints given only in conversation exist nowhere else.
- Never compress text that will be used as an exact edit anchor, an identifier, a path, or a quoted requirement.

### Gaps
- No head-to-head study found of LCM, ARC, masking and default compaction on the same coding benchmark with billed cost.
- LCM's sample size was not found in the text reviewed, and its 32K to 128K per-length gains are not given in text.
- No 2025 to 2026 independent benchmark of the whole LLMLingua family with provider caching turned on was found.
- arXiv 2608.11392 and TRACE (compression instability) were not read in full.

## Cache-aware design: what do "Don't Break the Cache", "Token Reduction Is Not Cost Reduction" and related work show?

### Takeaway
A stable cached prefix delivers most of the input savings, and in cached agent runs, reducing visible tokens may not reduce the bill. Keep tools, system text and history append-only, put volatile values last, and evaluate by billed cost per successful task.

### Cited Findings

**Don't Break the Cache (arXiv 2601.06007, v2, January 2026)**
- **[Measured]** DeepResearch Bench, 10,000-token system prompt, 40 sessions per model and condition, more than 500 total. Pricing as of early January 2026 — [paper](https://arxiv.org/html/2601.06007v2):

| Model | Full context: cost / TTFT | System prompt only | Exclude tool results |
|---|---|---|---|
| GPT-5.2 | −79.3% / +9.5% | −81.4% / +10.5% | −79.6% / +13.0% |
| Claude Sonnet 4.5 | −77.8% / +21.8% | −78.5% / +22.9% | −78.1% / +20.9% |
| Gemini 2.5 Pro | −38.3% / +6.0% | −41.4% / +6.1% | −27.8% / −2.9% |
| GPT-4o | −47.8% / −8.8% | −45.9% / +30.9% | −46.8% / +28.1% |

  Cost is change in cost; TTFT is improvement, so negative TTFT is a regression.
- **[Measured]** Strategy differences on cost are only about 2 to 4 points for most models because the system prompt is cached in all of them. Full-context caching regressed GPT-4o TTFT, which the authors attribute to writes on dynamic tool content — [paper](https://arxiv.org/html/2601.06007v2).
- **[Measured]** Savings were positive from 500 to 50,000 prompt tokens. At 50,000 tokens GPT-5.2 saved 89% and Sonnet 4.5 saved 88%. At 500 tokens, below threshold, several models showed 10% to 18% TTFT regressions — [paper](https://arxiv.org/html/2601.06007v2).
- **Caution:** The paper's abstract ranges do not match all table values: Gemini's best TTFT gain is 6.1% and one Gemini cost result is 27.8%. It does not test tool-definition order. The recommendation to put dynamic values at the end of the system prompt is guidance, not an isolated experiment — [paper](https://arxiv.org/html/2601.06007v2).

**Token Reduction Is Not Cost Reduction (Weinberger and Hozez, arXiv 2607.12161, v1 2026-07-13, v5 2026-08-12 — [record](https://arxiv.org/abs/2607.12161))**
- **[Measured]** 2,848 analyzed Claude Code runs of 2,908, across 103 tasks and 7 repositories, on Haiku 4.5, Sonnet 5 and Opus 4.8, using provider-billed cost — [paper](https://arxiv.org/html/2607.12161).
- **[Measured]** Baseline bill: cache creation 44.3%, cache reads 35.4%, output 10.4%, uncached input 1.3%, unattributed 8.7%. Tool outputs were about 3.3% of cost — [paper](https://arxiv.org/html/2607.12161).

| Arm | Billed cost vs baseline (95% CI) | Success | Mean turns | Cost per success vs baseline |
|---|---|---|---|---|
| Baseline | — | 96.2% | 4.46 | 1.000 |
| RTK v0.44.1 | −2.7% (−5.6, −0.1) | 96.4% | 4.32 | 0.968 (0.937, 1.004) |
| RTK-ML | +6.8% (+2.8, +11.3) | 97.5% | 4.69 | 1.051 (1.006, 1.100) |
| Headroom v0.27.0 | +48.4% (+42.3, +55.0) | 97.8% | 4.20 | 1.464 (1.397, 1.531) |

- **[Measured]** RTK-ML cut estimated raw tool-output tokens by 38.4% and still cost more. Task-level reduction barely predicted cost change (Pearson r=0.15, CI crossing zero) — [paper](https://arxiv.org/html/2607.12161).
- **Mechanism caution:** The paper attributes RTK-ML's cost increase to changed trajectories, such as extra searches, re-reads and turns. It does not demonstrate cache invalidation, and Headroom's internal cause is not established. Cache carryover between arms cannot be fully excluded because consecutive runs fell inside the five-minute TTL — [paper](https://arxiv.org/html/2607.12161).

**Layout rules from providers and harness builders**
- **[Documented]** Anthropic's prefix order is `tools` → `system` → `messages`, and a change invalidates that level and later ones. Tool definition changes invalidate everything; `tool_choice` and image changes invalidate only messages — [Anthropic](https://platform.claude.com/docs/en/build-with-claude/prompt-caching).
- **[Documented]** Anthropic: a breakpoint on a per-request value such as a timestamp writes a new entry each time and never hits. One breakpoint at the end of static content is usually enough — [Anthropic](https://platform.claude.com/docs/en/build-with-claude/prompt-caching).
- **[Documented]** OpenAI lists tool names, descriptions, schemas, ordering and tool-specific instructions as settings to preserve — [OpenAI](https://developers.openai.com/api/docs/guides/prompt-caching).
- **[Claimed, first-party engineering post, 2026-04-30]** Claude Code's layout is static system and tools, then CLAUDE.md, session context, and messages. Past breaks included a timestamp in the static prompt, non-deterministic tool ordering, and changed tool parameters. It uses messages for updates, deferred tool stubs, and compaction forks that reuse the parent's prefix, and it treats drops in hit rate as incidents — [Anthropic engineering](https://claude.dev/blog/lessons-from-building-claude-code-prompt-caching-is-everything/).
- **[Claimed, practitioner case]** One moved dynamic identifier reportedly raised a security workload's hit rate from 7% to 74% and cut the monthly bill 59% — [DigitalOcean, search-level](https://www.digitalocean.com/community/tutorials/prompt-caching-in-practice-hit-rate).

### Inferences
- Treat request serialization as a tested contract: sort tools deterministically, keep JSON stable, and put clocks, IDs and per-run selections after the last stable boundary.
- A compressor that rewrites old history breaks the prefix from the first edit onward. One that transforms a tool result once, before it enters history, keeps append-only behaviour, although it can still change the trajectory.
- Report cost as cache read, cache write, uncached input and output separately, plus cost per success. Token counts alone can point the wrong way.

### Gaps
- No controlled study found that isolates tool-order instability as a variable; evidence is provider guidance and engineering reports.
- No public study found that cleanly separates cache-invalidation cost from trajectory change for context compressors.

## Exact provider semantics and headless CLI behaviour

### Takeaway
Caching is now exact-prefix everywhere documented, but the reuse boundary differs. Anthropic and current OpenAI models document block or message boundaries, so a shared start inside one user text block should not be assumed to hit. Google, Moonshot and xAI do not document that case. Put the shared content in its own block and end a breakpoint there.

### Cited Findings

**Summary table (as fetched 2026-11-18)**

| Provider | Activation | Minimum | Lifetime | Read / write price vs input | Usage fields |
|---|---|---|---|---|---|
| Anthropic | Explicit `cache_control` on blocks, or one top-level field for automatic placement | 512 to 4,096 tokens by model | 5 min default; 1 h option; refreshed on use | Read 0.1x (0.05x Opus 5.5 and Sonnet 5.5; 0.025x Fable and Mythos 5.1 and 5.2). Write 1.25x (5 min), 2x (1 h) | `cache_read_input_tokens`, `cache_creation_input_tokens`, `cache_creation.*` |
| OpenAI GPT-5.6 and later | Implicit or explicit mode | 1,024 visible input tokens | `30m` default; `12h` in Responses | Read 0.1x, or 0.05x on GPT-6.1 Sol and Terra and GPT-6.2 Astra and Sol. Write 1.25x (30m), 2.5x (12h) | `cached_tokens`, `cache_write_tokens`, TTL details |
| OpenAI earlier models | Implicit | Varies by request settings | `in_memory` about 5 to 10 min idle, up to 1 h; `24h` typically about 30 min, up to 24 h | Model-dependent read rate; no write charge | `cached_tokens`, rounded down to a multiple of 128 |
| Gemini API | Implicit on 2.5 and newer | 4,096 for Gemini 4 models, 3.8 Flash, 3.6 Flash, 3.1 Pro Preview; 2,048 for 2.5 | Not stated on AI Studio page | Discount not stated on AI Studio page; Vertex says 90% | `usage.total_cached_tokens` (AI Studio); `cachedContentTokenCount` (Vertex) |
| Moonshot/Kimi | Automatic prefix cache | Conflicting; see below | `5m` default, `1h` option; TTL fixed at first write | Read 0.1x. kimi-k3: miss $3.00, 5m write $3.00, 1h write $6.00, hit $0.30 per 1M | `cached_tokens`, `cache_write_tokens`; Anthropic-style fields on Messages |
| xAI | Automatic per docs and search excerpts; routing ID recommended | Not stated on pages read | No fixed TTL; can be evicted at any time | Cached rate on pricing page; write charge not stated | `prompt_tokens_details.cached_tokens`, `input_tokens_details.cached_tokens` |

Sources: [Anthropic](https://platform.claude.com/docs/en/build-with-claude/prompt-caching), [OpenAI](https://developers.openai.com/api/docs/guides/prompt-caching), [Gemini API](https://aistudio.google.com/docs/caching), [Vertex AI](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/context-cache/context-cache-overview), [Kimi](https://platform.kimi.ai/docs/guide/context-caching.md), [xAI best practices](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/best-practices), [xAI usage](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/usage-and-pricing).

**Does a shared start inside one user text block hit?**

| Provider | Documented answer | Status |
|---|---|---|
| Anthropic | A marked block writes one entry, a hash of the prefix ending at that block, and no earlier positions. Lookback checks block positions. A hit needs identical content through the marked block | **No by documented design**; the docs do not address the mid-block case in those words — [source](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) |
| OpenAI GPT-5.6+ | Implicit breakpoints fall at the end of eligible messages. Extending a message moves the old endpoint inside it, and without an explicit breakpoint there the prefix is not reused | **No, unless an explicit breakpoint exists at that boundary** — [source](https://developers.openai.com/api/docs/guides/prompt-caching) |
| OpenAI earlier models | Breakpoints occur at model-dependent intervals counted from the start of the hidden system message | **Not stated.** Interval placement suggests a hit could end inside a block — [source](https://developers.openai.com/api/docs/guides/prompt-caching) |
| Gemini | No rule on content, part or token granularity | **Silent.** On Gemini 4, a turn adding fewer than 2,050 tokens has the last 2,050 tokens of the previous request reprocessed at the standard rate — [source](https://aistudio.google.com/docs/caching) |
| Moonshot/Kimi | Content after a changed position cannot be reused; storage is in blocks of unspecified size, and a partial block is not written | **Not confirmed.** Partial reuse seems plausible, at unknown granularity — [source](https://platform.kimi.ai/docs/guide/use-context-caching-feature-of-kimi-api.md) |
| xAI | Append new messages; editing, deleting or reordering earlier ones breaks the prefix | **Silent** on mid-message matching — [source](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/best-practices) |

**Anthropic details [Documented]** — [source](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)
- Up to 4 breakpoints. If all 4 are explicit, adding top-level automatic caching returns a 400.
- Lookback checks at most 20 positions per breakpoint, including the breakpoint. On the Claude API, a run of consecutive `tool_use` blocks counts as one position, as does a run of `tool_result` blocks.
- Minimums: 512 for Fable 5.2, Mythos 5.2, Fable 5.1, Mythos 5.1, Opus 5.5, Opus 5, Sonnet 5.5, Fable 5, Mythos 5, Haiku 5.5. 1,024 for Opus 4.8, Sonnet 5, Sonnet 4.6, Sonnet 4.5, Opus 4.1, Opus 4, Sonnet 4. 2,048 for Mythos Preview, Opus 4.7. 4,096 for Opus 4.6, Opus 4.5, Haiku 4.5. Shorter prompts are processed uncached with no error.
- TTL runs from the start of the request that writes or reads the entry. An entry is available only after the first response starts, so parallel requests need the first response to begin.
- Automatic caching is available except on Amazon Bedrock `bedrock-runtime`. It targets the last cacheable block and walks backward if that block is ineligible.
- Caches are never shared between organizations. Claude API, Claude Platform on AWS and Foundry isolate by workspace; Google-managed Claude and Bedrock isolate by organization.

**OpenAI details [Documented]** — [source](https://developers.openai.com/api/docs/guides/prompt-caching)
- The widely repeated "1,024 tokens, then 128-token increments" is no longer a complete description. 1,024 is the stated GPT-5.6+ minimum. 128 now appears as the rounding of reported `cached_tokens` on earlier models, whose minimum varies with settings such as tools, images and output schemas.
- GPT-5.6+ implicit lookup considers up to 20 earlier eligible message endings plus the end of the initial developer block. Each request can create up to four writes, and an implicit breakpoint uses one. Explicit lookup covers the first 2 and latest 50 explicit breakpoints. The explicit marker is `prompt_cache_breakpoint: {"mode": "explicit"}`.
- `prompt_cache_key` influences routing on pre-5.6 models but does not guarantee a hit. Guidance is about 15 requests per minute per key, since higher traffic can overflow. On GPT-5.6+ it is optional and mainly for cache accounting.
- Retention defaults on earlier models: `24h` without Zero Data Retention, `in_memory` with it. Caches are not shared across organizations or regional processing boundaries.

**Google details [Documented]**
- Gemini API: explicit caching is not supported in the Interactions API or for Gemini 4 models, and remains available on supported Gemini 3 and 2.5 models through generateContent — [Gemini API](https://aistudio.google.com/docs/caching). This applies to `gemini-4-hydrogen`, which the page lists at a 4,096-token minimum.
- Vertex AI: implicit caching gives a 90% discount with no storage charge. Explicit caching is 90% on Gemini 2.5 or later and 75% on Gemini 2.0, has storage charges, and defaults to 60 minutes. Some Gemini 3 models need 6,144 tokens for implicit caching. A cached prefix expires no later than 24 hours after creation — [Vertex AI](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/context-cache/context-cache-overview).
- **Conflict:** a July 2026 keepalive paper uses a 0.25x implicit read ratio for Google — [Khailo](https://arxiv.org/pdf/2607.19214), against Vertex's documented 90% discount. Confirm on the price page for the exact model and platform.

**Moonshot and xAI details**
- **[Documented]** Kimi's Messages endpoint honours `cache_control` only at the top level. `kimi-k3.1-pro` and `kimi-k3` support cache writes — [Kimi](https://platform.kimi.ai/docs/guide/context-caching.md). The Kimi academy page was updated 2026-09-28 — [Kimi academy](https://www.kimi.com/academy/best-practices-for-context-caching).
- **Conflict:** a search excerpt attributed a 256-token threshold to official Kimi docs — [search lead](https://platform.kimi.ai/docs/guide/use-context-caching-feature-of-kimi-api.md). The three Kimi pages fetched do not state it; 256 appears there only as a `max_tokens` example. Treat it as unverified.
- **[Documented]** xAI recommends `x-grok-conv-id`, or `prompt_cache_key` on Responses, to route to the same server. Hits are not guaranteed — [xAI](https://docs.x.ai/developers/advanced-api-usage/prompt-caching/best-practices).

**Claude Code**
- **[Documented]** Layers are system prompt with tools, then project context, then conversation. Plan mode and skills are appended as messages. Model changes, most effort changes, fast-mode changes, some tool-set changes, compaction and image pruning can invalidate the cache — [Claude Code](https://code.claude.com/docs/en/prompt-caching).
- **[Documented]** Fresh-run reuse is effectively limited to one machine and directory. Parallel sessions in the same directory can share. Sequential sessions share only when the startup git-status snapshot matches, because the system context carries auto-memory paths, working directory, platform, shell, OS, branch and recent commits — [Claude Code](https://code.claude.com/docs/en/prompt-caching).
- **[Documented]** `-p` runs are in the main-conversation bucket. Default TTL is 1 hour on a subscription within plan usage and 5 minutes with usage credits, API key or cloud provider. Subagents default to 5 minutes. `promptCacheTtl`, `CLAUDE_CODE_PROMPT_CACHE_TTL`, `subagentPromptCacheTtl` and `CLAUDE_CODE_SUBAGENT_PROMPT_CACHE_TTL` control this — [Claude Code](https://code.claude.com/docs/en/prompt-caching).
- **[Documented]** `claude -p "hello" --output-format json` reports `usage.cache_creation`, split into `ephemeral_1h_input_tokens` and `ephemeral_5m_input_tokens`. `/usage`, status-line data and OpenTelemetry expose cache measures — [Claude Code](https://code.claude.com/docs/en/prompt-caching).
- **[Claimed, user report]** Issue 63930 describes full cache recreation in long sessions after parallel tool calls and deferred-tool materialization — [issue mirror](https://claudeissues.com/issue/63930-prompt-cache-fully-re-created-after-turns-with-many-parallel-tool-calls-cache-re).

**Codex CLI**
- **[Source, main branch checked 2026-11-18]** Default `PromptCacheKey` is `SessionDerived`; hosts can use `Custom(String)` or `Omit` — [thread_manager.rs](https://github.com/openai/codex/blob/main/codex-rs/core/src/thread_manager.rs), [client.rs](https://github.com/openai/codex/blob/main/codex-rs/core/src/client.rs).
- **[Source]** `codex exec` JSON `Usage` has `input_tokens`, `cached_input_tokens`, `cache_write_input_tokens`, `output_tokens` and `reasoning_output_tokens` — [exec_events.rs](https://github.com/openai/codex/blob/main/codex-rs/exec/src/exec_events.rs). The locally installed version is 0.144.1, which may differ from main.
- **[Measured by issue reporters, not confirmed by OpenAI]** Issue 35300 (opened 2026-07-25, still open in the fetched copy): on a replay with a stable prefix of about 9k tokens and a changed final user message, unmodified requests got 0% hits, while an explicit breakpoint on the last stable block cached 8,937 tokens (98.6%). Within one session, append-only turns reached about 98%. On a Bedrock Mantle cross-session test, only breakpoint plus stable key produced hits (0%, 96%, 97%). Other users report that some backends reject the field — [issue](https://github.com/openai/codex/issues/35300).

**Gemini CLI**
- **[Documented]** Token caching applies with API-key authentication (Gemini API key or Vertex AI), not OAuth, and `/stats` shows savings when cached tokens exist. The page says Antigravity CLI replaced Gemini CLI on 2026-06-18 for unpaid-tier and Google One users, and that Gemini CLI is not optimized for Gemini 4 Argon — [Gemini CLI](https://geminicli.com/docs/cli/token-caching/).
- **[Source, installed 0.68.0]** The stream-JSON formatter's `stats` include `total_tokens`, `input_tokens`, `output_tokens`, `cached`, `input`, `duration_ms`, `tool_calls` and per-model values. `cached` is populated from `cachedContentTokenCount` — local bundle `~/.local/lib/node_modules/@google/gemini-cli/bundle/chunk-2OHBIZ5H.js`, lines 287989 and 391138 onward. This corrects a [third-party cheatsheet](https://littlebearapps.com/help/untether/gemini-stream-json-cheatsheet/) whose example shows no cached field.

### Inferences
- Portable layout: place the shared content in its own leading block or message, close it there, and put the query or variable material in a later block. On Anthropic, mark the shared block. On GPT-5.6+, use an explicit breakpoint when the shared block is not the end of an eligible message.
- For fresh headless runs, cache reuse depends on the harness's prefix as much as on your prompt. The working directory, git snapshot, tool set, auth path and session-derived key can each prevent reuse.
- A definitive per-provider answer for the silent cases needs a small probe: send `A+B` then `A+C` in one text block, with `A` above the minimum, and repeat with `A` as a separate block. Compare cache fields over several trials, because routing can produce false misses.

### Gaps
- Google, Moonshot and xAI do not document mid-block or mid-part hit behaviour on the pages read. Anthropic and OpenAI earlier models also lack a direct sentence on this exact case.
- Claude Code's docs read here do not publish its exact breakpoint count or positions. A third-party claim of four was not verified.
- Claude Code's guidance for improving reuse across users and machines was not read. OpenAI's Codex agent-loop post returned HTTP 403, so only secondary summaries were seen.
- Whether current Codex main can emit `prompt_cache_breakpoint` was not established; the issue says 0.160.1 did not as of 2026-10-06.
- xAI's minimum size, TTL and write pricing were not found. A Moonshot forum mention of `prompt_cache_key` was not confirmed in its docs.

## Cache hit-rate modelling, eviction, TTL economics, and joint context selection

### Takeaway
Simple write-versus-read arithmetic explains most TTL choices, but real retention differs by provider and is only partly specified. I found no paper that solves context selection jointly with prefix-cache state, which looks like an open and useful problem.

### Cited Findings
- **[Measured, single author, arXiv v2 stamped 2026-07-24; PDF body dated 2026-07-27]** Khailo tested Anthropic, OpenAI, Google and DeepSeek, with prefixes of 40k and 100k tokens and idle gaps up to 40 minutes. No provider evicted before its documented TTL — [paper](https://arxiv.org/pdf/2607.19214).
- **[Measured]** Retention: Anthropic was warm through 300 s and gone by 360 s. DeepSeek collapsed by 540 to 600 s. OpenAI was about 50% warm at 15 to 20 minutes and fully cold by 30. Google stayed between 33% and 83% warm at every gap, which the author treats as routing probability rather than retention — [paper](https://arxiv.org/pdf/2607.19214).
- **[Model]** With read ratio `r`, re-prefill ratio `w`, ping interval `τ` and idle time `I`, keepalive costs `(I/τ + 1)r` per input token against `w` for letting the entry die. Break-even is `Imax ≈ τ(w/r − 1)`: about 46 minutes for Anthropic's 5-minute tier at `τ≈240 s`, about 36 minutes for OpenAI and DeepSeek, and about 12 minutes for Google at the paper's `r=0.25` — [paper](https://arxiv.org/pdf/2607.19214).
- **[Measured]** A 30-second keepalive held a median of at least 98% cached in every tested group. At 600 s Anthropic's baseline was 0/48 warm against 40/40 with keepalive. A 240-second interval cut ping spend 7.8x, and a 100k Anthropic prefix cost about $3.60 per hour at 30 s against about $0.45 at 240 s — [paper](https://arxiv.org/pdf/2607.19214).
- **[Measured]** At a 30-minute pause, keepalive saved 1.56x on Anthropic at 240 s and up to 2.45x on OpenAI at 480 s. It lost money on DeepSeek because re-prefill was so cheap. Reported TTFT fell from 2.9 to 1.5 s on Anthropic, 4.4 to 1.3 s on OpenAI, and 5.4 to 2.0 s on DeepSeek — [paper](https://arxiv.org/pdf/2607.19214).
- **[Claimed]** The same author argues widespread keepalive would weaken LRU eviction and push providers toward charging for residency — [paper](https://arxiv.org/pdf/2607.19214). This is argument, not measurement.
- **[Measured by reporter, not a controlled study]** One Codex user's AWS costs for 2026-07-01 to 07-23 show 258.8M cache-write tokens at $1,780.33, or 90.0% of $1,977.58 model spend, with 28.6M read tokens at $15.71. The reporter says this does not prove a single cause — [issue 35300](https://github.com/openai/codex/issues/35300).
- **[Search-level]** Nearby selection work: Budget-Aware Routing frames token-limited unit selection as knapsack-constrained subset selection, without cached prices — [ACL Findings 2026](https://aclanthology.org/2026.findings-acl.2114.pdf). The Efficiency Frontier models preprocessing reuse through amortized cost — [arXiv 2605.23071](https://arxiv.org/html/2605.23071). "Pricing the Context Window" prices a cached prefix token at a fraction `q` of a fresh token and uses a shadow price, but it is non-peer-reviewed and returned HTTP 403 when fetched — [Preprints.org](https://www.preprints.org/manuscript/202610.0614).
- **[Search-level]** Semantic caching is a different risk class. vCache adds user-defined error-rate guarantees because fixed similarity thresholds are unreliable — [arXiv 2502.03771](https://arxiv.org/pdf/2502.03771). Adversarial prompts can force false hits by key collision — [arXiv 2601.23088](https://arxiv.org/pdf/2601.23088). Provider prefix caching reuses processed prompt state and still generates a fresh answer; OpenAI describes storing encrypted KV tensors — [OpenAI](https://developers.openai.com/api/docs/guides/prompt-caching).

### Inferences
These are my derivations from the documented prices, not published results.
- Break-even reuse count: a write pays back after `n > (w−1)/(1−r)` reads. With `r=0.1`: `w=1.25` needs 1 read, `w=2` needs 2, `w=2.5` needs 2. At `w=1` (kimi-k3 5m, or OpenAI earlier models with no write charge) there is no write premium to recover.
- For one possible future reuse, writing at `w=1.25, r=0.1` pays when reuse probability exceeds about 28%. At `w=2` it needs a second expected reuse to pay back.
- Expected request input cost is approximately `C = P_read·r + P_write·w + P_uncached·1`, where the `P` values are shares of input tokens. Estimate hit probability from the gap distribution against measured retention, not only the nominal TTL.
- Cache-aware context selection is not an ordinary knapsack with a cheap item price. An item gets the cached price only if every earlier item in the prefix is unchanged, so the feasible cached set is prefix-closed. A practical solver can enumerate candidate reuse boundaries `k`, keep the prefix through `k` at read price, and solve a knapsack for the suffix at write or uncached price, within a token limit and with a penalty for added context.
- Add a quality cost per retained token. A cached token is cheap, but the next section shows that it is not free in answer quality.

### Gaps
- No paper was found that jointly optimizes context selection, ordering and provider cache state with measured answer quality.
- Providers do not publish eviction policies in the docs read, and xAI explicitly gives no fixed TTL.
- The keepalive result is one author's study, and its price inputs may be stale. CacheWise, its cited serving-side trace study, was not read.
- No authoritative cross-provider hit-rate model validated on production agent traces was found. Published hit rates such as those in vendor blogs are not comparable.
- No controlled comparison of semantic-cache false-positive rates against prefix caching on the same workload was found.

## Long-context degradation and position effects

### Takeaway
More input can hurt even when it fits and even when the needed evidence is present. Position matters, and placing the question last is sensible and cache-friendly, but it does not cancel the length penalty.

### Cited Findings
- **[Measured]** Lost in the Middle (Liu et al., 2023, TACL): on multi-document QA and key-value retrieval, performance is often highest when relevant information is at the start or end and drops when it is in the middle of long contexts — [abstract](https://arxiv.org/abs/2307.03172).
- **[Measured, vendor report, 2025-07-14]** Chroma's Context Rot tested 18 models on Needle in a Haystack, LongMemEval and Repeated Words. Performance degraded with input length, faster with lower needle-question similarity, and even one distractor hurt. On LongMemEval's 306 prompts (about 113k tokens full against about 300 focused), every model scored higher with focused input. Shuffling haystack sentences improved performance. Many results appear only in charts, and only two haystack topics were used — [Chroma](https://www.trychroma.com/research/context-rot).
- **[Search-level, peer reviewed: Findings of EMNLP 2025]** Du et al. report 13.9% to 85% degradation across five models on math, QA and coding, even when models could perfectly retrieve the evidence and inputs stayed within limits. The effect persisted with whitespace replacement and masking of irrelevant tokens, and with evidence placed directly before the question. Reciting the evidence and then answering in a shorter prompt improved GPT-4o by up to 4% on RULER — [arXiv 2510.05381](https://arxiv.org/abs/2510.05381v1).
- **[Measured, small study, 2026-03-16, MathAI 2026]** With a 256-token answer-bearing window placed last and distractors prepended, 200 instances per dataset, exact-match scores at 256 and 2,048 tokens were: gpt-4.1 HotpotQA 0.575 to 0.305; gpt-4.1-mini HotpotQA 0.505 to 0.260; gpt-4.1 SQuAD 0.635 to 0.570; gpt-4.1-mini SQuAD 0.545 to 0.520 — [paper summary](https://www.alphaxiv.org/abs/2603.15723.md). The paper has no position analysis, and its abstract's "nearly twice" understates its own Table 3 gap.
- **[Claimed]** Anthropic says queries at the end can improve response quality by up to 30 percent in tests, especially with complex multi-document inputs, and gives no method on the page — [Anthropic](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices).
- **[Measured, nuance]** Compaction threshold showed no measurable effect for two 1M-context models, while the smallest-window model did worse at lower thresholds on some benchmarks, so "less context is always better" is also false — [Salle et al.](https://arxiv.org/abs/2610.31778).

### Inferences
- Stable reference first, question last satisfies both prefix caching and the best-supported position advice.
- Do not treat a cheap cached token as harmless. Keep cached background only when its expected usefulness outweighs its distractor cost.
- For tasks needing several facts, a two-stage flow may help: locate or recite evidence first, then answer from a short context. Evidence for this is limited.
- Newer long-context models may be more robust. Most controlled degradation data here predates current frontier models, so re-test before setting hard limits.

### Gaps
- No controlled 2026 study was found that varies both cached status and context length on current frontier models and reports answer quality.
- The exact Du et al. tables and Chroma's model-level values were not extracted. A blog-reported Chroma figure for GPT-4o was not verified and is omitted.
- Anthropic's 30 percent figure cannot be assessed without its test design.
