<!-- Local PR issue-guard preflight body only. It has not been published. -->

## Summary

Preserve unrelated generic SQL availability when two exact-unique warehouse
table names collide under DuckDB's canonical identifier rules. Unscoped flow
reads now return the existing typed warehouse ambiguity and connection remedy.

## Linked Issue

Refs #4069

## Stacked PR

- Parent issue: #3897
- Parent branch: feat/3897-flow-connection-scope-nm
- PR base branch: feat/3897-flow-connection-scope-nm
- Sub-issue: #4069

## Verification

- Committed real Parquet/DuckDB RED and minimum GREEN evidence.
- Full affected packages and focused race matrix passed after the transport CPU
  gate cleared.
- GSD manual verify-work/code-review, vet/lint, canonical generators,
  docs/help checks, and candidate diff check passed.
- The inherited seven-line #3897 Markdown whitespace finding remains outside
  this child range.

## Pipeline

No no-mistakes run, PR creation, or CI observation has occurred. Those
deliberately remain pending Firstmate coordination after transport Sol r4.

## Safety

No credentials, provider calls, transport registration, production DuckDB or
Parquet wiring, reverse-ETL dispatch, or connector certification claim is
included.
