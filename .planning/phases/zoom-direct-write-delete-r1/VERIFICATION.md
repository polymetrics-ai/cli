# Verification — Zoom direct-write/delete deferral

- [x] Source inventory proves 61 non-upload REST writes, including 18 deletes.
- [x] Every deferred REST write is listed individually as recoverable `foundation-gap`.
- [x] Eight load-bearing upload operations are recorded as `schema-incompatible` with a foundation-gap recovery path.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/zoom` passes after invalid draft declarations are removed.
- [x] `make verify` passes before push.

## Evidence

- Source inventory: `jq` over PR #3951 operations produced 69 REST writes, 18 deletes, and eight load-bearing upload/multipart operations; 61 REST writes remain after those upload rejections.
- `connectorgen validate` rejected every attempted operation-backed REST coverage as unsupported: `covered_by.operations` is fixed-GraphQL-only and `covered_by.write` requires a `writes.json` reverse-ETL action.
- `rejections.json`: 61 exact operation IDs, all `foundation-gap`, `recoverable: true`; its count check reports 18 deletes.
- `foundation-gaps.json`: exact minimal foundation locations and recovery condition, without an engine patch.
- Final `make verify` — pass (exit 0), including the repository-wide test suite, generated-file checks, connector boundary, connector canon, and release checks.
