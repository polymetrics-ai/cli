# Test Plan

- Unit: verify every catalog entry has runtime capabilities.
- Unit: verify enabled GitHub capabilities are read/write/ETL/reverse ETL capable.
- Unit: verify planned connectors are metadata-only with an unsupported reason.
- CLI: verify catalog JSON contains runtime capabilities.
- CLI: verify catalog-only manual renders runtime capability details.
- Docs: run docs generation and validation.
- Local verification: run `go test ./...`, `go vet ./...`, `go build -o pm ./cmd/pm`, and `make verify`.
