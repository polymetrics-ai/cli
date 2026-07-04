# Overview

Oveit is an event ticketing/access-control platform. This bundle is a Tier-1 declarative migration
of `internal/connectors/oveit` (legacy Go package, read-only): it reads `events`, `orders`, and
`attendees` from legacy's own fictional API shape (`GET https://oveit.com/api/{events,orders,
attendees}`, HTTP Basic auth). The legacy package stays registered and unchanged until wave6's
registry flip.

## API surface (Pass B) — the real API does not match legacy's shape at all

`metadata.json`'s `docs_url` (previously `https://oveit.com/api/`) 404s live; the real, current
Oveit API documentation was located at `https://l.oveit.com/api-documentation/` (found via web
search, since the stale docs_url no longer resolves to any content) and `docs_url` has been
updated to point there. Researching it surfaced a fact this bundle's original migration predates:
**the real Oveit API bears no resemblance to what legacy — and, by inheritance, this bundle —
implement.** The real API is entirely `POST`-based against paths like `/seller/events`,
`/events/attendees`, `/crm/orders`, `/tickets/checkin`, `/wallet/credit`, etc., authenticated by a
token-exchange flow (`POST /seller/login` with an organizer email+password returns an access token
with its own `expires_at`, sent as a request **parameter** — or, per one endpoint's own docs, a
Bearer header — on every subsequent call), never HTTP Basic auth against endpoints literally named
`/events`/`/orders`/`/attendees`. Legacy's own package comment already flags its approach as
approximate: "a conservative read-only connector for documented Oveit API resources... sends
\[email/password\] as HTTP Basic auth."

`api_surface.json` now documents the REAL, full Oveit endpoint surface (recovered from
`l.oveit.com`'s Events/Tickets/Attendees/RFID/Virtual-wallets/Advanced-data/Turnstile/Channel-
Partner-Management doc pages) instead of repeating the fictional 3-endpoint surface legacy
invented. Every real endpoint is excluded as `requires_elevated_scope` or `duplicate_of`: the root
cause is that Oveit's real `POST /seller/login` token-exchange auth has no equivalent among this
dialect's declarative auth modes (`bearer`/`basic`/`api_key_header`/`api_key_query`/
`oauth2_client_credentials`) — it is a textbook token-exchange `AuthHook` shape (design conventions
Tier-2 table: "token-exchange auth (GitHub App JWT→installation token)") — and this pass's
instructions explicitly forbid creating a new hook package. This bundle's 3 existing streams keep
legacy's fictional shape completely unchanged (not migrated to the real API), since correctly
migrating them would require exactly the hook this pass cannot add. This gap is reported, not
silently worked around — see Known limits.

## Auth setup

Legacy requires both `email` (a plain config value, not a secret) and `password` (a secret),
combined into HTTP Basic auth (`Authorization: Basic base64(email:password)`). This bundle wires
the identical shape via `streams.json` `base.auth`: `{"mode":"basic","username":"{{ config.email
}}","password":"{{ secrets.password }}"}`. Both `email` and `password` are `required` in
`spec.json`, matching legacy's hard error when either is missing
(`oveit.go:147-149`, `"oveit connector requires config email and secret password"`).

`base_url` defaults to `https://oveit.com/api` (legacy's `defaultBaseURL`), materialized via
`spec.json`'s `"default"` mechanism — an unset `base_url` now round-trips to the same default
legacy applied in code.

## Streams notes

All three streams (`events`, `orders`, `attendees`) share an identical record shape and pagination
behavior, matching legacy's single `streamEndpoint`/`record()` mapping applied uniformly across all
three endpoints. Records are extracted from the top-level `data` array. Primary key is `id`; there
is no incremental cursor (legacy never filters or advances reads by a timestamp field — every read
is a full stream read), so no `x-cursor-field`/`incremental` block is declared.

Pagination is `cursor` (`token_path`) reading the next page number from the response body's
`next_page` field, matching legacy's own `harvest` loop (`oveit.go:94-121`): the cursor param name
is `page`, and legacy stops as soon as `next_page` is absent/empty. `per_page` is sent on every
request via `config.page_size` (default 100, legacy's `defaultPageSize`); legacy additionally
enforces a hard max of 500, which is documentation-only in this bundle (see Known limits).

## Write actions & risks

None. Oveit's legacy connector is read-only (`Write` always returns `ErrUnsupportedOperation`);
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **First-page request omits the `page` query param entirely.** Legacy's `harvest` loop explicitly
  sends `page=1` on its first request. The engine's `cursor`(`token_path`) paginator
  (`tokenPathCursor.Start()`) always issues the first request with no cursor param at all, only
  adding `page=<token>` once a `next_page` value is read back from a prior response. This is a
  request-shape difference only, never an emitted-record-data difference (Oveit's API, like any
  standard paginated REST list endpoint, treats an absent page param as "first page" — legacy's own
  fixture/test harness never distinguishes an explicit `page=1` from an absent one). Documented per
  the parity-deviation meta-rule (`docs/migration/conventions.md` §5); acceptable since it cannot
  change accepted-input behavior.
- **`page_size` upper bound (500) is not enforced by this bundle.** Legacy's `pageSize` helper
  rejects a `page_size` config value outside `[1, 500]` with a hard config error
  (`oveit.go:182-192`). The engine's declarative config layer has no numeric-range validation
  primitive for `spec.json` properties; `connectorgen validate`'s schema-shape checks cover type,
  not bounds. Not modeled as a value constraint; documented as scope narrowing since an
  out-of-range `page_size` value is a caller-configuration error, not something an emitted record
  would ever reflect. A caller supplying a wildly out-of-range value would simply have it forwarded
  verbatim to Oveit's `per_page` query param and rely on Oveit's own server-side clamping/error
  response.
- **`max_pages` "all"/"unlimited" string aliases are not modeled.** Legacy's `maxPages` helper
  accepts the literal strings `"all"`/`"unlimited"` (case-insensitive) as synonyms for "0 = no cap"
  in addition to an empty value. This bundle has no `max_pages` config property at all (pagination
  is capped only by the API's own `next_page` exhaustion, matching the common case of an unset
  `max_pages` in legacy) — a caller needing a hard page-count cap can rely on `MaxPages`-equivalent
  behavior once/if this bundle adds a `max_pages` spec property in a later Pass B increment.
- **BLOCKED (`AUTH_COMPLEX`): the real Oveit API cannot be reached at all with this bundle's
  current auth, and no real endpoint from the real API (events, orders, attendees, tickets,
  check-in/scan, RFID pairing, virtual wallets, turnstile, channel-partner-management) can be added
  as a new stream or write action until this is resolved.** Oveit's real auth is `POST
  /seller/login` (organizer email+password) → an access token with its own `expires_at`, sent as a
  request parameter (or Bearer header, per one endpoint) on every subsequent call — a
  token-exchange flow this dialect's declarative `auth` modes cannot express (no
  `oauth2_client_credentials`-shaped grant fits; it needs a genuine `AuthHook`, e.g. github's
  RS256-JWT-to-installation-token exchange). This pass's instructions forbid creating a new hook
  package, so this is reported as a blocker rather than worked around: resolving it requires either
  a Tier-2 `hooks/oveit/` package (a future increment, out of this pass's scope) or accepting that
  this connector's true capability ceiling is the 3 streams legacy already (fictionally)
  approximates. See `api_surface.json` for the full real-endpoint-by-real-endpoint breakdown.
