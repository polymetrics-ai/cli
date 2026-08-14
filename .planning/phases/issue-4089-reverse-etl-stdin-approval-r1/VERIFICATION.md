# #4089 — verification checklist

## Lifecycle

- [x] Issue #4089 and parent #3988 read.
- [x] GSD adapter doctor, command-source resolution, prompt generation, and agent contract check passed.
- [x] Required Go CLI, security, testing, documentation, and website skills loaded; inline/manual GSD fallback recorded.
- [x] Existing bounded stdin carrier and both current argv call sites verified.

## Before completion

- [x] Red test recorded before production edit.
- [x] Both command paths require a bare stdin marker and use the shared carrier.
- [x] Empty, oversized, malformed, retired-argv, and replay inputs reject before a write-side effect.
- [x] Six independent secret-surface checks pass: argv, environment, files, logs, receipts, and evidence.
- [x] Plan → preview → run and exact-once replay behavior remain proven.
- [x] Runtime help, CLI manual, generated skills, website source/data, and CLI transcript are updated.
- [x] Repository stale-syntax scan finds no retired argv approval examples in active source, docs, website, generated data, or tests.
- [x] Targeted package tests, vet, docs, connector boundary, and split local gates are green.
- [x] Inline `verify-work` and `code-review` records contain no unresolved finding.

## Commands and result

```text
PASS  go test -timeout 20m ./internal/cli -count=1
PASS  go test -timeout 20m ./internal/app -count=1
PASS  go test -timeout 20m ./internal/connectors/certify -count=1
PASS  go test -timeout 20m ./internal/safety -count=1
PASS  focused cmd/connectorgen coverage check
PASS  go vet ./internal/cli ./internal/app ./internal/connectors/certify ./internal/safety ./cmd/connectorgen
PASS  go build ./cmd/pm
PASS  pm help reverse; pm reverse; pm reverse --help
PASS  docs and website generators; website blog catalog test
PASS  make tidy-check, docs-check, agent-contract-check, connectorgen-validate,
      connectorgen-surface-sync, connector-boundary, release-workflow-check,
      lint, smoke-no-build
PASS  rg --pcre2 -- '--approve(?=\\s|=|$)' active source/docs/website (only intentional retired-flag regression assertions remain)
```

## Rebased CI continuation

```text
PASS  rebased cleanly onto origin/integration/4015-mvp-flat-r1
PASS  GOTOOLCHAIN=go1.25.13 go run golang.org/x/vuln/cmd/govulncheck@latest ./...
PASS  GOTOOLCHAIN=go1.25.13 go test -count=1 -timeout 20m ./internal/app
PASS  GOTOOLCHAIN=go1.25.13 go test -count=1 -timeout 20m ./internal/cli \
      -run '^TestReverseETLApprovalUsesBoundedStdin$'
```
