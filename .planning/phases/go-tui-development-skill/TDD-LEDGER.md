# TDD Ledger

Phase: `go-tui-development-skill`

## Baseline

- Current repository has one project skill, `.agents/skills/caveman/SKILL.md`.
- Required skill routing and the task-skill matrix have no Go TUI task/group.
- No test currently validates project skill frontmatter, trigger coverage, links, or routing.
- Polymetrics product code and issue inventory still require inspection before making a framework
  recommendation.

## Red: Go TUI Skill Contract

- Status: planned.
- Test location: `internal/agentdocs/go_tui_skill_test.go` (test-only package).
- Expected contract:
  - skill frontmatter contains `name` and a force-triggering `description`;
  - required library and task trigger terms are present;
  - every required local reference exists and linked Markdown paths resolve;
  - `required-skills-routing.md` and `task-skill-matrix.yaml` route Go TUI work;
  - task matrix parses as YAML and names the new skill;
  - skill retains non-interactive, accessibility, restoration, cancellation, safe-output,
    deterministic-test, and definition-of-done gates.
- Red command: `go test ./internal/agentdocs`.
- Expected failure: skill and routing do not exist yet.

## Green: Skill, References, Routing, And Research

- Status: pending.
- Implementation target: project skill, four evergreen references, dated source ledger, dated
  Polymetrics gap map, and two routing surfaces.
- Green command: `go test ./internal/agentdocs`.

## Refactor

- Status: pending.
- Keep routine `SKILL.md` concise; move detailed matrices/checklists/guides into references.
- Remove duplicate issue-specific content from the evergreen skill.
- Recheck Markdown links, source dates, issue status, YAML parsing, and `git diff --check`.
