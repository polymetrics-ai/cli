# TDD Ledger — Issue #185

## Red result

Added tests for:

- commandrunner maps `--id` on Freshchat `user fetch` to a direct-read request body (`ids` array) for `POST /users/fetch`;
- engine direct read can safely execute the bounded Freshchat users-fetch POST policy.

Initial command:

```bash
gofmt -w internal/connectors/commandrunner/runner_test.go internal/connectors/engine/direct_read_test.go
go test ./internal/connectors/commandrunner -run TestRunFreshchatUsersFetchDirectReadCommand
go test ./internal/connectors/engine -run TestDirectReadFreshchatUsersFetchPOST
```

Expected failure observed: direct-read request body support was absent (`connectors.DirectReadRequest` had no `Body` field), and direct-read POST policies were not implemented.

## Green result

Implemented the narrow body-mapped Freshchat direct-read policy and marked Freshchat `user fetch` implemented.

```bash
go test ./internal/connectors/commandrunner -run 'TestRunFreshchatUsersFetchDirectReadCommand|TestRunDirectReadRejectsUnsafeEndpointMetadata|TestRunDirectReadRequiresOutputPolicy'
go test ./internal/connectors/engine -run 'TestDirectReadFreshchatUsersFetchPOST|TestDirectReadRejectsMutationMethod'
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestValidate_APISurface|TestFreshchatAPISurfaceLedger'
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results:

- Commandrunner focused tests: pass.
- Engine focused tests: pass.
- Connectorgen focused tests: pass.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: pass, `547 connector(s) checked, 0 findings`.

Follow-up full-gate finding: `go test ./...` exposed stale conformance validation that required all direct reads to be GET-only. Added the conformance policy allowlist for `freshchat_users_fetch` and re-ran:

```bash
go test ./internal/connectors/conformance -run 'TestConformance/freshchat|TestCheckSurfaceComplete'
go test ./cmd/connectorgen -run 'TestValidate_APISurface|TestFreshchatAPISurfaceLedger'
go run ./cmd/connectorgen validate internal/connectors/defs
```

All passed.

## Verification ledger

Focused green gates and full handoff gates pass:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```
