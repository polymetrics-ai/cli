# Batch R1 vNext source-lock cutover

## Task Delivery Header

- Issue: Refs #4283 — chore(connectors): pin and declare daily-use public API cohort
- Base branch: `main` (API-confirmed PR #4294 base)
- Merges into: `main` through the existing PR #4294
- Delivery: push only independently green commits to the existing `origin/fm/cli-top100-declaration-batch-r1` head; do not open another PR and do not merge.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`
- Task: independently finish the vNext legacy cleanup without recovering the excluded local work, then migrate Bitbucket, CircleCI, Docker Hub, Jira, Notion, Sentry, Stripe, and Vercel one source lock at a time through the one permitted execution route.
- Verification: record truthful RED/GREEN evidence before every production slice; run focused reference/native/engine/CLI/build checks, broader engine/app/CLI checks, deterministic reference and migrated-lock renders, generated docs/skills/website checks, residual and secret/local-state scans, and the exact-SHA review procedure in `VERIFICATION.md` and `REVIEW-CONVERGENCE.md`.

## Non-negotiable architecture

- `source.lock.json` is immutable authoring and evidence input only. It is never embedded in or read by runtime.
- One pipeline owns generation: vNext source lock → canonical per-operation descriptor with shared schema registry and request/response refs → deterministic execution files (`metadata`, `spec`, streams and schemas, writes, operations, CLI surface, optional sync transport and rate limits).
- Runtime reads only execution JSON. Legacy importers, retained artifacts, source projections, certification records, root declaration ledgers, source locks, hashes, and retention-only admission are removed from runtime/admission with **no compatibility reader, fallback, feature flag, or second route**.
- Existing engine, commandrunner, REST/GraphQL/multipart encoders, credential/approval boundary, DuckDB/warehouse path, and sync transport machinery remain when they execute declared artifacts.
- Diagnostics may report authoring omissions, but only malformed/missing required execution JSON, ambiguous bindings, missing actual encoders/executors, invalid bounded routes/schemas, invalid invocation/approval/auth, and incompatible sync executor/mode may block runtime.
- Each rendered connector explicitly declares supported and empty lanes. A source operation may feed multiple lanes. Direct operations remain direct; binary, ETL, reverse-ETL, and sync transport semantics stay distinct.

## Deletion scope

1. Remove legacy `connectorgen` source import/projection/materialization/retention, declaration-admission, operation-evidence, and certification command paths and their runtime-coupled helpers/tests.
2. Remove source-lock, retained artifact, certification, API-surface evidence, and global admission-ledger embedding/loading from `defs` and `engine`.
3. Replace direct-read and deferred visibility gates that consult source/certification ledgers with execution-bundle self-consistency checks.
4. Keep source facts only in connector-local vNext locks and deterministic authoring tests; keep execution facts only in rendered JSON.
5. Update Foundation Atlas entries in the same change: retire the two legacy source foundations and record the vNext authoring renderer. No new runtime foundation is planned.

## TDD slices and test matrix

| Slice | Happy | Bad | Edge |
| --- | --- | --- | --- |
| JSON-only runtime | GitHub, GitLab, and Asana discover and reach the credential/approval boundary with provider I/O disabled. | A selected connector with malformed or missing required execution JSON is rejected. | One malformed connector does not suppress unrelated healthy connectors. |
| No legacy admission | Commands bind from operations/CLI execution JSON without source locks, importers, certification, retained evidence, or root ledgers. | Ambiguous command/operation bindings and missing encoders still fail. | A documented deferred command reports its concrete foundation gap rather than being hidden by certification or retention state. |
| Deterministic renderer | Re-rendering a vNext lock is byte-stable and equals checked-in execution JSON. | Unknown schema refs, invalid bounded routes, and contradictory lane bindings fail authoring validation. | One source operation populates multiple supported lanes without turning a direct operation into a warehouse pipeline. |
| Lane coverage | Direct read/write, binary download/upload, ETL, reverse ETL, sync transport, and explicit empty lanes are surfaced. | A sync mode without its actual executor is rejected. | Optional sync transport/rate-limit files remain absent when explicitly unsupported and deterministic when configured. |

## Rollout order

1. Commit characterization tests red-first for GitHub, GitLab, and Asana.
2. Implement the vNext canonical descriptor/renderer and remove every legacy runtime/admission dependency locally.
3. Materialize and verify GitHub, GitLab, and Asana; commit and push the green reference cohort.
4. Migrate green connector-local cohorts in order: Bitbucket, CircleCI, Docker Hub, Jira, Notion, Sentry, Stripe, Vercel. Push normally after each verified cohort.
5. Preserve source mapping if an actual shared executor is absent and report the exact gap before implementing any new genuine shared runtime foundation.

## Lifecycle record

- Restarted from the authoritative remote checkpoint `0b214b79eeb871238ce8454cd7b896e71e2746a7`, proved reachable from `origin/fm/cli-top100-declaration-batch-r1` before any edit. The excluded `/private/tmp/cli-batch1-vnext-legacy-cutover-r1` worktree is neither a source nor a recovery target.
- The prior TDD ledger's pending rows conflict with the claimed reference-cohort green checkpoint. Those claims are not carried forward as executable evidence: this continuation starts with a new native fixture-bypass RED proof and records its matching GREEN result before production cleanup.
- `scripts/gsd doctor` was run at restart and exited 1 solely because `.gsd/prompts/issue-122-rebootstrap.md` is absent. `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` passed. The generated `discuss-phase` and `plan-phase --tdd` prompts require `.planning/ROADMAP.md`, which this established custom phase does not contain. The documented inline/manual-GSD fallback therefore records discussion, plan, TDD, execution, verification, and review evidence in this phase directory; it is not a lifecycle waiver.
- Loaded guidance: repository skill routing; project and Firstmate `connector-lane-build-order`; `go-engineering` and its ETL/security guidance; `tdd`; Firstmate `firstmate-exhaustive-review`; GSD Pi adapter; and CLI help/docs/website parity. `skill://golang-how-to` and the repo-named task-specific Go skills are unavailable in this session, so the available `go-engineering` guidance is the recorded substitute rather than a false claim that those skills were loaded.
- Architecture inputs: `docs/connector-canon/SOURCE-LOCK-VNEXT.md`, current vNext loader/renderer tests, reference locks, the connector lane contract, this plan, and `VERIFICATION.md`. Historical source/certification/retention rules are deletion inventory, not current runtime authority.

## Commit checkpoints

1. Tracked plan and TDD ledger.
2. Characterization/red-test contract.
3. Legacy removal plus vNext renderer green.
4. GitHub, GitLab, and Asana rendered bundles and proof green; normal push to the existing remote head.
5. Deterministic connector-local rollout commits and pushes for the remaining Batch R1 cohort.
6. Final verification, code review, generated docs/help checks, and no-legacy dependency audit.
