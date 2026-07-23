# Verification Checklist

## Base And Scope Correction

- [x] Disposable worktree isolation verified.
- [x] Original branch created and original `origin/main` SHA recorded.
- [x] Captain correction read before further edits.
- [x] Current `origin/feat/cli-architecture-v2` fetched.
- [x] Existing commit preserved by rebase onto parent SHA
  `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`.
- [x] Existing `bubble-tea-tui-design`, design docs, ADRs, #397 artifacts, and routing inspected.
- [x] Direct parent-orchestrator agent absence and bounded advisory fallback recorded.
- [x] Separate `skill-gap-spec.md` audit and adoption decision read before skill design froze.
- [x] No generic `go-tui-development` skill added.

## Program Research

- [x] Current issue #397 and draft PR #438 inspected read-only with `gh-axi`.
- [x] Relevant TUI issues and stacked PRs inventoried with `gh-axi`.
- [x] Current parent code/dependency state compared with default-branch baseline.
- [x] Authoritative TUI/library/application sources sampled with `chrome-devtools-axi` and
  read-only `gh-axi`; existing issue #462 decision found sufficient unless the audit identifies a
  concrete gap.
- [x] Dated artifact records parent base, actual/speculative distinction, source URLs, delivery
  gaps, recommendations, risks, sequence, and not-implemented boundary.
- [x] Volatile issue/queue/head/review status is absent from evergreen skill prose.

## Skill And Routing Contract

- [x] `cli-architecture-v2-delivery/SKILL.md` follows project frontmatter conventions.
- [x] Description force-triggers #397, PR #438, CLI Architecture v2, phase/subissue work,
  Cobra/Viper, events/TUI/accessibility, slog/OpenTelemetry, stacked orchestration, GSD/TDD,
  exact-head review, and parent readiness.
- [x] Skill starts with live issue/PR/code/phase inspection rather than a stale phase assumption.
- [x] Track A routes to ADR 0002 and CLI/Cobra/Viper/parity skills.
- [x] Track B routes to ADR 0003 and `bubble-tea-tui-design` without duplicating it.
- [x] Track C routes to ADR 0004 and observability/performance/security skills.
- [x] Parent queue/integration/readiness ownership remains with the live parent orchestrator.
- [x] Three focused references cover state/dependencies, per-phase delivery, and parent
  integration/review without duplicating specialist implementation.
- [x] Required routing and task matrix select the skill for #397 work.
- [x] MUST/SHOULD/MAY and a program-slice definition of done are explicit.

## Automated Validation

- [x] Focused RED failed before skill/routing implementation (missing `SKILL.md`, exit 1).
- [x] Focused GREEN passes directly and through `make cli-architecture-v2-skill-check`.
- [x] Frontmatter, trigger terms, Markdown/local links, YAML routing, and contradiction markers are
  validated without a new dependency.
- [x] Focused target is not wired into global `make verify`.
- [x] `scripts/tests/pi-model-routing.sh`
- [ ] `go test ./...`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check`
- [ ] `go.mod` and `go.sum` unchanged from the corrected parent base.

## Safety And Delivery

- [x] No product CLI/TUI behavior, help/manual/website output, runtime dependency, or shared parent
  artifact changed.
- [ ] No secret, credentialed connector check, runtime lifecycle action, generic write tool, or
  reverse ETL execution.
- [ ] #419 remains deferred and no dependency approval is implied.
- [ ] No Claude or GitHub Copilot review invoked.
- [ ] Changes committed on the worker branch.
- [x] Firstmate instructed the no-mistakes shipping stage after focused implementation/validation.
- [ ] Final stacked PR targets `feat/cli-architecture-v2`, uses `Refs #397`, and remains unmerged.
