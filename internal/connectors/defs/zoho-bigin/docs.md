# Overview

Zoho Bigin is a Tier-2 (AuthHook) migration repairing the `AUTH_COMPLEX` quarantine entry recorded
in `docs/migration/quarantine.json` ("Legacy performs an OAuth2 refresh_token grant token exchange
... before every read/check. The engine's declarative auth dialect only supports the oauth2 [client
credentials grant]"). It reads Zoho Bigin pipelines, module records, and field metadata via the
Zoho OAuth 2.0 **refresh-token grant** only — the 3-legged consent/acquisition dance is out of scope
(the refresh token arrives as a pre-issued secret; the credentials layer already owns acquisition/
storage), matching gmail's precedent (`internal/connectors/hooks/gmail/hooks.go`). This bundle
migrates `internal/connectors/zoho-bigin` (the hand-written connector it replaces); the legacy
package stays registered and unchanged until wave6's registry flip. Read-only: legacy
`zoho_bigin.go`'s `Write` always returns `ErrUnsupportedOperation`, and this bundle declares
`capabilities.write: false` with no `writes.json` to match.

## Auth setup

Provide three secrets: `client_id`, `client_secret`, and `client_refresh_token` (long-lived; never
logged) — all three are `required` in `spec.json`, matching legacy's `requireOAuth` check (unlike
gmail, Zoho Bigin's legacy connector treats `client_secret` as mandatory, not optional).
`hooks/zoho-bigin/hooks.go` implements `AuthHook`, copying gmail's hook pattern
(`docs/migration/conventions.md` §1's Tier-2 table: token-exchange auth) adapted for zoho-bigin's
own required-field shape: it POSTs `grant_type=refresh_token` + `client_id` + `client_secret` +
`refresh_token` to `token_url` (default `https://accounts.zoho.com/oauth/v2/token`,
config-overridable), caches the resulting access token until 60 seconds before its declared expiry,
and sets `Authorization: Zoho-oauthtoken <access_token>` on every request (Zoho's own header scheme
— legacy's `refreshToken` decodes `access_token` from the JSON response and the read path applies it
via `connsdk.Bearer`, which legacy itself sends as a plain `Bearer <token>` header; this bundle
instead uses Zoho's documented `Zoho-oauthtoken` scheme directly in the hook, since a custom
AuthHook is not constrained to `connsdk.Bearer`'s prefix — this is a stricter-correctness match to
Zoho's own published API contract, not a deviation from any legacy-observable behavior since legacy
only replayed the raw access token string it received).

`token_url` MUST resolve to an `https://` URL: the hook fails closed on a non-https or unparseable
override rather than sending the refresh token/client secret to an attacker-chosen endpoint. This
mirrors legacy's `validateURL` (`zoho_bigin.go:233-241`) but tightens it to https-only in the hook
(legacy's `validateURL` also accepted plain `http`) — see Known limits.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook": "zoho-bigin",
...}` — legacy has no alternate auth path (no static API key, no public/no-auth fallback), so there
is no `when`-gated bypass to declare.

## Streams notes

Three streams, all primary-keyed on `id`, none paginated (`base.pagination: {"type": "none"}`,
matching legacy: `zoho_bigin.go`'s `Read` issues exactly one request per stream, with no
page/offset/cursor query parameter anywhere):

- `pipelines` — `GET /Pipelines`, records at `data`. Schema projection (default mode) matches
  legacy's `mapPipeline` exactly: `id`, `name`, `display_value`, no rename needed (raw field names
  already match the schema).
- `records` — `GET /{{ config.module_name }}` (defaults to `Deals`, matching legacy's
  `zoho_bigin.go:103-107` fallback when `module_name` is unset), records at `data`. Declared
  `projection: "passthrough"` rather than schema projection — see Known limits for why.
- `fields` — `GET /settings/fields`, records at `fields`. Declared `projection: "passthrough"` for
  the same reason as `records` — see Known limits.

## Write actions & risks

None — Zoho Bigin is read-only. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`zoho_bigin.go:125-127`).

## Known limits

- **`records` and `fields` streams do not reproduce legacy's multi-field name/id coalesce.**
  Legacy's `mapRecord` derives its `name` output field as `first(item["name"], item["Deal_Name"],
  item["display_value"])` (first non-empty value wins across three differently-shaped raw fields,
  since different Zoho Bigin modules use different display-name conventions), and `mapField`
  derives `id` as `first(item["id"], item["api_name"])`. The engine's `computed_fields` dialect has
  no coalesce/fallback-across-multiple-source-fields primitive (`docs/migration/conventions.md` §3:
  every `computed_fields` entry is a single template resolved against one reference or literal, with
  only "skip if THIS entry's source is absent" tolerance, never "try field A, else field B, else field
  C"). Declaring only the first-priority field (e.g. `name` alone) would silently drop records where
  legacy would have fallen back to `Deal_Name`/`display_value`/`api_name` — an accepted-input
  emitted-DATA change, not cosmetic. Both streams are instead declared `projection: "passthrough"`:
  every raw field (`id`, `name`, `Deal_Name`, `display_value`, `api_name`, `display_label`, and any
  other module-specific field) survives verbatim, strictly more permissive than legacy (a downstream
  consumer can reproduce legacy's exact coalesce priority itself, or read the specific field it
  needs) and never drops data legacy would have emitted for any accepted input. This is documented
  here per `docs/migration/conventions.md` §5's parity-deviation ledger convention; classified
  ACCEPTABLE (never drops/changes data for any legacy-accepted input, differs only in also exposing
  additional raw fields legacy's narrower projection discarded).
- **`token_url` https-only enforcement is stricter than legacy's `validateURL`** (which accepted
  plain `http` too, `zoho_bigin.go:233-241`): the hook only accepts `https://` overrides. Never
  stricter for any *production* Zoho OAuth endpoint, which is always https; strictly safer for the
  one new SSRF-adjacent secret-bearing surface this migration adds. See the parity-deviation ledger
  in `docs/migration/conventions.md` §5.
- **`data_center` is not modeled as a config key.** Legacy's own test fixtures set a `data_center`
  config value, but `zoho_bigin.go` never reads it anywhere (dead config in legacy itself, not just
  in this migration) — `base_url` is the sole, already-correct override mechanism for a
  region-specific data center (e.g. `https://www.zohoapis.eu/bigin/v2`). Not declared in `spec.json`
  per `docs/migration/conventions.md` F6 (a spec property with no wired template is dead config).
- **`TestConformance/zoho-bigin`'s dynamic (fixture-replay) checks are `skip_dynamic`'d** for the
  identical reason as gmail's bundle-level marker: this bundle's *sole* auth candidate is `mode:
  custom`, and conformance's synthetic config can never carry a real `https` `token_url` — the
  AuthHook's own https-only guard means no synthetic secret value can ever satisfy it, so every
  auth-resolving dynamic check would fail identically and uninformatively regardless of hook wiring.
  `hooks/zoho-bigin/hooks_test.go` is the authoritative substitute proof for the AuthHook's real
  OAuth2 refresh-grant behavior (form shape, caching/expiry, https enforcement, error paths, secret
  redaction) — the same gmail precedent this bundle's `metadata.json` `conformance.reason` names.
