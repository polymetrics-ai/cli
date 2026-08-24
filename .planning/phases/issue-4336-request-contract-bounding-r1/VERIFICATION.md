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
- [x] `go vet ./...` with bounded concurrency.
- [x] `go build ./cmd/connectorgen` and `go build ./cmd/pm`.
- [x] Serialized `go test -timeout 20m ./cmd/connectorgen`.
- [x] Focused generated CLI/help/inspection parity tests. The full
      `internal/cli` suite's two concurrent fresh-child build cases failed only
      under suite concurrency and each passed when rerun serialized; this is
      recorded as an environmental full-suite limitation, not a package pass.
- [x] After merging final `main`, reproduced the CI generated-skills drift,
      regenerated only `docs/skills/pm-twenty/SKILL.md`, proved a second
      generator pass byte-stable, and passed
      `TestSkillsGenerateMatchesTrackedSkills` in 4.376s.
- [x] After Firstmate's subsequent merge of `main` at `60f85ae94`, the exact
      tracked-skills test passed in 4.942s, `surface-sync --check` reported zero
      drift across 553 connectors, and full `cmd/connectorgen` passed in
      183.024s with `GOFLAGS=-p=2`.
- [x] Applicable individual `make verify` gates from AGENTS.md.
- [x] `git diff --check origin/main...HEAD`.

## Lifecycle and delivery

- [x] Execute-phase prompt executed inline and evidenced.
- [x] Verify-work prompt executed inline; no deliverable gap remained.
- [x] `golang-lint` loaded; code-review prompt executed inline and three
      warning findings fixed with red/green tests.
- [x] PR body records issue, Red/Green evidence, skills, gates, review status,
      and accepted no-mistakes record or explicit task prohibition.
- [x] Pull request #4345 opened against `main`; API-reported base verified
      `main` and the branch was updated through `60f85ae94` as directed.
- [ ] Required GitHub checks and automated review coverage complete.

## CLI help/manual/docs/website parity

- [x] Effective PM cap visible in connector help/inspection fixture.
- [x] PM-labelled runtime error covered.
- [x] `pm help gong`, bare `pm gong`, and `pm gong calls get --help` checked.
- [x] Generated connector manuals/skills and CLI docs updated; no command or
      flag set changed and the branch contains no `website/**` diff, so website
      projection is not applicable to this additive policy-provenance text.
