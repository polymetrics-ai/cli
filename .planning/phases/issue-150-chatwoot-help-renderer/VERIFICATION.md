# Verification: Chatwoot Help Renderer And Docs Parity

## Completed targeted gates

```bash
./pm help docs
```

Result: pass; docs command help reviewed before generation/validation.

```bash
./pm connectors inspect chatwoot | grep -E 'COMMAND SURFACE|Usage: pm chatwoot|conversation list|message send'
grep -E 'COMMAND SURFACE|Usage: pm chatwoot|conversation list|message send' docs/connectors/chatwoot/MANUAL.md
```

Result: pass; runtime manual and checked-in Chatwoot manual expose the command surface.

```bash
go test ./internal/connectors -run Guide -count=1
go test ./internal/connectors/bundleregistry -run 'GitHubGuide|ChatwootGuide' -count=1
go test ./internal/connectors/bundleregistry -run ChatwootGuide -count=1
```

Result: pass.

```bash
( cd website && npm run test:unit -- tests/api/connector-data.test.ts )
( cd website && npm run typecheck )
( cd website && npm run build )
```

Result: pass.

```bash
./pm docs validate --connectors-dir docs/connectors
go run ./cmd/connectorgen validate internal/connectors/defs
git diff --check
```

Result: pass.

## Package-manager note

`pnpm test:unit` could not run before dependencies were installed, and `pnpm install --frozen-lockfile` is currently blocked by `ERR_PNPM_LOCKFILE_CONFIG_MISMATCH`. Used `npm ci` against the checked-in `website/package-lock.json` for local website tests; no dependency files changed.

## Completed full handoff gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Result: pass. `make verify` completed `go test -timeout 20m ./...`, `go build ./cmd/pm`, connector docs validation, smoke test, `golangci-lint`, and `go run ./cmd/connectorgen validate internal/connectors/defs`.
