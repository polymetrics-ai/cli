# Code review — provider-cited rate-limit declarations R1

## Method

The generated `code-review` prompt requires a numeric phase and a GSD reviewer role. This task
uses a named foundation phase and the worker contract forbids lifecycle-role spawning, so review
was performed inline as the documented fallback.

Reviewed the production embed change and all 25 declaration files against the rate-limit schema,
their connector `spec.json` scope properties, and the cited provider-policy research recorded in
the resumability ledger. Re-ran engine, commandrunner, generator, lint, boundary, docs, smoke,
and release gates.

## Findings

No actionable findings.

- All declared policies have clean HTTPS provider URLs and ISO retrieval dates.
- Both declared scopes use existing non-secret `account_id` fields.
- The remaining policies use `unknown` with a nonblank reason and no policy payload.
- The embed test demonstrated the declaration would be absent without the `defs.FS` wildcard,
  then passed after it was added.
- The population recheck joined every declared connector to a `done` sweep record with an
  enumerable provider artifact; no dead or retired surface received a declaration.
