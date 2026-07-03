# Overview

Uptick is a fresh Tier-2 (AuthHook) migration at legacy capability parity, porting
`internal/connectors/uptick` (`uptick.go` + `streams.go`). Uptick is field service management
software (https://developer.uptick.com/); this bundle reads tasks, clients, properties, invoices,
and assets through the Uptick REST API v2.14. Uptick authenticates with a short-lived bearer access
token obtained by exchanging `client_id`/`client_secret`/`username`/`password` at
`{base_url}/api/oauth2/token/` — an OAuth 2.0 **Resource Owner Password Credentials grant**, which
the engine's built-in `oauth2_client_credentials` auth mode cannot express (that mode only performs
a client-credentials grant — no `username`/`password` fields at all — and always sends
`grant_type=client_credentials`, never `grant_type=password`). This is the same shape as
`internal/connectors/hooks/strava`/`gmail`/`snapchat-marketing`'s pilot AuthHooks;
`hooks/uptick/hooks.go` ports legacy's `passwordGrantAuth` almost verbatim. Read-only: legacy's
`Write` always returns `ErrUnsupportedOperation` (`uptick.go:118-120`) since the Uptick upstream
source supports only full_refresh/incremental reads, and this bundle declares `capabilities.write:
false` with no `writes.json` to match. The legacy package stays registered and unchanged until the
wave6 registry flip.

## Auth setup

Provide `base_url` (required; a per-tenant Uptick instance, e.g. `https://demo-fire.onuptick.com`),
`username` (plain config, not a secret), and three secrets: `client_id`, `client_secret`, and
`password` (never logged). `hooks/uptick/hooks.go` implements `AuthHook`, mirroring legacy
`uptick.go`'s `passwordGrantAuth`: it POSTs `grant_type=password` + `username` + `password` +
`client_id` + `client_secret` to `{base_url}/api/oauth2/token/`, caches the resulting access token
until 60 seconds before its declared expiry (falling back to a 1-hour TTL when the token response
carries no `expires_in`, matching legacy's `uptick.go:404-408` exactly), and sets `Authorization:
Bearer <access_token>` on every request.

`base_url` is validated for a well-formed `http(s)://` URL with a host (matching legacy's
`uptickBaseURL`, `uptick.go:451-467`, which accepts plain `http` for local test servers as well as
`https`) before any network access — this bounds SSRF risk exactly like legacy. Unlike
strava/snapchat-marketing, `base_url` has **no default** here (legacy requires it: an Uptick
instance is per-tenant, so there is no sensible shared default), matching `uptick.go:454`'s
required-and-erroring-when-absent behavior.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook": "uptick",
...}` — legacy has no alternate auth path (no static API key, no public/no-auth fallback), matching
strava/gmail/snapchat-marketing's identical single-candidate shape.

## Streams notes

Five streams, all primary-keyed on `id`, all pagination via `links.next` (an absolute next-page URL
read from the response body — `pagination.type: next_url`), matching legacy's `harvest`'s exact
convention (`uptick.go:171-228`): `tasks`, `clients`, `properties`, `invoices`, `assets`, each
reading `/api/v2.14/<resource>/` with a `fields[<Type>]` sparse-fieldset query param bounding the
response columns to legacy's exact per-stream field list (`streams.go:22-53`), `ordering=-updated`,
and `page_size` (default `100`, `spec.json`'s `default` materialization, conventions.md §3 —
matches legacy's `uptickDefaultPageSize`, `uptick.go:39`).

Every field name on every stream's raw Uptick API response already matches its schema property name
exactly (`is_active`, `contact_email`, `daily_budget_micro`-style snake_case throughout) — plain
schema-mode projection copies every field by exact key match with **zero** `computed_fields` needed
for any of the five streams, preserving each field's native JSON type (numbers stay numbers,
booleans stay booleans), matching legacy's field-built `connectors.Record{...}` mapping exactly
(`streams.go:191-282`, each `mapRecord` a pure passthrough of `item[<key>]` per field).

Every stream's incremental cursor field is `updated` (`x-cursor-field`), matching legacy's
`CursorFields: []string{"updated"}` (`streams.go`) for all five streams. Unlike snapchat-marketing/
strava, Uptick's legacy `harvest` DOES send a server-side incremental filter
(`updatedsince=<lower_bound>`, `uptick.go:167,183-185`) whenever a lower bound resolves (state
cursor, falling back to `start_date` config) — so every stream declares an `incremental` block
(`cursor_field: updated`, `request_param: updatedsince`, `start_config_key: start_date`), matching
conventions.md §8 rule 2's truth table exactly (legacy sends a server-side filter → `request_param`
declared).

## Write actions & risks

None — Uptick is read-only. `capabilities.write: false`, no `writes.json` file, matching legacy's
`ErrUnsupportedOperation` (`uptick.go:118-120`): "the Uptick upstream source supports only
full_refresh/incremental reads" (package doc comment, `uptick.go:12-13`).

## Known limits

- **`TestConformance/uptick`'s dynamic (fixture-replay) checks are genuinely `skip_dynamic`'d, for
  the identical reason strava's/gmail's/snapchat-marketing's are** (see
  `internal/connectors/defs/strava/docs.md`'s Known limits): the bundle's *sole* auth candidate is
  `mode: custom`, and conformance's synthetic config can never carry a real username/password that
  round-trips through a live (or even a fixture-replayed) OAuth token exchange — the AuthHook always
  attempts a real HTTP POST to `{base_url}/api/oauth2/token/` to mint an access token, which
  conformance's static-fixture replay harness has no mechanism to intercept for a non-declarative
  auth path. Every auth-resolving dynamic check would therefore fail identically and
  uninformatively regardless of hook wiring. `paritytest/uptick` (which wires the real `AuthHook`
  via `engine.HooksFor("uptick")`, matching strava/gmail/snapchat-marketing's precedent) is the
  authoritative parity/correctness bar for this connector's auth + read path.
- **No config-driven `max_pages` runtime override**. Legacy accepted a caller-supplied
  `max_pages` config override (`uptickMaxPages`, `uptick.go:484-497`, supporting `0`/`all`/
  `unlimited` for uncapped, defaulting to a bounded `uptickMaxPagesGuard = 10000` loop-safety cap
  when unset) in addition to the short-page/absent-`links.next` stop signal. The engine's `next_url`
  pagination spec has no `max_pages` field at all (unlike `page_number`'s static integer) — there is
  no per-request config-driven override mechanism, identical to searxng's/strava's/
  snapchat-marketing's documented `page_size`/`max_pages` gap (conventions.md §1). `max_pages` is
  therefore NOT declared in `spec.json`; the engine's own `next_url` paginator loop-guards against
  infinite pagination independently (a repeated-URL loop guard, `paginate.go`'s `nextURL.Next`),
  which is a strictly stronger termination guarantee than legacy's page-count cap alone for any
  legacy-accepted input — never data-changing, since both sides still terminate on the identical
  absent-`links.next` signal for any real (non-adversarial, non-looping) API response sequence.
