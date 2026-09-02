# Batch R1 vNext source-lock cutover

## Task Delivery Header

- Parent issue: Refs #4325 — Batch R1 scalable synchronization execution amendment
- Current child: Refs #4423 — N1 executable proof baseline
- Base branch: `main` (API-confirmed PR #4294 base)
- Merges into: `main` through the existing PR #4294
- Delivery: one writer pushes only independently green, ordinary fast-forward commits to the existing `origin/fm/cli-top100-declaration-batch-r1` head; do not open a per-slice PR, force-push, or merge.
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

1. G0: freeze the direct-parent delivery amendment, parent/base/denominator, and local certification-tree disposition; commit and normally push the planning-only checkpoint.
2. #4423 N1: commit characterization tests red-first for GitHub, GitLab, and Asana, then restore the executable proof baseline without a runtime behavior change.
3. Implement the vNext canonical descriptor/renderer and remove every legacy runtime/admission dependency locally only through the later authorized child sequence.
4. Materialize and verify GitHub, GitLab, and Asana; commit and push the green reference cohort only when its later child gate is reached.
5. Migrate green connector-local cohorts in order: Bitbucket, CircleCI, Docker Hub, Jira, Notion, Sentry, Stripe, Vercel only after the prerequisite children are green.
6. Preserve source mapping if an actual shared executor is absent and report the exact gap before implementing any new genuine shared runtime foundation.

## G0 direct-parent delivery amendment

- Authority: #4325 comment `5500153864` (2026-09-01T20:41:45Z) requires the full issue lifecycle and exactly one code writer, with ordinary fast-forward commits directly to the existing Batch R1 parent branch. #4294 comment `5500165004` records that routing correction; the PR body was not changed after `gh-axi pr edit` failed before mutation on deprecated `projectCards`.
- Immutable delivery denominator: **4,341 primary retained source operations**, as published on #4325. N1 is a proof-baseline repair only; it neither changes that denominator nor re-pins a source lock, execution manifest, generated connector artifact, or runtime behavior.
- Immutable delivery base: after `git fetch origin fm/cli-top100-declaration-batch-r1`, both `HEAD` and `origin/fm/cli-top100-declaration-batch-r1` resolve to `d260b725ce6f53403961d7af1ef48ea6651cdd66`; its merge base with `origin/main` is `813f457a925f7ee3fe3bea101a43e445992c8552`. This continuation does not rebase or recreate any excluded local work.
- Certification-tree disposition: `HEAD` and the index contain no `internal/connectors/certifications/` path. The frozen checkpoint `0b214b79eeb871238ce8454cd7b896e71e2746a7` deleted the former tracked legacy certification tree. The sole current untracked item, `internal/connectors/certifications/.fingerprint-salt`, has no Git history and is not ignored; its opaque local provenance is not repository ownership. It remains unstaged, unread, unmodified, and out of scope for G0 and N1. No certification route is restored or retained.

## Lifecycle record

- Restarted from the authoritative remote checkpoint `0b214b79eeb871238ce8454cd7b896e71e2746a7`, proved reachable from `origin/fm/cli-top100-declaration-batch-r1` before any edit. The excluded `/private/tmp/cli-batch1-vnext-legacy-cutover-r1` worktree is neither a source nor a recovery target.
- The prior TDD ledger's pending rows conflict with the claimed reference-cohort green checkpoint. Those claims are not carried forward as executable evidence: this continuation starts with a new native fixture-bypass RED proof and records its matching GREEN result before production cleanup.
- `scripts/gsd doctor` was run at restart and exited 1 solely because `.gsd/prompts/issue-122-rebootstrap.md` is absent. `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` passed. The generated `discuss-phase` and `plan-phase --tdd` prompts require `.planning/ROADMAP.md`, which this established custom phase does not contain. The documented inline/manual-GSD fallback therefore records discussion, plan, TDD, execution, verification, and review evidence in this phase directory; it is not a lifecycle waiver.
- Loaded guidance: repository skill routing; project and Firstmate `connector-lane-build-order`; `go-engineering` and its ETL/security guidance; `tdd`; Firstmate `firstmate-exhaustive-review`; GSD Pi adapter; and CLI help/docs/website parity. `skill://golang-how-to` and the repo-named task-specific Go skills are unavailable in this session, so the available `go-engineering` guidance is the recorded substitute rather than a false claim that those skills were loaded.
- Architecture inputs: `docs/connector-canon/SOURCE-LOCK-VNEXT.md`, current vNext loader/renderer tests, reference locks, the connector lane contract, this plan, and `VERIFICATION.md`. Historical source/certification/retention rules are deletion inventory, not current runtime authority.
- Independent Firstmate review over `1655123262586b2eaa395aa75b0e54bd7c4558bd..c5bf5c5d544e85dcca5eac3ebed45ba78ad7fb33` returned **BLOCK** as an N1-to-S1A unlock. It confirms the N1 proof repairs, but it does not certify N1 or authorize the old fixture/native/registry paths. Captain applies the already-accepted D1/D2 decisions through the S1A correction map.
- S1A must add and execute RED gates before production deletion: exact-count fixture credential-boundary refusal, no-skip production registry/binary implemented-command sweep, executor-identity collision rejection, and hostile-origin/no-secret-send. Then correct the Atlas to name open whole-bundle publication and strict source-lock-decoding gaps unless the approved D2 transaction implements them, make execution manifests authoritative, remove API native overwrites and delegating hooks, and preserve execution-JSON-only defs plus existing credential, approval, and bounded-request controls.

## Commit checkpoints

1. G0 planning-only direct-parent amendment and local-state disposition; ordinary push to the existing remote head.
2. #4423 N1 characterization/red-test contract and green executable proof baseline, with no runtime behavior change.
3. Later authorized legacy removal plus vNext renderer green.
4. Later GitHub, GitLab, and Asana rendered bundles and proof green; normal push to the existing remote head.
5. Later deterministic connector-local rollout commits and pushes for the remaining Batch R1 cohort.
6. Final verification, code review, generated docs/help checks, and no-legacy dependency audit.
