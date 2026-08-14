# #4069 TDD ledger

**Fresh delivery lineage:** correction 2 / 5 is active; correction 1 / 5 was
completed before this correction.
**Specification owner:** #4066 at its terminal 5 / 5; this is not loop 6
**Starting head:** `659efd8a0d69f26b55fcbd3c02150e995c159519`
**Correction-1 canonical finish-plan SHA-256:**
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`

| Slice | RED contract | GREEN contract | Status |
|---|---|---|---|
| 1. Two exact-unique canonical equivalents | Real Parquet `acme/records` and `globex/RECORDS` cause generic `SELECT 1` to fail while an omitted flow surfaces raw DuckDB duplicate-view text. | Scoped reads return their owner rows, generic `SELECT 1` succeeds, and omitted flow reads return typed `*warehouse.AmbiguousTableError` with the flow `connection` remedy. | GREEN 2026-08-12 |
| 2. Existing #4066 boundary | The correction must not alter the existing alias, three-table case-variant, exact ambiguity, action-source, reverse/read, or schedule behavior. | Targeted pre-pause fence and resumed full affected/race selectors pass after the transport CPU gate. | GREEN 2026-08-12 |
| 3. Refactor | No refactor is accepted before GREEN. | The policy remains snapshot-derived, deterministic, and free of SQL-text filtering or extra inventory scans. | GREEN review 2026-08-12 |

## RED command

```text
go test -timeout 20m ./internal/cli -run '^TestFlowCaseEquivalentUniqueTablesPreserveGenericSQLAndTypedAmbiguity$' -count=1
```

Expected RED: exit 1. The fixture must establish that both exact spellings are
individually unique, the selected reads work, then show `SELECT 1` fails at
`register view` with DuckDB's canonical duplicate error and the omitted flow
does not contain `*warehouse.AmbiguousTableError`.

## GREEN command

```text
go test -timeout 20m ./internal/cli -run '^TestFlowCaseEquivalentUniqueTablesPreserveGenericSQLAndTypedAmbiguity$' -count=1
```

Expected GREEN: exit 0. The records remain isolated by selected connection,
the unrelated generic query succeeds, and each omitted-flow assertion exposes
the typed ambiguity plus the truthful flow manifest remedy with no checkpoint
or returned rows.

## Evidence rules

The committed RED must be captured before any production implementation edit.
The only allowed local state is temporary test-owned Parquet/DuckDB data; no
credential, provider, reverse-ETL, action dispatch, or transport state is
created. Trace files contain command/result summaries only and never secrets.

## Recorded RED

At production head `f30dfcefc`, the command exited 1 as required. The selected
`acme/records` and `globex/RECORDS` subtests both passed and returned only their
own row. The unrelated generic `SELECT 1` failed before execution with:

```text
register view "records": Catalog Error: View with name "records" already exists!
```

Both omitted-flow forms then failed the typed-error assertion because the same
raw DuckDB error reached the flow engine instead of an
`*warehouse.AmbiguousTableError`. The non-secret command record is
`traces/red-case-equivalent-unique-tables.txt`.

## Recorded GREEN

The exact targeted command passed after the policy began deriving
case-equivalent exact-name groups from the existing resolver snapshot. The
fixture retained both selected owner rows, generic `SELECT 1` returned one row,
and omitted unquoted `records` plus quoted `RECORDS` each returned an error
chain containing `*warehouse.AmbiguousTableError` for `records` and the flow
`connection` remedy. The rejected paths returned no flow rows and wrote no
success checkpoint.

The short #4066 flow matrix, focused app query matrix, and schedule re-entry
selectors also passed before the CPU pause. After the gate cleared,
`internal/app`, `internal/cli`, `internal/flow`, `internal/warehouse`,
and `internal/schedule` all passed in full, and the focused
#3897/#4066/#4069 race matrix passed. The non-secret records are in
`traces/green-case-equivalent-unique-tables.txt` and
`traces/resume-broad-verification.txt`.

## Correction 1 / 5 — accepted same-owner policy

The audited d902 head leaves a same-owner destination inventory unguarded. This
is the first correction in the fresh #4069 lineage; the prior cross-owner RED
and GREEN rows remain historical evidence and must not be rewritten.

| Slice | RED contract | GREEN contract | Status |
|---|---|---|---|
| C1. New connection invariant | One local-warehouse request contains distinct streams whose effective tables are `records` and `RECORDS`; creation currently succeeds or persists. Capture state bytes/count/revision and require `errors.As` to the new type. | Creation rejects after defaults and before ID/save; all captured persisted state remains unchanged. Exact duplicate spelling and non-local destination controls preserve their current behavior. | GREEN + broad local 2026-08-12 |
| C2. Legacy sync fence | A persisted same-owner collision opens unchanged, then its next stream run currently begins/persists a run and can mutate WAL/table state. | Open does not rewrite legacy state; either stream is rejected before `beginRun` and no run/checkpoint/stream/owner/directory/WAL/temp/Parquet state changes. | GREEN + broad local 2026-08-12 |
| C3. SQL policy | Generic/selected bare and quoted collision references currently either bind a survivor or encounter raw registration behavior; a one-owner `AmbiguousTableError` is not a truthful remedy. | A dedicated typed same-owner collision is returned only for the colliding key; `SELECT 1`, unrelated tables, and a real generated-alias collision control remain executable. | GREEN + broad local 2026-08-12 |
| C4. Flow and schedule boundary | Unscoped and selected bare/quoted collision flows can complete or report the wrong error/checkpoint outcome. | Each fails without a success checkpoint, including schedule re-entry; inherited cross-owner flow/action/reverse/schedule behavior remains green. | GREEN + broad local 2026-08-12 |
| C5. Exact physical reads | A case-insensitive physical path can have one surviving spelling even though legacy state declares two. | Direct query, action, and reverse reads use only resolver-proven physical spellings and refuse the missing variant without aliasing it. | GREEN + broad local 2026-08-12 |

### Correction 1 RED command set

The committed failing command record is
`traces/correction-1-same-owner-red.txt`. It includes the five named cases in
`PLAN.md`, selected app/CLI/flow/warehouse selectors, and the preserved
cross-owner/generated-alias controls. RED evidence uses real test-owned
Parquet/DuckDB state and asserts types plus persisted mutation boundaries, not
merely command status.

## Recorded correction 1 RED

At planning checkpoint `465e02911`, with the canonical finish-plan snapshot
SHA-256 recorded above, these focused commands exited 1 as expected:

```text
go test -timeout 20m ./internal/app -run 'Test(CreateConnectionRejectsSameOwnerCaseEquivalentDestinationTables|LegacySameOwnerCaseEquivalent)' -count=1
go test -timeout 20m ./internal/cli -run '^TestFlowLegacySameOwnerCaseEquivalentInventoryStopsAtTheTypedBoundary$' -count=1
```

The app suite proves the current code accepts a new `records`/`RECORDS`
inventory, starts a legacy run, and returns nil or an ordinary missing-table
error for generic, selected, quoted, and generated-alias SQL instead of the
required typed one-owner collision. The direct resolver-backed query/action/
reverse physical-spelling control passed. The flow suite proves omitted and
selected bare/quoted forms currently complete with nil error; its two-attempt
case is the schedule re-entry regression. Full non-secret output is recorded
in `traces/correction-1-same-owner-red.txt`. No production file is part of the
RED checkpoint.

### Correction 1 GREEN rule

GREEN may begin only after the committed RED checkpoint. The implementation
must share deterministic ASCII identifier-key semantics across creation, legacy
sync preflight, and query policy; it must not migrate/rewrite data, parse SQL,
Unicode-fold, reserve flat aliases, access credentials, or claim certification.

## Recorded correction 1 targeted GREEN

The new shared ASCII identifier helper is used by local-warehouse creation,
credential-free legacy sync preflight, and the per-query DuckDB policy. The
policy receives the declared local-warehouse collision snapshot beside the
resolver snapshot; it suppresses only collided bare/generated bindings and
lets a resolver-visible real alias win. Direct query/action/reverse reads do
not enter that SQL policy and remain exact resolver lookups.

The focused target plus inherited #4066/#3897 selector fence passed. The
non-secret command record is `traces/correction-1-same-owner-green.txt`.
Broader affected-package, race, GSD review, docs/help/website, generator, and
issue-guard verification remain required before this correction can be handed
to no-mistakes.

## Authorized recovery gap — Website generated data

This is a deterministic generated-output gap in correction **1 / 5**, not
correction 2. The existing RED/GREEN behavior evidence remains intact.

| Slice | RED contract | GREEN contract | Status |
|---|---|---|---|
| C1-GD. Website aggregate | Exact #4071 head has four duplicate Website failures because the repository generator writes an uncommitted `website/lib/docs.generated.ts`; no source or runtime behavior is implicated. | The no-install repository generator changes exactly that aggregate to SHA-256 `38c8230504d3b83671c58af11dc3b69bf2eea08cf14fe7d3424dba48e3aa2106`; a second invocation is clean and affected checks pass. | GREEN 2026-08-12 |

### C1-GD recorded RED

At #4071 head `9a5b23fe14aba16d04f55b28ff52be0a5940cb68`, `gh-axi pr checks
4071` reports `29 passed, 4 failed, 12 skipped, 1 pending, 46 total`. The
four failures are two `Website checks` and two `Website generated data` jobs.
Exact `pull_request` run `31579519649` fails `Verify generated website data`;
run `31579519744` successfully runs the generator and then fails `Check
generated files`. Both report that `website/lib/docs.generated.ts` is stale.
The causal, non-secret trace is
`traces/correction-1-website-generated-data-red.txt`.

### C1-GD GREEN command and boundary

```text
cd website && npm run gen:website-data
```

Before committing, inspect `git diff --name-only -- website` and require the
sole path `website/lib/docs.generated.ts`; require its SHA-256 to equal
`38c8230504d3b83671c58af11dc3b69bf2eea08cf14fe7d3424dba48e3aa2106`.
Any other changed website path or hash is a stop condition. Do not install
dependencies, edit source documentation, or modify production behavior.

### C1-GD recorded GREEN

After the committed RED/GSD plan checkpoint `393b46bdc`, two consecutive
`cd website && npm run gen:website-data` invocations produced the exact
expected SHA-256
`38c8230504d3b83671c58af11dc3b69bf2eea08cf14fe7d3424dba48e3aa2106`.
Both times `git diff --name-only -- website` contained exactly
`website/lib/docs.generated.ts`; `git diff --check` passed. The output diff is
the audited two serialized doc-body replacements and no source document,
runtime, connector, dependency, or configuration path changed.

Affected no-install Website verification passed: ESLint exits 0 with the
known 13 warnings in untouched files, `npm run typecheck`, 80 unit tests, 27
script tests, and `npm run build` all pass. The build's existing default-auth
diagnostic remains non-fatal and does not cause a secret to be supplied.
Browser E2E is intentionally left to natural GitHub CI because the checked-in
workflow first installs Playwright Chromium and this recovery forbids installs.
The non-secret command/result record is
`traces/correction-1-website-generated-data-green.txt`.

## Recorded correction 1 broad verification

The affected `internal/app`, `internal/cli`, `internal/flow`,
`internal/warehouse`, and `internal/schedule` packages all passed with
`-count=1 -timeout 20m`. Focused race selectors passed for the correction's
creation/legacy SQL boundary and the inherited #3897/#4066 query, alias,
flow, selector, and schedule matrix. The explicit real-table control added in
`f2eacc769` materializes `RECORDS__<owner-id>` in the legacy same-owner
fixture and confirms that a resolver-visible real alias remains queryable
while the colliding base and invented aliases return
`*warehouse.SameOwnerCaseEquivalentTableError`.

`gofmt`, candidate diff checks, vet, configured lint, tidy, build/smoke,
contract/generator/surface-sync/certification/release checks, docs/manual
generation, website typecheck/lint/build, the local issue guard, and the
manual inline GSD verify-work/code-review fallback all completed. Exact
commands, inherited website warnings, and the canonical finish-plan hash are
recorded in `traces/correction-1-broad-verification.txt`. No no-mistakes,
push, PR mutation, exact-head CI, or independent Sol audit was started by
this local gate.

## Correction 2 / 5 — authoritative flow help/manual parity

**Canonical finish-plan SHA-256:**
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`

| Slice | RED contract | GREEN contract | Status |
|---|---|---|---|
| C2. Authoritative flow manual | Existing `TestGoldenDocsGenerateMatchesTrackedCLIManuals` exits 1 because generated `flow.md` from `flowHelp` omits the approved three-line case-equivalent fail-closed remedy now tracked in `docs/cli/flow.md`; map iteration can also expose the same-commit `connectionsHelp` `its`/`any` drift first. | `flowHelp` and the one-word companion `connectionsHelp` source produce the exact tracked manuals; generated CLI markdown and golden transcripts are updated by their owners, and no manual source edit remains. | GREEN targeted 2026-08-12 |
| C2. Website generated output | The existing website source/aggregate may be stale after the CLI-manual correction. | Run the checked-in website generator twice; retain only owned output and record a no-change result when CLI manuals are not an input. | GREEN targeted 2026-08-12 |

### Correction 2 RED command

```text
go test -timeout 20m ./internal/cli -run '^TestGoldenDocsGenerateMatchesTrackedCLIManuals$' -count=1
```

Expected RED: exit 1 with `generated docs drift for flow.md`; the tracked
manual contains the three case-equivalent fail-closed lines, while the
authoritative embedded `flowHelp` generation does not. The detailed
non-secret command result belongs in
`traces/correction-2-flow-manual-golden-drift-red.txt` before any `flowHelp`
edit.

The first run also recorded the same pipeline-document commit's one-word
`connections.md` source drift. Because the test stops at one randomly chosen
map entry, a second run captured the exact `flow.md` CI failure; both are in
the causal RED trace. The companion correction stays limited to the embedded
manual source and does not change runtime behavior.

### Correction 2 GREEN rule

The only hand-authored behavior text is the existing `flowHelp`/`connectionsHelp`
source. Use
`pm docs generate`, the opt-in golden transcript writer, and
`website`'s checked-in generator for all derived files. Preserve all prior
commits and runtime behavior; do not weaken the golden, manually edit
`docs/cli/flow.md`, rerun GitHub CI, or modify #4071 topology.

## Recorded correction 2 targeted GREEN

`internal/cli/docs.go` now supplies the three approved `flowHelp` lines and
the same-commit `connectionsHelp` `any local sync` wording. Two invocations of
`pm docs generate` reproduce the existing tracked CLI manuals without changing
them, proving those manuals are derived rather than hand-edited in this
correction. The opt-in transcript generator updated only the six expected
connection/flow help, bare-manual, and JSON-manual fixture values; its second
invocation was clean. Two `website` data-generator invocations were clean,
showing that its existing source/aggregate has no additional dependency on the
CLI manual files.

The focused manual/transcript/docs selectors passed, as did direct `pm help
flow`, bare `pm flow`, and `pm flow --json` inspection. The exact commands and
scoped diff are in `traces/correction-2-flow-manual-golden-drift-green.txt`.

## Recorded correction 2 broad local/GSD verification

The exact manual golden selector and the full `internal/cli` package pass with
`-timeout 20m`. `go vet ./...`, `golangci-lint run ./internal/cli`, `make lint`,
`gofmt`, and `go build ./cmd/pm` pass. The docs validator, agent-contract
check, connector validator, and canonical surface-sync check pass. Website
lint exits 0 with 13 inherited warnings in untouched files; its typecheck, 80
unit tests, 27 script tests, and production build pass without an install or a
secret. The checked-in website generator is a second clean no-change run.

`git diff --check` for the correction-2 range is clean. The exact range has no
hand-authored `docs/cli/flow.md` or website source change: it is limited to the
authoritative help source, regenerated golden transcript values, and GSD/TDD
evidence. `scripts/verify-gsd-workflow` passes against `origin/main`. The
manual inline `verify-work` and `code-review` steps pass with no candidate gap
or finding. Full non-secret command records are in
`traces/correction-2-flow-manual-broad-verification.txt`.

## Correction 3 / 5 — destination-scoped legacy collision admission

| Slice | RED contract | GREEN contract | Status |
|---|---|---|---|
| C3. Non-local ETL isolation | With a legacy local-warehouse `records`/`RECORDS` collision present, a distinct non-local connection is rejected before its source or destination executes. | The unrelated non-local ETL loads one record, sends one destination batch, and performs one source read; it never visits the local inventory guard. | GREEN 2026-08-14 |
| C3. Local invariant control | A legacy local-warehouse collision must still be refused before `beginRun` with its typed same-owner error and no persisted mutation. | The existing C2 legacy sync test remains green without a changed reason or weakened assertion. | GREEN 2026-08-14 |

### Correction 3 RED command

```text
go test -timeout 20m ./internal/app -run '^TestLegacyLocalWarehouseCollisionDoesNotBlockNonLocalETL$' -count=1 -v
```

Expected RED against production source exact head
`3b75f4a62fd8d743ec883a5b824164374f661857`: exit 1 and an error chain
containing `*warehouse.SameOwnerCaseEquivalentTableError` before the unrelated
source or destination executes. The test itself is restored from the final
commit's deletion; no production source is changed until that failure is
captured.

### Correction 3 GREEN commands

```text
go test -timeout 20m ./internal/app -run '^(TestLegacyLocalWarehouseCollisionDoesNotBlockNonLocalETL|TestLegacySameOwnerCaseEquivalentInventorySurvivesOpenAndFailsBeforeMutation)$' -count=1 -v
go test -timeout 20m ./internal/app -count=1
```

GREEN requires the non-local false positive to disappear and the true local
positive to retain the same typed reason and pre-mutation boundary. The guard
is scoped using the generic materialization interface only; no connector name
or warehouse literal is permitted.

### Recorded correction 3 RED

After restoring the deleted test, its selector ran against the unchanged
`RunETL` production source from exact #4071 head
`3b75f4a62fd8d743ec883a5b824164374f661857` (the only earlier commit in this
worktree is the planning checkpoint). It exited 1: `RunETL(non-local)` was
rejected by the legacy `acme` `RECORDS`/`records` local-warehouse collision
before the non-local source or destination could execute. The full non-secret
output is `traces/correction-3-nonlocal-collision-red.txt`.

### Recorded correction 3 GREEN

Verified production commit `a49ae952d52a6edf9786cc90e02825e2414f5b44`
restores the selected connection's generic materialization check around the
preflight; GSD evidence checkpoint `5a4f3d23645d3046d02055273fedd9a8b9dd67d1`
adds only the resulting evidence. Child-PR delivery remains pending and does
not change #4071. The focused pair passed: the restored false-positive
regression loaded one record, acknowledged one batch, and made one source read; the existing local
true-positive still passed its `errors.As` check for
`*warehouse.SameOwnerCaseEquivalentTableError` before `beginRun` or state
mutation. The full touched `internal/app` package passed in 98.497 seconds.
The non-secret command records are
`traces/correction-3-destination-scoped-collision-green.txt`.
