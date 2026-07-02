# Overview

PokeAPI is a wave2 fan-out declarative-HTTP migration. It reads Pokemon, types, abilities, and
moves from the public PokeAPI (`GET https://pokeapi.co/api/v2/<resource>`). This bundle is a
capability-parity port of `internal/connectors/pokeapi` (the hand-written connector it migrates);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

No credentials are required: PokeAPI is a public, open reference API. `base_url` defaults to
`https://pokeapi.co/api/v2` and may be overridden for tests/proxies.

## Streams notes

All four streams (`pokemon`, `type`→`types`, `ability`→`abilities`, `move`→`moves`) are named-
resource list endpoints returning a `{count, next, previous, results[]}` envelope; records are
extracted from `results`. Every record computes an `id` field via the `last_path_segment` filter
over the raw `url` (`{{ record.url | last_path_segment }}`), matching legacy's
`idFromURL(urlValue)` (`pokeapi.go:151-158`), which trims a trailing slash and returns the final
`/`-delimited segment (PokeAPI's own resource-id-in-URL convention, e.g.
`.../pokemon/1/` -> `"1"`). `name` and `url` pass through schema projection unchanged. Primary key
is `name` (matching legacy's `PrimaryKey: []string{"name"}` stream declaration — PokeAPI resource
names are the stable identifier across pagination, not the numeric `id` derived from the URL).

Pagination is `offset_limit` (`limit`/`offset` query params, matching legacy's
`connsdk.OffsetPaginator{LimitParam:"limit", OffsetParam:"offset", PageSize:pageSize}`,
`pokeapi.go:93`), 100 records per page and a default `max_pages: 3` cap, both matching legacy's
`defaultPageSize`/`defaultMaxPages` constants exactly.

## Write actions & risks

None. PokeAPI's legacy connector is read-only (`Capabilities.Write: false`); this bundle ships no
`writes.json`.

## Known limits

None beyond the Pass B API-surface narrowing recorded in `api_surface.json`. Every field legacy's
`namedResourceRecord` emits (`id`, `name`, `url`) is modeled without approximation.
