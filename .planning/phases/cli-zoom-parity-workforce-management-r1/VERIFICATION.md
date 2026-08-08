# Verification Checklist — Zoom Workforce Management parity, R1

## Planned checks

- [x] Live artifact, operation count, byte count, hash, and ledger comparison recorded before RED.
- [x] RED capture committed before all production declaration/foundation changes.
- [x] CSV foundation proves valid CSV reaches the provider while malformed or non-`.csv` sources
  are rejected before network dispatch; existing JSON validation remains green.
- [x] All 18 command paths pass real commandrunner preflight.
- [x] Eleven direct reads and seven direct writes run against isolated exact fixtures.
- [x] Both DELETEs assert 204 status-only semantics and require destructive confirmation.
- [x] Endpoint ledger reconciliation is confined to `provider_module=workforce-management`; zero
  rows are `unsafe_or_disallowed`.
- [x] Generated CLI docs/site output retains Zoom-only changes after whole-file generation.
- [x] Fresh `pm` binary reaches base, namespace, provider group, and all 18 command help routes.
- [x] Scoped local gates, inline verify-work, and manual code review complete.

## Captured results

- `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...` passed (`15.171s`), including
  real-runner tests for all 18 provider paths and exact plan/preview/approval/execute fixtures.
- `go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/connsdk
  ./internal/connectors/commandrunner ./cmd/connectorgen` passed; the numeric-range foundation,
  CSV foundation, command runner, and generator are jointly green.
- `go test -count=1 -timeout 20m ./internal/cli` passed. Scoped `go vet` passed for engine,
  connsdk, commandrunner, Zoom definitions, connectorgen, and `pm`.
- `make lint`, `make tidy-check`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make connector-boundary`, and `make release-workflow-check` passed. The smoke workflow created
  a temporary isolated project successfully.
- A fresh binary returned zero for `pm help zoom`, bare `pm zoom`, bare
  `pm zoom workforce-management`, and every exact Workforce Management `--help` route.
- `surface-reconcile` covered exactly 18 Workforce Management endpoints (11 reads / 7 writes),
  with no residual local block and zero `unsafe_or_disallowed` Zoom rows. The generated endpoint
  ledger changed only `.zoom` and contains the eleven direct-read paths.
- Connector manuals were regenerated and validated. Website generator output changed only the
  Zoom records; both generated website JSON files were compared after removing Zoom and were
  byte-equivalent to `HEAD`.
- Inline manual `verify-work` and `code-review` checked the closed request schemas, source 1–4
  numeric range, CSV caps/header/snapshot validation, redaction, typed confirmation, no invented
  paging flags, generated-file scope, and endpoint-ledger locality. No findings remain.
