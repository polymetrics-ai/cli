# Verification checklist — request-contract execution envelopes

## Behavior and safety

- [ ] Gong-shaped valid source imports through the real importer path.
- [ ] Exact provider schema is retained with no synthetic `maxLength`.
- [ ] Every executable common input has a positive, versioned PM envelope.
- [ ] Common missing string/numeric/array bounds are not gap spam.
- [ ] Dynamic/composition/untyped serialization gaps remain merge-blocking.
- [ ] Runtime rejects over-cap values before auth/network and never truncates.
- [ ] Error names PM execution policy and exact measurement unit.
- [ ] Exact numeric lexemes reach the wire unchanged.
- [ ] Existing command effective limits are not silently changed.
- [ ] No connector-specific code or connector definition edit.

## TDD and local gates

- [ ] Red importer/envelope trace recorded before production edits.
- [ ] Green focused `cmd/connectorgen` tests with `GOFLAGS=-p=3`.
- [ ] Green focused engine tests with `GOFLAGS=-p=3`.
- [ ] Deliberate sabotage makes the new bounding tests fail; restored green.
- [ ] `gofmt` on changed Go files.
- [ ] `go vet ./...` with bounded concurrency.
- [ ] `go build ./cmd/connectorgen` and `go build ./cmd/pm`.
- [ ] Serialized `go test -timeout 20m ./cmd/connectorgen`.
- [ ] Serialized `go test -timeout 20m ./internal/cli` if CLI output changes.
- [ ] Applicable individual `make verify` gates from AGENTS.md.
- [ ] `git diff --check origin/main...HEAD`.

## Lifecycle and delivery

- [ ] Execute-phase prompt executed inline and evidenced.
- [ ] Verify-work prompt executed inline; gaps planned/executed if found.
- [ ] `golang-lint` loaded; code-review prompt executed inline.
- [ ] PR body records issue, Red/Green evidence, skills, gates, review status,
      and accepted no-mistakes record or explicit task prohibition.
- [ ] Pull request opened against `main`; API-reported base verified `main`.
- [ ] Required GitHub checks and automated review coverage complete.

## CLI help/manual/docs/website parity

- [ ] Effective PM cap visible in connector help/inspection fixture.
- [ ] PM-labelled runtime error covered.
- [ ] `pm help <topic>`, bare namespace, and affected `--help` checked.
- [ ] Docs/website/generated artifacts either updated together or explicitly
      proved unchanged/not applicable.
