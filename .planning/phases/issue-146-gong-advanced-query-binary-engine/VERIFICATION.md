# Verification — issue-146-gong-advanced-query-binary-engine

- [x] `go test ./internal/connectors/engine -run DirectRead -count=1`
- [x] `go test ./internal/connectors/commandrunner -count=1`
- [x] `go test ./cmd/connectorgen -run Gong -count=1`
- [x] `go test ./cmd/connectorgen -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [x] `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs`
- [x] Docs/help parity: `pm help docs` reviewed; `pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs` passed; Gong connector manual/skill and website generated catalog updated.

## 2026-07-10 engine-support planning verification (not yet run)

Planned gates for the future implementation slice:

- [ ] `go test ./internal/connectors/connsdk -run Multipart -count=1`
- [ ] `go test ./internal/connectors/engine -run 'OperationDirectRead|WriteJSONArray|WriteMultipart|DirectRead' -count=1`
- [ ] `go test ./internal/connectors/commandrunner -run 'OperationDirectRead|Multipart|JSONArray' -count=1`
- [ ] `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`
- [ ] `go test -timeout 20m ./...`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] CLI/help/docs parity for any command availability flip: `pm help gong`, `pm connector gong <namespace> --help`, connector docs/manual/website generated data.
