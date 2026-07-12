# TDD Ledger: Issues #107-#110 Jira Full Surface

## Preflight

```bash
scripts/gsd prompt plan-phase issue-107-110-jira-full-surface --json
scripts/gsd prompt programming-loop init --phase issue-107-110-jira-full-surface --dry-run
```

Result: plan prompt generated; programming-loop command unavailable in this adapter, so manual GSD/TDD fallback is active.

## Planned red tests

```bash
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedJiraFullSurface' -count=1
go test ./internal/connectors/engine -run 'TestJiraFullSurfacePolicy' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs/jira --json
```

Expected initial failures:

- `operations.json` is absent or incomplete for Jira.
- Jira `api_surface.json` does not enumerate all 620 official OpenAPI operations in operation-ledger mode.
- Direct-read/write/binary/sensitive policy metadata is not complete for #107-#110.

## Red evidence

```bash
gofmt -w internal/connectors/engine/bundle_test.go internal/connectors/engine/direct_read_test.go
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedJiraFullSurface|TestBundleLoadDiskJiraAPISurfaceFullCoverage|TestDirectReadRedactsGenericJSONSensitiveFields' -count=1
```

Result: failed as expected.

```text
Jira Operations = 0, want official OpenAPI count 620
Jira Surface operation_ledger_version = 0, want 1
DirectRead: direct read output policy "json_redacted" is not supported
```

## Green evidence

Targeted green checks passed:

```bash
gofmt -w cmd/connectorgen/validate.go internal/connectors/commandrunner/runner.go internal/connectors/engine/bundle.go internal/connectors/engine/bundle_test.go internal/connectors/engine/direct_read.go internal/connectors/engine/direct_read_test.go
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedJiraFullSurface|TestBundleLoadDiskJiraAPISurfaceFullCoverage|TestDirectReadRedactsGenericJSONSensitiveFields' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/engine -count=1
go test ./internal/connectors/commandrunner -count=1
go test ./internal/cli -run 'TestJiraCommandSurfaceRunsGeneratedDirectRead|TestJiraCommandSurfaceRunsStreamBackedCommands|TestJiraConnectorCommandSurfaceHelp' -count=1
go test ./cmd/connectorgen -count=1
```

Implemented full-surface counts:

```text
operations=620: rest_write=333, rest_read=271, rest_query=7, binary_download=5, file_upload=4
api_surface=620: GET=276, POST=135, PUT=119, DELETE=90
api_surface coverage: write=333, direct_read=268, stream=3, blocked=16
writes=333, destructive confirmations=256
cli commands=303, implemented=271
```

Generated docs/website data checks passed:

```bash
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
go run ./cmd/pm docs validate --connectors-dir docs/connectors
cd website && pnpm gen:website-data && pnpm test:unit -- connector-data
```

Full verification passed after registry-load optimization:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
cd website && pnpm build
```

## Refactor evidence

- Added `json_redacted` direct-read output policy for generic bounded JSON GET endpoints and recursive redaction of content/download/secret-shaped fields.
- Added `rest_query` operation kind for safe REST read-query POST operations that need fixed body-variable execution (#109) instead of misclassifying them as writes.
- Added Jira `operations.json`, `writes.json`, operation-ledger `api_surface.json`, and generated fixed-endpoint direct-read command metadata from the official Jira Cloud OpenAPI.
- Kept binary downloads and file uploads inventoried with max-byte policy but blocked by default until a safe local file executor is enabled.
- Cached embedded bundle loading in `internal/connectors/bundleregistry` so repeated CLI/test app opens do not re-parse the expanded full-surface metadata; this fixed the initial `go test ./...` timeout in `internal/connectors/certify`.
