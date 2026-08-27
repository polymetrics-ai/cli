# Issue #4359 — Verification checklist

## Targeted proof

- [ ] `go test -timeout 20m ./internal/connectors/engine -run 'CompositeProviderPath|CommandBinding' -count=1`
- [ ] `go test -timeout 20m ./internal/connectors/commandrunner -run 'CircleCI|ImplementedCommand' -count=1`
- [ ] `go test -timeout 20m ./internal/app -run 'CircleCI|TransportPreflight' -count=1`
- [ ] `go test -timeout 20m ./internal/cli -run 'CircleCI|DeclarationAdmission' -count=1`
- [ ] `go build -o ./bin/pm ./cmd/pm`
- [ ] Eleven fresh-project, credential-free `pm circleci … --json` runs reach exactly `missing --credential` after valid declared fixture input; no provider request.

## Generated/source checks

- [ ] `go test -timeout 20m ./cmd/connectorgen -count=1`
- [ ] `go run ./cmd/connectorgen source-import circleci --check` (read-only source lock once Batch-1 is present; otherwise record its unavailable-on-base reason)
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check`
- [ ] `go run ./cmd/connectorgen declaration-admission internal/connectors/defs`
- [ ] Operation evidence check/generator applicable to current main.

## Repository checks

- [ ] `gofmt -w` on changed Go files
- [ ] `go vet ./...`
- [ ] `make tidy-check`
- [ ] `make lint`
- [ ] `make docs-check-no-build`
- [ ] `make smoke-no-build`
- [ ] `make agent-contract-check`
- [ ] `make connector-boundary`
- [ ] `make release-workflow-check`
- [ ] Website type/docs check as applicable
- [ ] `git diff --check`

## Review and delivery

- [ ] Fresh-context Codex audit replaces unavailable Claude audit and records exact immutable source SHA.
- [ ] Fresh-context Codex re-review records final code SHA and all six lane results.
- [ ] PR base API response is exactly `main`.
