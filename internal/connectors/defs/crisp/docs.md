# Overview

Crisp is an API connector scaffold for the official Crisp REST API v1. #205 records the official non-HEAD surface as metadata only: 220 operations with method split GET=91, POST=47, PATCH=44, PUT=12, DELETE=26.

No executable ETL stream, direct-read command, binary transfer, or reverse-ETL write action is exposed by this scaffold. Follow-up #204 subissues replace blocked ledger rows with explicit `covered_by` stream, direct-read, write, or binary policy entries.

Service API documentation: https://docs.crisp.chat/references/rest-api/v1/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.crisp.chat`.
- `identifier` (required, secret, string); Crisp REST API identifier for HTTP Basic authentication. Never logged.
- `key` (required, secret, string); Crisp REST API key for HTTP Basic authentication. Never logged.
- `website_id` (required, string); Crisp website identifier for website-scoped operations.
- `start_date` (optional, string); future incremental lower-bound timestamp.
- `page_size` and `max_pages` (optional, string); future bounded read hints.

Add credentials from environment variables or stdin. Do not put secret values in prompt text, logs, docs, fixtures, shell history, or command-line flags.

Connection checks call GET `/v1/website/{ config.website_id }` when check execution is used in a later slice.

## Streams notes

#205 declares no streams. Candidate ETL streams are classified in #207 after the operation ledger is reviewed for durable collection reads, pagination, cursor fields, primary keys, schemas, and sanitized fixtures.

The initial API ledger keeps every official operation blocked by default so the connector cannot overclaim read coverage before schemas and fixture-backed validation exist.

## Write actions & risks

#205 declares no write actions. Crisp mutations include conversation, people, campaign, helpdesk, plugin, subscription, website administration, import/export, and destructive delete-style operations.

Future reverse-ETL actions must be named actions with explicit record schemas, path fields, redaction, risk text, plan preview, approval token, and typed confirmation for destructive/admin operations. No raw generic HTTP write or direct write escape hatch is allowed.

## Known limits

- This is a metadata-only scaffold for issue #205; no ETL/read/write behavior is implemented yet.
- `api_surface.json` uses `operation_ledger_version: 1` with blocked metadata rows until #207, #208, #209, #210, and #211 implement or finalize classifications.
- CLI surface entries are planned help/discovery metadata only. Runtime help/manual/website parity is completed in #206.
- No live Crisp API calls or credentialed checks were run for this scaffold.
