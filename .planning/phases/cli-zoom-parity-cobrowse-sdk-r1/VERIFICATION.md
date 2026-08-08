# Verification Checklist — Zoom Cobrowse SDK parity, R1

## Lifecycle

- [x] Required skills and connector/CLI/GSD guidance are loaded for the parent run and recorded in `PLAN.md`.
- [x] GSD command provenance is resolved with `scripts/gsd sources`.
- [x] Live artifact URL, retrieval time, HTTP status, byte count, and exact four-operation audit are recorded before RED.
- [ ] RED test failure is captured, committed, and pushed before production declarations.
- [ ] GREEN implementation is committed and pushed.
- [ ] Inline verify-work and code-review evidence are recorded.

## Source parity

- [x] All four live Cobrowse SDK operations match the derived ledger (delta `0`).
- [x] No `unsafe_or_disallowed` disposition is permitted.
- [ ] Four GETs have exactly one covered declaration each.
- [ ] `from`/`to` appear only on the explicitly documented report reads.
- [ ] No page, page-size, token, or undeclared date query flag is exposed.

## Runtime/docs checks

- [ ] Focused Zoom, conformance, commandrunner, vet, and build checks pass.
- [ ] Every route is reachable through the built binary; safe reads reach Zoom as provider `401`.
- [ ] `surface-sync --check`, Zoom/full validation, docs/website validation, and scoped generated-file review pass.
