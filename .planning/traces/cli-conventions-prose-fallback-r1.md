# Request-contract evidence enforcement trace

**Issue:** #3734  
**Branch:** `fm/cli-conventions-prose-fallback-r1`  
**Status:** `completed_with_warnings` — mechanism complete; repository-wide coverage gate is red
until the separately dispatched connector backfill is complete.

## Workflow provenance

This is a retrospective recovery trace, not a claim that the repository GSD artifact existed
before implementation. The task began as documentation-only and was expanded through captain
decisions while no-mistakes owned the implementation branch. The missing pre-code GSD artifact is
a workflow defect recorded here rather than hidden.

`scripts/gsd doctor` passed on 2026-08-05. The documented shell entry point was unavailable:

```text
scripts/gsd prompt programming-loop init --phase cli-conventions-prose-fallback-r1 --dry-run
scripts/gsd: unknown GSD command: programming-loop
```

The recovery therefore used the manual-GSD fallback required by
`.agents/agentic-delivery/references/gsd-pi-adapter.md`, with the no-mistakes run as the actual
review/fix/test executor. No dependencies, credentials, live writes, or quality-gate reductions
were used.

## Plan and boundaries

1. Define the five evidence tiers and hollow-command guardrails in the authoring convention.
2. Enforce operation-side source tiers, anchored field citations, explicit empty bodies, and
   retained write-action linkage.
3. Align authoring, preflight, and runtime around the same normalized schema, JSON Pointer,
   query-name, body-merge, and effective-dispatch semantics.
4. Commit measured coverage artifacts and leave strict enforcement red pending connector backfill.
5. Do not touch `internal/connectors/engine/write.go` or
   `internal/connectors/engine/schema/writes.schema.json`.

## TDD and review ledger

The no-mistakes run `01KZ73EZQMMWH8AZYNXS1RCWTJ` performed 18 source-review fix rounds. Each
accepted behavior finding was closed with regression coverage in the acting layer before the next
round. The resulting tests cover request-contract loading, required-query citations, write-field
mapping injectivity, root arrays, effective hook ownership, schema composition, JSON Pointer
assembly, static/dynamic body merge, CLI input domains, and runtime preflight.

Focused commands recorded by the pipeline include:

```text
go test ./internal/connectors/engine -run '^(TestBundleLoadParsesRequestContractAndNoBody|TestBundleLoadRequiresRequestContractForRESTOperations|TestBundleLoadRequiresRequiredQueryCitations|TestBundleLoadRejectsInvalidRequestContracts)$' -count=1
go test ./cmd/connectorgen -run '^(TestValidate_CLISurfaceNoBodyRequestContractRequiresFlags|TestValidate_CLISurfaceRequestContractFlagsApplyToAllOperations)$' -count=1
go test ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1
go run ./cmd/pm connectors inspect stripe --json
```

The last two commands are deliberately red on the strict mechanism-only rollout because checked-in
connector definitions have not yet been backfilled. Full verification is therefore honestly
recorded as incomplete, not passed. The coverage manifest records 6,927 unclaimed write actions
and 4,300 REST operations without request-contract evidence.

## Stale-intent recovery

The run retained its original documentation-only `--intent` after captain decisions authorized
schema and validation enforcement. AXI could not amend a running intent or reject the test phase's
non-gated auto-fix. Commit `7f43e5b7` therefore removed the authorized mechanism. The captain
rejected that finding; `a132076f` reverts it without dropping pipeline history. Tree equivalence
was verified with `git diff 50980590e..a132076f6`, which is empty. The only later source delta is
the delivery-pipeline finding in `docs/migration/request-contract-outbound-coverage.md`.

## Orchestration decisions

| Cycle | Decision | Reason |
| --- | --- | --- |
| Recovery planning | `local_critical_path` | One pipeline-owned branch and shared engine/schema files made a second mutating worker inappropriate. |
| Stale-intent repair | `local_critical_path` | The captain required an exact additive revert of one pipeline commit; no independent write scope existed. |
| Verification | `local_critical_path` | Evidence was already produced by the active no-mistakes run; the handoff verifies commit/tree identity and the deliberate rollout failure. |
| Summary | `local_critical_path` | One issue, one draft PR, and one status-file owner. |

## Verification state

- `git diff --check`: passed after the recovery documentation change.
- Protected-path diff check: passed; neither concurrent-lane file changed.
- no-mistakes source review: passed at `50980590` after 18 fix rounds.
- Full `make verify`: expected red until request-contract metadata is backfilled; it must not be
  marked green or weakened in this mechanism PR.
- Draft PR: #3733; human-gated and not merge-ready.
