# PLAN — Issue 405 TTY gate and NDJSON progress

Sub-issue: https://github.com/polymetrics-ai/cli/issues/405
Parent: https://github.com/polymetrics-ai/cli/issues/397 / PR #438
Branch: `feat/405-tty-ndjson-progress`
Base branch: `feat/cli-architecture-v2`
Worker directory: `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-405-tty-ndjson-progress`
Starting head: `d8e532eb20d24d982c09772ecae48abb3bb64271`
Mode: bounded mutating worker in isolated cwd.

## Required reading complete

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`
- `docs/plans/universal-programming-loop-prd.md`, `docs/prompts/universal-programming-loop-prompts.md`
- `docs/plans/cli-architecture-v2-improvement-plan.md` Stage 7 / Pillar B
- `docs/prompts/cli-architecture-v2-gsd-execution-prompt.md` Stage 7
- `docs/design/tui-ux-design.md` gate, palette, glyph, agent-parity sections
- `docs/adr/0002-cobra-viper-cli-framework.md`, `docs/adr/0003-interactive-tui-layer.md`, `docs/adr/0004-opentelemetry-observability.md`
- Issue #405 body via `gh issue view 405 --json ...`

## GSD adapter evidence

- `scripts/gsd doctor` passed.
- `scripts/gsd list` passed and showed 69 commands.
- `scripts/gsd prompt plan-phase 405 --skip-research` generated the official `/gsd-plan-phase 405 --skip-research` prompt.
- `scripts/gsd prompt programming-loop init --phase 405 --dry-run` failed: `scripts/gsd: unknown GSD command: programming-loop`.
- Adapter gap fallback: loaded `.pi/prompts/pm-gsd-loop.md`; running the GSD universal programming loop inline/manual. Record each cycle as `local_critical_path`, not `spawned`.

## Skills loaded

Routing source: `.agents/agentic-delivery/references/required-skills-routing.md`.

- `gsd-core`
- `caveman` for compact handoff only
- `golang-how-to`
- `golang-cli`
- `golang-spf13-cobra`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-documentation`
- UI/design source docs: `docs/design/tui-ux-design.md`, `docs/adr/0003-interactive-tui-layer.md`

Stack implementation skill note: `.pi/skills/go-implementation/SKILL.md` was requested by worker instructions but is absent in this checkout (`ENOENT`); loaded `gsd-core` plus required Go skills from required-skills-routing instead.

## Scope

Allowed writes:

- `internal/cli/**` run options/global UI flags and tests
- `internal/ui/**` TTY detection and style foundation
- `docs/cli/**` generated/manual CLI parity docs for changed help text
- `website/**` CLI reference parity docs
- issue-local `.planning/phases/405-tty-ndjson-progress/**`
- `go.mod`/`go.sum` only for the Stage 7-approved `golang.org/x/term` dependency

Explicit non-scope:

- no shared parent orchestration artifacts (`.planning/traces/cli-architecture-v2-orchestration-state.yaml`)
- no #423 perf namespace work
- no #410 telemetry/OTel span work
- no connector bundle changes
- no reverse ETL gate changes
- no new dependencies beyond approved `golang.org/x/term`
- no credentialed/runtime-backed checks

## Acceptance criteria mapping

1. TUI activates only under ADR conditions: stdout TTY, not `--json`, not `--plain`, not `--no-input`, `PM_NO_TUI` unset, `CI` unset, `TERM != dumb`.
2. `cli.RunWithOptions` added; existing `Run` delegates with plain mode so current tests stay on the plain path.
3. `--plain`, `--no-input`, `--progress ndjson` parsed, documented, and tested.
4. `--progress ndjson` wires sanitized `events.NDJSON` to stderr only; JSON stdout remains one final envelope.
5. `internal/ui/styles/**` palette/glyph foundation supports no-color and ASCII fallback constraints.

## Slice plan

### Slice 1 — UI detection and style foundation

Red tests first:

- `internal/ui` table tests for ADR gate matrix (`stdout` pipe, `--json`, `--plain`, `--no-input`, `PM_NO_TUI`, `CI`, `TERM=dumb`, happy TTY).
- `internal/ui/styles` tests for no-color palette degradation, accessible/ANSI palette selection, and ASCII glyph/rail fallback.

Implementation:

- Add `internal/ui` detector with injectable env and stdout-TTY facts.
- Add `internal/ui/styles` semantic tokens and glyph vocabulary with ASCII fallback.
- Use `golang.org/x/term` in CLI detection only; no charm deps.

### Slice 2 — CLI RunWithOptions and global flags

Red tests first:

- `RunWithOptions` auto mode observes simulated TTY detection while `Run` remains plain by construction.
- Global `--plain` and `--no-input` are stripped before command dispatch and force plain detection.
- invalid `--progress` values produce validation errors without dispatch.

Implementation:

- Add `RunOptions`, `RunMode`, and `RunWithOptions`.
- Extend global parser for `--plain`, `--no-input`, `--progress ndjson`.
- Keep config bootstrap semantics for `--root` and `--json` unchanged.

### Slice 3 — NDJSON progress to stderr

Red tests first:

- `--progress ndjson` with a local fixture flow emits NDJSON progress on stderr and exactly one JSON result on stdout.
- Progress output is absent from stdout; progress events decode line-by-line from stderr.
- `CI=1`, `PM_NO_TUI=1`, and `--plain` still keep plain output.

Implementation:

- Wrap invocation context with `events.NewNDJSON(stderr)` only for `--progress ndjson`.
- Leave default emitter as `Nop`; no polling, services, or credentials.

### Slice 4 — CLI help/docs/website parity

Red/validation first:

- Tests assert runtime help documents global UI/progress flags.
- Generated `docs/cli/**` diff catches drift.
- Grep website docs for `--plain`, `--no-input`, `--progress ndjson`.

Implementation:

- Update root/config/ETL/flow help text as applicable.
- Regenerate `docs/cli/**` for changed docs-map manuals.
- Update `website/content/docs/cli-reference.mdx`.

## Commit / push checkpoints

1. Plan artifact checkpoint.
2. Red-test checkpoint after failing tests are captured.
3. Green implementation checkpoint after focused gates pass.
4. Docs/website parity checkpoint after help/docs tests pass.
5. Final verification checkpoint after full local gates; push to `origin feat/405-tty-ndjson-progress` only, never `main`.
