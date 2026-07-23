#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

python3 - "$repo_root" <<'PY'
from __future__ import annotations

import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1])
errors: list[str] = []


def read(relative: str) -> str:
    path = root / relative
    if not path.is_file():
        errors.append(f"missing canonical file: {relative}")
        return ""
    return path.read_text()


def require(relative: str, *needles: str) -> None:
    text = read(relative)
    for needle in needles:
        if needle not in text:
            errors.append(f"{relative}: missing required route marker {needle!r}")


def forbid(relative: str, patterns: dict[str, str]) -> None:
    text = read(relative)
    for label, pattern in patterns.items():
        if re.search(pattern, text, flags=re.IGNORECASE):
            errors.append(f"{relative}: forward PM route still contains {label}")


local_review = ".agents/agentic-delivery/workflows/local-codex-review-loop.md"
local_prompt = ".agents/agentic-delivery/prompts/local-codex-review-prompt.md"
require(
    local_review,
    "fresh-context",
    "exact base",
    "exact head",
    "disposition",
    "re-review",
    "Shepherd",
)
require(local_prompt, "exact base", "exact head", "read-only", "disposition")

forward_paths = [
    ".agents/agentic-delivery/contracts/parent-orchestrator-contract.md",
    ".agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md",
    ".agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md",
    ".agents/agentic-delivery/workflows/pi-active-orchestration-loop.md",
    ".agents/agentic-delivery/workflows/pi-autonomous-orchestration-loop.md",
    ".agents/agentic-delivery/workflows/codex-active-orchestration-loop.md",
    ".pi/prompts/pm-orchestrate.md",
    ".pi/prompts/pm-auto-loop.md",
    ".pi/prompts/pm-gsd-loop.md",
    ".pi/prompts/pm-review-loop.md",
    ".agents/agentic-delivery/agents/coordination/parent-issue-orchestrator.agent.yaml",
]
legacy_review_patterns = {
    "Claude review requirement": r"claude-review-loop|@claude\s+review|claude_(?:auto|manual|required)",
    "Copilot review route": r"copilot(?:_backup|\s+(?:backup\s+)?review)",
    "legacy PM disposition agent": r"pm-claude-review-disposition",
}
unavailable_command_patterns = {
    "unavailable programming-loop command": (
        r"/gsd-programming-loop|scripts/gsd\s+prompt\s+programming-loop|"
        r"\bgsd-programming-loop\b"
    ),
}
for relative in forward_paths:
    require(relative, "local-codex-review-loop.md", "shepherd-validator.md")
    forbid(relative, legacy_review_patterns)
    forbid(relative, unavailable_command_patterns)

require(
    ".agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md",
    "Canonical PM fallback",
    "/pm-orchestrate",
    "local-codex-review-loop.md",
    "shepherd-validator.md",
)
require(
    ".agents/agentic-delivery/references/gsd-pi-adapter.md",
    "programming-loop",
    "/pm-orchestrate",
    "local-codex-review-loop.md",
    "shepherd-validator.md",
)
require(
    ".agents/agentic-delivery/schemas/orchestration-state.schema.yaml",
    "local_codex",
    "shepherd",
    "legacy_review_route",
)

reviewer = read(".pi/agents/pm-reviewer.md")
if not re.search(r"^tools:.*\bbash\b", reviewer, flags=re.MULTILINE):
    errors.append(".pi/agents/pm-reviewer.md: exact-head reviewer lacks read-only bash/git access")
for marker in ("fresh-context", "exact base", "exact head", "Do not modify"):
    if marker not in reviewer:
        errors.append(f".pi/agents/pm-reviewer.md: missing {marker!r}")

legacy_agent = read(".pi/agents/pm-claude-review-disposition.md")
for marker in ("Deprecated", "local-codex-review-loop.md", "pm-reviewer"):
    if marker not in legacy_agent:
        errors.append(f".pi/agents/pm-claude-review-disposition.md: missing migration marker {marker!r}")

for relative in (
    ".agents/agentic-delivery/workflows/automated-review-routing-loop.md",
    ".agents/agentic-delivery/workflows/claude-review-loop.md",
):
    require(relative, "Legacy", "local-codex-review-loop.md", "not part of the canonical PM")

model_test = read("scripts/tests/pi-model-routing.sh")
if "scripts/tests/pm-orchestrator-contract.sh" not in model_test:
    errors.append("scripts/tests/pi-model-routing.sh: focused PM contract is not in make verify path")

if errors:
    raise SystemExit("PM orchestrator contract violations:\n- " + "\n- ".join(errors))

print(
    "pm orchestrator contract ok: one owner; unavailable-command fallback; "
    "exact-head local Codex review; independent Shepherd; no Claude/Copilot PM coverage"
)
PY
