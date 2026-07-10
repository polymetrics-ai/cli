# Verification — issue #253 Gong top-level JSON array request bodies

- [ ] `go test ./internal/connectors/engine -run 'WriteJSONArray|Write' -count=1`
- [ ] `go test ./internal/connectors/commandrunner -run JSONArray -count=1`
- [ ] `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [ ] CLI/help/docs parity for any command availability flip.
