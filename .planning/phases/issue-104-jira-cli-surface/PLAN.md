# Plan: Jira CLI Surface Metadata

Parent issue: #81
Sub-issue: #104
Parent branch: `feat/81-jira-cli-parity`
Issue branch from tracker: `feat/jira-cli-surface`
Current execution: local critical-path slice in the parent checkout; split to the tracker branch before PR if the diff grows beyond #104.

## Objective

Produce and validate `internal/connectors/defs/jira/cli_surface.json` as docs/help metadata. Map Jira-provider-like commands into safe app intents without adding dispatcher behavior, writes, raw HTTP access, credential checks, or live Jira calls.

## GSD / Runtime Evidence

- Planning command: `scripts/gsd prompt plan-phase issue-104-jira-cli-surface --skip-research` generated successfully.
- Programming-loop command attempted: `scripts/gsd prompt programming-loop init --phase issue-104-jira-cli-surface --dry-run` failed with `unknown GSD command: programming-loop`.
- Manual GSD fallback active. Follow plan → red → green → refactor → verify → commit/push.
- Spawn decision: `local_critical_path` because current harness has no Pi `subagent` tool and the parent PR is not open yet.

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

## Scope

In scope:

- Add a red embedded-bundle test proving Jira command-surface metadata is expected at runtime.
- Add `internal/connectors/defs/jira/cli_surface.json`.
- Keep implemented commands limited to existing streams and covered endpoints:
  - `issue list` → stream `issues`, `GET /rest/api/3/search`
  - `project list` → stream `projects`, `GET /rest/api/3/project/search`
  - `user list` → stream `users`, `GET /rest/api/3/users/search`
- Classify representative provider-like Jira commands as `partial`, `planned`, `unsupported_api`, `unsupported_local`, or `unsafe_or_disallowed` when not safely executable in this slice.
- Document reverse-ETL-related commands as non-executable metadata only. Do not add `writes.json` or write refs.

Out of scope:

- Full 620-operation ledger (#107).
- Direct-read execution (#108).
- Runtime help renderer/docs generation (#105).
- Stream command dispatch (#106).
- GraphQL/body-variable execution (#109).
- Sensitive/admin reverse-ETL policy implementation (#110).
- Credentialed Jira checks or live API calls.

## Red Test

Add a test before `cli_surface.json` exists:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
```

Expected first failure:

```text
Jira CLISurface is nil; defs.FS must embed cli_surface.json
```

## Green Implementation

1. Add `TestBundleLoadEmbeddedJiraCLISurface` near the GitHub embedded CLI-surface test.
2. Add `internal/connectors/defs/jira/cli_surface.json` matching `cli_surface.schema.json`.
3. Validate references:
   - implemented ETL commands have a stream reference;
   - any `api_surface` reference exists and matches the stream-covered endpoint;
   - no command exposes `raw_api` or `direct_write` as implemented;
   - no examples include secret-looking values.
4. Run focused tests and validation.

## Refactor Notes

- Keep metadata conservative. `cli_surface.json` is not a promise that every Jira command is executable.
- Prefer `planned` for safe future direct reads and `unsafe_or_disallowed` for admin/destructive/sensitive writes until #107/#110 classify them.
- Use Jira command words familiar to Jira users (`issue`, `project`, `user`, `sprint`, `board`, `workflow`, `attachment`) but keep source links to Atlassian docs/OpenAPI rather than relying on a third-party CLI as authority.

## Verification

Targeted:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedJiraCLISurface -count=1
go test ./internal/connectors/engine -run CLISurface -count=1
go test ./cmd/connectorgen -run CLISurface -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Broader before handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI Help / Docs / Website Parity

Applies: yes, but #104 is metadata-only.

- Runtime help: baseline checked with `go run ./cmd/pm help connectors`.
- Connector inspection: baseline checked with `go run ./cmd/pm connectors inspect jira --json`.
- Bare namespace behavior: not changed in #104.
- `docs/cli/**`: not changed unless metadata consumers require docs updates.
- `website/**`: not changed unless metadata consumers require docs updates.
- Generated help/manual artifacts: not changed in #104.

## Safety Gates

- No secrets in examples or fixtures.
- No credentialed Jira checks.
- No new dependencies.
- No generic raw HTTP write.
- No write action enablement.
- No reverse ETL execution.
- No destructive/admin external actions.
