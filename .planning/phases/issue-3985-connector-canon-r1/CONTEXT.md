# Issue 3985 — connector implementation canon and delivery gate: context

- **Gathered:** 2026-08-10
- **Status:** Ready for implementation
- **Source:** GitHub issue #3985, captain rulings, and the accepted reports listed below.

<domain>
## Phase boundary

Create one repository-tracked source of truth for connector delivery, make the runtime
executability admission check visible as an explicit required gate, archive superseded material
without deleting it, and update public documentation. This is a documentation and delivery-guard
foundation; it does not implement PostgreSQL, GitHub, or any other connector.
</domain>

<decisions>
## Locked decisions

- The canon root is `docs/connector-canon/`. It contains the index, end-to-end procedure,
  archive index, and reproducibility record.
- The accepted source reports are copied from the captain's shared workspace into tracked
  `data/` paths so a clean clone does not silently depend on that workspace. Their SHA-256
  provenance is recorded in the canon index.
- `data/cli-github-etl-reverse-etl-gap-map-r1/report.md` is **VOID for coverage counts**:
  it measured `main` (341 operations), not the parity branch (768). Its architecture observations
  remain archived context only. It must never supply a coverage number.
- `data/cli-blocked-source-recovery-tiers-r1/brief.md` is archived as actively wrong for its
  unsubstantiated “195 genuinely blocked” baseline. The current `docs/migration/quarantine.json`
  contains 15 entries; the procedure and public docs will not repeat 195.
- A file named `certification.json` is not certification. The honest current baseline is zero
  accepted live certification artifacts.
- Every implementation flow is warehouse-mediated: API -> warehouse -> API, API -> warehouse ->
  database, database -> warehouse -> API, and database -> warehouse -> database. A connector may
  be both ends of the API flow; direct source-to-destination delivery is not an alternative path.
- `TestEveryImplementedCommandPassesRuntimePreflight` is the real no-network admission sweep.
  It gets a dedicated `make connector-runtime-preflight` target; both aggregate verification
  targets cover the same sweep once through `test`. It is admission/reachability evidence, not
  fixture or live correctness.
- The mandatory FOUNDATION CHECK is explicit: an implementation may declare a capability only after
  its prerequisite runtime, executor, conformance, and proof exist and execute. A missing
  foundation is a separately filed issue; it is never worked around or relabelled implemented.
- No new dependency, credentialed provider call, reverse-ETL execution, generic SQL/HTTP/shell
  write surface, or change to the PostgreSQL/GitHub parity lanes is in scope.
</decisions>

<canonical_refs>
## Canonical references

- `data/captain.md` — binding captain rulings, with historical stale count claims corrected by
  this index rather than silently edited.
- `data/cli-database-connector-framework-design-r1/report.md` — typed database framework,
  PostgreSQL reference driver, no generic SQL executor, no second sync-mode enum, no separate repo.
- `data/cli-cdc-large-transaction-strategy-r1/report.md` — PostgreSQL 14+ `pgoutput` v2,
  bounded stage, complete receipt before acknowledgement, fail-closed quota, no cursor fallback.
- `data/cli-cdc-bidirectional-changefeed-design-r1/report.md` — one workset contract with two
  different producers; explicit tombstones only in phase one.
- `data/cli-postgres-parity-issue-tree-r1/report.md` — Postgres Parity #3972 and 11 sub-issues.
- `data/cli-daily-use-top50-connectors-r1/report.md` — daily-use certification targets and the
  zero-certified correction.
- `docs/migration/conventions.md` and
  `docs/architecture/connector-architecture-v2-design.md` — bundle authoring/runtime mechanics.
- `internal/connectors/commandrunner/runner_test.go` —
  `TestEveryImplementedCommandPassesRuntimePreflight`, the real runtime admission sweep.
</canonical_refs>

<deferred>
## Deferred work

- The generated capability and flow matrices belong to #3984.
- The PostgreSQL shared foundations and eleven parity units belong to #3972 and its sub-issues.
- GitHub parity is a separate live lane; this phase only archives the wrong-branch research report.
- Accepted live proof for any connector remains a future certification job; this phase must not
  fabricate it.
</deferred>

---

*Issue: #3985*
