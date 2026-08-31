# Context — Batch R1 retained-source mapping bridge

## Task Delivery Header

- Issue: Refs #4325 — Batch R1 source-first, complete multi-lane connector parity.
- Base branch: `origin/fm/cli-top100-declaration-batch-r1@925180aeb093b1c655ac4c1dbba52d7d81c47b07`.
- Merges into: `codex/4283-retained-source-mapping-r1` → `fm/cli-top100-declaration-batch-r1` → `main`.
- Delivery: Candidate branch committed and pushed for independent review; no parent integration or main merge.
- Working branch: `codex/4283-retained-source-mapping-r1`.
- Task: Add a generic, retained-source mapping-only command for structurally complete Batch R1 v2 source locks. It must reconstruct and verify source facts in memory, reconcile all exact source IDs against the local seven-lane matrix, and produce only a mapping report/retention-only contract value. It must not make an execution claim.
- Verification: Focused red/green/refactor Go tests, `go test -race`, `go vet`, JSON checks, cohort check, agent-contract check, help check, and changed-path diff check.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Eight frozen v2 locks are admitted only through complete retained provider evidence | live | Test reconstructs every eligible lock in memory and proves exactly 8 connectors / 2,340 REST source IDs; malformed or incomplete evidence returns an error. |
| The local mapping matrix reconciles all source IDs through exactly seven non-executable lanes | live | Test builds a retention-only contract from each matrix and `ValidateRetentionOnly` plus `ReconcileSourceOperations` succeeds; mutations show missing/duplicate/unknown IDs fail. |
| Legacy `canonical_evidence` remains a normal-import boundary | live | Test proves the default canonical predicate remains false for a valid retained v2 lock and source-import remains unmodified. |
| Mapping admission cannot become runtime admission | live | Test proves no descriptor, root enabled contract, engine bundle, source projection, materializer, or runtime operation is invoked; resulting contract has zero implemented cells and source-lock-only artifacts. |
| The command is deterministic and constrained to repository-owned inputs | live | Test checks stable JSON report/contract encoding and rejects connector/path/matrix selector defects. |

## Decisions already fixed by the assigned task

1. This is mapping/materialization-only, not runtime implementation.
2. Do not relax `sourceImportLockHasCanonicalEvidenceContract`, set `canonical_evidence`, or call `runSourceImport`, `runSourceMaterialize`, source projection, or engine bundle loading.
3. Do not write connector JSON. The candidate command is check/report-only; deterministic retention-only sidecar equivalence is tested through pure in-memory serialization, not persisted under `internal/connectors/defs`.
4. Only generic matrix variants are supported: `source_operations` with `source_id`/`lanes`, and `operations` with `source_id` or CircleCI's `source_operation_id`/`cells`.
5. Matrix schema v1 accepts only REST-only source-lock schema v2 inputs with a complete object `source_contract` and object `source_operation` fragments. GraphQL is rejected explicitly rather than fetched or guessed.
6. The pre-existing Foundation Atlas entry `source.retention-import.v1` is reused. No shared runtime contract changes, so no Atlas edit is authorized or needed.

## Cohort / ownership ledger

| Connector | Frozen source rows | Owned input paths |
| --- | ---: | --- |
| Bitbucket | 297 | `defs/bitbucket/sources/*-source-lock.json`, `*-source-lane-matrix.json` |
| CircleCI | 111 | `defs/circleci/sources/*-source-lock.json`, `*-source-lane-matrix.json` |
| Docker Hub | 54 | `defs/dockerhub/sources/*-source-lock.json`, `*-source-lane-matrix.json` |
| Jira | 617 | `defs/jira/sources/*-source-lock.json`, `*-source-lane-matrix.json` |
| Notion | 49 | `defs/notion/sources/*-source-lock.json`, `*-source-lane-matrix.json` |
| Sentry | 223 | `defs/sentry/sources/*-source-lock.json`, `*-source-lane-matrix.json` |
| Stripe | 589 | `defs/stripe/sources/*-source-lock.json`, `*-source-lane-matrix.json` |
| Vercel | 400 | `defs/vercel/sources/*-source-lock.json`, `*-source-lane-matrix.json` |
| **Total** | **2,340** | The frozen Batch R1 cohort manifest binds all file hashes and source-ID digests. |

The implementation may read those inputs but must not modify them. Changed production paths are limited to `cmd/connectorgen/retainedsourcemapping.go`, its test, and `cmd/connectorgen/main.go`; GSD evidence lives in this phase directory.

## Inline GSD fallback

The repository-local GSD adapter was healthy (`scripts/gsd doctor`), and prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were resolved. This Codex worker cannot use the Pi role runtime and is prohibited from spawning roles, so it performs the generated workflow inline and records each phase in this directory.

## Required skills loaded

- `connector-lane-build-order`
- `go-engineering`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`

## CLI parity disposition

`connectorgen retained-source-mapping` is an authoring/developer command. Its root usage and subcommand help need tests. It does not change `pm`, any public connector command, generated PM manual, website, completion, credential flow, reverse-ETL flow, or runtime JSON. Those public surfaces are not applicable.
