# Overview

Snapchat Marketing is a fresh Tier-2 (AuthHook + `fan_out`) migration at legacy capability parity,
porting `internal/connectors/snapchat-marketing` (`snapchat_marketing.go` + `streams.go` +
`auth.go`). It reads the authenticated user's organizations, ad accounts (under configured
organizations), and campaigns/ad squads/ads (under configured ad accounts) through the Snapchat Ads
API v1. Snapchat Marketing authenticates with a short-lived bearer access token obtained by
exchanging a long-lived `refresh_token` (plus `client_id`/`client_secret`) at
`https://accounts.snapchat.com/login/oauth2/access_token` — an OAuth 2.0 **refresh-token grant**,
which the engine's built-in `oauth2_client_credentials` auth mode cannot express (that mode only
performs a client-credentials grant, never `grant_type=refresh_token`). This is the same shape as
`internal/connectors/hooks/strava`/`gmail`'s pilot AuthHooks; `hooks/snapchat-marketing/hooks.go`
ports the pattern almost verbatim for Snapchat's own token endpoint/field names. Read-only: legacy's
`Write` always returns `ErrUnsupportedOperation` (`snapchat_marketing.go:148-150`) since the Snapchat
Ads API exposes no obvious safe reverse-ETL writes here, and this bundle declares
`capabilities.write: false` with no `writes.json` to match. The legacy package stays registered and
unchanged until the wave6 registry flip.

## Auth setup

Provide three secrets: `client_id`, `client_secret`, and `refresh_token` (long-lived; never logged).
`hooks/snapchat-marketing/hooks.go` implements `AuthHook`, mirroring legacy `auth.go`'s
`refreshTokenAuth`: it POSTs `grant_type=refresh_token` + `refresh_token` + `client_id` +
`client_secret` to `token_url` (default
`https://accounts.snapchat.com/login/oauth2/access_token`, config-overridable), caches the resulting
access token until 60 seconds before its declared expiry (falling back to a 1-hour TTL when the
token response carries no `expires_in`, matching legacy's `auth.go:104-108` exactly), and sets
`Authorization: Bearer <access_token>` on every request.

`token_url` and `base_url` are validated for a well-formed `http(s)://` URL with a host (matching
legacy's `validateHTTPURL`, `snapchat_marketing.go:381-393`, which accepts plain `http` for local
test servers as well as `https`) before any network access — this bounds SSRF risk on both
overridable URLs exactly like legacy.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook":
"snapchat-marketing", ...}` — legacy has no alternate auth path (no static API key, no
public/no-auth fallback), matching strava/gmail's identical single-candidate shape.

`organization_ids` (plain config, comma-separated, not a secret) is required only for the
`adaccounts` stream; `ad_account_ids` (same shape) is required only for the `campaigns`/`adsquads`/
`ads` streams — matching legacy's `streamPaths`' per-scope id-list requirement
(`snapchat_marketing.go:232-259`), which errors only when reading a stream that actually needs the
missing id list, not globally.

## Streams notes

Five streams, all primary-keyed on `id`, all pagination via `paging.next_link` (an absolute
next-page URL read from the response body — `pagination.type: next_url`), matching legacy's
`harvest`'s exact convention (`snapchat_marketing.go:152-198,38`):

- `organizations` (`GET /organizations`, top-level, no fan-out) — every organization the
  authenticated user can access.
- `adaccounts` (`GET /organizations/{organization_id}/adaccounts`) — fans out over
  `organization_ids` (comma-separated config, `fan_out.ids_from.config_key`), substituting each id
  into the path via `{{ fanout.id }}` (`fan_out.into.path_var`) — matching legacy's `streamPaths`
  `"adaccounts"` scope, which builds one request path per configured organization id
  (`snapchat_marketing.go:236-245`).
- `campaigns`/`adsquads`/`ads` (`GET /adaccounts/{ad_account_id}/<resource>`) — each fans out over
  `ad_account_ids` the same way, matching legacy's `"adaccount"` scope
  (`snapchat_marketing.go:246-255`).

**Every list response wraps its array under a plural key and wraps each element under a singular
envelope key** (Snapchat's own convention, e.g. `{"campaigns":[{"sub_request_status":"SUCCESS",
"campaign":{...}}]}`) — `records.path` selects the plural array (`"campaigns"`), but the raw array
elements are still envelope-wrapped; this bundle uses the default **schema-mode projection**
(no raw top-level key ever matches a declared schema property, since every real field lives one
level deeper under the singular envelope key) plus a full `computed_fields` map of bare
`{{ record.<singular>.<field> }}` references per stream — e.g. `"id": "{{ record.campaign.id }}"` —
which both unwraps the envelope AND (per the typed-extraction rule, conventions.md §3) preserves
each field's native JSON type for numeric fields (`daily_budget_micro`, `lifetime_spend_cap_micro`,
`bid_micro`), reproducing legacy's `unwrapEnvelopes` + per-stream `mapRecord` exactly
(`snapchat_marketing.go:200-228`, `streams.go:157-231`) with zero envelope keys
(`campaign`/`sub_request_status`/etc.) or extraneous fields leaking into the emitted record.

No stream declares an `incremental` block — legacy's `harvest` never sends a server-side
incremental filter parameter on any of the five streams (matches conventions.md §8 rule 2's truth
table: `x-cursor-field` in schema only, no `incremental` block, since legacy publishes
`CursorFields: []string{"updated_at"}` for every stream, `streams.go`, but never wires a
request-side filter for it).

## Write actions & risks

None — Snapchat Marketing is read-only. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`snapchat_marketing.go:148-150`): the Snapchat Ads API exposes no
obvious safe reverse-ETL write surface for this connector's read-only scope.

## Known limits

- **`TestConformance/snapchat-marketing`'s dynamic (fixture-replay) checks are genuinely
  `skip_dynamic`'d, for the identical reason strava's/gmail's are** (see
  `internal/connectors/defs/strava/docs.md`'s Known limits): the bundle's *sole* auth candidate is
  `mode: custom`, and conformance's synthetic config can never carry a real refresh token that
  round-trips through a live (or even a fixture-replayed) OAuth token exchange — the AuthHook always
  attempts a real HTTP POST to `token_url` to mint an access token, which conformance's
  static-fixture replay harness has no mechanism to intercept for a non-declarative auth path. Every
  auth-resolving dynamic check would therefore fail identically and uninformatively regardless of
  hook wiring. `paritytest/snapchat-marketing` (which wires the real `AuthHook` via
  `engine.HooksFor("snapchat-marketing")`, matching strava/gmail's precedent) is the authoritative
  parity/correctness bar for this connector's auth + read path.
- **No config-driven `page_size`/`max_pages` runtime override**. Legacy hard-codes a `limit=50`
  query param and bounds its pagination loop at `snapchatMaxPages = 1000`
  (`snapchat_marketing.go:35,157-158`) with no caller-supplied override for either. The engine's
  `next_url` pagination spec has no `page_size`/`max_pages` fields at all (unlike `page_number`'s
  static integers) — there is no per-request config-driven override mechanism, identical to
  searxng's/strava's documented `page_size`/`max_pages` gap (conventions.md §1). Neither property is
  declared in `spec.json`; the bundle hard-codes legacy's own static `limit=50` per-stream query
  param, matching legacy's behavior exactly (legacy also never exposed either as config).
- **`adaccounts`/`campaigns`/`adsquads`/`ads` reading without the relevant id-list config
  (`organization_ids`/`ad_account_ids`) is a DOCUMENTED PARITY DEVIATION, not identical to legacy.**
  Legacy's `streamPaths` hard-errors when the relevant id list is unset
  (`snapchat_marketing.go:239,249`). The engine's `fan_out.ids_from.config_key` resolves an
  empty/absent config value to a zero-length id list (`read.go`'s `splitTrimmedCSV`) and silently
  emits ZERO records (no sub-sequence ever runs) rather than erroring — see
  `paritytest/snapchat-marketing`'s `TestParitySnapchatMarketing_MissingAdAccountIDs`. ACCEPTABLE
  per conventions.md §5's meta-rule: this never changes emitted-record DATA for any legacy-accepted
  input (both sides require the id list to emit any records at all); it only changes the
  empty-config-value failure mode from a hard error to a silent empty read.
