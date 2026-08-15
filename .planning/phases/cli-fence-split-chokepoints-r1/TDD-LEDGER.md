# TDD ledger — Fence split chokepoints r1

| ID | Guarantee | Red evidence | Green proof |
| --- | --- | --- | --- |
| R1 | Connector-local certification truth | A shard-reconstruction test could not compile because no sharding API existed. | Passed: the two shards reconstruct the scoped capability and flow aggregate; baseline is derived in memory. |
| R2 | Stable source anchors | A source-anchor test found `file:line` anchors. | Passed: every generated Go anchor is `file:Symbol`; a one-line executor insertion left all generated hashes unchanged. |
| R3 | Honest scoped drift detection | A changed/missing shard could evade the old aggregate-only checker. | Passed: `certification-matrix --check` validates and byte-compares allowlisted shards; non-allowlisted evidence is filtered from scoped generation. |
| R4 | App split is mechanical | Existing app ETL and Open-path tests are the baseline behavioural contract. | Passed: the same focused app suite passes after moving only composition and mode-dispatch blocks. |
| R5 | PostgreSQL capability parity | Existing PostgreSQL tests require CDC to remain false. | Passed: the row table applies the same `CDC=false` override and the native package suite passes. |
| R6 | Incremental lane isolation | A scoped generator run had no API or rewrote every shard. | Passed: the generator core test runs GitHub scope and confirms PostgreSQL plus status are byte-identical. |
| R7 | CI boundary and standard-library security | CI rejected the intentional GitHub allowlist literal as shared connector policy, and `govulncheck` found reachable Go 1.25.12 standard-library vulnerabilities. | Passed: the one-match, expiring boundary exception keeps the explicit allowlist reviewable; whole-tree boundary checks are clean and Go 1.25.13 `govulncheck` reports zero reachable vulnerabilities. |
| R8 | Generator isolation from non-allowlisted runtime ledger entries | `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestCertificationCheckIgnoresMalformedNonAllowlistedRuntimeLedgerEntry$'` failed at `b9e294a2`: its copied real source tree changed only `mysql` in `operation_endpoint_ledger.json` to `[{"unexpected":true}]`, then the real `go run ./cmd/connectorgen certification-matrix --check` panicked through `native/postgres.New()`. | Passed: the same command-level fixture now exits 0 with no stderr after the PostgreSQL matrix source is constructed from the generator's already-scoped bundle. `certification-matrix --check` confirms the committed shards remain byte-identical, and `git diff --exit-code 2df18ee -- internal/connectors/engine/bundle.go` is empty. |

## Red command

```sh
go test -count=1 ./cmd/connectorgen -run 'TestCertification(ShardsRoundTripGeneratedMatrices|ScopedGenerationLeavesOtherShardByteIdentical|SourceAnchorsUseSymbols)'
```

The pre-implementation failure is retained in `traces/red-run.txt`. The identical focused command
must pass after implementation.

## R8 red command

The regression runs the command path in an isolated copy of the real generator
inputs, mutates only the real non-allowlisted `mysql` ledger entry, and runs the
actual `go run ./cmd/connectorgen certification-matrix --check` command. It failed
at `b9e294a2` with the documented native PostgreSQL factory panic and passes after
the generator-only construction change.
