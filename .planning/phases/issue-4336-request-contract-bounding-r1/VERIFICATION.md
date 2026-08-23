# Verification checklist — request-contract execution envelopes

## Behavior and safety

- [x] Gong-shaped valid source imports through the real importer path.
- [x] Exact provider schema is retained with no synthetic `maxLength`.
- [x] Every executable common input has a positive, versioned PM envelope.
- [x] Common missing string/numeric/array bounds are not gap spam.
- [x] Dynamic/composition/untyped serialization gaps remain merge-blocking.
- [x] Runtime rejects over-cap values before auth/network and never truncates.
- [x] Error names PM execution policy and exact measurement unit.
- [x] Exact numeric lexemes reach the wire unchanged.
- [x] Existing command effective limits are not silently changed.
- [x] No connector-specific code or connector definition edit.

## TDD and local gates

- [x] Red importer/envelope trace recorded before production edits.
- [x] Green focused `cmd/connectorgen` tests with `GOFLAGS=-p=3`.
- [x] Green focused engine tests with `GOFLAGS=-p=3`.
- [x] Deliberate sabotage makes the new bounding tests fail; restored green.
- [x] `gofmt` on changed Go files.
- [ ] `go vet ./...` with bounded concurrency.
- [ ] `go build ./cmd/connectorgen` and `go build ./cmd/pm`.
- [ ] Serialized `go test -timeout 20m ./cmd/connectorgen`.
- [ ] Serialized `go test -timeout 20m ./internal/cli` if CLI output changes.
- [ ] Applicable individual `make verify` gates from AGENTS.md.
- [ ] `git diff --check origin/main...HEAD`.

## Lifecycle and delivery

- [x] Execute-phase prompt executed inline and evidenced.
- [ ] Verify-work prompt executed inline; gaps planned/executed if found.
- [ ] `golang-lint` loaded; code-review prompt executed inline.
- [ ] PR body records issue, Red/Green evidence, skills, gates, review status,
      and accepted no-mistakes record or explicit task prohibition.
- [ ] Pull request opened against `main`; API-reported base verified `main`.
- [ ] Required GitHub checks and automated review coverage complete.

## CLI help/manual/docs/website parity

- [x] Effective PM cap visible in connector help/inspection fixture.
- [x] PM-labelled runtime error covered.
- [ ] `pm help <topic>`, bare namespace, and affected `--help` checked.
- [ ] Docs/website/generated artifacts either updated together or explicitly
      proved unchanged/not applicable.
