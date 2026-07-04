# Outreach

## Overview

Reads and writes Outreach REST API v2 resources through `https://api.outreach.io/api/v2`. Outreach is a JSON:API API; this bundle now covers the full documented standard-resource surface from the official OpenAPI document plus the documented custom-object list/detail/create/update/delete surface. The four legacy streams (`prospects`, `accounts`, `sequences`, `mailings`) keep their original schema projection and computed fields so legacy emitted records remain unchanged.

## Auth setup

Provide an Outreach OAuth2 **access token** (`secrets.access_token`); it is sent as `Authorization: Bearer <access_token>` on every request and is never logged. Legacy performed a refresh-token grant outside the connector; this declarative bundle still consumes only the already-issued access token because the engine has no refresh-token-grant auth mode. Requests send `Accept` and `Content-Type` as `application/vnd.api+json`, matching Outreach's JSON:API documentation.

## Streams notes

There are 96 streams. Standard collection endpoints read `data[]`, send `page[size]=100`, and follow Outreach `links.next` with the engine's same-host `next_url` guard. Standard detail endpoints read the single `data` object and use optional config IDs such as `account_id` or `task_id`. New Pass B streams use `projection: passthrough` so the full JSON:API resource envelope (`id`, `type`, `attributes`, `relationships`, `links`) is emitted without inventing field mappings from hundreds of vendor attributes. The legacy four streams remain schema-projected to `id`, `type`, `email`, `name`, `created_at`, and `updated_at`; `name` uses the same `attributes.name` then `attributes.displayName` fallback as legacy.

Custom-object streams are generic: set `custom_object_name` for `/customObjects/{objectName}` and `custom_object_record_id` for its detail endpoint. The `schema_definitions` stream reads `/schema` as a keyed object so callers can inspect standard and authenticated-tenant custom object definitions.

## Write actions & risks

There are 163 write actions covering documented JSON:API POST/PATCH/DELETE endpoints, including resource CRUD, batch actions, task actions, sequence/sequence-template activation, mailbox actions, imports, webhooks, and custom objects. JSON:API writes accept a `data` object and path/query identifiers as top-level record fields; path/query fields are excluded from the request body. Delete and bulk-destroy style actions are marked destructive and require approval.

## Known limits

- The legacy `page_size`/`max_pages` config surface is not carried over for `next_url` pagination; the bundle sends the legacy default `page[size]=100` and the engine follows `links.next` until Outreach returns an empty/missing next link. These runtime knobs are intentionally not declared in `spec.json` because the current pagination dialect cannot consume them.
- Custom objects are tenant-defined. The generic custom-object streams and writes require the caller to provide the internal object name; the bundle cannot enumerate tenant-specific object types statically. Use `schema_definitions` to read the schema endpoint for available object definitions.
- Bulk action endpoints expose only documented required `actionParams` query fields in write paths. Optional deep-object query flags such as `filter` and `skipConfirmation` are not modeled because the write dialect has no optional query-parameter omission primitive for action-specific query strings.
- Fixture pagination remains single-page for `next_url` streams, using the sanctioned `links.next: ""` terminator fixture pattern.
