# Batch R1 vNext source-lock cutover

## Task Delivery Header

- Issue: Refs #4283 — chore(connectors): pin and declare daily-use public API cohort
- Base branch: `fm/cli-top100-declaration-batch-r1`
- Merges into: `fm/cli-batch1-vnext-legacy-cutover-r1` → `fm/cli-top100-declaration-batch-r1` → `main` through the existing PR #4294
- Delivery: push only green commits to the existing `origin/fm/cli-top100-declaration-batch-r1` head; do not open another PR and do not merge.
- Working branch: `fm/cli-batch1-vnext-legacy-cutover-r1`
- Task: replace legacy source import, projection, certification, evidence-ledger, and retention-only admission with one compact immutable vNext source-lock authoring model whose deterministic output is the existing runtime execution JSON bundle.
- Verification: focused projector/loader/preflight/commandrunner tests, definition validation, generated-output checks, broader connector/CLI tests, builds, and connector-local all-lane assertions.

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

- `scripts/gsd doctor` and `scripts/gsd sources` for discuss, plan, execute, verify, and code-review passed before planning.
- Generated GSD discuss/plan prompts are being executed inline because the task is assigned to one isolated CLI worktree and prohibits borrowing another worktree. This is the repository's manual-GSD fallback, not a lifecycle waiver.
- Loaded skills: `gsd-ns-workflow`, `golang-how-to`, `go-engineering`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`, and `no-mistakes`.
- Architecture inputs reviewed: the compact-JSON research report, connector canon, migration conventions, v2 architecture, declaration admission/evidence procedures, and Foundation Atlas. Historical source/certification rules are deletion inventory, not current architecture authority.

## Commit checkpoints

1. Tracked plan and TDD ledger.
2. Characterization/red-test contract.
3. Legacy removal plus vNext renderer green.
4. GitHub, GitLab, and Asana rendered bundles and proof green; normal push to the existing remote head.
5. Deterministic connector-local rollout commits and pushes for the remaining Batch R1 cohort.
6. Final verification, code review, generated docs/help checks, and no-legacy dependency audit.
