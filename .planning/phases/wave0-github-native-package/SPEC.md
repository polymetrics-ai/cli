# SPEC — GitHub Native Package + Data-Driven Registry

## Package layout (target)
```
internal/connectors/github/
  github.go         # Connector type: Name/Metadata/Check/Catalog/Read/Write (+Write actions)
  auth.go           # auth modes (token / github_app JWT→installation token / public)
  streams.go        # stream + field + record definitions
  github_test.go    # package-level tests (moved/adapted from connectors/github_test.go)
internal/connectors/
  registry_gen.go   # self-registration: factories appended via RegisterFactory(); NewRegistry consumes them
```

## Registration mechanism
- Add to `package connectors` a package-level factory registry:
  - `func RegisterFactory(name string, f func() Connector)` appends to an ordered slice/map.
  - `NewRegistry()` keeps the built-in primitives (Sample/File/Warehouse/Outbox) and then iterates
    registered factories, then the existing `enabled` catalog-alias logic — preserving today's
    behavior (github + source-github both resolve, CatalogAliasConnector still applies).
- `internal/connectors/github` exposes `func New() connectors.Connector` and, in an `init()`,
  calls `connectors.RegisterFactory("github", New)`. A blank import in `registry_gen.go`
  (`_ "polymetrics/internal/connectors/github"`) wires it in.
- Import-cycle: the github package imports `connectors` (for types/interface) and `connsdk`;
  `connectors` must NOT import `github` except as a blank import in `registry_gen.go` — which is
  fine because blank imports only run `init()` and introduce no symbol dependency that the
  github package depends back on at compile time of `connectors`' own symbols. (Standard Go
  plugin-registry pattern.)

## Behavior parity (must hold)
- `Github{}.Name()` == "github"; metadata, capabilities, auth modes, streams, write actions,
  and conformance identical to current behavior.
- `source-github` catalog entry (implementation_status=enabled, pm_connector_name=github) still
  produces a working `CatalogAliasConnector` targeting the github connector.
- `pm connectors inspect github --json` → kind "Connector" with full manifest (already fixed).

## connsdk adoption (where it cleanly applies, optional this phase)
- The github HTTP request helper (`doJSONWithAuth`) and Link-header pagination map onto
  `connsdk.Requester` + `connsdk.LinkHeaderPaginator`. Adopt opportunistically WITHOUT changing
  observable behavior or weakening the GitHub App JWT auth (which stays GitHub-specific).
- If adoption risks behavior drift, keep the proven GitHub HTTP code as-is inside the package and
  defer connsdk refactor — parity and green tests outrank refactor purity this phase.

## Out of scope
DuckDB, catalog pair-merge, other connectors.
