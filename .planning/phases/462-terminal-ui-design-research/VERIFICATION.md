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

- [ ] Phase artifacts reopened before delegated docs/design edits.
- [ ] GSD adapter health recorded: `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 462 --skip-research`.
- [ ] Manual universal-loop fallback recorded because `scripts/gsd prompt programming-loop ...` is absent.
- [ ] Loaded skills and missing `.pi/skills/go-implementation` mismatch recorded.
- [ ] RED grep inventory recorded in `TDD-LEDGER.md`.
- [ ] Bare namespace contract grep: no bare `pm query`/bare `pm reverse` TUI-launch wording remains.
- [ ] Explicit interactive subcommands are consistent: `pm query grid`, `pm reverse guide`.
- [ ] Approval-token contract grep: no wording says final frames/transcripts/logs/accessibility/JSON/shell-equivalent text prints tokens.
- [ ] Dependency rows/rosters encode `#462` or `D-TUI` directly for #408, #409, #411, #412,
  #414, #416, #418, and #463 where applicable.
- [ ] #462 status says provisionally integrated / review blocked with PR #465, head
  `6853fee28e0208381b49931fb1f5dfec42ee50ef`, Claude disabled, Copilot quota exhausted,
  fallback human, accepted correction PR pending.
- [ ] Query export path contract includes typed read-only export, project-scoped default,
  control-character/traversal/broad-path rejection, clean/confined path, symlink race rejection,
  no overwrite default, TTY confirmation, noninteractive `--force`, sanitized command echo, and
  exact `--no-input` guidance.
- [ ] Skill quick validation passes.
- [ ] JSON/YAML/Markdown syntax checks pass as applicable.
- [ ] `scripts/gsd doctor` passes at final verification.
- [ ] `git diff --check` passes.
- [ ] Exact scope check shows no `cmd/**`, `internal/**`, `go.mod`, `go.sum`, `website/**`,
  `docs/cli/**`, or nondelegated parent phase artifact changes.
- [ ] `make docs-check` run if safe/feasible; if not, blocker recorded.

### Review route status

Claude review remains unavailable (`disabled_manually` on PR #465 context). Copilot backup already
reported quota exhausted and must not be retried in this blocker window. Human/parent orchestrator
review remains pending for the accepted correction PR.
