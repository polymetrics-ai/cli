# Local Verification

Verdict: PASS by the repository-mandated separated local gates.

The GSD programming-loop verifier was invoked with `verify --execute`. Its repo-profile command
`go test ./...` exceeded the helper's per-command window and returned `exit=n/a` after listing only
passing packages. `AGENTS.md` explicitly warns that this monolithic command is routinely cut off
and requires changed packages plus `internal/cli` to run separately. It was not repeated. The
completed separated evidence is below; CI carries the full 550+ connector suite.

| Check | Status | Command | Notes |
| --- | --- | --- | --- |
| Dependency consistency | PASS | `make tidy-check` | `go.mod` and `go.sum` unchanged. |
| Format | PASS | `gofmt -w internal/app internal/cli internal/connectors`; `gofmt -l cmd internal` | GSD format execution also passed. |
| Diff hygiene | PASS | `git diff --check` | No whitespace errors. |
| Static analysis | PASS | `go vet ./...` | Executed directly and by GSD. |
| Lint | PASS | `make lint` | golangci-lint: 0 issues. |
| Core unit packages | PASS | `go test ./internal/connectors ./internal/connectors/engine ./internal/connectors/commandrunner` | Includes closed schemas, direct bypass, and `rest_write` seam. |
| App | PASS | `go test ./internal/app` | Full package, 33 seconds. |
| CLI | PASS | `go test ./internal/cli` | Full package, 381.346 seconds. |
| Conformance/approved fixtures | PASS | `go test ./internal/connectors/conformance ./internal/connectors/defs/asana ./internal/connectors/defs/zendesk-support` | Local replay only; no live provider or credentials. |
| Build | PASS | `go build ./cmd/pm` |  |
| Smoke | PASS | `make smoke-no-build` | Local ETL plus safe reverse-ETL flow completed. |
| Docs | PASS | `make docs-check-no-build` | Connector docs validation passed. |
| Website docs generator | PASS | `node website/scripts/gen-docs-data.mjs` | Generated TypeScript matches MDX source. |
| Connector validation | PASS | `make connectorgen-validate` | 550 connectors, 0 findings. |
| Surface sync | PASS | `make connectorgen-surface-sync` | 550 connectors, 0 drift. |
| Boundary | PASS | `make connector-boundary` | Clean, 0 findings/warnings. |
| Release workflow | PASS | `make release-workflow-check` | Assertions passed. |
| CLI help parity | PASS | `./pm help reverse`; `./pm reverse`; `./pm github repo deploy-key delete --help` | Runtime help, bare namespace, and canonical public command executed. |

## GSD traces

- `scripts/gsd prompt plan-phase delete-confirmation-foundation-r1 --skip-research --tdd`
- `scripts/gsd prompt execute-phase delete-confirmation-foundation-r1`
- `scripts/gsd prompt verify-work delete-confirmation-foundation-r1`
- `programming-loop.mjs verify --phase delete-confirmation-foundation-r1 --execute`
- Strict TDD gate: `.planning/phases/delete-confirmation-foundation-r1/TDD-GATE.json` (`passed:true`)

No dependency install, live credential, live provider call, browser UAT, or runtime service was
required for this Go-only engine/CLI foundation.

## Review-fix verification

The trusted-input hardening round rechecked only the packages in its changed area. The initial
focused command passed connectors, vault, engine, GitHub hooks, conformance, app, Asana, and
Zendesk Support, then exposed two zero-record SQS preview regressions. After correcting the shared
no-op prepared-write path, the final focused command passed:

`go test ./internal/connectors/engine ./internal/connectors/native/amazon-sqs -count=1`

The subsequent rollback audit made consumption stable for the sealed plan identity and retained the
opaque nonce in the authenticated marker. Its final targeted check passed:

`go test ./internal/connectors ./internal/vault ./internal/app -run 'Test(ProcessWriteApprovalRequiresSealedPlanAndPersistentConsumption|WriteApprovalConsumptionMarkerIsMonotonic|ConsumedApprovalCannotReplayFromRolledBackStateSnapshot)$' -count=1`
