Closes #4352

## Intent

Provide a shared, closed source-bound read execution foundation. A mapped source GET can become a normal credential-gated fixed operation only when its exact source identity, method, path, typed inputs, and executor semantics are declared.

## What Changed

- Added `source_operation` bindings (source ID + GET method + connector-relative path) to declarative operations and command metadata.
- Extended source projection to consider non-mutating GETs without changing mutation/write projection. It promotes only existing, exact declarations; it does not create a route or generic HTTP surface.
- Added no-network engine preflight for source-bound direct reads and stream ETL commands. Stream proof requires one declared composite, the exact source route, records/schema, and pagination.
- Added a stable `missing_foundation=source-bound-read-execution-r1:` disposition for incomplete typed read contracts before provider dispatch.
- Added a source-cited partial-coverage disposition for an already implemented
  mutation whose locked request contract still needs a named shared foundation.
  It requires exact source ID/method/path and a matching incomplete action; it
  cannot downgrade, invent, or conceal an operation.
- Reconciled #4357's source-cited write-disabled mutation artifacts. They need
  explicit `metadata.capabilities.write=false` and retained provider citation;
  complete declared delete/reverse-ETL actions retain their executable lane.
- Captain correction `021.msg` supersedes the historical 9/100 partition:
  source import materializes all source-complete retained Asana GET contracts
  in their true executor lane—106 bounded direct reads and 12 exact
  record/schema/pagination-backed ETL streams. Only
  `asana.rest.getMembership` remains deferred, explicitly naming
  `cli-openapi30-reference-sibling-foundation-r1`.
- The ETL fan-out proof uses exact locked identities and locations:
  - `asana.rest.getProjectStatusesForProject` — `paths["/projects/{project_gid}/project_statuses"].get`
  - `asana.rest.getSectionsForProject` — `paths["/projects/{project_gid}/sections"].get`
  - `asana.rest.getStoriesForTask` — `paths["/tasks/{task_gid}/stories"].get`

The source locations above are from `data/connector-operation-mapping-reports/100-connectors/batch-1/asana.json`.

## Asana surface

The locked source has 249 operations. The current executable count is **212**:

| Intent | Implemented |
| --- | ---: |
| `direct_read` | 106 |
| `etl` | 12 |
| `reverse_etl` | 94 |
| **Total** | **212** |

**37** operations remain unavailable only with their exact provider citation
and concrete `missing_foundation` (or the explicitly unsupported `/batch`
wrapper). The historical claim that the 21 source-complete mutations were
non-executable has been removed.

## TDD / GSD Evidence

- Lifecycle: inline/manual `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`, recorded in `.planning/phases/issue-4352-source-bound-read-execution-r1/`. Inline fallback was required because compatible isolated workers were unavailable and the canonical contract forbids role spawning.
- Red: source projection test failed before implementation because `get_access_requests` had no source binding.
- Green: source projection, route-substitution/no-dispatch, missing-foundation/no-dispatch, direct-vs-stream, and Asana preflight controls pass. Full evidence is in `TDD-LEDGER.md` and `VERIFICATION.md`.
- Frozen findings AUDIT-001 through AUDIT-006 were reread and carried forward.
  The R5 independent pass applies only to immutable parent `566ab08…`; this
  reconciliation requires a fresh audit of the pushed exact head.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
  `golang-concurrency`, `golang-documentation`, `golang-lint`,
  `vercel-react-best-practices`, and `vercel-composition-patterns`.

## Testing

- Focused conflict tests passed in 19.500s; full changed packages passed:
  `cmd/connectorgen` (177.850s), engine (9.928s), commandrunner (22.072s),
  Asana defs (5.582s), App (267.112s), and CLI (456.149s).
- Source import verified 249 Asana operations; validation reported 553
  connectors/zero findings; surface sync made zero corrections; operation
  evidence reported 1,774 rows/five rollups with fixed-100 green.
- Declaration admission, runtime preflight, canon, connector-boundary,
  certification subject/matrix/candidates/sweep, scoped vet, tidy, lint,
  agent-contract, docs validation, and 34 website script tests passed.
- Rebuilt `pm`; in 212 isolated initialized credential-free projects, all 212
  commands reached `missing --credential` (exit 1), with no unknown, blocked,
  or provider result.

`npm --prefix website run typecheck` could not start because `tsc` is absent.
Aggregate `go test ./...` and `make verify` were not run locally due the
runner's per-command limit; they remain CI checks and are not claimed as local
passes.

## CLI / Docs / Website

- Verified `pm help docs`, bare `pm asana`, `pm asana access-requests get-access-requests --help`, and `pm docs --help`.
- Refreshed generated Asana manual/skill, connector catalog, and website connector data. No top-level command-tree change was required.

## Safety / Scope

- No credentials were read, logged, or stored; no provider calls were made.
- No arbitrary URL, method, path, header, body, shell, HTTP write, SQL write, or curl escape hatch was added.
- Legacy ETL, direct-write, reverse-ETL, binary, and delete behavior is covered by the runner/Asana regressions.
- Every eligible historical Batch-1 planned GET is now generated from its
  concrete locked contract, not merely status-promoted. A future lock refresh
  must rerun source import/check and leave any non-scalar/header/body or absent
  typed contract deferred with a concrete named foundation; it must never use
  historical `planned` metadata to hide an otherwise capable source operation.
- The 19 source-complete DELETEs and two no-body POST mutations are now
  promoted through the existing typed reverse-ETL/delete path. The executable
  total is 212 (106 direct reads, 12 ETL, 94 reverse ETL); 37 only retain exact
  source-cited unavailable foundations. No provider behavior was invoked.

## Follow-up / Integration

- Current `main` `b9b2478…` is integrated through normal Git ancestry. Its
  write-disabled mutation-artifact policy is preserved alongside this repair;
  no source lock or provider bytes were rewritten.
- Do not merge from this task. Obtain a fresh independent Codex audit of the
  pushed exact head before human merge consideration.
- Automated review route: Claude automatic review is expected on PR open. Confirm it covers the final commit range, disposition any actionable findings, and use Copilot only if Claude is unavailable.

## Checkpoints

- `80f36f8c1` — planning/TDD checkpoint
- `6f41f52ae` — implementation, generated artifacts, and verification evidence
- current-main reconciliation — regenerated outputs and 212-command
  credential-boundary census

## R6 repair

- All 37 unavailable Asana rows now return only their exact generator-owned
  `missing_foundation` or `not_applicable=generic_batch_wrapper` disposition,
  before lifecycle/executor fallback. The runner refuses arbitrary prose as a
  declaration authority. Rebuilt-binary schema, encoding, OpenAPI-sibling, and
  batch examples all exit 7 with their declared reason.
- `Connector.Read` and `ReadWithOutcome` now structurally preflight
  source-bound ETL identity before authentication admission, while keeping the
  existing inner check as defense in depth. The adapter regression asserts zero
  auth and requester calls on structural source-path drift; the shared
  preflight retains its existing method/records/pagination controls.
