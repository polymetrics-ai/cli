# TDD Ledger — Issue #205 Crisp CLI surface metadata

## Manual GSD/TDD fallback

`gsd-programming-loop` is not registered in the current `scripts/gsd` command registry. This issue follows the manual universal programming loop: plan, capture red validation, implement smallest metadata scaffold, run targeted validation, then update verification evidence.

## Ledger

| Step | Command | Expected | Actual | Status |
|---|---|---|---|---|
| Red validation | `go run ./cmd/connectorgen validate internal/connectors/defs/crisp` | fails because bundle absent | failed as expected: `validate: read root: open .: no such file or directory` | complete |
| Implement metadata scaffold | write Crisp `metadata.json`, `spec.json`, empty `streams.json`, `api_surface.json`, `cli_surface.json`, `docs.md` | files added | added 220-operation blocked metadata scaffold | complete |
| Targeted green | `tmp=$(mktemp -d); cp -R internal/connectors/defs/crisp "$tmp/crisp"; go run ./cmd/connectorgen validate "$tmp"` | pass | `connectorgen validate: 1 connector(s) checked, 0 findings` | complete |
| Fleet green | `go run ./cmd/connectorgen validate internal/connectors/defs` | pass | `connectorgen validate: 548 connector(s) checked, 0 findings` | complete |
| Conformance smoke | `go test ./internal/connectors/conformance -run 'TestConformance/crisp'` | pass | `ok polymetrics.ai/internal/connectors/conformance 3.733s` | complete |
| Inspect smoke | `go run ./cmd/pm connectors inspect crisp --json` parsed via Python | pass | output `connector.name=crisp` | complete |
| Catalog count regression | `go test ./...` after adding bundle | failed on fixed connector-count assertions (`count 551`, bundle count `547`) and default 10m certify timeout | updated expected counts/help/docs artifacts; re-ran with project timeout | complete |
| Full local gate | `make verify` | pass | passed: fmt, tidy-check, vet, test, build, docs validate, smoke, lint, connectorgen validate | complete |

## Behavior-change note

#205 is metadata/validation only. It does not add executable streams, direct-read dispatch, write actions, binary downloads, or reverse-ETL execution. Later issues own behavior tests. The branch also updates catalog-count tests/docs because adding any declarative bundle increments runtime catalog totals.
