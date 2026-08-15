# Verification — #3860 polling-watermark truth surfaces

| Area | Command/evidence | Status |
| --- | --- | --- |
| TDD red/green | Focused surface tests and `TDD-LEDGER.md` | pending |
| CLI/connectors regression | `go test -timeout 20m ./internal/cli/... ./internal/connectors/...` | pending |
| Fresh binary | `go build -o ./bin/pm ./cmd/pm` before regeneration | pending |
| Runtime parity | `./bin/pm help connectors`; `./bin/pm connectors`; `./bin/pm connectors --help`; inspect/catalog commands | pending |
| Generated docs/website | Sanctioned generator plus exact diff review and stale-claim grep | pending |
| Live database lane | Supplied `POLYMETRICS_DATABASE_INTEGRATION=1 ... go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/connectors/native/postgres` | pending |
| Repository gates | individual `make verify` component gates | pending |
| GSD verification/review | generated verify-work and code-review prompts executed inline, with findings dispositioned | pending |
| Delivery | rebase, push, PR to exact integration base, API read-back | pending |

## Safety and scope

- No credentials, provider calls, generic SQL/HTTP/shell write, reverse ETL execution, dependency, or driver change.
- #4125, #4136, #4090, and #4154 remain excluded.
- Generated artifacts are only modified by their sanctioned generator after building a fresh binary.
