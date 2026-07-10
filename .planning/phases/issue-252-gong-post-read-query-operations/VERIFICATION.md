# Verification — issue #252 Gong typed POST read-query operation execution

- [ ] `go test ./internal/connectors/engine -run 'OperationDirectRead|DirectRead' -count=1`
- [ ] `go test ./internal/connectors/commandrunner -run OperationDirectRead -count=1`
- [ ] `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [ ] CLI/help/docs parity for any command availability flip.
