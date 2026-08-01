# Overview

Reads Intercom contacts, companies, conversations, admins, and tags through the Intercom REST API, and records the complete official Intercom OpenAPI 2.16 operation surface in connector-owned metadata.

Implemented fixture-backed ETL streams: `contacts`, `companies`, `conversations`, `admins`, `tags`.

The official inventory contains 231 operations: 55 ETL/read, 114 reverse-ETL write, 42 direct read/query/search, 7 binary/export, 12 CDC/changefeed-like, and 1 duplicate/not-applicable row. See `OFFICIAL_INVENTORY.md` and `api_surface.json` for the op-level ledger.

Service API documentation:

- https://developers.intercom.com/docs/references/2.16/rest-api/api.intercom.io
- https://developers.intercom.com/page-data/shared/oas-docs/references/%402.16/rest-api/api.intercom.io.yaml.json
- https://developers.intercom.com/page-data/docs/references/rest-api/api.intercom.io/data.json

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Intercom access token. Used only for Bearer auth; never logged.
- `api_version` (optional, string); optional `Intercom-Version` header value. When unset, the header is omitted and Intercom uses the account's default API version.
- `base_url` (optional, string); default `https://api.intercom.io`; format `uri`; Intercom API base URL override for tests or proxies.
- `page_size` (optional, string); default `50`; records per page (1-150) for the legacy-parity read streams.

Secret fields are redacted in logs and write previews: `access_token`.

Authentication behavior: Bearer token authentication using `secrets.access_token`. Connection checks call `GET /admins`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `starting_after`; next token from `pages.next.starting_after`.

- `contacts`: `GET /contacts` - records path `data`; query `per_page=50`; fixture-backed two-page cursor pagination.
- `companies`: `GET /companies` - records path `data`; query `per_page=50`; fixture-backed.
- `conversations`: `GET /conversations` - records path `data`; query `per_page=50`; fixture-backed.
- `admins`: `GET /admins` - records path `data`; fixture-backed.
- `tags`: `GET /tags` - records path `data`; fixture-backed.

Additional official read/detail/search/binary/changefeed operations are ledgered as typed blocked/planned operation rows rather than guessed as ETL streams without verified record shapes.

## Write actions & risks

`writes.json` declares 114 typed Intercom reverse-ETL write actions from the official OpenAPI mutation set, including `submit_fin_csat` for `POST /fin/csat`. They are provider-specific actions, not a generic HTTP write tool.

Safety requirements for every live write:

1. Create a reverse-ETL plan.
2. Preview the resolved action and records.
3. Provide the explicit approval token.
4. Execute only through the reverse-ETL runner.
5. For actions with `confirm: "destructive"`, provide the typed confirmation value `destructive` before any provider request is dispatched.

DELETE, redact, merge, detach/remove, archive/block, cancel, and similar destructive actions are included when represented by typed schemas and destructive confirmation. They are not blanket-excluded as unsafe.

Fixture-backed write request-shape evidence is connector-local and does not certify live Intercom behavior. Live certification remains `0` until a separately approved live-safe executor records redacted artifacts.

## Known limits

- No live Intercom credentials, provider calls, writes, binary downloads, or certification were used for this inventory.
- Direct read/query/search, binary/export, and CDC/changefeed-like operations are blocked by default in `api_surface.json`/`operations.json` until shared foundations and fixtures prove safe execution.
- CDC/changefeed truthfulness and state handling depend on #2986 and #2988 before this connector can claim CDC execution.
- The generated write schemas are simplified to the engine's supported draft-07 subset (`type`, `properties`, `required`, `items`, `enum`, `format`, `description`, `additionalProperties`). OpenAPI `oneOf`/`anyOf` request variants are merged into a typed property vocabulary without variant-specific validation.
- Certification metadata is fixture-only; `certification.json` intentionally has no live write pairings.
