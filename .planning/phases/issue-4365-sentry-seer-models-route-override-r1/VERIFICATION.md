# Issue #4365 verification checklist

- [x] `go test -timeout 20m ./internal/connectors/engine -run '^TestSentrySeerModels' -count=1` (happy/bad/edge route tests green)
- [x] `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`
- [x] `go test -timeout 20m ./internal/cli -run '^TestSentrySeerModels(CommandStopsBeforeProviderIOWithoutCredential|HelpAndBareNamespaces)$' -count=1` (Sentry help/namespace/credential-spy tests green)
- [x] Root help golden transcript variants regenerated and focused `TestGoldenTranscripts` rerun green
- [x] `go test -timeout 20m ./internal/cli -count=1` — green (`453.157s`)
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [x] `go run ./cmd/connectorgen surface-sync` then `--check`
- [x] `go run ./cmd/connectorgen declaration-admission internal/connectors/defs --json` (existing one-connector source cohort remains green)
- [x] `go run ./cmd/connectorgen operation-evidence . --check`
- [x] `go run ./cmd/connectorgen surface-reconcile internal/connectors/defs/sentry --check --json` (no pending Sentry operation rows; global check has 3,844 pre-existing unrelated reclassifications)
- [x] Whole-tree `connector-boundary` report clean; branch-relative boundary will be rerun after the implementation commit so the base-diff includes tracked changes
- [x] generated connector docs/manual command and `docs-check`
- [x] `pm help sentry`, `pm sentry`, `pm sentry seer`, and `pm sentry seer list-models --help`
- [x] built-binary missing-credential/zero-transport proof in an isolated initialized project
- [x] `gofmt`, `go vet ./...`, scoped package tests, `go build ./cmd/pm`, tidy/lint/smoke/release/docs/contract/connector/certification gates, website typecheck/lint/build, and `git diff --check`
- [x] Inline GSD `execute-phase`, `verify-work`, and `code-review` fallback records (Pi cannot provide the required isolated workers); independent exact-head audit will be requested on the PR
