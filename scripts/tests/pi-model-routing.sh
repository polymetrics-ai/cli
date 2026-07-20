#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

python3 - "$repo_root" <<'PY'
import json
import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1])
model = "openai-codex/gpt-5.6-sol"
implementation_agents = {
    "pm-docs-writer",
    "pm-gsd-worker",
    "pm-issue-worker",
}


def frontmatter(path: pathlib.Path) -> dict[str, str]:
    text = path.read_text()
    if not text.startswith("---\n"):
        raise AssertionError(f"{path}: missing YAML frontmatter")
    try:
        raw = text.split("---\n", 2)[1]
    except IndexError as exc:
        raise AssertionError(f"{path}: unterminated YAML frontmatter") from exc
    values: dict[str, str] = {}
    for line in raw.splitlines():
        if ":" not in line:
            continue
        key, value = line.split(":", 1)
        values[key.strip()] = value.strip()
    return values


agent_paths = sorted((root / ".pi" / "agents").glob("*.md"))
if not agent_paths:
    raise AssertionError("no project Pi agents found")

seen: set[str] = set()
for path in agent_paths:
    values = frontmatter(path)
    name = values.get("name", "")
    seen.add(name)
    expected_thinking = "high" if name in implementation_agents else "xhigh"
    assert values.get("model") == model, (
        f"{path}: model={values.get('model')!r}, want {model!r}"
    )
    assert values.get("thinking") == expected_thinking, (
        f"{path}: thinking={values.get('thinking')!r}, want {expected_thinking!r}"
    )

missing = implementation_agents - seen
assert not missing, f"missing implementation agents: {sorted(missing)}"

settings = json.loads((root / ".pi" / "settings.json").read_text())
assert settings["defaultProvider"] == "openai-codex", "Pi default provider must be openai-codex"
assert settings["defaultModel"] == "gpt-5.6-sol", "Pi default model must be gpt-5.6-sol"
assert settings["defaultThinkingLevel"] == "xhigh", "Pi main-session default must be xhigh"

config = json.loads((root / ".planning" / "config.json").read_text())
assert config["parallelization"]["max_concurrent_agents"] == 4, (
    "parallelization.max_concurrent_agents must match Pi's four-worker concurrency cap"
)
required_gsd_overrides = {
    "gsd-loop-coordinator",
    "gsd-loop-planner",
    "gsd-loop-reviewer",
    "gsd-loop-backend",
    "gsd-planner",
    "gsd-plan-checker",
    "gsd-verifier",
    "gsd-executor",
    "gsd-code-fixer",
    "gsd-doc-writer",
}
assert required_gsd_overrides <= set(config["model_overrides"]), (
    "missing required GSD model overrides: "
    f"{sorted(required_gsd_overrides - set(config['model_overrides']))}"
)
for agent, configured_model in config["model_overrides"].items():
    assert configured_model == model, (
        f".planning/config.json model_overrides.{agent}={configured_model!r}, want {model!r}"
    )
effort = config["effort"]
assert effort["default"] == "xhigh", "GSD default effort must be xhigh"
assert effort["routing_tier_defaults"] == {
    "light": "xhigh",
    "standard": "xhigh",
    "heavy": "xhigh",
}, "all non-overridden GSD routing tiers must use xhigh"
expected_high_overrides = {
    "gsd-loop-backend",
    "gsd-executor",
    "gsd-code-fixer",
    "gsd-doc-writer",
}
assert {
    agent for agent, configured_effort in effort["agent_overrides"].items()
    if configured_effort == "high"
} == expected_high_overrides, "only GSD implementation roles may use high"

driver = (root / "scripts" / "pi-shepherd-loop.sh").read_text()
required_driver_patterns = {
    "Sol coordinator default": r'ORCH_MODEL="\$\{ORCH_MODEL:-openai-codex/gpt-5\.6-sol\}"',
    "xhigh coordinator default": r'ORCH_THINKING="\$\{ORCH_THINKING:-xhigh\}"',
    "explicit coordinator thinking": r'--thinking "\$ORCH_THINKING"',
    "Sol/xhigh validator default": (
        r'VALIDATOR_ARGS="\$\{VALIDATOR_ARGS:---model openai-codex/gpt-5\.6-sol '
        r'--thinking xhigh '
    ),
}
for label, pattern in required_driver_patterns.items():
    assert re.search(pattern, driver), f"pi-shepherd-loop.sh missing {label}"

auto_driver = (root / "scripts" / "pi-auto-loop.sh").read_text()
required_auto_driver_patterns = {
    "Sol coordinator default": r'ORCH_MODEL="\$\{ORCH_MODEL:-openai-codex/gpt-5\.6-sol\}"',
    "xhigh coordinator default": r'ORCH_THINKING="\$\{ORCH_THINKING:-xhigh\}"',
    "explicit coordinator thinking": r'--thinking "\$ORCH_THINKING"',
}
for label, pattern in required_auto_driver_patterns.items():
    assert re.search(pattern, auto_driver), f"pi-auto-loop.sh missing {label}"

active_policy_paths = [
    root / ".pi" / "README.md",
    root / ".pi" / "prompts" / "pm-auto-loop.md",
    root / ".pi" / "prompts" / "pm-orchestrate.md",
    root / ".agents" / "agentic-delivery" / "prompts" / "claude-orchestrator.md",
    root / ".agents" / "agentic-delivery" / "workflows" / "pi-autonomous-orchestration-loop.md",
    root / "docs" / "prompts" / "universal-programming-loop-prompts.md",
    root / "scripts" / "claude-auto-loop.sh",
    root / "scripts" / "pi-auto-loop.sh",
    root / "scripts" / "pi-shepherd-loop.sh",
]
stale = ("gpt-5.4-mini", "gpt-5.5")
for path in active_policy_paths:
    text = path.read_text()
    for value in stale:
        assert value not in text, f"{path}: stale active model route {value}"

print(
    f"pi model routing ok: {len(agent_paths)} agents; "
    "implementation=sol/high; all other roles=sol/xhigh; concurrency=4"
)
PY
