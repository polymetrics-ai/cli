# Overview

Chameleon is a wave2 fan-out migration, expanded in Pass B to the full documented API surface. This
bundle reads Chameleon in-product experiences (surveys, tours, launchers, tooltips, embeds),
targeting/configuration data (segments, tags, event names), operational data (deliveries,
webhooks), and account data (companies) through the Chameleon v3 REST API. It also writes 8
9 mutations: publish/unpublish for each of the 5 experience types, plus create/delete for Deliveries
and Webhooks. Legacy `internal/connectors/chameleon` (which stays registered and unchanged until
wave6's registry flip) covers only the original 5 read streams and is entirely read-only; this
bundle now covers substantially more of Chameleon's real surface than legacy did.

**Pagination correction (Pass B research finding, not a new deviation)**: legacy's own code comment
states Chameleon's "upstream API has no documented cursor pagination," and the prior wave2 bundle
inherited that same offset/limit assumption. Full-surface research against Chameleon's own Developer
Hub (`https://developers.chameleon.io/apis/*`) shows this was incorrect — every list endpoint
documents `limit`/`before`/`after` cursor pagination, returning a `cursor: {limit, before}` object
in the response body whose `before` value is the next page's cursor token. This bundle corrects
`base.pagination` to `type: cursor` with `token_path: cursor.before` accordingly (see `conventions.md`
§3's `cursor`/`token_path` pagination shape). This is a genuine correctness fix, not a data-changing
deviation: it changes how many requests a sync issues to walk a large collection (previously,
`offset`-based pagination past the first ~2-3 pages would in practice miss or duplicate records
against a real API that never actually implemented offset semantics server-side, since offset/limit
was never real) — every record ever returned is unaffected in shape, only correctly enumerated now.
The 5 original streams' record schemas intentionally retain the legacy emitted field set even where
the newer Chameleon docs use different names. For `surveys`, `tours`, `launchers`, and `tooltips`
that means the legacy experience fields (`title`, `type`, `state`, `is_live`, `created_at`,
`updated_at`) rather than the broader documented `name`/`position`/`published_at` surface; for
`segments` it means the legacy `description` field rather than the newer `items`/`items_op`
definition. This keeps the bundle faithful to `internal/connectors/chameleon`'s emitted record data
during the pre-cutover migration.

## Auth setup

Provide the Chameleon account secret via the `api_key` secret; it is sent as the
`X-Account-Secret` header (`auth.mode: api_key_header`), matching legacy's
`connsdk.APIKeyHeader(chameleonSecretHeader, secret, "")` with no value prefix. `base_url`
defaults to `https://api.chameleon.io/v3` (legacy's `chameleonDefaultBaseURL`).

## Streams notes

All cursor-paginated streams (`surveys`, `tours`, `launchers`, `tooltips`, `segments`, `embeds`,
`event_names`, `tags`, `deliveries`) share the same shape: `GET` against the Chameleon list
endpoint, records at the stream's own top-level JSON field (matching the real documented envelope
key per resource — `surveys`/`tours`/`launchers`/`tooltips`/`segments`/`embeds`/`event_names`/
`tags`/`deliveries` respectively), `limit: 50` (Chameleon's documented default, also this bundle's
fixed page size), `base.pagination.type: cursor` with `cursor_param: before` and
`token_path: cursor.before` (the real documented next-page token location). `webhooks` declares
`pagination.type: none` and a fixed `kind: webhook` query param — Chameleon's own docs describe its
list endpoint as returning "a complete (un-paginated) list" and require `kind` to select `webhook`
vs `zapier_hook` subscriptions; this bundle covers the `webhook` kind only (the more common
integration shape; `zapier_hook` entries are Zapier-managed and out of scope). `companies`
(`/analyze/companies`) is also cursor-paginated per its documented `cursor` response envelope but
declares no `incremental.cursor_field` — Chameleon's Company schema documents only `created_at`,
no `updated_at`, so there is no real update-tracking timestamp to declare as a cursor.

Every stream with a documented `updated_at` field (`surveys`/`tours`/`launchers`/`tooltips`/
`segments`/`embeds`/`event_names`/`tags`/`deliveries`) declares `incremental.cursor_field:
updated_at` with NO `request_param` and NO `client_filtered` — Chameleon's REST API documents no
server-side "updated since" filter parameter for any of these list endpoints, so every sync (full
or incremental) walks every page unfiltered; the bare `cursor_field` declaration exists only so the
engine derives `incremental_append` sync-mode eligibility from the schema's own cursor field,
matching the one precedent this codebase already established for APIs with no real incremental
filter (see castor-edc/chargedesk's identical bare-`cursor_field`-no-filter shape). Primary key for
every stream is `id`.

`spec.json` intentionally does NOT declare a `limit`/`max_pages` runtime-configurable property
(unlike legacy, which accepts a `config.limit` page-size override and a `config.max_pages`
override): `PaginationSpec.PageSize`/`MaxPages` are read exclusively from `streams.json`'s static
`pagination` JSON literal, never from a `config.*`-templated value (F6, `conventions.md`: a
declared-but-unwireable spec property is worse than an absent one). See Known limits.

## Write actions & risks

Nine write actions were added in Pass B, all gated by approval (`metadata.json`'s
`capabilities.write: true`, `risk.write`). None existed in legacy, which is entirely read-only
(`Write` stub returning `connectors.ErrUnsupportedOperation`) — these are new capability, not a
parity port:

- `publish_survey`/`publish_tour`/`publish_launcher`/`publish_tooltip`/`publish_embed`
  (`PATCH /edit/{surveys,tours,launchers,tooltips,embeds}/:id`) — each sets `published_at` to a
  timestamp to publish the experience live to end-users, or to `null` to unpublish it. This is the
  single mutation every one of Chameleon's 5 experience-type reference pages documents identically
  (the `+`/`-` Environment/Tag-id prefix mutations on the same PATCH endpoint are excluded — see
  `api_surface.json`'s `tags/bulk` exclusion below for the reasoning that generalizes to these too).
- `create_delivery`/`delete_delivery` (`POST`/`DELETE /edit/deliveries`) — `create_delivery`
  directly triggers a Tour or Microsurvey for one specific end-user (by `profile_id`, `uid`, or
  `email`); `delete_delivery` cancels a not-yet-triggered pending Delivery (Chameleon's own docs:
  "once a Delivery is marked as triggered... the delivery can no-longer be... deleted").
- `create_webhook`/`delete_webhook` (`POST`/`DELETE /edit/webhooks`) — creates or removes an
  outbound webhook subscription that POSTs Chameleon event data (tour completions, survey
  responses, etc.) to a caller-supplied HTTPS URL.

Every other documented Chameleon mutation is excluded — see `api_surface.json` for the specific
category+reason per endpoint: bulk/mixed-target tag operations and Delivery updates are
`out_of_scope` (dialect/scope limitations); Company and Profile deletion/clearing endpoints are
`destructive_admin` (irreversible data loss, deliberately excluded from this connector's
conservative write surface).

## Known limits

- **Legacy schemas for the 5 pre-existing streams are preserved.** The newer Chameleon docs show a
  richer/different field set for the experience resources, but the migration target for these
  streams is the legacy Go connector's emitted records. `surveys`/`tours`/`launchers`/`tooltips`
  therefore declare only `id`, `title`, `type`, `state`, `is_live`, `created_at`, and
  `updated_at`; `segments` declares `id`, `name`, `description`, `created_at`, and `updated_at`.
  Expanded Pass B-only streams keep their documented schemas because legacy had no emitted record
  surface for them.
- **Pagination corrected from offset/limit to cursor** (see Overview) — a genuine bug fix. No
  `page_size`/`max_pages` runtime override is exposed (matching every other bundle's F6 rationale);
  every read uses the fixed `limit: 50` page size baked into `streams.json`.
- **`profiles`/`profiles/count` are out of scope**: Chameleon's `/analyze/profiles` is a
  search-only endpoint (no plain unfiltered list, unlike `/analyze/companies`), and its documented
  example response omits the `cursor` envelope every other list endpoint returns — real pagination
  behavior for an unfiltered "list everyone" call is not confidently documented. `/analyze/
  profiles/count` returns a single aggregate integer, not enumerable rows.
- **`demos` has no documented REST list/retrieve endpoint at all**: Chameleon's Product Demos
  reference page documents only schemas and CRM-sync/webhook behavior, with no HTTP Request section
  for any GET endpoint — Demos are Chrome-extension-recorded and consumed via CRM sync or outbound
  webhooks only.
- **`steps` and `elements` have no standalone endpoints**: both are documented explicitly as
  embedded objects within Tours/Microsurveys/Embeddables (`elements`' own doc states "they are not
  listed or retrieved independently").
- Company/Profile deletion and Profile-forget (privacy erasure) endpoints are excluded as
  `destructive_admin` — out of scope for this connector's conservative write surface.
