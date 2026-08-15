# Connector action identity, standing job approval, and rate refusal context

**Issues:** #3994, #3992, #4049
**Base:** `integration/4015-mvp-flat-r1`
**Mode:** inline/manual GSD fallback; the canonical single-worker contract forbids role spawning.

## Captain correction applied

The first plan for this combined phase interpreted #3992 as requiring a newly issued, single-use
grant for every scheduled firing. The 2026-08-15 launch correction explicitly rejects that model.
The committed RED checkpoint for per-fire grants is retained as historical evidence, but R2-R4 in
that ledger are superseded here and the grant implementation/tests are removed before proceeding.

Approval attaches once to an existing reverse-ETL job through its durable
`AuthorizationRecord`. A flow composes existing jobs and retains only their safe references. A
schedule retains only the name of an existing flow. Every firing loads those jobs again and
revalidates credential revision, source/table scope, mappings, destination action/configuration,
confirmation policy, expiry, and revocation before any provider request.

## Outcome

1. A connector action carries a payload-bound prepared-execution identity from the production flow
   path through its durable receipt. This identity is execution evidence, not authority.
2. Flow creation positively resolves sync jobs to existing connections and action jobs to existing
   reverse plans with live standing authorization before writing the flow.
3. Schedule creation and installation positively resolve an existing, valid flow before writing a
   schedule or backend entry. Schedule files and rendered commands contain no authorization
   reference, token, credential, secret, or secret-derived preimage.
4. The installed command remains `pm --root <root> flow run <name> --json`. The named flow is
   associated with its single schedule at runtime so the existing persisted fire lease still parks
   drift, cancellation, ambiguous acknowledgement, and cleanup failure without inventing another
   authorization state machine.
5. An unavailable coordinator under `require_shared` returns an `errors.As`-visible
   `*connsdk.RateBudgetRefusalError` with code `shared_coordinator_unavailable` before transport.

## Locked decisions

- `AuthorizationRecord` is the sole durable approval. No per-fire grant, approval object, approval
  token, MAC, raw credential, payload, or raw configuration is added to schedule/flow state.
- A sync job reference names an existing App connection. An action job reference names an existing
  `ReversePlan` whose one-time plan/preview/proceed already minted a standing authorization. The
  action's executable scope is hydrated from that plan rather than trusted from duplicated inline
  manifest fields.
- Missing, malformed, unrecognised, unapproved, expired, revoked, or drifted job references return
  typed errors naming the reference. Creation writes nothing on refusal.
- A prepared-execution identity is distinct from the content-free scope identity. It binds the
  standing authorization and current mapped payload, write target, preview, flow/step, and run.
- A safe prepared-execution lease prevents concurrent/replayed dispatch of the same prepared
  execution. It is non-authoritative replay/ambiguity state, never a token or grant. Refusal before
  dispatch releases it; once dispatch may have occurred it remains halted.
- The existing schedule fire lease owns running/succeeded/parked ordering. Terminal state is durable
  before lock cleanup; partial or ambiguous writes do not advance a flow checkpoint or auto-replay.
- Rate-budget refusal preserves cancellation/deadline errors and the existing typed shared
  coordinator cause in its unwrap chain.
- #4125 and #4158 remain excluded.

## Production composition

- Flow action: `cmd/pm.main -> cli.Run -> runFlow -> flowRun -> flow.Engine.Run ->
  connectorFlowActionRunner -> app.App.ExecuteAuthorizedFlowAction -> typed connector Write`.
- Scheduled flow: `cmd/pm.main -> cli.Run -> runFlow -> flowRun -> schedule.BeginFire -> the same
  connector action path -> schedule.FireLease.Complete/Park`.
- GitHub rate admission: `cmd/pm.main -> cli.Run -> App connector composition -> GitHub WriteHook ->
  engine.Runtime.RequesterFor -> connsdk.Requester admission -> HTTP transport`.

## Live-proof constraint

The base publishes PostgreSQL as source-only; #4158 is excluded. Available GitHub/PostgreSQL
environment capability is detected without rendering credential values. Where live execution is
unavailable, the issue/PR evidence names the exact gap and the hermetic production-path proof used
instead.
