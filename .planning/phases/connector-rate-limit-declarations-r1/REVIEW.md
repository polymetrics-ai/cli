# Code review — provider-cited rate-limit declarations R1

## Method

The generated `code-review` prompt requires a numeric phase and a GSD reviewer role. This task
uses a named foundation phase and the worker contract forbids lifecycle-role spawning, so review
was performed inline as the documented fallback.

Reviewed the production embed change, the retained 24 first-batch files, and all 25 second-batch
files against the rate-limit schema, their connector `spec.json` scope properties, and the cited
provider-policy research recorded in the resumability ledger. Re-ran engine, commandrunner,
generator, lint, boundary, docs, smoke, and release gates.

## Findings

No actionable findings.

- All declared policies have clean HTTPS provider URLs and ISO retrieval dates.
- The three declared policies use existing non-secret account scopes: `account_id` for Harvest and
  CallRail, and account-specific `base_url` for Aha!.
- The remaining policies use `unknown` with a nonblank reason and no policy payload.
- The embed test demonstrated the declaration would be absent without the `defs.FS` wildcard,
  then passed after it was added.
- The population recheck joined every current declaration to a `done` sweep record with an
  enumerable provider artifact. It removed Vercel because it is absent from the sweep; no dead or
  retired surface received a declaration.
