# Official Source Evidence — Lucid ELD #1950

Fetched at: 2026-07-30T01:56:20.128346+00:00

## Sources
- openapi: https://api.drivehos.app/partner/swagger/doc.json status=200 content_type=application/json bytes=19882 sha256=1b3756f4c69c9133e24754a856d2fe9ec2b08768edd5dec25b899f564ddb7ec4
- swagger_ui: https://api.drivehos.app/partner/swagger status=200 content_type=text/html bytes=3609 sha256=f5f4f0a04d06a641aaafd89d531bbaf22c7ae91c338e6c611cfc6cec69bfe060
- withterminal: https://docs.withterminal.com/providers/tsp/lucid-eld status=200 content_type=text/html; charset=utf-8 bytes=424597 sha256=bb31d3cf65ae2fe7bf33a3b7591f6f258a56a487cf81ac1d47f44f8ad07ec5af
- withterminal_markdown: https://docs.withterminal.com/providers/tsp/lucid-eld.md status=200 content_type=text/markdown; charset=utf-8 bytes=13495 sha256=df18b088f94701c22763e354b624cb45aec46327d0100a2448ae385e611b9821

## OpenAPI inventory

- Swagger version: 2.0
- host:
- basePath: ''
- schemes: []
- securityDefinitions: []
- global security: []
- definitions: handlers.ResponseV2
- operations: 8

- GET /v2/company-info — Get company info — params: X-API-Provider-Key:header:string, X-API-Company-Key:header:string; responses: 200, 401, 500
- GET /v2/driver/{driver_id} — Get driver by ID — params: X-API-Provider-Key:header:string, X-API-Company-Key:header:string, driver_id:path:string; responses: 200, 401, 500
- GET /v2/drivers — Get drivers list — params: X-API-Provider-Key:header:string, X-API-Company-Key:header:string, limit:query:integer, page:query:integer; responses: 200, 500
- GET /v2/latest-driver-status — Get drivers last status — params: X-API-Provider-Key:header:string, X-API-Company-Key:header:string, driver_id:query:string, limit:query:integer, page:query:integer; responses: 200, 401, 500
- GET /v2/latest-vehicle-status — Get last vehicle status — params: X-API-Provider-Key:header:string, X-API-Company-Key:header:string, vehicle_id:query:string, limit:query:integer, page:query:integer; responses: 200, 400, 401, 500
- GET /v2/vehicle-location-history/{vehicle_id} — Get vehicle location history — params: X-API-Provider-Key:header:string, X-API-Company-Key:header:string, vehicle_id:path:string, start_date:query:string, end_date:query:string, next_page_token:query:string, limit:query:integer; responses: 200, 400, 401, 500
- GET /v2/vehicle/{vehicle_id} — Get vehicle by ID — params: X-API-Provider-Key:header:string, X-API-Company-Key:header:string, vehicle_id:path:string; responses: 200, 401, 500
- GET /v2/vehicles — Get list of vehicles — params: X-API-Provider-Key:header:string, X-API-Company-Key:header:string, status:query:string, limit:query:integer, page:query:integer; responses: 200, 401, 500
