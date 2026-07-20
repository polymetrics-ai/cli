# TDD Ledger — Phase 408 flow/ETL dashboards

Issue: #408  
Mode: manual universal-loop fallback after `scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`.

## Loaded skills

- `gsd-core`
- `bubble-tea-tui-design` + references: interaction/layout, charts/dashboards, testing/accessibility, inspiration study
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-spf13-cobra`
- `caveman` for final handoff only

## RED plan

Before production edits, capture failing tests/validation for:

1. `internal/ui` dashboard model frames:
   - success final frame
   - failure final frame with redacted/sanitized error
   - cancellation final frame after Done
   - wide/standard/compact/guard layouts
   - no-color/ASCII/reduced-motion/accessibility frames
2. Event channel bridge:
   - progress throttling/coalescing
   - lifecycle events not dropped
   - channel close sends final Done/error state
3. Cancellation:
   - `ctrl+c` cancels runner context
   - model waits for final event before quitting
4. CLI activation matrix for `flow run` and `etl run`:
   - eligible dual-TTY activates dashboard
   - `--plain`, `--json`, `--no-input`, `CI=1`, `PM_NO_TUI=1`, `TERM=dumb`, stdin-piped, stdout-piped bypass dashboard/prompt paths
   - no ANSI on machine paths
   - stdout/stderr/exit parity for plain existing behavior
5. Help/docs parity:
   - `pm help flow`, `pm help etl`, bare `pm flow`, bare `pm etl`, and command help reflect behavior or remain unchanged when not applicable

## Evidence log

| Cycle | Type | Command | Result | Notes |
|---|---|---|---|---|
| plan | GSD preflight | `scripts/gsd doctor` | PASS | Adapter healthy. |
| plan | GSD preflight | `scripts/gsd list` | PASS | 69 commands listed. |
| plan | GSD plan prompt | `scripts/gsd prompt plan-phase 408 --skip-research` | PASS | Wrote `/tmp/gsd-plan-408.txt`. |
| plan | GSD programming loop | `scripts/gsd prompt programming-loop init --phase 408-flow-etl-dashboards --dry-run` | FAIL | `scripts/gsd: unknown GSD command: programming-loop`; manual universal-loop fallback recorded. |
| plan | parent sync | `git fetch origin feat/cli-architecture-v2 && git merge --ff-only origin/feat/cli-architecture-v2` | PASS | Branch fast-forwarded to `b77d8f49` before production edits. |

## RED evidence

Pending. Next action: inspect current `internal/ui`, `internal/events`, `internal/cli/flow_cli.go`, and `internal/cli/etl_cli.go`, then add focused failing tests before production edits.

## GREEN evidence

Pending.

## REFACTOR evidence

Pending.
