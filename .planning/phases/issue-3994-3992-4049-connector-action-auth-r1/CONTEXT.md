# Connector action identity, per-fire grant, and rate refusal context

**Issues:** #3994, #3992, #4049  
**Base:** `integration/4015-mvp-flat-r1`  
**Mode:** inline/manual GSD fallback; the canonical single-worker contract forbids role spawning.

## Outcome

Close the three adjacent audited residuals without changing the standing authorization model:

1. a connector action carries a payload-bound prepared-execution identity from the production
   flow path through its durable receipt;
2. every firing derives and consumes exactly one short-lived grant from the already-approved,
   content-free authorization scope before the physical write;
3. an unavailable coordinator under `require_shared` returns an `errors.As`-visible
   `*connsdk.RateBudgetRefusalError` whose code is `shared_coordinator_unavailable`, before
   transport.

## Locked decisions

- The existing `AuthorizationRecord` remains the durable, revocable, content-free approval. No
  payload, record count, run ID, token, credential, or raw configuration is added to it.
- A prepared-execution identity is distinct from the scope identity. It binds the authorization
  reference and scope identity to the actual mapped payload, destination write target, preview
  digest, flow/step identity, and run/firing identity.
- A firing grant is opaque, short-lived, authenticated by the existing project approval authority,
  and single-consume across processes. It is never serialized into schedule state, argv, JSON,
  logs, or receipts.
- Scope/revocation/expiry, destination validation, preview, context cancellation, and grant
  authentication all precede durable grant consumption. A refusal at those gates performs zero
  sends/writes, advances no checkpoint, and consumes no grant.
- The grant is consumed immediately before the provider/database write. Cancellation, process
  death, or an ambiguous/partial write after that point parks the firing and cannot replay the
  grant automatically.
- A successful receipt exposes only safe identities: prepared execution, firing, authorization,
  connector/action, and acknowledgement/read-back timestamps.
- Schedule terminal state is durable before its lease cleanup. Failure after a potentially
  non-idempotent write parks before lock cleanup; a successful write/read-back/checkpoint records
  success before removing the lock.
- Rate-budget refusal is an SDK contract layered over the existing safe coordinator cause. It
  preserves `context.Canceled`/deadline errors and retains `SharedRateLimitUnavailableError` in the
  unwrap chain for compatibility.
- #4125 and #4158 are excluded and will not be changed.

## Production composition

- Flow action: `cmd/pm.main -> cli.Run -> runFlow -> flowRun -> flow.Engine.Run ->
  connectorFlowActionRunner -> app.App.ExecuteAuthorizedFlowAction -> typed connector Write`.
- Schedule: `cmd/pm.main -> cli.Run -> runSchedule -> runScheduleFire -> flowRun -> the same flow
  action path`, with schedule lease persistence around the firing.
- GitHub rate admission: `cmd/pm.main -> cli.Run -> app/connector composition -> GitHub WriteHook
  -> engine.Runtime.RequesterFor -> connsdk.Requester admission -> HTTP transport`.

## Current-base constraint

The base publishes PostgreSQL as source-only: `internal/connectors/defs/postgres/sync_transport.json`
has no destination transport and `metadata.json` has `write: false`. Therefore a production R2
GitHub-to-PostgreSQL scheduled destination cannot truthfully pass on this branch. #3982 owns that
destination publication, while #4158 owns the currently failing live managed-target assertion. This
task will prove grant carriage through the real binary and action path, run any available PostgreSQL
read/control evidence, and record the exact R2 limitation without editing either excluded issue.

## Required edge contract

Cancellation, process death, replay, grant expiry/revocation, approval refusal, coordinator
unavailability, concurrent grant consumption, and partial-write cleanup ordering each need a typed
outcome plus an observable negative side-effect assertion. Hermetic tests use only local transports,
injected clocks, and isolated project roots; live GitHub/PostgreSQL evidence is run only when the
required environment is present and never renders credentials.

