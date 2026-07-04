# Overview

Brevo (formerly Sendinblue) is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to
the full practical read/write surface. It reads Brevo contacts, email campaigns, contact lists,
contact segments, senders, sender domains, CRM companies, CRM deals, and webhooks through the
Brevo REST API v3 (`https://api.brevo.com/v3/...`), and writes contact/list/sender/company/deal/
webhook lifecycle mutations. This bundle originally migrated `internal/connectors/brevo` (the
hand-written connector, read-only); the legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide a Brevo API key via the `api_key` secret; it is sent as the `api-key` header
(`api_key_header` auth mode, matching legacy's `connsdk.APIKeyHeader("api-key", secret, "")`) and
is never logged. `base_url` defaults to `https://api.brevo.com/v3` and may be overridden for
tests/proxies.

## Streams notes

`contacts` and `email_campaigns` (Brevo's `/emailCampaigns` endpoint; stream renamed to snake_case
per this repo's naming convention, §2) share the same shape: `GET` list endpoints paginated with Brevo's
offset/limit convention (`pagination.type: offset_limit`, `limit_param: limit`, `offset_param:
offset`, `page_size: 100` — matches legacy's `brevoDefaultPageSize`); records live at `contacts`/
`campaigns` respectively (legacy's `recordsPath`). Both support Brevo's `modifiedSince` incremental
filter (`incremental.request_param: modifiedSince`, `param_format: rfc3339` — sent verbatim as the
persisted cursor or, on a fresh sync, the RFC3339 `start_date` config value, identical to legacy's
`incrementalLowerBound`). `contacts_lists` (`GET /contacts/lists`, records at `lists`) is also
offset/limit paginated but has no incremental filter, matching legacy (`supportsModifiedSince:
false`). `senders` (`GET /senders`, records at `senders`) is a single-request, non-paginated
endpoint (legacy's `paginated: false`) — no `pagination` block is declared for it, matching the
engine's `none` default.

**Pass B additions.** `senders_domains` (`GET /senders/domains`, records at `domains`) is
single-request, non-paginated, no incremental filter. `contacts_segments` (`GET
/contacts/segments`, records at `segments`) is offset/limit paginated like `contacts_lists`, no
incremental filter (Brevo's segments endpoint has no `modifiedSince`-equivalent). `webhooks` (`GET
/webhooks`, records at `webhooks`) is single-request, non-paginated; its schema declares
`x-cursor-field: modifiedAt` (a genuinely sortable field present on every record) but the stream
itself has no `incremental` block since Brevo's `/webhooks` GET accepts no server-side
modified-since filter (§8 rule 2: no server-side filter → no `incremental` block, keep
`x-cursor-field` in the schema only). `companies` (`GET /companies`, records at `items`) is
Brevo's one `page`-numbered (not offset/limit) list endpoint in this bundle
(`pagination.type: page_number`, `page_param: page`, `size_param: limit`, `start_page: 1`) and
supports `modifiedSince`; its cursor value lives at the nested `attributes.last_updated_at` path,
which the engine's `incremental.cursor_field`/`x-cursor-field` machinery requires to be a
TOP-LEVEL schema property (flat `raw[cursorField]` map lookup, both
`conformance`'s `checkCursorAdvances` and the client-filtered read path) — a `computed_fields`
entry (`"last_updated_at": "{{ record.attributes.last_updated_at }}"`) lifts it to a top-level
field via typed bare-reference extraction before projection, exactly matching the "cursor field
must be a top-level schema property" convention. `crm_deals` (`GET /crm/deals`, records at
`items`) is offset/limit paginated and supports `modifiedSince`; its cursor value similarly lives
at the nested `attributes.last_updated_date` path and is lifted the same way
(`computed_fields.last_updated_date`).

## Write actions & risks

Fourteen write actions, none present in legacy (legacy shipped `capabilities.write: false`):

- **`create_contact`** / **`update_contact`** / **`delete_contact`** — contact lifecycle.
  `update_contact` mutates attributes, list membership (`listIds`/`unlinkListIds`), or blacklist
  status — changing `emailBlacklisted`/`smsBlacklisted` affects real send eligibility immediately.
  `delete_contact` is irreversible (removes engagement history too); `delete.missing_ok_status:
  [404]` treats an already-absent contact as a successful idempotent delete.
- **`create_contacts_list`** — creates a new list under an existing folder (no update/delete
  endpoint exists in the API for lists; see `api_surface.json`'s `duplicate_of`/`destructive_admin`
  entries for why those are excluded rather than modeled).
- **`create_sender`** / **`update_sender`** / **`delete_sender`** — sender identity lifecycle.
  `create_sender` triggers a real verification email to the target address; `update_sender`
  affects every campaign that references the sender going forward; `delete_sender` is irreversible.
- **`create_company`** / **`update_company`** — CRM company lifecycle (no delete write modeled;
  `DELETE /companies/{id}` is `requires_elevated_scope`-excluded pending admin-role review, per
  `api_surface.json`).
- **`create_deal`** / **`update_deal`** — CRM deal lifecycle (same delete-exclusion reasoning as
  companies).
- **`create_webhook`** / **`update_webhook`** / **`delete_webhook`** — webhook subscription
  lifecycle. Creating or updating a webhook's `url`/`events` registers/redirects live event
  delivery (opens/clicks/bounces/unsubscribes/list-additions) to an external endpoint of the
  caller's choosing — review the target before enabling, per `metadata.json`'s `risk.write`.
  `delete_webhook` is irreversible; `delete.missing_ok_status: [404]` treats an already-absent
  webhook as a successful idempotent delete.

Every action's per-record `risk` string in `writes.json` is the authoritative, reviewable summary;
`metadata.json`'s `risk.write`/`risk.approval` roll these up for the connector as a whole.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (1-1000,
  default 100) and `max_pages` (0/all/unlimited or a positive integer) as config-driven overrides
  read fresh on every `Read` call (`brevoPageSize`/`brevoMaxPages`). The engine's `offset_limit`
  paginator reads its page size only from `pagination.page_size` (a fixed literal baked into
  `streams.json`, sourced once at bundle-load time), with no `config.*`-templated override
  mechanism — matching the exact limitation documented in this repo's `bitly` bundle for its
  `next_url` paginator. This bundle declares `page_size: 100` to match legacy's real default
  exactly; a caller cannot raise or lower it at read time as legacy allowed. `spec.json` still
  declares `page_size`/`max_pages` for documentation continuity with legacy's config surface, but
  neither is wired into any template (F6-adjacent: these are the same "legacy had this knob, the
  engine's chosen pagination type cannot express it" shape as bitly's `page_size`, not silently
  dropped dead config).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a
  `previous_cursor` field onto fixture-mode records when a prior cursor is set. This bundle's
  schemas and parity target the live wire shape only; the engine's own conformance/fixture-replay
  harness supersedes legacy's in-code fixture-mode path.
- The `contacts` stream's 2-page conformance fixture (`fixtures/streams/contacts/{page_1,
  page_2}.json`) uses a full 100-record page 1 (matching the real `page_size: 100`) followed by a
  1-record short page 2, so pagination truly terminates on Brevo's own short-page signal rather than
  an artificially-lowered page size — this keeps the fixture's wire shape identical to a real
  production page count instead of trading fixture verbosity for behavioral accuracy. The
  Pass B `contacts_segments`/`companies`/`crm_deals` 2-page fixtures follow the identical
  full-page/short-page pattern.
- **No write action deletes a CRM company or deal.** `DELETE /companies/{id}` and
  `DELETE /crm/deals/{id}` both exist in Brevo's API but are excluded (`api_surface.json`,
  `requires_elevated_scope`) rather than modeled as `delete_company`/`delete_deal` write actions —
  irreversible CRM-record deletion is gated behind the CRM admin/owner role in Brevo's own
  permission model, and this pass draws the write-surface line at create/update for CRM objects
  pending a dedicated elevated-scope review. Contact, sender, and webhook deletes ARE modeled
  (`delete_contact`/`delete_sender`/`delete_webhook`) since those are ordinary account-level
  mutations with no equivalent elevated-role gate in Brevo's docs.
- **No write action creates a webhook filtered by `channel`/scoped update beyond url/description/
  events.** Brevo's `updateWebhook` body accepts the same field set as `createWebhook` minus
  `type` (immutable after creation); `update_webhook`'s `record_schema` reflects this — a record
  attempting to change `type` post-creation is simply ignored by the body-construction rule (not
  included in the allow-checked field set), matching the API's own immutability, not silently
  dropped data.
