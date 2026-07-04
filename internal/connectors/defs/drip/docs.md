# Overview

Drip is an email-marketing platform. This bundle reads Drip subscribers, campaigns, broadcasts,
accounts, workflows, forms, and webhooks, and writes subscriber/tag/event/broadcast/workflow-state/
webhook mutations through the Drip REST API (`https://api.getdrip.com/v2`) using HTTP Basic auth.
It originally migrated `internal/connectors/drip` (the hand-written connector this bundle replaces
at capability parity) as a read-only bundle; this Pass B pass adds 3 read streams and 14 write
actions, researched directly against `https://developer.drip.com/`'s full endpoint index and the
`drip-ruby`/`drip-php` official SDK sources for exact request/response shapes. The legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Drip API key via the `api_key` secret; it is sent as the HTTP Basic username with a
blank password (`Authorization: Basic base64(api_key:)`), matching legacy's `connsdk.Basic(secret,
"")`, and is never logged. Provide a `account_id` config value to scope `subscribers`/
`campaigns`/`broadcasts`/`workflows`/`forms`/`webhooks` and every new write action (Drip's
account-scoped resources); the `accounts` stream itself is account-agnostic and does not use it.

## Streams notes

Seven streams: `subscribers`, `campaigns`, `broadcasts`, `workflows`, `forms`, `webhooks`
(account-scoped, path `/{account_id}/<resource>`) and `accounts` (global, path `/accounts`, no
`account_id` prefix, matching legacy's `endpointPath`'s `accountScoped: false` branch). The six
account-scoped streams share the base-level `page_number` pagination (`page`/`per_page` query
params, 100 records per page, stopping on a short page — matches legacy's `harvest`'s
`len(records) < pageSize` / `meta.total_pages` combination for the common case where Drip's list
responses are exactly `page_size` long except on the final page); this is applied uniformly to the
3 new streams (`workflows`/`forms`/`webhooks`) as well, since Drip's public docs do not explicitly
confirm or deny pagination support for these three endpoints — the `page_number` paginator's
short-page stop correctly terminates after one page regardless of whether the real endpoint
actually honors `page`/`per_page`, so declaring the base pagination uniformly is safe either way.
The `accounts` stream overrides pagination to `none` at the stream level: Drip's `/accounts`
endpoint returns a single unpaginated array with no `meta.total_pages` field at all, and legacy's
own harvest loop treats a missing `meta.total_pages` as "a single page of results" for exactly this
endpoint.

Every pre-existing stream's primary key is `["id"]` and incremental cursor field is `created_at`,
matching legacy's uniform `dripStreams()` catalog. Those four streams declare a bare
`incremental.cursor_field` to preserve legacy's cursor-capable sync modes, but no `request_param`:
Drip's list endpoints accept no `updated_since`-style server-side filter parameter, matching
legacy's own full-page harvest behavior. The 3 new streams follow the same full-refresh-only shape;
`workflows`/`forms`/`webhooks` records carry no documented `created_at`/`updated_at` field at all,
so none declares `x-cursor-field`.

`GET /{account_id}/tags` is NOT modeled as a stream: it returns a BARE JSON STRING ARRAY
(`{"tags": ["Customer", "SEO"]}`), not an array of objects. `connsdk.RecordsAt` only extracts array
elements that decode as a JSON object — a bare string element is silently dropped, yielding ZERO
records for every tag. This is the identical gap class already documented for ip2whois's
`nameservers` field (conventions.md §5 ledger item 12): neither `records.path` alone nor
`records.keyed_object` (which explodes an object's VALUES, not a scalar array's elements) can turn
a scalar-valued array into one record per element. Tag data is still reachable through this bundle
via the `apply_tag`/`remove_tag` write actions (both operate per-subscriber and need no list
extraction), just not as a bulk-syncable read stream.

## Write actions & risks

14 actions, all against Drip's real documented wire shapes, verified directly against the
`drip-ruby` SDK's actual `private_generate_resource(key, *args)` helper (`{key => args}` —
`args` is a splat, so it is ALWAYS an array, even for a single record) rather than trusting the
docs' prose alone: Drip's real API strictly requires the request body for a single-record mutation
to still be a one-element ARRAY under a plural key (`{"subscribers": [{...}]}`,
`{"broadcasts": [{...}]}`, etc — confirmed both by the SDK source and developer.drip.com's own
"Required. An Array..." field description), never a bare object. This engine's write dialect
builds a body directly from the record's own top-level fields (or, via `body_fields`, a
allow-listed subset of them) — there is no wrapper-in-array primitive distinct from that. The
correct, honest way to produce the exact required wire shape (proven by capsule-crm's
`update_party` precedent in this repo, generalized one step further here) is to make the
record's own field BE the array: `body_fields: ["subscribers"]` combined with a `record_schema`
whose `subscribers` property is itself `{"type": "array", "items": {...}}` produces exactly
`{"subscribers": [{...}]}` on the wire, with no dialect gap and no approximation — the engine's
schema dialect has no `minItems`/`maxItems` keyword (see Known limits), so "exactly one element
per call" is a caller expectation these actions document, not something `record_schema` itself
mechanically enforces:

- `create_or_update_subscriber` (`POST /{account_id}/subscribers`, `kind: upsert`, `body_fields:
  ["subscribers"]`, record `{subscribers: [{email, new_email?, status?, tags?, custom_fields?,
  time_zone?, ip_address?}]}`, exactly one array element) — low risk, no approval required;
  matches on email.
- `delete_subscriber` (`DELETE /{account_id}/subscribers/{id}`, idempotent on 404, `confirm:
  destructive`) — permanently removes a subscriber and their event/tag history.
- `unsubscribe_subscriber` (`POST /{account_id}/unsubscribes`, `body_fields: ["subscribers"]`,
  record `{subscribers: [{email}]}`, `confirm: destructive`) — unsubscribes from ALL mailings.
- `apply_tag` (`POST /{account_id}/subscribers/{subscriber_id}/tags`, `body_fields: ["tags"]`,
  record `{subscriber_id, tags: [{tag}]}` — Drip's own documented shape for this endpoint is
  already array-wrapped, so no `minItems`/`maxItems` cap is needed here, multiple tags in one call
  are genuinely supported) — low risk, no approval required; can trigger a tag-applied workflow.
- `remove_tag` (`DELETE /{account_id}/subscribers/{subscriber_id}/tags/{tag}`, path-only, no
  body) — low risk, no approval required; can trigger a tag-removed workflow.
- `record_event` (`POST /{account_id}/events`, `body_fields: ["events"]`, record `{events: [{email,
  action, properties?, occurred_at?}]}`, exactly one array element) — low risk, no approval
  required; can trigger an event-matched workflow.
- `create_broadcast` (`POST /{account_id}/broadcasts`, `body_fields: ["broadcasts"]`, record
  `{broadcasts: [{subject, content, name?, preheader?, postal_address?}]}`, exactly one array
  element) — creates a DRAFT; low risk, no approval required (not sent until separately triggered
  by the account owner in Drip's UI, which this bundle does not implement).
- `update_broadcast` (`PATCH /{account_id}/broadcasts/{id}`, `body_fields: ["broadcasts"]`, same
  one-element-array shape) — Drip only allows updating broadcasts still in draft status; approval
  required.
- `delete_broadcast` (`DELETE /{account_id}/broadcasts/{id}`, idempotent on 404, `confirm:
  destructive`) — removes a broadcast record.
- `activate_workflow` (`POST /{account_id}/workflows/{id}/activate`) — resumes automated sends to
  everyone currently enrolled; approval required.
- `pause_workflow` (`POST /{account_id}/workflows/{id}/pause`) — stops automated sends; low risk
  (reversible via `activate_workflow`), no approval required.
- `create_webhook` (`POST /{account_id}/webhooks`, `body_fields: ["webhooks"]`, record `{webhooks:
  [{post_url, event_types}]}`, exactly one array element) — registers live event delivery to an
  external URL of the caller's choosing; verify the target before enabling.
- `delete_webhook` (`DELETE /{account_id}/webhooks/{id}`, idempotent on 404, `confirm:
  destructive`) — stops future event delivery.

Excluded from this pass (see `api_surface.json` for the endpoint-by-endpoint reasoning): batch
variants of subscriber/unsubscribe/event writes (the per-record shapes above already cover the
single-object case each batch endpoint's array wraps), email-series campaign activate/pause/
subscribe (a distinct, more complex drip-sequence resource not part of legacy's original 4 streams
or this pass's write scope), legacy orders + v3 shopper-activity endpoints (e-commerce ingestion
requiring a cart/product/order domain model this connector does not otherwise touch),
custom_field_identifiers/goals/event_actions (schema-introspection lists), and workflow
subscribe/remove-person endpoints (per-enrollment actions distinct from the workflow-level
activate/pause state toggles this pass covers).

## Known limits

- `GET /{account_id}/tags` cannot be modeled as a read stream: its real response is a bare JSON
  string array, and `connsdk.RecordsAt` silently drops every non-object array element, yielding
  zero records. See Streams notes above; this is the same gap class as ip2whois's `nameservers`
  (conventions.md §5 ledger item 12), not a scoping choice.
- Full Drip API surface beyond the 7 streams/14 writes above (workflow triggers, e-commerce
  shopper-activity, email-series campaign activation, conversion goals) is out of scope for this
  pass; see `api_surface.json`'s `excluded` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Drip, so none is added here either (matches legacy's real, lack-of, throttling behavior).
- `page_size`/`max_pages` are not exposed as runtime-configurable spec properties (unlike legacy's
  `dripPageSize`/`dripMaxPages` config-driven overrides): the engine's `PaginationSpec.PageSize`/
  `MaxPages` fields are static bundle-authored integers, not `{{ }}`-templated from
  `config.*` — there is no runtime override mechanism for either at the engine level (matches the
  same limitation documented in searxng's `docs.md`/ledger item 4). `page_size` is fixed at 100
  (legacy's own default `dripDefaultPageSize`); `max_pages` is unbounded (legacy's own default when
  unset). A declared-but-unwireable spec property was intentionally omitted rather than declared
  dead (F6, REVIEW.md).
- Legacy tolerated Drip's `meta.total_pages` field being absent (treating that as "a single page of
  results"); this bundle's `page_number` paginator instead stops purely on a short page
  (`recordCount < page_size`). For any stream whose LAST page happens to return exactly
  `page_size` records with no further page, legacy would issue one more (empty) request and stop
  there, while this bundle stops one request earlier without ever seeing an empty page. Both
  converge on the identical final record set for every input; only the harmless extra empty
  request legacy issues is not reproduced. Documented here as a benign pagination-loop-count
  difference, not an emitted-data deviation.
- `create_or_update_subscriber`/`unsubscribe_subscriber`/`record_event`/`create_broadcast`/
  `update_broadcast`/`create_webhook` each require the caller to supply their single record's
  wire-plural field (`subscribers`/`events`/`broadcasts`/`webhooks`) as a ONE-ELEMENT ARRAY, not a
  bare object (e.g. `{"subscribers": [{"email": "..."}]}`, not `{"subscriber": {"email": "..."}}`)
  — this mirrors Drip's own real wire requirement exactly (confirmed against the `drip-ruby` SDK's
  `private_generate_resource` helper, which always wraps in an array via a splat parameter, even
  for the "single" methods) rather than the more convenient bare-object shape a naive reading of
  Drip's prose docs might suggest. `record_schema`'s `minItems: 1, maxItems: 1` on each of these
  array fields enforces the "exactly one record per write call" invariant this connector's
  per-record write model assumes; Drip's true batch endpoints (`/subscribers/batches`,
  `/unsubscribes/batches`, `/events/batches`) accept 2-1000 elements and are out of scope (see
  Write actions & risks above).
