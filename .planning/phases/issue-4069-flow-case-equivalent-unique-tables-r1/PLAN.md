---
issue: 4069
parent_issue: 3897
specification_owner: 4066
phase: issue-4069-flow-case-equivalent-unique-tables-r1
type: tdd
wave: 1
branch: fix/4069-flow-case-equivalent-unique-tables-r1
base_branch: feat/3897-flow-connection-scope-nm
immutable_start_head: 659efd8a0d69f26b55fcbd3c02150e995c159519
correction_budget: 0/5
files_modified:
  - .planning/phases/issue-4069-flow-case-equivalent-unique-tables-r1/
  - internal/app/query_engine_duckdb.go
  - internal/cli/flow_cli_test.go
autonomous: true
---

# #4069 TDD plan — canonical-equivalent unique table names

## GSD lifecycle record

Resolved commands:

- `scripts/gsd prompt discuss-phase issue-4069-flow-case-equivalent-unique-tables-r1 --auto`
- `scripts/gsd prompt plan-phase issue-4069-flow-case-equivalent-unique-tables-r1 --tdd`
- `scripts/gsd prompt execute-phase issue-4069-flow-case-equivalent-unique-tables-r1`
- `scripts/gsd prompt verify-work issue-4069-flow-case-equivalent-unique-tables-r1`
- `scripts/gsd prompt code-review issue-4069-flow-case-equivalent-unique-tables-r1`

The adapter is healthy and the commands resolve, but the archived ROADMAP does
not register this issue as a numbered phase and compatible Pi GSD execution is
unavailable in this worker. Under the canonical no-delegation contract, the
generated prompts are executed inline and recorded in this phase. This is a
manual execution mechanism only; it does not waive discuss, plan, strict TDD,
verification, or review.

Required skills: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-database`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-lint`.

## Scope

Change only the per-query DuckDB view policy and its regression. Preserve the
existing resolver snapshot, exact-name ownership semantics, generated aliases,
flow remedy boundary, action selectors, reverse/read selectors, schedule
re-entry, and #4066's existing collision matrices.

## Slice 1 — RED: exact-unique names still collide in DuckDB

Add `TestFlowCaseEquivalentUniqueTablesPreserveGenericSQLAndTypedAmbiguity` to
the CLI flow regression suite. Build a new real local warehouse fixture with:

- `acme/records` containing `acme-case-unique`;
- `globex/RECORDS` containing `globex-case-unique`;
- no second owner for either exact spelling.

The test must prove that both scoped app query paths return only their owner
row, then demonstrate the defect with generic `SELECT 1` and an omitted flow
against canonical-equivalent bare identifiers. The RED assertion is only valid
when `SELECT 1` returns DuckDB's duplicate-view registration error and the
flow error does not satisfy `errors.As(..., *warehouse.AmbiguousTableError)`.
Record the command output and commit the failing test before production edits.

## Slice 2 — GREEN: canonical collision policy from one resolver snapshot

Extend `newQueryViewPolicy` and view registration without parsing or rewriting
caller SQL:

1. derive canonical DuckDB identifier groups from `resolver.Tables()`;
2. identify groups with different exact table spellings that collide in
   DuckDB, while retaining each original spelling and owner identity;
3. suppress conflicting unscoped bare view registration, allowing generic
   owner aliases where the existing policy permits them;
4. for an unscoped flow, map a canonical bare lookup to a new typed
   `warehouse.AmbiguousTableError` made from the same captured owners;
5. leave connection-scoped queries on the zero policy so only their selected
   table can bind.

Rerun the exact RED selector. GREEN requires the generic `SELECT 1` control,
both selected rows, and omitted-flow typed ambiguity/remedy to pass. No
provider, credential, transport, or production wiring path may be touched.

## Slice 3 — REFACTOR while green

Keep canonical-key bookkeeping concrete and private to the existing query
policy. Preserve deterministic logical-table error spelling, defensive copies
of connection lists, and the existing `errors.As` chain. Avoid a new package,
interface, generic SQL capability, or warehouse inventory scan.

## Targeted verification before CPU pause

- `go test -timeout 20m ./internal/cli -run '^TestFlowCaseEquivalentUniqueTablesPreserveGenericSQLAndTypedAmbiguity$' -count=1`
- the short affected #4066 selector as a regression fence, without broad
  package or race execution while the transport lane holds the heavy CPU slot.
- `gofmt` for touched Go files and `git diff --check` for the #4069 range.

Commit the green correction with `Refs #4069`, update this ledger and
RUN-STATE, then append the required paused status for Firstmate. Do not start
the full app/CLI suites, broad race sweeps, repository-wide verification, or
no-mistakes until the transport CPU gate clears.

## Resume verification

After Firstmate resumes this worker, run proportionate affected-package and
focused-race tests, `git diff --check` for the child range, GSD verify/review,
vet/lint, canonical generator/surface-sync/certification checks, applicable
docs/help/website checks, and the PR issue guard. Distinguish the known
inherited seven-line #3897 whitespace defect from this child range; do not
repair it here. Then use `no-mistakes axi run --help` and drive the allowed
pipeline without `--yes`.

## CLI/docs/website parity

Not applicable to production edits: no public command, flag, output, help,
manifest syntax, manual, or website documentation changes. The resume
verification records affected docs/help/website gate results and this explicit
exemption in the PR body.

## Commit checkpoints

1. `docs(4069): plan canonical-equivalent warehouse correction`
2. `test(4069): reproduce case-equivalent unique warehouse table failure`
3. `fix(4069): preserve query availability for case-equivalent tables`
4. `docs(4069): record broad correction verification`
