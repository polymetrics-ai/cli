#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

python3 - "$repo_root" <<'PY'
import pathlib
import re
import sys

root = pathlib.Path(sys.argv[1])
skill_dir = root / ".agents" / "skills" / "cli-architecture-v2-delivery"
skill_path = skill_dir / "SKILL.md"
reference_names = (
    "state-and-dependency-model.md",
    "phase-delivery-checklist.md",
    "parent-integration-and-review.md",
)
required_paths = [
    skill_path,
    skill_dir / "agents" / "openai.yaml",
    *(skill_dir / "references" / name for name in reference_names),
]
for path in required_paths:
    assert path.is_file(), f"missing required CLI Architecture v2 delivery skill file: {path}"

skill = skill_path.read_text()
assert skill.startswith("---\n"), "SKILL.md must start with YAML frontmatter"
parts = skill.split("---\n", 2)
assert len(parts) == 3, "SKILL.md frontmatter is unterminated"
frontmatter = parts[1]
assert re.search(r"(?m)^name:\s*cli-architecture-v2-delivery\s*$", frontmatter), (
    "SKILL.md frontmatter name must be cli-architecture-v2-delivery"
)
for trigger in (
    "CLI Architecture v2",
    "#397",
    "#438",
    "feat/cli-architecture-v2",
    "S0",
    "P01",
    "P22",
    "P18B",
    "D-TUI",
    "#398",
    "#437",
    "#453",
    "#462",
    "#463",
    "#469",
    "Cobra",
    "Viper",
    "typed configuration",
    "event bus",
    "TTY",
    "NDJSON",
    "Bubble Tea",
    "accessibility",
    "slog",
    "OpenTelemetry",
    "plain/JSON",
    "stacked PRs",
    "GSD",
    "TDD",
    "exact-head review",
    "Shepherd",
    "parent readiness",
):
    assert trigger in frontmatter, f"frontmatter description missing force trigger {trigger}"

required_skill_terms = (
    "parent-branch implementation truth",
    "parent_branch_satisfied_at",
    "active_ready",
    "dependency_blocked",
    "human_decision_blocked",
    "integrated_review_debt",
    "deferred_by_human",
    "default_branch_complete",
    "scripts/gsd doctor",
    "scripts/gsd list",
    "scripts/gsd sources",
    "manual universal loop",
    "PLAN",
    "TDD",
    "VERIFY",
    "exact-head",
    "stdout",
    "stderr",
    "JSON",
    "NDJSON",
    "Shepherd",
    "active review constraints",
    "plan → preview → approval → execute",
    "dependency approval",
    "human-only",
    "Definition of Done",
    "MUST",
    "SHOULD",
    "MAY",
    "bubble-tea-tui-design",
    "golang-spf13-cobra",
    "golang-spf13-viper",
    "golang-observability",
    "golang-benchmark",
    "golang-performance",
    "caveman",
)
for term in required_skill_terms:
    assert term in skill, f"SKILL.md missing required delivery contract term: {term}"

assert not re.search(r"\b[0-9a-f]{40}\b", skill, re.IGNORECASE), (
    "evergreen SKILL.md must not contain a 40-hex commit SHA"
)
assert "ready_queue" not in skill, "evergreen SKILL.md must not embed a current ready_queue"
for forbidden in (
    "invoke Claude",
    "invoke GitHub Copilot",
    "generic shell write",
    "generic HTTP write",
    "generic SQL write",
):
    assert forbidden not in skill, f"SKILL.md contains forbidden instruction: {forbidden}"
assert "go-tui-development" not in skill, "delivery skill must not route to a duplicate generic TUI skill"

agents = (root / "AGENTS.md").read_text()
required_routing = (
    root / ".agents" / "agentic-delivery" / "references" / "required-skills-routing.md"
).read_text()
matrix = (
    root / ".agents" / "agentic-delivery" / "matrices" / "task-skill-matrix.yaml"
).read_text()
bubble = (root / ".agents" / "skills" / "bubble-tea-tui-design" / "SKILL.md").read_text()
for label, text in (
    ("AGENTS.md", agents),
    ("required-skills-routing.md", required_routing),
    ("task-skill-matrix.yaml", matrix),
    ("bubble-tea-tui-design/SKILL.md", bubble),
):
    assert "cli-architecture-v2-delivery" in text, f"{label} does not route the delivery skill"
assert "cli_architecture_v2_delivery" in matrix, "task matrix missing delivery skill group"
assert "cli_architecture_v2_program" in matrix, "task matrix missing program task type"
assert "cli_architecture_v2_tui" in matrix, "task matrix missing TUI task type"

makefile = (root / "Makefile").read_text()
assert re.search(r"(?m)^cli-architecture-v2-skill-check:\s*$", makefile), (
    "Makefile missing focused cli-architecture-v2-skill-check target"
)
def make_rules(text: str, target: str) -> list[str]:
    lines = text.splitlines()
    starts = [
        index
        for index, line in enumerate(lines)
        if re.match(rf"^{re.escape(target)}\s*:", line)
    ]
    assert starts, f"Makefile missing {target} target"
    blocks: list[str] = []
    for start in starts:
        block = [lines[start]]
        for line in lines[start + 1 :]:
            if line and not line[0].isspace() and re.match(r"^[A-Za-z0-9_.-]+\s*:", line):
                break
            block.append(line)
        blocks.append("\n".join(block))
    return blocks

verify_rules = make_rules(makefile, "verify")
assert all("cli-architecture-v2-skill-check" not in rule for rule in verify_rules), (
    "focused skill check must not join any global make verify declaration or recipe without approval"
)
# Regressions: recipe-line and duplicate-target wiring must both be visible.
recipe_probe = "verify: test\n\t$(MAKE) cli-architecture-v2-skill-check\nnext:\n\ttrue\n"
assert "cli-architecture-v2-skill-check" in "\n".join(make_rules(recipe_probe, "verify")), (
    "Make rule parser must include recipe lines"
)
duplicate_probe = (
    "verify: test\n\ttrue\nother:\n\ttrue\n"
    "verify: cli-architecture-v2-skill-check\n\t$(MAKE) cli-architecture-v2-skill-check\n"
)
duplicate_rules = make_rules(duplicate_probe, "verify")
assert len(duplicate_rules) == 2, "Make rule parser must inspect duplicate target declarations"
assert "cli-architecture-v2-skill-check" in "\n".join(duplicate_rules), (
    "Make rule parser must include duplicate-target prerequisites and recipes"
)

for path in [skill_path, *(skill_dir / "references" / name for name in reference_names)]:
    text = path.read_text()
    for target in re.findall(r"\[[^\]]+\]\(([^)]+)\)", text):
        if target.startswith(("http://", "https://", "mailto:", "#")):
            continue
        clean = target.split("#", 1)[0]
        if not clean:
            continue
        resolved = (path.parent / clean).resolve()
        assert resolved.exists(), f"broken relative link in {path}: {target}"

print("CLI Architecture v2 delivery skill content and links ok")
PY

tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/cli-architecture-v2-skill-yaml.XXXXXX")"
tmp_go="$tmp_dir/main.go"
trap 'rm -rf "$tmp_dir"' EXIT
cat >"$tmp_go" <<'GO'
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	for _, path := range os.Args[1:] {
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		if strings.HasSuffix(path, "SKILL.md") {
			const opening = "---\n"
			if !bytes.HasPrefix(data, []byte(opening)) {
				panic(fmt.Errorf("parse %s: missing frontmatter", path))
			}
			rest := data[len(opening):]
			end := bytes.Index(rest, []byte("\n---\n"))
			if end < 0 {
				panic(fmt.Errorf("parse %s: unterminated frontmatter", path))
			}
			data = rest[:end]
		}
		var value any
		if err := yaml.Unmarshal(data, &value); err != nil {
			panic(fmt.Errorf("parse %s: %w", path, err))
		}
	}
}
GO
(
    cd "$repo_root"
    go run "$tmp_go" \
        .agents/skills/cli-architecture-v2-delivery/SKILL.md \
        .agents/skills/cli-architecture-v2-delivery/agents/openai.yaml \
        .agents/agentic-delivery/matrices/task-skill-matrix.yaml
)

printf 'CLI Architecture v2 delivery skill YAML ok\n'
