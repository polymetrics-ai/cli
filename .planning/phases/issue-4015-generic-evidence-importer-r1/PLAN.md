# Generic Certification Evidence Importer

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification publication
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: PR #4216 is rebased onto verified integration head `c9791db4d`; it contains the generic importer and schema-v2 observed-operation proof path. Live GitHub publication is deferred only until credential rotation supplies a safe fresh proof artifact.
- Working branch: `fm/cli-generic-evidence-importer-r1`
- Task: Make `connectorgen certification-evidence` import completed proof-bearing reports for any connector without shared-code connector-name branches; stage GitHub's declaration-owned read binding and generic stage selector for the 34 previously proven live reads, without fabricating the missing fresh proof artifact; preserve PostgreSQL evidence bytes; and prove redaction, missing-evidence, and broken-evidence negative controls.
- Verification: Focused Go red/green tests in `./cmd/connectorgen`, `go run ./cmd/connectorgen certification-matrix`, generated-file and boundary checks, a second definition-shaped importer invocation, GitHub matrix count inspection, `gofmt`, `go vet`, `make` gates individually, and an inline security review.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Generic import maps completed observed proof into accepted evidence | fake | Deterministic report fixtures are required so CI never accesses live credentials; tests assert the serialized record passes `validateAcceptedEvidence` and matches the matrix capability key. |
| Captured secrets and response bodies cannot escape evidence | fake | Planted values in a request header, URL query, and response body must not occur in emitted JSON; fixtures are essential to prove a known secret is absent without committing a real one. |
| GitHub read publication is staged but honest | live deferred | The v2 importer can publish an `observed_operations` record from a fresh direct-read proof. The 34 previously proven reads have no recoverable proof artifact, and a replacement run waits for credential rotation; no record is fabricated. |
| Existing PostgreSQL proof remains byte-stable | live | A byte comparison confirms all fourteen PostgreSQL records remain unchanged from `c9791db4d`, whose v2 schema records their actual credential scope. |
| A third connector needs no new Go branch | fake | A distinct connector-shaped fixture and definition-owned source binding are imported through the same generic command; a deterministic fixture is used because its report is not a claim of live provider execution. |
| Missing or invalid evidence does not produce `live_tested` | fake | Matrix tests and a command-level broken-record control show the named capability changes red, then returns green after restoration. |

## Locked Decisions and Scope

- The verified diagnosis in `data/production-parity-shared-context.md` is authoritative: HTTP capture and proof-bearing record construction already exist; only the API-report-to-evidence importer is missing.
- Do not depend on PR 4198. Do not add connector identifiers or allowlist changes in shared Go. Connector-owned information belongs in definition files.
- Preserve the 12 pre-existing PostgreSQL transport records byte-identically (and do not rewrite the two later CDC records) unless an unavoidable format change is explained in the final diff.
- Publish only GitHub operations recorded as completed with observed proof. On refreshed integration base `c9791db4d`, the candidate projection is regenerated and byte-stable. The known 34 proven-live reads remain unpublished until a rotated credential can produce a new proof artifact. This slice does not claim to close the 1,571-command gate.
- Never print, serialize, or commit a credential, an authorization header, raw URL token, or raw provider response body.
- No new dependency, no ambient GitHub CLI login, and no write outside the disposable fixture scope.
- Schema v2 from #4215 resolves the former asserted-versus-verified credential-scope contract: a direct-read proof may state only `observed_operations`. The remaining publication prerequisite is a fresh report plus matching redacted proof, not a scope assertion.

## GSD and Skill Record

- Resolved GSD command sources: `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` passed.
- Inline/manual fallback: `gsd-sdk query init.phase-op issue-4015-generic-evidence-importer-r1` reports `phase_found: false`; the issue is a direct-PR delivery slice rather than a roadmap phase. The canonical adapter therefore cannot create phase artifacts or isolated workers. This worker executes the same discuss → TDD plan → execute → verify → review lifecycle inline and records it here.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-database`, `golang-graphql`, `golang-documentation`, and `golang-lint`.

## Execution Plan

1. Verify the cited source contracts and inspect existing accepted-evidence validation, report schema, database importer, and definition source bindings.
2. Add failing importer and negative-control tests, including secret redaction and matrix-accounting controls; record the red state in `TDD-LEDGER.md`.
3. Replace database-specific command plumbing with definition-derived generic input and run the records through the existing proof builder/accepted-evidence validator.
4. Exercise the generic report importer with validated proof fixtures, prove PostgreSQL byte stability plus a second connector-shaped import, and stage GitHub's observed-operation binding. Defer its live proof only while the credential rotation prevents a safe new run.
5. Execute focused and repository gates, perform the deliberate broken-evidence demonstration, run verification and code review inline, commit/push, open the direct PR, and read its API base back.
