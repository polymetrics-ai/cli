# Overview

Reads Facebook Marketing ad accounts, campaigns, and ads through the Graph API. This bundle migrates
`internal/connectors/facebook-marketing/facebook_marketing.go` (the legacy hand-written connector,
which stays registered and unchanged until wave6's registry flip). The catalog's inventory
(`docs/migration/inventory.json`) labels this connector `runtime_kind: "native_go"`, but the legacy
Go source is the ground truth and it is a plain `connsdk.Requester`-based HTTP connector, not a
protocol-native SDK client — there is no SQL/queue/custom-binary protocol here, only `GET` requests
against the Graph API's REST surface with Bearer auth and an absolute-next-page-URL pagination
convention. It is also read-only: `Capabilities.Write` is `false` and `Write()` unconditionally
returns `connectors.ErrUnsupportedOperation`.

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

All three streams share the same `limit=100` (matching legacy's `facebookDefaultPageSize`) and the
same pagination convention: Facebook's Graph API returns a fully-qualified absolute next-page URL at
`paging.next`, so `base.pagination` declares `next_url`/`next_url_path: "paging.next"` once for every
stream (legacy's `harvest`, `facebook_marketing.go:113-146`, follows `paging.next` verbatim the same
way for every endpoint). An absent/empty `paging.next` stops pagination, matching legacy's own
`strings.TrimSpace(next) == ""` stop check.

`campaigns`/`ads` require `config.ad_account_id` — an unresolved reference hard-errors exactly like
legacy's own explicit check (`"facebook-marketing connector requires config ad_account_id for this
stream"`); the engine's failure is an unresolved-key error naming `ad_account_id` instead, the same
config-validation-parity-by-classification-not-exact-text precedent already accepted for postgres
(parity-deviation ledger entry 9, `docs/migration/conventions.md` §5).

Neither stream is incremental in legacy (no cursor field, no server-side created/updated filter) —
no schema declares `x-cursor-field`.

## Write actions & risks

None. Legacy `facebook_marketing.go` is read-only: `Capabilities.Write: false`,
`Write()` always returns `connectors.ErrUnsupportedOperation`. `capabilities.write` is `false` and
this bundle ships no `writes.json`, despite the catalog inventory's `runtime_kind: native_go` label
(see "Overview" — the label describes legacy's original hand-written-Go implementation style, not a
non-REST protocol or a write capability; neither is true of the actual source).

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
