# Overview

ConvertKit (Kit) reads subscribers, forms, sequences, tags, broadcasts, custom fields, and
purchases, and writes subscriber/tag/form/sequence/broadcast/custom-field/purchase/webhook
mutations, through the ConvertKit v3 REST API. This bundle migrates
`internal/connectors/convertkit` (the legacy hand-written connector, kept registered and unchanged
until wave6's registry flip) to a declarative defs bundle at capability parity, then expands past
that legacy-parity floor with a Pass B full-surface review of the live v3 API reference.

Kit's v4 API (`api.kit.com/v4`) exists and is the vendor's forward path, but v4 uses an
incompatible auth scheme (an `X-Kit-Api-Key` header or OAuth2, not the `api_secret` query-param
scheme this bundle's auth candidates are built on) and a distinct credential namespace (v4 keys are
not interchangeable with v3 keys, per Kit's own upgrade guide). Kit's docs confirm v3 "is still
available for use" today even though deprecated, so this Pass B expansion stays within v3 rather
than silently rewriting the connector onto an incompatible v4 auth model — that would be a breaking
connector replacement, not a surface expansion.

## Auth setup

Provide the ConvertKit v3 API secret via the `api_key` secret; it is sent as the `api_secret` query
parameter on every request (read AND write) and never logged. Legacy resolved the same credential
from any of three secret names in precedence order (`api_key`, then `access_token`, then
`api_secret`, first non-empty wins) to tolerate differently-named credential storage across catalog
versions. This bundle reproduces that exact precedence with a 3-candidate `base.auth` list, each
candidate gated by a `when` clause on its own secret key (`docs/migration/conventions.md`'s
dual-auth first-match-wins ordering pattern) — `api_key` is declared first, `access_token` second,
`api_secret` last, matching legacy's fallback order exactly.

## Streams notes

`subscribers` and `broadcasts` use ConvertKit's page-based pagination (`page`/`per_page` query
params, `page_number` pagination type, `page_size: 50` matching legacy's default). `purchases` is
also page-based but the real API documents no `per_page` override — only `page` — so its
`streams.json` pagination declares `size_param: ""` (never sends a size param), matching the
real API's own `total_purchases`/`page`/`total_pages` list shape (50 purchases per page is the
API's own fixed default). The engine's `page_number` paginator stops on a short (or empty) page —
legacy's own stop condition (for the original 5 streams) combines a short/empty page with the
response's `total_pages` field, but in every real ConvertKit response shape a final page IS the
short/empty one once `total_pages` is reached, so this is a faithful, non-data-changing
representation of the same stop signal.

`forms`, `sequences`, `tags`, and `custom_fields` return their full collection in a single
unpaginated array under their own resource key (`base.pagination` is overridden per-stream to
`type: none`) — matches legacy's `paginated: false` endpoint table exactly for the first 3;
`custom_fields` is a new Pass B stream with the identical unpaginated shape per its own
documentation (an account has at most 140 custom fields).

Every original ConvertKit object schema declares `x-cursor-field: "created_at"`, matching legacy's
published catalog (`CursorFields: ["created_at"]` on every stream via `convertkitStreams()` in
`internal/connectors/convertkit/streams.go`), and every stream now also declares a bare
`incremental: { cursor_field: "created_at" }` block with no `request_param` — legacy's
`Read`/`harvest` never sends a server-side incremental filter (the upstream source is full-refresh
only in practice; ConvertKit's v3 list endpoints accept no since/updated-since query param), so no
`request_param` is added, but the engine derives `Manifest.SyncModes`'
`incremental_append`/`incremental_append_deduped` strictly from `stream.Incremental != nil`
(`DerivedSyncModes`, `internal/connectors/engine/connector.go`), not from schema `x-cursor-field`
alone. The new `purchases` stream declares the same shape keyed on its own real cursor field,
`transaction_time` (a purchase's own genuine per-record timestamp, per the v3 purchases response
schema); `custom_fields` declares no `x-cursor-field`/`incremental` block at all, since its real API
response carries no timestamp field of any kind (`id`/`name`/`key`/`label` only).

## Write actions & risks

Thirteen write actions are exposed, all `external mutation` / `approval required` per
`metadata.json`'s `capabilities.write: true`:

- `update_subscriber` (`PUT /subscribers/{id}`) — mutates a subscriber's name/email/custom fields.
- `create_tag` (`POST /tags`) — creates a new account tag.
- `tag_subscriber` (`POST /tags/{tag_id}/subscribe`) — applies a tag to a subscriber by email,
  creating the subscriber if the email is new (matches the real API's upsert-by-email semantics).
- `remove_tag_from_subscriber` (`DELETE /subscribers/{subscriber_id}/tags/{tag_id}`) — removes a
  tag from a subscriber by numeric id (the email-addressed `POST .../unsubscribe` variant is a
  duplicate mutation path and is not separately exposed — see `api_surface.json`).
- `subscribe_to_form` / `subscribe_to_sequence` (`POST /forms|sequences/{id}/subscribe`) — subscribe
  an email to a form/sequence, creating the subscriber if new.
- `create_broadcast` / `update_broadcast` / `delete_broadcast` — full broadcast lifecycle. A
  scheduled broadcast (`send_at`/`published_at` set) sends live email to the account's real
  subscriber list; `create_broadcast`/`update_broadcast` are flagged accordingly.
- `create_custom_field` / `update_custom_field` — custom-field lifecycle (label rename only for
  update, matching the real API: "the key will change but the name remains").
- `create_purchase` — records a purchase-tracking transaction (nested `purchase` object body,
  matching the real API's request shape exactly).
- `create_webhook` / `delete_webhook` (`POST`/`DELETE /automations/hooks`) — webhook lifecycle; the
  real v3 path is still `/automations/hooks` (the `/webhooks` path is v4-only).

No `delete_subscriber`/`delete_custom_field`/global-`unsubscribe` action is exposed: custom-field
deletion "permanently removes the field and all associated subscriber data" per Kit's own docs
(`destructive_admin`), and global unsubscribe drops the email from every form/sequence/tag on the
account at once with no reverse-ETL use case (`destructive_admin`) — see `api_surface.json`.

## Known limits

- Legacy's configurable `page_size` (1-50, default 50) and `max_pages` (0/all/unlimited or a
  positive integer cap) config knobs are not modeled: `streams.json`'s `pagination.page_size` is a
  fixed JSON literal with no config-driven override mechanism (same class of limitation as
  searxng's `page_size`/`max_pages`, `docs/migration/conventions.md`'s read-only/no-auth golden). A
  declared-but-unwireable config key is worse than an absent one (F6, REVIEW.md), so neither is
  declared in `spec.json`; every paginated stream is fixed at `page_size: 50`, matching legacy's
  own default (and, for `purchases`, the real API's own fixed default) exactly.
- v4-only surfaces (email templates, accounts/current, segments, courses-as-sequences-v4-shape) are
  out of scope: this bundle's auth is v3-specific and the two API versions are not credential- or
  request-shape-compatible; see the Overview section above.
- `fixtures/streams/subscribers/**` and `fixtures/streams/broadcasts/**` ship a full 50-record
  first page (matching the real `page_size: 50` stop threshold exactly, per
  `docs/migration/conventions.md` §4's "no string-ification workaround" / real-wire-shape rule) so
  `pagination_terminates` proves the engine's actual production page-size threshold, not an
  artificially shrunk one. `fixtures/streams/purchases/**` ships a genuine 2-page fixture (2
  records then 1) since a full 50-record page for this new stream would be pure fixture bulk with
  no additional coverage value beyond what `subscribers`/`broadcasts` already prove for the same
  paginator.
