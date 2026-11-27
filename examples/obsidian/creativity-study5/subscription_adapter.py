#!/usr/bin/env python3
"""One fresh, tool-free study request through the existing Codex or Claude CLI."""
import json
import os
import shutil
from pathlib import Path
import subprocess
import sys
import tempfile

MODELS = {"gpt-6.1-sol": "openai", "claude-opus-5-5": "anthropic"}
RULE = ("Answer only the supplied study packet. Return a JSON object keyed by its task IDs, "
        "with exactly the requested answer or rating for each ID. No tools or external context.")

def call(request, model):
    family = MODELS[model]
    prompt = RULE + "\n" + json.dumps(request, ensure_ascii=False)
    with tempfile.TemporaryDirectory(prefix="cx-study-session-") as folder:
        env = dict(os.environ)
        # The study uses the existing subscriptions. Do not route to a metered API key.
        for key in ("OPENAI_API_KEY", "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN",
                    "ANTHROPIC_BASE_URL", "OPENAI_BASE_URL"):
            env.pop(key, None)
        if family == "openai":
            home = Path(folder, "codex-home")
            home.mkdir()
            auth = Path(env.get("CODEX_HOME", Path.home() / ".codex"), "auth.json")
            (home / "auth.json").symlink_to(auth)
            catalog = auth.with_name("models_cache.json")
            if catalog.is_file():
                shutil.copyfile(catalog, home / "models_cache.json")
            env["CODEX_HOME"] = str(home)
            argv = ["codex", "exec", "-C", folder, "--skip-git-repo-check", "--ephemeral",
                    "--sandbox", "read-only", "-m", model, "--json",
                    "-c", 'model_reasoning_effort="high"', "-c", 'approval_policy="never"',
                    "-c", 'web_search="disabled"', "-c", 'model_provider="openai_inline"',
                    "-c", 'model_providers.openai_inline.name="openai_inline"',
                    "-c", "model_providers.openai_inline.requires_openai_auth=true",
                    "-c", "model_providers.openai_inline.supports_websockets=true"]
            for feature in ("shell_tool", "apps", "plugins", "remote_plugin", "memories",
                            "browser_use", "browser_use_external", "in_app_browser", "computer_use",
                            "image_generation", "tool_suggest", "chronicle", "multi_agent"):
                argv.extend(["--disable", feature])
            argv.append("-")
        else:
            argv = ["claude", "-p", "--model", model, "--effort", "high", "--tools", "",
                    "--strict-mcp-config", "--setting-sources", "", "--disable-slash-commands",
                    "--no-session-persistence", "--output-format", "json", "--system-prompt", RULE]
        p = subprocess.run(argv, input=prompt, capture_output=True, text=True, cwd=folder,
                           env=env, timeout=300)
        if p.returncode:
            # Model errors, not raw prompts or session/user metadata.
            details = []
            for line in p.stdout.splitlines():
                try:
                    obj = json.loads(line)
                except ValueError:
                    continue
                if obj.get("type") in ("error", "turn.failed") or obj.get("is_error"):
                    details.append(str(obj.get("message") or obj.get("error") or obj.get("result")))
            raise ValueError("provider exit %d: %s" % (p.returncode, (" ".join(details) or p.stderr)[-550:]))
        if family == "openai":
            events = [json.loads(line) for line in p.stdout.splitlines() if line.strip()]
            messages = []
            for event in events:
                item = event.get("item") or {}
                if item.get("type") == "error":
                    # Codex labels its context-window metadata warning as an item error.
                    # This changes the local size estimate, not the explicitly requested model.
                    if str(item.get("message", "")).startswith("Model metadata for `" + model + "` not found. Defaulting to fallback metadata"):
                        continue
                    raise ValueError("provider event: " + str(item.get("message"))[:450])
                if item and item.get("type") not in ("reasoning", "agent_message"):
                    raise ValueError("unexpected reader action: " + str(item.get("type")))
                if event.get("type") == "item.completed" and item.get("type") == "agent_message":
                    messages.append(item["text"])
            completed = [e for e in events if e.get("type") == "turn.completed"]
            if len(completed) != 1 or not messages:
                raise ValueError("incomplete reader turn")
            usage = completed[0]["usage"]
            raw = messages[-1]
        else:
            obj = json.loads(p.stdout)
            if obj.get("is_error") or model not in obj.get("modelUsage", {}):
                raise ValueError("reviewer failed or returned unexpected model telemetry")
            raw = obj["result"]
            usage = dict(obj["usage"])
            usage["input_tokens"] += usage.get("cache_read_input_tokens", 0) + usage.get("cache_creation_input_tokens", 0)
        raw = raw.strip()
        if raw.startswith("```") and raw.endswith("```"):
            raw = raw.split("\n", 1)[1].rsplit("```", 1)[0]
        answers = json.loads(raw)
        if set(answers) == {"answers"}:
            answers = answers["answers"]
        return {"model": model, "family": family, "answers": answers,
                "usage": {k: usage[k] for k in ("input_tokens", "output_tokens")}}


if __name__ == "__main__":
    request = json.load(sys.stdin)
    model = os.environ.get("CX_STUDY_MODEL") or request.get("model")
    if model not in MODELS:
        sys.exit("Set CX_STUDY_MODEL to one of: " + ", ".join(MODELS))
    try:
        print(json.dumps(call(request, model), ensure_ascii=False))
    except (OSError, ValueError, KeyError, subprocess.TimeoutExpired) as exc:
        sys.exit(str(exc))
