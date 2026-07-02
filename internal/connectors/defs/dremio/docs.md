# Overview

Dremio reads catalog entries, reflections, sources, and users through the Dremio REST API
(defaulting to the Dremio Cloud US root, `https://api.dremio.cloud/v0`). This bundle migrates
`internal/connectors/dremio` (legacy) at capability parity; the legacy package stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide a Dremio Personal Access Token (PAT) via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`, `bearer` auth mode) and never logged, matching legacy's
`connsdk.Bearer(secret)` wiring exactly.

## Streams notes

All 4 streams (`catalog`, `reflections`, `sources` (path `/source`), `users` (path `/user`)) share
the same envelope: `GET` against the resource path, records at `data`, primary key `["id"]`.
Pagination is `cursor` with `token_path: nextPageToken` and `cursor_param: pageToken` — Dremio
returns `{"data":[...], "nextPageToken":"..."}`; the next page is requested with
`pageToken=<nextPageToken>`, and the engine's `tokenPathCursor` paginator stops on a
null/absent/non-advancing token exactly like legacy's `harvest` loop (`strings.TrimSpace(next) ==
""`), plus an empty-page stop. No `stop_path` is declared: legacy's stop condition is driven purely
by the token itself, so this bundle preserves that exact stop-on-empty-token-only behavior
(conventions.md §3).

Every request sends `maxResults` (default 100, matching legacy's `dremioDefaultPageSize`) via each
stream's static `query`. None of the four streams have a genuine incremental filter in the legacy
connector (its own doc comment: "none of these list endpoints expose a reliable incremental
cursor"), so no schema declares `x-cursor-field` and no stream declares an `incremental` block —
these streams are full-refresh only, matching legacy's `dremioStreams()` (no `CursorFields` set on
any of the 4 `connectors.Stream` definitions).

## Write actions & risks

None. Dremio's write endpoints (SQL job submission, source/reflection create/update/delete) have no
legacy-parity implementation to migrate; legacy itself is read-only (`Capabilities.Write: false`),
so no `writes.json` is shipped.

## Known limits

- Only the 4 legacy-parity read streams (`catalog`, `reflections`, `sources`, `users`) are
  implemented; the full Dremio API surface (SQL execution, source/reflection mutation) is out of
  scope for this wave — see `api_surface.json`'s `excluded` entries.
- Dremio Cloud vs. Dremio Software/self-hosted vs. EU-region deployments use different base URLs;
  legacy resolves only a flat `base_url` config override with no derivation logic, and this bundle
  matches that exactly (`base_url` default is the Dremio Cloud US root; any other deployment must
  set `base_url` explicitly).
