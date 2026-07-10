# Verification — issue #254 Gong bounded typed multipart uploads

- [ ] `go test ./internal/connectors/connsdk -run Multipart -count=1`
- [ ] `go test ./internal/connectors/engine -run 'WriteMultipart|Write' -count=1`
- [ ] `go test ./internal/connectors/commandrunner -run Multipart -count=1`
- [ ] `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [ ] CLI/help/docs parity for any command availability flip.
