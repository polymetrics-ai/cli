# PLAN — connector-guardrail-remediation-r1

## GSD mode

- Adapter health: `scripts/gsd doctor` passed 2026-08-02 in Pi worktree.
- Planning command traces:
  - `scripts/gsd prompt map-codebase --fast` → `traces/gsd-map-codebase-fast.prompt.md`
  - `scripts/gsd prompt plan-phase connector-guardrail-remediation-r1 --skip-research` → `traces/gsd-plan-phase.prompt.md`
  - `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` → `traces/gsd-execute-phase.prompt.md`
- Fallback note: `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run` returned `unknown GSD command: programming-loop`. The repo-local adapter is otherwise healthy, so this phase uses the manual GSD universal runtime loop plus the available `execute-phase` prompt trace. TDD, subagent spawning, review, no-mistakes, and state-ledger requirements are not weakened.

## Required skills loaded

- `gsd-core`
- `caveman`
- `no-mistakes`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-lint`
- `golang-context`
- `golang-concurrency`

## Required policy/context loaded

- `AGENTS.md`
- Captain authorization decision and first-eight audit report
- Parent-orchestrator contract, parent orchestration loop, stacked PR workflow, GSD universal loop, Pi active orchestration loop, automated review routing, Claude review loop
- GSD Pi adapter reference, required-skills routing, CLI help/docs/website parity reference, worker handoff template, orchestration-state schema
- Connector migration handoff, conventions, architecture v2 design, connector boundary guard docs

## Objective

Create and integrate a parent remediation PR that enforces connector implementation path ownership and records forward remediation for all first-eight connector campaign findings, without rewriting `main` history and without touching the five active connector branches.

## Parent artifacts

- Parent issue: #3579
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Parent PR: #3580 draft PR to `main`; parent merge remains human-gated
- State ledger: `.planning/phases/connector-guardrail-remediation-r1/RUN-STATE.json`

## Slice plan

### Slice 1a — icon registry single-source foundation (#3595)

Write scope:

- `internal/connectors/icon_data.json`
- `internal/connectors/icons.go`
- `cmd/iconregistrygen/**`
- `website/scripts/gen-connector-bundles.mjs`
- `website/scripts/fetch-simple-icons.mjs`
- deletion of `website/data/icon_overrides.json`
- canonical SVG assets under `docs/connectors/icons/**`; `website/public/connectors/**` icons are generated/copied outputs only
- focused tests and authoritative docs

Outcomes:

- One authored connector-to-icon registry with canonical bare connector identifiers.
- No `source-*` / `destination-*` mapping fallback in runtime, website, fetching, or ownership consumers.
- Curated Simple Icons and all fetch/review metadata live in the canonical registry and canonical docs asset tree.
- Website output contains no mapping or SVG that exists only in the website tree.
- Source/destination collapses are audited without silent prefix stripping or ordering choices.

TDD first evidence: failing registry/consumer/ownership tests or proofs before production edits.

### Slice 1 — target-scope contract and core validator

Write scope:

- `cmd/connectorgen/**` for a new or extended validator command
- `internal/connectors/boundary/**` for reusable changed-path ownership validation logic
- focused tests under the same packages
- minimal docs for validator usage

Outcomes:

- Machine-readable connector implementation scope contains exactly one connector slug.
- Changed paths auto-detect connector implementation diffs from `internal/connectors/defs/<slug>/`, connector docs, generated connector outputs, hooks/native paths, fixtures/tests, and planning artifacts.
- The guard never trusts label/tag/scope omission as authority.
- Closed allowed path classes reject shared runtime/tooling, unrelated connectors, unrelated docs, and unrelated generated churn.
- Regression tests include positive target defs/fixtures/docs and negative shared runtime, unrelated defs, unrelated generated docs, and label-bypass cases.

TDD first evidence: add failing validator unit tests and CLI command tests before production implementation.

### Slice 2 — GitHub Actions, label/tag, local hook, required remote gate

Write scope:

- `.github/workflows/**`
- repository hook setup/config files already used by the repo
- `Makefile` guard target wiring
- narrow docs/runbook updates for the new guard

Outcomes:

- Ergonomic `connector-implementation` label/tag/scope presentation exists, but the remote check auto-detects relevant diffs and fails even without the label.
- GitHub Actions check runs on relevant PRs and `main` pushes and calls the same validator as local hooks.
- Local pre-commit/pre-push hook provides fast feedback and documents bypassability.
- Required GitHub ruleset/branch protection requires the guard check for `main`; if permission is denied, record exact blocker and do not claim completion.

TDD first evidence: add failing workflow/command fixture or validator invocation test before wiring.

### Slice 3 — PM orchestrator and no-mistakes integration (provisionally integrated via #3588)

Write scope:

- `.agents/**` delivery contracts/workflows/references/templates
- `.pi/prompts/**` and `.pi/agents/**` only where connector-lane routing text is needed
- issue/PR template files under `.github/**`
- no-mistakes guidance docs/config files if present in repo

Outcomes:

- Connector implementation instructions stop when shared foundation work is required.
- Connector PRs cannot absorb generic shared runtime/tooling or unrelated connector changes.
- Separate foundation sub-issue/PR flow is documented and tied to guard exceptions.
- Worker handoffs and PR bodies include target connector scope and guard evidence.

TDD first evidence: documentation/config validation or grep-based tests before production guidance edits, or docs-only exemption recorded if no executable check exists.

### Slice 4 — HubSpot and Bitbucket forward remediation

Write scope:

- forward remediation tests/artifacts for PR #3529 and #3531
- only the specific shared CLI/app/manifest/runtime paths identified in the audit if a code correction is required and non-overlapping with Slice 5
- remediation ledger entries

Outcomes:

- Disposition every shared CLI/app/manifest/runtime path identified for HubSpot and Bitbucket.
- Preserve demonstrably valid foundations with ownership/tests.
- Remove or correct harmful/unrelated churn with forward commits only.

TDD first evidence: add failing regression/golden/test or ledger consistency check before any corrective code changes.

### Slice 5 — Stripe, Freshchat, and Google Ads shared engine/runner/connectorgen remediation

Write scope:

- `internal/connectors/engine/**`, `internal/connectors/commandrunner/**`, `cmd/connectorgen/**`, `internal/connectors/command_surface.go`, and targeted tests only as needed for the audited PR #3530/#3535/#3536 dispositions
- remediation ledger entries

Outcomes:

- Disposition shared engine/write, commandrunner, connectorgen, command-surface changes.
- Preserve required general foundations only with explicit ownership/tests.
- Remove or correct unrelated behavior.
- Avoid overlap with Slice 4.

TDD first evidence: focused failing tests for behavior preserved/corrected before code changes.

### Slice 6 — Zendesk Support and Google Ads unrelated-connector/generated remediation

Write scope:

- generated/manual/docs/website remediation for PR #3532 and #3535 only
- target connector/generated ledger entries
- generator validation tests if needed

Outcomes:

- Disposition unrelated connector manuals/skills and broad generated/website churn.
- Remove or regenerate only demonstrably unrelated/stale outputs.
- Preserve valid shared indexes and target outputs.

TDD first evidence: add guard fixture/ledger test proving unrelated connector docs/generated paths fail, plus any generator-specific regression before broad generated changes.

### Slice 7 — historical audit ledger and end-to-end enforcement proof

Write scope:

- `docs/migration/**` or `.planning/phases/connector-guardrail-remediation-r1/**` ledger/proof artifacts
- guardrail fixture files/tests under validator package

Outcomes:

- Record all eight dispositions, including compliant Xero and Asana baselines.
- Add positive/negative fixtures proving allowed and forbidden path classes.
- Record labels cannot bypass auto-detection.
- Record explicit foundation PR path.
- Produce end-to-end proof and verification transcript.

TDD first evidence: ledger schema/check test or guard proof fixture fails before implementation.

## Dependency graph

- Slice 1 blocks Slice 2 and Slice 7 enforcement proof.
- Slice 1 should run before remediation slices for guard feedback, but remediation evidence can scout/read in parallel if write scopes are disjoint.
- Slice 4, Slice 5, Slice 6 are disjoint by path family and may run in parallel after parent PR exists.
- Slice 3 may run in parallel with remediation after Slice 1's public contract stabilizes.

## Spawn policy

- Use Pi `subagent` with `agentScope: "both"`, `confirmProjectAgents: false` only after reviewing project agents.
- Max four concurrent mutating workers.
- Each worker gets one issue, one branch, one worktree/cwd, one write scope, one handoff template.
- Orchestrator owns parent issue/PR body, state ledger, merge arbitration, and parent branch integration.

## Human gates / blockers

- Do not modify/reset/rebase/merge any active connector branches.
- Do not merge parent PR to `main`.
- Do not use secrets or credentialed connector checks.
- Do not add dependencies.
- Escalate branch protection/ruleset permission denial exactly.
- Escalate any no-mistakes daemon error; do not restart daemon.
