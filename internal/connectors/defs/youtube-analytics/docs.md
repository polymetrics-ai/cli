# Overview

YouTube Analytics is a Tier-2 (AuthHook) migration of `internal/connectors/youtube-analytics`
(legacy `youtubeanalytics.Connector`). It reads the YouTube Reporting API's three core
control-plane resources — scheduled reporting jobs, the report types available to schedule, and
the generated reports for a job — via the Google OAuth 2.0 **refresh-token grant** only; the
3-legged consent/acquisition dance is out of scope (the refresh token arrives as a pre-issued
secret; the credentials layer already owns acquisition/storage), matching gmail's precedent
exactly. This bundle is engine-vs-legacy parity-tested against `internal/connectors/youtube-analytics`
(the hand-written connector it migrates); the legacy package stays registered and unchanged until
wave6's registry flip. Read-only: legacy `youtube_analytics.go:172-174` always returns
`ErrUnsupportedOperation` from `Write`, and this bundle declares `capabilities.write: false` with
no `writes.json` to match.

## Auth setup

Provide three secrets: `client_id`, `client_secret` (optional — some Google OAuth client types
issue refresh tokens that don't require a client secret at token-refresh time), and
`refresh_token` (long-lived; never logged). `hooks/youtube-analytics/hooks.go` implements
`AuthHook`, porting legacy `auth.go`'s `oauthRefreshAuth` almost verbatim (byte-for-byte the same
shape gmail's hook already ports): it POSTs `grant_type=refresh_token` + `refresh_token` +
`client_id` [+ `client_secret`, `scope`] to `token_url` (default
`https://oauth2.googleapis.com/token`, config-overridable), caches the resulting access token
until 60 seconds before its declared expiry, and sets `Authorization: Bearer <access_token>` on
every request. `client_secret` is omitted from the token-request form when unset (matches legacy's
`if a.clientSecret != ""` guard, `auth.go:85-86`).

`token_url` MUST resolve to an `https://` URL (THREAT-MODEL.md Delta 2, gmail precedent): the hook
fails closed on a non-https or unparseable override rather than sending the refresh token/client
secret to an attacker-chosen endpoint. This is stricter than legacy's own `validatedURL`
(`youtube_analytics.go:280-296`), which also accepted plain `http` — documented as a parity
deviation in Known limits below (never stricter for any real Google OAuth endpoint, which is
always https).

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook":
"youtube-analytics", ...}` — legacy has no alternate auth path (no static API key, no
public/no-auth fallback), so there is no `when`-gated bypass to declare, matching gmail's shape
exactly (unlike github's token-OR-app_jwt "auto" resolution).

## Streams notes

Three streams, all primary-keyed on `id`, all paginated via the Reporting API's
`pageToken`/`nextPageToken` cursor convention (`pagination.type: cursor`, `cursor_param:
pageToken`, `token_path: nextPageToken`) — identical to legacy's shared
`connsdk.CursorPaginator` wiring (`youtube_analytics.go:159-166`):

- `jobs` (`GET /jobs`) — scheduled reporting jobs for the channel or content owner.
- `report_types` (`GET /reportTypes`) — report types available to be scheduled as jobs.
- `reports` (`GET /jobs/{{ config.job_id }}/reports`) — generated reports for a job; requires the
  `job_id` config value (matches legacy's `resolveResource`, `youtube_analytics.go:241-250`, which
  hard-errors when `job_id` is unset for this stream — the same failure mode this bundle produces
  via a plain, non-`omit_when_absent` `{{ config.job_id }}` path template: an absent `job_id` is a
  hard interpolation error, not a silently-wrong request).

Every stream sends `pageSize` per `config.page_size` (default 100) and, when `config.content_owner_id`
is set, `onBehalfOfContentOwner` (an `omit_when_absent`-dialect query entry, conventions.md §3 —
sent only when configured, exactly matching legacy's `applyContentOwner`,
`youtube_analytics.go:254-258`, which adds the query param conditionally rather than sending it
empty).

`computed_fields` rename each stream's camelCase raw fields to the schema's snake_case names
(`reportTypeId` -> `report_type_id`, `createTime` -> `create_time`, `systemManaged` ->
`system_managed`, etc.), matching legacy's `jobRecord`/`reportTypeRecord`/`reportRecord`
(`streams.go:90-120`) field-for-field.

**No incremental sync mode** (sentry/gmail precedent): legacy publishes `CursorFields` on the
catalog surface (`create_time` for `jobs`/`reports`) but `Read` never applies a state-based filter
anywhere — there is no `incrementalLowerBound`-equivalent call in `youtube_analytics.go`'s `Read`,
and `InitialState` always seeds an empty cursor. This bundle matches that exactly by declaring
**no `incremental` block on any stream** (full_refresh only); adding one would be new, unrequested
capability, not a migration.

## Write actions & risks

None — YouTube Analytics is read-only. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`youtube_analytics.go:172-174`).

## Known limits

- **`token_url` https-only enforcement is stricter than legacy's `validatedURL`** (which accepted
  plain `http` too, `youtube_analytics.go:280-296`): the hook only accepts `https://` overrides.
  Documented as a parity deviation (never stricter for any production Google OAuth endpoint, which
  is always https; strictly safer for the one new SSRF-adjacent secret-bearing surface this hook
  introduces). See the parity-deviation ledger in `docs/migration/conventions.md` §5.
- **`base_url` https-only enforcement is NOT added** — unlike `token_url`, the Reporting API host
  itself (`config.base_url`) has no dedicated hook-side guard; it flows straight into
  `streams.json`'s `base.url` template and is resolved by the engine's ordinary requester
  construction, exactly matching legacy's own `baseURL`/`validatedURL` (which accepts http OR
  https for the data-plane host, only `token_url`'s secret-bearing exchange gets the stricter
  guard in legacy too). No deviation here.
- **Bundle-level `skip_dynamic` marker** (matches gmail exactly, conventions.md §4): this bundle's
  *sole* auth candidate is `mode: custom`, and conformance's synthetic config can never carry a
  real `https` `token_url` — the AuthHook's own https-only guard means no synthetic secret value
  can ever satisfy it, so every auth-resolving dynamic check (`check_fixture`, every
  `read_fixture_nonempty:<stream>`, `pagination_terminates`, `records_match_schema`,
  `cursor_advances`) would fail identically and uninformatively regardless of hook wiring.
  `internal/connectors/paritytest/youtube-analytics` (which wires the real `AuthHook` via
  `engine.HooksFor("youtube-analytics")`) is the authoritative parity/correctness bar for this
  connector's auth + read path — `TestConformance/youtube-analytics` passes via the marker-skip
  path, not a bypassed/expected-fail path.
- Full YouTube Reporting API surface (job creation/deletion, bulk report CSV/GZIP media download,
  the separate YouTube Analytics querying API) is out of scope for this pass; see
  `api_surface.json`'s `excluded` entries. Only the 3 legacy-parity read streams are implemented.
