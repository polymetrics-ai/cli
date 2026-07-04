# Overview

SendPulse is a Pass B full-surface expansion of the wave2 fan-out migration. This bundle reads
SendPulse address books, campaigns, senders, per-book subscriber emails, and the account-wide
email blacklist, and writes address-book/email-membership/campaign/sender/blacklist lifecycle
mutations, through the SendPulse REST API. It migrates `internal/connectors/sendpulse` (the legacy
hand-written connector, which stays registered and unchanged until wave6's registry flip); legacy
was read-only (`Write` always returned `connectors.ErrUnsupportedOperation`), so every write action
here is genuinely new capability, not a ported behavior.

## Auth setup

Provide `client_id`/`client_secret` secrets (SendPulse API app credentials); the bundle exchanges
them for a bearer token via OAuth2 client-credentials (`auth.mode: oauth2_client_credentials`)
against `token_url`, matching legacy's `connsdk.OAuth2ClientCredentials`. `token_url` defaults to
`https://api.sendpulse.com/oauth/access_token` (legacy's own hardcoded literal-append default);
`base_url` independently defaults to `https://api.sendpulse.com`. See Known limits for the
narrowed case where only one of the two is overridden.

## Streams notes

`addressbooks`, `campaigns`, `senders` (the original 3 legacy-parity streams): unchanged shape,
all declare `"projection": "passthrough"` (legacy's `Read` emits every raw decoded field verbatim
via `connsdk.Harvest`, with no `mapRecord`-style field-building — schema-mode projection would
silently drop any undeclared field, a parity regression). `GET` against the SendPulse list
endpoint, records at the JSON body root (`records.path: "."`). Pagination is `page_number`
(`page_param: page`, `size_param: limit`, `start_page: 1`, `page_size: 100`), `max_pages` left
unbounded (see Known limits). Primary keys: `addressbooks`/`campaigns` use `id`, `senders` uses
`email`.

`blacklist` (`GET /blacklist`) is a new top-level list stream. SendPulse's real blacklist endpoint
takes NO pagination parameters at all (confirmed against the official PHP client's
`getBlackList(): return $this->get('blacklist');` — no `limit`/`offset`/`page` argument exists),
so this stream declares a `pagination: {"type": "none"}` override rather than inheriting the base
`page_number` spec — the base spec's `page`/`limit` params would be genuinely-dead query noise
SendPulse's real API silently ignores, and declaring `pagination: none` is the honest
representation instead of sending parameters that do nothing.

`emails_in_book` (`GET /addressbooks/{{ fanout.id }}/emails`) is a sub-resource fan-out over
`addressbooks`: SendPulse's real per-subscriber-email data is only reachable scoped to a single
address book id, with no cross-book list endpoint (see Known limits for the excluded cross-book
`/emails/{email}` global-lookup variant). `fan_out.ids_from.request` issues the SAME paginated `GET
/addressbooks` request the `addressbooks` stream itself uses (`records_path: "."`, `id_field:
"id"`, reusing this stream's effective pagination — the base `page_number` spec, unbounded, since
`emails_in_book` declares no pagination override of its own), then `into.path_var: "id"`
threads each discovered book id into `/addressbooks/{{ fanout.id }}/emails`, and `stamp_field:
"book_id"` writes the source book id (always a STRING — see sendowl's docs.md for the identical
fan_out-dialect-wide constraint) onto every emitted subscriber record after projection.

None of the 5 streams are incremental: neither legacy nor SendPulse's real list endpoints for any
of these resources support a server-side updated-since filter.

## Write actions & risks

10 write actions, none of which existed in legacy:

- **`create_addressbook`/`update_addressbook`/`delete_addressbook`** (`POST /addressbooks`, `PUT
  /addressbooks/{{ record.id }}`, `DELETE /addressbooks/{{ record.id }}`, `body_type: json`
  throughout — SendPulse's real API sends a JSON body on every POST/PUT/DELETE request, confirmed
  against the official PHP client's `sendRequest`, which sets `Content-Type: application/json` and
  `json_encode`s the body for every non-GET method including DELETE): address-book lifecycle.
  `delete_addressbook` is `idempotent`/`missing_ok_status: [404]`.
- **`add_emails_to_book`/`remove_emails_from_book`** (`POST`/`DELETE
  /addressbooks/{{ record.id }}/emails`, `body_fields: ["emails"]`): subscribe/unsubscribe email
  addresses to/from a book. `emails` is declared a bare `"type": "array"` with no `items`
  constraint: SendPulse's real `addEmails` accepts EITHER a bare email string or a `{email,
  variables}` object per array entry (confirmed against the PHP client's `addEmails`, which passes
  the caller's array through unmodified) — a `items: {"type": "object"}` constraint would wrongly
  reject the plain-string shape.
- **`create_campaign`** (`POST /campaigns`, `body_type: json`): creates a new email campaign
  against a real address book. Depending on account send-scheduling settings this may trigger an
  actual send to real subscribers — the highest-impact action in this bundle.
- **`cancel_campaign`** (`DELETE /campaigns/{{ record.id }}`): cancels a scheduled/in-progress
  campaign (SendPulse's real cancel endpoint is `DELETE`, not a `PATCH`/status-field update,
  confirmed against the PHP client's `cancelCampaign`).
- **`add_sender`/`remove_sender`** (`POST`/`DELETE /senders`, no path id — SendPulse identifies the
  sender to remove by an `email` BODY field on the DELETE request, not a path segment, confirmed
  against the PHP client's `removeSender`): sender-identity lifecycle. `add_sender` triggers a
  real activation email SendPulse sends to the new sender address.
  `remove_sender` breaks any campaign still referencing that sender.
- **`add_to_blacklist`/`remove_from_blacklist`** (`POST`/`DELETE /blacklist`, `emails` as a
  BASE64-ENCODED STRING body field, not a plain string or array — confirmed against the PHP
  client's `addToBlackList`/`removeFromBlackList`, both of which `base64_encode()` the caller's
  raw address(es) before sending): account-wide send-suppression lifecycle. The bundle's
  `record_schema` declares `emails` as a plain `"type": "string"` (the caller is responsible for
  supplying the already-base64-encoded value, matching what the real wire body actually carries —
  this dialect has no `base64`-encoding filter usable on write-body fields, only on
  `computed_fields`/interpolated read-side templates, so the encoding step cannot be performed
  declaratively on the write path).

## Known limits

- **Dynamic (fixture-replay) conformance checks are marked `skip_dynamic` at the bundle level**
  (unchanged from the prior wave2 shape — see `metadata.json`'s `conformance` block).
  `oauth2_client_credentials` auth requires a real `token_url`; conformance's synthetic
  non-secret config value is not a resolvable URL, so every auth-resolving dynamic check,
  including `write_request_shape:*` for the 10 new write actions, is skipped for the identical
  reason as the original 3 read streams. The write actions are net-new capability with no legacy
  behavior to structurally compare against at all; their request-shape correctness is instead
  documented by `fixtures/writes/*.json` (each recording the real SendPulse wire shape per the
  official PHP client) and asserted statically by `connectorgen validate`'s
  `write_schemas_valid`/`write_path_fields` checks.
- **`blacklist` and `emails_in_book`'s own per-id requests are declared with NO pagination**
  (`blacklist` explicitly via `pagination: {"type": "none"}`; `emails_in_book`'s per-book request
  inherits the base `page_number` spec, unbounded) — matching each endpoint's REAL documented
  parameter set rather than blanket-applying the base spec to every new stream.
- **`max_pages` is left unbounded** (undeclared in `base.pagination`) rather than baked in at
  legacy's own default of 1 — see the prior wave2 rationale (unchanged): `PaginationSpec.MaxPages`
  cannot be wired to config at all, so "leave unbounded" (matching the `elasticemail`-style
  default elsewhere in this codebase) was chosen over silently truncating every real sync to a
  single page. This also governs the `emails_in_book` fan-out's id-listing request (walks every
  page of `/addressbooks`, not just the first 100) and the per-book-id sub-sequence.
- `token_url`'s default is a fixed literal, not a `base_url`-derived value — unchanged from the
  prior wave2 shape; see that section's original rationale (a caller overriding `base_url` alone
  must also override `token_url` to match).
- **`emails` on `add_to_blacklist`/`remove_from_blacklist` must be pre-base64-encoded by the
  caller.** This dialect's write-body construction has no encoding-filter mechanism (the `base64`
  interpolation filter only applies to `{{ }}` template resolution on read-side
  paths/headers/query/computed_fields, never to a write action's `record_schema`-validated body
  fields) — a caller must supply the value already encoded, matching exactly what the real
  SendPulse wire body expects, so no data-parity gap exists, only an authoring-ergonomics one.
- Full SendPulse API surface still has documented exclusions beyond what's covered here (the
  entire SMTP/transactional-email, bulk-SMS, web-push, and Automation-360-event product surfaces,
  plus narrower per-record variable/analytics sub-endpoints) — see `api_surface.json`'s
  per-endpoint `excluded` entries, each with a specific closed-vocabulary category and reason (no
  blanket "Pass B capability expansion" placeholders remain).
