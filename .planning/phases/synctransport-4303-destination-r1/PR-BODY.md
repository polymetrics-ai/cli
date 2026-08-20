Refs #4303

## Intent

Make connector-owned reverse-ETL destinations declarable and reachable through
the actual persisted application and CLI path, without introducing a generic
HTTP writer.

## What Changed

- Adds `declarative_api/declarative_typed_destination`, a shared factory for
  schema-backed `writes.json` actions selected only by destination
  declarations and per-connector conformance evidence.
- Keeps GitHub's issue-label route on its specialized typed adapter and
  provider-state read-back while composing it through the common evidence and
  factory loop.
- Persists one exact `destination_action` on a stream when a connector offers
  multiple eligible typed actions. Runtime callers cannot supply or override
  an action, route, method, body, source mapping, connector, or evidence.
- Binds plan, preview, approval, workset ownership, acknowledgement, receipts,
  and result projection to that definition-owned action.
- Documents the destination declaration and persisted-selection contracts in
  the sync-transport, CLI, skill, and website documentation.

## Application Dispatch Proof

- Production-shaped synthetic bundles prove two actions in one connector and a
  third action in another connector can each be persisted, planned, previewed,
  approved, applied, acknowledged, read back, and checkpointed independently.
- Missing, foreign, stale, unlisted, malformed, wrong-mode, wrong-source, and
  wrong-schema selections fail before provider I/O. `change_capture` remains
  ineligible as a destination mode.
- Exact selected-action schemas admit their own snake_case or camelCase input
  fields and reject cross-action, cross-connector, generic, shell, HTTP, and
  undeclared names before I/O.
- Completed runs retain ordinary provider status, headers, nested fields, large
  numeric values, and tier-specific fields. Credential material is represented
  only by an in-place `{ "masked": true }` marker.

## GitHub Compatibility

The real GitHub issue-label route passed its API proof before the bespoke
composition was retired: add, set, keyed replay, acknowledgement, checkpoint,
and independent provider-state label read-back all remained intact.

## TDD and GSD

- Red: synthetic declarative destinations initially failed because their exact
  executor was not registered; persisted multi-action dispatch initially lacked
  a closed selected-action identity.
- Green: generic composition, distinct evidence, refusal-before-I/O cases,
  persisted multi-action selection, exact action-schema mappings, approval,
  acknowledgement, and complete result projection pass.
- GSD lifecycle ran inline: `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review` prompts were resolved with
  `scripts/gsd`. Compatible isolated workers were unavailable and the canonical
  single-worker contract forbids role spawning; the fallback is recorded under
  `.planning/phases/synctransport-4303-destination-r1/`.
- Skills: `golang-how-to`, `golang-cli`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
  and `golang-documentation`.

## Safety

- No connector definitions changed.
- No generic HTTP, SQL, shell, arbitrary-operation, route, verb, or body write
  surface was added.
- The reverse path remains plan → preview → approval → execute.
- Real GitHub credentials were confined to the explicitly requested live-test
  process environment and were never written or displayed.

## Testing

- Focused persisted-dispatch, source-binding, connector result, engine output,
  and CLI help/golden tests passed.
- `make connector-boundary` was detached and polled locally: clean, 552
  connectors, 294 files, no findings or warnings.
- `make verify` passed locally, including the complete Go test tree, docs,
  smoke, lint, generator/certification, connector-boundary, canon, and release
  checks.
- The fresh Website Data failure identified stale generated documentation, not a
  runtime or connector defect. `cd website && pnpm run gen:website-data`
  regenerated `website/lib/docs.generated.ts`; a second run was idempotent.
  Website lint, typecheck, unit tests, script tests, and production build all
  passed locally.
- `go test -tags=databaseintegration -count=1 -timeout 20m -v ./internal/cli -run '^TestPMBinaryExecutesLivePostgresWarehouseGitHubIssueLabels$'` passed: add, set, keyed replay, independent API label read-back, acknowledgement, and checkpoint.
- `git diff --check main...HEAD` passed.

## Review

Inline standard review found no actionable issue. Automated review primary
route remains `claude_auto`; no fallback review was requested. The fresh local
revalidation and review record are in the phase artifacts.

## Commit Checkpoints

- `58c32446d` docs(synctransport): plan typed destinations (Refs #4303)
- `c6f03c937` feat(synctransport): compose typed destinations (Refs #4303)
- `609f23bb3` feat(synctransport): compose typed reverse-ETL destinations
- `c8e75083c` docs(synctransport): record destination revalidation (Refs #4303)
- Generated website-data repair (this commit; Refs #4303)
