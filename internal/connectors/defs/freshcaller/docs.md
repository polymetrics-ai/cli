# Overview

Freshcaller is a declarative-HTTP bundle migrated from `internal/connectors/freshcaller` (the
hand-written legacy connector, which stays registered and unchanged until wave6's registry flip).
It reads Freshcaller calls, agents, teams, and phone numbers through the Freshcaller REST API v1.
Legacy is a plain connsdk-based HTTP connector (single token-header auth, page-number pagination,
no writes, no auth/stream hooks needed) — a pure Tier-1 declarative bundle fully expresses it; no
Tier-2 hook or Tier-3 native package is warranted. It is read-only in both legacy and this bundle
(`capabilities.write: false`, no `writes.json`).

## Auth setup

Legacy authenticates with a single API key, sent as the `Authorization` header in the form
`Token token=<api_key>` (`connsdk.APIKeyHeader("Authorization", token, "Token token=")`,
`freshcaller.go:172`). This bundle's `base.auth` reproduces the identical header shape via
`api_key_header` mode with `prefix: "Token token="`:

```json
{ "mode": "api_key_header", "header": "Authorization", "value": "{{ secrets.api_key }}", "prefix": "Token token=" }
```

`base_url` defaults to `https://api.freshcaller.com/api/v1`, matching legacy's `defaultBaseURL`
constant exactly.

## Streams notes

All 4 streams (`calls`, `agents`, `teams`, `numbers`) share the identical shape: a flat
`GET <resource>` request, `records.path` naming the resource's own top-level array key
(matching legacy's `streamEndpoints[...].recordsPath`, which is always identical to the resource
name), and `page`/`page_size` page-number pagination with a short-page stop — matching legacy's
`harvest`, which advances `page` until a page returns fewer than `pageSize` records. `page_size`
defaults to 100 and is bounded 1-100 in `spec.json`, matching legacy's
`defaultPageSize`/`maxPageSize` constants (`freshcaller.go:21-22`); `max_pages` defaults to
unbounded (`0`), matching legacy's `maxPages` config parsing (`0`/`all`/`unlimited` all mean
unbounded).

`calls` is the only stream with a cursor field (`x-cursor-field: call_time`, matching legacy's
`CursorFields: []string{"call_time"}`); no `incremental` block is declared since legacy never
actually filters `calls` requests by `call_time` server-side (it is catalog metadata only in
`streams()`, not a wired request-param filter) — matching the auth0/users precedent for the
identical "catalog cursor field with no wired incremental filter" shape.

Field mapping is a direct 1:1 projection of legacy's `mapRecord` functions:

- `calls`: `id`, `direction`, `status`, `call_time`, `duration`, `agent_id`, `phone_number`
  (`callRecord`, `streams.go:27-29`).
- `agents`: `id`, `name`, `email`, `status` (`agentRecord`, `streams.go:31-33`).
- `teams`: `id`, `name` (`teamRecord`, `streams.go:35-37`).
- `numbers`: `id`, `phone_number`, `name` (`numberRecord`, `streams.go:39-41`).

No `computed_fields` renames are needed — every schema field name matches the raw API field name
verbatim.

## Write actions & risks

None. Freshcaller is a read-only source in this bundle, matching legacy
(`Capabilities: connectors.Capabilities{..., Write: false}`, `freshcaller.go:47`).

## Known limits

- Only the 4 legacy-parity streams are implemented; the broader Freshcaller API surface (call
  metrics, IVR/call routing configuration, business calendars, webhooks) is out of scope for this
  wave — see `api_surface.json`'s `excluded` entries.
- `calls`'s `call_time` cursor field is catalog metadata only (no server-side incremental filter),
  matching legacy exactly — not a deviation, since legacy itself never wires it into a request
  parameter.
