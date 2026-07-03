# Overview

Twilio is a Tier-2 (StreamHook) migration of `internal/connectors/twilio` (legacy
`twilio.Connector`). It reads Twilio messages, calls, recordings, conferences, and account usage
records through the Twilio REST API (`2010-04-01`), read-only. This bundle is engine-vs-legacy
parity-tested against `internal/connectors/twilio` (the hand-written connector it migrates); the
legacy package stays registered and unchanged until wave6's registry flip. Read-only: legacy
`twilio.go:99-101` always returns `ErrUnsupportedOperation` from `Write` (sending messages/placing
calls are side-effecting actions inappropriate for a generic reverse-ETL source), and this bundle
declares `capabilities.write: false` with no `writes.json` to match.

Twilio was originally quarantined under an `OTHER`/no-reason marker (`docs/migration/
quarantine.json`); investigation for this pass found the connector's actual blocker is **not**
auth (Twilio's HTTP Basic account_sid/auth_token pair is fully declarative-expressible, `mode:
basic`) but **pagination**: see "Streams notes" below.

## Auth setup

Provide two secrets: `account_sid` (used as both the HTTP Basic username AND to scope every
account-relative resource path, e.g. `/Accounts/{account_sid}/Messages.json`) and `auth_token`
(used only as the HTTP Basic password). Both are declared `x-secret: true` and never logged. This
is a **fully declarative** `mode: basic` auth candidate — no `AuthHook` is needed for this
connector, unlike gmail/youtube-analytics's OAuth refresh-grant shape:

```json
{ "mode": "basic", "username": "{{ secrets.account_sid }}", "password": "{{ secrets.auth_token }}" }
```

`base.check` probes `GET /Accounts/{{ secrets.account_sid }}/Messages.json` (matching legacy's
`Check`, `twilio.go:88-92`, which reads a bounded 1-record page of Messages to confirm auth and
connectivity without mutating anything).

## Streams notes

Five streams, every one primary-keyed on `sid` (`usage_records` on `category`, matching legacy's
`PrimaryKey: []string{"category"}` — Twilio's usage-records resource has no per-record `sid`):
`messages`, `calls`, `recordings`, `conferences`, `usage_records`. Every stream reads an
account-scoped resource path (`/Accounts/{{ secrets.account_sid }}/<Resource>.json`) and returns a
JSON object holding the records array under a resource-named key (`records.path`), matching
Twilio's real list-endpoint envelope exactly.

**Pagination — Tier-2 StreamHook, not declarative (genuine `ENGINE_GAP`, same shape as
`docs/migration/quarantine.json`'s `rootly`/`safetyculture` entries)**: Twilio's list pagination
follows a `next_page_uri` field read from the response body
(`twilio.go:161,193-206`/`TestReadPaginatesAndAuthenticates`) — and the real wire value is a
**host-relative URL** (e.g. `"/2010-04-01/Accounts/AC_test/Messages.json?Page=1&PageSize=2"`), not
an absolute one. The engine's only declarative pagination type that reads a next-page URL from the
response body, `next_url` (`engine/paginate.go`'s `nextURL`/`checkOrigin`), enforces a same-origin
SSRF guard that **fail-closed rejects any next-page URL with an empty `Host`** — `checkOrigin`
returns `"next URL ... has no host; rejecting (fail closed)"` for exactly this shape, which is
correct guard behavior for a genuinely cross-host redirect attempt but incorrectly also rejects
Twilio's own legitimate host-relative convention. There is no dialect escape hatch (no
"host-relative next_url" variant, no `allow_relative` flag) — declaring `next_url` for any Twilio
stream would silently stop pagination after page 1 in production (the guard's rejection surfaces
as a sticky `Err()`, which `readOneSequence` treats as pagination-terminated-with-error), which is
strictly worse than not declaring pagination at all. This is a narrower, StreamHook-specific
repeat of the identical structural gap `rootly`/`safetyculture` hit for their own relative
`links.next`/`next_page` body-cursor fields.

`hooks/twilio/hooks.go` implements `StreamHook.ReadStream`, porting legacy's `harvest`
(`twilio.go:161-208`) verbatim: issue the first request with a `PageSize` query param (from
`config.page_size`, default 50), extract records from the resource-named key, then follow
`next_page_uri` via `absoluteURL` (host-relative URLs are resolved against the SAME requester's
resolved base origin — never a caller-controlled host, so the SSRF surface `next_url`'s guard
protects against does not reopen here: the hook only ever follows a path Twilio itself returned,
scoped to the connection's own configured host) until `next_page_uri` is null/empty, or
`config.max_pages` is reached.

Every stream carries a `"conformance": {"skip_dynamic": true, "reason": "..."}` marker
(`docs/migration/conventions.md` §4, sentry/monday precedent): `streams.json`'s declared
`base.pagination: {"type": "none"}` and each stream's plain `path`/`records.path` are **not** the
real production dispatch path — every real `Read()` call routes through the StreamHook
(`ReadStream` always returns `handled=true`), so a declarative fixture replay could never exercise
the real host-relative `next_page_uri`-follow behavior at all. The authoritative substitute named
in each marker is `paritytest/twilio` (`TestParityTwilio_MessagesTwoPagePagination`, live
`httptest.Server`-driven) and `hooks/twilio/hooks_test.go`.

No incremental sync mode is sent as a request parameter anywhere in legacy: `CursorFields` is
published on the catalog/manifest surface only (`date_sent`/`start_time`/`date_created`/
`start_date` per stream), but `harvest` never forwards a state cursor into a request — there is no
`incrementalLowerBound`-equivalent call anywhere in the legacy package (`InitialState` always seeds
an empty cursor, `twilio.go:110-117`). This bundle matches that exactly by declaring **no
`incremental` block on any stream** (full_refresh only) while still declaring `x-cursor-field` on
each schema (matching sentry's identical precedent: catalog-surface-only cursor fields, never
wired to a request).

## Write actions & risks

None — Twilio is read-only. `capabilities.write: false`, no `writes.json` file, matching legacy's
`ErrUnsupportedOperation` (`twilio.go:99-101`).

## Known limits

- **Per-stream `skip_dynamic` markers, not a bundle-level marker** (unlike gmail/youtube-analytics):
  Twilio's `check` and `auth` ARE declarative and dynamically exercisable (`mode: basic` resolves
  cleanly against conformance's synthetic secrets) — only the STREAM read path is StreamHook-only,
  so `check_fixture` and any auth-only dynamic check still run normally; only
  `read_fixture_nonempty:<stream>`/`pagination_terminates`/`records_match_schema`/
  `cursor_advances` are skipped for these hook-dispatched streams (conventions.md §4's per-stream
  marker semantics), matching monday/sentry's shape exactly.
- **Host-relative `next_url` pagination remains a genuine, unresolved `ENGINE_GAP`** for any future
  Tier-1 attempt: see "Streams notes" above. A dialect addition (e.g. an `allow_relative_next_url`
  flag on `PaginationSpec` that resolves a relative next-page URL against the requester's own
  resolved base origin, mirroring exactly what this hook already does safely) would let this
  connector — and `rootly`/`safetyculture` — drop their StreamHooks entirely. Not pursued as an
  engine increment in this pass; this is the 3rd connector to hit the identical shape
  (twilio/rootly/safetyculture), which meets `conventions.md` §6's recurrence bar for a future
  engine-mini-wave candidate, but that increment is out of scope for this Tier-2 hook-authoring
  pass.
- Full Twilio API surface (TaskRouter, Verify, Sync, Video, sending messages/placing calls,
  account/phone-number management) is out of scope for this pass; see `api_surface.json`'s
  `excluded` entries. Only the 5 legacy-parity read streams are implemented.
