# Overview

Reads Facebook Marketing ad accounts, campaigns, ads, ad sets, ad creatives, custom audiences, and
performance insights, and creates/updates campaigns and ad sets, through the Graph API. This bundle
originally migrated `internal/connectors/facebook-marketing/facebook_marketing.go` (the legacy
hand-written connector, which stays registered and unchanged until wave6's registry flip) as a
read-only bundle, and has since been expanded (Pass B full-surface pass) with 4 new read streams
(`ad_sets`, `ad_creatives`, `custom_audiences`, `insights`) and 3 new write actions
(`create_campaign`, `update_campaign`, `create_ad_set`). The catalog's inventory
(`docs/migration/inventory.json`) labels this connector `runtime_kind: "native_go"`, but the legacy
Go source is the ground truth and it is a plain `connsdk.Requester`-based HTTP connector, not a
protocol-native SDK client — there is no SQL/queue/custom-binary protocol here, only Graph API REST
calls with Bearer auth and an absolute-next-page-URL pagination convention. **Legacy itself is
read-only** (`Capabilities.Write` is `false`, `Write()` unconditionally returns
`connectors.ErrUnsupportedOperation`); the write actions this bundle now declares are a genuinely
NEW capability sourced directly from the documented Marketing API, not a migration of any existing
legacy write path — there is no legacy write behavior to preserve parity with.

**Tier justification**: every legacy behavior (auth, path construction, pagination, record mapping)
is expressible in `streams.json`/`spec.json` alone — bearer auth, `next_url` pagination reading
`paging.next`, and per-stream `fields`/`limit` query params all have direct dialect equivalents (see
below). This is a pure **Tier 1** declarative bundle: zero Go, matching ≥90% of connectors per
`docs/migration/conventions.md` §1's target.

## Auth setup

Provide a Facebook Graph API `access_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <access_token>`) and never logged, matching legacy's
`connsdk.Bearer(token)` (`facebook_marketing.go:179`). `base_url` defaults to
`https://graph.facebook.com/v20.0` (legacy's `facebookDefaultBaseURL`) and may be overridden for
tests/proxies.

## Streams notes

- **`ad_accounts`** — `GET /me/adaccounts`, not account-scoped (legacy's
  `facebookStreamEndpoints["ad_accounts"].accountScoped == false`). Records at `data`, fields
  `id,account_id,name,account_status,currency,timezone_name`.
- **`campaigns`** — `GET /{{ config.ad_account_id }}/campaigns`, account-scoped. Records at `data`,
  fields `id,name,status,effective_status,objective,created_time,updated_time`.
- **`ads`** — `GET /{{ config.ad_account_id }}/ads`, account-scoped. Records at `data`, fields
  `id,name,status,effective_status,created_time,updated_time`.
- **`ad_sets`** (new) — `GET /{{ config.ad_account_id }}/adsets`, account-scoped. Records at
  `data`, fields `id,name,campaign_id,status,effective_status,daily_budget,lifetime_budget,
  billing_event,optimization_goal,bid_amount,start_time,end_time,created_time,updated_time`
  (the standard Ad Set object fields — Graph API reference: `developers.facebook.com/docs/
  marketing-api/reference/ad-campaign/`).
- **`ad_creatives`** (new) — `GET /{{ config.ad_account_id }}/adcreatives`, account-scoped.
  Records at `data`, fields `id,name,object_story_id,object_type,thumbnail_url,status`.
- **`custom_audiences`** (new) — `GET /{{ config.ad_account_id }}/customaudiences`,
  account-scoped. Records at `data`, fields `id,name,subtype,description,
  approximate_count_lower_bound,approximate_count_upper_bound,operation_status,time_created,
  time_updated`.
- **`insights`** (new) — `GET /{{ config.ad_account_id }}/insights`, account-scoped, `level=ad`
  (the finest-grained level; `campaign_id`/`adset_id`/`ad_id` are all present on every row, so
  campaign- or adset-level aggregates are derivable downstream without a separate stream per
  level — see `api_surface.json`'s `duplicate_of` exclusions for the per-level insights edges).
  `date_preset=last_30d` is a static default (Facebook's own GET-with-no-params default window);
  no incremental cursor is declared — the Insights API's real "freshness" semantics (attribution
  windows causing already-fetched date ranges to change retroactively) do not map cleanly onto the
  engine's `incremental.cursor_field`/`request_param` model, so this stream is deliberately
  full-refresh only. Insight rows have no natural `id` field of their own (Facebook does not assign
  one); `computed_fields` synthesizes one via `"id": "{{ record.ad_id }}{{ record.date_start }}"`
  (a mixed multi-reference template, so per the typed-extraction rule this always stringifies,
  which is exactly what a synthetic composite key needs).

All streams share the same `limit=100` (matching legacy's `facebookDefaultPageSize`) and the
same pagination convention: Facebook's Graph API returns a fully-qualified absolute next-page URL at
`paging.next`, so `base.pagination` declares `next_url`/`next_url_path: "paging.next"` once for every
stream (legacy's `harvest`, `facebook_marketing.go:113-146`, follows `paging.next` verbatim the same
way for every endpoint it covers; the 4 new streams follow the identical convention, confirmed
against current Marketing API documentation independent of legacy, which never implemented them). An
absent/empty `paging.next` stops pagination, matching legacy's own `strings.TrimSpace(next) == ""`
stop check.

`campaigns`/`ads`/`ad_sets`/`ad_creatives`/`custom_audiences`/`insights` all require
`config.ad_account_id` — an unresolved reference hard-errors exactly like legacy's own explicit
check (`"facebook-marketing connector requires config ad_account_id for this stream"`) for the 2
streams legacy covers; the engine's failure is an unresolved-key error naming `ad_account_id`
instead, the same config-validation-parity-by-classification-not-exact-text precedent already
accepted for postgres (parity-deviation ledger entry 9, `docs/migration/conventions.md` §5).

No stream is incremental (no cursor field, no server-side created/updated filter) — no schema
declares `x-cursor-field`. This matches legacy exactly for `ad_accounts`/`campaigns`/`ads`; for the
4 new streams it is a genuine API property (Graph API list edges support cursor-based *pagination*
via `paging.next`, but not a server-side "updated since" filter parameter), not a narrowing of an
existing capability.

## Write actions & risks

`capabilities.write` is now `true` (Pass B expansion added `writes.json`). All 3 actions use
`body_type: "form"` — Facebook's Graph API accepts form-encoded POST bodies universally, including
for array/object-valued fields (`special_ad_categories`, `targeting`), by JSON-stringifying the
field value; `write.go`'s `buildForm`/`stringifyAny` already does exactly this (JSON-marshal any
non-string record value before setting the form field), so no new engine behavior was needed — this
is the same pattern facebook-pages' `create_post` write action already established for this API
family.

- `create_campaign` — `POST /{{ config.ad_account_id }}/campaigns`. Required: `name`, `objective`,
  `status`, `special_ad_categories` (send `[]` when the campaign has no special ad category
  declaration — Meta requires the field to be present even when empty). `status: "PAUSED"` is the
  safe default Meta's own quickstart guide recommends while a campaign's ad sets/ads are still
  being assembled.
- `update_campaign` — `POST /{{ record.id }}` (path_fields: `id`; Facebook's Graph API updates a
  node in place via `POST` to the node's own id, not a nested collection path — there is no
  separate PUT/PATCH verb for this API). Commonly used to pause/resume spend (`status`) or adjust
  budget.
- `create_ad_set` — `POST /{{ config.ad_account_id }}/adsets`. Required: `name`, `campaign_id`,
  `billing_event`, `optimization_goal`, `targeting`, `status`.

All 3 actions carry a `risk` noting they mutate a LIVE ad account and can incur real ad spend once
an ad set/campaign is active with ads attached — approval-gated. Ad/ad-creative creation, ad-set
update, and campaign/ad-set deletion are deliberately NOT included in this pass; see
`api_surface.json` for the specific reason each was held back (mostly: undocumented-beyond-a-skeleton
required-field contracts for creative/ad payloads, or destructive-admin deletes deferred pending a
dedicated review).

## Known limits

- **`ad_account_id`'s `act_` prefix auto-completion is not modeled.** Legacy's `facebookResource`
  (`facebook_marketing.go:215-227`) prepends the literal `"act_"` to a configured `ad_account_id`
  that doesn't already start with it. The engine's declarative path templating
  (`stream.Path`/`InterpolatePath`) substitutes `{{ config.ad_account_id }}` verbatim with no
  conditional-prefix/string-transform primitive (unlike `urlencode`/`base64`/`join:<sep>`, none of
  the dialect's filters perform a conditional literal-prepend). This bundle's `spec.json` documents
  that `ad_account_id` must be supplied already including the `act_` prefix. This is a config-SHAPE
  narrowing (a caller who previously configured a bare numeric id must now include the prefix), never
  an emitted-record-DATA change for any config value already in the prefixed form (which legacy
  itself also accepted unchanged, per its own `!strings.HasPrefix(accountID, "act_")` guard).
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-500,
  default 100) and `max_pages` (0/unlimited or a positive cap) as config-driven overrides
  (`facebookPageSize`/`facebookMaxPages`). The engine's `next_url` paginator has no config-driven
  page-size or request-count-cap knob at all (the identical, already-ledgered limitation documented
  on aircall's and bitly's bundles, this repo's other `next_url` migrations); `page_size`/`max_pages`
  are therefore not declared in `spec.json`, and this bundle sends Facebook's own default
  (`limit=100`) as a static per-stream query literal. `max_pages` is left unbounded
  (`base.pagination` declares no `max_pages`), matching legacy's own default (unset/`all`/`unlimited`
  configs all resolve to `0` = unbounded in `maxPagesConfig`).
- **`limit`/`fields` are re-sent on every page request, unlike legacy's reset-to-nil-after-page-1.**
  Legacy's `harvest` (`facebook_marketing.go:113-146`) sets `query = nil` once it starts following an
  absolute `next` URL, relying on Facebook's own `next` URL to carry every needed query parameter.
  The engine's `readDeclarative` loop instead merges `stream.Query` (here, `limit`+`fields`) onto
  EVERY page request unconditionally, and `connsdk.Requester`'s URL resolution re-applies that merged
  query onto the absolute next-page URL (replacing any same-named param already present) — the
  identical, already-ledgered divergence documented on aircall's and bitly's `next_url` bundles this
  wave. This is benign in DATA terms only because Facebook's own `paging.next` URL already carries
  the identical `limit`/`fields` values the engine re-applies (the replace is idempotent); if a
  future Graph API version ever varied `fields` mid-pagination, this bundle's request would diverge
  from legacy's — today it does not.
- **`next_url` fixtures are single-page, per the sanctioned exception (conventions.md §4).** A
  `next_url` stream's next-page URL is the replay server's own runtime address, unknown until the
  harness picks a port — a static fixture file cannot embed the correct absolute URL for a second
  page. Every stream in this bundle ships a single-page fixture (satisfies `fixtures_present`/
  `read_fixture_nonempty`); `pagination_terminates` passes on the first stream (`ad_accounts`) with
  its single page (`hits == len(pages) == 1`), which is not a 2-page pagination proof but is not a
  false failure either. Real 2-page `next_url`-following correctness for this exact request shape is
  proven by legacy's own existing test (`internal/connectors/facebook-marketing/
  facebook_marketing_test.go`), plus the engine's own generic `next_url` paginator unit tests
  (`internal/connectors/engine/paginate_test.go`) and read-path integration test
  (`internal/connectors/engine/read_test.go`'s `TestReadNextURLPaginationSetsBaseHostFromRequester`).
  This wave does not add a new `paritytest/facebook-marketing` package (out of scope per this wave's
  JSON-only mandate, matching aircall's identical decision); a future wave adding hand-written parity
  suites should follow bitly's/calendly's `TestParity<Name>_..._TwoPagePagination` pattern.
- **`insights` has no natural incremental cursor and no legacy behavior to compare against** (it is
  a genuinely new stream). It is intentionally full-refresh only; see "Streams notes" above for why
  the Insights API's attribution-window semantics don't map onto `incremental.cursor_field`.
- **Write actions are unproven against a live ad account** (there is no legacy write path this
  bundle's writes could be parity-tested against, unlike every read stream). `create_campaign`/
  `update_campaign`/`create_ad_set`'s required-field lists and `body_type: "form"` JSON-stringify
  behavior are sourced from current Marketing API documentation and the community's public
  quickstart guides, not from a legacy Go implementation — flagged here per conventions.md's
  Known-limits requirement to record exactly this kind of non-parity-provable new capability.
- **Ad and ad-creative creation are not covered** (see `api_surface.json`): both require a
  fully-specified nested payload (`adcreatives` needs `object_story_spec` with a Page id and
  link/image/video reference; `ads` needs a pre-existing creative id plus full targeting) whose
  complete required-field contract is scattered across several product-specific guides rather than
  one canonical reference — deferred rather than guessing an incomplete contract that would fail
  silently-but-confusingly against the real API.
