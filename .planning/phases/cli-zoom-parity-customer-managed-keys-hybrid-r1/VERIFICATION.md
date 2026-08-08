# Verification Checklist — Zoom Customer Managed Keys Hybrid parity, R1

## Lifecycle

- [x] Required skills and connector/CLI/GSD guidance are loaded for the parent run and recorded in `PLAN.md`.
- [x] GSD command provenance is resolved with `scripts/gsd sources`.
- [x] Live API/auth artifact URL, retrieval time, HTTP status, byte count, and exact operation audit are recorded before RED.
- [x] RED test failure is captured, committed, and pushed before production declarations (`5a0172053`).
- [x] Separate direct-write safety, operation-origin/auth, and endpoint-ledger foundations are red/green tested, committed, and pushed (`0987a58bc`, `833a2d9d4`, `410eb1bb7`).
- [x] GREEN implementation and scoped generated docs are committed and pushed (`cec675503`, `f86e6a480`).
- [x] Review-found inherited customer-host header leak is RED/GREEN committed and pushed (`5c9518918`, `dfa221bcd`).
- [x] Inline verify-work and code-review evidence are recorded.

## Source parity

- [x] The one Customer Managed Keys Hybrid operation matches the live artifact and derives exactly one covered Zoom ledger row.
- [x] No `unsafe_or_disallowed` disposition is introduced (`0` Zoom rows).
- [x] The command exposes exactly the two documented required request-body fields and no pagination input.
- [x] Key-connector JWT configuration is explicit and cannot send the standard OAuth token or inherited secret header to a custom customer host.
- [x] Successful response and provider error output redact generic and declared sensitive values.

## Runtime/docs checks

- [x] Focused engine, commandrunner, app, Zoom, conformance, certify, vet, lint, and build checks pass.
- [x] The built binary reaches all namespace/help routes and the declared plan lifecycle without local unknown command/flag failures.
- [x] The isolated fixture invocation proves declared POST/auth/path wiring with no emitted synthetic secret.
- [x] `surface-sync --check`, scoped/full validation, docs/website validation, and scoped generated-file review pass.
