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

CLI help/manual/website parity is not applicable: the work neither changes a `pm` command nor materializes a runnable connector surface. The audit-repair fixture deliberately projects a pre-existing, blocked declaration only to prove the source descriptor reaches the registered runtime boundary before credential or provider I/O; it adds no user-facing command, flag, help, manual, or website surface. Generator/docs checks remain required.

## TDD slices

1. **Red — retained Stripe operation evidence.** Add focused importer coverage containing the exact Stripe `GetAccount` GET and `DeleteAccountsAccount` DELETE source identities, plus a nested local `$ref` chain. Assert the current global preflight aborts with `reference depth limit exceeded`; record the failing command.
2. **Green — source-local resolution.** Add source-document-local normalized-reference memoization and retain the existing per-traversal cycle/reference/resource checks. Keep all-document reference preflight, but let a typed finite depth outcome reach per-operation resolution rather than aborting unrelated source operations.
3. **Green — bounded operation-local gap.** Introduce a typed finite-depth error. A byte-backed v1/v2 lock may declare only a lower `rest.reference_depth_limit`; for a gap-enabled lock, turn that error into a source-cited skeletal descriptor carrying `cli-source-descriptor-foundation-r1`, its exact operation location, and a merge-blocking runtime gap. Preserve hard rejection for malformed, external, cyclic, ambiguous, target-kind, and resource-exhaustion errors.
4. **Refactor — lock projection and generator controls.** Ensure operation identity validation accepts the retained descriptor, source-descriptor serialization is stable, and source-projection cannot materialize a command from an incomplete descriptor. Update verification evidence only as a consequence of these tested behaviors.
5. **CI-repair Red/Green — current-main preflight without depth abort.** After integrating #4358, restore its complete source-grammar preflight and reproduce the full package failure. Keep its validation of unused malformed/dynamic references and non-response expansion exhaustion, but recognize only typed finite depth exhaustion for gap-enabled locks. Continue scanning later component entries so depth cannot mask an independent external or malformed reference; let the per-operation importer produce the exact incomplete descriptor.
6. **Independent-audit repair Red — retained Stripe corpus.** Restore the immutable Batch 1 Stripe source evidence already retained in repository history: its source lock, retained-artifact manifest, exact 7,967,776-byte content-addressed artifact, and source crosswalk/disposition evidence. Add a red test that uses the connector-owned retained fetcher, asserts all 589 locked descriptors are emitted with unique IDs, preserves known GET/DELETE identities, and observes the actual response-reference depth condition as one operation-local `cli-source-descriptor-foundation-r1` gap per locked source operation.
7. **Independent-audit repair Green — unused depth disposition.** Keep full document grammar scanning. A source-contract-gap lock may skip only typed depth while importing source operations; every unreferenced over-depth component must subsequently be either rejected or recorded as a deterministic source-cited document gap. Add a synthetic fixture whose normal source operation remains complete while an unused nested reference cannot vanish unnoticed.
8. **Independent-audit repair Green — registry-to-preflight proof.** Project an actual incomplete retained source descriptor through a declaration-owned bundle, register the engine-backed connector through the repository registry seam, then assert `commandrunner.Preflight` returns the exact source descriptor `missing_foundation` before credential resolution, executor dispatch, or provider I/O. This is a blocked-state proof, not command materialization.

## Verification plan

- `go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImportStripeReferenceDepthOperationLocal|TestSourceImportRejectsUnsafeReferences' -count=1`
- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceImportPreflightsUnusedGrammarObjects|TestSourceImportReservesNonResponseReferenceExpansionBeforeCloning|TestSourceImportStripeReferenceDepthOperationLocal|TestSourceImportRejectsUnsafeReferences)$' -count=1`
- `go test -timeout 20m ./cmd/connectorgen -count=1`
- `go test -timeout 20m ./internal/connectors/engine -count=1`
- `go test -timeout 20m ./internal/connectors/commandrunner -count=1`
- `go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImportRetainedStripeCorpus|TestSourceImportRetainsUnusedDepthDisposition|TestSourceProjectionRetainedStripeDepthGapStopsAtRegistryPreflight' -count=1`
- `go vet ./...`; `go build ./cmd/connectorgen`; `go build ./cmd/pm`
- individual `make` gates: tidy-check, lint, docs-check, smoke-no-build, agent-contract-check, connectorgen-validate, connectorgen-surface-sync, connector-boundary, release-workflow-check
- operation-evidence and declaration-admission commands identified from the Makefile.

## Commit checkpoints

1. Commit planning evidence after the red test is established and its failure recorded.
2. Commit the tested implementation after focused tests and changed-package tests pass.
3. Commit any review fix only when it changes production or test behavior; do not create evidence-only commits after push.
4. The audit repair lands as a code/test/source-evidence commit after its red/green checks; do not push or create an audit-evidence-only commit.
