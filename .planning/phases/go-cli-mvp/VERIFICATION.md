# Verification

## Commands

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/poly
make smoke
make verify
```

## Result

All commands passed.

`make verify` ran:

- format
- vet
- tests
- build
- end-to-end smoke flow

The smoke flow verified:

- project init
- credential creation from environment variables
- connection creation
- catalog refresh
- ETL from `sample.customers` into local warehouse
- query of local warehouse rows
- reverse ETL plan generation
- approval-token execution
- outbox write creation

## Residual Risks

- No CI configuration yet.
- No SQLite, DuckDB, OS keychain, Parquet, or real SaaS connector integration yet.
- Credential vault uses a local generated key file for the MVP. Production should move to OS keychain or passphrase/KMS-backed envelope encryption.

