# PLAN — issue #3721 clean project-local Codex workers

Parent: #3714. Branch: `fm/cli-agents-wave-codex-r1`. Base:
`refactor/3714-canonical-delivery-flow`. The eventual sub-PR base is the same parent branch.

## GSD path

- `scripts/gsd doctor` and `scripts/gsd list`: passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`: passed.
- Generated prompts:
  - `scripts/gsd prompt discuss-phase issue-3721-codex-workers --auto`
  - `scripts/gsd prompt plan-phase issue-3721-codex-workers --tdd --skip-research`
  - `scripts/gsd prompt execute-phase issue-3721-codex-workers --interactive`
- Inline/manual execution is required by the canonical single-worker/no-delegation contract. No
  GSD or specialist role is spawned.

## Required skills loaded

- `github-issue-first-delivery`
- `no-mistakes` v1.41.2
- `golang-how-to`, `golang-testing`, `golang-cli`, `golang-error-handling`, `golang-security`,
  `golang-safety`, and `golang-code-style`
- `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and
  `gsd-code-review`, constrained to inline execution

## Scope

Canonical source, generator, validation tests, generated Codex TOML projections, and concise
project-trust documentation only. Do not delete the legacy project role files, touch any home
directory, add dependencies, or change CLI user-facing behavior.

## RED → GREEN → REFACTOR slices

### Slice 1 — Codex isolation and format contract

1. **RED:** Add a negative delegation test against the current projection rendering. It must fail
   because no worker configuration sets `agents.enabled = false`, which officially defaults to
   enabled and permits reaching an injected ambient agent.
2. **GREEN:** Extend the canonical contract with verified Codex projection facts: standalone
   format, mandatory fields, trusted-project requirement, documented precedence, explicitly
   undocumented filename collision behavior, and the exact delegation setting. Add typed
   validation and a Codex TOML renderer that emits both roles from this source.
3. **REFACTOR:** Keep shared Markdown projections intact, make projection rendering mode explicit,
   and use bounded root-relative creation/replacement so sync can create required generated files
   without following a path outside the selected root.

### Slice 2 — Drift, parser, and discovery evidence

1. **RED:** Add tests that model an unrelated ambient agent and fail while the emitted config
   leaves multi-agent tools enabled.
2. **GREEN:** Add TOML parse/schema tests using the repository's existing TOML parser transitively
   supplied by Viper, drift tests that reject removal of `agents.enabled = false`, and sync tests
   that create only the two required Codex output files.
3. **REFACTOR:** Generate the exact target files with `go run ./cmd/agentcontractgen sync`, run the
   contract check, and record what the static test proves versus runtime conditions it cannot
   prove (notably project trust and inherited non-agent configuration).

### Slice 3 — Trust bootstrap and handoff evidence

1. Document trust prerequisite/fail-closed behavior beside the canonical delivery contract without
   duplicating generated policy by hand.
2. Add the current Codex CLI discovery smoke and a no-global-path diff audit to verification.
3. Run the required GSD verify/review prompts inline, record any gaps, then use the child
   no-mistakes command with `--skip=push,pr,ci` after committed green work. Open no PR against
   `main`; use `gh-axi` only to create a sub-PR to the parent branch if the outer worker lifecycle
   authorizes that stage.

## Verification plan

- Focused RED/GREEN: `go test ./internal/agentcontract ./cmd/agentcontractgen`.
- Canonical check: `go run ./cmd/agentcontractgen check`.
- Formatting/static: `gofmt -w cmd internal`; `go vet ./internal/agentcontract ./cmd/agentcontractgen`.
- Repository gates: `make tidy-check`, `make lint`, `make agent-contract-check`, `make docs-check-no-build`,
  `make smoke-no-build`, and `make release-workflow-check` as applicable.
- Codex smoke: parse the generated TOML in unit tests; run `codex doctor --json` from this trusted
  worktree and record any project-discovery evidence without reading or changing user config.
- Scope audit: `git diff --check`, generated-file drift check, and a changed-path check proving no
  `~/.codex` path is affected.

## Commit checkpoints

1. Plan/context/TDD ledger.
2. Executed RED delegation test evidence.
3. Green canonical source, renderer, tests, and generated files.
4. Refactor/docs/verification/review fixes.
