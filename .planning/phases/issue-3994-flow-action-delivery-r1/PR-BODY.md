Closes #3994

## Intent

Route approved flow action steps through the common connector write path on the
durable, revocable authorization identity from #4132.

## What changed

- `flow run` builds a connector-backed `StepActionRunner` and accepts an opaque
  `--authorization` reference when it is not embedded in the action config.
- An action re-derives and requires the durable scope before connector dispatch,
  maps selected warehouse rows, calls typed validation, revalidates destructive
  previews, passes real durable approval evidence to typed `Write`, reads the
  target back, persists an opaque receipt, then permits checkpoint success.
- `HTTPActionRunner`/`DestURL` remain legacy flow-package test fixtures only;
  neither occurs in the app/CLI production call path.
- Help, generated CLI manual/transcript, website docs, and the GSD/TDD evidence
  describe the durable authorization, target read-back, and receipt lifecycle.

## Red / Green / Refactor evidence

- **Red:** before the app seam existed, the new action tests failed to compile
  because the public connector-backed action API and receipt surface were
  absent.
- **Green:** `TestConnectorFlowActionRunnerScopeDriftStopsBeforeTargetRequest`
  proves changed scope causes zero connector validation/write/read-back calls,
  zero receipts, and no checkpoint. The action fixture validates the real
  durable approval evidence it receives at write time.
- **Green:** receipt tests prove the receipt count is zero during acknowledged
  write and target read-back; the flow composition test observes checkpoint
  success only after the receipt-returning action completes.
- **Refactor:** legacy HTTP runner comments make its test-only status explicit;
  the manifest's documented `upsert` default is preserved by the new runner.

## Verification

- `go test -timeout 20m ./internal/flow/... ./internal/app/...`
- `go test -timeout 20m ./internal/cli`
- `go vet ./...`
- `go build ./cmd/pm`
- `make tidy-check docs-check agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check`
- `make smoke-no-build`
- `make lint`
- `npm --prefix website run typecheck`
- `npm --prefix website run lint` (passes with pre-existing warnings in unrelated website component files)
- `./pm help flow`, `./pm flow --help`, and `./pm flow`

## Delivery and safety

- Base: `integration/4015-mvp-flat-r1`; rebased on the current integration tip
  before push. The API-reported base will be recorded after PR creation.
- No credential, token, payload content, or raw destination configuration is
  persisted in the flow receipt. No generic HTTP, SQL, URL, or raw-operation
  write surface is introduced.
- Live provider evidence is deferred to the captain-runbook procedure on #3994:
  no authorized real-provider credential exists in this isolated worktree and
  none was requested or exposed. The hermetic proof establishes typed writes,
  zero-send scope refusal, approval evidence, read-back, receipt, and checkpoint
  ordering; it does not claim a live provider result.

## GSD and review

- Inline/manual fallback recorded in the planning artifacts because this
  isolated runtime cannot run the compatible Pi workers and the contract keeps
  this task to one worker.
- Resolved: `scripts/gsd prompt discuss-phase 3994`, `plan-phase 3994 --tdd`,
  `execute-phase 3994`, `verify-work 3994`, and `code-review 3994`.
- Required skills: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and `golang-context`.
- Automated review route: `claude_auto` on PR open; status pending. Copilot is
  not requested unless Claude is unavailable.
