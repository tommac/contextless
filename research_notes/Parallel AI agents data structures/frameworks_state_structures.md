# Data structures for splitting work and sharing state in open multi-agent LLM frameworks (2023 to 2026)

Research date: 2026-11-17. Documentation pages were read as they stood on that date; where a page states a version, it is given. Paper facts are tied to the arXiv version read. "Measured" means a number reported in a paper or evaluation; "claimed" means stated in docs or announcements without a measurement.

## 1. What is the core data structure in each framework, and is execution truly concurrent?

### Takeaway
Three families exist. (a) Turn-taking over a shared transcript (ChatDev, AutoGen 0.2 group chat, AutoGen AgentChat teams, Magentic-One, Swarm, OpenAI Agents SDK, CAMEL role-playing, CrewAI sequential): one agent acts at a time and the "shared state" is the message history, plus a ledger or context dict in some. (b) Pregel-style superstep graphs (LangGraph, Microsoft Agent Framework workflows): nodes triggered in the same superstep run concurrently behind a synchronization barrier, with typed state channels or buffered state updates. (c) Event or message-driven runtimes (AutoGen Core, LlamaIndex Workflows, Google ADK ParallelAgent, MetaGPT's pool): concurrency is available but the framework gives weaker or no ordering guarantees. In every case found, "concurrent" means concurrent LLM/tool calls (asyncio-style), not parallel computation over shared memory.

### Cited Findings

**MetaGPT (paper v7, Nov 2024; first version Aug 2023)**
- Structure: a global shared message pool with publish/subscribe. "we introduce a shared message pool that allows all agents to exchange messages directly"; roles "select information to follow based on their role profiles" — [MetaGPT paper v7](https://arxiv.org/html/2308.00352v7)
- Messages are structured artifacts, not chat: "agents in MetaGPT communicate through documents and diagrams (structured outputs) rather than dialogue." The chain is PRD (Product Manager) to system design (Architect) to task list (Project Manager) to code (Engineer) to tests (QA Engineer) — [MetaGPT paper v7](https://arxiv.org/html/2308.00352v7)
- Activation rule is dependency based: "an agent activates its action only after receiving all its prerequisite dependencies" — [MetaGPT paper v7](https://arxiv.org/html/2308.00352v7)
- Execution in the paper's SOP is sequential: SOPs "enables all agents to work in a sequential manner"; the abstract calls it an "assembly line paradigm" — [MetaGPT paper v7](https://arxiv.org/html/2308.00352v7), [arXiv abstract](https://arxiv.org/abs/2308.00352)
- In code, subscription is by the Action that caused a message: a Role calls `self._watch([UserRequirement])` and "If there is a `Message` caused by the particular `Action`s the `Role` `_watch`, then it is a valid observation"; messages carry `cause_by`; output is "encapsulated in a `Message` to be finally `publish_message` to the `Environment`"; the team runs in rounds via `team.run(n_round=...)` — [MetaGPT docs, Multi-Agent 101](https://docs.deepwisdom.ai/main/en/guide/tutorials/multi_agent_101.html)

**ChatDev (paper v5, Jun 2024; first version Jul 2023)**
- Structure: a "chat chain" of three sequential phases (design, coding, testing) split into five sequential subtasks; each subtask is a two-agent dialogue between an instructor and an assistant — [ChatDev paper v5](https://arxiv.org/html/2307.07924)
- State passed between phases is only the solution: the chain "only transmits the solutions from previous phases as long-term memories"; "Short-term memory is utilized to sustain the continuity of the dialogue within a single phase" — [ChatDev paper v5](https://arxiv.org/html/2307.07924)
- Execution is strict turn taking; a subtask ends after two unchanged code modifications or 10 rounds. No parallel execution is described — [ChatDev paper v5](https://arxiv.org/html/2307.07924)

**AutoGen 0.2 (paper v2, Oct 2023)**
- Structure: "conversable agents" that each maintain "internal context based on sent and received messages"; control flow is the conversation itself via auto-reply: "Once an agent receives a message from another agent, it automatically invokes generate_reply and sends the reply back to the sender unless a termination condition is satisfied" — [AutoGen paper v2](https://arxiv.org/html/2308.08155)
- Group chat: a GroupChatManager repeats three steps: select a speaker, collect that speaker's response, broadcast it to the other agents; agents "share the same context" — [AutoGen paper v2](https://arxiv.org/html/2308.08155)

**AutoGen 0.4 Core and AgentChat (released Jan 2025)**
- Core is "a scalable, event-driven actor framework for creating agentic workflows"; "Agents now communicate through asynchronous messages, supporting both event-driven and request/response interaction patterns" — [AutoGen 0.4 launch post, 17 Jan 2025](https://devblogs.microsoft.com/autogen/autogen-reimagined-launching-autogen-0-4/)
- "At the base is AutoGen Core. This layer implements the actor model for agents"; the design "decouples how the messages are delivered between the agents from how the agents handle them" — [Microsoft Research article, 25 Feb 2025](https://www.microsoft.com/en-us/research/articles/autogen-v0-4-reimagining-the-foundation-of-agentic-ai-for-scale-extensibility-and-robustness/)
- An agent is "a software entity that communicates via messages, maintains its own state"; state is private to the agent — [AutoGen Core concepts](https://microsoft.github.io/autogen/stable/user-guide/core-user-guide/core-concepts/agent-and-multi-agent-application.html)
- Two delivery modes: "direct messaging or broadcast". A topic is (type, source); "A subscription maps topic to agent IDs"; with type-based subscription the runtime derives the agent ID from agent type plus topic source and "If the agent with the ID does not exist, the runtime will create it" — [Topic and Subscription](https://microsoft.github.io/autogen/stable/user-guide/core-user-guide/core-concepts/topic-and-subscription.html)
- Runtimes: `SingleThreadedAgentRuntime` for a single process; a distributed runtime that "consists of a host servicer and multiple workers" — [Agent Runtime Environments](https://microsoft.github.io/autogen/stable/user-guide/core-user-guide/core-concepts/architecture.html)
- AgentChat teams built on Core are still turn taking: in RoundRobinGroupChat "all agents share the same context and take turns responding in a round-robin fashion" and "Each agent, during its turn, broadcasts its response to all other agents"; SelectorGroupChat "selects the next speaker using a ChatCompletion model after each message"; Swarm "uses HandoffMessage to signal transitions between agents" — [AgentChat Teams tutorial](https://microsoft.github.io/autogen/stable/user-guide/agentchat-user-guide/tutorial/teams.html)

**Magentic-One (paper Nov 2024, built on AutoGen 0.4)**
- Structure: an Orchestrator with two ledgers. The task ledger (outer loop) holds given or verified facts, facts to look up, facts to derive, educated guesses, and a step-by-step plan. The progress ledger (inner loop) answers five questions each iteration: is the request satisfied, is the team looping, is forward progress being made, which agent speaks next, what instruction to give it — [Magentic-One paper](https://arxiv.org/html/2411.04468)
- A stall counter increments on loops or lack of progress; when it exceeds the threshold (2 in the experiments) the Orchestrator leaves the inner loop, reflects, and rewrites the task ledger and plan. Agents clear their contexts after each plan update — [Magentic-One paper](https://arxiv.org/html/2411.04468)
- One next speaker and one instruction are chosen per inner-loop iteration, so workers take turns — [Magentic-One paper](https://arxiv.org/html/2411.04468)

**Microsoft Agent Framework (GA 1.0 April 2026; docs dated Sep to Oct 2026)**
- Structure: a directed graph of typed executors connected by edges; "The framework uses a modified Pregel execution model — a Bulk Synchronous Parallel (BSP) approach with superstep-based processing" — [MAF Workflow Builder and Execution](https://learn.microsoft.com/en-us/agent-framework/concepts/workflows/builder-and-execution)
- Each superstep collects pending messages, routes them, "Runs all target executors concurrently within the superstep", "Waits for all executors to complete before advancing (synchronization barrier)", and queues new messages for the next superstep — [MAF Workflow Builder and Execution](https://learn.microsoft.com/en-us/agent-framework/concepts/workflows/builder-and-execution)
- Consequence of the barrier: with a fan-out to a chain and to one long-running executor, "the chained path cannot advance until the long-running executor completes" — [MAF Workflow Builder and Execution](https://learn.microsoft.com/en-us/agent-framework/concepts/workflows/builder-and-execution)
- Python also has an experimental functional API where parallelism is plain `asyncio.gather`; prebuilt orchestrations are "sequential, concurrent, handoff, group-chat, and Magentic patterns" — [MAF Workflow concepts](https://learn.microsoft.com/en-us/agent-framework/concepts/workflows/), [MAF Workflow capabilities](https://learn.microsoft.com/en-us/agent-framework/workflows/)

**LangGraph (docs read Nov 2026; version 1.x)**
- Structure: a state graph whose state is a set of channels, one per key, each optionally with a reducer. "LangGraph runs graphs with message passing, in the style of Google's Pregel system" — [LangGraph Graph API](https://docs.langchain.com/oss/python/langgraph/graph-api)
- Concurrency unit is the superstep: "Nodes that run in parallel share one super-step"; "A node becomes active when it receives input on an incoming edge"; "The run ends when every node is inactive and no messages remain in transit" — [LangGraph Graph API](https://docs.langchain.com/oss/python/langgraph/graph-api)
- Map-reduce fan-out: "Return one or more Send objects from a conditional edge"; each Send names a target node and carries its own input; "LangGraph runs those invocations in parallel"; reducers merge the workers' updates — [LangGraph Graph API](https://docs.langchain.com/oss/python/langgraph/graph-api)
- `Command` combines a state update and routing in one return: "Use Command when a single step must both update state and choose what runs next" — [LangGraph Graph API](https://docs.langchain.com/oss/python/langgraph/graph-api)
- Recursion limit: "Starting in version 1.1.4, the default is 10,007 super-steps"; "Between 1.0.6 and 1.1.3, the default is 10,000 steps" — [LangGraph Graph API](https://docs.langchain.com/oss/python/langgraph/graph-api)

**CrewAI (docs read Nov 2026)**
- Crews: two processes. Sequential "Executes tasks sequentially"; hierarchical "Organizes tasks in a managerial hierarchy" and needs `manager_llm` or `manager_agent`. The `Process` enum is restricted to `sequential` and `hierarchical`; no consensual process is listed — [CrewAI Processes](https://docs.crewai.com/concepts/processes)
- Data passed between tasks is a `TaskOutput` (`raw`, `pydantic`, `json_dict`, `agent`, `messages`, `summary`); "the output of one task is automatically relayed into the next one" and `context` selects which earlier outputs a task receives — [CrewAI Tasks](https://docs.crewai.com/concepts/tasks)
- Limited concurrency via `async_execution`: "the crew will not wait for it to be completed to continue with the next task"; a later task that lists it in `context` waits for it — [CrewAI Tasks](https://docs.crewai.com/concepts/tasks)
- Flows: an event-style method graph (`@start`, `@listen`, `@router`, `or_`, `and_`) over one state object, either an unstructured dict or a Pydantic model; "Each flow state automatically gets a unique ID"; "All satisfied `@start()` methods will execute (often in parallel) when the Flow begins or resumes" — [CrewAI Flows](https://docs.crewai.com/concepts/flows)

**OpenAI Swarm (Oct 2024, experimental) and OpenAI Agents SDK (Mar 2025 onward)**
- Swarm: two primitives, Agents and handoffs; "An `Agent` can hand off to another `Agent` by returning it in a `function`". Shared state is `context_variables`, "A dictionary of additional context variables, available to functions and Agent instructions", updatable by returning a `Result`. The loop is sequential and tool calls run in order: "If multiple functions are called by the `Agent`, they will be executed in that order". It "is entirely powered by the Chat Completions API and is hence stateless between calls" — [Swarm README](https://github.com/openai/swarm)
- Agents SDK: "Handoffs are represented as tools to the LLM" (for example `transfer_to_refund_agent`); the receiving agent takes over the conversation "and gets to see the entire previous conversation history", unless an `input_filter` returns a modified `HandoffInputData` — [Agents SDK Handoffs](https://openai.github.io/openai-agents-python/handoffs/)
- Loop: "If the LLM requests a handoff, we update the current agent and input, and re-run the loop"; "If the LLM produces tool calls, we run those tool calls, append the results, and re-run the loop"; `MaxTurnsExceeded` bounds it. Concurrency exists only at the tool level: `parallel_tool_calls` and `tool_execution.max_function_tool_concurrency` — [Agents SDK Running agents](https://openai.github.io/openai-agents-python/running_agents/)

**AgentVerse (paper v3, Oct 2023)**
- Structure: a four-stage loop: Expert Recruitment, Collaborative Decision-Making, Action Execution, Evaluation; "If the goal remains unmet, the feedback returns to the initial expert recruitment stage". Decision-making is either horizontal (each agent refines, a function integrates) or vertical (one solver, several reviewers). Shared state is an environment state updated as s_new = T(s_old, A) — [AgentVerse paper v3](https://arxiv.org/html/2308.10848)
- The paper does not state that action execution is parallel — [AgentVerse paper v3](https://arxiv.org/html/2308.10848)

**CAMEL (paper Mar 2023, NeurIPS 2023; Workforce docs read Nov 2026)**
- Original structure: "a novel communicative agent framework named role-playing" using "inception prompting to guide chat agents toward task completion"; a two-party turn-taking dialogue — [CAMEL abstract](https://arxiv.org/abs/2303.17760)
- Later Workforce module: a `task_agent` "breaks the main task into smaller, self-contained subtasks", a `coordinator_agent` "assigns each subtask to the most suitable worker", "Workers execute their assigned tasks, often in parallel", and "A task's result is stored and can be used as a dependency for other tasks". Worker memory is isolated by default: "`share_memory`: If `True`, `SingleAgentWorker` instances will share memory" — [CAMEL Workforce docs](https://docs.camel-ai.org/key_modules/workforce)

**Google ADK (docs read Nov 2026, ADK 2.x)**
- ParallelAgent "initiates the run_async() method of each sub-agent present in the sub_agents list concurrently"; "Each sub-agent operates in its own execution branch" with "no automatic sharing of conversation history between these branches" — [ADK Parallel agents](https://adk.dev/agents/workflow-agents/parallel-agents/)
- The shared structure is session state: "Session state is shared: all branches run against the same session"; results are handed over by giving each sub-agent an `output_key`. "Because the branches interleave, the event order is not deterministic" — [ADK Parallel agents](https://adk.dev/agents/workflow-agents/parallel-agents/)
- `session.state` "is a collection (dictionary or Map) holding key-value pairs" with scope prefixes: none (session), `user:`, `app:`, `temp:` (one invocation). Updates travel as `EventActions.state_delta` on events applied by `append_event` — [ADK State](https://adk.dev/sessions/state/)

**LlamaIndex Workflows (docs read Nov 2026)**
- "A workflow is an event-driven, step-based way to control the execution flow of an application"; each step "receives an event, does some work, and returns another event" — [LlamaIndex Workflows](https://developers.llamaindex.ai/python/llamaagents/workflows/)
- Fan-out and fan-in: "Concurrent work is a step that returns `list[Event]`, paired with another step that accepts `list[Event]`"; `num_workers` "controls how many copies of that step can process events at once"; results arrive "in completion order". `ctx.send_event` emits "the moment you call it"; `ctx.collect_events` "returns `None` until all three have arrived" — [LlamaIndex concurrent execution](https://developers.llamaindex.ai/python/llamaagents/workflows/concurrent_execution/)
- Shared state is a per-run store reached through `ctx.store` (get/set), optionally a typed Pydantic model — [LlamaIndex managing state](https://developers.llamaindex.ai/python/llamaagents/workflows/managing_state)

### Inferences
- Summary table (derived from the findings above):

| Framework | Plan carrier | Shared state | Message carrier | Truly concurrent agents? |
|---|---|---|---|---|
| MetaGPT | Fixed SOP; role subscriptions | Shared message pool of structured documents | Pub/sub `Message` with `cause_by` | No in the paper's SOP (sequential) |
| ChatDev | Fixed chat chain | Phase solutions only | Two-agent dialogue | No |
| AutoGen 0.2 | Emergent from conversation; speaker selection | Shared group transcript | Chat messages, broadcast by manager | No |
| AutoGen 0.4 Core | Application-defined | None shared; per-agent private state | Typed async messages, direct or topic broadcast | Runtime allows it; AgentChat teams take turns |
| Magentic-One | Task ledger | Task ledger + progress ledger + transcript | Orchestrator instruction per turn | No |
| MS Agent Framework | Executor graph | Scoped workflow state, buffered per superstep | Typed messages along edges | Yes, within a superstep |
| LangGraph | Graph edges, `Send`, `Command` | Channels with reducers | State updates (writes to channels) | Yes, within a superstep |
| CrewAI Crew | Task list / manager agent | None; `TaskOutput` via `context` | Task outputs | Only `async_execution` tasks |
| CrewAI Flow | Decorated method graph | One dict or Pydantic state | Method return values | Start methods "often in parallel" |
| Swarm / Agents SDK | None; model picks handoff | Conversation history; context dict | Handoff tool call | No (tools may run concurrently) |
| AgentVerse | Recruit/decide/act/evaluate loop | Environment state | Group discussion | Not stated |
| CAMEL Workforce | Task tree with dependencies | Stored task results; memory off by default | Task assignments | "often in parallel" |
| Google ADK | Agent tree / graph | `session.state` dict | Events with `state_delta` | Yes, ParallelAgent branches |
| LlamaIndex Workflows | Step graph inferred from event types | `ctx.store` | Typed events | Yes, async steps and workers |

- Only LangGraph and Microsoft Agent Framework give a named consistency model (BSP supersteps). ADK and LlamaIndex give concurrency with explicit "order is not deterministic" caveats. The transcript-based frameworks avoid the problem by never running two writers at once.
- AutoGen 0.4 separates the two layers: the Core runtime is an actor system with no shared state, while the team abstractions on top reintroduce a shared transcript and turn order. A statement such as "AutoGen 0.4 is concurrent" is therefore true of Core and false of the default AgentChat teams.
- The examples in the LlamaIndex docs use `asyncio` primitives and describe limits as per process, which indicates asyncio concurrency in one process, not multi-core parallelism. The docs do not say this in so many words.

### Gaps
- MetaGPT source code details (whether roles within one round are awaited together, the per-role message buffer, `sent_from`/`send_to` routing) were not confirmed from a primary page in this session; the tutorial page does not describe loop internals.
- AutoGen Core docs do not state a message ordering guarantee or whether one agent instance processes several messages concurrently; the pages read are silent on both.
- The Magentic-One paper's last 8,000 characters and the AutoGen paper's appendices were not read; the exact pub/sub wiring of Magentic-One on autogen-core is not confirmed.
- CAMEL Workforce docs carry no version number, and the mechanism named "task channel" in the source code was not confirmed from the docs page.
- OpenAI's original March 2025 announcement returned HTTP 403, so its wording is known only from secondary sources.

## 2. How does each handle two agents writing the same state key at once?

### Takeaway
LangGraph is the only framework found that makes a same-key concurrent write a hard error unless a reducer is declared. Microsoft Agent Framework buffers state updates to the superstep boundary but tells users to synchronize their own shared state. LlamaIndex offers an explicit lock (`edit_state`). Google ADK serializes event appends per session in its database service and otherwise warns of races. The remaining frameworks have no conflict rule because they have no concurrent writers or no shared keyed state.

### Cited Findings
- LangGraph default: "With no reducer, updates overwrite the previous value"; "The default reducer replaces the left value with the right"; a reducer takes the current value and the node update and combines them — [LangGraph Graph API](https://docs.langchain.com/oss/python/langgraph/graph-api)
- LangGraph conflict: the error `INVALID_CONCURRENT_GRAPH_UPDATE` occurs when a StateGraph "received concurrent updates to its state from multiple nodes to a state property that doesn't support it"; if "multiple nodes in e.g. a fanout within a single step return values for "some_key", the graph will throw this error". Fix: annotate the key with a reducer, for example `Annotated[list, operator.add]`; "If your graph executes nodes in parallel, make sure you have defined relevant state keys with a reducer" — [LangGraph error reference](https://docs.langchain.com/oss/python/langgraph/errors/INVALID_CONCURRENT_GRAPH_UPDATE)
- LangGraph message lists: `add_messages` is used instead of `operator.add` because "Plain operator.add only appends, so it cannot update an existing message in place during human-in-the-loop edits" — [LangGraph Graph API](https://docs.langchain.com/oss/python/langgraph/graph-api)
- Microsoft Agent Framework (.NET): state writes are queued. "The executor that calls `QueueStateUpdateAsync` can read the updated value immediately in the same handler. Other executors see that update starting in the next superstep." Scopes are private per executor unless a shared scope name is used — [MAF State](https://learn.microsoft.com/en-us/agent-framework/concepts/workflows/state)
- Microsoft Agent Framework: "Buffered state updates: Queued workflow-state changes are committed between supersteps", but "These guarantees don't fix execution or event ordering among concurrent executors within one superstep. Synchronize access to application-owned shared state rather than assuming the superstep barrier prevents all races" — [MAF Workflow Builder and Execution](https://learn.microsoft.com/en-us/agent-framework/concepts/workflows/builder-and-execution)
- LlamaIndex Workflows: when "state can be manipulated by multiple steps running at the same time", "lock the state with `edit_state()` so the update is atomic"; "No other step can edit the state while the block is running" — [LlamaIndex managing state](https://developers.llamaindex.ai/python/llamaagents/workflows/managing_state)
- Google ADK: "Because the branches run concurrently, do not rely on one branch reading a key another branch is still writing"; "If you need communication or data sharing between these agents, you must implement it explicitly" — [ADK Parallel agents](https://adk.dev/agents/workflow-agents/parallel-agents/)
- Google ADK state page: `DatabaseSessionService` takes a per-session lock to serialize concurrent updates; `InMemorySessionService` is described as not safe for multi-threaded use; direct modification of `session.state` outside the event path is "Not Thread-Safe" and "bypasses the ADK's event tracking and can lead to lost data" — [ADK State](https://adk.dev/sessions/state/)
- CrewAI Flows: the Flows page has no statement on concurrent state modification, locking or thread safety; persistence is described as "Transaction-based state updates for data integrity" — [CrewAI Flows](https://docs.crewai.com/concepts/flows)
- CrewAI Crews: no shared mutable state is described; information moves through task outputs and `context` — [CrewAI Tasks](https://docs.crewai.com/concepts/tasks)
- AutoGen Core: each agent "maintains its own state"; a published message reaches each subscribed agent once — [AutoGen Core concepts](https://microsoft.github.io/autogen/stable/user-guide/core-user-guide/core-concepts/agent-and-multi-agent-application.html), [Topic and Subscription](https://microsoft.github.io/autogen/stable/user-guide/core-user-guide/core-concepts/topic-and-subscription.html)
- Swarm: functions run in order and may each return updated `context_variables` — [Swarm README](https://github.com/openai/swarm)
- CAMEL Workforce: worker memory is separate unless `share_memory=True` — [CAMEL Workforce docs](https://docs.camel-ai.org/key_modules/workforce)

### Inferences
- Conflict policy by framework:

| Framework | Policy for concurrent same-key writes |
|---|---|
| LangGraph | Reducer if declared; otherwise error at the superstep (not silent last-write-wins) |
| MS Agent Framework | Buffered and committed between supersteps; winner among same-step writers not documented; user told to synchronize |
| LlamaIndex Workflows | Opt-in lock via `edit_state()`; plain `get`/`set` is unprotected read-modify-write |
| Google ADK | Per-session lock on append in the database service; otherwise effectively last write wins on a key, with a documented warning |
| CrewAI Flows | Not documented |
| AutoGen Core | Not applicable: no shared state, only messages |
| MetaGPT, ChatDev, AutoGen group chat, Magentic-One, Swarm, Agents SDK | Not applicable: single writer per turn, append-only transcript |

- The transcript frameworks' state is append-only and single-writer, so the only "conflict" they can have is semantic (two agents disagree in text), which no framework resolves structurally. Magentic-One resolves it by having one Orchestrator own both ledgers.
- LangGraph's overwrite default is last-write-wins only across supersteps; within one superstep it refuses. This distinction is easy to lose in summaries that say "LangGraph is last-write-wins by default".
- ADK's recommended pattern (distinct `output_key` per parallel branch, then a later agent reads all keys) is a convention that avoids conflicts by partitioning the key space, not a mechanism that detects them.

### Gaps
- Microsoft Agent Framework docs do not say which update wins when two executors queue an update to the same shared-scope key in one superstep.
- ADK docs do not use the phrase "last-write-wins"; behaviour for two branches writing the same key through `state_delta` is not spelled out.
- CrewAI Flow behaviour when two parallel listeners mutate the same state field is undocumented on the pages read; source code was not inspected.
- LangGraph's atomicity claim (a superstep's writes are applied all together or not at all) is commonly stated but was not found verbatim on the pages read.

## 3. What is persisted, and can a run be resumed or replayed?

### Takeaway
LangGraph and Microsoft Agent Framework persist a full snapshot at every superstep boundary and support resume, replay from an earlier checkpoint and forking. LlamaIndex serializes the context (state store plus in-flight events) but leaves the checkpoint loop to the user and is at-least-once. ADK persists an event log plus state through its session service. CrewAI Flows persist state per method to SQLite. The research systems of 2023 (MetaGPT, ChatDev, AgentVerse, CAMEL) did not specify resumable execution, and MetaGPT's paper lists that as a limitation.

### Cited Findings
- LangGraph: "A checkpoint is a snapshot of the graph state saved at each super-step." A `StateSnapshot` holds `values`, `next`, `config` (`thread_id`, `checkpoint_ns`, `checkpoint_id`), `metadata` (`source`, `writes`, `step`), `created_at`, `parent_config`, `tasks` — [LangGraph checkpointers](https://docs.langchain.com/oss/python/langgraph/checkpointers)
- LangGraph replay: "Invoke the graph with a prior `checkpoint_id` to re-run nodes after that checkpoint"; "Nodes before the checkpoint are skipped (their results are already saved)". State updates create a new checkpoint and leave the original, enabling forks — [LangGraph checkpointers](https://docs.langchain.com/oss/python/langgraph/checkpointers)
- LangGraph partial failure: "LangGraph stores pending checkpoint writes from any other nodes that completed successfully at that super-step"; "When you resume graph execution from that super-step you don't re-run the successful nodes" — [LangGraph checkpointers](https://docs.langchain.com/oss/python/langgraph/checkpointers)
- LangGraph durability modes: `"exit"` persists only when execution exits; `"async"` persists while the next step runs; `"sync"` persists before the next step starts — [LangGraph checkpointers](https://docs.langchain.com/oss/python/langgraph/checkpointers)
- LangGraph granularity limit: "With a checkpointer, LangGraph saves checkpoints at super-step boundaries, not mid-function inside a node"; after an interrupt or retry a node re-runs from the start — [LangGraph Graph API](https://docs.langchain.com/oss/python/langgraph/graph-api)
- LangGraph backends: `InMemorySaver` (lost on restart), `SqliteSaver` (development), `PostgresSaver` (recommended for production); a separate Store holds cross-thread data; "each subgraph manages its own checkpoint namespace" — [LangGraph persistence](https://docs.langchain.com/oss/python/langgraph/persistence)
- Microsoft Agent Framework: "Checkpoints are created at the end of each superstep, after all executors in that superstep have completed their execution", capturing "The current state of all executors", "All pending messages in the workflow for the next superstep", "Pending requests and responses", "Shared states" — [MAF Checkpoints](https://learn.microsoft.com/en-us/agent-framework/workflows/checkpoints)
- Microsoft Agent Framework: "Starting in Python version 1.13.0, workflows also create an entry checkpoint before the first superstep to record the workflow input... These checkpoints make the complete workflow run replayable." Storage: in-memory, file, Azure Cosmos DB. Rehydration requires the same topology and executor identities. Python stores use pickle behind a restricted unpickler — [MAF Checkpoints](https://learn.microsoft.com/en-us/agent-framework/workflows/checkpoints)
- LlamaIndex: a snapshot (`Context.to_dict()`) captures "the events still in flight plus the state store". There is no built-in checkpointer; the user saves on `StepStateChanged`. "A step that was mid-execution when you captured the state is rewound and runs again from the top." "Resume is at-least-once, and step side effects need to be safe to repeat." A DBOS runtime plugin journals step transitions — [LlamaIndex durable workflows](https://developers.llamaindex.ai/python/llamaagents/workflows/durable_workflows)
- Google ADK: with `InMemorySessionService` state "is lost on restart"; with `DatabaseSessionService` or `VertexAiSessionService` state "is saved reliably"; `temp:` keys are never persisted — [ADK State](https://adk.dev/sessions/state/)
- CrewAI Flows: "The @persist decorator enables automatic state persistence"; "SQLiteFlowPersistence is the default storage backend"; resume with `kickoff(inputs={"id": <uuid>})`, fork with `kickoff(restore_from_state_id=<uuid>)`; an unknown ID "falls back silently" — [CrewAI Flows](https://docs.crewai.com/concepts/flows)
- AutoGen AgentChat: "Teams are stateful and maintains the conversation history and context after each run, unless you reset the team"; a further `run()` continues "from the next agent in the round-robin order" — [AgentChat Teams tutorial](https://microsoft.github.io/autogen/stable/user-guide/agentchat-user-guide/tutorial/teams.html)
- OpenAI Agents SDK: history is carried by `to_input_list()`, Sessions, `conversation_id` or `previous_response_id`; a paused run resumes from a `RunState`; durable execution is delegated to Temporal, Restate or DBOS integrations — [Agents SDK Running agents](https://openai.github.io/openai-agents-python/running_agents/)
- Swarm: stateless between calls; the caller passes messages and context back in — [Swarm README](https://github.com/openai/swarm)
- MetaGPT paper limitation: users find it hard to "interrupt the running process of each agent, or set the starting running point (checkpoint)" — [MetaGPT paper v7](https://arxiv.org/html/2308.00352v7)

### Inferences
- Replay fidelity ranks roughly: LangGraph and MAF (per-superstep snapshots with lineage, fork and replay) > LlamaIndex (snapshot, user-driven, at-least-once) and CrewAI Flows (per-method state snapshot) > ADK (durable event log and state, but no documented step-level replay on the pages read) > Agents SDK (conversation history plus external durable engines) > Swarm and the 2023 research systems (none).
- In both BSP frameworks the checkpoint unit is the superstep, so work inside a long node or executor is not recoverable; both docs advise accordingly (LangGraph re-runs the node from the top; MAF advises consolidating chains into one executor, which enlarges the unrecoverable unit). There is a direct trade-off between barrier latency and checkpoint granularity.
- None of the frameworks persists a write-ahead log of individual state mutations with causality; LangGraph's `metadata.writes` and `parent_config` are the closest thing to a per-step provenance record.

### Gaps
- AutoGen `save_state`/`load_state` exist in the API but were not on the page read; what exactly is serialized was not verified.
- MetaGPT's "Serialization & Breakpoint Recovery" doc page exists as a sidebar link but its content was not read.
- CrewAI's interaction between `@persist` and concurrently running methods is undocumented.
- ADK 2.x graph workflows may add checkpoint and resume features; not verified from a primary page.

## 4. What benchmarks or measurements exist for each design, and what was only claimed?

### Takeaway
Only the research-origin systems have measured results, and those measure end-task success of the whole system, never the coordination data structure in isolation, except for two ablations: Magentic-One without its ledgers (a 31% drop) and MetaGPT role and feedback ablations. The engineering frameworks (LangGraph, CrewAI, Agents SDK, ADK, LlamaIndex, Microsoft Agent Framework) publish semantics, not measurements. Independent work (MAST, the Google scaling study) reports high failure rates and task-dependent gains or losses.

### Cited Findings

**Measured by the authors**
- MetaGPT: Pass@1 of 85.9% on HumanEval and 87.7% on MBPP. Executable feedback adds +4.2% (HumanEval) and +5.4% (MBPP) — [MetaGPT paper v7](https://arxiv.org/html/2308.00352v7)
- MetaGPT on SoftwareDev (ChatDev / MetaGPT without feedback / MetaGPT): executability 2.25 / 3.67 / 3.75 (scale to 4); running time 762 / 503 / 541 s; tokens 19,292 / 24,613 / 31,255; code files 1.9 / 4.6 / 5.1; lines 77.5 / 194.6 / 251.4; tokens per line 248.9 / 126.5 / 124.3; human revision cost 2.5 / 2.25 / 0.83 — [MetaGPT paper v7](https://arxiv.org/html/2308.00352v7)
- MetaGPT role ablation: a lone Engineer scores executability 1.0 with 10 revisions; the four-role team 4.0 with 2.5 revisions at higher cost ($1.385 vs $0.915) — [MetaGPT paper v7](https://arxiv.org/html/2308.00352v7)
- ChatDev on SRDD (1,200 prompts): completeness / executability / consistency / quality are 0.5600 / 0.8800 / 0.8021 / 0.3953 for ChatDev, 0.4834 / 0.4145 / 0.7601 / 0.1523 for MetaGPT, 0.5022 / 0.3583 / 0.7887 / 0.1419 for GPT-Engineer. ChatDev used 148.2 s and 22,949 tokens per task; MetaGPT 154.0 s and 29,279 tokens; GPT-Engineer 15.6 s and 7,183 tokens — [ChatDev paper v5](https://arxiv.org/html/2307.07924)
- ChatDev ablations: removing roles drops quality to 0.2212; removing communicative dehallucination to 0.3094 — [ChatDev paper v5](https://arxiv.org/html/2307.07924)
- Conflict between the two papers: MetaGPT reports beating ChatDev on its SoftwareDev set (executability 3.75 vs 2.25), while ChatDev reports beating MetaGPT on SRDD (executability 0.88 vs 0.41) — [MetaGPT paper v7](https://arxiv.org/html/2308.00352v7); contradicted by [ChatDev paper v5](https://arxiv.org/html/2307.07924)
- AutoGen 0.2: 69.48% on the 5,000-problem MATH test set against 55.18% for GPT-4 alone; "around 15%" gain on 134 unseen ALFWorld tasks from adding a grounding agent; OptiGuide workflow code reduced "from over 430 lines to 100 lines"; unsafe-code F1 up 8% (GPT-4) and 35% (GPT-3.5-turbo) with the multi-agent design — [AutoGen paper v2](https://arxiv.org/html/2308.08155)
- AutoGen group chat: on 12 hand-made tasks, role-play speaker selection gave "a higher success rate and fewer LLM calls" than a task-based prompt; no figures in the main text — [AutoGen paper v2](https://arxiv.org/html/2308.08155)
- Magentic-One task completion: GAIA test (300 questions) 38.00±5.5 with GPT-4o plus o1 and 32.33±5.3 with GPT-4o; AssistantBench test (181) accuracy 27.7±6.5 and 25.3±6.3; WebArena (812 tasks) 32.8±3.2 with GPT-4o. Comparisons: omne v0.1 40.53±5.6 and humans 92.00±3.1 on GAIA; Jace.AI 57.1±3.4, WebPilot 37.2±3.3 and humans 78.2±2.8 on WebArena — [Magentic-One paper](https://arxiv.org/html/2411.04468)
- Magentic-One ablation (GAIA validation, GPT-4o): replacing the ledger Orchestrator with a simple group-chat orchestrator reduces performance by 31%; removing one worker agent reduces it by 21% to 39% — [Magentic-One paper](https://arxiv.org/html/2411.04468)
- Magentic-One error analysis: most frequent codes are persistent-inefficient-actions, insufficient-verification-steps and inefficient-navigation-attempts — [Magentic-One paper](https://arxiv.org/html/2411.04468)
- AgentVerse (CoT / Solo / Group), GPT-4: FED 95.4 / 95.8 / 96.8; Commongen-Challenge 95.9 / 99.0 / 99.1; MGSM 95.2 / 96.0 / 95.2; Logic Grid Puzzles 59.5 / 64.0 / 66.5; HumanEval pass@1 83.5 / 87.2 / 89.0. GPT-3.5-Turbo: FED 81.6 / 81.1 / 85.1; Commongen 76.6 / 93.6 / 92.3; MGSM 80.4 / 82.4 / 80.8; HumanEval 73.8 / 74.4 / 75.6 — [AgentVerse paper v3](https://arxiv.org/html/2308.10848)
- AgentVerse negative result: with GPT-3.5-Turbo the group underperforms a solo agent "in two of three tasks"; about 10% of MGSM errors are attributed to one agent being swayed by another's wrong feedback — [AgentVerse paper v3](https://arxiv.org/html/2308.10848)

**Independent evaluations**
- MAST ("Why Do Multi-Agent LLM Systems Fail?"): 1,642 annotated traces across seven open-source systems (MetaGPT, ChatDev, HyperAgent, AppWorld, AG2 MathChat, Magentic-One, OpenManus); a reported "41% to 86.7% failure rate"; 14 failure modes in three categories, the largest single modes being FM-1.3 at 15.7%, FM-2.6 at 13.2%, FM-1.5 at 12.4% and FM-1.1 at 11.8% (category sums, computed here and not stated by the paper: system design about 44%, inter-agent misalignment about 32%, task verification about 24%); annotator agreement κ = 0.88 — [MAST paper](https://arxiv.org/html/2503.13657)
- MAST conclusion: "many MAS failures arise from the challenges in organizational design and agent coordination rather than the limitations of individual agents"; targeted fixes gave ChatDev +9.4% (workflow change) and +15.6% (added verification) but completion rates "still remain low" — [MAST paper](https://arxiv.org/html/2503.13657)
- "Towards a Science of Scaling Agent Systems" (v1 9 Dec 2025, v3 8 Apr 2026): 260 configurations, six benchmarks, three model families, five architectures (single, independent, centralized, decentralized, hybrid). Change against a single agent ranges from +80.8% on decomposable financial reasoning to -70.0% on sequential planning; architectures without centralized verification propagate errors more; a fitted model picks the best architecture for 87% of held-out configurations (cross-validated R² 0.373) — [Scaling study](https://arxiv.org/abs/2512.08296)

**Claimed, not measured**
- CAMEL abstract makes qualitative claims about role-playing as a scalable way to "generate conversational data"; no numbers in the abstract — [CAMEL abstract](https://arxiv.org/abs/2303.17760)
- AutoGen 0.4 is described as improving "code quality, robustness, usability, and scalability"; no benchmark accompanies the claim — [AutoGen 0.4 launch post](https://devblogs.microsoft.com/autogen/autogen-reimagined-launching-autogen-0-4/)
- Microsoft Agent Framework lists BSP "guarantees" (step boundaries, checkpoint boundaries, buffered state) as design properties, with no performance measurement — [MAF Workflow Builder and Execution](https://learn.microsoft.com/en-us/agent-framework/concepts/workflows/builder-and-execution)
- CrewAI states "Transaction-based state updates for data integrity" for Flow persistence without evidence or a definition of the transaction scope — [CrewAI Flows](https://docs.crewai.com/concepts/flows)

### Inferences
- The only measurements that isolate a coordination data structure are (1) Magentic-One's ledger ablation, -31% without the two ledgers, and (2) MetaGPT's finding that structured documents plus roles lift executability from 1.0 to 4.0. Both support "explicit structured plan and progress state beats a free transcript", on single benchmarks with 2023 to 2024 models.
- No framework has published a measurement of its concurrency model: nothing was found on wall-clock speed-up from parallel supersteps, conflict rates, or correctness under concurrent writes. Statements about parallelism benefits are design claims.
- MetaGPT and ChatDev each win on the benchmark their own authors built, with different executability definitions (a 1 to 4 score vs a fraction that runs). These results cannot be compared directly and neither should be cited as showing one design is better.
- MAST's top failure modes (by the paper's mode names as I recall them, not re-verified in this session: FM-1.3 step repetition, FM-1.5 unaware of termination conditions, FM-2.6 reasoning-action mismatch, FM-1.1 disobeying task specification) and Magentic-One's own error analysis (repeated unproductive actions, insufficient verification) agree, even though Magentic-One has an explicit loop detector and stall counter. A progress ledger reduces but does not remove looping.
- The scaling study's spread (+80.8% to -70.0%) suggests the value of parallel fan-out depends on whether the task decomposes, which matches the frameworks' own design split between map-reduce primitives (LangGraph `Send`, ADK ParallelAgent, LlamaIndex fan-out) and sequential handoffs.

### Gaps
- The text names of the MAST failure-mode codes (FM-1.1 and so on) were not captured in the text read; only codes and percentages are confirmed.
- MAST per-framework failure rates are in a figure whose numbers were not in the text read; only the 41% to 86.7% range is confirmed.
- The scaling study's frequently quoted error-amplification figures (17.2x independent vs 4.4x centralized) and author affiliations were not on the abstract page and are unverified here.
- No independent benchmark was found for LangGraph, CrewAI, OpenAI Agents SDK, Google ADK, LlamaIndex Workflows or Microsoft Agent Framework as coordination substrates.
- CAMEL's full-paper human and GPT-4 evaluation numbers were not read; AutoGen's 120-problem level-5 MATH figures were not in the text read.
- All measured results predate 2025 model generations; the size of the effects with newer models is unknown.

## 5. Which design choices were abandoned or changed between versions, and why?

### Takeaway
The clear direction from 2023 to 2026 is away from implicit control flow hidden in a conversation and towards explicit graphs or event systems with typed state and checkpoints. AutoGen moved from synchronous conversation to an actor runtime (0.4, Jan 2025) and was then folded into Microsoft Agent Framework with a Pregel superstep engine (GA April 2026). Swarm was replaced by the Agents SDK. ADK 2.0 moved from a hierarchical agent executor to a graph engine. CrewAI added Flows beside Crews.

### Cited Findings
- AutoGen 0.2 to 0.4: v0.4 is "a complete redesign of the AutoGen library"; AgentChat is "a replacement for AutoGen v0.2" — [AutoGen 0.4 launch post, 17 Jan 2025](https://devblogs.microsoft.com/autogen/autogen-reimagined-launching-autogen-0-4/)
- Stated reasons: users needed "greater modularity and the ability to reuse agents seamlessly" and "better support for debugging and scaling their agentic solutions"; the new design brings "asynchronous message exchange between agents" and "affordances to observe and control agent behavior" — [Microsoft Research article, 25 Feb 2025](https://www.microsoft.com/en-us/research/articles/autogen-v0-4-reimagining-the-foundation-of-agentic-ai-for-scale-extensibility-and-robustness/)
- AutoGen to Microsoft Agent Framework: MAF 1.0 reached general availability in early April 2026 for .NET and Python as the successor to AutoGen and Semantic Kernel, which remain in maintenance mode (bug and security fixes, no new features) — [Visual Studio Magazine, 6 Apr 2026](https://visualstudiomagazine.com/articles/2026/04/06/microsoft-ships-production-ready-agent-framework-1-0-for-net-and-python.aspx), [VentureBeat](https://venturebeat.com/ai/microsoft-retires-autogen-and-debuts-agent-framework-to-unify-and-govern)
- The engine changed again in MAF: from AutoGen Core's free-running actor messaging to "a modified Pregel execution model — a Bulk Synchronous Parallel (BSP) approach", justified by defined step boundaries, checkpoint boundaries and buffered state updates — [MAF Workflow Builder and Execution](https://learn.microsoft.com/en-us/agent-framework/concepts/workflows/builder-and-execution)
- MAF keeps the earlier patterns as prebuilt orchestrations with stable names `Concurrent`, `Sequential`, `GroupChat`, `Handoff`, `Magentic` — [MAF Checkpoints](https://learn.microsoft.com/en-us/agent-framework/workflows/checkpoints)
- MAF Python 1.13.0 changed checkpointing to add entry checkpoints so "the complete workflow run [is] replayable", with "minor breaking changes" to iteration counts and checkpoint ordering — [MAF Checkpoints](https://learn.microsoft.com/en-us/agent-framework/workflows/checkpoints)
- Swarm: "Swarm is now replaced by the OpenAI Agents SDK, which is a production-ready evolution of Swarm"; Swarm itself is titled "experimental, educational" — [Swarm README](https://github.com/openai/swarm)
- Agents SDK changes to the handoff structure: handoffs became tools with typed inputs (`input_type`), per-handoff `input_filter`, and an opt-in beta `nest_handoff_history` in which "the runner compacts summarizable history into ordered assistant summary segments" — [Agents SDK Handoffs](https://openai.github.io/openai-agents-python/handoffs/)
- Google ADK: the ParallelAgent page notes that template workflows are superseded by graph-based workflows; ADK 2.0 docs describe a move from a hierarchical agent executor to a graph-based execution engine — [ADK Parallel agents](https://adk.dev/agents/workflow-agents/parallel-agents/), [ADK 2.0 overview](https://adk.dev/2.0/)
- CrewAI: the `Process` enum now lists only sequential and hierarchical; Flows carry their own version history (`crewai run` for flows "Starting from version 0.103.0"; human feedback "requires CrewAI version 1.8.0 or higher") — [CrewAI Processes](https://docs.crewai.com/concepts/processes), [CrewAI Flows](https://docs.crewai.com/concepts/flows)
- LlamaIndex: the current durable-workflows page says there is no built-in checkpointer and the user runs the snapshot loop, with a DBOS runtime as the managed alternative — [LlamaIndex durable workflows](https://developers.llamaindex.ai/python/llamaagents/workflows/durable_workflows)
- LangGraph: default recursion limit raised to 10,000 steps between 1.0.6 and 1.1.3 and 10,007 from 1.1.4 — [LangGraph Graph API](https://docs.langchain.com/oss/python/langgraph/graph-api)
- MetaGPT paper revisions: seven arXiv versions from 1 Aug 2023 to 1 Nov 2024 — [arXiv abstract](https://arxiv.org/abs/2308.00352)

### Inferences
- Timeline: 2023, conversation as control flow (CAMEL, ChatDev, AutoGen 0.2, AgentVerse) and SOP pipelines (MetaGPT). 2024, explicit state objects (LangGraph channels and reducers, Magentic-One ledgers, CrewAI Flows, LlamaIndex Workflows) and handoff minimalism (Swarm). 2025, runtimes (AutoGen 0.4 actor Core, Agents SDK, ADK). 2026, convergence on superstep graphs with checkpoints (MAF 1.0, ADK 2.0 graph engine, LangGraph 1.x).
- AutoGen's history is two successive corrections: first from synchronous chat to an unconstrained actor runtime, then to BSP supersteps. The second step trades the actor model's freedom for checkpointable step boundaries, which suggests that resumability and observability, not raw concurrency, drove the design.
- Microsoft and LangChain arrived independently at the same Pregel model, and Google at a graph engine, which is the strongest available signal on which structure practitioners settled on; it is convergence of design, not measured superiority.
- The handoff line (Swarm to Agents SDK) kept its single-active-agent, shared-transcript structure and added controls on what history crosses a handoff. The persistent problem there is context size and relevance, not write conflicts.
- Statements likely to be outdated: anything describing AutoGen 0.2 `GroupChat` as current; Swarm as a usable framework; CrewAI's "consensual" process (planned in early docs, absent from the current enum); LlamaIndex `WorkflowCheckpointer` (absent from the current durable-workflows page); ADK `SequentialAgent`/`ParallelAgent`/`LoopAgent` as the primary workflow mechanism.

### Gaps
- The exact GA date of Microsoft Agent Framework 1.0 differs across secondary sources (2 or 3 April 2026), and one secondary source says AutoGen was archived while Microsoft's wording is maintenance mode. Microsoft's own announcement was not fetched.
- OpenAI's primary announcement of the Agents SDK (March 2025) could not be fetched (HTTP 403); the date rests on the Swarm README notice and secondary sources. One secondary source gives March 2026, which conflicts with the rest.
- ADK 2.0 release dates (reported as May 2026 for Python, June 2026 for Go) and whether ParallelAgent is formally deprecated or a thin wrapper over the graph engine are disputed among secondary sources; no primary release note was read.
- The AG2 community fork of AutoGen 0.2, ChatDev's later graph-based versions (MacNet, ChatDev 2.0), MetaGPT's later products and LangGraph's 1.0 release date were not researched in this session.
- No primary source was found giving the reason CrewAI dropped the consensual process or LlamaIndex removed a built-in checkpointer.
