# Overview

Zoom connector parity for the official Zoom Meeting API reference page. The source inventory is the packed OpenAPI 3.0 document embedded in Zoom Developer Docs at `https://developers.zoom.us/docs/api/rest/reference/zoom-api/methods/`.

This bundle covers all 184 official operations exactly once:

| lane | official count | executable surface |
| --- | ---: | --- |
| ETL/read streams | 48 | 48 declared streams with sanitized fixture replay |
| Direct/provider reads | 29 | fixed `rest_read` command operations with bounded redacted JSON output |
| Binary/file lane | 29 | fixed binary/file-lane reads or typed write actions; no raw file passthrough |
| Reverse ETL writes | 78 | typed write actions; plus 13 mutating binary/file-lane operations implemented as typed writes |
| CDC/changefeed | 0 | no Zoom changefeed operation in the landed audit |
| Excluded/not-applicable | 0 | no exclusions in the landed audit |

## Auth setup

Use a Zoom OAuth access token saved through Polymetrics credentials. The `access_token` field is marked `x-secret` and is sent only as `Authorization: Bearer ...`; never pass it as a CLI flag or commit it to fixtures.

Configuration fields:

- `base_url` defaults to `https://api.zoom.us/v2`.
- `page_size` defaults to `100` for paginated stream reads.
- Path-scoped operations use typed config or command flags for documented Zoom path parameters such as `userId`, `meetingId`, `webinarId`, `deviceId`, and related IDs.

Connection checks call `GET /users?page_size=1` against the configured base URL.

## Streams notes

The connector declares 48 Zoom streams. Paginated list streams use Zoom's `next_page_token` cursor convention where documented and replay fixtures prove the paginator terminates. Detail-style GET operations assigned to ETL are represented as single-object streams with explicit path-parameter config rather than provider-search passthrough.

All stream fixtures are synthetic (`example.invalid`, fixture IDs, and non-secret timestamps). They are recorded-real-shape enough for engine pagination/projection/conformance without containing live Zoom data.

## Write actions & risks

The connector declares 91 typed write actions: the landed audit's 78 reverse-ETL write operations plus 13 mutating binary/file-lane operations that still change provider state. Every write action has a closed record schema, fixed method/path, risk text, and command metadata. DELETE actions use `confirm: destructive`, redact path identifiers in previews/errors, and treat `404` as an idempotent already-absent cleanup response.

Reverse ETL execution remains plan → preview → explicit approval → execute. This bundle does not expose a raw method/path/body command, arbitrary GraphQL, shell, SQL, generic HTTP, local file passthrough, or unrestricted binary escape hatch.

## Known limits

- Fixture replay is not live certification. `certified=0` until a separately approved live-safe certification run records redacted provider evidence.
- The existing direct-read executor supports bounded JSON `rest_read` operations. File-transfer-style binary downloads remain represented with fixed metadata and safety notes; this slice does not edit shared binary execution runtime.
- The official source fetched during implementation has the same 184-operation total as the landed r2 audit but a newer page ETag/build id: ETag `"08322c4f0fa086914cd5d144268b61bf"`, Last-Modified `Fri, 31 Jul 2026 17:15:56 GMT`, Next build id `2026-07-31T11-05-34-06-00`, extracted OpenAPI SHA256 `7490bae1a0815d82721af85b894e90f5662489c0aacaa2300277758562213ab9`.
