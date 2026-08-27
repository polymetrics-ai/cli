## Summary

- replace connector-wide OpenAPI component preflight with bounded,
  source-document-local reference normalization and target memoization;
- retain only an exact over-depth Stripe operation as a source-cited,
  merge-blocked `cli-source-descriptor-foundation-r1` descriptor;
- retain hard rejection for malformed, external, ambiguous, cyclic, and
  resource-exhausting references.

Refs #4336

## Scope and safety

This is importer foundation work only. It does not change Stripe definitions,
source locks, retained source bytes/hashes/certificates, lane mappings,
credentials, delete/reverse-ETL policy, runtime safety, or provider I/O. It
does not add a raw HTTP/arbitrary-route/body/shell escape hatch and it does not
claim a `pm stripe` command is runnable. The full retained Stripe artifact is
not on current `main`; focused hermetic tests preserve the exact known Stripe
source URL, operation IDs, methods, paths, and source locations.

## GSD / TDD evidence

- `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
  `code-review` prompts were generated with `scripts/gsd` and executed inline;
  inline fallback is recorded because compatible isolated roles are forbidden
  by the canonical project contract.
- Red: the old importer aborted all source operations at document preflight
  with `reference depth limit exceeded`.
- Green: focused tests prove complete Stripe GET/DELETE retention, resolved
  nested contracts, operation-local gap retention, source-projection block,
  unsafe-reference hard rejection, canonical memoization, and counted
  traversal.

Required skills used: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, and `golang-lint`.

## Verification

See `.planning/phases/cli-provider-dialect-tolerance-foundation-r1/VERIFICATION.md`
for exact commands/results. Passing local gates include focused and changed
package tests, engine/commandrunner regressions, vet, builds, docs, smoke,
lint, agent contract, generator validation, surface sync, declaration
admission, operation evidence, certification, connector boundary/canon, and
release workflow checks. Full `go test ./...` and monolithic `make verify` are
intentionally delegated to CI per the repository's per-command worker
guidance.
