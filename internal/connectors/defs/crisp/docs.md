# Overview

Crisp is an API connector for the official Crisp REST API v1. The connector records and covers the official non-HEAD surface: 220 operations with method split GET=91, POST=47, PATCH=44, PUT=12, DELETE=26.

GET operations are implemented as bounded JSON direct-read commands. POST/PUT/PATCH/DELETE operations are implemented as explicit named reverse-ETL write actions with fixed method/path metadata and approval gates. No raw generic HTTP writer, shell tool, SQL writer, or unbounded binary downloader is exposed.

Service API documentation: https://docs.crisp.chat/references/rest-api/v1/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.crisp.chat`.
- `identifier` (required, secret, string); Crisp REST API identifier for HTTP Basic authentication. Never logged.
- `key` (required, secret, string); Crisp REST API key for HTTP Basic authentication. Never logged.
- `website_id` (required, string); Crisp website identifier for website-scoped operations.
- `start_date` (optional, string); future incremental lower-bound timestamp.
- `page_size` and `max_pages` (optional, integer); future bounded read hints.

Add credentials from environment variables or stdin. Do not put secret values in prompt text, logs, docs, fixtures, shell history, or command-line flags.

Connection checks call GET `/v1/website/{{ config.website_id }}`.

## Streams notes

This all-ops slice implements GET operations as direct reads, not long-running ETL streams. Stream materialization remains a follow-up refinement for durable collection reads that need schemas, pagination state, cursors, primary keys, and sanitized fixtures.

The direct-read commands use fixed official Crisp endpoints, explicit path/query flags, and the provider-neutral `json_response` output policy capped by the direct-read byte limit.

## Write actions & risks

The connector declares 129 explicit reverse-ETL write actions for Crisp mutations. Each action has a fixed method/path, validates declared path fields, and runs only through the reverse-ETL lifecycle: plan, preview, approval token, and execute. Destructive/admin/file-style actions include destructive confirmation metadata where applicable.

For POST/PUT/PATCH actions, non-path record fields are sent as the JSON request body by the declarative engine. Later issue slices may tighten individual body schemas from provider body documentation, but this slice does not introduce any raw generic HTTP endpoint selector.

## Known limits

- No live Crisp API calls or credentialed checks were run while authoring this connector.
- GET direct reads are bounded JSON responses; binary/file transfer policy remains conservative and does not add an unbounded local file downloader.
- Durable ETL streams with cursor state and sanitized fixtures are still a future refinement on top of the direct-read coverage.
- Mutation body schemas are transport-safe scaffolds keyed by fixed actions; provider-specific semantic field constraints can be tightened incrementally without exposing a generic writer.
