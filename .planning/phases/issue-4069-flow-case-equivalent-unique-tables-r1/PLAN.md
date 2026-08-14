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
correction_budget: 3/5
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
  - website/lib/docs.generated.ts
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

## Authorized recovery gap — deterministic Website aggregate

### Authority, ownership, and scope

Captain authorization at
`data/cli-github-4066-case-equivalent-unique-tables-r1/authorized-recovery-2026-08-12.md`
authorizes this existing correction-1 #4069 phase to recover custody from the
terminal duplicate-PR monitor and close the generated-data gap. This is not a
new correction loop: the ledger remains **1 / 5**, #4071 remains the existing
draft stacked PR, and no issue, branch, PR, or transport work is created.

The only production-adjacent output allowed by this gap is the deterministic
aggregate `website/lib/docs.generated.ts`. It is derived from the already
committed #4069 additions in `website/content/docs/query.mdx` and
`website/content/docs/cli-reference.mdx`; neither source page nor any runtime
or connector behavior may change. The canonical finish-plan SHA-256 remains
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`.

### C1-GD-RED — exact-head generated-data failure

At canonical #4071 head
`9a5b23fe14aba16d04f55b28ff52be0a5940cb68`, both `Website CI/CD` run
`31579519649` and `Website Data` run `31579519744` regenerated
`website/lib/docs.generated.ts`, then failed because the generated aggregate
was dirty. The live `gh-axi pr checks 4071` result reports the two duplicate
`Website checks` and two duplicate `Website generated data` failures; all four
have the same cause. The non-secret RED trace records the run IDs, steps, and
expected aggregate hash before the generator is run locally.

### C1-GD-GREEN — one owned, deterministic aggregate

Run only the repository-owned no-install command:

```text
cd website && npm run gen:website-data
```

GREEN is valid only when the website-only changed-path set is exactly
`website/lib/docs.generated.ts` and its SHA-256 is exactly
`38c8230504d3b83671c58af11dc3b69bf2eea08cf14fe7d3424dba48e3aa2106`.
Then prove a second generator invocation is idempotent, run `git diff --check`,
the affected website checks, and the documented inline/manual GSD
`verify-work` and `code-review` fallback. Stage evidence and generated output
by explicit path only.

### Delivery recovery

After the correction is committed and the worktree is clean, start exactly one
fresh no-mistakes run without `--yes` and with `--skip=pr,ci`; this preserves
#4071 topology and avoids the known default-base duplicate-PR defect. PR and
CI steps must be recorded as skipped, not represented as pipeline-owned green.
Push only through that run, then monitor natural exact-head GitHub checks on
existing draft #4071 with `gh-axi`. No GitHub workflow rerun, PR edit, retarget,
ready-for-review transition, or merge is in scope.

### Required skills and execution mode

This generator-only Website-data correction uses the loaded `no-mistakes`,
`gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`,
and `vercel-react-best-practices` guidance. No React component, page, visual,
dependency, or runtime code changes, so component-composition and visual-design
implementation guidance is not applicable. The existing single-worker
inline/manual GSD fallback remains in force because compatible isolated Pi
roles are unavailable and delegation is forbidden.

## Correction 2 / 5 — flow manual generator parity gap

### Authority and causal failure

Captain decision `[key=flow-manual-golden-drift]` authorizes correction **2 /
5** in the existing #4069 lineage. At exact #4071 head
`678e294568a8a010a460ecb05fe11a42e1eb40f2`, Verify run `31590254616`
failed `internal/cli.TestGoldenDocsGenerateMatchesTrackedCLIManuals` because
the three approved fail-closed source-selection lines appeared in
`docs/cli/flow.md` but not in the authoritative `flowHelp` string. The
canonical finish-plan digest remains
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`.

### C2-RED — reproduce the real generator ownership failure

Before editing any runtime/help source, run:

```text
go test -timeout 20m ./internal/cli -run '^TestGoldenDocsGenerateMatchesTrackedCLIManuals$' -count=1
```

RED is valid only when it exits 1 and reports `generated docs drift for
flow.md`, with the tracked manual containing exactly these missing lines:

```text
  A case-equivalent spelling whose owner cannot be decided also fails closed;
  set "connection" to a known healthy owner rather than relying on an
  unscoped query.
```

Record the non-secret result in
`traces/correction-2-flow-manual-golden-drift-red.txt`. The pre-existing
golden is the causal regression test; no unrelated test or behavior change is
needed to make this RED meaningful.

The golden iterates a Go map and stops at the first mismatch. A first local RED
run therefore reported the companion `connections.md` mismatch introduced by
the same pipeline-owned document commit: tracked text says `any local sync`
while `connectionsHelp` still says `its local sync`. A second independent run
reproduced the exact CI `flow.md` failure above. This is not a new behavior or
delivery concern: the same generated-manual source must use the already
tracked, more precise wording so the single golden can become green.

### C2-GREEN — fix source, never generated markdown

Add the exact three lines to `internal/cli/docs.go`'s `flowHelp`, immediately
after the existing omitted-owner refusal. In the same authoritative source,
change only the companion `connectionsHelp` phrase from `its local sync` to
the already tracked `any local sync`. Do not edit either checked-in markdown
manual as a source. Then invoke the repository owners in this order:

```text
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1
cd website && npm run gen:website-data
```

The CLI generation may change `docs/cli/flow.md` and the transcript generator
may change `internal/cli/testdata/golden_transcripts.json`. The website
generator is required to prove whether any existing website aggregate depends
on this documentation state; it must not cause a hand-authored website source
change. Stop if the generated diff includes unrelated source, connector,
credential, transport, or production behavior paths.

### C2 acceptance and verification

1. The exact RED selector fails before `flowHelp` changes and passes after
   regeneration.
2. `pm help flow`, bare `pm flow`, and `pm flow --json` carry the same
   fail-closed wording through regenerated golden transcripts.
3. `docs/cli/flow.md` is byte-equivalent to `pm docs generate` output; no
   manual markdown ownership is introduced.
4. Existing website source continues to describe the accepted policy; run its
   checked-in data generator and retain only output it owns.
5. Run focused CLI/manual tests, affected Go tests, docs/generator/lint gates,
   inline GSD verify-work/code-review, then commit every correction-2
   artifact. Only then start exactly one fresh no-mistakes run with
   `--skip=pr,ci`, without `--yes`.

### Correction 2 checkpoint sequence

9. `docs(4069): plan flow manual generator correction`
10. `test(4069): reproduce flow manual generator drift` (committed RED)
11. `fix(4069): generate fail-closed flow manual` (minimum GREEN)
12. `docs(4069): record flow manual correction verification`

### Correction 2 execution record

The planned checkpoint, causal RED, and minimum GREEN are committed at
`6a30fa3cf`, `03faea3c0`, and `f9b5654c8` respectively. The final evidence
checkpoint records the focused/manual, broad local, generator, lint, and
inline GSD verification before one fresh `no-mistakes` run. It does not alter
the existing draft #4071 or request a GitHub workflow rerun.

## Correction 3 / 5 — destination-scoped legacy collision admission

### Authority, decision, and scope

Independent merge verification of exact #4071 head
`3b75f4a62fd8d743ec883a5b824164374f661857` found that its last commit removed
the selected-destination condition around the legacy local-warehouse
collision preflight. That broadens the intended same-connection protection:
an unrelated non-local ETL is denied by a legacy local-warehouse collision.
This correction restores the locked C1 rule to leave non-local destinations
unchanged. It remains correction 3 / 5 in #4069 and does not create, retarget,
merge, close, or otherwise mutate #4071.

The condition must use the existing credential-free
`connectionMaterializesLocalWarehouse` materialization abstraction. Do not
inspect connector names, use a warehouse literal, add connector-specific code
outside `internal/connectors/defs/<connector>/`, or alter the existing
same-connection local-warehouse collision invariant.

### C3-RED — restore the deleted non-local execution regression

Restore `TestLegacyLocalWarehouseCollisionDoesNotBlockNonLocalETL` before an
`internal/app` production edit. It must create an unrelated non-local
destination, seed a separate legacy local-warehouse `records`/`RECORDS`
inventory, and require the non-local run to load one record, acknowledge one
batch, and make one source read. Against exact source head `3b75f4a6`, the
selector must fail with `*warehouse.SameOwnerCaseEquivalentTableError`; record
the full non-secret output in `traces/correction-3-nonlocal-collision-red.txt`.

### C3-GREEN — gate only the selected local destination

At the start of `RunETL`, retain
`validateConfiguredLocalWarehouseDestinationTables` only when the selected
connection materializes the local warehouse. The local condition must run
before `prefixedID` and `beginRun`, retaining the existing typed refusal and
zero-mutation guarantee for a true same-connection local collision. The
non-local branch must bypass this inventory preflight entirely; it does not
open a local warehouse or weaken the local invariant.

### C3 acceptance and verification

1. The restored selector is RED on production source exact head `3b75f4a6`
   and GREEN after the materialization-scoped condition is restored.
2. `TestLegacySameOwnerCaseEquivalentInventorySurvivesOpenAndFailsBeforeMutation`
   remains GREEN and its `errors.As` assertion still proves
   `*warehouse.SameOwnerCaseEquivalentTableError` before any run mutation.
3. Run the full `internal/app` package, `gofmt`, `git diff --check`, `go vet`,
   the configured lint scope, `go build ./cmd/pm`, and
   `scripts/verify-gsd-workflow`. The behavior changes no public command,
   flag, output, help, manual, or website contract, so CLI/docs/website source
   edits are not applicable.
4. Execute the resolved `verify-work` and `code-review` prompts inline under
   the existing no-delegation fallback. At delivery end, record the child PR's
   actual head SHA in `RUN-STATE.json` and the published child PR body; do not
   mutate #4071.

### C3 checkpoint sequence

13. `docs(4069): plan destination-scoped collision correction`
14. `test(4069): reproduce non-local collision bypass regression` (RED)
15. `fix(4069): scope legacy collision preflight by destination` (GREEN)
16. `docs(4069): record destination-scoped collision verification`
