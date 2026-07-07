# GitHub CodeRabbit Review Fixes

## Objective

Resolve still-valid automated review findings on parent PR #49 without expanding the GitHub CLI
parity scope.

## Scope

- Remove redundant connector command write precheck in `internal/cli/cli.go`.
- Move reverse CLI test handler failures onto the main test goroutine.
- Add GitHub issue-delete scope metadata to `internal/connectors/defs/github/operations.json`.
- Extract duplicated website CLI surface mapping into a shared script helper.
- Decline unsupported `api_surface.schema.json` conditional schema change because the embedded
  schema compiler rejects `allOf`/conditional keywords and Go validation already enforces the
  invariant.

## Safety

- No live GitHub API calls.
- No secret values, fixtures, or credentials added.
- No production connector behavior expansion.
