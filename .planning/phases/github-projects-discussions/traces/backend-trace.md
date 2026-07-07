# Agent Trace: backend

> DRAFT — populated from phase artifacts during the review-fix slice. This trace should be
> cross-checked against actual run logs before the phase is marked complete.

## Rendered Prompt Or Prompt Reference

docs/prompts/universal-programming-loop-prompts.md (migration executor template), .planning/phases/github-projects-discussions/PROMPTS.md

## Files Inspected

internal/connectors/engine/bundle.go, internal/connectors/engine/graphql.go, internal/connectors/engine/read_test.go, internal/connectors/engine/bundle_test.go, internal/connectors/defs/github/operations.json

## Actions Taken

Added GraphQL query.* namespace, omit_when_empty support, default/type validation, DraftIssue fragment, and corresponding unit tests.

## Commands Run

go test ./internal/connectors/engine ./cmd/connectorgen -count=1; go run ./cmd/connectorgen validate internal/connectors/defs --json

## Findings

All engine tests pass; connectorgen validate reports no GitHub findings.

## Handoff Summary

Engine/bundle changes verified and committed.

## Verification Evidence

TDD-LEDGER.md red/green entries and VERIFICATION.md engine/connectorgen rows.

## Unresolved Risks

None remaining after local gates.
