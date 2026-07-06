# Overview

Reads Trello boards, lists, and checklists through the Trello REST API. Cards and actions are
blocked (see docs.md Known limits).

Readable streams: `boards`, `lists`, `checklists`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.atlassian.com/cloud/trello/rest/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.trello.com/1`; format `uri`; Trello API base
  URL override for tests or proxies.
- `board_ids` (optional, string); Comma-separated list of Trello board IDs to scope the
  lists/checklists streams to. Required for those two streams (no auto-discovery fallback here; see
  docs.md Known limits) - the boards stream itself is unaffected and always lists every board the
  authenticated member can access.
- `key` (required, secret, string); Trello API key, sent as the 'key' query parameter on every
  request. Never logged.
- `token` (required, secret, string); Trello API token, sent as the 'token' query parameter on every
  request. Never logged.

Secret fields are redacted in logs and write previews: `key`, `token`.

Default configuration values: `base_url=https://api.trello.com/1`.

Authentication behavior:

- API key authentication in query parameter `key` using `secrets.key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/members/me` with query `fields`=`id`; `token`=`{{ secrets.token }}`.

## Streams notes

Default pagination: single request; no pagination.

- `boards`: GET `/members/me/boards` - records at response root; query `token`=`{{ secrets.token
  }}`.
- `lists`: GET `/boards/{{ fanout.id }}/lists` - records at response root; query `token`=`{{
  secrets.token }}`; fan-out; ids from config field `board_ids`; id inserted into the request path;
  stamps `idBoard`.
- `checklists`: GET `/boards/{{ fanout.id }}/checklists` - records at response root; query
  `token`=`{{ secrets.token }}`; fan-out; ids from config field `board_ids`; id inserted into the
  request path; stamps `idBoard`.

## Write actions & risks

This connector is read-only. Read behavior: external Trello API read of board/list/checklist data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=3.
