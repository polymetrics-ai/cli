# TDD ledger — Airtable official API parity

## Red/green plan

| Slice | Red evidence before production edit | Green evidence target | Final evidence |
| --- | --- | --- | --- |
| Audit ledger | Current `api_surface.json` had 30 rows, not the 103 official OpenAPI operations. | Bundle validation accepts 103 partitioned rows. | `api_surface.json` tracks all 103 audited operations: 31 streams, 70 writes, 1 direct read, 1 blocked CSV import; root `connectorgen validate` has 0 findings. |
| Streams/fixtures | Current bundle had 5 streams and fixtures only for those streams. | Conformance `TestConformance/airtable` passes with fixtures for every executable stream. | 31 streams with sanitized fixtures; conformance passed. |
| Writes/fixtures | Current bundle had 12 write actions and lacked broad destructive/admin coverage. | Conformance write request-shape checks pass for every executable write action. | 70 typed write actions with fixtures; conformance write request-shape checks passed. |
| Direct read CLI | Current Airtable bundle had no `operations.json`/`cli_surface.json`, so `pm airtable hyperdb get-records` was unknown. | `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` includes Airtable direct read coverage and passes. | Added HyperDB operation/CLI command and fixture server test; CLI dynamic/golden gate passed. |
| Docs/catalog | Current generated docs/catalogs reported 5 streams / 12 writes. | Generated docs/catalogs report actual post-change stream/write/operation counts. | Regenerated Airtable docs, connector catalog, website connector data, and golden transcripts. |

## Captured red baseline

- `internal/connectors/defs/airtable/api_surface.json`: 30 endpoint rows; 17 covered; 13 excluded; no operation rows.
- `internal/connectors/defs/airtable/streams.json`: 5 streams.
- `internal/connectors/defs/airtable/writes.json`: 12 actions.
- `cli_surface.json`: missing.
- `operations.json`: missing.
- `certification.json`: missing.
- Official OpenAPI re-audit: 103 operations with lane counts `27/69/1/1/5`.

## Verification ledger

See `VERIFICATION.md` for command results. The recovered tree now passes the exact single-bundle validate command after the HyperDB direct-read CLI flag was marked required to match the typed `body.primaryKeys` schema; root validation and all local fixture gates also passed before the recovery merge commit.
