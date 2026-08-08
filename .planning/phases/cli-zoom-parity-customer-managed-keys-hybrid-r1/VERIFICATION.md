# Verification Checklist — Zoom Customer Managed Keys Hybrid parity, R1

## Lifecycle

- [x] Required skills and connector/CLI/GSD guidance are loaded for the parent run and recorded in `PLAN.md`.
- [x] GSD command provenance is resolved with `scripts/gsd sources`.
- [x] Live API/auth artifact URL, retrieval time, HTTP status, byte count, and exact operation audit are recorded before RED.
- [ ] RED test failure is captured, committed, and pushed before production declarations.
- [ ] Separate direct-write safety foundation is red/green tested, committed, and pushed.
- [ ] GREEN implementation is committed and pushed.
- [ ] Inline verify-work and code-review evidence are recorded.

## Source parity

- [ ] The one Customer Managed Keys Hybrid operation matches the live artifact and derives exactly one covered Zoom ledger row.
- [ ] No `unsafe_or_disallowed` disposition is introduced.
- [ ] The command exposes exactly the two documented required request-body fields and no pagination input.
- [ ] Key-connector JWT configuration is explicit and cannot send the standard OAuth token to a custom customer host.
- [ ] Successful response and provider error output redact generic and declared sensitive values.

## Runtime/docs checks

- [ ] Focused engine, commandrunner, app, Zoom, conformance, vet, and build checks pass.
- [ ] The built binary reaches all namespace/help routes and the declared plan lifecycle without local unknown command/flag failures.
- [ ] The isolated fixture invocation proves declared POST/auth/path wiring with no emitted synthetic secret.
- [ ] `surface-sync --check`, Zoom/full validation, docs/website validation, and scoped generated-file review pass.
