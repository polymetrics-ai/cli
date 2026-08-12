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
correction_budget: 1/5
canonical_finish_plan_sha256: 939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408
files_modified:
  - .planning/phases/issue-4069-flow-case-equivalent-unique-tables-r1/
  - internal/app/app.go
  - internal/app/query_engine_duckdb.go
  - internal/app/warehouse_destination_collision.go
  - internal/app/types.go
  - internal/warehouse/layout.go
  - internal/cli/docs.go
  - internal/app/warehouse_connection_isolation_test.go
  - internal/cli/flow_cli_test.go
  - docs/cli/connections.md
  - docs/cli/query.md
  - website/content/docs/query.mdx
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
`golang-structs-interfaces`, `golang-lint`, and `golang-documentation`.

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

## Correction 1 / 5 — same-owner case-equivalent warehouse inventory

### Authority and scope

The independent Sol audit of PR #4071 at
`d9022359e7b7bc2f7eb262c16177b52010681192` found that the prior GREEN only
covers cross-owner canonical equivalence. This correction implements accepted
policy 1 from
`data/cli-github-4069-same-owner-case-policy-r1/report.md`, under the existing
#4069 issue and draft #4071. It preserves every prior commit and starts no new
issue, branch, worktree, PR, or #4066 correction loop.

The active no-mistakes CI monitor was inspected before planning. It reports the
local branch synchronized with its pipeline head and its version-matched help
explicitly permits post-pipeline follow-up commits. The monitor remains active
and is never aborted, restarted, updated, or answered by hand.

### Slice C1-RED — commit the missing same-owner matrix before production code

Add and commit failing tests before editing `internal/app` production code.
Each typed assertion uses `errors.As`; each mutability assertion records actual
state/file outcomes rather than relying only on exit status.

1. `TestCreateConnectionRejectsSameOwnerASCIICaseEquivalentDestinationTablesBeforeSave`
   creates a two-stream local-warehouse request whose effective destinations
   are `records` and `RECORDS`. It expects a typed same-owner collision after
   defaults and before ID/save, with connection count, persisted bytes, and
   revision unchanged. Exact duplicate spelling and non-local destinations
   remain controls, not accidental new policy.
2. `TestLegacySameOwnerCaseEquivalentStateLoadsWithoutRewriteAndSyncRefusesBeforeMutation`
   seeds one persisted owner with distinct stream names and the two destination
   spellings. `Open` must preserve it. Either stream's next ETL must return the
   typed error before `beginRun`, run/stream/checkpoint state, owner record,
   directory, WAL, temporary file, or Parquet file mutation.
3. `TestLegacySameOwnerCaseEquivalentSQLUsesDeclaredInventory`
   covers generic and selected SQL, bare and quoted forms. It expects the new
   type rather than DuckDB `Catalog Error` or a one-owner
   `AmbiguousTableError`, while `SELECT 1` remains green and a real
   `RECORDS__<owner-id>` table keeps priority over a generated alias.
4. `TestSameOwnerCaseEquivalentFlowFailsWithoutSuccessCheckpoint` covers
   unscoped and selected bare/quoted flow queries, returned failure/result,
   absent success checkpoints, and schedule re-entry through the same flow
   boundary. Existing cross-owner action/reverse/schedule selectors remain
   regression controls.
5. `TestExactTableReadDoesNotAliasMissingCaseVariantOnCaseInsensitiveFilesystem`
   proves exact `QueryTable`, action, and reverse reads retain only physical
   spellings the resolver can prove. It materializes one spelling on the host
   filesystem and refuses the missing case variant without silently mapping it
   to the survivor.

Commit this RED slice as `test(4069): reproduce same-owner table collision`
with the `TDD-LEDGER.md` RED command/trace before the implementation commit.

### Slice C1-GREEN — one shared invariant and typed query boundary

Implement the smallest shape that makes the RED matrix green:

1. Keep one shared deterministic ASCII DuckDB identifier-key helper, avoiding
   Unicode folding or a new alias namespace.
2. Add one typed same-owner destination-table collision error containing the
   display connection and sorted exact spellings/stream names but no path,
   credential, or raw connector data.
3. Validate a whole local-warehouse connection's effective stream inventory
   after defaults in `CreateConnection`, before generating an ID or saving.
   Validate the same configured inventory at the beginning of `RunETL`, before
   `beginRun`. Leave non-local destinations unchanged.
4. Preserve legacy `Open`; derive its query policy from declared destinations
   plus the immutable resolver snapshot. Suppress only the collision's bare and
   generated bindings, return the typed error through replacement-scan lookup
   for bare/quoted collision references, and let unrelated SQL initialize.
5. Preserve exact direct/action/reverse lookup and all inherited #4066/#3897
   cross-owner, `_unattributed`, generated-alias, selector, and schedule
   behavior. Do not parse/rewrite SQL, migrate data, reserve aliases, add
   provider work, transport registration, production wiring, or certification
   claims.

Commit the minimum GREEN implementation as
`fix(4069): guard same-owner case-equivalent tables`. Refactor only with all
focused selectors green.

### Slice C1-docs and verification

Document the rejected creation configuration and legacy collision read/recovery
path through the embedded source `internal/cli/docs.go`, generated
`docs/cli/connections.md` and `docs/cli/query.md`, the corresponding golden
transcripts, and `website/content/docs/cli-reference.mdx` plus
`website/content/docs/query.mdx`. No new command or flag is introduced, so the
runtime-help wording/golden changes are applicable only because the generated
help already documents this behavior; run and record `pm connections`,
`pm connections create --help`, `pm query`, `pm query run --help`, manual and
website checks rather than claiming a blanket exemption.

After GREEN, run targeted tests first, then the affected app/CLI/flow/warehouse
packages, focused race selectors, `gofmt`, candidate `git diff --check`, vet,
configured lint, generator/surface-sync/certification checks, docs/help/website
checks, the PR issue guard, inline `verify-work`, and inline `code-review`.
Record inherited findings separately. Drive no-mistakes without `--yes` only
after this correction is committed and local gates are green; preserve #4071 as
a draft against `feat/3897-flow-connection-scope-nm` and require exact-head CI.

### Correction 1 checkpoint sequence

5. `docs(4069): plan same-owner case-equivalent correction`
6. `test(4069): reproduce same-owner table collision` (committed RED)
7. `fix(4069): guard same-owner warehouse collisions` (minimum GREEN)
8. `docs(4069): record same-owner correction verification`

The prior four checkpoints remain immutable evidence for the cross-owner
correction. This is correction 1 / 5 in #4069, leaving four correction loops.

### Correction 1 execution record

- `465e02911` — correction-1 plan, ledger, run-state, and verification update
  before production edits.
- `6e0a4a7cb` — committed strict same-owner RED with real local
  Parquet/DuckDB state.
- `06414fa44` — minimum production GREEN.
- `b182d559e` — concrete flow typed-error assertion.
- `f2eacc769` — causal real-alias control proving a resolver-visible physical
  `RECORDS__<owner-id>` table wins over an invented alias.
- `bb7c2458c` — embedded/generated/website documentation and golden parity.
- The next commit records broad local/GSD verification only. It does not start
  no-mistakes, push, mutate draft PR #4071, or claim exact-head CI/Sol audit.

### CLI/docs/website parity supersession

The earlier no-public-surface exemption no longer applies. The change modifies
creation and SQL error contracts, so connection/query manual and website docs
are in scope. No connector credential, reverse-ETL execution, or new generic
write surface is permitted.
