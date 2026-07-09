# Plan: GitLab CLI Surface Metadata (#83)

Parent issue: #78
Parent branch / active branch: `feat/78-gitlab-cli-parity`
Connector: `gitlab`
Primary artifact: `internal/connectors/defs/gitlab/cli_surface.json`

## GSD Command Evidence

- Parent planning prompt: `scripts/gsd prompt plan-phase issue-78-gitlab-cli-parity --skip-research`.
- Lane execution prompt: `scripts/gsd prompt execute-phase issue-83-gitlab-cli-surface --tdd`.
- Programming-loop fallback: `scripts/gsd prompt programming-loop init --phase issue-78-gitlab-cli-parity --dry-run` is unavailable in this adapter registry (`unknown GSD command: programming-loop`). This lane follows the manual GSD universal loop: plan, red test, green implementation, refactor, verify, record evidence.

## Required Skills Loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-spf13-cobra` (loaded for CLI command behavior awareness; no Cobra edits expected in #83).

## Scope

Add production GitLab `cli_surface.json` metadata that maps glab-inspired command families to safe Polymetrics app intents without enabling unsafe execution.

In scope:

- Implemented ETL command metadata for the four existing GitLab streams:
  - `project list` → `projects`
  - `group list` → `groups`
  - `user list` → `users`
  - `issue list` → `issues`
- Planned/partial docs-only command metadata for likely future direct reads and reverse ETL, with risk/approval notes where writes are described.
- Unsupported/local workflow metadata for commands that require local git/browser/config/alias behavior.
- Explicitly unsafe/excluded metadata for raw API escape hatches and destructive/admin surfaces.

Out of scope for #83:

- Adding new streams, writes, direct-read operations, GraphQL documents, binary downloads, or local workflow executors.
- Runtime help rendering and docs/website parity beyond metadata-only notes; #84 owns rendered help/docs parity.
- Credentialed GitLab checks.
- Reverse ETL execution.

## Red / Green / Refactor Plan

1. Red: add `TestBundleLoadEmbeddedGitLabCLISurface` in `internal/connectors/engine/bundle_test.go` proving the embedded GitLab bundle exposes non-nil CLI surface metadata and the four current stream-backed commands.
2. Run the focused red test and record failure in `TDD-LEDGER.md`.
3. Green: create `internal/connectors/defs/gitlab/cli_surface.json` with schema-valid, secret-free metadata.
4. Run focused test and `go run ./cmd/connectorgen validate internal/connectors/defs --json`.
5. Refactor: keep metadata minimal and honest; avoid overclaiming executable writes/direct reads.
6. Update this phase's verification and run state.

## CLI Help / Docs / Website Parity Checklist

- Runtime help (`pm gitlab --help`, `pm help gitlab`): deferred to #84 unless existing generic help automatically reflects `cli_surface.json` without new code.
- Bare namespace behavior: deferred to #84/#85 for command runtime behavior.
- `docs/cli/**`: deferred to #84.
- `website/**`: deferred to #84 unless website bundle generation requires metadata fixtures.
- Generated help/manual artifacts: deferred to #84.
- Safety note for this lane: metadata must not claim generic raw API writes or executable destructive/admin writes.

## Verification Checklist

Focused:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedGitLabCLISurface -count=1
go test ./cmd/connectorgen ./internal/connectors/engine -run 'CLISurface|GitLab' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Before handoff/commit if this branch accumulates production code beyond metadata:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Safety Gates

- No secrets in examples; do not include access-token-looking literals.
- No credentialed checks or live writes.
- No new dependencies.
- No generic raw HTTP write, generic shell write, generic SQL write, or arbitrary GraphQL mutation escape hatch.
- Binary/file transfer remains blocked or planned until bounded output policy exists.
- Sensitive/admin/destructive writes must remain blocked or described as reverse ETL plan → preview → approval → execute only.
