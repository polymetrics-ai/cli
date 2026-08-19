# #4069 — Context

**Gathered:** 2026-08-12
**Status:** Correction 1 / 5 planned; strict same-owner RED pending
**Mode:** Inline/manual GSD fallback in an issue-numbered phase

## Locked decisions

- The audited #4060 head is immutable: implementation begins only from
  `659efd8a0d69f26b55fcbd3c02150e995c159519` on the dedicated #4069 branch.
- #4066 remains the contract owner. This child starts its own 0 / 5 ledger;
  it does not reopen #4066 or consume a sixth correction there.
- The regression uses real local Parquet/DuckDB data: one owner has `records`,
  one has `RECORDS`, and each exact spelling is unique in the resolver.
- Explicit flow/app reads continue to use the manifest connection selector.
  Omitted flow reads must preserve `errors.As` to
  `*warehouse.AmbiguousTableError` so the flow engine can add its truthful
  `connection` remedy.
- Generic SQL remains deliberately available. The required control is an
  unrelated `SELECT 1`; do not remove generic aliases or parse/rewrite caller
  SQL to solve the collision.
- The solution is derived from the query's existing immutable
  `warehouse.TableResolver` snapshot. No second warehouse scan, hand-authored
  SQL filter, or provider-side behavior is allowed.

## Implementation direction

`newQueryViewPolicy` currently recognizes only an exact name for which
`resolver.Find(name, "")` is already ambiguous. The correction will represent
canonical DuckDB-name collisions from the same snapshot, suppressing duplicate
bare view registrations while retaining generic owner alias behavior. For an
unscoped flow, the canonical collision resolves to a fresh typed ambiguity
whose logical table spelling is deterministic and whose connections come from
the captured owners. Connection-scoped requests keep the zero policy and
remain isolated.

## Required skills

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-database`,
`golang-design-patterns`, and `golang-structs-interfaces` were loaded. The
later review pass additionally loads the review-required lint guidance.

## CLI/docs parity disposition

No command, flag, output schema, manifest field, help text, manual page, or
website contract changes. This is an internal fail-closed query-binding
correction. Runtime/docs/website generation remains not applicable to the
implementation diff; later verification will run the repository docs/help
checks that are affected by the final diff and record that result.

## Manual-GSD fallback

The required adapter and all lifecycle commands resolve, but this issue is not
an active numbered ROADMAP phase and this execution environment cannot run the
Pi workflow directly. The canonical single-worker contract also forbids
spawning a GSD role. The generated prompts are executed inline: decisions are
recorded here, the plan and ledger precede production edits, and RED/GREEN,
verify-work, and code-review evidence will be captured in this phase.

## Correction 1 / 5 locked decisions

- Sol's final audit at `d9022359e7b7bc2f7eb262c16177b52010681192` is a
  release-blocking acceptance finding. It is owned by #4069's existing fresh
  lineage and does not authorize another issue, branch, worktree, or PR.
- Apply the accepted policy from
  `data/cli-github-4069-same-owner-case-policy-r1/report.md`: reject distinct
  exact spellings with the same deterministic ASCII DuckDB identifier key
  within one local-warehouse connection after defaults but before save. An
  exact duplicate spelling keeps its existing behavior and is not this error.
- `App.Open` must preserve legacy state without a migration, rename, delete,
  fold, or fatal project-open rejection. `RunETL` must validate the complete
  configured local-warehouse inventory before `beginRun`, so a legacy collision
  cannot create a run, checkpoint, owner record, WAL, temporary file, or
  Parquet mutation.
- The query policy must combine the connection's declared effective inventory
  with one immutable resolver snapshot. It suppresses only a legacy
  same-owner collision's bare/generated bindings and returns a new typed error
  for bare or quoted references to that key. It must not parse or rewrite SQL,
  Unicode-fold, reserve a flat alias namespace, or make unrelated `SELECT 1`
  fail.
- A same-owner collision is not `*warehouse.AmbiguousTableError`: choosing the
  already selected connection cannot resolve it. Its error tells the operator
  to use an exact physical read where available or create replacement
  connection(s) whose destination tables differ by more than ASCII letter case.
- Connection creation and query error behavior are now user-facing. The
  original docs-N/A disposition is superseded: update connection/query manual
  and website guidance, then record runtime-help, manual, website, and
  generated-surface applicability honestly.

## Correction 2 / 5 — authoritative flow manual generator

**Command:** `scripts/gsd prompt plan-phase issue-4069-flow-case-equivalent-unique-tables-r1 --gaps`
**Execution:** inline/manual under the existing no-delegation fallback. The
captain authorized this correction after exact #4071 head
`678e294568a8a010a460ecb05fe11a42e1eb40f2` failed the generated-manual
golden gate.

| Area | Alternatives considered | Selected decision |
|---|---|---|
| Source of truth | Keep the manually edited `docs/cli/flow.md`, weaken the golden, or update embedded help | Add the three approved lines to `internal/cli/docs.go`'s `flowHelp`; the checked-in manual remains generator output only. |
| RED evidence | Treat the CI result as a transient documentation warning or reproduce the same golden assertion | Run `TestGoldenDocsGenerateMatchesTrackedCLIManuals` before the source edit and record its `flow.md` drift exactly. |
| Dependent artifacts | Hand-edit the tracked manual/transcript/website aggregate or use their owners | Run `pm docs generate`, the golden-transcript update path, and the existing website data generator; retain only their scoped outputs. |
| Delivery topology | Start another issue/branch/PR or reuse the existing child | Keep #4069/#4071 and correction 2 / 5; no PR metadata or workflow rerun is authorized. |

## Correction 3 / 5 — destination-scoped legacy collision admission

**Command:** `scripts/gsd prompt discuss-phase
issue-4069-flow-case-equivalent-unique-tables-r1`
**Execution:** inline/manual under the existing no-delegation fallback. The
independent merge report is the finished acceptance evidence for exact #4071
head `3b75f4a62fd8d743ec883a5b824164374f661857`.

| Area | Alternatives considered | Selected decision |
|---|---|---|
| Guard scope | Validate all configured local inventories for every ETL, remove the legacy guard, or scope it to the selected materialization | Preserve the guard only for an ETL whose selected connection materializes the local warehouse. |
| Architecture | Check connector names or warehouse literals in `RunETL`, or use the generic materialization abstraction | Use the existing credential-free `connectionMaterializesLocalWarehouse` method only. |
| Regression | Rely on the independent audit or restore its deleted executable control | Restore `TestLegacyLocalWarehouseCollisionDoesNotBlockNonLocalETL`, capture RED on exact source head, then retain it as committed GREEN coverage. |
| Local safety | Allow the non-local path by weakening all collision validation | Keep the existing true-positive legacy local test and typed pre-mutation refusal unchanged. |
