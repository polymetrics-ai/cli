# Overview

EasyPost is a wave2 fan-out declarative-HTTP migration. It reads EasyPost shipments, trackers,
addresses, parcels, and insurances through the EasyPost REST API
(`GET https://api.easypost.com/v2/...`). This bundle migrates `internal/connectors/easypost` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip. EasyPost is read-only: legacy exposes no reverse-ETL writes, so `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Auth setup

Provide the EasyPost API key via the `username` secret; it is sent as the HTTP Basic username with
an empty password (`Authorization: Basic base64(<api_key>:)`), matching legacy's
`connsdk.Basic(secret, "")` exactly (`easypost.go`'s `requester`). `base_url` defaults to
`https://api.easypost.com/v2` and may be overridden for tests/proxies.

## Streams notes

All 5 streams (`shipments`, `trackers`, `addresses`, `parcels`, `insurances`) share the same list
shape: `GET` against the resource-named endpoint, records at the resource-named JSON key
(`{"shipments":[...],"has_more":bool}`, etc.), and EasyPost's `before_id`/`has_more`
newest-first-then-page-older convention (`pagination.type: cursor` with
`last_record_field: id` and `stop_path: has_more`) — the next page's `before_id` is the `id` of the
LAST record on the current page, and pagination stops when `has_more` is falsy, matching legacy's
`harvest` function exactly. Every stream declares `incremental.cursor_field: created_at` with
`request_param: start_datetime` and `start_config_key: start_date`: the incremental lower bound
(persisted cursor, or the `start_date` config value on a fresh sync) is sent as `start_datetime`
only when it resolves to a non-empty value — `read.go`'s `buildInitialQuery` only sets
`request_param` when the formatted lower bound is non-empty, reproducing legacy's
`startDateBound` (`cursor` first, else `config.start_date`, else omitted entirely) with no
extra dialect needed.

## Write actions & risks

None. EasyPost is a read-only source in legacy (`easypost.go`'s package doc: reverse-ETL writes
would mutate live shipping/insurance objects, so writes are explicitly unsupported);
`capabilities.write` is `false` and no `writes.json` is shipped.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (default
  100, capped at 100) and `max_pages` (0/all/unlimited = unbounded) as config-driven overrides
  (`easypostPageSize`/`easypostMaxPages` in `easypost.go`). The engine's `cursor` paginator has no
  config-driven page-size or max-pages knob, so this bundle sends legacy's own default
  (`page_size=100`) as a static per-stream query literal and does not declare `page_size`/`max_pages`
  in `spec.json` at all (a declared-but-unwireable config key is worse than an absent one, per
  conventions.md F6 precedent). Pagination is bounded only by EasyPost's own `has_more` stop
  signal, matching EasyPost's real termination behavior.
- **Legacy's fixture-mode-only synthetic fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps a `previous_cursor` field (echoing
  `req.State["cursor"]`) that has no equivalent in the live wire shape. This bundle's schemas and
  fixtures target the LIVE record shape only (`easypost.go`'s `harvest`/`mapRecord` functions), per
  the bitly-pilot precedent (`docs/migration/conventions.md`'s worked example): the engine's own
  fixture-replay conformance harness supersedes the need for an in-connector fixture-mode branch.
