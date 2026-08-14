# #4089 — reverse-ETL approval stdin carrier context

## Locked decisions

- Keep the existing plan → preview → proceed authorization model, including single-use consumption, unchanged.
- Replace only the CLI transport channel: a reverse-ETL run requires the bare `--approval-token-stdin` marker and accepts the token as one bounded line on standard input.
- Reuse `readApprovalTokenFromStdin`; do not add a second secret reader, environment fallback, file carrier, or connector-specific path.
- Reject missing, valued, legacy argv, empty, oversized, multiline, and replay inputs before the first write-side effect.
- The implementation stays generic in `internal/cli`; no connector literal or special case is permitted.

## Code context

- `internal/cli/etl_transport.go` owns the existing 4 KiB, one-line stdin carrier.
- `internal/cli/cli.go` has the two reverse-ETL argv call sites: generic connector-command execution and `pm reverse run`.
- `internal/app.RunReverseETL` remains the authority for persisted-preview, authorization, and exact-once replay semantics.

## Canonical refs

- `AGENTS.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `internal/cli/etl_transport.go`
- `internal/cli/cli.go`
- `internal/app/app.go`

## Manual GSD fallback

The named issue phase is outside the numeric roadmap, and the repository's single-worker contract forbids the GSD roles that the generated workflows would otherwise start. The prompts were generated through `scripts/gsd prompt` and the lifecycle is executed inline with the locked issue decisions above.
