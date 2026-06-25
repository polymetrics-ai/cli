# PLAN — GitHub Native Package + Data-Driven Registry

Tasks are checkbox-tracked. Behavior tasks require red-first evidence in TDD-LEDGER.md (TDD gate).
Test tasks (`type: test`) author the red tests.

## Wave A — Registry self-registration mechanism
- [ ] id: t-registry-factory type: test — Red test: a factory registered via `RegisterFactory` is resolvable from `NewRegistry().Get(...)`.
- [ ] id: b-registry-factory type: behavior — Add `RegisterFactory(name, func() Connector)` + ordered factory store in package connectors; `NewRegistry()` consumes registered factories alongside built-ins and the existing enabled-catalog-alias loop. No behavior change for current connectors.

## Wave B — GitHub package migration
- [ ] id: t-github-package type: test — Red test in `internal/connectors/github/` asserting `github.New().Name()=="github"`, capabilities Read+Write, catalog streams, httptest-backed read, and write-action validate accept/reject.
- [ ] id: b-github-package type: behavior — Move github.go/github_auth.go/github_streams.go into `internal/connectors/github/` (package github) exporting `New() connectors.Connector`; translate `connectors.*` type references; relocate GitHub-only helpers; keep shared helpers in connectors. Adopt connsdk where behavior-neutral.
- [ ] id: b-registry-wiring type: behavior — `init()` calls `connectors.RegisterFactory("github", New)`; add `registry_gen.go` blank-importing the github package; remove `r.Register(Github{})` from `NewRegistry()`; delete the old package-connectors github*.go once green; move/adapt the old github_test.go.

## Wave C — Verification
- [ ] id: t-parity type: test — Parity/regression: `make verify` green; `NewRegistry().Get("github")` and `.Get("source-github")` both resolve with Read+Write; `pm connectors inspect github --json` → kind "Connector".

## Ordering / dependencies
A → B → C. Wave A lands first (small, safe). Wave B's heavy file-move is delegated to a backend
subagent. Wave C is the deterministic gate.

## Rollback
Pure refactor; revert phase commits. No data/schema/dependency changes. See RUNBOOK.md.
