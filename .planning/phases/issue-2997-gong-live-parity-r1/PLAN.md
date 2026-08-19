# Plan: Gong release-0.3.0 live parity reconciliation

## Task Delivery Header

- Issue: Refs #2997 — Gong official API parity parent (historical parent; this incremental certification work remains linked for provenance).
- Base branch: `main`.
- Merges into: `main`.
- Delivery: Draft PR #3552 is refreshed on `fm/cli-gong-parity-wave01-r1`, remains unmerged, and truthfully records local gates plus any external live-certification blocker.
- Working branch: `fm/cli-gong-parity-wave01-r1`.
- Task: Reconcile the preserved Gong parity branch with current main, the published typed-destination foundation, and Batch 2/3 source-lock maps; keep every current provider operation exactly once in the declaration surface and prove the executable surfaces through the built CLI without credentials.
- Verification: provider inventory comparison; source/disposition integrity; targeted Gong definition, conformance, commandrunner, CLI, and application tests; built-binary credential-free preflight sweep; generator/surface checks; repository static gates; connector-boundary; and no-mistakes after the firstmate gate.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every current public OpenAPI operation is source-locked and classified exactly once | live | A credential-free public OpenAPI fetch produces 69 method/path/operation-ID rows which exactly match the committed source lock and disposition rows. |
| All direct-read declarations are executable through the installed CLI dispatch path | live | A built `pm` run in a fresh project reaches the credential preflight for every fully supplied direct-read command; none returns `unknown command` or an API-surface binding block. |
| Provider writes remain typed, declarative, and approval-bound | live | Definition validation and plan/preview/approval tests reject unbound or destructive actions before provider I/O; declarations name only fixed operations and schema-bound fields. |
| Reverse-ETL transport uses the shared definition-selected foundation | fake | Actual Gong reverse-ETL requires an approved disposable credential reference. Connector-local destination declarations and credential-free admission tests prove the non-provider path; live apply/readback is blocked until that reference is supplied. |
| Live provider certification covers real persisted App paths and cleanup | fake | No approved disposable Gong credential/store reference is available in this worktree. No secret will be requested, printed, persisted, or put in command arguments. |
| Every provider-defined operation is source-traced, mapped, enabled, and reachable | live | The 69-row official source lock, disposition map, API surface, operation ledger, generated CLI surface, help/manual, and website projection agree. A safety tier, scope, or destructive classification can require typed confirmation but cannot make a declared operation unreachable. |
| All six applicable execution surfaces reconcile | live | ETL, reverse ETL, direct read, direct write, binary download, and binary upload are each proven through their declaration-owned runtime path, or marked inapplicable only when the official source audit proves that Gong exposes no operation in that surface. |

## Locked discussion decisions

`discuss-phase` is executed inline under the task's autonomous instruction. The user has already
locked the relevant product and safety decisions: retain all 69 official operations; use exact,
declaration-selected commands; keep ordinary response data complete; model destructive writes with
typed confirmation and the reverse plan → preview → approval → execute flow; and do not introduce
Gong-named runtime policy branches. The only external dependency is a non-echoing disposable Gong
credential reference for live certification. It is not safe to invent or request one in this task.

## GSD command path and manual fallback

- `scripts/gsd doctor` passed.
- `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and
  `code-review` resolved to the checked-in command registry and official documentation.
- Manual-GSD fallback: the canonical single-worker contract and this runtime prohibit spawning the
  required isolated GSD roles. The lifecycle is therefore recorded and executed inline: this plan
  is `discuss-phase` and `plan-phase --tdd`; the red/green ledger records `execute-phase`; the
  verification checklist records `verify-work`; and `REVIEW.md` will record `code-review`.

## Required skills loaded

- `no-mistakes`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-context`, `golang-concurrency`, and `golang-documentation`.
- Repository guidance: `required-skills-routing.md`, `gsd-pi-adapter.md`,
  `cli-help-docs-website-parity.md`, `docs/migration/HANDOFF-CODEX.md`,
  `docs/migration/conventions.md`, and `docs/architecture/connector-architecture-v2-design.md`.

## Source audit

- Official artifact: `https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=`.
- Current fetch: OpenAPI `3.0.1`, title `Gong API`, version `V2`; 59 paths and 69 operations:
  DELETE 3, GET 29, PATCH 1, POST 28, PUT 8.
- The Batch 2/3 lock at
  `internal/connectors/defs/gong/sources/gong-operation-source-lock.json` has the same complete
  69-row method/path/operation-ID/deprecation inventory. Gong changes JSON serialization between
  requests while retaining the same bytes/count and semantic inventory, so raw digest drift alone
  is not a provider-contract change. The lock remains the captured immutable source artifact;
  this phase records semantic refetch proof rather than replacing it with a nondeterministic hash.

## Execution slices

1. Reconcile foundations without discarding the preserved branch: merge current `origin/main`,
   then the published `origin/fm/cli-reverse-etl-destination-r1`, then
   `origin/fm/cli-map-batch23-r1`; resolve only Gong- and evidence-related conflicts. The named
   structured-body/source-importer/header-binary foundations are not published in this repository
   at planning time (`#4305`, `#4306`, and `#4307` return GitHub `NOT_FOUND`) and are recorded as
   unavailable rather than copied into shared runtime code.
2. Reconcile Gong's `api_surface.json`, source lock/disposition map, `operations.json`,
   `cli_surface.json`, streams, write actions, schemas, fixtures, and generated connector docs
   against the 69-row source inventory. Add only declaration-owned mappings and fixtures.
3. Add RED tests for every remaining invalid runtime surface—especially missing direct-read exact
   API-surface bindings—then make the connector definitions pass the real runtime preflight and
   prove credential-free built-binary routing.
4. Add Gong-owned source/destination transport and certification declarations where the merged
   foundations admit them. Multipart fixture replay is paused on the accepted provider-neutral
   `cli-closed-operation-runtime-r1` F2/F4 approval-digest requirement; do not add a
   provider-specific shared-code or fixture bypass. Run plan/preview/approval/execute only through
   credential-free refusal paths until a disposable credential reference is supplied.
5. Run focused gates, repository static gates, generated artifacts, connector boundary, full
   verification where the command timeout permits, a source/diff review, and the no-mistakes
   pipeline when firstmate authorizes its final gate.

## Captain hard certification gate

The branch is not merge-ready until the following evidence is complete. This gate supersedes any
historical parity count or declaration-only validation result.

1. **Source and mapping:** every current official provider operation is traceable to one immutable
   source-lock row, one disposition, and the exact supported declaration/API/CLI surface. No source
   row is hidden by tier, scope, safety, or inconvenience.
2. **Enabled behavior:** every declared supported operation reaches its real command/App dispatch
   path. Destructive or high-risk operations use runtime metadata, typed confirmation, and the
   existing plan → preview → explicit approval → execute path—not `availability` or a connector
   policy branch as a substitute for reachability.
3. **Six-surface reconciliation:** separately certify ETL, reverse ETL, direct read, direct
   write, binary download, and binary upload. A surface is `not_applicable` only with an exact
   source-audit finding; otherwise it remains an unmet requirement. Binary upload includes
   multipart after `cli-closed-operation-runtime-r1` publishes its approval-digest contract.
4. **Discovery and output:** prove every supported operation is reachable from runtime CLI help,
   generated manual/docs, and website discovery projections. Retain all ordinary non-secret
   provider response fields; credential/transport-secret masking must preserve the masked field
   with an explicit marker rather than dropping it.
5. **Live evidence after foundations publish:** use only a non-echoing approved disposable Gong
   credential reference and the persisted App path. Exercise supported CRUD and application
   commands, each applicable surface, pagination, required-input errors, reverse-ETL
   plan/preview/approval/apply/readback, binary paths, and cleanup. Record bounded counts,
   classifications, and non-secret fingerprints—not customer payloads or credentials.

## CLI/docs parity

Connector declarations generate the Gong manual, skill, website data, and runtime connector help.
Any resulting command/flag/output changes require `pm help gong`, `pm gong`, relevant command
`--help`, generated-doc drift checks, and docs/website verification. No standalone generic CLI
surface may be added.
