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
  human fallback, and correction PR #467 starting head/status captured explicitly.
- Query export path contract is explicit and preserves no generic write-tool boundary.

## Review correction GREEN ledger — 2026-07-20

| Check | Evidence | Status |
|---|---|---|
| Bare namespace contract | contradiction grep for old bare `pm query`/`pm reverse` TUI-launch wording returned no current-doc matches outside the historical RED ledger | Pass |
| Explicit subcommands | `pm query grid` and `pm reverse guide` markers present in design docs, ADR, plan, prompt, Pi prompt, skill, and phase artifacts | Pass |
| Approval token secrecy | old token-display wording absent from current docs; new contract marks approval tokens sensitive one-time values, ephemeral in-memory only, never rendered in final frames/transcripts/logs/screenshots/accessibility/JSON/shell-equivalent text/fixtures | Pass |
| Dependency rows | Python roster check confirmed direct `#462`/`D-TUI` markers for #408, #409, #411, #412, #414, #416, #418, and #463 across roadmap, backlog, source plan, and execution prompt | Pass |
| Query export path | marker check confirmed typed read-only export, project-scoped default, control-character/traversal/broad-path/symlink rejection, no-overwrite default, confirmation only when stdin/stdout are TTYs, noninteractive `--force`, sanitized command echo, and exact `--no-input` guidance | Pass |
| Skill validation | PyYAML frontmatter/reference check printed `Skill is valid!` | Pass |
| JSON syntax | `python3 -m json.tool .planning/phases/462-terminal-ui-design-research/RUN-STATE.json` | Pass |
| Scope | diff against `c91b90cf9671b5caabc0ef4ec24d81897f870458` contains only delegated docs/skill/#462 phase artifacts and no `go.mod`, `go.sum`, `cmd`, `internal`, `website`, or `docs/cli` changes | Pass |
| Whitespace | `git diff --check` | Pass |
| GSD health | `scripts/gsd doctor` printed `ok commands 69` | Pass |
| Docs gate | `make docs-check` built `./cmd/pm` and printed `Validated connector docs in docs/connectors` | Pass |

No production behavior test was required; the correction is docs/planning/skill only.

## Correction PR #467 TTY-gate/state RED ledger — 2026-07-20

GSD/Pi evidence recaptured before delegated source/design/skill docs edits:

```text
scripts/gsd doctor                         # pass, ok commands 69
scripts/gsd prompt plan-phase 462 --skip-research > /tmp/gsd-plan-462-correction-467.txt
wc -c /tmp/gsd-plan-462-correction-467.txt # 10692
scripts/gsd prompt programming-loop init --phase 462 --dry-run
# scripts/gsd: unknown GSD command: programming-loop
```

Manual universal-loop fallback remains active for this correction because `programming-loop` is not
registered in the shell adapter. Execution decision: `local_critical_path` — one assigned worker,
isolated cwd/branch, no subagent tool, docs-only accepted-review correction. PR #467 starting
snapshot: open at `e8286ea83a76ac2c6f6257c6e2d40fd21af81640`, CI green at that head,
human/parent review pending.

Loaded skills for this correction: `gsd-core`, `caveman`, `bubble-tea-tui-design`,
`golang-how-to`, `golang-cli`, `golang-testing`, `golang-documentation`, `golang-security`,
`golang-safety`, `golang-context`, `golang-concurrency`, and `golang-error-handling`. Skill-route
mismatch persists: `.pi/skills/go-implementation/SKILL.md` is absent.

### TTY gate missing-contract inventory

```text
RED: TTY gate contract inconsistent
docs/adr/0003-interactive-tui-layer.md:38:2. **The TUI is affirmatively gated.** `ui.Detect` enables it only when stdout is a real
docs/design/tui-ux-design.md:23:and always exists first. The TTY door opens only when `ui.Detect` says so
docs/design/tui-ux-design.md:488:- **Gate**: `ui.Detect` table tests (pipes are non-TTY by construction; env/flag matrix).
docs/plans/cli-architecture-v2-improvement-plan.md:63:**D. TUIs must be gated, degradable, and honest.** TTY-only activation; colorprofile
docs/plans/cli-architecture-v2-improvement-plan.md:247:- **TTY gate** `ui.Detect`: TUI only when stdout is a TTY ∧ ¬`--json` ∧ ¬`--plain` ∧
docs/prompts/cli-architecture-v2-gsd-execution-prompt.md:127:ui.Detect per ADR-0003 §2 (TTY ∧ ¬json ∧ ¬plain ∧ ¬no-input ∧ ¬PM_NO_TUI ∧ ¬CI ∧ TERM≠dumb);
.agents/skills/bubble-tea-tui-design/references/testing-and-accessibility.md:12:`CI=1`, `PM_NO_TUI=1`, `--plain`, `--json`, and pipes bypass the TUI.
MISSING future RED test marker: stdin-piped+stdout-TTY
MISSING future RED test marker: stdout-piped
```

RED status: docs-contract grep failed as expected. The defect is documentation/prompt
contradiction only: some delegated sources required stdout TTY or ambiguous `TTY`, while the plan
correctly required both stdin and stdout TTYs. No production Go behavior test is required in this
correction, but future TUI implementation issues must start with failing tests for
`stdin-piped+stdout-TTY` fallback, `stdout-piped`, `CI`, `--json`, `--plain`, and `--no-input`.

### Expected GREEN evidence for #467 accepted findings

- All normative Bubble Tea/Huh/prompt activation language says both stdin and stdout must be TTYs.
- Piped/non-TTY stdin always takes deterministic plain/noninteractive fallback, never consumes
  scripted stdin unexpectedly, never hangs, and never opens `/dev/tty` to bypass the gate.
- Existing disables stay normative: stdout-piped/non-TTY, `CI`, `PM_NO_TUI`, `TERM=dumb`, `--json`,
  `--plain`, and `--no-input` all bypass TUI/prompt activation.
- Future RED matrix is recorded in design docs, skill/test reference, execution/Pi prompts, and
  phase verification.
- Query grid/reverse guide, query export, approval-token secrecy, and accessibility/plain behavior
  remain explicit and unchanged.
- Run state records #467 open at starting head `e8286ea83a76ac2c6f6257c6e2d40fd21af81640`, CI green
  at that head, and review status human/parent pending; no artifact claims an invented final head.

## Correction PR #467 GREEN ledger — 2026-07-20

| Check | Evidence | Status |
|---|---|---|
| TTY gate alignment | Python docs-contract contradiction grep over ADR, design docs, skill references, roadmap/backlog/Pi prompts, execution prompt, and phase artifacts printed `PASS docs-contract contradiction grep: stdin+stdout TTY gate aligned; future RED matrix present` | Pass |
| Future RED matrix | Markers present for `stdin-piped+stdout-TTY`, `stdout-piped`, `CI`, `--json`, `--plain`, and `--no-input` in design docs, skill/test reference, execution prompt, Pi prompt, roadmap, and phase artifacts | Pass |
| Scripted stdin safety | Delegated sources say piped/non-TTY stdin falls back to deterministic plain/noninteractive behavior without consuming scripted stdin, hanging, or using `/dev/tty` | Pass |
| Query/reverse/accessibility preservation | Marker check confirmed `pm query grid`, bare `pm query` help, `pm reverse guide`, bare `pm reverse` help, accessibility/plain, approval-token secrecy, and typed read-only query export contracts remain present | Pass |
| State honesty | RUN-STATE.json/RUN-STATE.md/SUMMARY record PR #467 open at starting head `e8286ea83a76ac2c6f6257c6e2d40fd21af81640`, CI green at that head, human/parent review pending, and no invented final-head claim | Pass |
| Skill validation | PyYAML frontmatter/reference check printed `Skill is valid!` | Pass |
| JSON syntax | `python3 -m json.tool .planning/phases/462-terminal-ui-design-research/RUN-STATE.json` | Pass |
| Scope | Exact scope check passed; no `cmd/**`, `internal/**`, `go.mod`, `go.sum`, `website/**`, `docs/cli/**`, parent #397 phase artifacts, or nondelegated docs changed | Pass |
| Whitespace | `git diff --check` | Pass |
| GSD health | `scripts/gsd doctor` printed `ok commands 69` | Pass |
| Docs gate | `make docs-check` built `./cmd/pm` and printed `Validated connector docs in docs/connectors` | Pass |

No production behavior test was required; this correction is docs/planning/skill/prompt only. Full
`make verify` was not run because scope stayed docs-only and the issue requested `make docs-check`.

## Follow-up PR #468 RED ledger — 2026-07-20

GSD/Pi evidence recaptured before delegated source/design/skill docs edits:

```text
scripts/gsd doctor                         # pass, ok commands 69
scripts/gsd prompt plan-phase 462 --skip-research > /tmp/gsd-plan-462-pr468.txt
wc -c /tmp/gsd-plan-462-pr468.txt          # 10716
scripts/gsd prompt programming-loop init --phase 462 --dry-run
# scripts/gsd: unknown GSD command: programming-loop
```

Manual universal-loop fallback remains active because `programming-loop` is not registered in the
shell adapter. Execution decision: `local_critical_path` — one assigned worker, isolated cwd/branch,
no subagent tool, docs-only accepted local review correction. PR state snapshot: PR #467 merged at
parent commit `93a117100c6421955262aa32794a91a158d267e1` from old head
`e8286ea83a76ac2c6f6257c6e2d40fd21af81640`; PR #468 open at starting head
`fd122c52458a6ef0db12f60f303c261ed2e63d4c`, human review pending. Git/GitHub remain the live
source of truth after that starting snapshot; no self-referential final-head claim is allowed.
Local sidecar review is local evidence only and does not count as external review coverage.

Loaded skills for this correction: `gsd-core`, `caveman`, `bubble-tea-tui-design`,
`golang-how-to`, `golang-cli`, `golang-testing`, `golang-documentation`, `golang-security`,
`golang-safety`, `golang-context`, `golang-concurrency`, and `golang-error-handling`.
Skill-route mismatch persists: `.pi/skills/go-implementation/SKILL.md` is absent.

### Follow-up missing-contract inventory

```text
RED: plain prompt contradiction scan (delegated docs/skills)
docs/design/tui-ux-design.md:317:- Accessible/plain: `pm query grid --plain` or accessibility mode offers a sequential prompt for
docs/design/tui-ux-design.md:363:- Accessible: browse falls back to a sequential prompt (`filter? capability?`) then plain
.agents/skills/bubble-tea-tui-design/references/testing-and-accessibility.md:14:   `--json`, `--plain`, and `--no-input`; each must bypass the TUI/prompt path without consuming

RED: Stage 16 shared TTY fallback RED matrix check
FAIL Stage 16 missing matrix markers: stdin+stdout TTY activation, stdin-piped+stdout-TTY, stdout-piped, CI, --json, --plain, --no-input

RED: stale artifact state check
FAIL artifacts missing follow-up markers: 93a117100c6421955262aa32794a91a158d267e1, fd122c52458a6ef0db12f60f303c261ed2e63d4c, PR #468
```

RED status: docs-contract grep failed as expected. The `testing-and-accessibility.md` hit is a
safe bypass-marker line, not a contradiction; it stays as a required bypass assertion. The defects
are documentation/prompt/artifact contradictions only, so no production Go behavior test is required
in this correction. Future TUI implementation issues must add failing tests for the shared fallback
matrix before code.

### Expected GREEN evidence for PR #468 accepted findings

- Query grid docs no longer state that `--plain` runs a sequential prompt; `--plain`, `--json`,
  and `--no-input` always bypass Bubble Tea, Huh, and prompts.
- Delegated docs/skill wording allows sequential prompts only in explicit accessible mode when both
  stdin and stdout are TTYs and no bypass flag is set.
- Bypass paths use deterministic table/summary output or exact required-flag errors only.
- Shared TUI preflight plus Stage 16 gate name stdin+stdout TTY activation,
  `stdin-piped+stdout-TTY`, `stdout-piped`, `CI`, `--json`, `--plain`, and `--no-input`, while
  surface-specific tests remain intact.
- Run state records #467 merged at parent commit `93a117100c6421955262aa32794a91a158d267e1` from
  old head `e8286ea83a76ac2c6f6257c6e2d40fd21af81640`; #468 open at starting head
  `fd122c52458a6ef0db12f60f303c261ed2e63d4c` with human review pending; GitHub live source and
  local-sidecar-not-external-coverage caveats present; no invented final head.

## Follow-up PR #468 GREEN ledger — 2026-07-20

| Check | Evidence | Status |
|---|---|---|
| Prompt bypass contract | Docs-contract grep over delegated docs/skill/prompt sources printed `PASS docs-contract contradiction grep: prompt bypass and shared/Stage16 matrix aligned` | Pass |
| Query grid `--plain` correction | Old `pm query grid --plain` sequential-prompt wording absent; bypass docs now require deterministic table/summary output or exact required-flag errors only | Pass |
| Explicit accessible mode | Sequential prompt wording is limited to explicit accessible mode after stdin+stdout TTY gate and no `--plain`/`--json`/`--no-input` bypass flag | Pass |
| Shared + Stage 16 matrix | Execution prompt TUI preflight and Stage 16 gate both include stdin+stdout TTY activation, `stdin-piped+stdout-TTY`, `stdout-piped`, `CI`, `--json`, `--plain`, and `--no-input` | Pass |
| Prior corrections preserved | Marker check confirmed `pm query grid`, `pm reverse guide`, bare contextual help, approval-token nondisclosure, typed read-only query export, direct `#462`/`D-TUI`, and no `/dev/tty` bypass remain present | Pass |
| Artifact state honesty | RUN-STATE/SUMMARY record PR #467 merged at parent commit `93a117100c6421955262aa32794a91a158d267e1` from old head `e8286ea83a76ac2c6f6257c6e2d40fd21af81640`, PR #468 open at starting head `fd122c52458a6ef0db12f60f303c261ed2e63d4c`, human review pending, GitHub as live source, local sidecar not external coverage, and no final-head claim | Pass |
| Skill validation | frontmatter/reference validation printed `Skill is valid!` | Pass |
| JSON syntax | `python3 -m json.tool .planning/phases/462-terminal-ui-design-research/RUN-STATE.json` | Pass |
| Scope | Exact scope check passed; no `cmd/**`, `internal/**`, `go.mod`, `go.sum`, `website/**`, `docs/cli/**`, parent #397 artifacts, or nondelegated docs changed | Pass |
| Whitespace | `git diff --check` | Pass |
| GSD health | `scripts/gsd doctor` printed `ok commands 69` | Pass |
| Docs gate | `make docs-check` built `./cmd/pm` and printed `Validated connector docs in docs/connectors` | Pass |

No production behavior test was required; this correction is docs/planning/skill/prompt only. Full
`make verify` was not run because scope stayed docs-only and the issue requested `make docs-check`.
