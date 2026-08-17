# Verification — generic live certification runner

## Planned checks

- `node --check scripts/certify-connector-live.mjs`
- `node scripts/certify-connector-live.mjs <connector> --definition-check`
- `node scripts/certify-connector-live.mjs <authorized-connector> --credential-env <environment-name>`
- `node scripts/certify-connector-live.mjs <different-connector> --credential-env <environment-name>`
- `go run ./cmd/connectorgen certification-matrix --check` after every accepted record and once after the run
- `git diff --check`

The full Go suite and `make verify` are deliberately deferred until the end of the live run, as authorized by the captain; this task has no Go production changes.

## Executed checks

- `node --check scripts/certify-connector-live.mjs` — passed.
- `node scripts/certify-connector-live.mjs github --definition-check` — passed: `commands=1571 candidates=122 eligible=122 credential_configs=2`.
- Definition-only invocation for every one of the 36 command-surface connectors — passed.
- `node scripts/certify-connector-live.mjs freshchat ...` — passed unchanged: `executed=0 certified=0 provider_refused=0 missing_fixture=1 product_defect=0`.
- Authorized live run: `executed=122 certified=38 provider_refused=80 missing_fixture=4 product_defect=0`; every accepted record was immediately checked by `go run ./cmd/connectorgen certification-matrix --check`, which also passed after the run.
- `go test -timeout 20m ./cmd/connectorgen` — passed (82.667s).
- `go test -timeout 20m ./internal/connectors/commandrunner` — passed (12.541s).
- `go test -timeout 20m ./internal/cli` — passed (511.153s).
- End-of-run repository gates passed: `make fmt tidy-check vet build docs-check-no-build smoke-no-build lint agent-contract-check connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connectorgen-certification-candidates connectorgen-certification-sweep connector-boundary connector-canon-check release-workflow-check`.
- `git diff --check` — pending immediately before commit.
- Post-PR App retry: selected 16 captain-probed App-200 rows, executed all 16, certified 15, and retained one HTTP 400 product-defect receipt. Final `certification-matrix --check` passed with 53 GitHub accepted records on disk.
