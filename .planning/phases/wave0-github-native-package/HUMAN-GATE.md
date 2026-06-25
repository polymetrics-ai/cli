# HUMAN GATE — DuckDB dependency + CGO (Wave 0 item 2)

## Why this is a gate
GSD protocol requires human approval before **adding a third-party dependency** and before a change
that alters the build contract. The DuckDB-backed warehouse + real `Querier` needs:
- a new Go module dependency (e.g. `github.com/marcboeker/go-duckdb`), and
- **CGO** (a C toolchain at build time), which breaks the repo's current "dependency-free,
  pure-Go, easy cross-compile" property.

## Proposed approach (for approval)
- Add go-duckdb behind a `duckdb` build tag. Default builds stay pure-Go and CGO-free, using the
  existing JSONL warehouse + simple `Querier`. `go build -tags duckdb` enables the DuckDB warehouse
  with real analytical SQL (joins/aggregations/window functions) and routes `app.QuerySQL` to it
  with SELECT-only safety (extend `validateNativeSelect`).
- This keeps `make verify` green in the default (CGO-free) configuration and adds a second
  `make verify-duckdb` lane for the tagged build.

## Status
- Phase `wave0-github-native-package` (item 1: GitHub native package + registry) is **completed/green**.
- Item 2 (DuckDB) is **blocked_human_gate** pending the decision below.

## Decision needed
Approve the CGO/go-duckdb dependency now (build-tag gated), choose a pure-Go alternative, or defer
DuckDB and proceed to Wave 1 connectors first (which do not require DuckDB).
