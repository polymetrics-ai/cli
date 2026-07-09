# Prompt Snapshot

## Coordinator

Use the GSD universal programming loop with Codex `gpt-5.5` and `xhigh` reasoning. Implement issue
#37 as a scoped sub-PR against `feat/44-github-cli-parity`. Preserve existing GitHub covered rows,
convert legacy exclusions to blocked operation-ledger metadata, and do not enable runtime execution
for newly modeled endpoints.

## Backend

Apply Go TDD. Add red tests for loader/schema support, connectorgen semantic validation, and GitHub
metrics before production code. Implement the minimal schema/types/validation changes needed for
`operation_ledger_version: 1`.

## Security

Confirm all operation-ledger rows are non-executable, blocked by default, and source-linked or noted
when sensitive/admin/destructive/disallowed. Do not add raw API, GraphQL, generic HTTP, SQL, shell,
or destructive/admin dispatch.

## Reviewer

Review the slice for exactly-one endpoint classification, preservation of existing covered rows,
validator/conformance parity, and command-runner safety.
