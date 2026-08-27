# Verification checklist — issue 4371

## Required local gates

- [ ] Focused cited-only non-executable/partial red then green tests.
- [ ] Ordinary OpenAPI/Swagger disposition stability and invalid citation matrix.
- [ ] Salesloft/Copper source import, source projection, operation evidence,
  validate, and surface-sync checks; or exact record that clean-worktree source
  locks are unavailable and no cohort output was regenerated.
- [ ] Real registry/commandrunner missing-foundation-before-credential/I/O
  proof for any generated unavailable command under test.
- [ ] `go test -timeout 20m ./cmd/connectorgen -count=1`.
- [ ] Changed dependent engine/commandrunner suites with `-timeout 20m`.
- [ ] `go vet ./...`, `go build ./cmd/connectorgen`, `go build ./cmd/pm`, and
  the individually applicable tidy/lint/docs/smoke/agent-contract/
  connectorgen/connector-boundary/release-workflow gates.
- [ ] `git diff --check` before commit and after current-main rebase.
- [ ] Inline GSD `verify-work`, `code-review`, and a separate exact-head Codex
  audit request/result recorded against the final code SHA.
- [ ] PR base API read-back equals `main`; CI/review are requested but no merge
  is performed.

## Results

Pending implementation.
