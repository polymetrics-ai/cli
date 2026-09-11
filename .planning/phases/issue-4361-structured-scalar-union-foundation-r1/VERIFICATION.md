# Verification checklist — issue 4361

- [x] Focused red test result recorded in `TDD-LEDGER.md`.
- [x] Focused engine tests green with `-timeout 20m`.
- [x] Focused commandrunner tests green with `-timeout 20m`.
- [x] Full changed `cmd/connectorgen` package tests green with `-timeout 20m` (181.143s after the projection repair).
- [x] `go test -timeout 20m ./internal/connectors/engine`, `./internal/connectors/commandrunner`, and `./internal/cli` are green; CLI completed in 426.480s.
- [x] `go build ./cmd/pm`, `go vet ./...`, `make tidy-check`, `make lint`, `make docs-check-no-build`, and `make smoke-no-build` are green.
- [x] Source/definition and generated checks are green: `connectorgen-validate`, `surface-sync`, `declaration-admission`, `operation-evidence`, GitHub parity artifacts, certification checks, connector canon, and release-workflow checks.
- [x] Current `origin/main` was fetched at `2165619ec8f5f9d4141b491b7a5a64bc460d0c71` and merged normally; it was already an ancestor of the branch.
- [x] `git diff --check` is green.
- [x] Credential-free built-binary census records a zero usable-surface delta: Twilio has 94 declared actions and Xero 87, but neither manifest has a command surface. No cited Twilio/Xero command was fabricated.
- [ ] Fresh-context independent Codex exact-head audit complete; every blocker fixed or explicitly absent. (Runs after PR open and head freeze, per firstmate instruction 003.)
- [ ] PR base verified from the GitHub API-equivalent response after opening.
