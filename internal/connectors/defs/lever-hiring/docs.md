# Overview

Reads Lever Hiring opportunities, postings, users, requisitions, stages, and related hiring records through the Lever Data API. This bundle also exposes fixed-target bounded direct reads and selected typed reverse-ETL write plans.

Service API documentation: https://hire.lever.co/developer/documentation.

Current official operation ledger: 106 HTTP operations (55 GET, 26 POST, 14 PUT, 11 DELETE). Implemented rows: 60 = 25 stream-backed reads + 21 bounded direct reads + 14 typed writes. Blocked/planned rows: 46. Validated rows: 0 (fixture-only; no live provider calls were made).

Lever additionally documents 10 webhook trigger/event names. Those are provider-pushed event payloads, not HTTP operations the connector can call, so they are tracked separately and are not part of the 106 counted above.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string): Lever OAuth2 access token. Bearer authentication is preferred when both credential types are present.
- `api_key` (optional, secret, string): Lever API key sent as HTTP Basic username with a blank password.
- `base_url` (optional, string): default `https://api.lever.co/v1`; use `https://api.sandbox.lever.co/v1` or a test proxy only via config, never by editing operation paths.
- `mode` (optional, string): retained for credential-free fixture/execution-contract compatibility.

Secret fields are redacted: `access_token`, `api_key`. Connection checks call `GET /postings?limit=1`.

## Streams notes

The connector declares 25 stream-backed read surfaces. Top-level collection streams use Lever cursor pagination (`limit`, `offset`, `next`, `hasNext`). Opportunity/posting sub-resource streams use the engine fan-out contract to list parent IDs from `/opportunities` or `/postings`, then request the fixed child collection path once per parent ID and stamp the parent ID onto emitted records.

Implemented stream names: `opportunities`, `postings`, `users`, `requisitions`, `stages`, `deleted_applications`, `archive_reasons`, `audit_events`, `disposition_stages`, `feedback_templates`, `deleted_opportunities`, `deleted_postings`, `form_templates`, `requisition_fields`, `sources`, `applications`, `opportunity_feedback`, `opportunity_interviews`, `opportunity_notes`, `opportunity_offers`, `opportunity_file_actions`, `opportunity_panels`, `opportunity_forms`, `opportunity_referrals`, `posting_users`.

Scalar-list and binary/file-family operations that cannot be represented as durable object records are kept out of ETL streams and are either bounded direct reads or blocked in the operation ledger.

## Write actions & risks

The connector declares 14 typed write actions: `update_feedback`, `delete_feedback`, `create_feedback_template`, `update_feedback_template`, `delete_feedback_template`, `create_form_template`, `update_form_template`, `delete_form_template`, `delete_note`, `delete_requisition`, `delete_requisition_field_options`, `delete_requisition_field`, `deactivate_user`, `reactivate_user`.

Writes are only available through reverse ETL plan -> preview -> explicit approval -> execute. Destructive/no-body actions use `confirm: destructive`. The bundle does not expose arbitrary request bodies, raw query strings, generic method/path/body, file bytes, shell commands, or passthrough HTTP tools.

Many official Lever mutations remain blocked because the current generic write executor does not support provider query parameters such as `perform_as`, or because the official request body is not closed enough to expose without a connector-specific schema/hook.

## Known limits

- Fixture-only evidence: no live Lever credentials, provider calls, provider writes, or validation run were used.
- Webhook trigger ingestion and webhook subscription lifecycle rows are blocked pending the shared webhook/CDC foundation (#2986/#2988).
- Lever file/resume/upload/download rows are blocked pending a bounded binary/multipart transfer executor; no generic file byte passthrough is exposed.
- Mutations requiring documented query/form parameters (for example `perform_as`) are blocked until shared write-query support or a connector hook can express them without a raw query escape hatch.
- The generated operation ledger is based on the current official Lever Developer documentation fetched for this slice and should be re-audited when Lever changes that page.
