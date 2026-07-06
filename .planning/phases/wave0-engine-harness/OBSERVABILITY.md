# OBSERVABILITY — wave0-engine-harness

Posture: this is a local-first CLI; wave0 adds **no new runtime telemetry** (no metrics endpoints,
no tracing, no log sinks). The observability surface of this phase is (a) structured error
context and (b) machine-readable report artifacts consumed by agents and the wave gates.

## 1. Engine error context (the debugging surface)

Every engine failure is a typed `engine.Error` (API-CONTRACT §2) carrying:

- `connector` — bundle name
- `stream` or `action` — which declarative unit failed
- `page` (read path) or `record_index` (write/validate path)
- `class` — from `error_map` (e.g. `rate_limited`) when a rule matched
- `hint` — bundle-authored operator hint, surfaced VERBATIM by the CLI (design §F.4)
- wrapped cause chain reaching `*connsdk.HTTPError` (`status`, redacted URL/body excerpt)

Guarantees (tested, T-04/T-08): messages pass `safety.RedactErrorText`; secrets never appear;
`errors.As`/`errors.Is` reach the HTTP cause. This replaces per-connector ad-hoc error prose and
is the primary signal migration agents get when a bundle misbehaves.

## 2. Report artifacts (machine-readable observability)

| Artifact | Producer | Consumer |
|---|---|---|
| Conformance v2 `Report` (per bundle: static+dynamic checks, JSON-marshalable) | `conformance.Run` / `TestConformance` output | wave gates, repair agents, adversarial reviewers |
| `connectorgen validate --json` findings (file+rule per defect) | `cmd/connectorgen` | agent self-check loop (orchestration-plan §Verification pyramid layer 1) |
| Certification report `.polymetrics/certifications/<name>.json` + `history/` | `certify.Runner` | maintainers; enablement generator in a later phase |
| `docs/migration/inventory.json` | `cmd/inventorygen` | wave1+ planners (bundle sizing/assignment) |
| `.planning/phases/wave0-engine-harness/{TDD-LEDGER.md,VERIFICATION.md,traces/*}` | executors/verifier | coordinator, reviewer |

All JSON artifacts carry enough identity (`connector`, check `name`, timestamps where relevant)
to be diffed run-over-run; certify reports additionally record per-stage
`{argv_redacted, exit_code, kind, duration_ms}` so a failing stage is reproducible by hand.

## 3. Gate visibility

Local gate summary comes from `make verify` (extended with `lint` + `connectorgen-validate`,
RUNBOOK §1) and `go test -cover ./internal/connectors/engine` for the coverage metric. There is
no CI yet (per `.planning/STATE.md`); the wave-close procedure in RUNBOOK §5 is the authoritative
record, committed with each wave.

## 4. Deferred (explicitly not wave0)

- Live certification budgets/rate-limit accounting (`budget` block) — with certify Tier-2 work.
- Leak ledger observability (`certify-ledger.jsonl`) — with the write-protocol phase.
- Any `pm` runtime logging changes, CI dashboards, or coverage trend tracking.
