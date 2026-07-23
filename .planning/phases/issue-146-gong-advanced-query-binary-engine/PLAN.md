# Gong advanced query/binary policy

Parent: #133
Issue: #146
Branch: `feat/133-gong-cli-parity`

## GSD path

- Adapter health: `scripts/gsd doctor` passed in this session.
- Requested workflow: `scripts/gsd prompt programming-loop ...` remains unavailable (`unknown GSD command: programming-loop` from earlier run); manual-GSD fallback is active and recorded.
- Planning prompt available: `scripts/gsd prompt plan-phase 133 --skip-research`.
- Required skills loaded: gsd-core, golang-how-to, golang-cli, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-testing, golang-documentation.

## Objective

Model POST read-query bodies and binary/multipart payloads as typed operations with bounded policies and explicit engine gaps where not executable.

## Constraints

- No secrets requested, printed, stored, or summarized.
- No credentialed Gong checks.
- No generic raw HTTP write, arbitrary GraphQL mutation body, generic shell write, or direct-write escape hatch.
- Reverse ETL remains plan -> preview -> approval -> execute; destructive actions require typed confirmation.
- Sensitive/admin/destructive operations are implemented as typed surfaces where the engine supports the request shape; non-JSON body/query/binary gaps remain typed operation metadata with bounded policies and exact blocker notes.

## Implementation plan

1. Add fail-first tests that assert the Gong surface has executable command coverage and safety policy coverage.
2. Extend shared direct-read output policy only as needed for bounded JSON redaction.
3. Generate/update Gong definitions from the public OpenAPI operation inventory.
4. Refresh connector docs/manual notes and planning state.
5. Run targeted validation and Go tests before broader gates.

## Verification target

- `go test ./internal/connectors/engine -run DirectRead -count=1`
- `go test ./cmd/connectorgen -run Gong -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`

## 2026-07-10 engine-support replan (analysis only)

User request: deeply analyze and plan safe engine support for Gong's remaining executor gaps:
POST read-query execution with fixed JSON bodies, multipart uploads, and top-level array request bodies.
No production code changes are in scope for this planning slice.

### Current code evidence

- `internal/connectors/engine/direct_read.go` only permits `GET`, rejects absolute URLs, enforces api_surface direct-read coverage, caps successful responses to 1 MiB, decodes JSON, and applies a fixed output policy (`json_redacted` now exists).
- `internal/connectors/commandrunner/runner.go` blocks every `cli_surface.command.operation` before dispatch; executable reads must be stream-backed or GET direct reads, and executable reverse ETL must reference `writes.json` actions.
- `internal/connectors/engine/write.go` supports JSON object, form, none, and GraphQL bodies only. It cannot emit multipart bodies or top-level JSON arrays from a write action.
- `internal/connectors/engine/bundle.go` includes `StreamSpec.Body`, but `read.go`'s `buildStreamRequestBody` currently returns a body only for GraphQL streams; non-GraphQL fixed POST bodies are effectively dormant.
- `internal/connectors/connsdk/http.go` can marshal arbitrary JSON with `Requester.Do`/`DoLimited`, so top-level arrays are technically possible at the HTTP layer, but no typed write/direct-read planner exposes them. There is no multipart request primitive.
- `internal/connectors/engine/bundle.go` and `operations.schema.json` already load `rest.content_type`, `rest.max_bytes`, and `rest.body_schema`; `validateOperationSemantics` permits POST `rest_read` metadata only when `body_schema` is present. These are metadata-only today.
- Gong now has 16 blocked operation rows: 13 POST read-query endpoints, 2 multipart uploads, and 1 top-level array write.

### Recommended implementation strategy

1. Preserve operation metadata as the safety gate, but do not add a generic raw API/write/upload command. An operation becomes executable only when its kind, method, content type, body schema, CLI flag mappings/default body, and output/approval policy all pass validator rules.
2. Add a typed operation-backed direct-read path for `rest_read` GET/POST rather than weakening ordinary GET direct reads. Command runner may dispatch `cmd.Operation` only when the connector implements a narrow operation-direct-reader interface and the operation is `rest_read` with `content_type: application/json` (for POST).
3. Build POST read bodies from connector-authored defaults/templates plus typed CLI flags mapped to `body.<field_path>`; forbid inline raw JSON/body flags. Compile and validate the materialized body against `rest.body_schema`, cap encoded body bytes separately from response `max_bytes`, then call `Requester.DoLimited` with `json_redacted` output.
4. Extend write actions for typed non-object bodies instead of bypassing reverse ETL: add schema-gated `body_type` variants such as `json_array` and `multipart`, with explicit source fields/parts. Existing JSON object writes remain unchanged by default.
5. Add `connsdk` multipart support as a streaming/reopenable request primitive with max-byte enforcement before send, per-part metadata, safe local path validation, and no logging of file contents. Dry-run/preview must stat files and redact path/content fields.
6. Bind approval to local payload identity for file/JSON-file inputs. At minimum store and verify size + modtime in the plan hash; for bounded JSON array files (Gong 1 MiB) compute SHA-256 before approval and recheck before execute. Large media uploads may use size+mtime with an explicit preview warning if content hashing is intentionally skipped.
7. Roll out incrementally: first engine/connsdk fixtures, then one low-risk POST read-query fixture, then top-level array JSON with a tiny fixture, then multipart upload with temp files, and only then flip Gong commands from `planned` to `implemented` where typed filters/parts are safe.

### Non-goals / blockers

- No `--body`, `--json`, `--upload`, `--url`, or generic HTTP method/path override.
- No raw arbitrary filter object for broad Gong POST reads unless every accepted field is typed and schema-validated.
- No credentialed Gong execution in local gates.
- No top-level array execution for schemas that remain `additionalProperties: true` without a tighter connector-owned schema or a bounded schema-validated file input.
