# Phase 462 Verification

## Research and design

- [x] Every user-requested reference application was installed/built and safely exercised.
- [x] Screenshots exist outside the repository and contain no intentionally supplied secret.
- [x] Bubble Tea v2/Bubbles/Lip Gloss primary guidance and GitHub CLI accessibility guidance reviewed.
- [x] NTCharts v2 quickstart runs in the isolated research clone; no Polymetrics dependency changed.
- [x] Adopt/adapt/avoid decisions prevent generic shell/HTTP/SQL writes and unsafe mutation shortcuts.
- [x] Chart grammar specifies axes, units, exact text, bounds/downsampling, and accessible table fallback.

## Repository contract

- [x] Skill validates and contains no scaffold TODOs.
- [x] Required skill routing names `bubble-tea-tui-design` for every TUI task.
- [x] Design, ADR, program plan, execution prompt, roadmap, backlog, and Pi prompt agree.
- [x] GSD phase artifacts are complete and identify the docs-only RED/GREEN/refactor path.
- [x] Live issues #408, #409, #411, #412, #414, #416, and #418 name the design gate and skill.
- [x] Dedicated chart issue #463 is a child of #411 and records the unapproved dependency status.

## Gates

- [x] `scripts/gsd doctor` — 69 commands healthy
- [x] skill `quick_validate.py` — `Skill is valid!`
- [x] YAML/frontmatter parse and forbidden placeholder scan
- [x] `git diff --check`
- [x] link/reference and affected-issue grep
- [x] `git diff --exit-code <start> -- go.mod go.sum cmd internal docs/cli website internal/connectors/defs`
- [x] `make docs-check` — build and connector-doc validation pass
- [x] branch committed/pushed and stacked PR #465 opened to `feat/cli-architecture-v2`

## CLI parity applicability

No runtime command, flag, output, help topic, bare namespace, generated CLI document, or website
page changes in this issue. The execution and worker prompts strengthen the parity requirements
for the later behavior-changing UI phases.

## Review correction checklist — 2026-07-20

### Required checks before handoff

- [x] Phase artifacts reopened before delegated docs/design edits.
- [x] GSD adapter health recorded: `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 462 --skip-research`.
- [x] Manual universal-loop fallback recorded because `scripts/gsd prompt programming-loop ...` is absent.
- [x] Loaded skills and missing `.pi/skills/go-implementation` mismatch recorded.
- [x] RED grep inventory recorded in `TDD-LEDGER.md`.
- [x] Bare namespace contract grep: no bare `pm query`/bare `pm reverse` TUI-launch wording remains
  in current docs outside the historical RED ledger.
- [x] Explicit interactive subcommands are consistent: `pm query grid`, `pm reverse guide`.
- [x] Approval-token contract grep: no current-doc wording says final frames/transcripts/logs/
  accessibility/JSON/shell-equivalent text prints tokens.
- [x] Dependency rows/rosters encode `#462` or `D-TUI` directly for #408, #409, #411, #412,
  #414, #416, #418, and #463 where applicable.
- [x] #462 status says provisionally integrated / review blocked with PR #465, head
  `6853fee28e0208381b49931fb1f5dfec42ee50ef`, Claude disabled, Copilot quota exhausted,
  fallback human, and correction PR #467 starting head/status captured explicitly.
- [x] Query export path contract includes typed read-only export, project-scoped default,
  control-character/traversal/broad-path rejection, clean/confined path, symlink race rejection,
  no overwrite default, confirmation only when stdin/stdout are TTYs, noninteractive `--force`,
  sanitized command echo, and exact `--no-input` guidance.
- [x] Skill quick validation passes: `Skill is valid!`.
- [x] JSON/YAML/Markdown syntax checks pass as applicable: PyYAML skill frontmatter and
  `python3 -m json.tool` for `RUN-STATE.json`.
- [x] `scripts/gsd doctor` passes at final verification (`ok commands 69`).
- [x] `git diff --check` passes.
- [x] Exact scope check shows no `cmd/**`, `internal/**`, `go.mod`, `go.sum`, `website/**`,
  `docs/cli/**`, or nondelegated parent phase artifact changes.
- [x] `make docs-check` ran: `go build ./cmd/pm`; `Validated connector docs in docs/connectors`.

### Review route status

Claude review remains unavailable (`disabled_manually` on PR #465 context). Copilot backup already
reported quota exhausted and must not be retried in this blocker window. Human/parent orchestrator
review remains pending for the accepted correction PR.

## Correction PR #467 TTY-gate/state checklist — 2026-07-20

### Required checks before handoff

- [x] Phase artifacts reopened before delegated source/design/skill docs edits.
- [x] GSD adapter health recorded: `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 462 --skip-research`.
- [x] Manual universal-loop fallback recorded because `scripts/gsd prompt programming-loop ...` is absent.
- [x] Loaded skills and missing `.pi/skills/go-implementation` mismatch recorded.
- [x] RED docs-contract grep recorded for stdout-only/ambiguous TUI gates and missing
  `stdin-piped+stdout-TTY` / `stdout-piped` future test markers.
- [x] Bubble Tea/Huh/prompt activation contract requires both stdin and stdout TTYs in ADR,
  design docs, source plan, execution prompt, skill references, roadmap/backlog/Pi prompts, and
  phase artifacts.
- [x] Piped/non-TTY stdin fallback is explicit: deterministic plain/noninteractive behavior, no
  scripted-stdin consumption, no hang, and no `/dev/tty` bypass.
- [x] Future RED test matrix is recorded for `stdin-piped+stdout-TTY`, `stdout-piped`, `CI`,
  `--json`, `--plain`, and `--no-input`.
- [x] Explicit `pm query grid`, `pm reverse guide`, query export, approval-token secrecy, and
  accessibility/plain fallback contracts remain present.
- [x] RUN-STATE.json/RUN-STATE.md/SUMMARY record #467 open at starting head
  `e8286ea83a76ac2c6f6257c6e2d40fd21af81640`, CI green at that head, and human/parent review
  pending; no generic pending placeholder or invented final-head claim remains.
- [x] Contradiction grep passes.
- [x] `python3 -m json.tool .planning/phases/462-terminal-ui-design-research/RUN-STATE.json` passes.
- [x] Skill quick validation passes: `Skill is valid!`.
- [x] Direct dependency/token/export contracts unchanged.
- [x] `git diff --check` passes.
- [x] Exact scope check shows no `cmd/**`, `internal/**`, `go.mod`, `go.sum`, `website/**`,
  `docs/cli/**`, parent #397 artifacts, or nondelegated docs changes.
- [x] `scripts/gsd doctor` passes at final verification (`ok commands 69`).
- [x] `make docs-check` passes.

### Review route status

Do not retry Claude or Copilot in this blocker window. Final route remains human/parent review and
parent integration gates after local finding disposition.
