# Verification Checklist — Zoom Quality Management parity, R1

## Lifecycle

- [x] Required skills and connector/CLI/GSD guidance are loaded for the parent run and recorded in `PLAN.md`.
- [x] GSD command provenance is resolved with `scripts/gsd sources`.
- [x] Live artifact URL, retrieval time, HTTP status, byte count, and exact six-operation audit are recorded before RED.
- [ ] RED test failure is captured, committed, and pushed before production declarations.
- [ ] GREEN implementation is committed and pushed.
- [ ] Inline verify-work and code-review evidence are recorded.

## Source parity

- [x] All six live Quality Management operations match the derived ledger (delta `0`).
- [x] No `unsafe_or_disallowed` disposition is permitted.
- [ ] Five GETs and one POST have exactly one covered declaration each.
- [ ] List reads have no source-invented paging/date input flags.
- [ ] POST carries every documented typed input and a closed nested request schema.

## Runtime/docs checks

- [ ] Focused Zoom, conformance, commandrunner, CLI, vet, and build checks pass.
- [ ] Every route is reachable through the built binary; safe reads reach Zoom as provider `401`, and POST remains preview-only.
- [ ] `surface-sync --check`, Zoom/full validation, docs validation, and scoped generated-file review pass.
