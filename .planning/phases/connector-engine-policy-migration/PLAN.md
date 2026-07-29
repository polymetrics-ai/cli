# Connector engine/direct-read policy migration

## Scope

Focused Issue-B migration from `/Users/karthiksivadas/karthik-agent-workspace/data/cli-connector-guard-rollout-migration-r1/report.md` after boundary guard PR #605. Remove GitHub-specific date-range and repository contents policy identifiers/behavior from shared engine, conformance, validator, and commandrunner runtime. Keep certification, generated-doc overhaul, native/hook hardening, PM Broker, website release, and unrelated connector work out of scope.

## GSD / issue trace

- GSD command path: `scripts/gsd doctor` + `scripts/gsd list`; `scripts/gsd prompt programming-loop ...` is referenced by repo policy but is not registered in `.gsd/commands.json` in this checkout (`scripts/gsd prompt programming-loop --help` returned unknown command). Manual GSD universal-loop fallback recorded here using `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.
- Orchestration decision: `local_critical_path` for plan/TDD/implementation because this is one focused shared-runtime slice in one isolated disposable worktree; read-only review/verification may be run after green tests.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-testing`, `golang-security`, `golang-safety`, `golang-code-style`, `golang-lint`, `golang-error-handling`.
- Issue search: `gh-axi search issues` for exact Issue-B/date-range/contents/provider-neutral terms found no exact focused standalone issue. Existing exception ledger maps date-range residue to #64 and repository-contents residue to #60; this branch will keep the PR focused and can reference those without broadening their feature scope.

## Baseline before edits

`go run ./cmd/connectorgen boundary . --json` (whole tree) before production edits:

- outcome: clean
- checked_files: 129
- findings: 0
- warnings: 0
- applied exceptions: 24
- target exceptions in scope: 12
  - `github-connectorgen-param-format`: `cmd/connectorgen/validate.go` `github_date_range` matches=1
  - `github-conformance-date-range-format`: `internal/connectors/conformance/dynamic.go` `github_date_range` matches=2
  - `github-engine-date-range-format`: `internal/connectors/engine/read.go` `github_date_range` matches=2
  - `github-connectorgen-output-policy-file`: `cmd/connectorgen/validate.go` `github_contents_file_metadata` matches=1
  - `github-connectorgen-output-policy-directory`: `cmd/connectorgen/validate.go` `github_contents_directory` matches=1
  - `github-commandrunner-output-policy-file`: `internal/connectors/commandrunner/runner.go` `github_contents_file_metadata` matches=1
  - `github-commandrunner-output-policy-directory`: `internal/connectors/commandrunner/runner.go` `github_contents_directory` matches=1
  - `github-engine-direct-read-policy-const-file`: `internal/connectors/engine/direct_read.go` `directReadPolicyGitHubContentsFileMetadata` matches=3
  - `github-engine-direct-read-policy-const-directory`: `internal/connectors/engine/direct_read.go` `directReadPolicyGitHubContentsDirectory` matches=3
  - `github-engine-direct-read-policy-value-file`: `internal/connectors/engine/direct_read.go` `github_contents_file_metadata` matches=1
  - `github-engine-direct-read-policy-value-directory`: `internal/connectors/engine/direct_read.go` `github_contents_directory` matches=1
  - `github-engine-direct-read-redaction-helper`: `internal/connectors/engine/direct_read.go` `redactGitHubContentsObject` matches=3

## Design

1. Replace `github_date_range` with a provider-neutral incremental contract:
   - add `IncrementalSpec.OperatorPrefix` (`json:"operator_prefix,omitempty"`) as bounded comparison-prefix metadata;
   - add generic `rfc3339_utc` timestamp formatting for lower-bound values requiring UTC normalization;
   - compose `param_format: "rfc3339_utc"` + `operator_prefix: ">="` for date-range-style query qualifiers.
2. Replace `github_contents_file_metadata`/`github_contents_directory` with generic repository contents policies:
   - `repository_contents_file_metadata`;
   - `repository_contents_directory`.
   Keep existing shape checks, content/download URL redaction, sensitive repository path blocking, and file-vs-directory behavior.
3. Update GitHub CLI surface metadata to use the generic policy names and add synthetic second-connector unit coverage using the generic policy names.
4. Update validation schemas/rules, conformance assertion helper, tests, and `docs/migration/conventions.md` as the authoritative contract.
5. Remove only the drained exception-ledger rows. Do not loosen scanner lexicon/rules/max-match/expiry logic.

## Non-goals

- No certification metadata migration.
- No generated-doc overhaul or website release work.
- No runtime-backed or credentialed connector checks.
- No new dependencies.
- No broad connector migrations beyond GitHub metadata required for the renamed generic policies.
