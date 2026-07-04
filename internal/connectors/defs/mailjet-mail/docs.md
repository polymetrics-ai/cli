# Overview

Mailjet Mail is a wave2 fan-out declarative-HTTP migration. It reads Mailjet contacts, contact
lists, messages, campaigns, and aggregated statistics through the Mailjet Email REST API v3
(`GET https://api.mailjet.com/v3/REST/...`). This bundle migrates
`internal/connectors/mailjet-mail` (the hand-written connector it replaces at capability parity);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Mailjet public API key via the `api_key` config value (sent as the HTTP Basic auth
username) and a Mailjet secret API key via the `api_key_secret` secret (sent as the HTTP Basic auth
password); `api_key_secret` is never logged. This matches legacy's
`connsdk.Basic(apiKey, secret)` exactly. `api_key` itself is not a credential the API accepts alone
(it is a public identifier paired with the secret), matching legacy's own choice to keep it a plain
config value rather than `x-secret`.

## Streams notes

All 5 streams (`contacts`, `contactslists`, `messages`, `campaigns`, `stats`) share the same shape:
`GET` against the Mailjet v3 REST resource (`contact`, `contactslist`, `message`, `campaign`,
`statcounters`), records at the `Data` array key of the `{Count, Total, Data:[...]}` envelope,
primary key `["ID"]`. None of the 5 legacy streams declare an incremental cursor field (the Mailjet
Email REST API supports full-refresh reads only for this connector, matching legacy's own comment);
this bundle declares no `incremental` block for any stream, matching legacy exactly.

Pagination follows Mailjet's `Limit`/`Offset` convention (`pagination.type: offset_limit`,
`limit_param: Limit`, `offset_param: Offset`, `page_size: 100`, matching legacy's
`mailjetDefaultPageSize`) — the next page's `Offset` advances by `Limit` (100) until a page returns
fewer than `Limit` records, exactly matching legacy's `connsdk.OffsetPaginator`.

## Write actions & risks

None. The Mailjet Email REST API source is read-only for this connector (legacy's `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`page_size` is not runtime-configurable per legacy's config-driven override.** Legacy exposes
  `page_size`/`max_pages` as config-driven overrides (`mailjetPageSize`/`mailjetMaxPages`,
  `mailjet-mail.go:261-289`). `spec.json` declares `page_size` (default `"100"`) for documentation
  parity, but the engine's `offset_limit` paginator reads its page size from the STATIC
  `streams.json` `pagination.page_size` field, not a per-request config value, so the declared
  `page_size` config property is not actually wired into the pagination page-size knob itself
  (only into the default-materialization surface). `max_pages` (legacy's hard request-count cap) is
  not modeled at all — the engine's `offset_limit` paginator has no `MaxPages`-equivalent knob
  wired to a config value; pagination is bounded only by the short-page stop signal, matching
  Mailjet's own real termination behavior.
- Full Mailjet Email API surface (sending mail, templates, senders, DNS validation, event
  webhooks, bulk contact-list management) is out of scope for wave2; see `api_surface.json`'s
  `api_surface.json` concrete exclusion entries. Only the 5
  legacy-parity read streams are implemented.
