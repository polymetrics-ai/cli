# ADR — DuckDB query engine behind a build tag

## Status
Accepted (Wave 0, item 2). Dependency add approved by human gate (HUMAN-GATE.md).

## Context
The platform needs real analytical SQL (joins/aggregations/window) for the query→analyze→act loop.
DuckDB is the natural fit but requires CGO (`github.com/marcboeker/go-duckdb`), which conflicts with
the repo's current pure-Go, dependency-light, easily-cross-compiled build.

## Decision
1. Add go-duckdb **behind a `duckdb` build tag**. Default builds remain pure-Go / CGO-free using the
   existing JSONL warehouse + `parseSelectAll`. `-tags duckdb` swaps in a DuckDB engine.
2. **Storage stays JSONL this phase.** DuckDB queries the JSONL files via `read_json_auto` views —
   no ETL write-path change, lowest risk. Parquet/native-DuckDB storage is a deferred optimization.
3. SELECT-only enforcement on all queries; DuckDB runs read-only over per-query views.
4. Two verify lanes: `make verify` (default, CGO-free) and `make verify-duckdb` (tagged, CGO).

## Consequences
- (+) Full analytical SQL when desired; default build keeps its portability guarantees.
- (+) Reversible: deleting the tagged files + `go mod tidy` removes the dependency.
- (−) Two build configurations to keep green; CGO toolchain required for the tagged lane.
- go-duckdb v1.8.5 ships prebuilt static libs for darwin/arm64 (and common platforms) — no DuckDB
  source compile.

## Alternatives rejected
- Pure-Go SQL engine (weaker analytics; user chose DuckDB).
- Make DuckDB the default (breaks CGO-free / cross-compile property).
- DuckDB-native storage now (larger, riskier write-path change; deferred).
