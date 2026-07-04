# Overview

Mailosaur is a wave2 fan-out declarative-HTTP migration. Mailosaur is an email and SMS testing
service; this bundle reads its virtual servers, the message summaries within a server, and account
usage transactions through the Mailosaur REST API (`GET https://mailosaur.com/api/...`). This
bundle migrates `internal/connectors/mailosaur` (the hand-written connector it replaces at
capability parity); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Mailosaur API key via the `password` secret; it is sent as the password of HTTP Basic
auth (username defaults to the literal `api`, matching Mailosaur's own convention, and is
overridable via the `username` config value). It is never logged.

## Streams notes

`servers` (`GET /servers`) is a single-page endpoint whose response body IS the array itself (no
enclosing envelope key) — `records.path: "."` matches legacy's `recordsPath: "."` /
`connsdk.RecordsAt(resp.Body, ".")`.

`messages` (`GET /messages`) is scoped to one Mailosaur virtual server: the required `server`
config value is sent as the `server` query param (matching legacy's `base.Set("server", server)`;
an absent `server` hard-errors on both sides — legacy: `"mailosaur stream \"messages\" requires
config server (server id)"`; engine: an unresolved `config.server` query-template key, per
conventions.md §5's config-validation-parity precedent). The optional `received_after` config value
is sent as `receivedAfter` (opt-in optional-query object dialect, `omit_when_absent: true`,
matching legacy's `if receivedAfter := ...; receivedAfter != "" && stream == "messages"`
conditional). Pagination follows Mailosaur's zero-indexed `page`/`itemsPerPage` convention
(`pagination.type: page_number`, `page_param: page`, `size_param: itemsPerPage`, `start_page: 0`,
`page_size: 50` matching legacy's `mailosaurDefaultPageSize`), records at `items` — a page
returning fewer than `itemsPerPage` records stops the read, matching legacy's `harvest`.

`transactions` (`GET /usage/transactions`) reports account transactional usage over Mailosaur's own
trailing 31-day window; legacy reads it as a single bounded GET with no pagination
(`streamEndpoint.paginated` unset/false for this endpoint, unlike `messages`) — this bundle
declares no `pagination` block for the stream (falling back to `base`'s absence, i.e. `none`),
matching legacy exactly. Records live at `items`.

## Write actions & risks

None. Mailosaur is a read-only source for this connector (legacy's `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` hard request-count cap
  (`mailosaurMaxPages`) on the `messages` stream's read loop. The engine's `page_number` paginator
  has no `MaxPages`-equivalent knob wired to a config value; pagination is bounded only by the
  short-page stop signal, matching Mailosaur's own real termination behavior. `max_pages` is not
  declared in `spec.json`.
- Full Mailosaur API surface (message content/attachment retrieval, spam analysis, server
  creation/deletion, message/server deletion, account plan limits) is out of scope for wave2; see
  `api_surface.json`'s `excluded` entries (`out_of_scope` for not implemented in this bundle,
  `destructive_admin` for delete mutations, `non_data_endpoint` for the account limits snapshot).
  Only the 3 legacy-parity read streams are implemented.
