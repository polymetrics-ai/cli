# TDD Ledger — issue #254 Gong bounded typed multipart uploads

## Red (planned before production code)

- connsdk multipart test proves boundary/content type, file part, auth/default headers, and response handling.
- connsdk safety tests fail for too-large, missing, traversal, and unsafe paths before network send.
- Engine multipart write sends only declared parts and enforces max bytes.
- Preview/redaction test for multipart fields.
- Validator rejects unsafe implemented multipart commands.

## Green

- Added stdlib-only `connsdk.Requester.DoMultipart` with boundary/content-type handling, auth reuse, retry-safe file reopening, pre-send regular-file and size checks.
- Added engine `body_type: multipart` write support with declared field/file parts, project-root path validation, symlink escape prevention, per-part and total byte caps.
- Added Gong `upload_call_media` and `upload_crm_entities` multipart write actions and implemented CLI surface entries.
- Extended commandrunner redaction to cover multipart file/path fields in previews.
- Added reverse-plan payload identity binding for local file uploads (path hash plus size/mtime) so changed files invalidate approvals before execution.

Evidence:
- `go test ./internal/connectors/connsdk -run Multipart -count=1`
- `go test ./internal/connectors/engine -run 'OperationDirectRead|WriteJSONArray|WriteMultipart|DirectRead|Write' -count=1`
- `go test ./internal/connectors/commandrunner -run 'OperationDirectRead|DirectRead|RedactRecord' -count=1`
- `go test ./internal/app -run PayloadIdentities -count=1`
- `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`

## Refactor

- Implemented multipart as connector-authored write metadata only; no generic upload command or arbitrary local file egress path was added.
- File content/path values remain redacted in command previews and docs.

## Skills

gsd-core, golang-how-to, golang-cli, golang-testing, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-context, golang-documentation.
