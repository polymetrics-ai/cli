# Overview

Campayn is a wave2 fan-out declarative-HTTP migration. It reads Campayn subscriber lists, email
campaigns, and calendar reports through the Campayn REST API. This bundle is migrated from
`internal/connectors/campayn` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip. **This is a partial migration**: two of legacy's five
streams (`forms`, `contacts`) are list-scoped substreams requiring a sub-resource fan-out read the
Tier-1 declarative dialect cannot express — see Known limits.

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

## Write actions & risks

None. Campayn is read-only in legacy (its write endpoints are documented as TODO upstream, per
the legacy package comment); `capabilities.write` is `false` and this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`forms` and `contacts` are NOT migrated in this bundle (blocked).** Legacy reads these two
  streams by first listing every subscriber list id (`listIDs`, `campayn.go:194-210`), then
  fetching `GET /lists/{list_id}/forms.json` or `GET /lists/{list_id}/contacts.json` once per list,
  stamping the parent `list_id` onto every emitted record (`readSubstream`, `campayn.go:161-191`).
  This is a sub-resource fan-out read (design §B.7's legitimate Tier-2 `StreamHook` trigger,
  `docs/migration/conventions.md` §1) — the Tier-1 declarative `streams.json` dialect has no
  mechanism to (a) issue a preliminary list-lists request, (b) fan out a per-stream read over each
  returned id, or (c) stamp a dynamically-discovered parent id onto every child record. Per this
  wave's hard rule (JSON + docs.md only, no Go/hooks), these two streams are left unmigrated; the
  legacy connector remains authoritative for them until a follow-up wave adds a
  `hooks/campayn/hooks.go` `StreamHook`. Blocker: `ENGINE_GAP` — see the structured result's
  `blockers[]` for this connector.
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
- Only the 3 top-level legacy-parity streams are implemented; the full known Campayn surface
  (forms, contacts, and any future endpoints) beyond these three is out of scope for this wave —
  see `api_surface.json`'s `excluded` entries.
