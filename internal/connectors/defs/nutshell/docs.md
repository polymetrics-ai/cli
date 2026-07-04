# Overview

This bundle covers the documented Nutshell REST API surface from the current official developer reference at https://developers.nutshell.com/docs/getting-started. The original legacy-parity streams (accounts, contacts, leads, activities, and users) keep their existing projected schemas and record fields. Pass B adds documented JSON list/detail streams for related CRM resources and write actions for object-body or no-body mutations that the declarative engine can express.

## Auth setup

Provide a Nutshell account `username` and API token as the `password` secret. Both are sent with HTTP Basic auth. The token is never written to fixtures, docs, or logs.

## Streams notes

`accounts`, `contacts`, `leads`, `activities`, `invoices`, `notes`, `quotes`, and `tasks` use Nutshell's 0-indexed `page[page]` plus `page[limit]` pagination where documented or preserved from the legacy connector. Other reference/detail endpoints use a stream-level `pagination: {"type": "none"}` override.

The five legacy streams continue to project the exact fields emitted by `internal/connectors/nutshell`. Newly added streams use `projection: "passthrough"` with permissive schemas because the current Nutshell OpenAPI blocks are inconsistent for several list endpoints; fixtures preserve the documented envelope names or root-array shape used by each endpoint. Detail streams take optional `*_id` config values such as `account_id`, `lead_id`, and `task_id`.

## Write actions & risks

`writes.json` includes create/update/lifecycle/delete actions whose request can be represented as a JSON object or no body: account/contact/lead/activity/audience/custom-field/note/product-category/source/tag/task creation or updates, undelete/reopen/watch lifecycle calls, and documented DELETE calls. Delete actions are marked destructive and include 404 missing-ok semantics for idempotent replay behavior.

The write surface mutates live Nutshell CRM data and should be used only through the normal reverse-ETL plan, preview, approval, execute flow.

## Known limits

- JSON Patch endpoints and the lead-installments endpoint require root JSON array request bodies. The current write dialect builds object bodies from records, so these are excluded in `api_surface.json` and recorded as a typed `ENGINE_GAP` in `docs/migration/quarantine.json`.
- `GET /stagesets/{id}/export` returns CSV rather than JSON records and is excluded as `binary_payload`.
- The old docs URL `https://app.nutshell.com/rest/` now returns 404; the API base URL remains `https://app.nutshell.com/rest`, while metadata points at the current official developer docs.
