# Overview

Campayn is a wave2 fan-out declarative-HTTP migration. It reads Campayn subscriber lists, signup
forms, contacts, email campaigns, and calendar reports through the Campayn REST API. This bundle
is migrated from `internal/connectors/campayn` (the hand-written connector); the legacy package
stays registered and unchanged until wave6's registry flip. All 5 legacy streams are now migrated:
the 3 top-level collections (`lists`, `emails`, `reports`) plus the 2 list-scoped substreams
(`forms`, `contacts`), which use the engine's `fan_out` dialect (S4 engine mini-wave item 2) — the
`ENGINE_GAP` that previously blocked these two streams is closed.

## Auth setup

Provide a Campayn API key via the `api_key` secret; it is sent as
`Authorization: TRUEREST apikey=<api_key>` (`streams.json` `base.auth`'s `api_key_header` mode
with `prefix: "TRUEREST apikey="`), matching legacy's
`connsdk.APIKeyHeader("Authorization", secret, campaignAuthPrefix)` (`campayn.go:273`). Never
logged.

`base_url` is **required** with no default — see Known limits for why the sub_domain-derived
default legacy computes cannot be expressed here.

## Streams notes

`lists` (`GET /lists.json`), `emails` (`GET /emails.json`), and `reports`
(`GET /reports/calendar.json`) are top-level collections read directly; Campayn returns each as a
bare top-level JSON array (`records.path: ""`), matching legacy's
`connsdk.RecordsAt(resp.Body, "")`. None are paginated or incremental — legacy's
`campaignStreamEndpoints` declares no pagination for any Campayn endpoint and no
`CursorFields` for any stream (Campayn's read API supports full refresh only), so no
`pagination`/`incremental` block is declared here either.

`forms` and `contacts` are list-scoped substreams: legacy's `readSubstream` (`campayn.go:161-191`)
first lists every subscriber list id (`listIDs`, `campayn.go:194-210`), then fetches
`GET /lists/{list_id}/forms.json` or `GET /lists/{list_id}/contacts.json` once per list, stamping
the parent `list_id` onto every emitted record. This bundle expresses the identical sequence via
`streams.json`'s `fan_out` block: `ids_from.request` issues a preliminary `GET /lists.json` (a bare
top-level array with no pagination, matching legacy's own single unpaginated list call — the
id-listing request declares no pagination block of its own, conventions.md §3), extracts `id` off
every returned list record, then `into.path_var: "list_id"` threads each resolved id into
`/lists/{{ fanout.id }}/<resource>.json`'s path, and `stamp_field: "list_id"` writes it onto every
emitted record after projection — matching legacy's stamped `list_id` field exactly. Like the
top-level streams, neither substream is paginated or incremental at the per-list level, matching
legacy's own single unpaginated per-list fetch.

## Write actions & risks

None. Campayn is read-only in legacy (its write endpoints are documented as TODO upstream, per
the legacy package comment); `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`base_url` has no default (a scope narrowing, not a data-parity change).** Legacy derives the
  base URL from a `sub_domain`/`domain` config value
  (`https://<sub_domain>.campayn.com/api/v1`) when `base_url` itself is unset
  (`campaignBaseURL`, `campayn.go:290-312`), with a validated single-DNS-label check to prevent
  host injection. The engine's `spec.json` `"default"` materialization
  (`docs/migration/conventions.md` §3) only fills in a FIXED literal for a genuinely-absent key —
  it has no mechanism to derive one config value's default from another config value at
  bundle-load or read time. Requiring `base_url` directly (documented here, not silently narrowed)
  is the honest representation; a future capability-expansion pass could revisit this if the
  dialect grows a base-URL-construction template mechanism.
- **`forms`/`contacts` fan-out has no list-count cap.** Legacy's `listIDs` has no defensive cap of
  its own either (unlike bigmailer's `bigmailerMaxBrands`), so the engine's `fan_out.ids_from.request`
  (which fully "paginates" the unpaginated `/lists.json` call to exhaustion — trivially, since it
  returns a bare array in one response) is exact parity here, not a deviation.
- All 5 known Campayn streams (lists, emails, reports, forms, contacts) are implemented; any future
  Campayn endpoints beyond these five are out of scope for this wave — see `api_surface.json`.
