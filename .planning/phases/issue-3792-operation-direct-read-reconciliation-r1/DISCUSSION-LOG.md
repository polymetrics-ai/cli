# Discussion log — #3792

## Settled constraints

- The direct-read executor already accepts only `rest_read` and
  `provider_search`; the missing admission step is command preflight.
- Preflight must use engine-owned metadata rather than a copied
  `connectorgen` rule, so `TestEveryImplementedCommandPassesRuntimePreflight`
  continues to sweep real runtime behavior.
- Direct-read policy is executable when the engine accepts the command's
  explicit policy for the declared operation endpoint. It is not blindly
  overwritten with an operation default: existing `surface-sync` deliberately
  preserves supported direct-read policies.
- Reconciliation binds an `api_surface` row to an implemented command only
  after the real `commandrunner.Preflight` succeeds. A declared, planned, or
  preflight-failing command remains a blocked reason, never coverage.
- Ambiguous/malformed source data, unsupported operation models, and bundle
  load failures are refusals; the tool leaves the source unchanged.

## Chosen shape

The engine exposes the closed, no-network
`OperationDirectReadPreflighter.PreflightOperationDirectRead` contract. It
uses the same static admission helper as execution to validate operation ID,
kind, method, connector-relative path, bounded cap, endpoint ledger presence,
and output policy. Commandrunner supplies the command's complete binding to
that runtime interface; it does not duplicate engine eligibility rules.

`connectorgen surface-reconcile` will load the disk bundle through the engine,
then call the real `commandrunner.Preflight` for each candidate command. Its
check mode is the review/report mode; write mode changes only deterministically
derived `operation`/`covered_by` classification fields and reasons.
