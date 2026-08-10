# Verification Checklist — Zoom AI Services documented-operation parity, R1

## Planned checks

- [x] Fresh artifact URL, timestamp, HTTP status, byte count, SHA-256, OpenAPI/server provenance,
  and zero source-to-ledger delta recorded before RED. The Markdown audit is HTTP 200 / 87,750
  bytes / SHA-256 `154631ef…f3ba367`; the actual endpoint OpenAPI is HTTP 200 / 97,885 bytes /
  SHA-256 `d2ba7fbd…62ab65`. Both are recorded in `PLAN.md`.
- [x] RED command-surface test was captured and committed before production declaration; the
  WebSocket reconciliation-bootstrap RED was separately captured in `2b79911aa`.
- [x] All 22 provider operations are covered: 12 `rest_read`, 9 `rest_write`, and 1 fixed Live
  Scribe `websocket_session`; 22 matching ledger rows are covered and zero remain blocked.
  Zoom has zero `unsafe_or_disallowed` rows.
- [x] The protocol foundation has standalone schema RED/GREEN evidence, no new dependency, fixed
  endpoint/subprotocol/frame bounds, and a narrowly scoped reconciliation bootstrap fix in
  `29e4e64c1` (plus lint test cleanup `8518509c3`).
- [x] Loopback coverage executes every REST operation: 12 reads prove fixed paths/query/redaction;
  six JSON writes and three `DELETE` cancellations prove plan → preview → approval → execute,
  typed destructive confirmation, status 204, and no invented response body.
- [x] The Live Scribe foundation test suite proves fixed handshake/subprotocol/auth/session-update/
  PCM frame ordering, bounds, redaction, cancellation, and fail-closed rejection behavior; its
  closed operations schema loader regression is green.
- [x] A freshly built binary reached `pm help zoom`, bare `pm zoom`, bare `pm zoom ai-services`,
  and all 22 exact `pm zoom ai-services … --help` routes. The main-rebase exit behavior was used;
  no unknown command can pass this check at exit 0. The same 22/22 + 3/3 sweep was repeated after
  rebasing the consumer onto foundation `dee258307` with main `f96a47e80` as an ancestor.
- [x] `surface-sync --check`, scoped `surface-reconcile --check`, full connector validation,
  generated docs/manual/site output, Zoom-only endpoint-ledger delta, and website catalog locality
  are verified. Generated files were regenerated, not hand-merged.
- [x] Scoped tests, vet, build, tidy, docs, smoke, lint, agent contract, connector validation,
  surface sync, boundary, release, and website typecheck gates passed. Inline/manual `verify-work`
  and `code-review` evidence is recorded in this phase directory because the canonical runtime
  forbids role spawning for this parent delivery.

## Manual verify-work record — 2026-08-10

Automated evidence fully covers the deliverables in `SUMMARY.md`; no human product judgment or
credentialed provider call is required. The source page and endpoint JSON were re-fetched from
Zoom, all outputs use synthetic values in loopback tests, and the command surface is observed from
a freshly built binary. Result: **pass**.
