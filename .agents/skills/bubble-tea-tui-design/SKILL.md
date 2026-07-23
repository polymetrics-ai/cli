---
name: bubble-tea-tui-design
description: Design, implement, test, or review Polymetrics terminal interfaces built with Bubble Tea v2, Bubbles v2, Lip Gloss v2, Huh, Glamour, or terminal charts. Use for any TUI model, dashboard, wizard, browser, query grid, pager, keymap, focus system, responsive terminal layout, accessibility pass, or interactive CLI issue/plan/prompt.
---

# Bubble Tea TUI Design

Use this skill to keep every Polymetrics terminal surface consistent, safe, keyboard-first,
accessible, and compatible with its plain/JSON sibling.

For work in CLI Architecture v2 issue #397 / PR #438, also load
[`cli-architecture-v2-delivery`](../cli-architecture-v2-delivery/SKILL.md) for parent-branch, GSD,
dependency, integration, review, and human-gate rules. This skill remains the source of truth for
TUI interaction and rendering design.

## Start here

1. Read `docs/design/tui-ux-design.md` and
   `docs/design/terminal-ui-research-and-design-system.md`.
2. Read the issue's GSD plan and the CLI help/docs/website parity instructions.
3. Load `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, and the other
   Go skills routed by the task.
4. Before production edits, activate `gsd-programming-loop` and record the RED contract.
5. Read the reference files selected by the task:
   - layout, focus, keymaps, or models: `references/interaction-and-layout.md`
   - charts or dashboards: `references/charts-and-dashboards.md`
   - test, accessibility, or review: `references/testing-and-accessibility.md`
   - visual inspiration or dependency choice: `references/inspiration-study.md`

## Non-negotiable contract

- Flags, plain text, JSON, and NDJSON are the API; the TUI is a TTY-gated projection.
  Bubble Tea and Huh prompt activation require both stdin and stdout TTYs, plus no `--json`,
  `--plain`, `--no-input`, `PM_NO_TUI`, `CI`, or `TERM=dumb`. `--plain`, `--json`, and
  `--no-input` always bypass Bubble Tea, Huh, and all prompts, producing deterministic output or
  exact required-flag errors only. Sequential prompts are allowed only in explicit accessible mode
  after the same stdin+stdout TTY gate passes and no bypass flag is set. Piped or non-TTY stdin
  always falls back to deterministic plain/noninteractive behavior; never consume scripted stdin
  unexpectedly, never hang for a prompt, and never open `/dev/tty` to bypass the gate. Ordinary
  bare namespaces render contextual help/subcommand summaries and exit 0. The human-first
  allowlist is `pm query` and `pm reverse`: on an eligible dual-TTY they open the same safe
  workspace as explicit aliases `pm query grid` and `pm reverse guide`; on every bypass path they
  render contextual help and exit 0. Help flags never launch a TUI, and invalid actions remain
  usage errors.
  Action subcommands may progressively prompt for missing fields after the same gate passes; for
  example, incomplete `pm credentials add [name]` and `pm connections create [name]` invocations
  may start guidance, while complete invocations execute directly. Invalid supplied values return
  direct validation errors instead of opening a repair wizard.
- Treat `--json --no-input` as the documented agent/automation invocation profile and add
  `--progress ndjson` only to long-running commands. Do not invent a global `--agent-mode`:
  `pm query run --agent-mode summary|stream` controls query result shape only. Agent paths must
  return one structured envelope or an actionable required-flag error and must never prompt.
- Use Bubble Tea's model/update/view architecture. Commands own asynchronous I/O; `Update`
  remains deterministic and never performs blocking work.
- Default to Normal mode. Enter Filter/Edit mode only when an input owns focus, display the
  mode, and let `esc` return to Normal before it backs out of the surface.
- Support arrows and Vim navigation: `j/k`, `h/l` where spatially meaningful, `gg/G`,
  `ctrl+u/ctrl+d`, `/`, `n/N`, `tab/shift+tab`, `enter`, `esc`, `?`, `q`, and `ctrl+c`.
- Keep short contextual help visible. `?` opens the complete keymap for the current mode.
- Sanitize and redact every dynamic view string before styling. Never render credentials,
  approval tokens, headers, request bodies, query strings, or raw secret-bearing errors.
  Approval tokens are sensitive one-time authorization values; a guided reverse flow may carry
  them only ephemerally in memory and must never print them in final frames, transcripts, logs,
  screenshots, accessibility output, JSON, shell-equivalent command text, or fixtures.
- Do not add generic shell, generic HTTP write, generic SQL write, or generic file-write actions.
  Query charts operate only on the existing read-only result path. Query export is a typed
  read-only output path with project confinement, no-overwrite default, confirmation/`--force`,
  sanitized command echo, and `--no-input` guidance.
- Mutations require plan/preview/approval/execute. Dangerous actions cannot be unlabelled
  single-key shortcuts.
- Color is reinforcement only. Pair it with a word and glyph; support `NO_COLOR`, ASCII,
  reduced motion, screen-reader-safe transcripts, and the plain fallback.
- Resize from `tea.WindowSizeMsg`; never hardcode a single terminal size or assume a dark
  background. Below the supported size, render a useful size guard.
- New Go dependencies are a human gate. Do not edit `go.mod` merely because a reference
  implementation uses a library.

## Component and state architecture

Keep a small root model that owns:

- current surface and mode;
- pane focus and responsive layout class;
- shared semantic key bindings and help model;
- sanitized view data and bounded result windows;
- cancellation/error/final state;
- child Bubbles models that receive only relevant messages.

Commands and application services emit typed events or return bounded data. Models must
not import business packages to discover data or perform writes. The command layer builds
small view-specific data structures and chooses plain versus TUI.

## Design workflow

1. Write the plain/JSON parity contract and failure/cancellation states.
2. Sketch 160×45, 100×30, and 80×24 frames plus the below-minimum guard.
3. Define modes, focus order, and key conflict rules before adding styling.
4. Define semantic content hierarchy: title/status, primary work area, contextual detail,
   help/status footer.
5. Implement the smallest model with keyboard navigation and plain fallback.
6. Add styling, charts, mouse support, and motion only after keyboard and transcript tests
   are green.
7. Run the accessibility, redaction, resize, cancellation, and parity matrix.
8. Record help/manual/website checks and exact GSD/TDD evidence in the phase artifacts.

## Review failure conditions

Block or correct a TUI change when it:

- steals printable keys while an editor/filter is focused;
- hides current focus, mode, units, selected item, or destructive impact;
- relies on color, animation, mouse input, Unicode width, or alt-screen state alone;
- performs I/O in `Update`/`View`, leaks goroutines, ignores cancellation, or floods events;
- renders unbounded query/result data or misleading charts without axes/units/text values;
- changes plain/JSON behavior, exit codes, stdout/stderr boundaries, or help parity;
- makes humans discover `query grid` or `reverse guide` before reaching the default TTY workspace,
  removes either explicit alias, or launches a bare workspace on a bypass path;
- launches guidance for a complete action, repairs an invalid supplied value by prompting, or lets
  an agent/non-TTY invocation prompt or consume unexpected stdin;
- adds a dependency not named and approved for the phase.
