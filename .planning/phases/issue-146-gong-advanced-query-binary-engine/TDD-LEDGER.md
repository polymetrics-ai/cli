# TDD Ledger — Gong advanced query/binary policy

## Red

- Added fail-first tests: `TestGongFullSurfaceCommandAndOperationCoverage`, `TestGongMetadataEnablesWriteCapability`, direct-read `json_redacted`/version-prefix tests. Initial run failed on missing cli_surface/writes/operations, write=false metadata, and unsupported `json_redacted` policy.

## Green

- Implemented Gong command surface, expanded streams, bounded GET direct reads, typed JSON writes, POST read-query/multipart operation metadata, generic JSON redaction, docs and website catalog updates. Targeted tests now pass.

## Refactor

- Kept multipart/top-level-array payloads as typed blocked operation metadata pending executor support; no raw write escape hatch added.

## Skills

gsd-core, golang-how-to, golang-cli, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-testing, golang-documentation

## 2026-07-10 engine-support planning ledger (analysis only)

No production code was changed in this slice. Planned fail-first coverage before any executor implementation:

Red tests to add first:
- `internal/connectors/engine`: non-GraphQL stream/read fixed bodies are honored where explicitly declared, and operation-backed POST read-query sends a connector-authored JSON body, validates it against `rest.body_schema`, enforces response max bytes, and applies `json_redacted`.
- `internal/connectors/commandrunner`: implemented operation-backed direct-read commands allow only typed `path.*`, `query.*`, and `body.*` mappings; unknown/raw body flags stay blocked.
- `internal/connectors/engine`: `body_type=json_array` writes marshal the selected record/file value as a top-level JSON array, not as an object wrapper, and reject schema mismatches.
- `internal/connectors/connsdk`: multipart requests stream bounded file parts with auth/default headers and correct `Content-Type` boundary; too-large, missing, traversal, and symlink/unsafe paths fail before network send.
- `internal/connectors/engine`: multipart write preview redacts file/path/content-like fields and binds approval to payload metadata.
- `cmd/connectorgen`: validator rejects implemented commands for unsupported operation shapes, raw body mappings, missing body schemas/templates, missing max-byte caps, and unsupported content types.
- Gong fixture/definition tests: blocked operations only become implemented after typed filters/parts/body schemas are present and docs/help parity is updated.

