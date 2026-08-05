# PLAN — clean project-local Claude workers

Issue: #3716. Parent: #3714. Branch: `fm/cli-agents-wave-claude-r1`.

## GSD path

- `scripts/gsd doctor`: passed.
- `scripts/gsd list`: passed; `programming-loop` is absent.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review|ship`:
  passed.
- `scripts/gsd prompt discuss-phase issue-3716-claude-workers --auto`: generated and executed
  inline as the discussion record.
- `scripts/gsd prompt plan-phase issue-3716-claude-workers --tdd --skip-research`: generated and
  executed inline through the plan and RED/GREEN/REFACTOR slices below.
- Generated `execute-phase`, `verify-work`, and `code-review` prompts were executed inline. The
  Pi adapter cannot supply the workflow's reviewer/UAT subagents, so their prescribed artifacts
  are recorded manually below; no agent role was spawned because the canonical base worker
  declares delegation `none`.

## Required skills loaded

- `no-mistakes` v1.41.2 — exact child-gate authority and non-default-base workaround.
- `github-issue-first-delivery` — issue/stacked PR linkage and hard-stop rules.
- `golang-how-to` — task-specific Go skill router.
- `golang-testing` — isolated red/green frontmatter, drift, and ambient-delegation regression
  tests.
- `golang-cli` — generator command behavior and stdout/stderr boundaries.
- `golang-error-handling`, `golang-safety`, and `golang-security` — bounded projection paths,
  defensive errors, and no expansion across the repository boundary.
- `golang-lint` — focused lint/review verification.
- `gsd-programming-loop` was read for its TDD lifecycle, but its absent adapter command is not
  invoked; this phase follows the canonical installed GSD sequence instead.

## Deliverable slices

### Slice 1 — RED: Claude projection and isolation contract

1. Add a focused test describing both Claude projections as required full files with parseable YAML
   frontmatter, exact minimal tool list, `permissionMode: default`, and no `Agent` capability.
2. Model an unrelated ambient agent in the test fixture and make the assertion fail against the
   current generator, which has no Claude frontmatter projection and therefore cannot prove that
   the ambient agent is unreachable.
3. Record the failing focused test command and failure reason in `TDD-LEDGER.md` before production
   code or canonical source changes.

### Slice 2 — GREEN: canonical policy, renderer, generation, and drift enforcement

1. Add a validated Claude harness policy to
   `.agents/agentic-delivery/canonical/delivery-contract.json`, including the official source URL,
   project discovery/precedence evidence, exact tool allowlist, and default permission mode.
2. Extend `internal/agentcontract` to render complete Markdown/YAML Claude files from that policy,
   validate their frontmatter, create missing required targets only through the existing
   root-contained atomic writer, and compare full files for drift.
3. Make the two Claude projection targets required, generate them with
   `go run ./cmd/agentcontractgen sync`, and rerun the focused tests plus the canonical check.
4. Add regression coverage that adding `Agent` to a generated file is rejected as drift and sync
   restores the no-delegation file.

### Slice 3 — REFACTOR: runtime smoke, documentation, and scope audit

1. Run the installed Claude CLI from a trusted project with each `--agent` name to prove project
   discovery. Use an isolated temporary home with ambient fixture agents and a bounded,
   non-mutating prompt that requests delegation; capture that the generated worker exposes no
   `Agent` tool rather than claiming precedence alone creates isolation.
2. Document the exact official behavior and caveats in generated projection metadata and phase
   verification: managed settings and CLI `--agents` override project definitions, and `Agent`
   omission blocks invocation but does not erase ambient definitions from their own scopes.
3. Run review, validation, diff, and path-scope checks. Do not touch installed plugins or any home
   directory.

## Verification

- RED/GREEN: `go test ./internal/agentcontract ./cmd/agentcontractgen`.
- Canonical/drift: `go run ./cmd/agentcontractgen sync` and
  `go run ./cmd/agentcontractgen check`.
- Claude smoke: installed `claude` v2.1.222 in this trusted project for both generated names;
  temporary ambient fixtures only, never real home/plugin paths.
- Formatting/static: `gofmt -w cmd internal`, `go vet ./internal/agentcontract ./cmd/agentcontractgen`,
  and `make lint`.
- Repository gates applicable to this scope: `make tidy-check`, `make docs-check-no-build`,
  `make smoke-no-build`, `make agent-contract-check`, `make connector-boundary`, and
  `make release-workflow-check`; connector validation/surface checks are run separately according
  to the repository timeout policy.
- Delivery: `no-mistakes axi run --intent <complete issue intent> --skip=push,pr,ci`, then use
  `gh-axi` to push/open a conventional-title sub-PR against
  `refactor/3714-canonical-delivery-flow`. Never use `--yes` or target `main`.

## Commit checkpoints

1. Plan/context/TDD checkpoint.
2. Executed RED test checkpoint with recorded failure.
3. Green canonical policy/generator/generated files checkpoint.
4. Review/verification fixes checkpoint, if needed.
