# Plan — Issue #186 Freshchat advanced query/binary engine

Refs #186, #180.

## Scope

Narrow #186 to Freshchat multipart/binary upload parity metadata and fail-closed runtime behavior:

- add typed `file_upload` operation metadata for the official Freshchat `/files/upload` and `/images/upload` endpoints;
- connect the visible `pm freshchat file upload` and `pm freshchat image upload` command-surface entries to those typed operations while keeping them non-executable until a bounded file-upload executor exists;
- prove commandrunner blocks the typed upload operations before credential resolution and without reading local files;
- preserve the operation ledger: the two upload endpoints stay blocked by default and no generic raw upload/HTTP body executor is exposed.

## Out of scope

- Implementing multipart upload execution.
- Reading local files, binary bytes, or credentials.
- Any Freshchat write execution or reverse ETL execution.
- Changing auth scopes or adding dependencies.

## Required skills

gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-structs-interfaces, golang-documentation.

## TDD slices

1. Red: embedded Freshchat operations should include two typed `file_upload` operations with positive `max_bytes` and approval text.
2. Red: commandrunner should block `pm freshchat file upload` and `pm freshchat image upload` as typed operation-backed commands before credential resolution.
3. Green: add `operations.json`, wire upload commands to operation IDs, update docs/generated surfaces, and validate the bundle.

## Verification

Focused:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatFileUploadOperations
go test ./internal/cli -run TestFreshchatUploadCommandsBlockTypedOperationsBeforeCredentialResolution
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestFreshchatAPISurfaceLedger|TestFreshchatBinaryUploadCommandsUseTypedOperations'
go run ./cmd/connectorgen validate internal/connectors/defs
```

Full handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```
