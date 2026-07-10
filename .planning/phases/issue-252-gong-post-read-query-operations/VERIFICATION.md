# Verification — issue #252 Gong typed POST read-query operation execution

- [x] `go test ./internal/connectors/engine -run 'OperationDirectRead|DirectRead' -count=1`
- [x] `go test ./internal/connectors/commandrunner -run 'OperationDirectRead|DirectRead|RedactRecord' -count=1`
- [x] `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [x] CLI/help/docs parity: Gong manual/skill regenerated/updated, website connector bundle regenerated, `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` passed, and `go run ./cmd/pm connectors inspect gong --json` inspected metadata without credentials.
- [x] Full gates covering this slice: `go test -timeout 20m ./...`, `go vet ./...`, `go build ./cmd/pm`, `make verify`.
