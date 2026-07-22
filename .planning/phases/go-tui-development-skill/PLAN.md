# PLAN: Go TUI Development Skill

## Objective

Research current Go terminal-UI libraries and production applications, compare the findings with
Polymetrics code and GitHub issues, and add a source-backed reusable skill that routes future Go TUI
work without adding a product feature or runtime dependency.

## Delivery Context

- Work branch: `fm/cli-go-tui-development-skill-r1`
- Base and analyzed SHA: `origin/main` at `873cd7b251f70c4a35a607a0d4e86051ea0fbd15`
- The branch was created before `git fetch origin main --prune`; fetched `origin/main` matched HEAD.
- Scope: `.agents/**`, a dated project research artifact, test-only validation, and these planning
  artifacts.
- Non-goals: no TUI feature, command/flag/help change, connector operation, runtime dependency,
  credential access, or GitHub mutation.

## GSD Activation

- `scripts/gsd doctor`: passed.
- `scripts/gsd list`: passed.
- `scripts/gsd sources programming-loop`: failed because the repo-local registry does not expose
  `programming-loop`.
- `scripts/gsd prompt quick --full "Research Go terminal UI libraries ..."`: generated and is being
  executed inline as the supported adapter fallback.
- Manual programming-loop fallback: this plan, `TDD-LEDGER.md`, and `VERIFICATION.md` record the
  test-first and verification evidence required by the missing command.

## Required Skills Used

- `gsd-core`: repo-local GSD workflow and fallback contract.
- `golang-how-to`: Go skill routing.
- `golang-cli`: Unix I/O, TTY, automation, and cancellation contracts.
- `golang-testing`: table-driven, deterministic, integration, race, fuzz, and leak tests.
- `golang-concurrency` and `golang-context`: event-loop effects, ownership, cancellation, and
  backpressure.
- `golang-safety`: panic, lifecycle, resource, and terminal-state restoration risks.
- `golang-code-style`: concise, reviewable guidance and any Go test code.
- `golang-security`: untrusted ANSI/OSC input, secrets, clipboard, paste, hyperlink, and log risks.
- `golang-documentation`: source-backed evergreen documentation with explicit modality.
- `no-mistakes`: final authoritative validation and PR delivery path; Claude and Copilot review are
  explicitly excluded by the task.

## Research And Inspection Slices

1. Inspect repository skill conventions, routing, agent schemas, validation patterns, command tree,
   terminal output, prompts, progress, tests, and dependencies.
2. Inventory relevant Polymetrics issues and PRs with read-only `gh-axi`; distinguish actual records
   from speculative future work.
3. Research authoritative library and production-application sources with `chrome-devtools-axi` and
   read-only GitHub metadata/source views.
4. Write a dated source ledger and project gap map with stable URLs, access date, evidence,
   recommendations, risks, acceptance tests, ordering, and a not-implemented boundary.
5. Add a concise `go-tui-development` skill plus selection, UX/accessibility,
   architecture/performance, and testing/compatibility references.
6. Wire the skill into required routing and the task-skill matrix.
7. Add a failing test-only contract for frontmatter, trigger coverage, internal links, routing, and
   YAML shape; then make the documentation/routing implementation pass.

## Red / Green / Refactor Plan

- Red: add a Go contract test that expects the new skill, required trigger terms, reference links,
  routing entries, and valid YAML before those files exist.
- Green: add the smallest complete skill, references, issue-gap artifact, and routing changes that
  satisfy the contract.
- Refactor: remove duplicated detail from `SKILL.md`, verify source claims and issue status, and make
  links relative and stable.

## Risks And Controls

- Library popularity can be mistaken for fit: use a multi-factor selection matrix, not stars.
- Issue mappings can rot: keep dated issue-specific detail outside the evergreen skill.
- Terminal claims can be version-sensitive: avoid unsourced API/version promises and retain dated
  source URLs.
- Accessibility can be overstated: require plain/non-TTY fallback and evidence, not a claim that a
  cell UI is screen-reader accessible.
- Test validation can become a prose parser: assert only durable contracts and parse YAML with the
  repository's existing `gopkg.in/yaml.v3` dependency.
- Product scope can creep: no command behavior, TUI package import, or `go.mod`/`go.sum` change.

## Commit Checkpoints

1. Plan and GSD fallback evidence.
2. Red skill-contract test.
3. Green research, skill, references, and routing.
4. Verification/refactor fixes and final evidence.
5. Final committed branch handed to firstmate; no-mistakes runs only after firstmate instructs it.
