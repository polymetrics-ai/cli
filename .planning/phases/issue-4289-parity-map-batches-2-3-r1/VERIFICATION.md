# Issue #4289 — Verification Checklist

- [x] Every listed bundle has a parsing source lock and a corrected declaration-disposition ledger.
- [x] The source-lock/map-integrity assertion proves all nineteen source inventories, API-surface bindings, row shapes, six-class totals, and rejected-operation reasons: `node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-parity-maps.mjs` → `verified 19 connectors / 3340 documented operations`.
- [x] Every documented DELETE is declared or carries an explicit disabled disposition; the generated report records its enabled/documented count per connector.
- [x] `go run ./cmd/connectorgen validate --json` → `connectors_checked: 552`, zero findings and warnings.
- [x] `go run ./cmd/connectorgen surface-sync --check` → 552 scanned, zero fields filled/corrected.
- [x] Targeted fixture/conformance test: `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/(grafana|trello|slack|n8n|google-calendar|gmail|twilio|amazon-sqs|elasticsearch|gong|google-ads|facebook-marketing|linkedin-ads|aircall|xero|paypal-transaction|gocardless|amazon-seller-partner|miro)$' -count=1` → PASS. Runtime preflight: `go test -timeout 20m -run '^TestEveryImplementedCommandPassesRuntimePreflight$' ./internal/connectors/commandrunner -count=1` → PASS.
- [x] Repository generated/snapshot checks: `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connectorgen-certification-candidates`, and `make connectorgen-certification-sweep` → PASS.
- [x] Detached `go run ./cmd/connectorgen boundary . --json` → clean; 292 files / 552 connectors / zero findings and warnings. Captured in `traces/connector-boundary.stdout` (stderr empty).
- [x] `go build ./cmd/pm`, `go vet ./...`, `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connector-canon-check`, and `make release-workflow-check` → PASS.
- [x] Inline `verify-work` and data-focused code review found no actionable map defects; see `REVIEW.md`.
- [x] No provider credential, live provider call, provider write, or live-certification claim appears in the diff or evidence. Public provider documentation was read only during source capture.
- [ ] PR uses `Refs #4289` and its GitHub API base is read back as `main` (pending push/open).

The full `go test -timeout 20m ./...` suite was intentionally left to CI: the repository instruction documents that the 550+ connector suite routinely exceeds this worker's per-command window and requires scoped local checks. The changed definition paths are covered by the targeted conformance run above, and the production preflight test sweeps every implemented command.
