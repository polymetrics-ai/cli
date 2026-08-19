# Verification — issue #4291

## Superseded verification record

The earlier complete-map validation is superseded by the 2026-08-19 source-lock completeness defect. PR #4296 is held. Its former `source_operation_count` / `declared_percent` figures were based on the legacy `api_surface.json` denominator and must not be used as provider-coverage evidence.

## Artifact-level red/green evidence

- **RED — batch 6:** `test ! -f` for every proposed batch-6 source lock and disposition ledger passed before implementation.
- **RED — batch 7:** `test ! -f` for every proposed batch-7 source lock and disposition ledger passed before implementation.
- **GREEN — initial map only:** the issue-local strict ledger-invariant check passed against the then-current `api_surface.json` denominator. It correctly checked parity classes, reachability, and reverse-ETL semantics, but it did not prove provider-source completeness. Its former total (**2,099 documented, 575 enabled, 348 commands, 448 writes, 211 deletes**) is therefore an initial crosswalk count, not a complete provider-surface claim.
- **GREEN — Salesloft comprehensive remap:** complete rendered-reference crawl passed: 315 public API-reference pages yielded **211 unique operations** (120 GET, 50 POST, 23 PUT, 18 DELETE), replacing the prior 12-operation/84,498-byte index-derived inventory. `counts.total`, `operations_found`, per-method counts, and `coverage_confidence` are recorded in the source lock; the regenerated API surface and disposition ledger each contain all 211 operations and omit `declared_percent`. The explicit map invariant passed at **211/211/211** (source lock/API surface/disposition rows), with 91 `direct_write` rows carrying the separate generic reverse-ETL foundation-gap attribute.
- **GREEN — official-spec remaps:** Iterable **148/148/148** (source lock/API surface/disposition rows), Klaviyo **345/345/345**, and Intercom **231/231/231**. Each source lock has `counts.total`, per-method counts, `operations_found`, and `complete_machine_readable_specification` evidence. Their former 4/9/10-row boundaries are not retained.
- **GREEN — Freshdesk complete reference remap:** **170/170/170** after parsing all 171 endpoint sections in the provider’s single 3.2MB rendered reference (query examples normalize to 170 method/path operations). This replaces the legacy 10-row API-surface boundary without treating a partial crawl as complete.
- **GREEN — Copper complete reference remap:** **89/89/89** after completing the provider-published 637-document rendered MkDocs corpus (32 GET, 35 POST, 11 PUT, 11 DELETE), replacing five synthetic `HOOK` rows. The five native stream bindings are the provider-documented `POST /v1/<resource>/search` operations proven by `internal/connectors/native/copper/streams.go:5-25`; no command/action is bound, so enabled remains zero.
- **GREEN — post-#4297 declaration reconciliation:** rebased onto `origin/main` at `51dd6d468` and ran `connectorgen surface-reconcile --check --json` for all 20 owned connectors. It found deterministic reason updates only in Help Scout (5), Gorgias (2), and Chatwoot (23), then applied and rechecked them cleanly. It covered **zero** rows: none of those 20 currently has an implemented, endpoint-matching runnable command whose runtime preflight proves the newly available executor capability. The audit therefore retains no false enabled claim; future mapped operations with such a binding will be promoted by this same runtime-backed pass.

## Repository checks

| Command | Result |
| --- | --- |
| `go run ./cmd/connectorgen validate` | pass — `552 connector(s) checked, 0 finding(s)` |
| `go run ./cmd/connectorgen surface-sync --check` | pass — `552 connector(s) scanned, 0 field(s) need synchronization` |
| `go test -timeout 20m ./cmd/connectorgen` | pass |
| `go test -timeout 20m ./internal/connectors/engine` | pass |
| `go test -timeout 20m ./internal/cli` | pass |
| `go vet ./...` | pass |
| `go build ./cmd/pm` | pass |
| `gofmt -w cmd internal && go mod tidy && git diff --exit-code -- go.mod go.sum` | pass — formatting/mod-tidy introduced no module drift |
| `./pm docs validate --connectors-dir docs/connectors` | pass |
| `make smoke-no-build` | pass |
| `make lint` | pass — `golangci-lint` reported `0 issues` |
| `make agent-contract-check` | pass |
| `make connectorgen-validate` | pass |
| `make connectorgen-surface-sync` | pass |
| `node --test scripts/tests/github-combined-operation-ledger.test.mjs scripts/tests/gen-github-graphql-parity.test.mjs scripts/tests/github-source-drift.test.mjs` | pass — 15 tests, 0 failures |
| `node scripts/gen-github-graphql-parity.mjs --check` | pass |
| `node scripts/github-combined-operation-ledger.mjs --check` | pass |
| `go run ./cmd/connectorgen certification-matrix --check` | pass |
| `go run ./cmd/connectorgen certification-candidates --connector github --check` | pass |
| `go run ./cmd/connectorgen certification-sweep --connector github --check` | pass |
| `go run ./cmd/connectorgen boundary . --json` (captured to a temporary log and polled to completion) | pass |
| `bash scripts/tests/connector-canon.sh` | pass |
| `./scripts/tests/pinned-build-dependencies.sh` | pass |
| `./scripts/tests/homebrew-release-notify.sh` | pass |
| `./scripts/tests/release-target-parity.sh` | pass |

`go test -timeout 20m ./...` and aggregate `make verify` were deliberately not run as single commands: the repository AGENTS instruction says agents under a per-command timeout must run changed packages plus `internal/cli` separately and execute `make verify`'s non-suite gates individually, because the full 550+ connector suite is routinely cut off and indistinguishable from a hang. The targeted package tests and each applicable gate above were run; CI retains the full suite.
