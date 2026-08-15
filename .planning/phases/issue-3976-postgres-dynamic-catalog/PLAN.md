# PLAN — Issue #3976: PostgreSQL dynamic typed catalog discovery

## R2 — resumable source reads (active 2026-08-16)

### Task delivery header

- Issue: `Refs #3976 — Postgres Parity: make full and cursor reads exact and resumable`.
- Base/PR target: `integration/4015-mvp-flat-r1` (dispatch head `ef3c71caf`).
- Branch: `fm/cli-3976-resumable-reads-r1`.
- One target connector: native PostgreSQL source reads only.
- Delivery: direct PR; no `no-mistakes` pipeline and no merge.

### R2 goal and ownership fence

Make the production PostgreSQL source construction path select #3858's shared resumable
polling source executor. Delete its parallel private paging loop rather than extending it.
The adapter may map PostgreSQL's typed catalog rows into the shared source seam, but must
not alter shared executor semantics, bundle schema/generation, arbitrary SQL support,
CDC, or target writes. A shared-contract gap is a required foundation split.

### TDD slices

1. **RED — production reach/happy path.** Add a test through the real `pm`/sync
   construction path proving a PostgreSQL incremental/cursor source selects the shared
   executor, emits the expected records, and persists/uses its resumable tuple. It
   fails while the private loop still owns the path.
2. **RED — typed pre-I/O refusal.** Add separately named tests for an unset stream
   cursor and stale/invalid checkpoint. Each asserts its concrete typed reason and
   zero source-session/query, checkpoint-write, and delivery calls.
3. **RED — cursor edge contract.** Add a nullable-cursor fixture and a restart fixture
   with equal cursor values. Assert no nullable row vanishes and the resumed combined
   identities are exact once-only; a live catalog mismatch is rejected specifically.
4. **GREEN — PostgreSQL adapter.** Route the native source reader through the closed
   shared source/polling seam, construct a live-catalog-validated per-stream cursor
   binding, and propagate the real checkpoint without swallowing recovery outcomes.
5. **GREEN — delete the duplicate loop.** Remove the private PostgreSQL paging/mode/
   checkpoint restrictions and retain only thin driver-specific catalog/read mapping.
6. **REFACTOR / review.** Keep the context, resource, parameterized query, defensive
   copy, and identifier-safety boundaries. Record production call-chain evidence,
   manual GSD verification/review, and focused gates.

### Test-contract matrix

| Class | Named R2 evidence | Observable assertion |
| --- | --- | --- |
| Happy path | shared-executor PostgreSQL incremental/resume | exact produced identities/rows and checkpoint-resume request, from the binary construction path |
| Bad path | missing stream cursor; stale/invalid checkpoint | specific typed refusal before any query, delivery, or checkpoint mutation |
| Edge case | nullable cursor; equal cursor/restart; two streams with different cursors | no omitted null record; exact-once combined identities; per-stream catalog binding |

The integration/container proof is deliberately pending: this host's runtime is reported
unreliable, so no shared runtime will be started or restarted. Unit/production-construction
evidence is required before delivery; a successful live-dbtest run, if an already available
explicit endpoint works, is supplemental and its exact output will be recorded.

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

## Required skills used

- `golang-how-to`, `golang-database`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-context`, and `golang-lint` for the
  native Go database boundary and its test/review gates.
- `github-issue-first-delivery` for the stacked issue/PR contract and
  `no-mistakes` for the post-commit delivery pipeline.
- `gsd-verify-work` and `gsd-code-review` for the generated-command manual
  fallbacks recorded in this phase directory.

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
base-owned opt-in `dbtest` runtime contract through Docker and Colima's direct
Unix endpoint. Its exact command and verbatim real-server output are retained
in `traces/live-reads-green.txt` and posted on issue #3976. This issue does not
change a CLI surface, so runtime help/manual/website parity is not applicable
unless implementation evidence shows a user-facing PostgreSQL catalog output
change.

## Branch / PR topology

- Child branch: `feat/3976-postgres-dynamic-catalog`.
- PR #4065 compares only to `integration/4015-mvp-flat-r1`, never `main`.
  Base head `fbd06e7d7c5c0632182e98cbb3a223ba25b19883` is retained by merge
  commit `0df3d5d4d`.
- The child uses `Refs #3976` plus `Refs #3972`; parent-stack bookkeeping is
  not a substitute for the PR's explicit integration comparison base.

## Live-proof resumption slice

The current branch is resumed on `integration/4015-mvp-flat-r1` at
`fbd06e7d7c5c0632182e98cbb3a223ba25b19883`; PR #4065 now targets that exact
integration branch. This focused GSD/TDD slice makes the existing PostgreSQL
source/catalog behavior executable through the base-owned Docker-or-Podman
`dbtest` harness without widening connector ownership.

1. **RED:** add an opt-in PostgreSQL Docker-harness assertion that requires an
   explicit runtime/Unix endpoint and observes seeded catalog, full-read, and
   cursor-read rows. Capture the pre-base harness-constructor incompatibility
   in a trace; it is historical RED evidence, not a claim that live proof is
   unavailable.
2. **GREEN:** adopt the base-owned `POLYMETRICS_CONTAINER_RUNTIME` /
   `POLYMETRICS_CONTAINER_ENDPOINT` contract, including the pinned Colima
   capacity probe, then seed deterministic rows and assert returned primary
   keys and values rather than process status.
3. **OBSERVE:** run the same live test for no `cursor_field`, a nonexistent
   cursor column, nullable cursor rows, and two relations with different
   cursor columns. Log the real outcomes without changing the captain-deferred
   connection-level cursor product contract.
4. **EXCLUSION:** do not re-enable the historical logical-replication CDC
   test. The merged capability fence intentionally returns unsupported and the
   historical test skips unconditionally; CDC execution is owned elsewhere.

The inline GSD fallback remains required (single canonical worker and no
numbered roadmap phase). The required skills for this resumption also include
`golang-concurrency` for context-bound harness cleanup. CLI/docs/website
parity is not applicable because no command, flag, help, or user-facing
connector definition changes.
