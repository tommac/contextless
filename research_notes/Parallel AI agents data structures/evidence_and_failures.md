# Empirical evidence on multi-agent LLM systems, 2023 to 2026: gains, losses, failures and coordination cost

Reading notes for the report writer:
- "Budget-matched" means the single-agent baseline was given comparable tokens or LLM calls. Results without that are marked NOT budget-matched.
- Percentages written "+80.8%" are relative changes unless "points" is stated.
- Several papers changed their numbers between arXiv versions, and some have internal inconsistencies. These are flagged where found.
- Most numbers were extracted from arXiv HTML or PDF text by an automated reader. Figures that exist only in images could not be read and are listed under Gaps.

## 1. "Why Do Multi-Agent LLM Systems Fail?" (Cemri et al., 2025) and the MAST taxonomy

### Takeaway
MAST is a 14-mode, 3-category failure taxonomy built from annotated traces of 7 open-source frameworks running GPT-4-class and Claude-3.x-class models; roughly 42 to 44% of labelled failures are specification or system-design problems, 32 to 37% inter-agent misalignment, and 21 to 24% verification problems. The paper's own interventions (better role prompts, an added verification step, topology change) bought single-digit to about 15-point gains and left most failures in place.

### Cited Findings
- Paper identity: Cemri, Pan, Yang, Agrawal, Chopra, Tiwari, Keutzer, Parameswaran, Klein, Ramchandran, Zaharia, Gonzalez, Stoica (UC Berkeley and collaborators). v1 17 Mar 2025, v2 22 Apr 2025, v3 26 Oct 2025 (NeurIPS 2025 datasets track version). The abstract reports 14 failure modes in 3 categories, 1,600+ annotated traces across 7 frameworks, a taxonomy developed from 150 traces, and human inter-annotator Cohen's kappa of 0.88. — [arXiv abstract](https://arxiv.org/abs/2503.13657)
- Frameworks in the main set: MetaGPT, ChatDev, HyperAgent, AppWorld, AG2 (MathChat), Magentic-One, OpenManus. v2 says OpenManus and Magentic-One were used as generalization tests, and its appendix also covers AutoKaggle, Multi-Agent Peer Review and MA-ToT. — [arXiv v2 HTML](https://arxiv.org/html/2503.13657v2)
- v3 dataset: 1,642 traces; benchmarks ProgramDev, ProgramDev-v2, SWE-Bench Lite, AppWorld Test-C, GSM-Plus, OlympiadBench, MMLU, GAIA; models GPT-4o, GPT-4, GPT-4o-mini, Claude-3.7-Sonnet, Qwen2.5-Coder-32B-Instruct, CodeLlama-7b-Instruct. — [arXiv v3 HTML](https://arxiv.org/html/2503.13657v3)
- Headline failure rate (measured): "41% to 86.7% failure rate on 7 state-of-the-art open-source MAS". Per-framework rates are only in a figure. — [arXiv v3 HTML](https://arxiv.org/html/2503.13657v3)
- Failure-mode frequencies, v2 (200+ traces). Category totals as stated in the paper: Specification Issues 41.77%, Inter-Agent Misalignment 36.94%, Task Verification 21.30%. — [arXiv v2 HTML](https://arxiv.org/html/2503.13657v2)

| Mode | v2 share | v3 share |
|---|---|---|
| FM-1.1 Disobey task specification | 10.98% | 11.8% |
| FM-1.2 Disobey role specification | 0.5% | 1.5% |
| FM-1.3 Step repetition | 17.14% | 15.7% |
| FM-1.4 Loss of conversation history | 3.33% | 2.80% |
| FM-1.5 Unaware of termination conditions | 9.82% | 12.4% |
| FM-2.1 Conversation reset | 2.33% | 2.20% |
| FM-2.2 Fail to ask for clarification | 11.65% | 6.80% |
| FM-2.3 Task derailment | 7.15% | 7.40% |
| FM-2.4 Information withholding | 1.66% | 0.85% |
| FM-2.5 Ignored other agent's input | 0.17% | 1.90% |
| FM-2.6 Reasoning-action mismatch | 13.98% | 13.2% |
| FM-3.1 Premature termination | 7.82% | 6.20% |
| FM-3.2 No or incomplete verification | 6.82% | 8.20% |
| FM-3.3 Incorrect verification | 6.66% | 9.10% |

- v3 column source: — [arXiv v3 HTML](https://arxiv.org/html/2503.13657v3). v3 names the first category "System Design Issues" rather than "Specification Issues". The v3 text does not state category totals; summing the v3 modes gives about 44.2% / 32.4% / 23.5% (my arithmetic, not the paper's).
- Qualitative per-framework pattern (v3): AppWorld often shows premature termination, OpenManus step repetition, HyperAgent step repetition and incorrect verification. — [arXiv v3 HTML](https://arxiv.org/html/2503.13657v3)
- Annotation reliability: human kappa 0.88 overall (round 1 was 0.24, round 2 averaged 0.92, round 3 averaged 0.84); kappa 0.79 on unseen systems and benchmarks. LLM annotator (OpenAI o1): zero-shot accuracy 0.89, kappa 0.58; few-shot accuracy 0.94, kappa 0.77, F1 0.80. — [arXiv v2 HTML](https://arxiv.org/html/2503.13657v2); [arXiv v3 HTML](https://arxiv.org/html/2503.13657v3)
- Interventions (measured, Table 4 of v2, accuracy %):

| System | Benchmark | Model | Baseline | Improved prompts | New topology |
|---|---|---|---|---|---|
| AG2 (MathChat) | GSM-Plus | GPT-4 | 84.75 ± 1.94 | 89.75 ± 1.44 | 85.50 ± 1.18 |
| AG2 (MathChat) | GSM-Plus | GPT-4o | 84.25 ± 1.86 | 89.00 ± 1.38 | 88.83 ± 1.51 |
| ChatDev | ProgramDev-v0 (32 tasks) | not stated in table | 25.0 | 34.4 | 40.6 |
| ChatDev | HumanEval | not stated in table | 89.6 | 90.3 | 91.5 |

  — [arXiv v2 HTML, Table 4 and Appendix F](https://arxiv.org/html/2503.13657v2)
- Significance: with GPT-4 the new-topology gain over baseline was not significant (Wilcoxon p = 0.4); with GPT-4o both interventions beat baseline at p = 0.03. AG2 runs used 200 GSM-Plus problems with six repetitions per configuration. — [arXiv v2 HTML, Appendix F](https://arxiv.org/html/2503.13657v2)
- The paper summarizes the ChatDev gains as +9.4% (role-specification or workflow change) and +15.6 points absolute from adding a high-level verification step on ProgramDev (25.0 to 40.6). — [arXiv v2 HTML](https://arxiv.org/html/2503.13657v2); [arXiv v3 HTML](https://arxiv.org/html/2503.13657v3)
- Authors' claim (not a measurement): the failures call for structural redesign of the system, since prompt-level fixes left ChatDev below 41% on ProgramDev. — [arXiv v2 HTML](https://arxiv.org/html/2503.13657v2)
- Related attribution evidence: the Who&When dataset (Zhang et al., ICML 2025) has failure logs from 127 LLM multi-agent systems. The best automated method names the responsible agent with 53.5% accuracy and the decisive error step with only 14.2%; some methods are below random, and o1 and DeepSeek R1 are reported as not practically usable for the task. — [arXiv 2505.00212](https://arxiv.org/abs/2505.00212); [PMLR v267](https://proceedings.mlr.press/v267/zhang25cq.html)

### Inferences
- About one fifth to one quarter of labelled failures are verification failures, and "incorrect verification" grew from 6.66% to 9.10% in the larger v3 sample: a verifier being present does not mean it works.
- Step repetition, unawareness of termination and reasoning-action mismatch together are about 41% in both versions. These are single-agent control problems that multi-agent structure exposes more often; they are not strictly communication problems.
- The low frequencies for "information withholding" and "ignored other agent's input" (each under 2%) suggest that outright communication refusal is rare; the commoner misalignment is failing to ask (6.8 to 11.7%) and derailment (about 7%).
- Mode shares are shares of labelled failures, not probabilities that a run hits that failure, so they cannot be multiplied by task counts.
- Automated failure diagnosis is weak (14.2% step accuracy), so a design that relies on post-hoc trace analysis to find the faulty agent will mostly not find it.

### Gaps
- Per-framework failure rates and per-framework mode distributions are only in figures (v2 Fig. 1, 8, 9; v3 Fig. 5) and were not readable.
- v2 is internally inconsistent on ChatDev's ProgramDev baseline: the text cites "33.33% correctness" in one place and "75% failure" in another, while Table 4 gives 25.0.
- v3 attributes the +9.4% ChatDev gain to two different interventions in different sections.
- No MAST results on 2026-generation models were found; the newest model in the dataset is Claude-3.7-Sonnet.

## 2. "Towards a Science of Scaling Agent Systems" (Kim et al., Google and collaborators, Dec 2025) and related scaling results

### Takeaway
Across controlled comparisons, multi-agent coordination averaged a slightly negative effect (mean -3.5%, very wide spread): it helped a decomposable financial-analysis task by up to +80.8% and hurt a sequential planning task by 39 to 70% in every multi-agent topology. Independent agents amplified errors 17.2x versus 4.4x for a centralized orchestrator, and all topologies cost 1.6 to 6.2x the tokens.

### Cited Findings
- Paper identity: Yubin Kim, Ken Gu, Paul Pu Liang, Xin Liu and about 16 others; v1 9 Dec 2025, v2 17 Dec 2025, v3 8 Apr 2026. v3 abstract: 260 configurations, 6 benchmarks, 5 architectures (single agent plus Independent, Centralized, Decentralized, Hybrid), 3 LLM families. — [arXiv abstract](https://arxiv.org/abs/2512.08296)
- v3 abstract headline numbers: +80.8% on decomposable financial reasoning, -70.0% on sequential planning; predictive model cross-validated R² = 0.373 across six benchmarks (0.413 with a task-grounded capability metric); predicts the best architecture on 87% of held-out configurations. — [arXiv abstract](https://arxiv.org/abs/2512.08296)
- VERSION CONFLICT: the v2 full text lists four benchmarks (Finance-Agent, BrowseComp-Plus, PlanCraft, Workbench) and reports cross-validated R² = 0.524 (± 0.033), training R² = 0.613. v3 widened to six benchmarks and the R² fell to 0.373. All detailed numbers below are from v2. — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2); [arXiv abstract (v3)](https://arxiv.org/abs/2512.08296)
- Performance versus single agent (v2, measured; mean scores in brackets where given):

| Benchmark (task type) | Independent | Centralized | Decentralized | Hybrid |
|---|---|---|---|---|
| Finance-Agent (decomposable analysis; single agent 0.349) | not extracted | +80.8% (0.631) | +74.5% (0.609) | +73.1% (0.604) |
| Workbench (tool use; single agent 0.629) | not extracted | -1.2% | +5.7% (0.664) | -1.2% |
| BrowseComp-Plus (web research; single agent 0.318) | -35% | +0.2% | +9.2% (0.347) | not extracted |
| PlanCraft (sequential planning; single agent 0.568) | -70.1% (0.170) | -50.3% (0.282) | -41.5% (0.332) | -39.1% (0.346) |

  — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Overall mean multi-agent change: -3.5% (95% CI -18.6% to +25.7%, standard deviation 45.2%). — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Error amplification relative to single agent (1.0): Independent 17.2x, Decentralized 7.8x, Hybrid 5.1x, Centralized 4.4x. — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Token overhead versus single agent: Independent +58%, Decentralized +263%, Centralized +285%, Hybrid +515% (described as 1.6 to 6.2x). — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Mean turns: single agent 7.2, Independent 11.4, Decentralized 26.1, Centralized 27.7, Hybrid 44.3. — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Scaling law for turns against number of agents n: T = 2.72 × (n + 0.5)^1.724, R² = 0.974, exponent 95% CI [1.685, 1.763] (super-linear). — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Efficiency, successes per 1,000 tokens: single agent 67.7, Independent 42.4, Decentralized 23.9, Centralized 21.5, Hybrid 13.6. — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Capability saturation: coordination gives diminishing or negative returns once the single-agent baseline exceeds about 45% (interaction coefficient -0.404, p < 0.001). — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Tool-coordination trade-off: efficiency × tools coefficient -0.267, overhead × tools -0.162 (both p < 0.001); tool-heavy tasks suffer more from coordination overhead. — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Message density: success fit S = 0.73 + 0.28 ln(c), R² = 0.68; performance plateaus near 0.39 messages per turn. — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Architecture selection accuracy 87% on held-out configurations against 20% random and 54% capability-only; out-of-sample check on GPT-5.2 gave mean absolute error 0.071. — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Authors' statement, not a fitted optimum: beyond about 3 to 4 agents, per-agent reasoning capacity becomes "prohibitively thin" under a fixed budget. — [arXiv v2 HTML](https://arxiv.org/html/2512.08296v2)
- Earlier scaling claim (MacNet, Qian et al., ICLR 2025; submitted June 2024, GPT-3.5-era): agents arranged as a directed acyclic graph, collaboration among over a thousand agents, irregular topologies outperform regular ones, overall performance follows a logistic growth curve in the number of agents, and "collaborative emergence" appears earlier than neural emergence. — [arXiv 2406.07155](https://arxiv.org/abs/2406.07155)
- "More Agents Is All You Need" (Li et al., Feb 2024) studied sizes up to 40 samples with majority voting and reports monotone improvement with no stated plateau (see section 3). — [arXiv 2402.05120v2](https://arxiv.org/html/2402.05120v2)

### Inferences
- The clearest task rule in the data: multi-agent helps when the task splits into independent sub-analyses and the single agent is weak (Finance-Agent baseline 0.349), and hurts when steps are sequentially dependent (PlanCraft), where every topology lost 39% or more.
- A central orchestrator is an error filter: it cut amplification from 17.2x to 4.4x. But it is the costliest non-hybrid option in tokens (+285%) and still lost on planning.
- The exponent of about 1.72 means coordination turns grow almost quadratically with team size, which supports small teams under any fixed budget.
- The predictive rule is weaker than first advertised: R² dropped from 0.524 to 0.373 when two benchmarks were added. The 87% architecture-selection figure should be treated as in-distribution.
- MacNet's logistic growth and "more agents" voting gains were measured on older models and on answer-aggregation or artifact-refinement tasks; they are not evidence about tool-using agent teams.

### Gaps
- The two benchmarks added in v3, and the v3 per-benchmark numbers, were not extracted; only the v3 abstract was read.
- Whether the comparisons were strictly token-matched is not clear from what was read; the paper reports token overheads, which implies multi-agent arms used more tokens in at least part of the study. Treat the performance deltas as NOT confirmed budget-matched.
- Latency in wall-clock seconds was not found; only turn counts.
- MacNet's numeric results and the agent count at which its curve saturates were not extracted.

## 3. Multi-agent debate and "more agents" results, and the later budget-matched rebuttals

### Takeaway
The original debate and voting papers (2023 to early 2024, GPT-3.5 and Llama-2 era) measured real gains over a single direct answer, typically 5 to 15 points, but were not budget-matched. Later evaluations with matched budgets or stronger baselines found debate rarely beats chain-of-thought or self-consistency: five debate frameworks beat chain-of-thought in only about 6 to 22% of 36 model-benchmark settings.

### Cited Findings
- Du et al., "Improving Factuality and Reasoning in Language Models through Multiagent Debate" (May 2023; ChatGPT/GPT-3.5; 3 agents, 2 rounds; NOT budget-matched against the single agent). Table 1: arithmetic 67.0 single, 72.1 reflection, 69.0 multi-agent majority, 81.8 debate; GSM8K 77.0 / 75.0 / 81.0 / 85.0; chess pawn-score gain 91.4 / 102.1 / 102.2 / 122.9. Table 2: biographies 66.0 single, 68.3 reflection, 73.8 debate; MMLU 63.9 / 57.7 / 71.1; chess move validity 29.3 / 38.8 / 45.2. — [arXiv 2305.14325](https://arxiv.org/pdf/2305.14325)
- Du et al. limited the main evaluation to 3 agents and 2 rounds "due to computational expense" and acknowledge debate is more costly. On GSM8K the majority-vote baseline with the same agents already reached 81.0 against debate's 85.0. — [arXiv 2305.14325](https://arxiv.org/pdf/2305.14325)
- Li et al., "More Agents Is All You Need" (Feb 2024; sampling and majority voting, no interaction; NOT budget-matched by design). Single query to ensemble of 40: Llama2-13B GSM8K 0.35 to 0.59, MATH 0.03 to 0.09, MMLU 0.42 to 0.51; Llama2-70B GSM8K 0.54 to 0.74, MMLU 0.55 to 0.60, HumanEval 0.24 to 0.33; GPT-3.5-Turbo GSM8K 0.73 to 0.85, MATH 0.29 to 0.39, Chess 0.51 to 0.55, MMLU 0.59 to 0.70, HumanEval 0.67 to 0.73. Single GPT-4: 0.88 / 0.40 / 0.65 / 0.77 / 0.88 on the same five. — [arXiv 2402.05120v2](https://arxiv.org/html/2402.05120v2)
- Same paper: Llama2-13B reaches Llama2-70B's GSM8K accuracy at ensemble size 15; Llama2-70B and GPT-3.5-Turbo reach their stronger counterparts at sizes 15 and 20. Relative gains are larger on harder tasks and weaker models (GSM8K: 69% / 37% / 16%; MATH: 200% / 120% / 34% for Llama2-13B / Llama2-70B / GPT-3.5). — [arXiv 2402.05120v2](https://arxiv.org/html/2402.05120v2)
- Same paper, voting versus debate: plain sampling-and-voting (size 40) beat standalone debate (capped at 10 agents due to communication overhead) in most cells, for example GPT-3.5 HumanEval 0.73 voting against 0.18 debate, and Llama2 debate scored 0 on HumanEval, attributed by the authors to noise from other agents' answers. — [arXiv 2402.05120v2](https://arxiv.org/html/2402.05120v2)
- Huang et al., "Large Language Models Cannot Self-Correct Reasoning Yet" (Oct 2023): on GSM8K, multi-agent debate underperforms self-consistency with an equivalent number of responses; the authors attribute debate's gain to voting-like consistency. — [arXiv 2310.01798](https://arxiv.org/pdf/2310.01798)
- Wang et al., "Reasoning in Token Economies" (EMNLP 2024; GPT-3.5, GPT-4-0613, Mistral-7B, Llama-2-70b, Mixtral; GSM8K, MATH, TheoremQA, CSQA, HotpotQA, Game of 24; BUDGET-MATCHED at 20 queries or 10k tokens per question). Debate used 6 agents × 3 rounds (18 queries), Reflexion 19 queries. Self-consistency "regularly outperforms more complex strategies when given equivalent budgets" on all datasets except HotpotQA, where it ties. Debate's gains flatten at round two and its answer entropy falls each round. — [arXiv 2406.06461](https://arxiv.org/html/2406.06461v3); [ACL Anthology](https://aclanthology.org/2024.emnlp-main.1112.pdf)
- Zhang et al., "Stop Overvaluing Multi-Agent Debate" (Feb 2025, v3 June 2025; 5 debate frameworks: Society of Mind, Multi-Persona, Exchange-of-Thought, ChatEval, AgentVerse; 9 benchmarks; GPT-4o-mini, Claude-3.5-Haiku, Llama3.1-8b, Llama3.1-70b). Win / tie / lose against chain-of-thought over 36 settings: SoM 13.9 / 44.4 / 41.7%; Multi-Persona 0 / 22.2 / 77.8%; EoT 5.6 / 36.1 / 58.3%; ChatEval 11.1 / 61.1 / 27.8%; AgentVerse 13.9 / 38.9 / 47.2%. — [arXiv 2502.08788](https://arxiv.org/pdf/2502.08788)
- Same paper, GPT-4o-mini accuracies (CoT / self-consistency / best debate framework): MMLU 80.73 / 82.13 / 80.40 (AgentVerse); MMLU-Pro 62.80 / 66.27 / 62.80; GSM8K 93.60 / 95.67 / 94.93 (SoM); MATH 72.87 / 73.96 / 75.93 (EoT); HumanEval 78.05 / n.a. / 85.57 (AgentVerse, which adds a code-execution stage); MBPP 62.26 / n.a. / 58.88. Direct single answer without CoT: MMLU 65.33, HumanEval 66.67. — [arXiv 2502.08788](https://arxiv.org/pdf/2502.08788)
- Same paper: debate frameworks do beat a single agent prompted for a direct answer (34 of 45 GPT-4o-mini conditions), which reproduces the original papers' comparisons; the reversal comes from using CoT as the baseline. — [arXiv 2502.08788](https://arxiv.org/pdf/2502.08788)
- Same paper, mechanism: Multi-Persona, ChatEval and AgentVerse correct many wrong answers but also flip many initially correct answers to wrong; SoM and EoT are conservative and correct little. When tokens are increased, self-consistency keeps improving while debate either shows no stable trend or stays below self-consistency at comparable tokens. — [arXiv 2502.08788](https://arxiv.org/pdf/2502.08788)
- Same paper's position (authors' claim, with supporting experiments on GPT-4o-mini plus Llama3.1-70b): mixing different models in a debate consistently improves existing debate frameworks. — [arXiv 2502.08788](https://arxiv.org/html/2502.08788v3)
- "The Cost of Consensus" (arXiv 2605.00914, April 2026 preprint, seen only via a search summary): debate consumed 2.1 to 3.4x more tokens than self-correction (up to 28,631 tokens per problem) for equal or lower accuracy. Provisional. — [Emergent Mind summary of arXiv 2605.00914](https://www.emergentmind.com/papers/2605.00914)
- Gao et al. re-ran the Du-style math debate with Gemini-2.0-Flash: GSM8K single agent 93.8% to 94.0%, debate 93.6% to 94.6%, with debate using 8,837 normalized tokens against 526 for the single agent (about 16.8x). — [arXiv 2505.18286](https://arxiv.org/html/2505.18286v1)

### Inferences
- The original gains were measured against the weakest baseline (one direct answer). Against chain-of-thought or self-consistency with similar spend, the average debate effect is near zero or negative on 2024-era small and mid models.
- Debate's main mechanical weakness is correlated convergence: diversity collapses after round one (entropy falls), and persuasive wrong answers flip correct ones. Independent sampling plus voting preserves diversity, which is why it uses extra budget better.
- Heterogeneous models are the one variant with consistent support, which fits the diversity explanation.
- Voting gains shrink as the base model gets stronger (GPT-3.5 gained 16% relative on GSM8K against 69% for Llama2-13B), and near-saturated benchmarks leave no room.

### Gaps
- Exact self-consistency versus debate accuracies in Wang et al. are only in figures.
- The heterogeneous-debate numeric gains in Zhang et al. were not extracted.
- No budget-matched debate study on 2026 frontier reasoning models was read in full; the April 2026 "Cost of Consensus" preprint was seen only as a summary.

## 4. Single agent with tools versus multi-agent at equal budget; token and latency overhead

### Takeaway
Where budgets are matched, a single agent usually ties or wins (average 0.418 against 0.369 to 0.388 for multi-agent variants at a 1,000 thinking-token budget on multi-hop question answering). Measured token overhead of multi-agent systems ranges from about 1.6x to 6.2x in controlled agentic tasks, 2x to over 20x on coding and math pipelines, and about 15x chat usage in Anthropic's production research system.

### Cited Findings
- Tran and Kiela (Stanford), "Single-Agent LLMs Outperform Multi-Agent Systems on Multi-Hop Reasoning Under Equal Thinking Token Budgets" (arXiv 2604.02460, 2 Apr 2026; Qwen3-30B, DeepSeek-R1-70B, Gemini-2.5-Flash, Gemini-2.5-Pro; FRAMES and MuSiQue 4-hop; BUDGET-MATCHED on thinking tokens only). Average accuracy across datasets and models:

| Thinking budget | Single | Single (long prompt) | Sequential | Subtask | Roles | Debate | Ensemble |
|---|---|---|---|---|---|---|---|
| 100 | 0.290 | 0.337 | 0.364 | 0.322 | 0.363 | 0.370 | 0.280 |
| 500 | 0.390 | 0.366 | 0.376 | 0.342 | 0.365 | 0.380 | 0.310 |
| 1k | 0.418 | 0.397 | 0.379 | 0.369 | 0.381 | 0.388 | 0.333 |
| 2k | 0.421 | 0.420 | 0.389 | 0.383 | 0.398 | 0.403 | 0.372 |
| 5k | 0.427 | 0.425 | 0.386 | 0.396 | 0.417 | 0.420 | 0.411 |
| 10k | 0.426 | 0.424 | 0.387 | 0.399 | 0.423 | 0.420 | 0.420 |

  — [arXiv 2604.02460](https://arxiv.org/html/2604.02460v1)
- Same paper, exceptions: at a 100-token budget multi-agent variants lead in most rows (the authors treat that budget as too small for useful reasoning); Gemini-2.5-Pro on MuSiQue at 1k shows Debate 0.444 and Roles 0.430 above the single agent's 0.413. Under heavy context corruption (masking or substitution at 0.7) the Sequential multi-agent variant beats the single agent; under deletion and distractors the single agent stays ahead. Those corruption results are reported qualitatively from a figure. — [arXiv 2604.02460](https://arxiv.org/html/2604.02460v1)
- Same paper's theoretical argument (authors' claim): by the data processing inequality, with a fixed budget and perfect context use a single agent is at least as information-efficient, and splitting context across agents adds communication bottlenecks. — [arXiv 2604.02460](https://arxiv.org/abs/2604.02460)
- Gao et al. (UIUC), "Single-agent or Multi-agent Systems? Why Not Both?" (arXiv 2505.18286, May 2025). Multi-agent advantage shrinks with newer models (NOT budget-matched): MetaGPT on HumanEval 67.0% single to 87.7% multi with ChatGPT, against 90.2% to 93.2% with Gemini-2.0-Flash; Self-Collaboration on MBPP 52.2% to 68.2% against 79.6% to 80.8%; math debate on GSM8K 77% to 85% against 93.8% to 94.6%. — [arXiv 2505.18286](https://arxiv.org/html/2505.18286v1)
- Same paper, multi-agent over single-agent token multiples (prefill / decode): MBPP 5.05x / 5.56x; HumanEval 5.14x / 3.51x; DS1000 4.42x / 8.97x; BigCodeBench 8.65x / 2.10x; BigCodeBench-Hard 14.82x / 2.51x; GSM8K 34.66x / 12.77x; AIME 220.22x / 11.05x; FinRobot 9.00x / 21.64x; HoVer 1.20x / 2.94x. — [arXiv 2505.18286](https://arxiv.org/html/2505.18286v1)
- Same paper, Gemini-2.0-Flash accuracy single / multi: GSM8K 94.0 / 93.6; HumanEval 90.2 / 93.3; MBPP 79.6 / 80.8; DS1000 62.9 / 62.3; BigCodeBench 47.4 / 46.0; BigCodeBench-Hard 22.3 / 23.0; AIME 25.0 / 38.3 (the one large multi-agent win, at about 20x the tokens: 81,159 against 4,010). — [arXiv 2505.18286](https://arxiv.org/html/2505.18286v1)
- Same paper: a cascade that tries the single agent first and escalates to multi-agent beat both on most tasks (for example DS1000 71.2 against 62.9 and 62.3; BigCodeBench-Hard 32.4 against 22.3 and 23.0) and cut tokens against always-multi-agent by about 3% to 88%. On 30 human-annotated software tasks with Gemini-2.5-Pro, the single agent won 15 against ChatDev (ChatDev won 1) and 13 against MetaGPT (MetaGPT won 0). — [arXiv 2505.18286](https://arxiv.org/html/2505.18286v1)
- Kim et al. scaling study: token overhead +58% (independent) to +515% (hybrid); successes per 1,000 tokens fall from 67.7 (single) to 13.6 (hybrid); turns rise from 7.2 to 44.3. — [arXiv 2512.08296v2](https://arxiv.org/html/2512.08296v2)
- Anthropic engineering post (13 June 2025; Claude Opus 4 lead with Claude Sonnet 4 subagents; internal research eval; NOT budget-matched, vendor self-report): multi-agent beat single-agent Opus 4 by 90.2%; agents use about 4x the tokens of chat and multi-agent systems about 15x; on BrowseComp, token usage alone explained 80% of performance variance and tokens plus tool-call count plus model choice explained 95%; parallel tool calling cut research time for complex queries by up to 90%. — [Anthropic engineering](https://www.anthropic.com/engineering/multi-agent-research-system)
- Same post names poor fits: tasks where all agents need the same context or have many inter-agent dependencies, and most coding tasks, which have fewer truly parallelizable parts than research. — [Anthropic engineering](https://www.anthropic.com/engineering/multi-agent-research-system)
- Xu et al., "Rethinking the Value of Multi-Agent Workflow: A Strong Single Agent Baseline" (arXiv 2601.12307, Jan 2026; 7 benchmarks across coding, math, QA, domain reasoning, planning and tool use): one LLM playing all roles in a multi-turn conversation with KV-cache reuse matches homogeneous multi-agent workflows at lower cost; it cannot reproduce heterogeneous-model workflows. — [arXiv 2601.12307](https://arxiv.org/abs/2601.12307)
- Wang et al. budget-aware study: ToT on Game of 24 with GPT-4 reached 76% at $159.87 per 100 instances; a cheaper GPT-3.5 evaluator gave 72% at $33.53. — [arXiv 2406.06461](https://arxiv.org/html/2406.06461v3)

### Inferences
- Anthropic's own data supports the budget explanation: if tokens explain 80% of variance, a large part of the 90.2% gain is extra tokens and parallel context windows, not coordination as such. Multi-agent there is a way to spend more tokens in parallel within context limits, and to cut wall-clock time.
- Latency is the one axis where multi-agent has a clear measured benefit (up to 90% less research time through parallelism), but only for breadth-first, independent subtasks.
- The cascade result suggests the practical default: single agent first, escalate only on hard or failed cases. Easy items gain nothing and pay 5 to 35x.
- Multi-agent's residual advantage appears where a single context is unreliable (corrupted or overloaded context) or where the single agent is far from saturation on a hard task (AIME 25.0 to 38.3).

### Gaps
- No controlled wall-clock latency comparison of single against multi-agent at equal token budget was found; turn counts and one vendor figure are the only evidence.
- No budget-matched comparison on SWE-bench, GAIA, WebArena or AssistantBench was found in this pass; leaderboard entries on those benchmarks mix models, scaffolds and budgets.
- Gao et al. has inconsistencies: the stated MetaGPT gain of "10.7%" does not match 67.0 to 87.7, and the text's token ranges do not match its own table. Their summary cost claims ("up to 20%" and "88.1%") also conflict.
- Tran and Kiela match thinking tokens only, not total tokens, and cover multi-hop QA only, with no tool use.
- The Xu et al. numbers were not extracted.

## 5. Verification and review agents, error propagation, and information loss in handoffs

### Takeaway
Evidence that reviewer or judge agents reliably catch errors is thin and mixed: an added verification step gave +15.6 points in one case study, but LLM verifiers showed 22.5 to 70.4% false-positive rates on matched benign traces, judges can replace correct answers with poisoned ones, and automated failure attribution finds the decisive step only 14.2% of the time. Unchecked parallel agents amplify errors most (17.2x); a central checker reduces but does not remove amplification (4.4x).

### Cited Findings
- MAST: verification failures are 21.3% (v2) to about 23.5% (v3, summed) of labelled failures; "incorrect verification" alone is 6.66% to 9.10% and "no or incomplete verification" 6.82% to 8.20%. Adding a high-level verification step to ChatDev raised ProgramDev success from 25.0 to 40.6. — [arXiv 2503.13657v2](https://arxiv.org/html/2503.13657v2); [arXiv 2503.13657v3](https://arxiv.org/html/2503.13657v3)
- Kim et al.: error amplification 17.2x for independent agents without cross-checking against 4.4x with a centralized orchestrator, 5.1x hybrid, 7.8x decentralized. — [arXiv 2512.08296v2](https://arxiv.org/html/2512.08296v2)
- Wang et al. (2024), self-evaluation accuracy of GPT-4-0613 as a judge of answers: on GSM8K it accepts correct answers 99.2% of the time but identifies incorrect ones only 15.6%; MATH 91.1% / 46.1%; TheoremQA 94.5% / 23.2%; HotpotQA 99.4% / 2.9%; CSQA 98.7% / 6.0%. In Tree-of-Thoughts, replacing the evaluator with a random one dropped top-1 score to 0.0, so evaluator quality is decisive. — [arXiv 2406.06461](https://arxiv.org/html/2406.06461v3)
- "Watching the Watchers" (Woo, NeurIPS 2026; fault injection in a simulated procurement and finance workflow, 8 injected violation types each with a matched benign twin, 5,760 LLM evaluations over three open-weight model families): deterministic event-local rules reach AUROC 0.95 to 0.98 on local violations but are at chance on cross-event relations; trace-level predicates reach 0.975 to 0.978 on three cross-event relations; the LLM verifier's false-positive rate on matched benign traces was 22.5 to 70.4%; value claims could not be resolved without an external authoritative source. — [NeurIPS 2026 virtual page](https://neurips.cc/virtual/2026/172886)
- "When the Judge Intervenes" (NeurIPS 2026, seen only as a search summary): in a Planner to Worker to Judge pipeline with poisoned evidence, every revision the GPT judge made replaced a correct answer with a poisoned one. Provisional. — [NeurIPS 2026 virtual page](https://neurips.cc/virtual/2026/173019)
- Who&When (ICML 2025): best automated attribution names the faulty agent 53.5% of the time and the decisive step 14.2%. — [arXiv 2505.00212](https://arxiv.org/abs/2505.00212)
- Zhang et al. (debate): judge-based and aggressive debate frameworks (Multi-Persona, ChatEval, AgentVerse) frequently flip initially correct answers to incorrect, offsetting their corrections. — [arXiv 2502.08788](https://arxiv.org/pdf/2502.08788)
- "Hallucination Cascade" (Jamshidi et al., arXiv 2606.07937, June 2026; GPT-5.3, DeepSeek-V3, LLaMA-3-70B; 500 cascades over 10 domains, 1,250 responses, 3-agent sequential refinement chains): normalized hallucination score fell from 0.422 at the first agent to 0.272 at the last (amplification factor 0.644, net attenuation), about 0.072 reduction per hop, while factual accuracy fell slightly from 0.789 to 0.769. — [arXiv 2606.07937](https://arxiv.org/abs/2606.07937)
- Gao et al. failure analysis: (a) performance is capped by the critical agent, with about 80% of examples either passed by both single and multi-agent or failed by both; (b) downstream agents get overloaded, and early-stopping experiments found only about 25% of upstream messages helped final accuracy; (c) information is lost in summarization, for example a correct intermediate answer in a math debate was dropped and later agents reinforced a wrong one. — [arXiv 2505.18286](https://arxiv.org/html/2505.18286v1)
- Anthropic: passing results through the coordinator creates a "game of telephone"; their mitigation is for subagents to write outputs to external storage and return lightweight references, which preserves fidelity and reduces token copying. Observed production failures included spawning up to 50 subagents for simple queries, endless searching for nonexistent sources, and subagents duplicating each other's work when task descriptions were vague. — [Anthropic engineering](https://www.anthropic.com/engineering/multi-agent-research-system)
- Code review tools (mostly vendor-run benchmarks, low independence): reported precision about 65 to 76% and recall about 51 to 52% for two commercial reviewers; one vendor reports false-positive rates of 22.5% for itself against 41 to 46% for competitors; a Kodus experiment raised recall from about 53% to 62% at the cost of false positives rising from 170 to 328. — [Softwareseni comparison](https://www.softwareseni.com/how-ai-code-review-tools-compare-on-real-bug-detection-and-noise); [kluster.ai benchmark](https://kluster.ai/blog/code-review-benchmark); [Kodus experiments](https://kodus.io/en/ai-code-review-recall/)
- MetaGPT ablation: adding an executable feedback step (run the code, feed errors back) improved Pass@1 by 4.2 points on HumanEval and 5.4 on MBPP, raised executability from 3.67 to 3.75 and cut human revision cost from 2.25 to 0.83. — [MetaGPT ICLR 2024 slides](https://iclr.cc/media/iclr-2024/Slides/19756_GyeFksR.pdf); [arXiv 2308.00352](https://arxiv.org/html/2308.00352v7)

### Inferences
- LLM judges are lenient by default: very high acceptance of correct answers and poor rejection of wrong ones (2.9 to 46% on 2023 GPT-4). When prompted to hunt for violations they swing to high false-positive rates (22.5 to 70.4%). Neither mode is a reliable gate on its own.
- Verification grounded in something external (code execution, deterministic rules, an authoritative data source) has the strongest measured effect: MetaGPT executable feedback, AgentVerse's execution stage on HumanEval, rule-based AUROC of 0.95 or more. Opinion-only review has the weakest.
- Rule-based checks and LLM checks fail in complementary places: rules miss cross-event relations unless written at trace level; LLM verifiers cover more but with many false alarms.
- Sequential refinement dampens hallucination modestly (about 36% over three hops) but loses a little true content each hop, consistent with summaries being lossy in both directions.
- Three separate sources report information loss or overload at handoffs (Gao path-level defects, Anthropic telephone effect, MAST loss of history and withholding). Only Gao gives a number (about 25% of upstream messages useful).

### Gaps
- No peer-reviewed study was found that reports both the catch rate and the false-positive rate of an independent reviewer agent on realistic multi-agent coding or research work with frontier 2026 models. The code-review figures are vendor benchmarks.
- "Watching the Watchers" true-positive rates for the LLM verifier were not on the page read.
- No quantitative measurement of information loss per handoff or per summary (for example, fraction of facts retained) was found.
- The "When the Judge Intervenes" finding was not verified against the paper.
- Whether a different-model reviewer catches more than a same-model reviewer was not found as a direct measurement; the only support is the heterogeneous-debate result.

## 6. Which communication structure works best: structured artifacts against free chat, orchestrator against peer-to-peer, and topology search

### Takeaway
Direct head-to-head evidence is limited. The best-controlled result is that a centralized orchestrator contains errors better than peers or independent agents (4.4x against 7.8x and 17.2x amplification) but no topology wins across tasks. Structured artifacts and executable feedback (MetaGPT) beat chat-style role play (ChatDev) in the authors' own evaluation, and automated topology search mainly buys lower cost for similar accuracy, with gains that a strong single agent can often match.

### Cited Findings
- Kim et al.: best topology depends on the task. Centralized was best on Finance-Agent (+80.8%), decentralized on BrowseComp-Plus (+9.2%) and Workbench (+5.7%), and all were worse than a single agent on PlanCraft. Independent agents were worst on error amplification (17.2x) and hybrid worst on cost (+515% tokens, 44.3 turns). — [arXiv 2512.08296v2](https://arxiv.org/html/2512.08296v2)
- MetaGPT (Hong et al., ICLR 2024; GPT-4 era; authors' own evaluation): agents exchange structured documents (requirements, design, interface specs) under a standard operating procedure. Reported 85.9% and 87.7% Pass@1 on HumanEval and MBPP. On its SoftwareDev benchmark, executability 3.75 of 4 against 2.25 for ChatDev and human revision cost 0.83 against 2.5 (the ChatDev comparison figures come from a secondary overview). — [arXiv 2308.00352](https://arxiv.org/html/2308.00352v7); [alphaXiv overview](https://www.alphaxiv.org/overview/2308.00352)
- Later re-test with a stronger base model: MetaGPT's HumanEval margin over a single agent shrank from 67.0 against 87.7 (ChatGPT) to 90.2 against 93.2 (Gemini-2.0-Flash); a Gemini-2.5-Pro single agent beat MetaGPT 13 to 0 on 30 human-annotated software tasks. — [arXiv 2505.18286](https://arxiv.org/html/2505.18286v1)
- AgentVerse in an independent evaluation: best of five debate frameworks on HumanEval with GPT-4o-mini (85.57 against CoT 78.05) because it includes an execution-evaluation stage, but below CoT on MATH (64.49 against 72.87) and MBPP (58.88 against 62.26), and it beat CoT in only 13.9% of 36 settings. — [arXiv 2502.08788](https://arxiv.org/pdf/2502.08788)
- MacNet (ICLR 2025): directed-acyclic-graph topologies; irregular topologies outperform regular ones; logistic growth in performance with agent count. — [arXiv 2406.07155](https://arxiv.org/abs/2406.07155)
- Topology and workflow search, cost comparison on MATH from the MaAS paper (ICML 2025; authors' own table): AFlow training cost $22.50, inference $1.66, accuracy 51.28%; MaAS training $3.38, inference $0.42, accuracy 51.82%; DyLAN training $13.01 and 508 minutes; GPTSwarm training $7.02 and 129 minutes; MaAS 53 minutes. — [arXiv 2502.04180](https://arxiv.org/pdf/2502.04180)
- AFlow (authors' claim): workflows found for GPT-4o-mini outperform GPT-4o on HumanEval (94.7% against 93.9%) at 4.55% of the inference cost. — [alphaXiv overview of arXiv 2410.10762](https://www.alphaxiv.org/overview/2410.10762)
- Design difference among search methods: ADAS and AFlow represent workflows as code (ADAS heuristic search, AFlow Monte Carlo tree search); MaAS samples a query-specific system from an "agentic supernet". — [arXiv 2502.04180](https://arxiv.org/pdf/2502.04180)
- Xu et al. (Jan 2026): using AFlow-discovered workflows, a single LLM executing the same workflow in one multi-turn conversation matched the multi-agent version at lower cost, when all agents share the same base model. — [arXiv 2601.12307](https://arxiv.org/abs/2601.12307)
- Anthropic production guidance (vendor report): orchestrator-worker with 3 to 5 parallel subagents; each subagent needs an objective, output format, tool guidance and clear task boundaries or it duplicates work; subagents should write to shared storage and return references instead of relaying through chat. — [Anthropic engineering](https://www.anthropic.com/engineering/multi-agent-research-system)
- MAST intervention: changing ChatDev's or AG2's topology gave gains comparable to prompt fixes (AG2 GPT-4o 84.25 to 88.83; ChatDev ProgramDev 25.0 to 40.6) and was not significant for AG2 with GPT-4. — [arXiv 2503.13657v2](https://arxiv.org/html/2503.13657v2)

### Inferences
- Across sources, what is passed matters more than who talks to whom: executable or checkable artifacts (code runs, structured specs, file references) show the most consistent benefits, while free-form multi-round chat shows diversity collapse and answer flipping.
- A centralized orchestrator is the safer default for error containment; peer-to-peer did slightly better on exploratory web and tool tasks, but the differences there (+5.7%, +9.2%) are small next to the variance (standard deviation 45.2%).
- Searched topologies (AFlow, MaAS, ADAS) differ little in accuracy from each other on the reported benchmark (51.28% against 51.82% on MATH); the measured differentiator is search and inference cost. Xu et al. suggest much of the benefit is the workflow (the sequence of steps), not the separation into agents.
- Almost all framework-level claims come from 2023 to 2024 base models; the re-tests available show the margins compress sharply on 2025 models.

### Gaps
- No controlled ablation was found that holds the agent team fixed and varies only the message format (structured artifact against free-form chat). MetaGPT against ChatDev differs in many other ways.
- GPTSwarm, DyLAN and ADAS accuracy numbers on their own benchmarks, and GAIA results for searched topologies, were not reliably extracted; one follow-up table gave conflicting GAIA figures for ADAS and AFlow and was left out.
- No multi-agent against single-agent comparison on GAIA, SWE-bench, WebArena or AssistantBench with matched model and budget was found in this pass.
- No evidence from late 2026 on whether frontier models change the centralized against decentralized ranking; the scaling study's out-of-sample check covered GPT-5.2 only as a prediction-error figure (mean absolute error 0.071).
