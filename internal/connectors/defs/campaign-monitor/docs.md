# Overview

Campaign Monitor is a wave2 fan-out declarative-HTTP migration. It reads Campaign Monitor clients,
campaigns, subscriber lists, and suppression lists through the createsend.com v3.3 REST API
(`GET https://api.createsend.com/api/v3.3/...`). This bundle targets capability parity with
`internal/connectors/campaign-monitor` (the hand-written connector it migrates, Go package
`campaignmonitor`); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Campaign Monitor authenticates with HTTP Basic auth: the account/client API key is the username,
and the password may be blank or a dummy value. Provide the API key via the `username` config value
(required). An optional `password` secret is sent as the Basic auth password when set (`auth`
candidate 1: `when: "{{ secrets.password }}"`); when `password` is unset, the second candidate
(`mode: basic`, no `when`, always matches) sends the literal dummy password `"x"` instead — matching
legacy's own `cmDefaultDummyPasswrd` fallback (`campaign_monitor.go:258-262`) exactly via the
dual-auth-candidate first-match-wins pattern (`docs/migration/conventions.md` §3's zendesk-support
precedent). `base_url` defaults to `https://api.createsend.com/api/v3.3` and may be overridden for
tests/proxies.

## Streams notes

`clients` (`GET /clients.json`) and `lists` (`GET /clients/{{ config.client_id }}/lists.json`) are
bare top-level JSON arrays with no pagination (`pagination.type: none`), matching legacy's
`endpoint.paged == false` / `readArray` branch. `campaigns` and `suppressionlist` use Campaign
Monitor's page/`NumberOfPages` envelope (`{Results:[...], PageNumber, NumberOfPages}`) —
`pagination.type: page_number` with `page_param: page`, `size_param: pagesize`, `page_size: 100`;
records live at `Results`. A page shorter than 100 records stops pagination, functionally equivalent
to legacy's own `current >= numberOfPages` stop check for every real response (the last page is
never longer than the requested `pagesize`); an exact-multiple-of-100 result set costs one extra,
empty-page request on the engine side that legacy's page-count check would have avoided, never a
data difference.

`lists` and `campaigns` (as well as `suppressionlist`) are scoped under `/clients/{{ config.client_id
}}/...`; `client_id` is urlencoded into the path by `InterpolatePath`'s per-segment default, matching
legacy's own `url.PathEscape(clientID)` in `resolveResource`. An absent `client_id` hard-errors on
both sides (legacy: `"campaign-monitor config client_id is required for this stream"`; engine: an
unresolved `config.client_id` path-template key). Cursor fields match legacy's own declarations:
`campaigns` -> `SentDate`, `suppressionlist` -> `Date`; `clients`/`lists` have none (full refresh),
matching legacy's `CursorFields: nil` for both.

## Write actions & risks

None. Campaign Monitor is exposed read-only, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`; `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config overrides
  (`cmPageSize`/`cmMaxPages`, `campaign_monitor.go:318-346`). The engine's `page_number` paginator's
  `PageSize` is a static bundle-authored int (not templated), and there is no `MaxPages`-equivalent
  config-driven knob either; `page_size` is fixed at legacy's own default (100) in `streams.json`'s
  base pagination block, and `max_pages` is unbounded (matching legacy's own
  `max_pages=0`/`all`/`unlimited` default), following bitly's identical documented scope-narrowing
  precedent (`docs/migration/conventions.md`).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) synthesizes deterministic records with no network access
  (`campaign_monitor.go:216-244`); this bundle's schemas and fixtures target the live path only.
- Full Campaign Monitor API surface (subscribers, transactional email, templates, segments, list
  subscription writes) is out of scope; see `api_surface.json`'s `excluded` entries.
