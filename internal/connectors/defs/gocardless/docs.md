# Overview

GoCardless is a wave2 fan-out declarative-HTTP migration. It reads GoCardless payments, mandates,
payouts, and refunds through the GoCardless REST API (`GET https://api.gocardless.com/...` or the
sandbox equivalent). This bundle migrates `internal/connectors/gocardless` (the hand-written
connector); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a GoCardless access token via the `access_token` secret; it is sent as
`Authorization: Bearer <access_token>` and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`gocardless.go:264`). `base_url` is **required** by this bundle (see
Known limits: legacy's live/sandbox environment-name derivation is not modeled) — set it to
`https://api.gocardless.com` for live or `https://api-sandbox.gocardless.com` for sandbox.
`gocardless_version` defaults to `2015-07-06` (legacy's `gocardlessDefaultVersion`) and is sent as
the `GoCardless-Version` header on every request, matching legacy's
`gocardlessVersion(cfg)`/`DefaultHeaders` exactly.

## Streams notes

All four streams (`payments`, `mandates`, `payouts`, `refunds`) share GoCardless's cursor
pagination (`pagination.type: cursor`, `cursor_param: after`, `token_path: meta.cursors.after`) —
GoCardless's list envelope shape `{"<resource>":[...], "meta":{"cursors":{"after":"<id>", ...}}}`,
matching legacy's `harvest` loop (`gocardless.go:156-198`) exactly: the next page is requested with
`after=<meta.cursors.after>` until that cursor is null/empty (no `stop_path` is declared — legacy
itself only checks the token's own emptiness, never a separate boolean stop signal, so this bundle
matches that exactly by leaving `stop_path` unset). Each stream sends `limit=50`
(`gocardlessDefaultPageSize`) as a static per-stream query value, matching stripe's `limit=100`
static-query precedent (`docs/migration/conventions.md`); `page_size`/`max_pages` runtime
config-driven overrides (`gocardlessPageSize`/`gocardlessMaxPages`, `gocardless.go:324-352`) are
not modeled — see Known limits.

Every stream declares `incremental.request_param: created_at[gt]` with
`start_config_key: start_date`, matching legacy's `incrementalLowerBound` (state cursor first,
falling back to the `start_date` config value; an empty result on the very first full sync means
no filter is sent at all — legacy's own `if createdGT != "" { base.Set("created_at[gt]", createdGT) }`
guard, `gocardless.go:159-161`). `param_format` is left at its `rfc3339` default (GoCardless's
`created_at[gt]` filter takes an RFC3339 timestamp verbatim, exactly what legacy sends).

GoCardless nests every relationship id under a `links` object (e.g.
`{"links":{"mandate":"MD123"}}`). Each stream's foreign-key fields (`payments.mandate`/`payout`,
`mandates.customer_bank_account`/`creditor`, `payouts.creditor`/`creditor_bank_account`,
`refunds.payment`/`mandate`) are derived via `computed_fields` bare single-reference templates
(e.g. `{{ record.links.mandate }}`), reproducing legacy's `linkField(item, "mandate")` primary path
(the `links` object is present and holds the key) exactly — see Known limits for the one
`linkField` behavior this bundle does not reproduce.

## Write actions & risks

None. GoCardless is read-only here (legacy's own `Capabilities.Write: false`, `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`gocardless_environment`-based live/sandbox base-URL derivation is dropped; `base_url` is
  required instead.** Legacy derives the live-vs-sandbox host from a `gocardless_environment`
  config value (`"live"` -> `gocardlessLiveBaseURL`, anything else -> sandbox,
  `gocardless.go:301-322`) when `base_url` itself is unset. The engine's `spec.json` `"default"`
  materialization mechanism (conventions.md §3) only supports a FIXED literal default, not one
  derived from another config key's value, so this cross-key derivation has no engine-side
  mechanism today (the identical documented limitation as chargebee's `site`-based URL and
  sentry's `hostname`-based URL). `base_url` is declared `required` in `spec.json` instead; the
  caller must supply the fully-formed host. ACCEPTABLE per conventions.md §5's meta-rule: this
  narrows the config surface (no config shape is silently mishandled) rather than changing any
  accepted input's emitted record data.
- **`linkField`'s top-level-field fallback is not modeled.** Legacy's `linkField(item, key)`
  prefers `item.links[key]` but falls back to a top-level `item[key]` scalar when the `links`
  object doesn't carry that key (`gocardless.go:182-189`). GoCardless's real, documented wire shape
  always nests every relationship id under `links` for these four resources (confirmed against
  legacy's own fixture-mode literal, `gocardless.go:226-233`, which always populates `links` for
  every relationship field this bundle reads) — the fallback branch is defensive-only and never
  fires for a genuine API response. This bundle's `computed_fields` reference `record.links.<key>`
  directly with no fallback; ACCEPTABLE per conventions.md §5's meta-rule (no real GoCardless input
  this bundle would receive exercises the fallback path differently).
- **`page_size`/`max_pages` are not runtime-configurable.** The engine's `cursor` (token_path)
  paginator has no config-driven page-size or request-count-cap knob (unlike
  `page_number`/`offset_limit`'s `PageSize`, there is no equivalent field the token_path variant
  reads at all); `limit=50` is a fixed per-stream static query value matching legacy's own default
  exactly, and pagination is bounded only by the empty-cursor stop signal, matching GoCardless's
  own real termination behavior when neither override is set.
- Full GoCardless API surface (customers, subscriptions, events, creditors, payment
  cancellation) is out of scope for wave2; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}` entries.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) stamps extra fields (`connector`, `fixture`,
  `previous_cursor`) onto every record (`gocardless.go:200-244`); this bundle's schemas and
  fixtures target the LIVE record shape only, per convention.
