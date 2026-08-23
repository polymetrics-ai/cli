# Source-lock operation import — context

## Task Delivery Header

- Issue: Refs #4306 — feat(connectorgen): import hash-locked provider operation contracts.
- Base branch: `main`.
- Merges into: `main`.
- Delivery: Branch `fm/cli-source-lock-import-r1` is committed locally with source-import implementation, its GSD/TDD evidence, and verified local gates; Firstmate owns the later no-mistakes pipeline and PR.
- Working branch: `fm/cli-source-lock-import-r1`.
- Task: Build the closed, connector-agnostic F0 importer that verifies a connector-owned lock before reading a public OpenAPI/source artifact and emits deterministic, declaration-ready operation descriptors. No production connector definitions, transport execution, credentials, or generic request controls are in scope.
- Verification: Focused red/green unit and command tests; `go test -timeout 20m ./cmd/connectorgen`; generator validation, golden/docs checks, `go vet ./...`, `go build ./cmd/pm`, `git diff --check`, completion-tracked `make connector-boundary`, and `make verify`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Only the connector-owned lock chooses the remote artifact | live | Tests assert that the importer fetches the lock URL exactly and rejects a lock outside the connector's `sources/` directory before any fetch. |
| Artifact bytes and digest are authoritative | live | A byte or SHA-256 mismatch returns a source-lock-refresh error and emits no descriptor file. |
| Fixed provider contracts become descriptors | fake | Checked-in synthetic OpenAPI JSON/YAML fixtures are necessary because production artifacts must not be cached; tests assert their emitted descriptor fields exactly. |
| IDs and serialization are deterministic | fake | Synthetic two-connector fixtures isolate ordering and omitted provider IDs; tests compare output bytes from repeated imports. |
| Unsafe/unsupported forms fail before generation | fake | Synthetic fixtures exercise each malformed source form without using live provider documents, which would be mutable and outside lock verification. |
| Downstream adoption is documented | live | Help and migration-document checks assert the import command and closed-source contract are discoverable without claiming a `pm` user command changed. |

## Decisions

- The source lock is the sole public-artifact locator. The command takes a connector and definitions root, resolves only that connector's `sources/<connector>-operation-source-lock.json`, and exposes neither a URL nor arbitrary source request fields.
- The importer supports only pinned JSON OpenAPI 3 / Swagger 2 and YAML OpenAPI 3 source forms using the existing pinned YAML dependency. It resolves only local JSON Pointer references under `#/`; unsupported or ambiguous schemas fail closed.
- Descriptor output is a canonical JSON intermediate. It records provider-owned `operation_id` unchanged, including `""`; `source_id` is derived only from connector, lower-case method, and connector-relative path when that provider ID is empty.
- The descriptor is a provider-contract artifact, not a command declaration or an execution surface. No provider call follows import; all test fetches are local/in-memory fixture seams.
- Captain clarification (2026-08-20): descriptor response/output schemas preserve every ordinary provider-declared response field and each declared status shape. Classification must not delete fields because they are rare, privileged, paid-tier, destructive, unfamiliar, or sensitive-looking. If a downstream runtime masks a credential/secret, its representation must retain field presence with an explicit marker; the importer never silently deletes ordinary provider fields.

## GSD and skills record

Resolved lifecycle: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `scripts/gsd prompt discuss-phase 4306`; `scripts/gsd prompt plan-phase 4306 --tdd`; and `go run ./cmd/agentcontractgen check`.

Inline/manual fallback: the direct implementation task has no compatible isolated Pi-role runtime, and the canonical delivery contract forbids role spawning. Discussion, planning, TDD execution, verification, and review will be recorded in this phase directory.

Skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-cli`, `golang-documentation`, `golang-lint`, `golang-swagger`, and `no-mistakes` (doctor only; the user prohibits starting its pipeline in this lane).

CLI/docs parity: `connectorgen source-import --help` changes developer-generator help only. No `pm` command, `docs/cli/**`, website page, generated `pm` manual, namespace behavior, completion, or credentialed connector behavior changes; each is intentionally not applicable and will be recorded in verification.
