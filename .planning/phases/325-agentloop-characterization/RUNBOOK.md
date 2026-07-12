# Runbook and Rollback: Phase 0

## Operator inspection

1. Run `go run ./cmd/loopctl safety status --json`.
2. Confirm `state=closed`, `run_enabled=false`, and `resume_enabled=false`.
3. Run sanitized replay only against tracked fixture `.json` files.
4. Do not attempt run/resume; denial exit 78 is expected until dependent hardening phases land.

## Failure response

- If either driver creates `.planning/auto-loop` or launches a stub/model while closed, treat it as
  a hard safety regression and stop.
- If fixture validation finds raw or sensitive content, delete the synthetic fixture and recreate
  it from typed event semantics; never copy a source transcript.
- If a later phase needs live enablement, it must implement its brokered authorization issue; do
  not add a Phase 0 env or argument escape hatch.

## Rollback

Revert the child commit on the parent integration branch if the stacked change must be removed.
Never restore the old fail-open drivers as an operational workaround. Parent-to-main rollback or
merge remains a human decision.
