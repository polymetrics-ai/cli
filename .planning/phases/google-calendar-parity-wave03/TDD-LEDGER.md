# Google Calendar parity TDD ledger

## Red / failing evidence targets

- `go run ./cmd/connectorgen validate internal/connectors/defs/google-calendar` should fail while the operation ledger references streams/writes/operations before their fixtures/schemas are complete.
- `go test ./internal/connectors/conformance -run 'TestConformance/google-calendar' -count=1` should fail until every executable stream/write fixture matches the declarative request shape and dynamic auth replay remains fixture-only.
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` should fail or require regenerated manual/catalog/CLI golden surfaces after the connector capabilities and command surface change.

## Green evidence

- `go run ./cmd/connectorgen validate internal/connectors/defs/google-calendar` passed with zero findings.
- `go test ./internal/connectors/conformance -run 'TestConformance/google-calendar' -count=1` passed with static + dynamic fixture replay and write request-shape coverage.
- `go test ./internal/connectors/hooks/google-calendar -count=1` passed, including conformance-only auth no-op, live synthetic-value refresh negative coverage, sanitized OAuth refresh exchange, redacted missing-refresh-token error, query-bearing write checks, EventDateTime validation, and fixture-only `freebusy query` operation direct read.
- `./pm connectors inspect google-calendar --json` reports `read=true`, `write=true`, 11 streams, and 26 write actions without reading credentials.

## Post-change count assertions verified

- Official operation total remains 38 from the discovery document.
- Implemented count is 38 executable `api_surface.json` covered rows: 11 streams, 26 named reverse-ETL write actions, and 1 direct-read operation.
- Fixture/tested count is 38 in fixture-only local coverage: 11 stream fixtures, 26 write request-shape fixtures with exact query assertions where declared, and 1 direct-read operation test.
- Blocked/planned operation rows are 0; `cdc=false` remains a runtime capability note because watch/stop operations only manage provider channels.
- Certified count remains 0; no live provider certification is performed.

## Red/green log

- Red baseline: `python3 - <<'PY' ... len(api_surface.endpoints) != 38 ... PY` reported `baseline api_surface_rows=4 covered=4 target=38` and exited 1 before production connector edits.
