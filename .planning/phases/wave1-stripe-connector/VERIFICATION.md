# Local Verification

- Verified by: orchestrator (independent of backend agent) — ran the gate + read the security surface.

| Check | Status | Command | Result |
| --- | --- | --- | --- |
| Format | passed | `gofmt -l internal` | clean |
| Static analysis | passed | `go vet ./...` | no findings |
| Stripe tests | passed | `go test ./internal/connectors/stripe/` | ok (read+pagination+auth, write-validate, registry) |
| Full suite | passed | `go test ./...` | 11 packages ok, 0 failures |
| connsdk addition | passed | `go test ./internal/connectors/connsdk/` | DoForm + shared-core tests green; JSON Do unchanged |
| Full gate | passed | `make verify` | exit 0; smoke ok |
| Catalog integrity | passed | jq + `native_conformance_test` | enabled=2 (github+stripe), planned=645, total/conformance length 647 |

## Parity (built binary)
- `pm connectors inspect stripe --json` → kind Connector, read+write.
- `pm connectors inspect source-stripe --json` → kind Connector, read+write (alias → live stripe).

## Modified existing tests — exact, not weakened
Enabling `source-stripe` changed counts/identity that pre-existing tests pinned. Updates verified exact:
- `catalog_test.go`: `Enabled==2 && PlannedNativePort==645` (was 1/646; total 647 unchanged).
- `catalog_cli_test.go`: asserts `"enabled": 2`, `"planned_native_port": 645`, plus NEW explicit
  `source-stripe` + `pm_connector_name: stripe` checks (stronger).
- `cli_test.go`: "not-found / not-yet-enabled" example swapped from `source-stripe` to the still-planned
  `source-strava` (valid; stripe is now live). Red-first stripe assertions untouched.

## Orchestrator review (VERIFY) — GO
- Secret safety: `client_secret` only feeds `connsdk.Bearer`; never logged/printed; errors carry no
  secret value. Manifest documents "never logged".
- SSRF: `base_url` validated (parse + http/https scheme + non-empty host); default api.stripe.com.
- Write allow-list: `{create_customer, update_customer}`; `ValidateWrite`/`resolveWriteAction` reject
  unknown actions; reverse-ETL stays plan→preview→approve→execute; fixture mode does no external call.
- connsdk `DoForm` added by extracting a shared `do` core — JSON `Do` behavior unchanged, +2 tests.

## TDD
3 behavior tasks (b-stripe-read, b-stripe-write, b-stripe-catalog) red-confirmed before code; now green.
