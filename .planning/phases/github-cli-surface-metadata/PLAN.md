# Plan: GitHub CLI Surface Metadata

## Branches

- Parent branch: `feat/44-github-cli-parity`
- Sub-issue branch: `feat/34-cli-surface-metadata`
- Parent issue: #44
- Sub-issue: #34

## Steps

1. Read project rules, issue #34, and GitHub CLI primary sources.
2. Add red tests for optional `cli_surface.json` parsing and validation.
3. Implement loader structs, meta-schema, and optional file parsing.
4. Implement `connectorgen validate` reference checks.
5. Add GitHub `cli_surface.json` metadata for the stable docs-only command surface.
6. Update GitHub connector docs and website-facing docs to mention the CLI surface metadata.
7. Run targeted tests and connector validation.
8. Update summary, verification, and run-state artifacts.

## Slice Boundaries

This slice is schema/docs only. Later issues own help rendering, stream-backed execution, direct
reads, GraphQL, sensitive/admin policy, and cross-connector rollout.

## Human Gates

- No new dependencies.
- No auth scope changes.
- No production deployment.
- No destructive external actions.
