---

## Outcome

Deliver complete, truthful PostgreSQL **database-connector parity** through the warehouse. Parity is
a certified behavioral contract, not an official-endpoint or SQL-statement count. A connector
implements its own source or target side against a persisted warehouse table/workset; it never
receives or dispatches directly to another connector.

The tree owns the PostgreSQL side of these four canonical paths:

1. API read → warehouse → API write — required as shared flow/mode conformance, without claiming a
   PostgreSQL capability where no PostgreSQL leg is used.
2. API read → warehouse → PostgreSQL write.
3. PostgreSQL read → warehouse → API write.
4. PostgreSQL read → warehouse → PostgreSQL write.

PostgreSQL change capture is a database-source extension of paths 3 and 4: PostgreSQL 14+
`pgoutput` v2 produces a bounded, transaction-preserving warehouse workset before any typed target
can consume it.

PostgreSQL is currently a read-only polling source. Its measured capabilities remain
`read: true, write: false, cdc: false`; `Postgres.Write` is unsupported; no database write executor
exists; and the CDC branch deliberately fails closed until bounded transaction staging exists.

## Binding decisions

- Implement the accepted typed database framework in `internal/connectors/database/` with
  PostgreSQL as reference driver. No generic SQL executor, second sync-mode enum, separate repo, or
  direct connector-to-connector route.
- Full and cursor-incremental reads must be exact, resumable, bounded, and materialized into the
  connection-owned warehouse before a target can see them.
- Change capture is PostgreSQL 14+ streamed `pgoutput` v2 with a bounded crash-recoverable stage,
  abort discard, whole-transaction warehouse receipt before LSN acknowledgement, fail-closed quota,
  and no cursor fallback.
- Target tables are created and owned by Polymetrics, structurally scoped by workspace + source
  connector + source connection ID, with an in-database ownership assertion. Refuse missing,
  unreadable, foreign, or drifted ownership; do not adopt arbitrary existing tables.
- Use only `internal/synccontract/mode.go`. Every one of its seven modes must be classified and
  proven truthfully: `full_overwrite`, `full_append`, `incremental_append`, `incremental_upsert`,
  and `incremental_dedupe` are phase-one PostgreSQL target modes; `incremental_dedupe_history` is a
  typed non-executable result until an immutable history contract exists; `change_capture` is a
  PostgreSQL source mode only, not a target write mode. A source or target must not borrow a mode
  from its counterpart without independently admitted evidence.
- Outbound continuous delivery derives keyed Parquet/DuckDB deltas from the warehouse and accepts
  explicit tombstones only. Physical absence is not a delete in phase one.
- Capability flags turn true only after real-binary and live-container certification. A generated
  surface, fixture, or issue checklist is not proof.

## Children and dependency order

| Key | Child | Class | Depends on |
| --- | --- | --- |
| F1 | #3974 | shared typed-framework foundation | — |
| F2 | #3981 | shared managed-target foundation | F1 |
| F3 | #3973 | shared write-session foundation | F1, F2, #3859 |
| F4 | #3975 | shared committed-transaction/warehouse receipt foundation | F1 |
| F5 | #3980 | shared immutable warehouse workset foundation | F1, F2 |
| P1 | #3976 | PostgreSQL warehouse-read adapter | F1, #3858 |
| P2 | #3982 | PostgreSQL managed-target write driver | F1, F2, F3, #3859 |
| P3 | #3977 | PostgreSQL CDC-to-warehouse producer | F1, F4, P1; incorporate PR #3967 |
| P4 | #3979 | PostgreSQL gap-free warehouse bootstrap | F4, F5, P1, P3 |
| P5 | #3983 | PostgreSQL warehouse-workset target delivery | F5, P2 |
| P6 | #3987 | four-flow/seven-mode warehouse conformance | P1–P5, #3864 |
| C1 | #3978 | PostgreSQL final certification and promotion | P1–P6 |

## Dependency-graph correction

#3987 is a new hard gate before #3978. The prior graph allowed certification after P1–P5 without an
owner for all four warehouse routes or the seven-mode truth table. The added gate does **not** change
or block active Wave A #3974: it runs only after the existing read, write, CDC, bootstrap, and
workset slices are ready. It moves final certification one integration wave later because
certification must consume executable flow/mode evidence rather than invent it.

Execution waves are therefore: A F1; B F2/F4 and P1 after #3858; C F3/F5/P3 once their individual
dependencies land; D P2; E P4/P5; F P6; G C1. Interface/conformance scaffolding for P6 may be
reviewed earlier, but its acceptance cannot close before P1–P5.

## Parent acceptance

- [ ] Every child is closed with its own issue-first GSD/TDD evidence and reviewed production PR.
- [ ] A real built `pm` binary proves the four named warehouse-mediated paths. Tests demonstrate
      that neither source nor target can form a direct source-to-destination hop.
- [ ] The flow/mode evidence contains all seven `synccontract.Mode` outcomes: five supported target
      modes, typed `incremental_dedupe_history` non-support, and PostgreSQL-only `change_capture`
      with no cursor fallback or target-mode claim.
- [ ] PostgreSQL CDC proves transaction boundaries, warehouse receipts before acknowledgement, slot
      state, quota failure, restart, and rebootstrap behavior against live PostgreSQL 14+ containers.
- [ ] Tests assert returned rows, target rows, transaction boundaries, receipts/checkpoints, owner
      identity, and failure behavior—not only exit status.
- [ ] Destructive, replay, quota, cancellation, schema-drift, ownership-collision, commit-unknown,
      disconnect, and restart cases are proven.
- [ ] `rate_limits.json` honestly says provider HTTP limits are not applicable; native resource caps
      and CDC slot/WAL telemetry are enforced and documented.
- [ ] Help/manual/docs/website/generated artifacts agree.
- [ ] `write` and `cdc` become true only in C1 after all evidence is green; `query` stays false.

## Prior work to reference, not duplicate

Reconcile #3118–#3125, #3811, polling/apply tree #3855–#3860, transport tree #3862/#3864, shared
changefeed/state foundations #2986/#2988 and children #3745–#3749, and PR #3967. The earlier
operation-count tree and unmerged writer predate the accepted managed-target and write-session design
and do not themselves establish parity. These children decompose #3811; they do not make a second
shared polling, apply, or direct transport executor.
