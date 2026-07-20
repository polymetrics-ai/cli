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

## Review correction RED ledger — 2026-07-20

GSD/Pi evidence captured before delegated docs/design edits:

```text
scripts/gsd doctor                         # pass, 69 commands
scripts/gsd prompt plan-phase 462 --skip-research > /tmp/gsd-plan-462.txt
wc -c /tmp/gsd-plan-462.txt                 # 10692
scripts/gsd prompt programming-loop init --phase 462 --dry-run
# scripts/gsd: unknown GSD command: programming-loop
.pi/skills/go-implementation/SKILL.md       # absent
```

Manual universal-loop fallback is active because `programming-loop` is not registered in the
repo-local GSD adapter. Execution decision for this worker cycle: `local_critical_path` — one
assigned worker, isolated cwd, no subagent tool, docs-only accepted-review correction.

Loaded skills recorded for this correction: `gsd-core`, `caveman`, `bubble-tea-tui-design`,
`golang-how-to`, `golang-cli`, `golang-testing`, `golang-documentation`, `golang-security`, plus
routed `golang-safety`, `golang-context`, and `golang-concurrency`. Skill-route mismatch:
`.pi/skills/go-implementation/SKILL.md` is absent; used available skill paths.

### Missing-contract inventory

```text
rg bare namespace / interactive launch:
docs/design/tui-ux-design.md:275:**`pm query` interactive (bare `pm query` on a TTY):** alt-screen browser.

rg approval token:
docs/prompts/cli-architecture-v2-gsd-execution-prompt.md:224:Guided reverse-ETL session ... tokens relayed
docs/plans/cli-architecture-v2-improvement-plan.md:258:guided reverse-ETL session (token relay handled in-session)
docs/plans/cli-architecture-v2-improvement-plan.md:371-372:No changes ... relays the existing tokens
docs/design/tui-ux-design.md:387-392:approval token is relayed ... Every intermediate ID/token is printed

rg dependency rows:
.planning/ROADMAP.md rows for #408/#409/#411/#412/#414/#416 omit direct #462 in Dependency gate;
.planning/traces/cli-architecture-v2-issue-backlog.md P10/P11/P14/P16/P18/P20 omit direct D-TUI;
docs/plans/cli-architecture-v2-improvement-plan.md phase table omits explicit design-gate text in rows;
docs/prompts/cli-architecture-v2-gsd-execution-prompt.md stage headings depend on prose preflight only.

rg status/review:
.planning/phases/462-terminal-ui-design-research/SUMMARY.md:3:Status: delivered in stacked PR #465; CI and automated review pending.
.planning/phases/462-terminal-ui-design-research/RUN-STATE.md:9:state: awaiting_ci_and_automated_review

rg query export:
docs/design/tui-ux-design.md:297-298:Export (`e`): write JSONL/CSV to a chosen path; always echoes ...
docs/design/terminal-ui-research-and-design-system.md:220:export serializes the underlying rows, never the glyph rendering;
```

RED status: docs-contract grep inventory failed as expected. No Go behavior test required because
this correction changes only delegated documentation/planning/skill artifacts.

### Expected GREEN evidence

- Bare namespace wording removed: no design text claims bare `pm query` or bare `pm reverse` starts
  a TUI; explicit interactive subcommands are `pm query grid` and `pm reverse guide`.
- Approval-token display wording removed and replaced by ephemeral in-memory one-time authorization
  contract with typed approval preserved.
- Dependency rows directly encode `#462`/`D-TUI` for affected TUI issues.
- Phase status says provisionally integrated / review blocked with PR #465, head
  `6853fee28e0208381b49931fb1f5dfec42ee50ef`, Claude disabled, Copilot quota exhausted,
  human fallback, accepted-correction PR pending.
- Query export path contract is explicit and preserves no generic write-tool boundary.

## Review correction GREEN ledger — 2026-07-20

| Check | Evidence | Status |
|---|---|---|
| Bare namespace contract | contradiction grep for old bare `pm query`/`pm reverse` TUI-launch wording returned no current-doc matches outside the historical RED ledger | Pass |
| Explicit subcommands | `pm query grid` and `pm reverse guide` markers present in design docs, ADR, plan, prompt, Pi prompt, skill, and phase artifacts | Pass |
| Approval token secrecy | old token-display wording absent from current docs; new contract marks approval tokens sensitive one-time values, ephemeral in-memory only, never rendered in final frames/transcripts/logs/screenshots/accessibility/JSON/shell-equivalent text/fixtures | Pass |
| Dependency rows | Python roster check confirmed direct `#462`/`D-TUI` markers for #408, #409, #411, #412, #414, #416, #418, and #463 across roadmap, backlog, source plan, and execution prompt | Pass |
| Query export path | marker check confirmed typed read-only export, project-scoped default, control-character/traversal/broad-path/symlink rejection, no-overwrite default, TTY confirmation/noninteractive `--force`, sanitized command echo, and exact `--no-input` guidance | Pass |
| Skill validation | PyYAML frontmatter/reference check printed `Skill is valid!` | Pass |
| JSON syntax | `python3 -m json.tool .planning/phases/462-terminal-ui-design-research/RUN-STATE.json` | Pass |
| Scope | diff against `c91b90cf9671b5caabc0ef4ec24d81897f870458` contains only delegated docs/skill/#462 phase artifacts and no `go.mod`, `go.sum`, `cmd`, `internal`, `website`, or `docs/cli` changes | Pass |
| Whitespace | `git diff --check` | Pass |
| GSD health | `scripts/gsd doctor` printed `ok commands 69` | Pass |
| Docs gate | `make docs-check` built `./cmd/pm` and printed `Validated connector docs in docs/connectors` | Pass |

No production behavior test was required; the correction is docs/planning/skill only.
