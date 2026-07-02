# Overview

Onfleet is a read-only declarative-HTTP migration (wave2 fan-out) of `internal/connectors/onfleet`
(the hand-written connector it replaces at capability parity). It reads tasks, workers, teams, hubs,
and administrators through the Onfleet v2 REST API. The legacy package stays registered and unchanged
until wave6's registry flip.

## Auth setup

Provide the Onfleet API key as the `api_key` secret. Onfleet authenticates with HTTP Basic auth where
the API key is the username and the password is intentionally blank
(`{"mode": "basic", "username": "{{ secrets.api_key }}", "password": ""}`), matching legacy
`onfleet.go`'s `connsdk.Basic(apiKey, "")` exactly — Onfleet expects a blank password, not a second
credential. The key is never logged.

## Streams notes

`tasks` reads `GET /tasks/all`, which returns `{lastId, tasks:[...]}`; this is a body-token cursor
(`pagination.type: cursor`, `token_path: lastId`, `cursor_param: lastId`) with no `stop_path` declared
— matching legacy's own stop condition exactly (an absent/empty `lastId` in the response body ends
pagination; the engine's `tokenPathCursor` already treats an empty/missing token as the stop signal
with no additional configuration needed).

`workers`, `teams`, `hubs`, and `administrators` are non-paginated top-level JSON arrays
(`records.path: "."`, `pagination: {"type": "none"}` overriding the base cursor pagination for these
4 streams), matching legacy's `arrayPath: ""` / `paginated: false` shape.

Every stream's primary key is `id`; `timeLastModified` is declared as `x-cursor-field` wherever legacy
published it (`tasks`, `workers`, `teams`, `administrators`) — `hubs` has no cursor field, matching
legacy's `CursorFields: nil`. As in legacy, no stream actually issues a server-side incremental filter
(Onfleet's API supports no query timestamp filter); every sync is a full refresh.

## Write actions & risks

None. Onfleet's write surface (creating/updating live delivery tasks) is operationally sensitive and
was deliberately never implemented in legacy (`Capabilities.Write: false` in both); this bundle ships
no `writes.json`.

## Known limits

- Onfleet's broader API surface (organization details, webhooks, task creation/mutation) is out of
  scope; see `api_surface.json`'s `excluded` entries.
- No stream supports a real incremental filter; `x-cursor-field` values are manifest-parity only
  (matching legacy's own published `CursorFields`, which legacy also never used to filter requests).
