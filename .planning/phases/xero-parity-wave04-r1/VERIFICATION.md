# Xero parity verification checklist

## Required local gates

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` with Xero findings clean.
- [x] `go test ./cmd/connectorgen -run 'TestXero' -count=1`.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/xero' -count=1`.
- [x] Focused CLI tests: `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=20m`.
- [x] `go build ./cmd/pm`.
- [x] Runtime docs/help spot checks using built `./pm`: `./pm help connectors`, `./pm connectors`, `./pm connectors inspect xero --json`, `./pm docs validate --connectors-dir docs/connectors`.
- [x] `make connector-boundary`.
- [x] `make verify`.
- [x] `git diff --check`.

## Safety verification

- [x] No live Xero/provider calls; only official public OpenAPI source fetches.
- [x] No secrets in fixtures/docs/generated surfaces.
- [x] No new dependencies.
- [x] No generic HTTP/SQL/GraphQL/shell/file/passthrough escape hatches.
- [x] Destructive writes use typed schemas, redaction where path identifiers appear, and `confirm: "destructive"`.
- [x] Fixture-only evidence remains uncertified.

## Results

- `go run ./cmd/connectorgen validate internal/connectors/defs`: 549 connector(s) checked, 0 findings.
- `go test ./cmd/connectorgen -run 'TestXero' -count=1`: passed.
- `go test ./internal/connectors/conformance -run 'TestConformance/xero' -count=1`: passed.
- `go test ./internal/connectors/engine -run 'TestXeroReportOperationsDirectRead' -count=1`: passed.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=20m`: passed.
- `go build ./cmd/pm`: passed.
- Built-CLI spot checks passed; `pm connectors inspect xero --json` reported 100 streams and 87 write actions.
- `make connector-boundary`: clean outcome.
- `make verify`: passed on the final run. Earlier uncached runs hit package-level timeouts in unrelated full-suite cert/CLI paths; after focused package runs warmed the test cache, the required make gate completed cleanly.
- `git diff --check`: passed.
- Issue addendum marker `xero-parity-wave04-r1-captain-policy-addendum` posted idempotently to #3102-#3109 via `gh-axi issue comment`.
