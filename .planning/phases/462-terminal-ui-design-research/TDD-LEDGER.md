# Phase 462 TDD Ledger

Issue: #462
Starting commit: `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`

## RED contract inventory

Before this issue's documentation/skill edits, the exact base-commit audit produced:

```text
fatal: path 'docs/design/terminal-ui-research-and-design-system.md' exists on disk, but not in '6c038bb4'
docs/design/terminal-ui-research-and-design-system.md exit=128
fatal: path '.agents/skills/bubble-tea-tui-design/SKILL.md' exists on disk, but not in '6c038bb4'
.agents/skills/bubble-tea-tui-design/SKILL.md exit=128
fatal: path '.planning/phases/462-terminal-ui-design-research/PLAN.md' does not exist in '6c038bb4'
.planning/phases/462-terminal-ui-design-research/PLAN.md exit=128
base_pi_prompt_contract exit=1
base_tui_contract exit=1
```

The `rg` audits searched the base Pi prompt for `bubble-tea-tui-design`,
`Normal/Filter/Edit`, or `ntcharts`, and the base TUI contract for `NORMAL · results`,
`ntcharts/v2`, or the new research document. Both failed as required.

## GREEN ledger

| Slice | Evidence | Status |
|---|---|---|
| Reference lab | bpytop, Conky, CAVA, LazyGit, LazyDocker, fzf, Gum, catalog, and NTCharts launched in isolated tmux windows; safe keyboard/help paths exercised | Complete |
| Screenshots | Deterministic ANSI→HTML→PNG captures for all requested references plus NTCharts, stored outside the repo | Complete |
| Design system | reference synthesis, modes, keys, focus, responsive classes, charts/dashboard grammar, Bubble Tea rules, verification matrix | Complete |
| Skill | `bubble-tea-tui-design` and four focused reference files created with skill-creator scaffolding | Complete |
| Program docs | design, ADR, plan, execution prompt, roadmap, issue backlog, and Pi prompt trace updated | Complete |
| Live issue routing | #408/#409/#411/#412/#414/#416/#418 updated; #462 nested under #397; chart #463 nested under #411 | Complete |

## Refactor and verification ledger

| Check | Expected | Status |
|---|---|---|
| skill quick validation | `Skill is valid!` with isolated PyYAML validator environment | Pass |
| YAML parse/frontmatter and no scaffold placeholder | skill validator and focused scan clean | Pass |
| `git diff --check` | exit 0 | Pass |
| GSD doctor/sources | 69 commands healthy; official plan-phase provenance printed | Pass |
| design/GSD/issue reference grep | every affected phase routed; live marker=true for seven UI issues | Pass |
| `go.mod`/`go.sum`/production diff | no delta under dependency, `cmd`, `internal`, CLI docs, website, or connector definitions | Pass |
| documentation gates | `make docs-check`: build + connector-doc validation pass | Pass |

No production test was appropriate because this phase changes only documentation, planning,
agent instructions, and a skill. Production UI issues retain strict test-first behavior.
