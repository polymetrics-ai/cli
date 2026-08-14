# VERIFICATION — issue #3865 verified-auth cohort fencing

Status: focused implementation verification passed; broader local gates and review remain.

- [x] Typed verified-auth result is the sole fence trigger.
- [x] A verified fence cancels same-cohort active contexts and rejects all later admissions before a fake sender increments.
- [x] An unrelated opaque cohort remains admitted.
- [x] Verified repair/test increases the epoch, retains the last fenced epoch as audit evidence, refuses stale members, and permits the new healthy epoch.
- [x] Restart/race proof observes zero post-fence admissions and sends under `-race`.
- [x] Formatter, vet/build, and applicable individual repository gates pass.
- [x] Inline GSD verify-work and code-review evidence is recorded.
- [ ] PR targets `integration/4015-mvp-flat-r1`; the GitHub API reports that exact base.

## Automated evidence

- Focused behavior: `go test -count=1 -run '^TestAuthCohortCoordinator' ./internal/coordination` passed.
- Required race proof: `go test -race -count=1 -timeout 20m ./internal/coordination/...` passed, including the restart/race test with 64 post-fence attempts and an exact zero-send assertion.
- Static/build: `gofmt` on changed Go files, `go vet ./internal/coordination/...`, `golangci-lint run ./internal/coordination/...`, and `go build ./cmd/pm` passed.
- Repository gates: `make tidy-check`, `make lint`, `make agent-contract-check`, `make docs-check`, `make smoke-no-build`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check` passed.
- CLI help/manual/website parity: not applicable; this adds no command, flag, help, documentation surface, or connector declaration.
