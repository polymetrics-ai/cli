# Overview

Countercyclical is a financial-intelligence platform for investment teams; this bundle reads its
investments, valuations, research memos, teams, assumptions, and pipelines streams, and creates
investments, through the Countercyclical REST API. It migrates
`internal/connectors/countercyclical` (the legacy hand-written connector, kept registered and
unchanged until wave6's registry flip) to a declarative defs bundle at capability parity, then
expands past that legacy-parity floor (the legacy connector mirrors Countercyclical's own
Airbyte-marketplace source connector, which documents only 3 streams) with a Pass B review of the
live developer documentation at `docs.countercyclical.io/developers`, which documents a materially
larger surface than the Airbyte source connector exposes.

## Auth setup

Provide the Countercyclical API key via the `api_key` secret; it is sent as the `apiKey` query
parameter on every request (read AND write; matching the upstream `ApiKeyAuthenticator`,
`inject_into request_parameter` convention, and confirmed identical for the write endpoint by the
live docs) and never logged.

## Streams notes

All 6 streams (`investments`, `valuations`, `memos`, `teams`, `assumptions`, `pipelines`) are `GET`
requests returning a root-level JSON array (`records.path: "."` selects the body root), primary key
`["id"]`. The live developer docs confirm every one of these endpoints accepts only the `apiKey`
query parameter — no `limit`/`offset`/`page` parameter is documented on any of them (the platform's
separate `pagination.md` page describes `limit`/`offset`/`data`+`meta`-wrapped pagination as a
capability "some" endpoints support, illustrated with a `/v1/workspaces` example that is not one of
this connector's streams; none of the 6 streams here are among the paginated ones). This bundle
nonetheless declares `pagination.type: offset_limit` (`limit`/`offset` query params, `page_size:
100`) as a defensive measure exactly mirroring legacy's own `harvest` loop, which sends
`limit`/`offset` and continues only while a full page (`== pageSize`) comes back — a short/empty
page (the expected, honest case for a genuinely unpaginated API) stops after exactly one request.
Legacy only sends `offset` when it is greater than 0 (page 1 omits it entirely); the engine's
`offset_limit` paginator always sends `offset=0` explicitly on the first request. This is a
documented, non-data-changing parity deviation: `offset=0` and an absent `offset` param are
semantically identical for every standard offset-pagination API (including this one, per its own
docs), and unknown/redundant query params are ignored by APIs of this shape — see the
parity-deviation ledger below.

`teams`/`assumptions`/`pipelines` are new Pass B streams with the identical unpaginated
root-array shape; their documented response examples are truncated with `...` (the docs do not
enumerate every field), so each schema declares only the fields the live docs actually show
(`teams`: `id`/`title`; `assumptions`: `id`/`name`/`discountRate`; `pipelines`: `id`/`name`) rather
than guessing at undocumented ones — see Known limits.

No stream declares an `incremental` block: the upstream manifest-only source connector this
migrates advertises full-refresh only (no incremental cursor), matching legacy's own
`streamCatalog()`, which leaves `CursorFields` empty for every stream; the live docs confirm none of
the 6 streams' list endpoints accept a since/updated-since filter parameter either.

## Write actions & risks

One write action, `create_investment` (`POST /integrations/make/actions/investments`), `external
mutation` / `approval required` per `metadata.json`'s `capabilities.write: true`. It creates a new
Investment in the caller's workspace from a single required field, `tickerSymbol`. This is the only
general-purpose creation endpoint the live docs publish for any of the 6 streams (valuations/memos/
teams/assumptions/pipelines document no creation endpoint at all); it happens to live under a
Make-integration-specific path, but its docs confirm it uses the same `apiKey` query-param auth as
every read endpoint (not a Make-specific OAuth flow), so it is a legitimate write action for this
connector rather than a Make-app-only capability. An functionally-identical Zapier-integration
endpoint (`POST /integrations/zapier/actions/investments`, same auth, same `tickerSymbol`-only
body) is documented too but not separately exposed — see `api_surface.json`'s `duplicate_of` entry.

No write action is exposed for valuations/memos/teams/assumptions/pipelines: the live docs document
no creation/update/delete endpoint for any of them.

## Known limits

- Legacy's configurable `page_size` (1-1000, default 100) and `max_pages` (0/all/unlimited or a
  positive integer cap) config knobs are not modeled: `streams.json`'s `pagination.page_size` is a
  fixed JSON literal with no config-driven override mechanism (same class of limitation as
  searxng's `page_size`/`max_pages`). A declared-but-unwireable config key is worse than an absent
  one (F6, REVIEW.md), so neither is declared in `spec.json`; the stop threshold is fixed at 100,
  matching legacy's own default exactly.
- Parity deviation (meta-rule, `docs/migration/conventions.md` §5): the engine's `offset_limit`
  paginator sends `offset=0` explicitly on the very first request; legacy omits the `offset` query
  param entirely when its value is 0. This never changes emitted record data for any input legacy
  itself would accept — `offset=0` and an absent `offset` are the same request to any
  standards-conforming offset-pagination API, and this API's own docs confirm unknown/redundant
  params are ignored. ACCEPTABLE.
- `teams`/`assumptions`/`pipelines` schemas are deliberately narrow: the live docs' own response
  examples are truncated (`{"id": "...", "name": "...", ...}`) and do not enumerate every field the
  real API returns. Rather than guess at undocumented fields (`SCHEMA_AMBIGUOUS` risk), each schema
  declares only the fields the docs actually show. A future capability-expansion pass with access to
  a live account (or a more complete OpenAPI-style spec, which Countercyclical does not currently
  publish) should widen these schemas to the real full field set.
- Webhooks are documented as a Settings-UI-managed Enterprise-only feature with no CRUD API
  endpoint published at all (16 event types across the 6 resource kinds, each `.created`/`.updated`/
  `.deleted`) — there is no `create_webhook`/`list_webhooks`/`delete_webhook` write action or stream
  to expose; see `api_surface.json`.
