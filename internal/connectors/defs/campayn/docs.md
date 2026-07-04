# Overview

Campayn reads and writes Campayn subscriber lists, signup forms, contacts, email campaigns, and
calendar reports through the Campayn REST API. This bundle is migrated from
`internal/connectors/campayn` (the hand-written connector); the legacy package stays registered and
unchanged until wave6's registry flip. All 5 legacy streams are migrated: the 3 top-level
collections (`lists`, `emails`, `reports`) plus the 2 list-scoped substreams (`forms`, `contacts`),
which use the engine's `fan_out` dialect (S4 engine mini-wave item 2) — the `ENGINE_GAP` that
previously blocked these two streams is closed.

**Pass B full-surface expansion** (`api_surface.json`, re-reviewed 2026-07-04 against the live
GitHub-hosted docs): Campayn's entire documented API surface is 6 endpoint-doc pages
(lists/contacts/forms/emails/reports/signup); every documented GET is now a stream (the two
single-item detail GETs — `/lists/{id}/forms/{form_id}.json`, `/contacts/{contact_id}.json` — are
`duplicate_of`/shape-mismatch exclusions, not gaps, see `api_surface.json`), and every documented,
dialect-expressible mutation is now a `writes.json` action: `add_contact` (`POST
/lists/{list_id}/contacts.json`), `update_contact` (`PUT /contacts/{contact_id}.json`), and
`unsubscribe_contact` (`POST /lists/{list_id}/unsubscribe.json`). The upstream README's own "Writing
through the API" section is literally marked `TODO` by Campayn itself; these 3 actions are
everything the endpoint docs actually specify a request/response shape for. `POST /signup` (creates
a new Campayn account, not a data record, and is explicitly gated behind separate privileged
authorization per the docs) is excluded as `requires_elevated_scope`.

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

`reports` additionally supports the documented `from`/`to` calendar-window query filters
(`stream.Query`'s optional-query dialect, `omit_when_absent: true` on both `config.report_from`/
`config.report_to`) — both params are UTC microtime (Unix seconds) values per the docs' own PHP
`strtotime` example; neither is sent when unset, matching the docs' "none, one, or both" filter
contract exactly.

## Write actions & risks

`capabilities.write` is now `true`; legacy was read-only only because its write endpoints were
undocumented at migration time (the legacy package comment called them "TODO upstream") — Pass B's
docs re-review confirms the endpoint docs (`endpoints/{contacts,lists}.md`) DO fully specify 3
mutations:

- `add_contact` (`POST /lists/{list_id}/contacts.json`) — adds a new contact to a list; low risk,
  no approval required.
- `update_contact` (`PUT /contacts/{contact_id}.json`) — replaces a contact's full field set; the
  docs explicitly warn any field omitted from the body is **removed** from the contact (not merged
  as a partial patch) — callers must resend the full desired field set every time. No approval
  required, but this is a wider-blast-radius PUT than a typical partial update.
- `unsubscribe_contact` (`POST /lists/{list_id}/unsubscribe.json`) — unsubscribes either one
  contact (by `id`) or every contact on the list sharing a given `email`; the docs note neither
  path is reflected in Campayn's own Reporting UI. No approval required.

No delete endpoint exists anywhere in the documented surface for any Campayn resource.

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
- All 5 known Campayn streams (lists, emails, reports, forms, contacts) are implemented, and every
  documented mutation is now covered by a write action — see `api_surface.json` for the full,
  closed-vocabulary accounting of every documented endpoint.
- `contacts`' schema still reflects only the fields the LIST endpoint documents (email, first_name,
  last_name, confirmed, image_url) — the single-contact detail endpoint's richer field set (title,
  company, address, phones[], sites[], social[], custom_fields[], etc.) is intentionally not
  modeled on this stream; see `api_surface.json`'s `duplicate_of` entry for
  `/contacts/{contact_id}.json`.
