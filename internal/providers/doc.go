// Package providers holds everything the orchestrator needs to talk to a
// headless CLI: provider profiles with their windows and fractions, one
// Adapter per CLI (claude, codex, gemini, kimi, grok, muse) plus a scripted
// fake, the macOS sandbox profile, the process runner with its watchdogs, the
// smoke runner, and usage parsing with a rate card.
//
// The package was copied from BookWriter v2, whose provider and routing
// specification and verified flag matrix are not in this repository.
//
// # Conventions the next packages can rely on
//
// Paths. Every path in an Invocation is absolute. Workspace is the packet
// directory (04 §4.3) and the child's working directory; RunDir is the run
// directory (04 §4.14) that receives stdout.log, stderr.log, events.jsonl,
// usage.json, argv.json, env.redacted.json, pgid, sandbox.sb and the
// compaction marker. HandoffPath defaults to <Workspace>/.cx/handoff.json
// when empty.
//
// Environment. The child environment is an allowlist: PATH, HOME, LANG,
// LC_ALL, LC_CTYPE, USER, LOGNAME, SHELL and TERM are copied from the parent,
// TMPDIR is set to <RunDir>/tmp, each adapter adds its own switches (for
// example DISABLE_COMPACT=1 for claude, GEMINI_API_KEY for gemini), and
// Invocation.Env is applied last and wins. No other variable leaks in.
//
// Events. ParseEvent maps one stream line to one Event. Kind is one of the
// Kind* constants; ContextTokens is set on any event that carries a context
// measurement (per-turn usage on claude and codex, usage_update on kimi)
// regardless of Kind, so the watchdog and PeakContext read every event.
// Tool events carry the tool name in Message, which is how RUN-NO-FOREIGN-TOOLS
// is checked. Raw is the original line.
//
// Usage. InputTokens is the total prompt side (uncached plus cached);
// CachedTokens is the part served from a prompt cache; CacheWriteTokens the
// part written to one; OutputTokens includes
// reasoning; ReasoningTokens is the reasoning subset where reported;
// PeakContext is the largest context measurement seen. FinalUsage returns
// ErrNoUsage when the output carries no usage at all, so the caller can charge
// the run at planned size (RUN-USAGE-RECORDED).
//
// Exit classes. The adapters classify in this order: any compaction event is
// TaintedCompaction; then an auth signature, a quota signature, and a context
// signature in error events, rate-limit events or stderr; then exit code 0 is
// Ok (or OkBlocked when the provider reports it stopped on its own cap) and
// anything else is Crash. Run overrides the class with Timeout, Stalled,
// TaintedCompaction (compaction kill or marker) or ContextBlocked (kill line)
// when its watchdog killed the process. InvalidHandoff and ScopeViolation are
// never produced here; they belong to collect, which reads the handoff and
// the diff.
//
// Sandbox. The seatbelt compiler on this macOS accepts only "*" and
// "localhost" as hosts in network rules (verified on 2026-10-21), so a
// per-host allowlist is enforced by a localhost CONNECT proxy that Run starts
// for sandboxed runs without Network; the profile admits only that port and
// the proxy refuses every host outside the provider's list. Research runs
// (Network true) get (allow network*).
//
// Kill. Run has no handle: cancel the context to kill the run. The outcome
// records why a run was killed in Killed (timeout, stalled, compaction,
// kill_line, cancelled).
package providers
