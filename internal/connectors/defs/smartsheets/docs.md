# Overview

Smartsheets reads the Smartsheet REST API v2 under `https://api.smartsheet.com/2.0`. The bundle keeps the legacy `sheets` and hook-backed `sheet_rows` streams and expands the documented API surface with streams for contacts, events, favorites, folders, groups, reports, shares, sheets, attachments, columns, comments, discussions, proofs, update requests, dashboards, users, webhooks, and workspaces.

The legacy Go connector remains read-only until the wave6 registry flip. Pass B adds declarative write actions for documented JSON mutations that the engine can express without query parameters, binary bodies, multipart payloads, or credential lifecycle hooks.

## Auth setup

Provide a Smartsheet API access token via the `access_token` secret. It is sent as `Authorization: Bearer <access_token>`, matching the legacy connector. `base_url` defaults to `https://api.smartsheet.com/2.0`.

Detail and nested streams use explicit config IDs such as `sheet_id`, `row_id`, `report_id`, `workspace_id`, and related path identifiers. `spreadsheet_id` is retained for the legacy-parity `sheet_rows` stream because the existing connector used that config key.

## Streams notes

`sheets` remains the first stream and preserves legacy passthrough behavior. `sheet_rows` remains StreamHook-handled because the legacy row shape flattens cells into dynamic field names using the sibling `columns[]` array in the same page body; the declarative dialect cannot use runtime data as output field names.

New streams are raw Smartsheet API objects with schema/passthrough-safe shapes. Ordinary index endpoints use Smartsheet `page`/`pageSize` pagination. `lastKey` endpoints use cursor pagination with `lastKey`. `events` uses `streamPosition`/`nextStreamPosition` and stops when `moreAvailable` is not true. Single-object detail/settings/path endpoints use one request with `pagination.type: none`.

Search endpoints are not streams because the OpenAPI requires a caller-supplied free-text `query`; they are ad hoc lookups, not fixed syncable list/detail resources.

## Write actions & risks

Write actions require the standard reverse-ETL plan, preview, approval, execute flow. Covered actions create, update, copy, move, publish, and delete supported Smartsheet objects including favorites, folders, groups, reports, sheets, URL attachments, automation rules, columns, comments, discussions, cross-sheet references, proofs, rows, summary fields, update requests, dashboards, alternate emails, webhooks, and workspaces.

Deletes are marked destructive and allow 404 as an idempotent missing result. Required path fields are removed from the JSON body; remaining record fields become the request body. Actions with no documented JSON body use `body_type: none`.

## Known limits

- Binary, octet-stream, and multipart upload endpoints are excluded because `writes.json` only supports JSON, form, and empty bodies.
- Writes that require query parameters, such as share mutations and row/summary-field bulk deletes, are excluded because write actions have no query parameter field.
- OAuth token issuance/revocation and notification-send endpoints are excluded as non-data side effects.
- Organization user lifecycle and whole-container destructive deletes are excluded as elevated-scope or destructive-admin operations.
- `sheet_rows` remains hook-backed for dynamic cell flattening; the hook is the documented Tier-2 escape hatch for this legacy record shape.
