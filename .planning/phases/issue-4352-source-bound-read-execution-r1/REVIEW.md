# Local review — #4352

## Scope and source-bound review

- No arbitrary HTTP method, URL, path, header, request body, shell, or curl escape hatch was added.
- The command source ID must equal the selected declaration's source ID; engine preflight also compares exact GET method and connector-relative path before dispatch.
- A source-bound direct read remains on the fixed `OperationDirectRead` executor and reaches the existing credential check only after no-network preflight.
- A source-bound collection cannot claim ETL solely because it is a GET: it needs a unique declared stream ETL composite, matching route, records path, schema, and pagination.
- Existing ETL commands do not require the new source binding. This was verified after a full runner regression caught and corrected the overly broad check.
- Source-bound stream interpolation permits only a whole-segment declared
  `{{ config.* }}` or `{{ fanout.id }}` where the locked path has a variable;
  literal route substitution and record-derived path interpolation are refused.
- Configured origins are not a source-bound escape hatch: the engine compares
  `base_url` with the connector's declared origin before authentication or
  requester construction.
- Mapping validation is lane-aware and capability-based. A complete source
  contract that is labelled planned causes source-import drift; a deferred read
  must retain a concrete named source foundation. Hash/byte/capture metadata
  remains raw-source integrity evidence, not mapping admission policy.
- The partial mutation-coverage disposition is exact-source-citation bound and
  accepts only an existing incomplete implemented action plus its named shared
  foundation. It rejects a complete action, an absent action, a non-mutating
  source operation, and an unrelated foundation; the legacy path-alias category
  additionally requires a demonstrable provider/local path-field-name mismatch.

## Findings

No unresolved code finding. The serialized source-lock audit and this branch's
complete `cmd/connectorgen` package test passed; the two repository-wide
commands noted in `VERIFICATION.md` remain CI/PR checks and are not treated as
passing.

## Required external review

Open a PR to `main` and let the repository's configured Claude review run. Confirm it covers the final commit range; disposition any actionable findings before human merge. The PR remains unmerged by this task.

## r4 local review disposition

The repair was reviewed against the six frozen findings. No raw paging flag or
generic provider write path remains; origin rejection occurs before App access;
the App/direct stream boundary validates source route semantics; promoted
deletes retain typed destructive confirmation and 404 idempotency; and generated
docs reflect the declaration counts. Local checks are green, including the
full generator package, source/operation evidence, runtime preflight, canon,
docs/website generation, and a 212/212 credential-boundary census. A fresh
independent Codex audit of the pushed SHA remains required before review
threads can be closed or a human considers merge.
