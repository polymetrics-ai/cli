# Verification — #4181 / #4171

## Result

PASS. The fresh `pm` binary completed the authenticated `rails/rails` commits
route at exactly 90,000 records after the authorization redesign. The durable
terminal state and an independent PostgreSQL connection both report 90,000
rows. The prior run-wide 15-minute approval-expiry symptom did not recur: this
run completed in 1,076.989768 seconds.

## Live scale evidence

| Measure | Verified result |
| --- | --- |
| Environment | Docker VM: 2 CPU, 2,054,631,424 bytes RAM |
| Source configuration | `commits`, 900 declared pages, 100 records/page |
| Durable extracted count | 90,000 |
| Durable warehouse/Parquet count | 90,000 |
| Durable PostgreSQL-applied count | 90,000 |
| Independent target read-back | separate PostgreSQL `SELECT count(*)` = 90,000 |
| Wall clock | 1,076.989768 s (17m56.989768s) |
| Overall throughput | 83.566254 records/s |
| GitHub extraction time | 539.100967784 s |
| Warehouse stage/reopen time | 7.244399840 s |
| PostgreSQL apply/read-back time | 517.652168624 s |
| Peak sampled `pm` RSS | 469,920 KiB (not a continuous profiler peak) |
| Staged Parquet | 16,674,647 logical bytes, 90 batch artifacts |
| PostgreSQL business relation | 58,859,520 bytes (`pg_total_relation_size`) |
| Checkpoint / receipt | final checkpoint present; one durable delivery receipt present |
| Provider rate headers after run | `limit=5000`, `remaining=5000`, `used=0` |

The source configuration makes the transport's data-page request count exactly
900. GitHub's authenticated rate headers reported no consumed budget after that
run; that conflicts with the completed 900 data-page requests, so those headers
are recorded as provider-reported headroom, not as a trustworthy meter of
consumption.

## Bottleneck attribution

GitHub page extraction is the marginal bottleneck: 539.10 s (50.06% of wall
time), narrowly ahead of PostgreSQL apply/read-back at 517.65 s (48.06%).
Warehouse stage/reopen is only 7.24 s (0.67%). This run is therefore not
warehouse-write-bound. PostgreSQL is a close secondary bottleneck; the timing
does not support claiming VM-CPU saturation, and the sampled 459 MiB process
RSS stayed well below the 2 GiB VM capacity.

## Safety properties verified

- A single-use preview token atomically mints a 24h-default, 24h–48h bounded,
  revocable shape authorization; it does not extend the short preview token.
- The PostgreSQL scope binds connection identity, stream/schema/table shape,
  credential revisions/configuration digests, action, confirmation policy, and
  exact expiry. It excludes records, payloads, raw credentials, and tokens.
- The durable scope is reloaded before every staged destination unit. A revoked
  second unit is refused before its warehouse/PostgreSQL side effect.
- HTTP provider pages and apply/read-back units each have a one-minute default
  context deadline. The timeout test proves an acknowledged first unit survives
  a timed-out second unit.
- Terminal success and failure persist extracted/Parquet/PostgreSQL counts and
  phase durations before deferred harness/project cleanup.

## Commands and outcomes

- `go test -count=1 ./internal/app ./internal/cli ./internal/synctransport ./internal/connectors/engine ./internal/connectors/native/postgres` — passed (the long CLI package run was allowed to finish under its documented timeout).
- Focused authorization, deadline, measurement, and CLI lifetime tests — passed.
- One-page latest-code live binary test — passed: durable/independent `100/100/100`.
- Exact 900-page live binary test — persisted `90000/90000/90000`; separate read-back `90000`.
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make lint`, `./pm docs validate --connectors-dir docs/connectors`, `npm run typecheck` — passed.
- `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connector-canon-check`, `make release-workflow-check` — passed.

The long-run tool wrapper was orphaned by the execution environment while its
`pm` child continued. Terminal state and independent PostgreSQL evidence were
collected before removing the exact named container and exact temporary test
directory. Docker/Colima were not restarted or reconfigured.
