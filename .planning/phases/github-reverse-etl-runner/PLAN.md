# GitHub Reverse ETL Runner Plan

Issue: #57
Parent: #44
Branch: `feat/57-github-reverse-etl-runner`

## Goal

Make GitHub command-surface entries with `intent: reverse_etl` and an existing write action create
approval-gated reverse ETL plans, preview those plans, and execute them only after approval.

## Scope

- Add a commandrunner write-command builder that maps declared `record.*` flags into one write
  record and validates it through the connector write validator.
- Keep generic `commandrunner.Run` non-mutating; it may run reads/direct reads but must not call
  connector writes.
- Add app-level connector command plans that reuse reverse-plan storage, token hashing, expiry,
  replay protection, status transitions, and write execution.
- Add provider-style CLI flow:
  - `pm github <write-command> ...` creates a plan.
  - `pm github <write-command> --plan <id> --preview --json` previews the stored plan.
  - `pm github <write-command> --plan <id> --approve <token> --json` executes the stored plan.
- Add initial GitHub flag mappings for `issue create`, `issue close`, and `repo deploy-key add`.
- Update GitHub connector docs and website data.

## Out Of Scope

- Adding new GitHub write actions.
- Executing operation-backed GraphQL, binary, XML, or local-git commands.
- Exposing secret/variable writes or raw API escape hatches.
- Making every existing `reverse_etl` command executable before it has explicit input mappings.

## Safety

- Plain provider-style write commands only create plans.
- Provider-style `--plan` execution must target a connector-command plan for the same connector and
  command path.
- JSON output redacts approval tokens, approval token hashes, and raw command payload records.
- The generic runner remains non-mutating.
- Command writes reuse existing approval-token replay and plan-hash checks.
- Commands without declared `record.*` mappings remain blocked.

## Verification Plan

- Red/green tests in `internal/connectors/commandrunner`.
- App approval regression tests remain green.
- CLI E2E test against `httptest` GitHub server proves plan, preview, bad approval rejection, and
  approved execution.
- Connector validation across all defs.
- Focused Go package tests and binary build.
