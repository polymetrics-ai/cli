# TDD ledger: Hubplanner parity wave05-r1

| Slice | Red / expected failing evidence | Green evidence | Notes |
| --- | --- | --- | --- |
| Baseline | Current `api_surface.json` had 7 covered rows and no operation-ledger mode; parent #3239 required 107 official rows. | `api_surface.json` now has 107 rows: 97 implemented, 10 blocked webhook events. | Manual GSD fallback because `scripts/gsd prompt programming-loop ...` is unavailable. |
| Streams | Adding streams without matching fixtures/surface coverage would fail `fixtures_present`, `read_fixture_nonempty`, and `surface_complete`. | `go test ./internal/connectors/conformance -run 'TestConformance/hubplanner' -count=1` passed. | 20 stream rows/fixtures; resources retains two-page pagination fixture. |
| Writes | New writes without closed schemas/fixtures would fail `write_schemas_valid` and `write_request_shape:<action>`. | Same conformance run passed 61 write fixtures. | Delete actions have `confirm: destructive`; no generic write path/body/query. |
| Direct reads / CLI | Implemented direct-read rows without command metadata would fail `surface_complete`. | `go run ./cmd/connectorgen validate internal/connectors/defs` and `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` passed. | 16 bounded operation direct-read commands, including the 2 official custom-field provider searches. |
| Final verification | N/A | `make verify` passed. | Full local gate includes gofmt, tidy-check, vet, tests, build, docs validate, smoke, lint, connectorgen validate, connector-boundary, release workflow check. |
