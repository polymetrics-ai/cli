# Overview

Smartwaiver is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full
documented v4 API surface. It reads Smartwaiver waivers, checkins, templates, published keys,
user info, and account settings, and sends webhook-configuration, webhook-resend, SMS, and
waiver-prefill mutations through the Smartwaiver v4 API (`https://api.smartwaiver.com/v4/...`,
documented at `https://api.smartwaiver.com/docs/v4/`). This bundle migrates
`internal/connectors/smartwaiver` (the hand-written connector); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Smartwaiver API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`), matching legacy's `connsdk.Bearer(key)`
(`smartwaiver.go:150`). `base_url` defaults to `https://api.smartwaiver.com` and may be overridden
for tests/proxies, matching legacy's own `validatedBaseURL` default.

## Streams notes

Six streams hit their own `GET` endpoint: `waivers` (`/v4/waivers`, records at
`waivers.waivers`), `checkins` (`/v4/checkins`, records at `checkins.checkins`), `templates`
(`/v4/templates`, records at `templates.templates`), `published_keys` (`/v4/keys/published`,
records at `published_keys.keys`), `user_info` (`/v4/info`, a single JSON object with no
records-array wrapper — `records.path: "."` returns the whole body as one record), and the
Pass-B-added `settings` (`/v4/settings`, also a single whole-object record via `records.path:
"."`, exposing the account's console `staticExpiration`/`rollingExpiration` waiver-expiry
policy) — matching legacy's `streamEndpoints` map's nested/flat `recordsPath` shapes exactly for
the original five, plus the same flat-object shape for the new one. Every stream declares
`projection: "passthrough"`: legacy's `readRecords` emits each decoded record verbatim
(`emit(connectors.Record(rec))`, `smartwaiver.go:122`, no field-building or `mapRecord` step), so
this bundle emits every raw field the API returns rather than narrowing to whatever subset
`schemas/*.json` declares — schema-mode projection on a verbatim-emitting legacy would silently
drop real API fields (`conventions.md` §8 rule 1).

None of the streams paginate in legacy (a single `r.Do` call per read, no loop) —
`pagination.type: none` is declared, one request per read — despite every request sending
`limit`/`offset=0` query params, matching legacy's `queryParams` (`smartwaiver.go:153-163`).
`limit` defaults to 100 (legacy's `defaultPageSize`) and is configurable via `page_size`.
`fromDts`/`toDts` are optional passthrough date filters, omitted entirely when unset, matching
legacy's `copyConfig` (only set when non-empty).

`Check` hits `/v4/me` (an account-identity probe distinct from the `user_info` stream's `/v4/info`
endpoint), matching legacy's `Check` (`smartwaiver.go:47`) exactly.

## Write actions & risks

Legacy's own connector is read-only, but Pass B full-surface expansion adds every
dialect-expressible mutation the real Smartwaiver v4 API documents (`api_surface.json`), each
requiring approval:

- **`set_webhook_config`** (`PUT /v4/webhooks/configure`, `update`): sets the account's webhook
  delivery endpoint and email-validation-required policy (optionally targeting a specific
  `webhookNumber` or creating a new one with `create: true`, up to 3 webhooks per account). Risk:
  changes where the account's near-real-time waiver-signed notifications are delivered — an
  operator-facing integration change, not a data mutation, but still capable of silently
  redirecting/losing webhook traffic if misconfigured.
- **`resend_webhook`** (`PUT /v4/webhooks/resend/{{ record.waiver_id }}`, `custom`, no body):
  re-triggers the new-waiver webhook for a given waiver ID. Smartwaiver documents this as a
  testing aid only and heavily rate-limits it (2 requests/minute, independent of the account's
  normal 100 rpm budget).
- **`send_sms`** (`POST /v4/sms`, `create`): sends a real outbound SMS containing a
  waiver-signing link to `number` for `templateId`. Smartwaiver rate-limits this per day for
  anti-spam/abuse prevention. Risk: an external, billable, real-world side effect (an actual text
  message to a phone number) — never dry-run this against a real number without approval.
  `number` is typed `string` in `record_schema` (not `integer`) since a phone number is not
  numeric data (leading `+`/formatting would be lost); Smartwaiver's own docs example happens to
  show a bare `0` placeholder, not a real shape constraint.
- **`prefill_template`** (`POST /v4/templates/{{ record.template_id }}/prefill`, `create`):
  generates a prefilled waiver-signing link (`participants[]`, `guardian`, address fields,
  `customWaiverFields`) — every top-level field in Smartwaiver's documented request body maps
  directly to a body field via the default JSON body-construction rule (§3), including the
  nested `participants` array and `guardian`/`customFields`/`customWaiverFields` objects, since
  `body_type: json` passes a record's non-path field values through as-is (arrays/objects
  survive; the dialect's body builder is not restricted to scalar fields). Risk: the request body
  and the returned prefilled-waiver URL both carry real participant PII (name, DOB, address,
  phone, custom-field answers) — treat both the write payload and its response as sensitive.

## Known limits

- **`page_size`'s upper bound (100) is not enforced by this bundle.** Legacy validates
  `page_size` is an integer between 1 and 100 in Go code (`smartwaiver.go:169-179`) before sending
  it as `limit`. The engine's declarative query dialect has no numeric-range validation
  primitive; an out-of-range `page_size` is sent to the API as-is rather than rejected
  client-side. This is a scope narrowing (client-side validation removed, not a data-shape
  change): the API itself is the ultimate arbiter of an invalid page size on both sides.
- **`start_date_2` (a secondary `fromDts` override config key) is not modeled.** Legacy accepts
  BOTH `start_date` and `start_date_2` as `fromDts` sources, applying `copyConfig` for
  `start_date` first and then `start_date_2` second (`smartwaiver.go:159-160`) — since both calls
  target the same `url.Values` key via `Set` (not `Add`), `start_date_2`, when present,
  unconditionally overwrites whatever `start_date` set, a two-config-key coalesce-with-precedence
  rule. The engine's `stream.Query` dialect has no mechanism to express "one of two optional
  config keys, second takes priority when both are set" as a single query entry — only one
  `template`/`default`/`omit_when_absent` triple can target one param name. This bundle therefore
  models only the primary `start_date` key (the documented, non-suffixed name); `start_date_2` is
  not declared in `spec.json` at all (a declared-but-unwireable key is worse than an absent one).
  Out-of-scope, not silently wrong: any sync relying solely on `start_date` is unaffected.
- **Response field sets in `schemas/*.json` are documentation, not a projection filter.**
  Smartwaiver's public API reference was consulted for real wire-shape field names (`waiverId`,
  `templateId`, `createdOn`, `checkinId`, `key`, `label`, `username`, etc.), but legacy's own
  `streams()` catalog only ever declared a generic 3-field shape (`waiverId`/`templateId`/
  `createdAt`) shared identically across all 5 streams regardless of each stream's real shape (a
  legacy catalog-authoring shortcut, not a real per-stream contract). This bundle's schemas
  instead document each stream's own real, distinct identity/timestamp fields, verified against
  Smartwaiver's public API documentation — but data completeness does not depend on how accurate
  or complete that documentation is: `projection: "passthrough"` (see Streams notes) emits every
  raw field the API actually returns regardless of what `schemas/*.json` declares, matching
  legacy's own verbatim-emit behavior exactly.
