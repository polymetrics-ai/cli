# Code review — PostgreSQL CDC restart recovery for 0.2.1

## Method and verdict

Inline manual review is the repository-mandated fallback because the canonical single-worker contract forbids spawning a reviewer role. The full diff from `origin/integration/4015-mvp-flat-r1` was reviewed after focused, live, and broader verification.

**Verdict:** pass — no actionable findings.

## Reviewed concerns

- The change does not make polling resumable or accept both checkpoint families. A bootstrap/CDC request accepts only the existing `logical_replication` mechanism, schema, protocol, and sealed PostgreSQL bootstrap barrier; a valid polling checkpoint still produces typed `invalid_checkpoint` rebootstrap.
- The transport wrapper now classifies the request before applying mechanism-specific validation. The same `postgresBootstrapTransportRequested` predicate controls both pre-I/O validation and dispatch, preventing the two decisions from drifting.
- The application-owned resume identity is still checked before the sealed native source identity is restored. `ReadCDC` then performs the existing system ID, timeline, publication, relation, schema fingerprint, LSN-retention, and live-source validation; no source-generation guard was bypassed.
- The extracted protocol helper contains only the checkpoint-family invariant formerly inside `validateCDCResume`; it does not replace the live source/generation checks.
- Receipt-before-checkpoint-before-source-ack ordering is unchanged. The live capability test and package failure-injection coverage remain green.
- The integration proof independently queries the managed target. Exact CDC counts are `1` before interruption, `1` at interruption, and `2` after restart; the post-restart key occurs once and the 1,001-row control table remains unchanged.
- PostgreSQL CDC remains declared `at_least_once`. The exact target multiplicity is evidence for the keyed managed-target route, not a new generic exactly-once claim.
- The immutable `postgres_cdc_r1-capability-cdc.json` record remains truthful for its four explicit source-CDC-to-warehouse facts and passed again with the fixed binary, so it was not rewritten.
- No command/help/docs/website surface, dependency, credential flow, release branch, or shared runtime lifecycle changed.

## Evidence reviewed

- Production: `internal/connectors/native/postgres/cdc.go` and `polling_transport_source.go`.
- Regression: `internal/connectors/native/postgres/transport_source_test.go`.
- Process-death and independent target proof: `internal/cli/postgres_transport_binary_integration_test.go`.
- Red/green traces and verification matrix in this phase directory.
- Required skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-database`, and `golang-lint`.

External pull-request review route: `claude_auto` on the non-draft direct PR. Its head SHA and disposition record belong on the PR so the review covers the final committed diff.
