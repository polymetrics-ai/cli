# Freshdesk connector

## Overview

Freshdesk REST API v2 connector for support tickets, contacts, companies, agents, groups, and the broader Freshdesk operation surface. The connector exposes the legacy five ETL streams plus fixed, connector-relative direct-read commands for JSON GET endpoints and named reverse-ETL write actions for JSON-expressible mutations.

## Auth setup

Create a Freshdesk API key in Freshdesk and store it as the `api_key` secret. Configure `base_url` as the full API base URL, for example `https://acme.freshdesk.com/api/v2`.

Service API documentation: https://developers.freshdesk.com/api/.

## Streams notes

Implemented streams:

- `tickets` (`GET /tickets`), incremental on `updated_at` via `updated_since` using `start_date`.
- `contacts` (`GET /contacts`).
- `companies` (`GET /companies`).
- `agents` (`GET /agents`).
- `groups` (`GET /groups`).

Additional GET endpoints are exposed as bounded direct reads under `pm freshdesk read ...`; they require fixed path/query flags, parse JSON only, and cap responses at 1 MiB.

## Write actions & risks

Freshdesk JSON-expressible POST/PUT/DELETE endpoints are exposed as named reverse-ETL actions under `pm freshdesk write ...`. These commands create a reverse-ETL command plan first; operators must preview and approve before execution. Delete/high-risk/admin actions declare `confirm: destructive` and require typed confirmation when executing the approved plan.

The connector intentionally does not expose a raw HTTP write surface. `POST /contacts/imports` remains blocked because Freshdesk requires a CSV multipart file upload; it needs a future bounded `file_upload` executor rather than JSON reverse ETL.

One advanced direct-read row, custom-object record filtering, remains blocked because Freshdesk encodes the custom field/operator as dynamic query parameter names (for example `age%5Bgte%5D=35`). That needs a typed query-builder policy rather than a raw query-string flag.

## Known limits

- No credentialed Freshdesk checks are run by static validation.
- Direct reads are JSON-only and capped at 1 MiB; binary/file downloads are not streamed to disk.
- Contact CSV import (`POST /contacts/imports`) is blocked until the connector engine has a bounded file-upload operation executor.
- Custom-object record filtering is blocked until direct reads support a typed dynamic-query builder for provider-specific filter syntax.
- `base_url` must include `/api/v2`; the engine does not derive it from a bare Freshdesk domain.
