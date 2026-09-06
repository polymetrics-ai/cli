# ADR 0002 — Cobra routing and invocation-scoped Viper configuration

- Status: Accepted (2026-07-24); steady-state wording updated 2026-09-01
- Foundation issues: #399, #400, #401, #402

## Context

The `pm` CLI needs one re-entrant command tree, stable machine-readable output,
typed configuration precedence, dynamic connector dispatch, and an approval
boundary that behaves identically from terminals, CI, cron, and agents.

## Decision

1. Build a fresh `spf13/cobra` command tree for every `cli.Run` invocation.
2. Keep one error funnel responsible for JSON envelopes, stderr, and exit codes.
3. Generate help and documentation from the current command model.
4. Dispatch connector operations from their rendered execution bundles.
5. Use `viper.New()` per invocation with an explicit environment allowlist.
6. Apply precedence as changed flag, explicit `POLYMETRICS_*` variable,
   matching `PM_*` alias, effective-root config file, then built-in default.
7. Keep credentials in the credential subsystem rather than typed config.
8. Require golden transcript review for intentional CLI behavior changes.

No compatibility dispatcher or fallback parser participates in current command
resolution. Connector commands resolve through the same execution artifact and
the same approval, authentication, and transport boundaries as other runtime
entry points.

## Consequences

- Command ownership and configuration precedence have one testable path.
- A fresh Cobra tree and Viper instance preserve in-process isolation.
- Machine output remains stable across interactive and automated callers.
- Golden transcripts make intentional interface changes reviewable.
