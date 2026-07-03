# Overview

Strava is a fresh Tier-2 (AuthHook) migration at legacy capability parity, porting
`internal/connectors/strava` (`strava.go` + `streams.go`). It reads the authenticated athlete's
profile, activities, lifetime stats, and clubs through the Strava v3 REST API. Strava authenticates
with a short-lived bearer access token obtained by exchanging a long-lived `refresh_token` (plus
`client_id`/`client_secret`) at `https://www.strava.com/oauth/token` — an OAuth 2.0
**refresh-token grant**, which the engine's built-in `oauth2_client_credentials` auth mode cannot
express (that mode only performs a client-credentials grant, never `grant_type=refresh_token`).
This is the same shape as `internal/connectors/hooks/gmail`'s pilot AuthHook; `hooks/strava/hooks.go`
ports gmail's `oauthRefreshAuth` pattern almost verbatim for Strava's own token endpoint/field names.
Read-only: legacy's `Write` always returns `ErrUnsupportedOperation` (`strava.go:97-99`) since the
Strava API exposes no obvious safe reverse-ETL writes, and this bundle declares `capabilities.write:
false` with no `writes.json` to match. The legacy package stays registered and unchanged until the
wave6 registry flip.

## Auth setup

Provide three secrets: `client_id`, `client_secret`, and `refresh_token` (long-lived; never logged).
`hooks/strava/hooks.go` implements `AuthHook`, mirroring legacy `strava.go`'s `refreshTokenAuth`: it
POSTs `grant_type=refresh_token` + `refresh_token` + `client_id` + `client_secret` to `token_url`
(default `https://www.strava.com/oauth/token`, config-overridable), caches the resulting access
token until 60 seconds before its declared expiry (falling back to a 6-hour TTL when the token
response carries neither `expires_in` nor `expires_at`, matching legacy's `strava.go:418-424`
exactly), and sets `Authorization: Bearer <access_token>` on every request.

`token_url` and `base_url` are validated for a well-formed `http(s)://` URL with a host (matching
legacy's `resolveHTTPURL`, `strava.go:494-510`, which accepts plain `http` for local test servers as
well as `https`) before any network access — this bounds SSRF risk on both overridable URLs exactly
like legacy.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook": "strava", ...}`
— legacy has no alternate auth path (no static API key, no public/no-auth fallback), matching
gmail's identical single-candidate shape.

`athlete_id` (plain config, not a secret) is required only for the `athlete_stats` stream, whose
resource path substitutes it directly (`/athletes/{{ config.athlete_id }}/stats`) — matching
legacy's `resolveResource`'s `{athlete_id}` placeholder substitution (`strava.go:430-442`), which
also only requires `athlete_id` when reading that specific stream, not globally.

## Streams notes

Four streams, all primary-keyed on `id`: `activities` (`/athlete/activities`, `page`/`per_page`
`page_number` pagination starting at page 1 with a 100-record page size, stopping on a short/empty
final page — legacy's `stravaDefaultPageSize`/`harvest`, `strava.go:37,185-214`) and `clubs`
(`/athlete/clubs`, identical pagination shape) are list endpoints; `athlete` (`/athlete`) and
`athlete_stats` (`/athletes/{athlete_id}/stats`) are singleton endpoints returning ONE JSON object
(`records.path: "."` + `single_object: true`, `pagination: {"type": "none"}` stream-level override)
— matching legacy's `endpoint.list == false` routing (`streamEndpoint.list`, `streams.go:13-14`,
dispatched via `readSingleton`, `strava.go:147-148,163-181`).

`activities`' incremental cursor field is `start_date` (`x-cursor-field`), matching legacy's
`CursorFields: []string{"start_date"}` (`streams.go:36`); `athlete`/`athlete_stats`/`clubs` declare
no cursor field, matching legacy's omission of `CursorFields` for those three streams
(`streams.go:39-57`). No stream declares an `incremental` block — legacy's `Read` never sends a
server-side incremental filter parameter on any of the four streams (matches conventions.md §8 rule
2's truth table: `x-cursor-field` in schema only, no `incremental` block, since legacy publishes the
cursor field but never wires a request-side filter for it).

Every field name on every stream's raw Strava API response already matches its schema property name
exactly (`sport_type`, `start_date`, `moving_time`, etc. — Strava's own wire shape is already
snake_case) — plain schema-mode projection copies every field by exact key match with **zero**
`computed_fields` needed for `activities`/`athlete`/`clubs`, preserving each field's native JSON type
(numbers stay numbers, booleans stay booleans), matching legacy's field-built
`connectors.Record{...}` mapping exactly (`streams.go:128-193`).

**`athlete_stats`' `id` field is synthesized, not copied from the raw response** (legacy's
`injectAthleteID`, `strava.go:444-461`): Strava's `athletes/{id}/stats` response carries no `id`
field of its own (the API's only identifier for that record is the athlete id in the request path),
so legacy stamps `athlete_id` (parsed as an int64 when numeric, else left as the raw string) onto
the record only when the raw object has no pre-existing `id` key — which is always true for this
endpoint's real response shape. This bundle's `athlete_stats` stream declares a single
`computed_fields` entry, `"id": "{{ config.athlete_id }}"`, stamping the SAME config-scoped value
(gap-loop cycle-1 item 2, conventions.md §3: `config.*` is wired into `computed_fields`' Vars
specifically for this "config value stamped onto every record" pattern, e.g. github's
`repository` field). The stamped value is always a STRING here (`computed_fields`' Interpolate
return type — no numeric-typed extraction path exists for a `config.*`-only reference, unlike a bare
`record.*` reference), whereas legacy stores it as an int64 when `athlete_id` parses as an integer;
schema types `["integer", "string"]` on `athlete_stats.id` (like `activities`/`athlete`/`clubs`'
`id`) tolerate either representation. This is documented as a parity deviation (see Known limits):
never data-changing for any legacy-accepted `athlete_id` value (both sides stamp the identical
underlying identifier; only the JSON representation of that identifier differs, string vs. integer),
and every consuming primary-key/cursor comparison in this codebase already treats a numeric-string
and its integer counterpart as compatible strings via JSON round-tripping.

## Write actions & risks

None — Strava is read-only. `capabilities.write: false`, no `writes.json` file, matching legacy's
`ErrUnsupportedOperation` (`strava.go:97-99`): the Strava API exposes no obvious safe reverse-ETL
write surface (activity/segment/route mutation endpoints are user-authored-content edits, not a
syncable business-data write).

## Known limits

- **`athlete_stats.id` is emitted as a JSON string (`"{{ config.athlete_id }}"`), not legacy's
  int64-when-numeric representation** — see the Streams notes explanation above. ACCEPTABLE per
  conventions.md §5's meta-rule: the underlying identifier value is identical either way, and the
  schema's `["integer", "string"]` union tolerates both representations; no legacy-accepted input
  produces a different logical id.
- **`TestConformance/strava`'s dynamic (fixture-replay) checks are genuinely `skip_dynamic`'d, for
  the identical reason gmail's are** (see `internal/connectors/defs/gmail/docs.md`'s Known limits):
  the bundle's *sole* auth candidate is `mode: custom`, and conformance's synthetic config can never
  carry a real refresh token that round-trips through a live (or even a fixture-replayed) OAuth
  token exchange — the AuthHook always attempts a real HTTP POST to `token_url` to mint an access
  token, which conformance's static-fixture replay harness has no mechanism to intercept for a
  non-declarative auth path. Every auth-resolving dynamic check would therefore fail identically and
  uninformatively regardless of hook wiring. `paritytest/strava` (which wires the real `AuthHook` via
  `engine.HooksFor("strava")`, matching gmail's precedent) is the authoritative parity/correctness
  bar for this connector's auth + read path.
- **No config-driven `page_size`/`max_pages` runtime override** for `activities`/`clubs`. Legacy
  accepted both as caller-supplied config overrides (`stravaPageSize`/`stravaMaxPages`,
  `strava.go:512-540`, with `max_pages` supporting `0`/`all`/`unlimited` for uncapped). The engine's
  `page_number` pagination spec's `page_size`/`max_pages` fields are static integers on
  `streams.json`'s `base.pagination` block, not `{{ }}`-templated — there is no per-request
  config-driven override mechanism for either (identical to searxng's and shopwired's documented
  `page_size`/`max_pages` gap, conventions.md §1). Both properties are therefore NOT declared in
  `spec.json`; the bundle hard-codes legacy's own default (`page_size: 100`, uncapped `max_pages`),
  matching legacy's behavior whenever the caller does not override either config key.
