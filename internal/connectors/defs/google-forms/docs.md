# Overview

Google Forms reads form metadata, form items (questions), and submitted responses for one or more
configured forms through the Google Forms API v1. This is a full legacy-parity migration of the
hand-written connector (`internal/connectors/google-forms`), which stays registered and unchanged
until wave6's registry flip. Read-only: legacy sets `Capabilities.Write = false` and `Write` always
returns `ErrUnsupportedOperation`; this bundle matches with `capabilities.write: false` and no
`writes.json`.

This is a **Tier-2 bundle** (AuthHook + CheckHook, `internal/connectors/hooks/google-forms/
hooks.go`) — the Forms API authenticates with a short-lived OAuth2 access token, and this connector
is configured with a long-lived refresh token (plus client id/secret) that must be exchanged for an
access token at Google's token endpoint, a signature/token-exchange auth scheme the declarative
`auth` dialect cannot express (`docs/migration/conventions.md` §1's sanctioned Tier-2 trigger list).
The 3 read streams themselves (`forms`, `form_items`, `responses`) are fully declarative once auth
is resolved — no `StreamHook`/`RecordHook`/`WriteHook` is needed, matching gmail's identical
AuthHook-only shape (`hooks/gmail/hooks.go`, `internal/connectors/hooks/google-forms/hooks.go`
mirrors it field-for-field for the refresh-token-grant machinery).

## Auth setup

Provide three secrets: `client_id`, `client_secret` (optional for some OAuth client types — see
below), and `client_refresh_token` (long-lived; never logged). `hooks/google-forms/hooks.go`
implements `AuthHook`, mirroring legacy `google-forms/googleforms.go`'s `oauthRefreshAuth`: it POSTs
`grant_type=refresh_token` + `refresh_token` + `client_id` [+ `client_secret`] to `token_url`
(default `https://oauth2.googleapis.com/token`, config-overridable), caches the resulting access
token until 60 seconds before its declared expiry, and sets `Authorization: Bearer <access_token>`
on every request. `client_secret` is omitted from the token-request form when unset (matches
legacy's `if a.clientSecret != ""` guard) — some Google OAuth client types (e.g. installed-app/
native clients) issue refresh tokens that don't require a client secret at token-refresh time.

`token_url` MUST resolve to an `https://` URL: the hook fails closed on a non-https or unparseable
override rather than sending the refresh token/client secret to an attacker-chosen endpoint — this
is stricter than legacy's `resolveHTTPURL` (which also accepted plain `http` for both `base_url` and
`token_url`); the tightened rule applies only to `token_url` (never `base_url`), mirroring gmail's
identical, already-ledgered deviation. This is the one new SSRF-adjacent surface this bundle adds.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook":
"google-forms", ...}` — legacy has no alternate auth path (no static API key, no public/no-auth
fallback), so there is no `when`-gated bypass to declare, matching gmail's identical
single-candidate shape.

`hooks/google-forms/hooks.go` also implements `CheckHook`: legacy's `Check`
(`googleforms.go:74-100`) reads the FIRST configured `form_id` (`formIDs[0]`) as a bounded metadata
read confirming auth/connectivity — the declarative check dialect has no "take the first element of
a comma-separated config value" primitive (unlike a fan_out stream, `base.check` has no `ids_from`
concept), so this one extra hook interface expresses it directly rather than approximating it with
a `"{{ config.form_id }}"` path template that would break for any multi-form_id config.

## Streams notes

All 3 streams fan out over the comma/space/newline-separated `form_id` config value via the
engine's `stream.fan_out` dialect (`ids_from.config_key: form_id`, `into.path_var: form_id`,
referenced in each stream's own `path` as `{{ fanout.id }}`), reproducing legacy's own per-form-id
loop (`readForms`/`readResponses`, `googleforms.go:170-249`) exactly.

`forms` (`GET /forms/{{ fanout.id }}`, `records.path: "."` — the whole form object is one record)
lists metadata for every configured form; primary key `form_id`. `computed_fields` rename the raw
API's camelCase `formId`/`revisionId`/`responderUri` to the schema's snake_case names and reach into
the nested `info` object for `title`/`documentTitle`/`description` — plain schema projection copies
by exact top-level key match only, so both the rename and the nested-object reach require
`computed_fields` (mirrors gmail's `drafts` stream reaching into its own nested `message` object).
`item_count` is derived with the `length` filter over the raw `items` array, matching legacy's
`mapFormRecord`.

`form_items` reads the IDENTICAL `GET /forms/{{ fanout.id }}` endpoint as `forms` (Google Forms has
no dedicated "list items" endpoint; item data is embedded in the form resource) but selects the
nested `items` array as its records path (`records.path: "items"`) instead of the whole object,
exploding it into one record per item — exactly legacy's `mapFormItemRecords`
(`streams.go:113-140`). `fan_out.stamp_field: "form_id"` writes the current form id onto every
emitted item record (items themselves carry no `formId` field in the raw API). `question_id` is a
`computed_fields` reach into the doubly-nested `questionItem.question.questionId` path, matching
legacy's identical type-asserted nested read; an item with no `questionItem` (e.g. a section break
or image item) silently omits `question_id` for that record (computed_fields' documented
absent-source-path tolerance), matching legacy's own `nil`-default behavior.

`responses` (`GET /forms/{{ fanout.id }}/responses`, `pagination.type: cursor` with
`token_path: nextPageToken`/`cursor_param: pageToken`, matching legacy's `readResponses` pagination
loop exactly) sends `pageSize` from `config.page_size` (default 5000, matching legacy's
`googleFormsDefaultPageSize`) and an optional `filter` query built from the engine's
`{{ incremental.lower_bound }}` reference (S3 engine mini-wave item 1): `"filter": {"template":
"timestamp >= {{ incremental.lower_bound }}", "omit_when_absent": true}` sends `filter=timestamp >=
<value>` exactly when the incremental lower bound resolves (a repeat sync's state cursor, or
`start_date` config on a fresh sync) and omits the `filter` param entirely on a from-scratch sync
with no `start_date` configured — matching legacy's `responseFilter` exactly (`googleforms.go:
452-464`: `bound == "" -> ""`, else `"timestamp >= " + bound`). `last_submitted_time` is the
declared `x-cursor-field`/`incremental.cursor_field`, with `start_config_key: start_date` — matching
legacy's own catalog (`CursorFields: []string{"last_submitted_time"}`).

`answers` (a nested object) and `total_score` (a number) are single **bare** `computed_fields`
references (`{{ record.answers }}`, `{{ record.totalScore }}`) — the gap-loop cycle-1 engine
mini-wave's typed `computed_fields` extraction (conventions.md §3) copies the raw JSON value
straight through for exactly this shape, preserving `answers`' native nested-object structure and
`total_score`'s native number type, matching legacy's verbatim `item["totalScore"]`/`item["answers"]`
assignment exactly (no stringification anywhere in either connector).

## Write actions & risks

None. `capabilities.write` is `false`; no `writes.json` is shipped. Legacy itself implements no
write path for Google Forms (`Write` always returns `ErrUnsupportedOperation`).

## Known limits

- **`token_url` https-only enforcement is stricter than legacy's `resolveHTTPURL`** (which accepted
  plain `http` for both `base_url` and `token_url`): the hook only accepts `https://` overrides for
  `token_url` specifically. This is documented as a parity deviation (never stricter for any
  *production* Google OAuth endpoint, which is always https; strictly safer for the one new
  SSRF-adjacent secret-bearing surface this bundle introduces) — identical in shape to gmail's
  already-ledgered deviation (`docs/migration/conventions.md` §5).
- **`TestConformance/google-forms`'s dynamic (fixture-replay) checks are `skip_dynamic`'d for the
  identical reason as gmail**: the sole auth candidate is `mode: custom` with no `when`-gated
  fallback, and conformance's synthetic config can never carry a real `https` `token_url` that the
  AuthHook's own https-only guard would accept — every auth-resolving dynamic check would otherwise
  fail identically and uninformatively. `hooks/google-forms/hooks_test.go` (which drives the real
  `AuthHook`/`CheckHook` directly via `httptest` servers) is the authoritative correctness bar for
  this connector's auth and check paths, matching gmail's precedent exactly.
- **`page_size`'s legacy-enforced bounds (1-5000, `googleFormsMaxPageSize`) are not statically
  validated by the engine dialect** — `spec.json` declares it a plain `integer` with a default; an
  out-of-range value is passed straight to the live API rather than rejected client-side the way
  legacy's `googleFormsPageSize` helper does. Not a data-parity issue (the API itself still rejects
  an invalid value), just a shifted validation boundary (same shape as google-tasks'/
  google-search-console's identical, already-ledgered deviations).
