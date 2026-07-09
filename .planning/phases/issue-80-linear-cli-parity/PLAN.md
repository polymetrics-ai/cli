# Issue #80 Parent Plan — Linear CLI parity

Date: 2026-07-09
Branch: `feat/80-linear-cli-parity`
Parent issue: https://github.com/polymetrics-ai/cli/issues/80
Sub-issues: #97-#103
Connector: `linear`

## GSD command path

- `scripts/gsd doctor` — passed.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — ran; command registry available (69 commands).
- `scripts/gsd prompt plan-phase issue-80-linear-cli-parity --skip-research` — generated the parent plan prompt.
- `scripts/gsd prompt programming-loop init --phase issue-80-linear-cli-parity --dry-run` — unavailable: `scripts/gsd: unknown GSD command: programming-loop`.

Manual-GSD fallback is active for the programming-loop portion because the repo-local adapter does not expose `programming-loop` in this checkout. The fallback still follows PLAN → RED → GREEN → REFACTOR → VERIFY and records evidence in this phase and sub-issue phase artifacts.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-spf13-cobra`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-graphql`
- `golang-documentation`
- `web-design-guidelines`
- `vercel-react-best-practices`
- `frontend-design` (for generated website/docs data parity; no hand-written UI changes in the first slice)

Required references read: `AGENTS.md`, issue and parent-orchestrator contracts, parent orchestration loop, stacked PR workflow, GSD universal runtime loop, GSD Pi adapter reference, automated review routing loop, CodeRabbit loop, CLI help/docs/website parity reference, connector migration handoff/conventions/design docs.

## Parent scope

Deliver a parent PR for Linear CLI parity with issue-scoped evidence, tests, validation, docs, and a final status table covering:

- official operations inventoried;
- operations mapped by app type: ETL, reverse ETL, direct read, binary/file, local workflow, blocked/excluded;
- operations implemented;
- operations intentionally blocked or deferred with reasons;
- certification/verification results.

## Branch and PR state

- Active branch: `feat/80-linear-cli-parity` (user-specified parent integration branch).
- Issue #80 body names `feat/linear-cli-parity`; this plan follows the user-provided branch and records the mismatch.
- Parent PR from `feat/80-linear-cli-parity` to `main`: not found at planning time.
- Sub-issue work may proceed locally as critical-path work in this checkout, but stacked PR review coverage remains pending until the parent PR exists.

## Ready queue and dependency order

| Issue | Lane | Write scope | Dependencies | Initial status |
|---:|---|---|---|---|
| #97 | CLI surface metadata | `internal/connectors/defs/linear/cli_surface.json`, focused tests/docs evidence | none | local critical path first |
| #98 | Help renderer/docs | help rendering, docs/website parity | #97 metadata | dependency blocked |
| #99 | Stream runner | command runner integration/tests | #97 metadata | dependency blocked |
| #100 | Operation ledger | `api_surface.json`/operation inventory | #97 metadata useful; can run after first slice | planned |
| #101 | Direct reads | direct-read metadata/executor safety | #100 and output policy | dependency blocked |
| #102 | GraphQL/advanced engine | fixed-document GraphQL/body variable support | #100 and safety review | planned/human-gate if engine-wide |
| #103 | Sensitive/admin policy | risk tiers/redaction/approval/blocked defaults | #100 operation classes | dependency blocked |

## Execution decision

No mutating subagent was spawned in this harness: `not_spawned_runtime_capability_missing` (the exposed toolset has no Pi `subagent` tool). The orchestrator takes `local_critical_path` for #97 in this checkout with separate phase artifacts.

## TDD slice plan

### Slice 1 — #97 Linear CLI surface metadata

1. RED: add a failing Go test proving the embedded Linear bundle exposes `cli_surface.json` and maps the four existing stream-backed commands to implemented ETL commands.
2. GREEN: add `internal/connectors/defs/linear/cli_surface.json` with safe, provider-inspired command metadata:
   - implemented ETL commands: `issue list`, `team list`, `project list`, `user list`;
   - planned direct-read commands only where safe and not currently implemented;
   - reverse-ETL and admin/destructive commands blocked/planned, not executable;
   - no generic raw GraphQL or write escape hatch.
3. REFACTOR: keep metadata schema-valid and docs-compatible; do not add dependencies or credentialed checks.
4. VERIFY: targeted engine test, `connectorgen validate internal/connectors/defs/linear`, then broader gates as time permits.

## Safety constraints

- No secrets in prompts, logs, fixtures, files, or test output.
- No credentialed Linear checks.
- No raw generic GraphQL mutation/read escape hatch.
- No new dependencies.
- No reverse ETL execution; writes remain plan → preview → approval → execute and blocked until explicit write actions/policy exist.
- Destructive/admin/elevated actions are blocked by default and human-gated.

## Commit checkpoints

1. Planning artifacts for #80 and #97.
2. Red test checkpoint for #97 when useful.
3. Green metadata implementation checkpoint after focused tests and validation pass.
4. Verification/documentation checkpoint after broader gates.

## Review plan

After local verification and PR creation/update, use automated review routing. For non-draft parent PRs targeting `main`, wait for CodeRabbit automatic review; do not post redundant manual commands. For stacked sub-PRs targeting the parent branch, record skipped/fallback coverage and use the parent PR review route when required.
