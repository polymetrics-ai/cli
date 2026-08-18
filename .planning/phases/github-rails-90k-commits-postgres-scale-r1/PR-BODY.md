Refs #4181
Refs #4171

## Summary

- Replaced this transport route's run-scoped approval expiry with a 24-hour-default,
  24–48-hour configurable, durable, shape-bound and revocable authorization.
- Bounded every declarative source page and PostgreSQL apply/read-back unit with a
  one-minute deadline, while reloading the durable authorization before each staged
  destination unit.
- Persisted extracted, Parquet-staged, PostgreSQL-applied counts and each phase's
  elapsed time on both completed and failed terminal runs, before test cleanup.

## Exact 90k live proof

| Measure | Result |
| --- | --- |
| Environment | Docker VM: 2 CPU / 2,054,631,424 bytes RAM |
| Rows extracted / Parquet / PostgreSQL | 90,000 / 90,000 / 90,000 |
| Independent PostgreSQL read-back | Separate connection: `SELECT count(*) = 90000` |
| Wall clock / throughput | 1,076.989768 s (17m56.989768s) / 83.566254 records/s |
| Extract / warehouse / PostgreSQL time | 539.100967784 s / 7.244399840 s / 517.652168624 s |
| Peak sampled `pm` RSS | 469,920 KiB |
| Parquet / PostgreSQL relation | 16,674,647 logical bytes / 58,859,520 bytes |
| Receipt / checkpoint | One durable receipt and final checkpoint; `batch_count=90` |
| API requests / reported headroom | 900 completed 100-record source pages; GitHub later reported `limit=5000`, `remaining=5000`, `used=0` |

GitHub extraction is the marginal bottleneck (50.06% of wall time), narrowly
ahead of PostgreSQL apply/read-back (48.06%); warehouse stage/reopen is 0.67%.
The contradictory rate headers are reported as provider headroom, not as a
trustworthy usage meter.

## Two-clock safety

- The one-time preview token remains single-use and atomically mints the durable
  authorization; it is neither logged nor persisted as a usable token.
- Authorization scope binds connection, stream/schema/table, credential
  revisions/config digests, action, confirmation policy, and expiry; it excludes
  payloads and raw credentials.
- Revocation reloads before the next staged destination unit, before its
  warehouse/PostgreSQL side effect.
- One page fetch and one apply/read-back unit each use a separate short deadline.

## Tests and delivery evidence

- Happy path: exact 90,000 fresh-binary transport and separate PostgreSQL
  read-back; latest-code one-page live binary route `100/100/100`.
- Bad path: malformed lifetime is a typed pre-I/O CLI rejection; stale/replayed
  approval and revoked authorization are refused before the next side effect.
- Edge cases: 24h/48h lifetime bounds, 900 pages, second unit timeout after a
  committed first unit, no-pre-stage zero measurement, and acknowledged
  finalization during concurrent state revisions.
- Passed: `go test -count=1 -timeout 20m ./internal/app` (222.274s),
  `./internal/cli` (415.041s), transport/engine/native Postgres packages,
  `go vet ./...`, `go build ./cmd/pm`, help transcripts, docs parity, plus the
  latest-code live binary test (24.090s).
- Manual GSD fallback, TDD ledger, verification, and review record:
  `.planning/phases/github-rails-90k-commits-postgres-scale-r1/`.

Claude auto-review is expected on PR open; this non-default-base PR records the
parent-review fallback in the review evidence.
