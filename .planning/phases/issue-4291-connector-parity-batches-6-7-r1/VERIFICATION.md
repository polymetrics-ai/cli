# Verification — issue #4291

## Artifact-level red/green evidence

- **RED — batch 6:** `test ! -f` for every proposed batch-6 source lock and disposition ledger passed before implementation.
- **RED — batch 7:** `test ! -f` for every proposed batch-7 source lock and disposition ledger passed before implementation.
- **GREEN — complete map:** the issue-local strict ledger-invariant check passed: all 20 canonical connector directories have exact source-lock and `api_surface.json` coverage, no duplicate `method/path`, no endpoint has `parity_class: reverse_etl`, each enabled endpoint has `covered_by.direct_read`, `covered_by.direct_reads`, or `covered_by.write`, every typed write action is an enabled `direct_write` with the required separate reverse-ETL gap, and unbound stream rows are `declaration-pending`. Totals: **2,099 documented, 575 enabled, 348 commands, 448 writes, 211 deletes**.

## Repository checks

| Command | Result |
| --- | --- |
| `go run ./cmd/connectorgen validate` | pass — `552 connector(s) checked, 0 finding(s)` |
| `go run ./cmd/connectorgen surface-sync --check` | pass — `552 connector(s) scanned, 0 field(s) need synchronization` |
| `go test -timeout 20m ./cmd/connectorgen` | pass |
| `go test -timeout 20m ./internal/connectors/engine` | pass |
| `go test -timeout 20m ./internal/cli` | pass |
| `go vet ./...` | pass |
| `go build ./cmd/pm` | pass |
| `gofmt -w cmd internal && go mod tidy && git diff --exit-code -- go.mod go.sum` | pass — formatting/mod-tidy introduced no module drift |
| `./pm docs validate --connectors-dir docs/connectors` | pass |
| `make smoke-no-build` | pass |
| `make lint` | pass — `golangci-lint` reported `0 issues` |
| `make agent-contract-check` | pass |
| `make connectorgen-validate` | pass |
| `make connectorgen-surface-sync` | pass |
| `node --test scripts/tests/github-combined-operation-ledger.test.mjs scripts/tests/gen-github-graphql-parity.test.mjs scripts/tests/github-source-drift.test.mjs` | pass — 15 tests, 0 failures |
| `node scripts/gen-github-graphql-parity.mjs --check` | pass |
| `node scripts/github-combined-operation-ledger.mjs --check` | pass |
| `go run ./cmd/connectorgen certification-matrix --check` | pass |
| `go run ./cmd/connectorgen certification-candidates --connector github --check` | pass |
| `go run ./cmd/connectorgen certification-sweep --connector github --check` | pass |
| `go run ./cmd/connectorgen boundary . --json` (captured to a temporary log and polled to completion) | pass |
| `bash scripts/tests/connector-canon.sh` | pass |
| `./scripts/tests/pinned-build-dependencies.sh` | pass |
| `./scripts/tests/homebrew-release-notify.sh` | pass |
| `./scripts/tests/release-target-parity.sh` | pass |

`go test -timeout 20m ./...` and aggregate `make verify` were deliberately not run as single commands: the repository AGENTS instruction says agents under a per-command timeout must run changed packages plus `internal/cli` separately and execute `make verify`'s non-suite gates individually, because the full 550+ connector suite is routinely cut off and indistinguishable from a hang. The targeted package tests and each applicable gate above were run; CI retains the full suite.
