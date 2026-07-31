# Overview

Front is modeled as a definition-owned connector bundle for the documented Front Platform API sources listed in `https://dev.frontapp.com/llms.txt` and the linked `reference/*.md` pages. This worker parsed the embedded OpenAPI snippets from those reference pages and reconciled 255 documented operation pages (254 unique method/path pairs; `PATCH /channels/{channel_id}` is documented twice and one row is retained as a duplicate ledger entry for count reconciliation).

Implemented read streams cover 115 fixed Front GET/changefeed surfaces. Implemented reverse ETL metadata covers 118 documented Core REST write operations plus 1 application event trigger surface. Direct/provider search metadata covers 5 bounded JSON operations. Binary attachment downloads (5) and application-channel/voice/plugin-adjacent operations (11) are represented as planned/blocked or not-applicable fixed operations without raw passthrough.

Service API documentation: https://dev.frontapp.com/llms.txt.

## Auth setup

Connection fields:

- `api_key` (required, secret, string): Front API token, sent as `Authorization: Bearer <api_key>`. Never log or store token values in fixtures.
- `base_url` (optional, string): default `https://api2.frontapp.com`.
- `page_limit` (optional, string): default `50`; sent as `limit` on paginated list streams.
- Optional path-parameter config fields such as `account_id`, `conversation_id`, `inbox_id`, and `team_id` are used only by parameterized streams. A parameterized stream fails closed if its required config value is absent.

Connection checks call `GET /inboxes` with bearer authentication.

## Streams notes

The bundle declares 115 Front read streams. List streams read records from `_results` and follow `_pagination.next`; single-resource streams use a single-object projection. Projection is `passthrough` for generated full-surface streams so the connector does not silently drop Front fields before a future schema-tightening pass.

The legacy fixture-backed streams remain available with their original names: `contacts`, `conversations`, `inboxes`, `tags`, `teammates`, and `channels`. Event/changefeed GET surfaces are represented as read streams; the mutating application event trigger is a fixed reverse-ETL action rather than a CDC reader claim. `metadata.capabilities.cdc` remains false because there is no streaming CDC implementation.

## Write actions & risks

`writes.json` declares fixed Front write actions only; there is no generic method/path/body escape hatch. Every action has a record schema, path field list where applicable, risk text, and `confirm: "destructive"` so execution stays behind the existing reverse ETL plan → preview → explicit approval → execute path plus typed destructive confirmation. DELETE actions declare idempotent 404 handling through `missing_ok_status: [404]`.

Direct/provider-search commands for analytics exports/reports and conversation search are fixed, bounded, JSON-redacted direct reads. Binary attachment download operations are recorded in `operations.json` with max-byte bounds but remain planned because this connector-local worker cannot add the shared binary file-output execution foundation.

## Known limits

- No live Front credentials or provider calls were used; this bundle is fixture/conformance backed and remains uncertified for live behavior.
- The generated stream schemas are intentionally permissive (`projection: passthrough`, `additionalProperties: true`) pending a future field-tightening pass; they preserve records rather than dropping undocumented fields.
- Parameterized streams require the corresponding config values and are not safe as an all-stream default sync without operator selection.
- Binary downloads are typed and bounded in metadata but not claimed executable until a shared binary-download execution contract is available.
- Application Channel, Voice Channel, and Plugin SDK-adjacent operations are present in the ledger as blocked/not-applicable rows so official counts are preserved without exposing unsupported writes.
- The duplicate `PATCH /channels/{channel_id}` docs row is recorded as a duplicate operation row; the executable Core API write action is declared once.
