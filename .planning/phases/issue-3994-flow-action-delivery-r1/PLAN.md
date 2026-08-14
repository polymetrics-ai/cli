# Issue #3994 — Connector-backed flow action delivery

## Task Delivery Header

- Issue: Closes #3994 — Flow: route approved actions through the common connector write path
- Base branch: integration/4015-mvp-flat-r1
- Merges into: integration/4015-mvp-flat-r1 → main
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its required checks green.
- Working branch: fm/cli-3994-flow-action-delivery-r1
- Task: Route approved flow action steps through the scoped warehouse source reader and typed connector write path, backed by the durable authorization scope introduced in #4132. A successful action is acknowledged, receipt-recorded, independently verified, then checkpointed.
- Verification: `go test -timeout 20m ./internal/flow/...`, `go test -timeout 20m ./internal/app/...`, targeted CLI tests, build/help/docs checks, and the required CI checks on the opened PR.

## Scope and decisions

- Reuse `internal/app/authorization.go` as the only durable, content-free authorization identity. Do not duplicate authorization records or token handling.
- Reuse the existing connection-scoped `App.ReadActionSource` path. An omitted owner must remain a fail-closed ambiguity before dispatch.
- A flow action must use the connector's typed `ValidateWrite` and `Write`; `HTTPActionRunner` remains legacy test coverage only and is never assembled by `flowRun`.
- The action runner must resolve the destination credential/runtime from the manifest. It must not accept a raw URL or generic HTTP/SQL writer.
- Approval is optional only after an existing durable authorization reference validates against the re-derived scope; scope, expiry, revocation, and confirmation failures happen before `Write`. The one-time approval token is consumed by the existing reverse-ETL approval lifecycle, not passed to a scheduled flow run.
- Receipt persistence follows a successful connector acknowledgement and precedes the engine checkpoint. Read-back is a required success condition.

## TDD slices

1. **RED:** add connector-backed action-runner tests which prove a changed scope yields zero destination validation/write/read-back calls and no receipt/checkpoint; prove receipt ordering after acknowledgement/read-back.
2. **GREEN:** add the narrowly typed app action execution seam using runtime resolution, scope authorization, record mapping, validation, approval evidence, write, and read-back.
3. **GREEN:** wire `flow run` to accept the durable `--authorization` reference and construct the connector-backed runner; remove `HTTPActionRunner` from the production construction path.
4. **REFACTOR:** retain only small consumer-owned interfaces and update the flow help/manual/web parity surfaces and generated transcript evidence.
5. **VERIFY:** execute scoped package tests and review the diff for raw URL/generic-write regressions. Live proof needs a captain-runbook credential; if unavailable, record the named deferred proof without secrets.

## Required skills and workflow

- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-context`.
- GSD commands resolved inline because the compatible isolated Pi runtime is unavailable and the delivery contract requires one worker: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, then `code-review`.
- CLI parity applies: runtime `pm flow`, `pm help flow`, `pm flow --help` where available, `docs/cli/flow.md`, website documentation/generated surface, and golden help output are checked or explicitly marked not applicable.

## Commit checkpoints

1. Planning/TDD evidence.
2. Failing-test evidence (where repository policy permits a red checkpoint).
3. Green implementation plus targeted package verification.
4. Review fixes and final verification evidence.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| An action uses the selected connection-scoped source rows | fake | A hermetic app/connector fixture is necessary because a real provider credential is unavailable; it asserts the destination received only the selected owner’s mapped row. |
| A changed authorization scope stops before provider dispatch | fake | The production `flowRun` composition test records target event counts and proves scope drift produces zero `ValidateWrite`, `Write`, and `Read` calls, plus zero receipts. The durable authorization foundation independently covers revocation, expiry, and one-time token replay. |
| Connector validation and typed write receive approval evidence | fake | A test connector records the typed request and verifies the evidence gate, impossible to observe from an external provider without a controlled target. |
| Receipt follows acknowledgement and read-back, and checkpoint follows both | fake | A connector fixture observes the app receipt count inside `Write` and `Read`, proving it is still zero until both complete; the production composition test then observes the successful checkpoint. |
| Production flow route contains no generic HTTP action dispatch | live | The real CLI construction test executes `flowRun` with a registered connector; its observable target mutation/read-back would be absent if the runner were not wired. |
| Help/manual/website describe the approval-gated action lifecycle | live | Actual generated manual/help and documentation checks assert the lifecycle wording and supported safe inputs are present. |

## Assertion Rule

Every test asserts an observable mutation, request count, ordering event, or recorded result. A returned nil error is never used as proof.
