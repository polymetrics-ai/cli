# Verification

Phase: native-runtime-capability-matrix

## Automated Checks

- `go test ./internal/connectors -run TestConnectorCatalog` passed.
- `go test ./internal/cli -run TestConnectorCatalog` passed.
- `go run ./cmd/pm-cataloggen` generated 647 connector definitions.
- `go build -o pm ./cmd/pm` passed.
- `./pm docs generate --dir docs/cli --connectors-dir docs/connectors` passed.
- `./pm docs validate --connectors-dir docs/connectors` passed.
- `node /Users/karthiksivadas/.codex/skills/gsd-programming-loop/scripts/tdd-gate.mjs --phase native-runtime-capability-matrix` passed.
- `go test ./...` passed.
- `go vet ./...` passed.
- `make verify` passed, including the sample ETL and reverse ETL smoke flow.
- `make install` installed `pm` to `/Users/karthiksivadas/.local/bin/pm`.

## CLI Spot Checks

- `./pm connectors inspect source-github --json` reports:
  - `implementation_status=enabled`
  - `runtime_capabilities.read=true`
  - `runtime_capabilities.write=true`
  - `runtime_capabilities.reverse_etl=true`
- `./pm connectors inspect destination-postgres --json` reports:
  - `implementation_status=planned_native_port`
  - `runtime_capabilities.read=false`
  - `runtime_capabilities.write=false`
  - non-empty `runtime_capabilities.unsupported_reason`
- `./pm connectors inspect destination-postgres` renders `RUNTIME CAPABILITIES` before configuration details.
- `./pm connectors catalog --type destination --stage generally_available --json` returns `count=9` and includes `destination-postgres`.

## GSD Helper Warning

`programming-loop.mjs verify --phase native-runtime-capability-matrix --execute` still reports missing inferred commands and `git diff --check` failure because this workspace is not a git worktree and the helper does not detect the Makefile checks from `docs/architecture/repo-profile.json`. The explicit local verification above passed.
