# PLAN — Issue #3976: PostgreSQL dynamic typed catalog discovery

## Goal

Replace PostgreSQL's coarse/static-shaped source catalog projection with live,
ordered typed discovery of the configured database/schema's supported base
tables. The shipping native PostgreSQL runtime must consume the #4034 typed
foundation rather than leave a disconnected descriptor model.

## GSD delivery record

- `scripts/gsd doctor` passed.
- Resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and
  `code-review` through `scripts/gsd sources`.
- Generated and executed inline/manual planning prompts:
  `scripts/gsd prompt discuss-phase issue-3976-postgres-dynamic-catalog` and
  `scripts/gsd prompt plan-phase issue-3976-postgres-dynamic-catalog --tdd`.
- Generated the inline/manual `execute-phase` prompt before entering the RED
  slice: `scripts/gsd prompt execute-phase issue-3976-postgres-dynamic-catalog`.
- `go run ./cmd/agentcontractgen check` passed.
- Manual inline fallback is required because this issue is not a numbered
  roadmap phase and the canonical contract permits exactly one worker; it does
  not weaken the TDD, verification, review, or no-mistakes gates.

## Allowed paths

- `internal/connectors/native/postgres/**` for source configuration,
  connection-owned catalog discovery, typed adapter wiring, and focused tests.
- Existing typed-foundation integration points under
  `internal/connectors/database/**` only when a PostgreSQL runtime adapter
  needs an already-defined public contract; no shared semantic expansion is
  allowed in this child.
- `internal/connectors/native/dbtest/**` only for reusable test-harness wiring
  needed by a live PostgreSQL source-catalog assertion.
- `.planning/phases/issue-3976-postgres-dynamic-catalog/**` for lifecycle,
  red/green/refactor evidence, verification, and PR body.

## Explicit exclusions and ownership guard

- No destination `CREATE`/`ALTER`, managed-target mutation, write executor, or
  type/value encoder (#3982).
- No Parquet schema compiler, DuckDB workset, materialization change, or
  warehouse schema-evolution rule (#3980).
- No outbound workset delivery, approval/receipt/baseline action (#3983).
- No pgoutput session, publication/slot/LSN acknowledgement, or CDC promotion
  (#3977).
- No seven-mode executor, route dispatcher, or capability promotion (#3987).
- No generic SQL/query/write surface, dependency, credential, or secret.

## Design constraints

1. Query the live PostgreSQL catalogs using context-aware, parameterized,
   bounded read-only queries. Close every `Rows`, check each `rows.Err()`, and
   return errors with safe context only.
2. Retain database, schema, relation, column, key, native type and modifier
   identity in structured typed values. Preserve deterministic ordering:
   schemas/relation names, column ordinal, and key ordinal.
3. Expose only configured-schema base tables. Do not merge similarly named
   tables across schemas or infer a static schema from a fixture or bundle.
4. Join the discovery result to the #4034 `database` catalog/fingerprint seam
   at the native PostgreSQL runtime boundary. The old coarse `connectors.Catalog`
   compatibility projection, if still required by callers, is derived from that
   same typed catalog rather than maintaining a second static model.
5. Map only explicitly supported native PostgreSQL shapes. Preserve native
   identity/modifiers for supported types; fail closed before returning a
   partially guessed catalog for unsafe/opaque unsupported shapes.
6. Preserve read-only capability truth: do not alter PostgreSQL write/query/CDC
   claims or make a downstream executor reachable.

## TDD slices and commit checkpoints

1. **Plan checkpoint (this commit):** record ownership audit, manual GSD
   fallback, explicit red/green/refactor contract, and verification checklist.
2. **RED checkpoint:** add a regression that invokes the current runtime
   catalog adapter against two materially different real PostgreSQL catalog
   shapes. It must fail because a static/coarse adapter cannot return distinct
   typed schema/table/column/key/type metadata. Preserve exact failing output in
   `traces/dynamic-catalog-red.txt` and commit the test before production code.
3. **GREEN checkpoint:** implement a bounded typed PostgreSQL catalog adapter
   that queries live catalog metadata, maps supported native/logical types,
   preserves ordered PK membership, and supplies a deterministic fingerprint.
   Rewire the production PostgreSQL catalog path to use this adapter.
4. **GREEN hardening checkpoint:** add unsupported type/unsafe-shape, nullable,
   composite key, schema isolation, deterministic order, and cancellation/error
   tests; prove no static field/table list remains on the #3976 runtime path.
5. **REFACTOR checkpoint:** consolidate query/scan/mapping helpers while
   preserving one typed source of truth, defensive copies, context propagation,
   and error safety. Update GSD evidence after each green slice.
6. **Review/fix checkpoint:** run generated `verify-work` and `code-review`
   manual fallbacks, disposition findings, then run no-mistakes without
   `--yes`; cap fresh correction loops at five.

## Required proof

- Independent PostgreSQL catalog queries are the live-test oracle, not an
  expected `Catalog()` object constructed by the code under test.
- Two schemas in one local PostgreSQL database differ materially (relations,
  column ordinals/nullability, ordered composite keys, and native/logical type
  details) and produce different stable discovered catalogs without code/config
  schema edits.
- Discovery preserves configured database/schema/table identity and does not
  collapse tables of the same name from independent schemas.
- Unsupported native types return a named rejection, never a string fallback.
- The typed foundation is invoked by the shipping PostgreSQL runtime boundary;
  no separate disconnected static model remains for this responsibility.

## Verification plan

Focused unit and live-harness checks will run first, then race, vet, build, and
individual repository gates listed in `VERIFICATION.md`. The live test uses the
existing opt-in local `dbtest` Podman contract when a direct local Podman
endpoint is available; otherwise the skip/failure reason is recorded without
claiming live proof. This issue does not change a CLI surface, so runtime
help/manual/website parity is not applicable unless implementation evidence
shows a user-facing PostgreSQL catalog output change.

## Branch / PR topology

- Parent issue/branch/PR: #3972, `feat/3972-postgres-parity`, #4017.
- Child branch: `feat/3976-postgres-dynamic-catalog`, based on current
  `origin/feat/3972-postgres-parity` at #4064 head `c2e013324`, retained by
  non-destructive merge commit `25bda3e73`.
- Draft child PR #4065 targets exactly `feat/3972-postgres-parity` and uses
  `Refs #3976` plus `Refs #3972`. It does not target `main` and is the only
  #3976 child PR.
