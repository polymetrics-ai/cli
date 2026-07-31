# Google Calendar parity summary

Status: implemented and locally verified.

## Delivered

- Replaced the read-only/quarantine Google Calendar bundle with a declarative Calendar API v3 parity surface.
- Added a complete `api_surface.json` operation ledger for all 38 official operations.
- Added 11 executable read streams, 26 named reverse-ETL write actions, and one bounded direct-read operation (`freebusy query`).
- Added sanitized check/read/write fixtures plus a fixture-only direct-read operation test.
- Replaced native delegation hooks with a Google Calendar OAuth refresh-token `AuthHook`; fixture auth no-op is gated by the internal conformance sentinel only, not user config.
- Promoted the google-calendar runtime registration to the engine bundle so inspect/catalog surfaces expose the new streams/writes.
- Regenerated Google Calendar connector docs/catalog/website/golden surfaces.
- Added schema `oneOf`/`anyOf` support needed for typed Calendar EventDateTime validation and conformance query assertions for query-bearing writes.
- Added manual-GSD artifacts, TDD ledger, and verification checklist for this phase.

## Counts

| total official ops | implemented | fixture/tested | blocked/planned | excluded/not-applicable | certified |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 38 | 38 | 38 | 0 | 0 | 0 |

Breakdown: 11 stream-backed GET operations, 26 typed reverse-ETL write/channel-management operations, 1 typed direct-read freeBusy operation.

## Known limits

- `cdc=false`: Calendar API watch/channel setup/stop operations are executable typed reverse-ETL actions, but webhook delivery, renewal, and changefeed state consumption remain shared runtime work outside this connector bundle.
- Certification remains fixture-only/live-unperformed; no credentialed provider checks were run.

## Verification

All required local gates passed; see `VERIFICATION.md`.
