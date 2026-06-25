# TDD Ledger

## Red tests before implementation

- `internal/vault/vault_test.go`: credentials are encrypted at rest and decrypted by field.
- `internal/app/app_test.go`: end-to-end local ETL and reverse ETL app workflow.
- `internal/cli/cli_test.go`: CLI help and JSON-oriented command behavior.

## Evidence

Initial tests are intentionally added before implementation and expected to fail until code is written.

