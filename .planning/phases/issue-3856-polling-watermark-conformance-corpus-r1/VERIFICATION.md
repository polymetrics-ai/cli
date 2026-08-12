# Verification — #3856 polling conformance corpus

## Planned gates

| Area | Command/check | Status |
| --- | --- | --- |
| R1 RED/GREEN | focused conformance test before/after implementation | pending |
| Polling engine | focused and full `internal/connectors/engine` tests | pending |
| Generic contract non-regression | `internal/synccontract` tests and unchanged `v1.json` digest | pending |
| App regression | `go test -timeout 20m ./internal/app -count=1` | pending |
| Runtime preflight sweep | focused commandrunner test | pending |
| Race safety | focused engine `-race` test | pending |
| Format/vet/build | scoped `gofmt`, `go vet ./...`, `go build ./cmd/pm` | pending |
| Repository gates | individual AGENTS.md gates | pending |
| Exact binary | clean child-head `pm` byte count, SHA-256, and applicable no-credential smoke | pending |
| GSD verify/review | generated prompt, durable verification/review report | pending |
| Independent audit | required Sol audit after final candidate | pending |
| no-mistakes | `--skip=push,pr,ci`, no `--yes` | pending |

## Explicit exclusions

Shepherd #3995 / PR #4062 is parked and excluded. No provider, credential,
database, live mutation, CDC promotion, or certification test occurs in this
corpus-only phase. Credential-backed Transport is deferred to the first
executable provider-boundary child and final #4019 gate.
