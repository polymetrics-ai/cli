# Overview

RingCentral is a wave2 fan-out declarative-HTTP migration. It reads RingCentral account
extensions, call-log records, messages, address-book contacts, and devices through the RingCentral
REST API v1.0 (`GET https://platform.ringcentral.com/restapi/v1.0/...`). This bundle targets
capability parity with `internal/connectors/ringcentral` (the hand-written connector it migrates);
the legacy package stays registered and unchanged until wave6's registry flip. Read-only
(`capabilities.write` is `false`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`).

## Auth setup

Provide an already-issued RingCentral OAuth2 access token via the `access_token` secret; it is
sent as an `Authorization: Bearer <access_token>` header, matching legacy's
`connsdk.Bearer(token)` (`ringcentral.go:172`) exactly. This bundle does not perform the OAuth2
authorization-code or JWT-bearer exchange itself — legacy never did either (it only ever consumes
a config-supplied `access_token` secret directly, with no token-refresh/exchange logic anywhere in
`ringcentral.go`). `base_url` defaults to `https://platform.ringcentral.com/restapi/v1.0` and may
be overridden for tests/proxies.

## Streams notes

All 5 streams (`extensions`, `call_log`, `messages`, `contacts`, `devices`) share the identical
shape: `GET` against a RingCentral account/extension-scoped list endpoint
(`/account/~/extension`, `/account/~/extension/~/call-log`,
`/account/~/extension/~/message-store`, `/account/~/extension/~/address-book/contact`,
`/account/~/device` — the `~` segments are RingCentral's own "current authorized
account/extension" convention, passed through literally, not templated), records at the response
body's `records` array, matching legacy's own `recordsPath: "records"` declaration for every
endpoint (`ringcentral.go:106-112`) exactly. Pagination is `page_number` (`page`/`perPage`,
`page_size: 100`), stopping on a short page exactly as legacy's `connsdk.PageNumberPaginator`
does.

Legacy applies five passthrough filters (`dateFrom`, `dateTo`, `type`, `messageType`,
`direction`) identically to every stream's request (`ringcentral.go:87-92`'s loop iterates a fixed
key list regardless of which stream is being read) — this bundle reproduces that exact blanket
behavior via `base.query`'s five `omit_when_absent` entries (shared across all streams), sent only
when the corresponding config value is set. RingCentral's real wire shape uses camelCase field
names (`extensionNumber`, `startTime`, `creationTime`, `firstName`, `lastName`), while legacy's own
`Catalog()` declares snake_case field names (`extension_number`, `start_time`, `creation_time`,
`first_name`, `last_name`) as the advisory stream shape — this bundle bridges that gap with
`computed_fields` renames (`"extension_number": "{{ record.extensionNumber }}"` etc.) on the 3
affected streams (`extensions`, `call_log`, `messages`, `contacts`), reproducing legacy's declared
field-name convention as the actual emitted shape. `computed_fields` also stamps a static `stream`
marker on every record, matching legacy's `mapRecord`'s `out["stream"] = stream`.

`start_time`/`creation_time` are declared as `x-cursor-field` on `call_log`/`messages`
respectively, matching legacy's own `CursorFields` Catalog declarations (`extensions`, `contacts`,
`devices` have no `CursorFields` in legacy). No `incremental` block is declared on any stream:
legacy's `Read` never reads a persisted sync cursor back into `dateFrom`/`dateTo`
(`harvest` reads only `req.Config.Config[key]`, never `req.State["cursor"]`) — it always resends
the exact same raw config value on every sync, with no forward advancement.

The `check` request (`GET /account/~`) carries no query params, matching legacy's
`r.DoJSON(ctx, http.MethodGet, "account/~", nil, nil, nil)` exactly (`ringcentral.go:47`).

## Write actions & risks

None. Legacy `ringcentral.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`page_size` bounded 1-1000, default 100; `max_pages` 0/all/unlimited for unbounded).
  The engine's `page_number` paginator reads `PaginationSpec.PageSize`/`MaxPages` as static
  bundle-authored integers, not config templates — there is no mechanism to wire a `spec.json`
  property into either field. This bundle sends `page_size: 100` (legacy's own default) as a
  static value in `streams.json`'s `base.pagination` block; neither `page_size` nor `max_pages` is
  declared in `spec.json` (F6: dead config is worse than absent config). Pagination is otherwise
  unbounded (matches legacy's `max_pages: 0` = unlimited default) other than the short-page stop
  signal.
- **Legacy's `id` fallback (`uri`/`extensionNumber`/`phoneNumber`/`name`) is not modeled.**
  Legacy's `mapRecord` falls back to a record's `uri`, `extensionNumber`, `phoneNumber`, or `name`
  field when `id` is absent. Every RingCentral resource this bundle reads always carries a numeric
  `id` in its real wire shape (legacy's own `Catalog`/`PrimaryKey` declarations assume `id`
  unconditionally for all 5 streams), so this fallback is defensive dead code against the real API
  — not exercised by any input legacy itself would realistically receive. Documented here for
  completeness, not implemented via a hook.
- **No OAuth2 authorization/token-refresh flow is modeled.** Legacy itself never implements one
  (it consumes a pre-issued `access_token` secret directly with no refresh logic); this bundle
  matches that exactly. Obtaining/refreshing the token is an operator/credential-provisioning
  concern outside this connector's scope, same as legacy.
- The full RingCentral API surface (SMS/fax send, call control, meeting/webinar management,
  presence updates, event subscriptions/webhooks) is out of scope for this wave; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
