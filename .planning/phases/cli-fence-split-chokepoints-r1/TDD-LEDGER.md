# TDD ledger — Fence split chokepoints r1

| ID | Guarantee | Red evidence | Green proof |
| --- | --- | --- | --- |
| R1 | Connector-local certification truth | A shard-reconstruction test could not compile because no sharding API existed. | Passed: the two shards reconstruct the scoped capability and flow aggregate; baseline is derived in memory. |
| R2 | Stable source anchors | A source-anchor test found `file:line` anchors. | Passed: every generated Go anchor is `file:Symbol`; a one-line executor insertion left all generated hashes unchanged. |
| R3 | Honest scoped drift detection | A changed/missing shard could evade the old aggregate-only checker. | Passed: `certification-matrix --check` validates and byte-compares allowlisted shards; non-allowlisted evidence is filtered from scoped generation. |
| R4 | App split is mechanical | Existing app ETL and Open-path tests are the baseline behavioural contract. | Passed: the same focused app suite passes after moving only composition and mode-dispatch blocks. |
| R5 | PostgreSQL capability parity | Existing PostgreSQL tests require CDC to remain false. | Passed: the row table applies the same `CDC=false` override and the native package suite passes. |
| R6 | Incremental lane isolation | A scoped generator run had no API or rewrote every shard. | Passed: the generator core test runs GitHub scope and confirms PostgreSQL plus status are byte-identical. |

## Red command

```sh
go test -count=1 ./cmd/connectorgen -run 'TestCertification(ShardsRoundTripGeneratedMatrices|ScopedGenerationLeavesOtherShardByteIdentical|SourceAnchorsUseSymbols)'
```

The pre-implementation failure is retained in `traces/red-run.txt`. The identical focused command
must pass after implementation.
