# Plan — Stripe provider-dialect tolerance foundation

## Scope

Shared `cmd/connectorgen` source-import behavior only, tracked as an incremental `Refs #4336` direct PR to `main`.

## Required skills

- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`

CLI help/manual/website parity is not applicable: the work neither changes a `pm` command nor materializes a connector surface. Generator/docs checks remain required.

## TDD slices

1. **Red — retained Stripe operation evidence.** Add focused importer coverage containing the exact Stripe `GetAccount` GET and `DeleteAccountsAccount` DELETE source identities, plus a nested local `$ref` chain. Assert the current global preflight aborts with `reference depth limit exceeded`; record the failing command.
2. **Green — source-local resolution.** Add source-document-local normalized-reference memoization and retain the existing per-traversal cycle/reference/resource checks. Replace all-document reference preflight with per-operation resolution so exact operations produce full descriptors.
3. **Green — bounded operation-local gap.** Introduce a typed finite-depth error. For v2/v3 gap-enabled locks only, turn that error into a source-cited skeletal descriptor carrying `cli-source-descriptor-foundation-r1`, its exact operation location, and a merge-blocking runtime gap. Preserve hard rejection for malformed, external, cyclic, ambiguous, target-kind, and resource-exhaustion errors.
4. **Refactor — lock projection and generator controls.** Ensure operation identity validation accepts the retained descriptor, source-descriptor serialization is stable, and source-projection cannot materialize a command from an incomplete descriptor. Update verification evidence only as a consequence of these tested behaviors.
5. **CI-repair Red/Green — current-main preflight without depth abort.** After integrating #4358, restore its complete source-grammar preflight and reproduce the full package failure. Keep its validation of unused malformed/dynamic references and non-response expansion exhaustion, but recognize only typed finite depth exhaustion for gap-enabled locks. Continue scanning later component entries so depth cannot mask an independent external or malformed reference; let the per-operation importer produce the exact incomplete descriptor.

## Verification plan

- `go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImportStripeReferenceDepthOperationLocal|TestSourceImportRejectsUnsafeReferences' -count=1`
- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceImportPreflightsUnusedGrammarObjects|TestSourceImportReservesNonResponseReferenceExpansionBeforeCloning|TestSourceImportStripeReferenceDepthOperationLocal|TestSourceImportRejectsUnsafeReferences)$' -count=1`
- `go test -timeout 20m ./cmd/connectorgen -count=1`
- `go test -timeout 20m ./internal/connectors/engine -count=1`
- `go test -timeout 20m ./internal/connectors/commandrunner -count=1`
- `go vet ./...`; `go build ./cmd/connectorgen`; `go build ./cmd/pm`
- individual `make` gates: tidy-check, lint, docs-check, smoke-no-build, agent-contract-check, connectorgen-validate, connectorgen-surface-sync, connector-boundary, release-workflow-check
- operation-evidence and declaration-admission commands identified from the Makefile.

## Commit checkpoints

1. Commit planning evidence after the red test is established and its failure recorded.
2. Commit the tested implementation after focused tests and changed-package tests pass.
3. Commit any review fix only when it changes production or test behavior; do not create evidence-only commits after push.
