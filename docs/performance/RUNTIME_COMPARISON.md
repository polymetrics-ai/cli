# Runtime Performance Comparison

Date: 2026-06-24

## What Dependency-Free Means

The dependency-free version is the local MVP path. It requires no database, cache, or workflow server.

It uses:

- local JSON state in `.polymetrics/state/state.json`
- AES-GCM encrypted vault files in `.polymetrics/vault`
- JSONL warehouse files in `.polymetrics/warehouse`
- JSONL reverse ETL outbox files in `.polymetrics/outbox`
- in-process ETL and reverse ETL orchestration

This mode is useful for local development, agent-safe command prototyping, offline demos, connector contract development, and small data movement workflows.

## What Runtime-Backed Means

The runtime-backed version keeps the local ETL loop but adds external runtime coordination:

- PostgreSQL run-ledger persistence
- DragonflyDB lease coordination
- Temporal health checks as the durable workflow orchestration target

Implemented commands:

```bash
./pm runtime doctor --json
./pm perf compare --iterations 50 --json
./pm perf compare --iterations 50 --runtime --json
./pm etl run --connection <name> --stream <stream> --runtime --json
```

## Measured Result

Command:

```bash
./pm perf compare --iterations 50 --json
```

Result:

```text
mode: dependency-free
iterations: 50
records: 150
duration: 38.28725ms
average: 765.745us per ETL loop
throughput: 3917.75 records/sec
```

Latest dependency-free verification during the agentic ETL phase:

```text
iterations: 50
records: 150
duration: 44.061125ms
average: 881.222us per ETL loop
throughput: 3404.36 records/sec
```

## Runtime-Backed Status

Command:

```bash
./pm perf compare --iterations 50 --runtime --json
```

Result:

```text
mode: runtime-backed
status: degraded
reason: runtime services are not all healthy
```

Runtime health:

```text
PostgreSQL localhost:15433: connection refused
DragonflyDB localhost:6379: ok
Temporal localhost:7233: connection refused
```

The local runtime stack could not be fully verified during the agentic ETL phase because `make runtime-up` did not finish the Temporal health wait within the verification window. The startup was interrupted and `make runtime-down` successfully removed the partial stack.

## Interpretation

The dependency-free path is fast for the current tiny sample connector because it performs all work in-process and writes small JSONL files locally.

The runtime-backed path is expected to have higher per-run overhead for small jobs because it adds network checks, lease coordination, and durable ledger writes. That overhead is intentional. It buys operational durability, coordination, and a path toward Temporal-managed retries and resumes.

For realistic large ETL jobs, the runtime-backed path should be evaluated on:

- recovery after interruption
- retry behavior
- run-ledger durability
- concurrent worker coordination
- large batch throughput
- backpressure behavior

## Re-run When Runtime Is Available

```bash
scripts/runtime.sh up
./pm runtime doctor --json
./pm perf compare --iterations 50 --runtime --json
scripts/runtime.sh down
```
