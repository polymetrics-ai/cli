---
phase: cli-3987-four-path-conformance-r1
issue: 3987
mode: tdd
status: complete
---

# Plan — #3987 four-path warehouse conformance

## Delivery and fallback

The inline/manual GSD fallback is required: the canonical delivery contract forbids role spawning and this runtime cannot provide a compatible isolated Pi role executor. Resolved commands are `discuss-phase --auto`, `plan-phase --tdd --auto`, `execute-phase`, `verify-work`, and `code-review` through `scripts/gsd prompt`.

Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-database`, `golang-code-style`, and `golang-lint`.

## Scope

Add only the conformance proof gap. The existing live/fresh-binary route proofs remain their owners’ evidence. The implementation will be an `internal/app` production-composition conformance test plus required GSD evidence; it will not alter connector runtime behavior, definitions, generated surfaces, live certification status, CLI output, or the out-of-scope certification roll-up.

## TDD slices

1. **RED — exact four-direction matrix.** Add a focused test that derives the required IDs from `certificationcatalog.FlowKinds`, resolves the four GitHub/PostgreSQL source/destination pairs through the real application registry, and asserts direction-specific source/destination references. It must fail before the matrix helper/coverage assertions exist and make a missing or substituted direction visible by its own subtest name.
2. **GREEN — warehouse boundary and current mode truth.** Extend the conformance proof to assert source/destination role distinction, sealed warehouse-stage/reopen requirements from the resolved production transport, current closed-mode admission, executable history routes, and `change_capture` source-only refusal from a PostgreSQL destination. Keep direct reads, credentials, and external writes out of the deterministic test.
3. **REFACTOR — fail-closed structure.** Remove duplicate fixture setup, retain a one-to-one map between generated flow kinds and named contracts, and make the result/roll-up assertion reject any missing, duplicate, failed, skipped, or unexecuted direction.
4. **Scratch proof.** After connector schema compilation succeeds, temporarily change only the API→API source binding to a still-schema-valid but wrong stream, run the named API→API matrix subtest to red, restore the exact declaration, and rerun the focused test green. Record both commands and outcomes in the ledger.

## Acceptance-to-proof mapping

| Acceptance | Test / observable result |
| --- | --- |
| Four named contracts | Named `api_to_api`, `api_to_database`, `database_to_api`, and `database_to_database` subtests resolve their exact real source/destination descriptors and cannot replace one another. |
| Warehouse mediation | The proof asserts the real orchestrator’s stage → reopen → apply → read-back → checkpoint topology and rejects a contract without a sealed receipt/workset boundary. |
| Current modes | `synccontract.AllModes()` is exhaustively accounted for: current declared/executable routes pass only where admitted; history is executed/admitted, never reclassified as stale non-executable. |
| `change_capture` | PostgreSQL’s change-capture source workset is distinct from normal destination modes; a PostgreSQL target request is refused before executor I/O. |
| Honest status | The matrix result model accepts `pass` only after its observed assertions; a nonexecuted/refused record has a concrete non-pass result and never contributes to the pass roll-up. |
| Can fail | Scratch binding defect fails the specifically named API→API subtest after schema validation. |

## Verification plan

```sh
go test -count=1 -timeout 20m ./internal/app -run 'TestWarehouseMediatedFourPathConformance'
go test -count=1 -timeout 20m ./internal/app ./internal/synctransport ./internal/connectors ./internal/synccontract
go test -count=1 -timeout 20m ./internal/cli
go vet ./internal/app ./internal/synctransport ./internal/connectors ./internal/synccontract ./internal/cli
go build ./cmd/pm
go run ./cmd/connectorgen validate internal/connectors/defs
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen certification-matrix --check
./connectorgen boundary . --json
make tidy-check lint docs-check-no-build smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-runtime-preflight connector-canon-check release-workflow-check
pnpm --dir website run gen:docs
pnpm --dir website run gen:docs
```

`make verify` and a single `go test ./...` are intentionally not run as one local command because repository guidance says they exceed the per-command execution window. CI remains the full-suite authority.

## CLI/docs parity

Not applicable: this change adds no command, flag, help text, user-visible output, manual, or website content. The existing command and generated artifact gates remain in the verification set.
