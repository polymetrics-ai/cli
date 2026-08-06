# Manual code review — #3792

## Route

The local Pi adapter did not provide the compatible isolated GSD code-review
worker, and the repository delivery contract forbids role spawning for this
lane. The single worker performed the documented manual fallback after all
local checks. No PR exists yet; firstmate owns the subsequent no-mistakes and
external-review lifecycle.

## Scope reviewed

- `OperationDirectReadPreflighter` wiring through connector, engine, native
  SQS/Ashby adapters, and `commandrunner.Preflight`.
- The shared engine admission helper to ensure execution and preflight use the
  same operation kind, method, path, cap, endpoint, and policy rules.
- `surface-reconcile` candidate matching, error-to-current-reason derivation,
  check-mode non-write behavior, and refusal of unsupported models.
- Tests, migration documentation, and the six-connector check-only report.

## Result

The follow-up review confirmed F1 was legitimate: production `defs.FS` omits
raw `api_surface.json`, so the previous optional-surface path could not prove
the required endpoint ledger. The repair adds a generated compact root
projection at the shared bundle-loading boundary and fails direct-read
preflight closed when that projection is missing, unresolved, incomplete, or
malformed. Focused engine, commandrunner, and connectorgen tests pass,
including the unchanged runtime sweep; no credential handling, network
execution in preflight/reconciliation, connector promotion, or targeted-ledger
write remains.
