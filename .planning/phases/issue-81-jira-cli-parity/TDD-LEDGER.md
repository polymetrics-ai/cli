# TDD Ledger: Jira CLI Parity Parent

## GSD Setup

- `scripts/gsd doctor`: passed.
- `scripts/gsd verify-pi`: passed.
- `scripts/gsd list --json`: ran; output exceeded harness display limit.
- `scripts/gsd prompt plan-phase issue-81-jira-cli-parity --skip-research`: generated prompt successfully.
- `scripts/gsd prompt programming-loop init --phase issue-81-jira-cli-parity --dry-run`: failed, `unknown GSD command: programming-loop`.
- Manual fallback active per `.agents/agentic-delivery/references/gsd-pi-adapter.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

## Red / Green Plan

Parent orchestration is mostly planning state. Behavior-changing work begins in issue #104.

### Issue #104 planned red test

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
```

Expected first failure: Jira bundle loads but `b.CLISurface == nil` because no `cli_surface.json` exists.

### Issue #104 green target

- Add Jira `cli_surface.json`.
- Re-run the focused engine test until it passes.
- Re-run CLI-surface validator and full connector definition validation.

## Evidence Log

| Time (UTC) | Cycle | Evidence |
| --- | --- | --- |
| 2026-07-09T12:49:58Z | plan | Parent issue #81 and sub-issues #104-#110 loaded with `gh issue view --json`. |
| 2026-07-09T12:49:58Z | plan | `go run ./cmd/pm help connectors` ran before connector inspection. |
| 2026-07-09T12:49:58Z | plan | `go run ./cmd/pm connectors inspect jira --json` confirmed metadata-only baseline: read=true, write=false, streams issues/projects/users. |
