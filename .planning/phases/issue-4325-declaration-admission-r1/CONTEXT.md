# Context — issue 4325 declaration-admission foundation

## Task Delivery Header

- Issue: Refs #4325 — restore Batch 1’s independent gate; this PR supplies only its shared source-declaration admission foundation.
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, containing a committed shared admission gate and its focused tests with local verification recorded.
- Working branch: fm/cli-declaration-admission-certification-r1
- Task: Add a deterministic, provider-I/O-free `connectorgen declaration-admission` certificate. It admits source rows only when each has a cited provider identity, a single lane, one discoverable CLI command, destructive metadata where applicable, and an implemented binding or named deferred foundation gap. Existing runtime/preflight, source lock, surface-sync, live certification, and connector-owned definitions remain separate.
- Verification: focused red/green tests in `cmd/connectorgen` and `internal/connectors/commandrunner`; `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/commandrunner`; generated/schema checks; selected `make` gates; built CLI discovery/refusal checks if a synthetic fixture can exercise them; final `git diff --check` and review.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Cited runnable reads are admitted | live | A fixture row linked to an implemented direct-read command is accepted; changing its lane, citation, endpoint, or binding produces a named admission finding. |
| Deferred writes, deletes, and binary rows remain visible | live | Deferred reverse-ETL, delete, download, and upload rows each pass only with a named foundation gap and discoverable command; an invocation is refused before provider I/O. |
| Missing, duplicate, stale, and base-path-mismatched declarations fail | live | Table-driven fixtures change each defect independently and assert the deterministic diagnostic. |
| No retained bytes, hashes, or typed body are admission requirements | live | A complete fixture with only a URL, exact document citation, identity, lane, command, and deferred state passes without those fields. |
| Admission remains distinct from runtime and live certification | live | Tests accept a zero-runnable all-deferred connector and assert that an `implemented` row without a runtime binding fails; existing runtime preflight and certification commands are not altered. |

## Discussion record

- Captain policy treats every provider operation as a declaration/discovery obligation, including writes, deletes, and binary operations. A source citation is sufficient for admission; byte retention and live proof are distinct certificates.
- Firstmate clarification: a deferred operation still owns a deterministic `cli_surface.json` command identity/path. A missing foundation is metadata, never an omitted operation.
- The existing source-import descriptor is intentionally artifact/hash-bound and runtime projection is stricter. Admission therefore receives a required, versioned repository source cohort plus a separate declaration catalog rather than weakening either existing contract; neither catalog requires retained bytes or a hash.
- Inline/manual GSD execution is used because this runner cannot use compatible isolated GSD roles and repository policy forbids role spawning.

## Skills and lifecycle

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.

Resolved command path: `scripts/gsd doctor`, `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`, plus each generated prompt. The task’s explicit contract supplies the discussed decisions; the lifecycle is performed inline and recorded here.
