# PLAN — Pi clean project-only workers

Issue: #3715. Parent: #3714. Branch: `fm/cli-agents-wave-pi-r1`.

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd list` and `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- `scripts/gsd prompt programming-loop ...`: unavailable, matching Wave 1's recorded adapter
  behavior. Inline/manual fallback is used and recorded.
- Generated and executed inline: `discuss-phase issue-3715-pi-project-workers --auto` and
  `plan-phase issue-3715-pi-project-workers --tdd --skip-research`.
- Generated `execute-phase`, `verify-work`, and `code-review` prompts were executed inline. The
  adapter cannot run its phase-number-only workflow against this issue-named phase, and this task
  forbids extra worker roles; `SUMMARY.md`, `UAT.md`, `VERIFICATION.md`, and `REVIEW.md` record the
  manual fallback, automated coverage, and review disposition.

## Required skills loaded

- `gsd-programming-loop` — TDD lifecycle and phase artifacts.
- `no-mistakes` v1.41.2 — child-local gate ownership and stacked-PR workaround.
- `golang-how-to`, `golang-cli`, `golang-code-style`, `golang-naming`, `golang-testing`,
  `golang-error-handling`, `golang-safety`, and `golang-security` — generator/checker change,
  test-first behavior, bounded paths, and child-tool safety.

## TDD slices

### Slice 1 — demonstrate ambient discovery and preserve child safety

1. RED: add a Pi extension executable test that loads the actual discovery module with a hostile
   user-agent fixture, the retained legacy project roles, and the existing bundled directory; it
   must fail because the requested clean roster includes ambient agents.
2. RED: cover that a child tool request containing `subagent` is stripped and that depth `1` is
   rejected before any subprocess spawn.
3. GREEN: introduce a canonical-derived `clean-project` scope; no bundled/user/legacy role may
   appear in its discovered roster, while the existing bounded tool/depth properties pass.
4. REFACTOR: keep `user`, `project`, and `both` aligned with the official example's semantics,
   remove bundled-role guidance, and retain project confirmation behavior for the new scope.

### Slice 2 — canonical full-file Pi projections and drift check

1. RED: add Go tests that make Pi projections required and assert missing projections fail; assert
   generator sync creates exact canonical Markdown/YAML files and that any whole-file drift fails.
2. GREEN: extend the canonical schema and generator to encode the clean mode and bounded Pi
   frontmatter, render both full files, mark both Pi projections required, and run sync.
3. REFACTOR: validate the exact role pair/tool allowlist in the canonical contract and preserve
   existing marked-block behavior for future Claude/Codex projections.

### Slice 3 — adapter documentation and verification

1. Update the Pi extension README to state official discovery format, scope precedence, clean-mode
   guarantee, tool allowlist behavior, project trust/confirmation, and the precise delegation
   boundary.
2. Run focused Go and Pi extension tests, generated-file drift check, formatting/static checks, and
   the repository's applicable individual verify gates.
3. Run inline GSD verify/review prompts, record finding dispositions, then commit the green slice.

## Verification plan

- `scripts/tests/pi-clean-project-agents.sh` — runtime red/green evidence using the installed Pi
  dependency and a hostile global-agent fixture; records no credentials or global state.
- `go test ./internal/agentcontract ./cmd/agentcontractgen`.
- `go run ./cmd/agentcontractgen sync` then `go run ./cmd/agentcontractgen check`.
- `go vet ./internal/agentcontract ./cmd/agentcontractgen` and `gofmt -w cmd internal`.
- Individual repository gates: `make tidy-check`, `make lint`, `make agent-contract-check`,
  `make docs-check-no-build`, `make smoke-no-build`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`, and
  `make release-workflow-check`.
- `git diff --check`, generated-file/status audit, and a changed-path audit confirming no home
  directory or Wave 6 legacy-role deletion.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. Red Pi and projection tests, with failing evidence in `TDD-LEDGER.md`.
3. Green canonical generator, clean discovery, and generated Pi workers.
4. Documentation, review, verification, and no-mistakes-fix checkpoint(s).
