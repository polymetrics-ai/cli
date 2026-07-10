# Verification — issue-146-gong-advanced-query-binary-engine

- [x] `go test ./internal/connectors/engine -run DirectRead -count=1`
- [x] `go test ./internal/connectors/commandrunner -count=1`
- [x] `go test ./cmd/connectorgen -run Gong -count=1`
- [x] `go test ./cmd/connectorgen -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [x] `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs`
- [x] Docs/help parity: `pm help docs` reviewed; `pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` passed; Gong connector manual/skill and website generated catalog updated.

## 2026-07-10 engine-support implementation verification (#252/#253/#254)

- [x] `go test ./internal/connectors/connsdk -run Multipart -count=1`
- [x] `go test ./internal/connectors/engine -run 'OperationDirectRead|WriteJSONArray|WriteMultipart|DirectRead|Write' -count=1`
- [x] `go test ./internal/connectors/commandrunner -run 'OperationDirectRead|DirectRead|RedactRecord' -count=1`
- [x] `go test ./internal/app -run PayloadIdentities -count=1`
- [x] `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [x] `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs`
- [x] `go test -timeout 20m ./...`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make verify`
- [x] CLI/help/docs parity: Gong connector manual/skill updated, website connector bundle regenerated, `go run ./cmd/pm connectors inspect gong --json` inspected metadata without credentials.
