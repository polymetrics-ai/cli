# Issue 476 Verification

Status: `pending`

No verification result is claimed before the declared commands run. Full verification requires a
successful `make verify`; focused checks alone do not set `verificationPassed`.

| Gate | Result | Evidence |
|---|---|---|
| focused issue tests | pending | — |
| full Shepherd tests | pending | — |
| strict TypeScript / Pi 0.80.6 | pending | — |
| Pi extension discovery | pending | — |
| diff hygiene | pending | — |
| `go vet ./...` | pending | — |
| `go test ./...` | pending | — |
| `go build ./cmd/pm` | pending | — |
| `make verify` | pending | — |
| exact remote head | pending | — |
| stacked ready PR | pending | — |

Runtime-backed services are not part of this issue and will not be started.

