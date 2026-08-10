# Context — Issue #3974: typed database foundation recovery

**Gathered:** 2026-08-11\
**Status:** Ready for TDD replay

## Discuss-phase record

`scripts/gsd prompt discuss-phase issue-3974-typed-database-foundation-recovery --auto`
was generated after the adapter health and command-source checks passed. #3974 is not a numbered
roadmap phase, and the canonical single-worker contract forbids the isolated GSD roles expected by
the official runtime. The repository-approved manual inline fallback is therefore used. The user
provided every material decision in the recovery brief, so no decision was reopened.

## Phase boundary

Recover the complete ordered #4014 aggregate range
`0be21e1293a9e5c913d9b4e23a846237594630c7^..9e763fe07a69b48f8ed88484ccb18580c63ae160`
onto `feat/3972-postgres-parity`, including its nested hardening commits. The result is a typed,
non-executing database foundation: strict definitions, logical types, structured identities and
read plans, resource bounds, native admission, and separate warehouse-mediated source/target legs.

## Locked decisions

- Replay all twelve commits in order. Do not import, replay, or duplicate #3864.
- Reconcile only the newer-canon overlaps: `docs/architecture/connector-architecture-v2-design.md`,
  `docs/migration/conventions.md`, and `internal/connectors/native/postgres/connector.go`.
- Current #4003 documentation canon wins; retain only the database-specific F1 clauses from the
  historical documentation.
- Preserve PostgreSQL's current TLS and fail-closed CDC behavior. `write`, `query`, and `cdc`
  remain false; this slice adds no production target, direct SQL/query surface, mode execution, or
  CDC v2.
- All data movement remains warehouse-mediated. No direct connector pair, zero-copy path, generic
  SQL executor, target DDL, write session, receipt, checkpoint, polling executor, or transport
  dispatcher belongs here.
- The child PR must target `feat/3972-postgres-parity`, retain audit linkage to #4014, and use
  `Refs #3974` plus `Refs #3972`; it remains unmerged.

## Canonical references

- `AGENTS.md` — issue-first delivery, GSD/TDD, scope, safety, and gate requirements.
- `docs/connector-canon/INDEX.md` — current connector delivery canon.
- `docs/architecture/connector-architecture-v2-design.md` — current architecture wording; #4003
  canon takes precedence over historical #4014 documentation.
- `docs/migration/conventions.md` — native database authoring constraints and `database.json`
  policy boundary.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-postgres-parity-topology-scout-r1/report.md`
  — reconstruction topology, exact range, focused test set, and non-goals.
- `/Users/karthiksivadas/karthik-agent-workspace/data/cli-connector-release-certification-r1/implementation-brief-template.md`
  — delivery topology, five-loop cap, and completion-report contract.
- `.agents/agentic-delivery/canonical/delivery-contract.json` — inline single-worker lifecycle.

## Existing code insights

- `internal/connectors/native/postgres/` is production-registered as a read-only snapshot source;
  its CDC foundation is deliberately blocked before source access.
- `internal/synccontract.Mode` is the only mode vocabulary; it must be reused, not replaced.
- `internal/warehouse` owns connection-scoped Parquet materialization and owner identity; F1 may
  type the boundary but may not add another materializer or transport path.
