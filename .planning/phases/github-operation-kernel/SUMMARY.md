# Summary: GitHub Operation Kernel

Issue: #56
Parent issue: #44
Branch: `feat/56-operation-kernel`

## Delivered

- Added optional connector `operations.json` metadata with typed operation kinds
  for REST, GraphQL, XML, binary/file transfer, controlled local git/file,
  browser, and composite operations.
- Added `operation` references to `cli_surface.json` commands and public command
  surface conversion.
- Added validation that implemented commands reference exactly one executable
  target: stream, write action, or operation.
- Added operation semantic validation for unique IDs, exact execution block
  matching, safe method classes, mutation approval metadata, and bounded binary
  operations.
- Added pre-credential command preflight so operation-backed/planned/unsafe
  commands block before app open or vault secret resolution.
- Added explicit `covered_by.direct_read(s)` API-surface coverage so blocked
  operation-ledger rows stay documentation-only.
- Added `operations.json` secret-shaped literal scanning.
- Added GitHub seed operation definitions for project listing, issue deletion,
  release asset download, and repository clone workflows.
- Kept execution fail-closed: operation-backed commands return a blocked
  command error until executor slices land in later issues.
- Documented the operation kernel architecture and safety boundaries.

## Verification

Green:

- `go test ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner`
- `go test ./internal/cli -run 'TestGitHubCommandSurfaceBlocksOperationBeforeCredentialResolution|TestGitHubCommandSurfaceBlocksReverseETLCommand|TestGitHubCommandSurfaceRunsStreamBackedIssueList|TestGitHubCommandSurfaceRunsDirectReadFile'`
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`
- `go test ./cmd/...`
- `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/bundleregistry ./internal/connectors/conformance`
- `go vet ./...`
- `go build ./cmd/pm`

Warning:

- `go test ./...` did not produce a final status because its PTY stayed open
  after no matching process was visible. It was interrupted and recorded as
  inconclusive.

## Next Slices

- #57: reverse ETL command executor with plan, preview, approval, execute.
- #58: GraphQL executor foundation.
- #60: binary/file transfer executor.
- #61: restricted local git executor.
- #62: hard parity gate.
