# Overview

Reads Huntr organization members, candidates, activities, notes, and actions through the Huntr REST
API.

Readable streams: `members`, `candidates`, `activities`, `notes`, `actions`.

This connector is read-only; no write actions are declared.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Huntr organization API key, sent as a Bearer token
  (Authorization: Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api.huntr.co/org`; format `uri`; Huntr
  organization API base URL override for tests or proxies.
- `page_size` (optional, integer); default `100`; Page size for the limit query parameter (1-100).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.huntr.co/org`, `page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/members` with query `limit`=`1`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `next`; next token from `next`.

- `members`: GET `/members` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `next`; next token from `next`.
- `candidates`: GET `/candidates` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `next`; next token from `next`.
- `activities`: GET `/activities` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `next`; next token from `next`.
- `notes`: GET `/notes/members` - records path `data`; query `limit`=`100`; cursor pagination;
  cursor parameter `next`; next token from `next`.
- `actions`: GET `/actions` - records path `data`; query `limit`=`100`; cursor pagination; cursor
  parameter `next`; next token from `next`.

## Write actions & risks

This connector is read-only. Read behavior: external Huntr organization API read of member,
candidate, activity, note, and action data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
