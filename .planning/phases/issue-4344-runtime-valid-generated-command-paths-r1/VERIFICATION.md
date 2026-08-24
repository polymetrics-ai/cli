# Verification checklist — #4344

- [ ] `GOFLAGS='-p=3' go test -count=1 -timeout 20m ./cmd/connectorgen`
- [ ] Targeted commandrunner / CLI behavioral tests with `-timeout 20m`
- [ ] Deliberate raw-source-ID reintroduction fails the new test, then restore
- [ ] `GOFLAGS='-p=3' go run ./cmd/connectorgen surface-sync --check`
- [ ] `GOFLAGS='-p=3' go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `GOFLAGS='-p=3' go run ./cmd/connectorgen operation-evidence --check`
- [ ] `make connector-runtime-preflight`, `make connector-canon-check`, `make connector-boundary`
- [ ] `GOFLAGS='-p=3' go build ./cmd/pm`; isolated credential-free command sweep
- [ ] `./pm docs validate --connectors-dir docs/connectors` and applicable runtime help/docs parity checks
- [ ] `gofmt`, `go vet`, lint, remaining safe `make verify` gates; record any environment block verbatim
- [ ] Inline verify-work and code-review evidence; PR API base read-back
