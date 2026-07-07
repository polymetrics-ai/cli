# Verification

## Commands

```bash
go test ./internal/cli -run 'TestGitHubCommandWriteUsesReversePlanApproval|TestConnectorCommand'
jq empty internal/connectors/defs/github/operations.json website/package.json
go run ./cmd/connectorgen validate internal/connectors/defs --json
node website/scripts/gen-connector-bundles.mjs && node website/scripts/gen-connector-catalog.mjs
```

## Result

All commands passed after removing the unsupported JSON Schema conditional-keyword edit.

## Review Disposition

- Accepted: redundant connector write precheck removed.
- Accepted: httptest handler no longer calls `t.Fatalf` on the handler goroutine.
- Accepted: `github.issue.delete` now declares `repo` and `public_repo` scopes.
- Accepted: website CLI surface mapping is shared by bundle and catalog generators.
- Declined for this slice: conditional `api_surface.schema.json` validation. The repo's embedded
  meta-schema compiler does not support `allOf`, `if`, `then`, or `anyOf`; the same invariant is
  already enforced by `cmd/connectorgen` validation and tests.
