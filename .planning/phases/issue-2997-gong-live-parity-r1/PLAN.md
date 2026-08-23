# Plan: Gong release-0.3.0 live parity reconciliation

## Task Delivery Header

- Issue: Refs #2997 — Gong official API parity parent (historical parent; this incremental certification work remains linked for provenance).
- Base branch: `main`.
- Merges into: `main`.
- Delivery: Draft PR #3552 is refreshed on `fm/cli-gong-parity-wave01-r1`, remains unmerged, and truthfully records local and live certification evidence, including any exact unsupported or paid-operation boundary.
- Working branch: `fm/cli-gong-parity-wave01-r1`.
- Task: Reconcile the preserved Gong parity branch with current main, the published typed-destination foundation, and Batch 2/3 source-lock maps; keep every current provider operation exactly once in the declaration surface and prove the executable surfaces through the built CLI without credentials.
- Verification: provider inventory comparison; source/disposition integrity; targeted Gong definition, conformance, commandrunner, CLI, and application tests; built-binary credential-free preflight sweep; the repository's `connectorgen certification-*` gates and `pm connectors certify` external-proof harness; generator/surface checks; repository static gates; connector-boundary; and no-mistakes after the firstmate gate.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every current public OpenAPI operation is source-locked and classified exactly once | live | A credential-free public OpenAPI fetch produces 69 method/path/operation-ID rows which exactly match the committed source lock and disposition rows. |
| All direct-read declarations are executable through the installed CLI dispatch path | live | A built `pm` run in a fresh project reaches the credential preflight for every fully supplied direct-read command; none returns `unknown command` or an API-surface binding block. |
| Provider writes remain typed, declarative, and approval-bound | live | Definition validation and plan/preview/approval tests reject unbound or destructive actions before provider I/O; declarations name only fixed operations and schema-bound fields. |
| Reverse-ETL transport uses the shared definition-selected foundation | live | The certification harness must drive the persisted App plan → preview → approval → execute path and record only bounded result classifications, readback references, and cleanup outcome. |
| Live provider certification covers real persisted App paths and cleanup | live | The captain has supplied an approved non-echoing disposable credential reference. The built CLI and certification harness must prove authentication, bounded safe reads, applicable writes/readback/cleanup, ETL, pagination, required-input behavior, and supported binary paths without serializing provider payloads or credential values. |
| Every provider-defined operation is source-traced, mapped, enabled, and reachable | live | The 69-row official source lock, disposition map, API surface, operation ledger, generated CLI surface, help/manual, and website projection agree. A safety tier, scope, or destructive classification can require typed confirmation but cannot make a declared operation unreachable. |
| All six applicable execution surfaces reconcile | live | ETL, reverse ETL, direct read, direct write, binary download, and binary upload are each proven through their declaration-owned runtime path, or marked inapplicable only when the official source audit proves that Gong exposes no operation in that surface. |

## Locked discussion decisions

`discuss-phase` is executed inline under the task's autonomous instruction. The user has already
locked the relevant product and safety decisions: retain all 69 official operations; use exact,
declaration-selected commands; keep ordinary response data complete; model destructive writes with
typed confirmation and the reverse plan → preview → approval → execute flow; and do not introduce
Gong-named runtime policy branches. The captain has supplied the approved non-echoing disposable
credential reference for live certification. Agentic endpoints remain categorically excluded from
live execution because they consume paid credits; a certification requirement that cannot avoid one
is a captain decision, not a connector-local exception. The captain further requires an ordinary
provider value to remain complete even when it equals a configured credential; only explicitly
declared secret fields may be masked. Foundation #4321 owns the shared runtime correction.

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
  `golang-context`, `golang-concurrency`, `golang-documentation`, and `golang-lint`.
- Repository guidance: `required-skills-routing.md`, `gsd-pi-adapter.md`,
  `cli-help-docs-website-parity.md`, `docs/migration/HANDOFF-CODEX.md`,
  `docs/migration/conventions.md`, and `docs/architecture/connector-architecture-v2-design.md`.

## Source audit

- Official artifact: `https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=`.
- Current credential-free refetch (2026-08-23): OpenAPI `3.0.1`, title `Gong API`, version `V2`;
  59 paths and 69 operations: DELETE 3, GET 29, PATCH 1, POST 28, PUT 8. The artifact SHA-256,
  byte count, and sorted semantic inventory fingerprint exactly match the committed source lock;
  the six-surface inventory and any provider-neutral source-import dependency are recorded in
  `SOURCE-AUDIT.md`.
- The Batch 2/3 lock at
  `internal/connectors/defs/gong/sources/gong-operation-source-lock.json` has the same complete
  69-row method/path/operation-ID/deprecation inventory. The source lock is refreshed to the
  current exact artifact and normalized to the strict source-import schema; the semantic
  comparison prevents a serialization-only change from being mistaken for an operation change.

## Execution slices

1. Reconcile foundations without discarding the preserved branch. Merge `origin/main` tree
   `6410fe59c` without rewriting history; it contains the final squashed structured-body,
   source-importer, typed-header/binary/status/text, reverse-action binding, and declaration-route
   foundations. Resolve only Gong- and evidence-related conflicts, and retain `main` for shared
   runtime and unrelated connector content.
2. Reconcile Gong's `api_surface.json`, source lock/disposition map, `operations.json`,
   `cli_surface.json`, streams, write actions, schemas, fixtures, and generated connector docs
   against the 69-row source inventory. Add only declaration-owned mappings and fixtures.
3. Add RED tests for every remaining invalid runtime surface—especially missing direct-read exact
   API-surface bindings—then make the connector definitions pass the real runtime preflight and
   prove credential-free built-binary routing.
4. Reconcile Gong-owned source/destination transport and certification declarations where the
   merged foundations admit them. The three multipart actions are declaration-owned and focused
   conformance covers their generic approval-digest path; do not add a provider-specific shared
   code or fixture bypass. Run the repository certification path with the approved credential only
   through non-echoing stdin/environment delivery, and never send an agentic operation.
5. Run focused gates, repository static gates, generated artifacts, connector boundary, full
   verification where the command timeout permits, a source/diff review, and the no-mistakes
   pipeline when firstmate authorizes its final gate.
6. Regenerate the Batch 2/3 machine-readable missing-foundation ledger after each foundation
   reconciliation. The closed Gong multipart capability has zero remaining Gong rows; unrelated
   open portfolio rows remain source-traced and are not a connector-local exemption.

## Execution outcome — 2026-08-23

- Slices 1–4 completed without a provider-specific shared-code change. The source parameter
  importer added the exact missing parameters for the three multipart operations, and Gong-owned
  certification metadata selects one ordinary, bounded typed direct read rather than either paid
  agentic source row.
- The built binary reached the credential gate for all 69 declared commands with zero unknown or
  unbound outcomes. Persisted-App authentication, bounded users ETL, bounded users-extensive
  direct read, required-input rejection, cursor-pagination rejection, and scoped external proof
  passed. Full details and only non-secret fingerprints are in `SOURCE-AUDIT.md` and
  `TDD-LEDGER.md`.
- The full live harness remains a partial observation, not a certification claim: 7 of 12 ETL
  append cells passed; five remain uncertified. No Gong mutation or agentic endpoint was called.
  Full parity awaits source-projected direct requiredness, self-cleaning write pairings, the five
  ETL cells, and a captain decision for the paid endpoints.
- The verification commands are recorded in `VERIFICATION.md`. Focused Gong gates, static gates,
  connector boundary, release workflow, generated candidates/sweep/subject, and the 69-command
  dispatch sweep passed. `go test ./...` and `make verify` both stop at unrelated generated-skill
  drift; source-import, validate, and surface-sync remain blocked by the fixed query-bearing Gong
  source URL. The shared certification matrix rejects Gong because it is not allowlisted.

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
   provider response fields. Only explicitly declared secret fields may be masked, and their
   masked field must retain an explicit marker rather than being dropped. A provider value that
   merely equals configured credential material remains ordinary provider truth.
5. **Live evidence after foundations publish:** use only a non-echoing approved disposable Gong
   credential reference and the persisted App path. Exercise supported CRUD and application
   commands, each applicable surface, pagination, required-input errors, reverse-ETL
   plan/preview/approval/apply/readback, binary paths, and cleanup. Record bounded counts,
   classifications, and non-secret fingerprints—not customer payloads or credentials.

6. **Paid endpoint exclusion:** do not execute any Gong agentic endpoint. It consumes paid credits;
   if a mandatory certification cell cannot be proven by ordinary REST endpoints, record that exact
   cell as uncertified and stop for captain direction.

7. **Collision-policy foundation:** shared result projections currently mask undeclared provider
   values when they equal configured credentials. This contradicts the captain's preservation rule;
   connector work must not add an exception. Foundation issue #4321 owns the provider-neutral
   red/green change while Gong records the dependency and continues independent certification.

## CLI/docs parity

Connector declarations generate the Gong manual, skill, website data, and runtime connector help.
Any resulting command/flag/output changes require `pm help gong`, `pm gong`, relevant command
`--help`, generated-doc drift checks, and docs/website verification. No standalone generic CLI
surface may be added.
