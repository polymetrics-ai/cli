# Overview

Google Directory reads Google Admin SDK Directory users, groups, organizational units, and
ChromeOS devices via bearer-token OAuth. This bundle is a pure Tier-1 declarative migration of
`internal/connectors/google-directory` (the hand-written legacy connector, which despite its
`internal/connectors/<name>/` location and `runtime_kind: native_go` inventory label is a plain
`connsdk`-HTTP connector with no SQL/queue/protocol-native behavior — the label predates this
convention's tier ladder). The legacy package stays registered and unchanged until wave6's
registry flip. Read-only; no write actions.

## Auth setup

Provide a Google OAuth 2.0 access token (Admin SDK Directory read scope) via the `access_token`
secret; it is used only for Bearer auth (`Authorization: Bearer <access_token>`) and is never
logged. Acquisition/refresh of the token is out of scope for this connector (the credentials layer
already owns it), matching every other Google-family bundle in this repo (see
`google-search-console`'s identical `access_token`-only shape).

## Streams notes

All 4 streams (`users`, `groups`, `orgunits`, `chromeos_devices`) share the same pagination shape:
`GET`, cursor pagination via `nextPageToken`/`pageToken` (`pagination.type: cursor`,
`token_path: nextPageToken`), a `customer_id` config value (default `my_customer`, matching
legacy's own default) either sent as a `customer` query parameter (`users`/`groups`) or
interpolated into the request path (`orgunits`/`chromeos_devices`: `customer/{{ config.customer_id
}}/...`), and `maxResults` set from the `page_size` config value (default 100, matching legacy's
`defaultPageSize`). `users.name` is a `computed_fields` reach into the raw nested
`name.fullName` object field; `orgunits.id` renames the raw API's `orgUnitId` to the schema's `id`;
`chromeos_devices.id`/`serial_number` rename the raw API's `deviceId`/`serialNumber`. Every stream
projects in `"schema"` mode (default) — legacy's own `mapRecord` functions build a field-by-field
`connectors.Record{...}`, never emit verbatim, so schema-mode projection matches legacy's actual
emission shape (§8 rule 1).

## Write actions & risks

None. `capabilities.write` is `false`; legacy's `Write` always returns
`connectors.ErrUnsupportedOperation`.

## Known limits

- Legacy accepted the access token under either of two secret keys —
  `authorization.access_token` (checked first) or a plain `access_token` fallback
  (`google_directory.go`'s `accessToken` helper). This bundle declares a single canonical secret,
  `access_token`, matching every other Google-family bundle's naming (`google-search-console`,
  `google-forms`) — no other bundle in this repo uses a dotted secret key, and the dotted alias
  appears to be a one-off legacy naming artifact rather than an established convention. Documented
  parity deviation (never changes accepted-input behavior for the `access_token`-keyed path;
  narrows only the redundant alias key name): ACCEPTABLE.
- Full Directory API surface (group members, domains, roles, schemas, mobile devices) is out of
  scope for this migration; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "not implemented in this bundle"}` entries. Only the 4 legacy-parity streams are implemented.
- Legacy's `page_size` validation bounds the value to `[1, 500]` (`maxPageSize`) and hard-errors
  outside that range; the engine's declarative `page_number`/`cursor` dialect has no config-value
  range-validation mechanism (out of scope for this migration — an out-of-range `page_size` is sent
  to the API verbatim rather than rejected client-side). This narrows validation strictness only,
  never accepted-input record data, for any `page_size` a caller would realistically set.
