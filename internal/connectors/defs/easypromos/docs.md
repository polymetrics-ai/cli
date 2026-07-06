# Overview

Reads Easypromos promotions, organizing brands, stages, users, participations, and prizes through
the Easypromos REST API.

Readable streams: `promotions`, `organizing_brands`, `stages`, `users`, `participations`, `prizes`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.easypromos.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.easypromosapp.com/v2`; format `uri`;
  Easypromos API base URL override for tests or proxies.
- `bearer_token` (required, secret, string); Easypromos API bearer token, sent as Authorization:
  Bearer <bearer_token>. Never logged.
- `mode` (optional, string).
- `promotion_id` (optional, string); Promotion id the per-promotion streams (stages, users,
  participations, prizes) are scoped to. Required for those streams; substituted into their
  /{resource}/{promotion_id} path.

Secret fields are redacted in logs and write previews: `bearer_token`.

Default configuration values: `base_url=https://api.easypromosapp.com/v2`.

Authentication behavior:

- Bearer token authentication using `secrets.bearer_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/promotions`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `next_cursor`; next token from
`paging.next_cursor`.

- `promotions`: GET `/promotions` - records path `items`; cursor pagination; cursor parameter
  `next_cursor`; next token from `paging.next_cursor`; computed output fields `created`,
  `default_language`, `description`, `end_date`, `id`, `organizing_brand_id`,
  `organizing_brand_name`, `promotion_type`, `start_date`, `status`, `timezone`, `title`, `url`.
- `organizing_brands`: GET `/organizing_brands` - records path `items`; cursor pagination; cursor
  parameter `next_cursor`; next token from `paging.next_cursor`.
- `stages`: GET `/stages/{{ config.promotion_id }}` - records path `items`; cursor pagination;
  cursor parameter `next_cursor`; next token from `paging.next_cursor`.
- `users`: GET `/users/{{ config.promotion_id }}` - records path `items`; cursor pagination; cursor
  parameter `next_cursor`; next token from `paging.next_cursor`.
- `participations`: GET `/participations/{{ config.promotion_id }}` - records path `items`; cursor
  pagination; cursor parameter `next_cursor`; next token from `paging.next_cursor`.
- `prizes`: GET `/prizes/{{ config.promotion_id }}` - records path `items`; cursor pagination;
  cursor parameter `next_cursor`; next token from `paging.next_cursor`; computed output fields
  `code`, `created`, `download_url`, `id`, `participation_id`, `prize_type_id`, `prize_type_name`,
  `redeem_url`, `stage_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Easypromos API read of promotion, user,
participation, and prize data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 6 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
