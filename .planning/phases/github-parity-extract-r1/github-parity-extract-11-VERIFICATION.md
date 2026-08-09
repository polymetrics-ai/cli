# Verification — plan 11 generated fixed GraphQL contracts

All evidence below is local and hermetic. No `pm` provider command, credential lookup, browser
action, fixture creation, cleanup retry, or provider write occurred in this checkpoint.

| Gate | Result |
| --- | --- |
| fixed GraphQL read/write redaction and one-document boundary regressions | PASS — `go test -timeout 20m ./internal/connectors/engine -count=1` |
| structured GraphQL-variable and command dispatch boundary | PASS — `go test -timeout 20m ./internal/connectors/commandrunner -count=1` |
| source generated CLI/API validation, including GraphQL read capability accounting | PASS — `go test -timeout 20m ./cmd/connectorgen -count=1` |
| static source and runtime operation transport coverage | PASS — `go test -timeout 20m ./internal/connectors/conformance ./internal/connectors/certify -count=1` |
| source-lock/GraphQL-generator/source-drift tests | PASS — 13 Node tests |
| generated GraphQL artifacts | PASS — `node scripts/gen-github-graphql-parity.mjs --check` |
| combined ledger | PASS — `1220 REST + 31 Query + 274 Mutation = 1525` |
| permanent artifact target | PASS — `make github-parity-artifacts-check` |
| all bundle validation | PASS — `connectorgen validate`: 551 connectors, 0 findings |
| derived-surface check | PASS — `surface-sync --check`: no drift |
| PM-only lab regression/boundary | PASS — 23 tests; one exact allowed target |
| changed-package vet and binary build | PASS |
| full `internal/app` and `internal/cli` suites | PASS |
| generated website script suite | PASS — 27 tests |
| whitespace/conflict scan | PASS — `git diff --check` |

Representative runtime help was checked from the freshly built binary:

- `pm github graphql` lists the 31 query and 274 mutation commands.
- `pm github graphql query viewer --help` reports the fixed direct-read operation.
- `pm github graphql mutation create-enterprise-organization --help` reports one required typed
  `--input` flag and the plan → preview → approval → execute lifecycle.
- `pm github graphql mutation delete-issue --help` remains explicitly
  `unsafe_or_disallowed` with the recorded provider/product block.

This is not current-head live evidence. The generated combined ledger intentionally reports
`live_proof: 0/1525`; the subsequent PM-only cohort work must create its own current-head terminal
evidence under the lab boundary.
