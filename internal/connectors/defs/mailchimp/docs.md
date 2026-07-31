# Overview

Mailchimp uses the official Marketing API Swagger root and all provider-owned path refs from https://api.mailchimp.com/schema/3.0/Swagger.json. This bundle models 298 current official operations: 79 ETL streams, 68 typed direct reads/search commands, 148 named reverse-ETL write actions, and 3 blocked operation-ledger rows.

The connector covers audiences/lists and members, campaigns, automations and customer journeys, reports and reporting resources, templates, file-manager metadata, batches, batch webhooks, ecommerce resources, landing pages, connected sites, verified domains, conversations, account exports, authorized apps, Facebook ads, and typed provider search.

## Auth setup

Connection fields:

- `data_center` (required): Mailchimp datacenter token such as `us6`; builds `https://<data_center>.api.mailchimp.com/3.0`.
- `access_token` (optional secret): OAuth bearer token; preferred when present.
- `api_key` (optional secret): API key used as HTTP Basic password when no bearer token is present.
- `start_date` (optional): RFC3339 lower bound used by supported top-level incremental streams.
- Optional identifier config such as `list_id`, `campaign_id`, `workflow_id`, `subscriber_hash`, `template_id`, `store_id`, and related path variables is used by nested streams and direct-read commands when a command flag does not supply the value.

Secrets must be provided from environment variables or stdin through `pm credentials add`; never paste secret values into chat, docs, logs, or shell history.

## Streams notes

The base reader uses bounded offset/count pagination with `count=100` and `max_pages=5` for fixture-safe and operator-bounded reads. Top-level `lists`, `campaigns`, `reports`, and `automations` retain incremental lower-bound parameters when a cursor or `start_date` is available. Nested streams are explicit ETL streams with schema projection and sanitized fixtures; they require the relevant identifier config when the stream path contains provider IDs.

Executable stream count: 79. Every executable stream has a sanitized fixture page; `lists` has a two-page fixture to prove offset pagination terminates.

## Write actions & risks

This bundle declares 148 named reverse-ETL actions. Each action has a closed `record_schema`, path fields for provider identifiers, risk text, redaction for sensitive identifier/content fields, and fixture-backed request-shape coverage. Destructive or externally irreversible actions such as deletes, archives, sends, publishes, pauses/starts, and triggers declare `confirm: "destructive"`. DELETE actions declare provider-idempotent 404 handling where supported by the HTTP delete semantics.

Reverse ETL remains: plan -> preview -> explicit approval -> execute. No action exposes a raw method/path/body escape hatch.

## Known limits

- `POST /batches` remains blocked because the official body is an arbitrary batch of method/path/body operations; exposing it would be a generic HTTP write passthrough forbidden by repo policy.
- `GET /` and `GET /ping` remain blocked as local metadata/health workflows; `pm connectors inspect` and `pm etl check` are the typed local surfaces.
- Fixture-only validation does not certify live provider behavior. Certified/live status remains `0` until a separately approved live executor runs with redacted artifacts.
- Nested direct reads rely on typed flags or matching config values for path variables; no generic raw API command is exposed.
