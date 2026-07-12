# Plan: Jira Help Renderer / Docs

Parent issue: #81
Sub-issue: #105
Parent branch: `feat/81-jira-cli-parity`
Issue branch from tracker: `feat/jira-help-renderer`
Current execution: local critical-path slice in the parent checkout; no Pi subagent tool is available in this harness.

## Objective

Render Jira provider-style connector help from `cli_surface.json` metadata through runtime CLI help, connector manuals, and website connector data without enabling new Jira command execution paths beyond metadata/help display.

## GSD / Runtime Evidence

- Planning command: `scripts/gsd prompt plan-phase issue-105-jira-help-renderer --skip-research` generated successfully.
- Programming-loop command attempted: `scripts/gsd prompt programming-loop init --phase issue-105-jira-help-renderer --dry-run` failed with `unknown GSD command: programming-loop`.
- Manual GSD fallback active. Follow plan → red → green → refactor → verify → commit/push.
- Spawn decision: `local_critical_path` because current harness has no Pi `subagent` tool. #104 is verified; #105 is dependency-ready.

## Required Skills Loaded

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
- `golang-documentation`
- `frontend-design`
- `web-design-guidelines`
- `vercel-react-best-practices`
- `vercel-composition-patterns`

## Scope

In scope:

- Add red tests for `pm jira --help` and bare `pm jira` rendering connector command-surface help successfully.
- Preserve invalid-command behavior for unknown connector command paths.
- Add/update runtime help JSON behavior for connector command surfaces where applicable.
- Regenerate connector manuals so `docs/connectors/jira/MANUAL.md` and `SKILL.md` include `COMMAND SURFACE`.
- Regenerate website connector data so Jira has non-null `cliSurface` metadata.
- Add website test coverage for Jira `cliSurface` route data.

Out of scope:

- New Jira command dispatch (#106).
- Full Jira operation ledger (#107).
- Direct-read execution (#108).
- GraphQL/body-variable execution (#109).
- Sensitive/admin write policy (#110).
- Credentialed Jira checks or live API calls.

## Red Tests

Planned red tests before production edits:

```bash
go test ./internal/cli -run 'TestJiraConnectorCommandSurfaceHelp|TestBareJiraConnectorCommandShowsHelp' -count=1
cd website && pnpm test:unit -- connector-data
```

Expected first failures:

- `pm jira --help` currently returns a usage/runtime error because top-level help dispatch only knows static docs topics.
- `pm jira` currently returns `missing connector command path` instead of contextual connector help.
- Website connector data currently lacks Jira `cliSurface` until generated data/tests are updated.

## Green Implementation

1. Add tests in `internal/cli/cli_test.go` for `pm jira --help`, bare `pm jira`, and optional `--json --help` connector help output.
2. Add a small helper that detects connector command-surface providers before falling back to static docs manuals.
3. Render `connectors.RenderConnectorManual(c)` for connector help/bare command display; keep `--json` envelope as `CommandManual`.
4. Regenerate CLI/connector docs: `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors`.
5. Regenerate website data: `cd website && pnpm gen:website-data`.
6. Add/adjust website connector-data tests for Jira `cliSurface`.
7. Run focused and full verification.

## Safety Gates

- No secrets in examples or generated data.
- No credentialed Jira checks.
- No new dependencies.
- No generic raw HTTP write.
- No Jira write action enablement.
- No reverse ETL execution.
- No destructive/admin external actions.
