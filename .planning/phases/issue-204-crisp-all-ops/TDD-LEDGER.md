# TDD Ledger — Issue #204 Crisp all-ops executable coverage

## GSD mode

- Plan command: `scripts/gsd prompt plan-phase 204 --skip-research`.
- Programming-loop command unavailable: `scripts/gsd prompt programming-loop init --phase issue-204-crisp-all-ops --dry-run` returned `unknown GSD command: programming-loop`.
- Manual GSD/TDD fallback active.

## Ledger

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Generic JSON direct-read policy | `go test ./internal/connectors/commandrunner -run TestRunDirectReadSupportsGenericJSONResponsePolicy -count=1` failed because `json_response` was unsupported | `go test ./internal/connectors/commandrunner ./internal/connectors/engine -count=1` passed after schema/runner/engine support and URI-template query stripping | complete |
| Crisp all-ops mapping | Existing #205 ledger had 220 `blocked` rows and no `writes.json` | Generated 91 implemented direct-read commands and 129 write actions; `go run ./cmd/connectorgen validate internal/connectors/defs` passed | complete |
| Safety/docs | Docs showed metadata-only blocked surface | Docs/catalog/manual now show read/write capabilities, 129 write actions, reverse-ETL approval gates, and destructive confirmation for spam reject/key-roll actions; `./pm docs validate --connectors-dir docs/connectors` passed | complete |

## Safety notes

No live Crisp calls. No credentials. No raw generic HTTP command. Write execution remains reverse-ETL plan → preview → approval → execute.
