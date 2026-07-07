# VERIFICATION — Issue #60 GitHub binary transfer

## Planned commands

Focused red/green:

```bash
go test ./internal/connectors/engine -run 'TestBinaryDownload|TestBundleLoadEmbeddedGitHubOperations'
go test ./internal/connectors/commandrunner -run 'TestRunBinary|TestRunGitHubBinary'
go run ./cmd/connectorgen validate internal/connectors/defs/github
go test ./internal/connectors/commandrunner ./internal/connectors/engine
go build ./cmd/pm
```

Broader optional gate before handoff if time permits:

```bash
go test ./internal/connectors/...
make verify
```

## Results

| Command | Result | Evidence |
| --- | --- | --- |
| `go test ./internal/connectors/engine -run 'TestBinaryDownload|TestBundleLoadEmbeddedGitHubOperations'` | pending | Red/green not run yet. |
| `go test ./internal/connectors/commandrunner -run 'TestRunBinary|TestRunGitHubBinary'` | pending | Red/green not run yet. |
| `go run ./cmd/connectorgen validate internal/connectors/defs/github` | pending | Not run yet. |
| `go test ./internal/connectors/commandrunner ./internal/connectors/engine` | pending | Not run yet. |
| `go build ./cmd/pm` | pending | Not run yet. |
| `go test ./internal/connectors/...` | pending | Not run yet. |
| `make verify` | pending | Not run yet. |

## Safety verification

- Secrets: no real secrets requested, read, printed, stored, or summarized.
- Filesystem writes: implementation must restrict writes to explicit safe destinations and use temp-file + rename.
- Binary bytes: implementation must return manifests and must not emit downloaded payload bytes to stdout by default.
- Uploads: deferred unless safe reverse-ETL approval semantics are explicitly added later.
