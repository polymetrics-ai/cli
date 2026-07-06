# Summary: GitHub CLI Surface Metadata

## Completed

- Added optional `cli_surface.json` parsing to connector bundles.
- Added a meta-schema and strict typed model for command groups, flags, command entries, API
  endpoint references, risks, approvals, and help topics.
- Added `connectorgen validate` checks for unknown streams, unknown writes, unknown API endpoint
  references, missing mappings for implemented ETL/reverse ETL commands, and secret-looking values
  in examples.
- Added GitHub production metadata at `internal/connectors/defs/github/cli_surface.json`.
- Updated GitHub connector notes, the public website docs, and `.agents` learning material.
- Fixed subagent review findings by embedding `cli_surface.json` in `defs.FS`, blocking unsafe
  implemented intents, requiring reverse ETL risk/approval metadata, rejecting excluded or
  mismatched endpoint references, and moving the website GitHub surface link out of the dispatched
  command map.
- Fixed CodeRabbit review finding on PR #48 by ensuring CLI-surface API endpoint references are
  validated whenever `api_surface.json` is present, including an intentionally empty endpoint list.
- Hardened the agentic delivery system so stacked sub-PRs cannot count skipped CodeRabbit status as
  approval and must be covered by sub-PR review records or parent-PR review records against `main`.

## Scope Control

This slice intentionally does not add `pm github ...` runtime dispatch, raw API execution, GraphQL
execution, direct reads, or direct writes. Later slices can render help and execute only commands
that resolve through this validated metadata.

## Accuracy Notes

- Current `gh ruleset` parity is limited to `check`, `list`, and `view`. GitHub repository ruleset
  create/update/delete remain connector-native reverse ETL write mappings, not current `gh` command
  paths.
- The `connection` flag in the metadata is a future `pm` connector convenience, not a GitHub CLI
  root global flag.
