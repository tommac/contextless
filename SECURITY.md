# Security

## What `cx` does on your machine

Read this before running `cx` with real models.

- **It starts vendor CLIs with their permission prompts turned off.** A
  prompt would hang a headless run, so each adapter passes the CLI's own
  bypass switch (for example `--permission-mode bypassPermissions`,
  `approval_policy="never"`, `--approval-mode yolo`). Inside its run, a model
  can read, write and execute whatever the process is allowed to.
- **On macOS a sandbox contains each run, and it is on by default.** The
  Seatbelt profile lets a run write only to its packet and run directories,
  read the system and the CLI's own configuration, and reach the network only
  through the hosts its provider needs (the `researcher` role gets the open
  network).
- **On Linux there is no sandbox.** `cx` refuses to start a CLI until you set
  `sandbox: false`, and then the CLIs run unconfined with prompts bypassed;
  `cx` prints a warning whenever it runs this way.
  Run it in a container, a virtual machine or a throwaway user account.
  Windows is not supported.
- **Your conversation goes to the providers you enable.** Every packet holds
  verbatim text from the workspace. Limit providers with `enable:` in the
  configuration.
- **Coder units edit a git worktree, not your checkout,** and a change
  outside the path allowlist is rejected. The verify command you configure
  runs outside the sandbox.
- **The dashboard (`cx ui`) and the MCP server (`cx serve -mcp`) expose the
  workspace.** The dashboard answers only on its own address and origin; do
  not bind it to a public interface. An MCP client can read the vault, append
  turns and fork workspaces.

API keys are read from the environment by the vendor CLIs or the opt-in HTTP
adapters. `cx` does not store them, and they are not written to packets.

## Reporting a vulnerability

Please do not open a public issue. Use GitHub's private reporting: the
"Report a vulnerability" button under this repository's Security tab. Say
what you ran, on which platform, and what happened. You should hear back
within a week.

Known open defects that are not security problems are listed in
`docs/ROADMAP.md`.
