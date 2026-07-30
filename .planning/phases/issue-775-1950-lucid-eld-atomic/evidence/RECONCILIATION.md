# Lucid ELD Source Reconciliation — Issue #1950

## Public source status

- Official OpenAPI JSON (`https://api.drivehos.app/partner/swagger/doc.json`) fetched unauthenticated on 2026-07-30: HTTP 200, SHA-256 `1b3756f4c69c9133e24754a856d2fe9ec2b08768edd5dec25b899f564ddb7ec4`, 8 HTTP operations.
- Official Swagger UI (`https://api.drivehos.app/partner/swagger`) fetched unauthenticated on 2026-07-30: HTTP 200, SHA-256 `f5f4f0a04d06a641aaafd89d531bbaf22c7ae91c338e6c611cfc6cec69bfe060`; page is Swagger UI shell loading the OpenAPI document.
- WithTerminal provider page (`https://docs.withterminal.com/providers/tsp/lucid-eld`, markdown alternate `/providers/tsp/lucid-eld.md`) fetched unauthenticated on 2026-07-30: provider code `lucid-eld`, available history 180 days, vehicle-location sample rate 60 seconds, crash reports not supported.

## Reconciliation rule

OpenAPI is wire-schema ground truth. WithTerminal is capability baseline only; it confirms supported model families (Driver, Vehicle, Latest Vehicle Location, Vehicle Location) but does not add Lucid wire paths beyond the official DriveHOS Swagger document.

## Official OpenAPI operation inventory

| Method | Path | OpenAPI summary | Parameters | Classification target |
|---|---|---|---|---|
| GET | `/v2/company-info` | Get company info | required headers `X-API-Provider-Key`, `X-API-Company-Key` | direct read `company info get` |
| GET | `/v2/driver/{driver_id}` | Get driver by ID | required headers, path `driver_id` | direct read `drivers get` |
| GET | `/v2/drivers` | Get drivers list | required headers, query `limit` default 100 max 1000, query `page` default 1 | stream `drivers` |
| GET | `/v2/latest-driver-status` | Get drivers last status | required headers, optional query `driver_id`, `limit` default 100 max 1000, `page` | direct read `latest driver statuses list` |
| GET | `/v2/latest-vehicle-status` | Get last vehicle status | required headers, optional query `vehicle_id`, `limit` default 100 max 200, `page` | direct read `latest vehicle statuses list` |
| GET | `/v2/vehicle-location-history/{vehicle_id}` | Get vehicle location history | required headers, path `vehicle_id`, required query `start_date`/`end_date` in `MM-DD-YYYY`, optional `next_page_token`, `limit` default 100 max 1000 | stream `vehicle_location_history` |
| GET | `/v2/vehicle/{vehicle_id}` | Get vehicle by ID | required headers, path `vehicle_id` | direct read `vehicles get` |
| GET | `/v2/vehicles` | Get list of vehicles | required headers, optional query `status` active/inactive, `limit` default 100 max 1000, `page` | stream `vehicles` |

## Pagination and auth notes for later lanes

- List/latest endpoints use page-number pagination (`page`, `limit`) except vehicle-location history, which uses body field `next_page_token` echoed as query `next_page_token`.
- Response envelope is `handlers.ResponseV2` with `data`, `description`, `next_page_token`, `page`, `size`, `status_code`, `total_elements`, and `total_pages`.
- Auth is represented as required headers in operation parameters, not OpenAPI `securityDefinitions`: `X-API-Provider-Key` and `X-API-Company-Key`. Both are credential-shaped and must be `x-secret` in #1951.
- No official OpenAPI POST/PUT/PATCH/DELETE operations, reports, webhook registration endpoints, binary/media endpoints, or prose-only extra Lucid endpoints were found in the public Swagger UI/OpenAPI pair. WithTerminal lists Camera Media and crash reports as not supported for Lucid ELD.

## Deferred implementation lanes

- #1951: connection spec/auth config and CLI metadata should consume the header names and date/limit/page constraints above.
- #1952: streams `drivers`, `vehicles`, `vehicle_location_history`.
- #1953: direct reads `company info get`, `drivers get`, `vehicles get`, `latest driver statuses list`, `latest vehicle statuses list`.
- #1954: no writes discovered; issue should remain no-op unless new official mutation evidence appears.
- #1955: docs/certification should cite this inventory and the public source hashes.
