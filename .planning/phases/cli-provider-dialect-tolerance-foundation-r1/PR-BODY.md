## Summary

- restore the exact pinned Stripe source evidence and generate its canonical
  589-operation descriptor through the retained-artifact importer;
- keep bounded reference processing source-local: a byte-backed source lock can
  only lower the reference-depth cap, and an unresolved response becomes an
  exact operation-local `cli-source-descriptor-foundation-r1` condition;
- preserve terminal rejection for hidden malformed, dynamic, external,
  wrong-kind, cyclic, ambiguous, and resource-exhausting references; and
- prove a retained Stripe condition reaches registry-backed commandrunner
  preflight before credentials, executor dispatch, or provider I/O.

Refs #4336

## Scope and safety

The restored artifact is byte-for-byte the historic Batch 1 Stripe capture:
SHA-256 `3653ad45bbec54fcbe461c541c908355b715018bdf455a0e11b27bedb2cbdee5`,
7,967,776 bytes, 589 locked REST source operations. The lock adds only the
clamp-only `reference_depth_limit: 1` importer policy; no source byte, digest,
credential, certificate, or provider request changed.

Generated source evidence keeps all 589 operations mapped through the source
descriptor, source crosswalk/disposition, Stripe API surface, and operation
evidence. No Stripe action, CLI command, route, generic HTTP/body/shell escape
hatch, credential behavior, delete/reverse-ETL behavior, or provider I/O is
added. Credential-bound runnable commands: **0**. Unresolved Stripe source
contracts: **589** precise source-descriptor missing-foundation conditions.

## GSD / TDD evidence

- `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
  `code-review` prompts were generated with `scripts/gsd` and executed inline;
  the documented fallback applies because compatible isolated roles are
  unavailable and forbidden by the canonical contract.
- Red: the previous importer aborted at document preflight on one finite
  reference-depth outcome, omitting unrelated source operations.
- Green: retained-corpus, unused-component, unsafe-depth, and registry
  preflight tests prove the exact boundary and no-I/O behavior.

Required skills used: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, and `golang-lint`.

## Verification

See `.planning/phases/cli-provider-dialect-tolerance-foundation-r1/VERIFICATION.md`
for exact commands/results. Passing local gates include focused and full
`cmd/connectorgen` tests, engine/commandrunner regressions, source import and
check, vet, builds, docs, smoke, lint, agent contract, generator validation,
surface sync, declaration admission, operation evidence, certification,
connector canon, release workflow, and connector-boundary checks. Full
`go test ./...` and monolithic `make verify` remain CI-owned per repository
per-command worker guidance.
