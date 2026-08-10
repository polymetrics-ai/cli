---

## Classification

**PostgreSQL-specific final certification and release gate.** No new core behavior should be invented
here; gaps found by certification return to their owning issue. #3987 supplies the required
warehouse-flow/mode conformance evidence before this issue begins final live proof.

## Scope

Certify the complete behavioral matrix and only then publish capability/surface truth:

- real-binary, warehouse-mediated flows for API read → warehouse → API write, API read → warehouse
  → PostgreSQL write, PostgreSQL read → warehouse → API write, and PostgreSQL read → warehouse →
  PostgreSQL write; the API→API flow proves shared routing only and does not promote PostgreSQL;
- PostgreSQL 14+ `change_capture` as a database-source path through `pgoutput` v2, bounded durable
  staging, connection-owned warehouse receipt, and acknowledgement only after that receipt;
- a complete `internal/synccontract.Mode` matrix: live proof for `full_overwrite`, `full_append`,
  `incremental_append`, `incremental_upsert`, and `incremental_dedupe`; typed non-executable proof
  for `incremental_dedupe_history`; and source-only handling for `change_capture` with no cursor
  fallback or target-mode claim;
- supported PostgreSQL version matrix through reusable `native/dbtest` Podman containers;
- positive mode/read/type/CDC coverage plus ownership, destructive safety, resource, crash/replay,
  unknown-commit, cancellation, slot/WAL and drift failures;
- PostgreSQL `certification.json`/database profile, honest `rate_limits.json`
  `state: not_applicable` with no-provider-HTTP reason, docs/help/manual/website/generated parity;
- final `write` and `cdc` capability flips only after all executable descriptors and the accepted
  live artifacts match.

Keep `query: false`. Record `incremental_dedupe_history`, cursor fallback, physical-absence deletes,
unadmitted API changefeed sinks, arbitrary existing targets and generic SQL as explicit non-support—
not hidden gaps.

## Dependencies

Hard dependencies: #3976, #3982, #3977, #3979, #3983, and **#3987**, including every transitive
shared foundation. #3987's new hard gate prevents this certification issue from certifying a direct
route or an unclassified sync mode. Cross-link old certification lane #3125 and broad PostgreSQL
umbrella #3811; this issue replaces their operation-count/broad claims with behavior-based proof.

## Acceptance and proof

- [ ] Build `pm` once and run the four named paths against live supported containers; assert exact
      source/warehouse/target records, types, transaction boundaries, receipts and checkpoints.
- [ ] Tests deliberately fail or are rejected when a source/target path skips the named warehouse
      workset, passes a counterpart connector directly, or loses plan/preview/approval state.
- [ ] Certification records all seven sync-mode outcomes: all five supported target modes, a typed
      phase-one rejection for `incremental_dedupe_history`, and PostgreSQL-only `change_capture`
      source proof through v2 staging/receipt/acknowledgement.
- [ ] Certification includes duplicate-cursor pages, explicit tombstones, owner collisions, stale
      approval, rollback, commit unknown, stage quota, slot lag/invalidation, restart, and
      schema/publication drift.
- [ ] Tests visibly skip before startup without live opt-in/explicit local Podman endpoint, but fail
      rather than skip when opted in and unavailable; all resources are uniquely owned and cleaned.
- [ ] Resource caps, timeouts, pool/batch/page/parameter bounds, CDC stage quota and slot/WAL
      telemetry are asserted. `rate_limits.json` is `not_applicable`, never `unknown`.
- [ ] `pm connectors inspect postgres --json`, catalog filters, runtime help, bare namespace/help,
      docs, website and generated artifacts all make identical claims.
- [ ] `write: true` and `cdc: true` land in the same green slice as matching executors and accepted
      live certification; `query` remains false.
- [ ] Full repository-required GSD/TDD, focused local gates, CI, automated review disposition and
      human release gate are recorded.
