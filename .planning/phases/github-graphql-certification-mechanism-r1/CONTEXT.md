## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification (P2 GraphQL mechanism).
- Base branch: `integration/4015-mvp-flat-r1` at `e7ae907ec6962920ebf42dc52c27c6014de6031a`.
- Merges into: `integration/4015-mvp-flat-r1` → human-gated integration → `main`.
- Delivery: A pull request from `fm/cli-graphql-certification-mechanism-r1` is open against the stated base, its API-reported base matches exactly, and local required checks are recorded.
- Working branch: `fm/cli-graphql-certification-mechanism-r1`.
- Task: Add a bounded, connector-neutral GraphQL certification stage. GraphQL schema, operation assertions, and classification data remain owned by connector definitions. Establish an honest, mutually exclusive 305-command split of schema-conformant, live-required, and fixture-bound commands; unexecuted commands must have a concrete non-pass status.
- Verification: Red/green tests prove a compiled schema and then a broken assertion fail; test the shared certification and consuming `cmd/connectorgen` packages; run validation, generated-file, boundary, lint, docs, and website-byte-stability gates; exercise the approved disposable GitHub credential only through the product path where a bounded read-only live probe is needed.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Connector-neutral schema certification | fake | Deterministic unit tests compile connector-owned schema data and validate declared GraphQL operations; fake provider transport is necessary to make CI deterministic and does not claim a live provider result. |
| Failure is observable after compilation | fake | A deliberately broken asserted response path is evaluated against a compiled, valid schema and makes the stage fail; without assertion evaluation it would pass. |
| 305-command disposition is honest | fake | A bundle-level test counts mutually exclusive dispositions and requires the total be 305; records have no `pass` status unless a live run produces an observable value. |
| Two bounded live read probes | live | The built `pm` binary executes two connector-owned read-only queries through the disposable identity and asserts declared produced values independently of exit status. |

## Discussion outcome

- Shared certification is justified only as an input-driven stage usable by any connector definition; no connector identifier is permitted in shared Go.
- Schema conformance establishes that a declared document is valid for the declared provider schema. It cannot establish a provider-produced value, authorization, fixture availability, or mutation effect; those remain live-required or fixture-bound.
- The task is a single GitHub connector lane with a permitted generic certification-stage addition. The branch is correctly based on the required integration ref.
- GSD was executed inline because this direct-PR task and the repository contract prohibit role spawning. Required skills loaded: `golang-how-to`, `golang-graphql`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
