# Verification — issue-146-gong-advanced-query-binary-engine

- [x] `go test ./internal/connectors/engine -run DirectRead -count=1`
- [x] `go test ./internal/connectors/commandrunner -count=1`
- [x] `go test ./cmd/connectorgen -run Gong -count=1`
- [x] `go test ./cmd/connectorgen -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [x] `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs`
- [x] Docs/help parity: `pm help docs` reviewed; `pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` passed; Gong connector manual/skill and website generated catalog updated.
