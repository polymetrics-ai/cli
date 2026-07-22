# TDD Ledger — Gong advanced query/binary policy

## Red

- Added fail-first tests: `TestGongFullSurfaceCommandAndOperationCoverage`, `TestGongMetadataEnablesWriteCapability`, direct-read `json_redacted`/version-prefix tests. Initial run failed on missing cli_surface/writes/operations, write=false metadata, and unsupported `json_redacted` policy.

## Green

- Implemented Gong command surface, expanded streams, bounded GET direct reads, typed JSON writes, POST read-query/multipart operation metadata, generic JSON redaction, docs and website catalog updates. Targeted tests now pass.

## Refactor

- Kept multipart/top-level-array payloads as typed blocked operation metadata pending executor support; no raw write escape hatch added.

## Skills

gsd-core, golang-how-to, golang-cli, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-testing, golang-documentation

## 2026-07-10 engine-support implementation ledger (#252/#253/#254)

Red:
- `internal/connectors/engine`: operation-backed POST read-query sends a connector-authored JSON body, validates `rest.body_schema`, enforces response max bytes, and applies `json_redacted`.
- `internal/connectors/commandrunner`: implemented operation-backed direct-read commands allow only typed `path.*`, `query.*`, and `body.*` mappings; unknown/raw body flags stay blocked.
- `internal/connectors/engine`: `body_type=json_array` writes marshal the selected record value as a top-level JSON array and reject schema mismatches before network send.
- `internal/connectors/connsdk`: multipart requests stream bounded file parts with auth/default headers, correct `Content-Type` boundary, and retry-safe file reopening; too-large paths fail before network send.
- `internal/connectors/engine`: multipart write support enforces project-root path safety, symlink escape prevention, and byte caps.
- `cmd/connectorgen`: validator rejects implemented commands for unsupported operation shapes, raw body mappings, missing body schemas, missing max-byte caps, and unsupported content types.

Green:
- Implemented operation direct-read, top-level JSON array writes, and bounded multipart writes without raw body/upload escape hatches.
- Flipped only safe Gong engine-shape commands; broad arbitrary-filter POST reads remain planned until safe typed filter flags are authored.
- Docs/manual/skill/website generated artifacts were updated.

