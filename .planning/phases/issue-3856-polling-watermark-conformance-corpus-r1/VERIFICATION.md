# Verification — #3856 polling conformance corpus

## Planned gates

| Area | Command/check | Status |
| --- | --- | --- |
| R1 RED/GREEN | focused conformance test before/after implementation | focused GREEN recorded; broad validation deferred |
| Polling engine | focused and full `internal/connectors/engine` tests | focused full engine pass; broad validation deferred |
| Generic contract non-regression | `internal/synccontract` tests and unchanged `v1.json` digest | focused fixture regression pass; generic corpus untouched |
| App regression | `go test -timeout 20m ./internal/app -count=1` | pending |
| Runtime preflight sweep | focused commandrunner test | pending |
| Race safety | focused engine `-race` test | pending |
| Format/vet/build | scoped `gofmt`, `go vet ./...`, `go build ./cmd/pm` | pending |
| Repository gates | individual AGENTS.md gates | pending |
| Exact binary | clean child-head `pm` byte count, SHA-256, and applicable no-credential smoke | pending |
| GSD verify/review | generated prompt, durable verification/review report | pending |
| Independent audit | required Sol audit after final candidate | pending |
| no-mistakes | `--skip=push,pr,ci`, no `--yes` | pending |

## Serialized validation hold

paused: #3856 focused implementation complete; awaiting serialized broad validation gate

The captain's resume record permits focused implementation and tests only at
this point. No broad repository suite, race-heavy sweep, no-mistakes run,
push, PR creation, CI, or merge has been started. The focused GREEN evidence
and exact commands are recorded in `TDD-LEDGER.md`.

## Explicit exclusions

Shepherd #3995 / PR #4062 is parked and excluded. No provider, credential,
database, live mutation, CDC promotion, or certification test occurs in this
corpus-only phase. Credential-backed Transport is deferred to the first
executable provider-boundary child and final #4019 gate.
