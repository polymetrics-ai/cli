# Overview

Freshdesk is a customer-support helpdesk platform. This bundle reads Freshdesk tickets, contacts,
companies, agents, and groups through the Freshdesk REST API v2. It is read-only, matching legacy
`internal/connectors/freshdesk` exactly (`Capabilities{Write: false}`).

## Auth setup

Provide a Freshdesk API key via the `api_key` secret. It is sent as the HTTP Basic username, with
the literal password `X` — Freshdesk's own documented convention (`api_key:X`) — via `base.auth`'s
`mode: basic`. Never logged.

## Streams notes

All 5 streams (`tickets`, `contacts`, `companies`, `agents`, `groups`) share the same shape: `GET`
against the Freshdesk list endpoint, records at the top-level JSON array (`records.path: "."`),
primary key `["id"]`, `x-cursor-field: updated_at` (every Freshdesk object exposes
`created_at`/`updated_at`, matching legacy's `CursorFields: ["updated_at"]` on every stream).
Pagination is `link_header` (RFC 5988 `Link: <url>; rel="next"`), identical to legacy's
`connsdk.LinkHeaderPaginator` usage.

Only `tickets` declares an `incremental` block (`cursor_field: updated_at`, `request_param:
updated_since`, `param_format: rfc3339`, `start_config_key: start_date`) — matching legacy exactly,
which gates its `updated_since` query param behind `if stream == "tickets"` and sends the
`start_date` config value verbatim (no reformatting), the same behavior `param_format: rfc3339`
(the default, sent-as-is) produces. The other 4 streams (`contacts`, `companies`, `agents`,
`groups`) apply no server-side incremental filter at all — again matching legacy, which never sets
`updated_since` for them — while still declaring `x-cursor-field: updated_at` for
`incremental_append_deduped` sync-mode eligibility, mirroring legacy's own stream catalog
declaration of `CursorFields` without any actual per-stream server-side filter.

## Write actions & risks

None. Freshdesk is exposed read-only, matching legacy's `Capabilities{Write: false}` and its
`Write` method, which always returns `connectors.ErrUnsupportedOperation`.

## Known limits

- Full Freshdesk API surface (conversations, time entries, solutions/articles, canned responses,
  business hours, webhooks, etc.) is out of scope for this wave; see `api_surface.json`'s
  `excluded: {category: out_of_scope, reason: "not implemented in this bundle"}` entries. Only the 5
  legacy-parity read streams are implemented.
- **`base_url` is required config, not derived from `domain`.** Legacy derives the base URL from a
  bare `domain` config value when `base_url` is unset (`https://<domain>/api/v2`), validating the
  domain has no `/?#` characters and stripping any `http(s)://` prefix. The engine's `spec.json`
  `"default"` materialization mechanism only supports a FIXED literal default, not one derived from
  another config value at read/check time — there is no declarative way to express "derive
  base_url from domain" without inventing Go for a single string-templating rule, which would be
  new undeclared logic outside the dialect. This bundle therefore requires the fully-formed
  `base_url` directly (e.g. `https://acme.freshdesk.com/api/v2`), matching the same accepted
  precedent as chargebee and repairshopr in this repo. This is a documented config-surface
  narrowing (never a data-parity change).
- **`link_header` pagination has no genuine 2-page fixture in this bundle.** The
  `fixtures/streams/**` replay harness (`internal/connectors/conformance/replay.go`'s
  `fixtureResponse`) has no mechanism to set custom response HEADERS from a fixture file at all
  (only `status`/`body` are recordable) — a `Link: <url>; rel="next"` response header, the entire
  signal `link_header` pagination follows, can never be emitted by the fixture replay server
  regardless of how the fixture JSON is shaped. This is the identical structural gap
  `docs/migration/conventions.md` §4 already documents and sanctions for `next_url` pagination
  (whose next-page URL is unknowable at fixture-authoring time) — here the URL itself is not even
  the blocker; the header transport mechanism is unavailable at all. Every stream therefore ships a
  single-page fixture (satisfies `fixtures_present`/`read_fixture_nonempty`/
  `pagination_terminates`, which passes trivially — no `Link` header in the response means the
  paginator correctly stops after page 1, proving termination but not proving a real second-page
  follow). Unlike the sanctioned `next_url` exception, this wave does not ship a live
  `paritytest/<name>` test to prove genuine 2-page `Link`-header-follow correctness (out of scope
  for this JSON+docs-only fan-out wave per the migration's hard rules) — that live-parity
  verification is deferred to a follow-up wave. This is a documented, honest scope gap, not a
  fabricated fixture or a silent correctness assumption: `internal/connectors/connsdk`'s own
  `LinkHeaderPaginator`/`parseLinkNext` unit tests (`connsdk/paginate_test.go`) already cover the
  Link-header-follow mechanism generically outside this bundle.
