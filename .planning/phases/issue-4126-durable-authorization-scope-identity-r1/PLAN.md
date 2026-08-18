# PLAN — Issue 4126: durable content-free authorization scope identity

## Task Delivery Header

- Issue: Closes #4126 — feat(app): add the durable content-free authorization scope identity
- Base branch: integration/4015-mvp-flat-r1
- Merges into: integration/4015-mvp-flat-r1 -> main
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its checks green; verify its API-reported base after opening.
- Working branch: fm/cli-4126-scope-identity-r1
- Task: Add the shared durable, revocable App authorization record and content-free scope identity; prove it can authorize repeated identical-scope reverse-plan dispatch without a token while refusing changed, revoked, expired, or replayed authorization before dispatch.
- Verification: `go test -timeout 20m ./internal/app/... ./internal/flow/... ./internal/schedule/...`, targeted authorization tests, required non-suite local gates, no-mistakes pipeline, and green PR checks.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Scope is stable when record content, count, and timestamps differ | live | A real App scope derivation before and after replacing the source warehouse rows with changed content, an added row, and a new timestamp yields an identical identity. |
| Every bound property changes the identity | live | Table-driven assertions alter exactly one of source connection, destination reference, credential revision, stream/table set, mappings, action, configuration digest, enabled operations, confirmation policy, or expiry and observe a different identity. |
| Identical scope runs unattended | fake | A per-test loopback GitHub endpoint is necessary because it is the hermetic existing destination that supports the real plan → preview → destructive-confirmation → provider-write path. It records one first write then two token-less changed-payload writes. |
| Changed scope stops before provider request | fake | A per-test loopback GitHub endpoint is necessary to count the real provider request without credentials or an external write. Mutated mappings return `AuthorizationScopeChangedError` and leave its send count unchanged. |
| Revocation and expiry stop with distinct typed reasons | fake | Separate per-test loopback GitHub endpoints are necessary to count the real provider request without external side effects. Revocation and expiry each return their matching typed error and leave their endpoint send count unchanged. |
| The single-use token is consumed once and replay is rejected | fake | A per-test loopback GitHub endpoint is necessary to observe the real destination request deterministically. First proceed sends once and creates the record; replay returns `AuthorizationTokenReplayError` with the count unchanged. |
| No secret, credential, or token material reaches storage or output | live | State-file and JSON-marshal assertions find the actual approval token and fixture secret absent while retaining only the non-secret authorization reference and derived identity. |

## Scope and exclusions

Allowed production paths are `internal/app/**`: content-free scope/record types,
state persistence, typed refusal errors, and the existing reverse-plan execution
seam. Required evidence files live in this phase directory.

Do not add the flow action runner, schedule firing/install surface, GitHub
destination modes, generic HTTP writes, generic SQL writes, credentials, or new
dependencies. `reversePlanHash` stays unchanged as the payload-bound
single-execution identity.

## Design

`AuthorizationScope` holds only the ten bound shape properties. It retains a
derived configuration digest and credential revision in place of raw material.
Canonical JSON plus SHA-256 produces a stable `ScopeIdentity`. The persisted
`AuthorizationRecord` holds an opaque generated record reference, the identity,
the safe scope, creation/revocation metadata, and expiry. It does not have a
token, raw credential, secret, record, count, timestamp, cursor, or run field.

The first reverse-plan proceed continues to perform all existing preview,
confirmation, seal, and token checks. Its atomic state transition consumes the
token and creates the record. Later token-less execution re-derives the scope
from current plan/runtime metadata, refreshes durable state, rejects revocation,
expiry, or the first differing bound property before source/provider dispatch,
and dispatches only after a match. A supplied token after consumption is a typed
replay refusal. The existing payload hash remains active for the one-time first
proceed and is deliberately not consulted on the authorized repeat path.

## TDD sequence

1. **Red:** Add App authorization tests for content-free scope identity, each
   bound property, token-free repeat, changed scope / revocation / expiry
   zero-send refusals, replay, and state/output redaction. Record the failing
   focused command.
2. **Green:** Add scope/record types, stable identity derivation, safe durable
   storage, revocation, and typed reasons.
3. **Green:** Wire the first atomic plan consumption to record creation and the
   token-less repeat path to pre-dispatch record validation.
4. **Regression:** Format, run targeted and required package tests, vet/build,
   repository gates, then rely on protected-branch CI for delivery checks.
5. **Verify/review:** Complete the verification checklist and manual deep
   cross-file review. Run GSD verify-work and code-review as the documented
   manual-inline fallback.

## GSD and skill evidence

Resolved and generated: `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review`. Required skills loaded: `golang-how-to`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-naming`,
`golang-error-handling`, `golang-security`, `golang-safety`, and
`golang-testing`.
