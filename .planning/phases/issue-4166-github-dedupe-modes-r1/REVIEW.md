# Code Review: GitHub dedupe modes r1

Manual inline `code-review` fallback: the delivery contract forbids spawning
the GSD reviewer/fixer roles in this worker environment.

## Reviewed scope

- Generator admission derivation and generated GitHub/PostgreSQL matrix shards.
- GitHub source-mode declaration, local warehouse apply strategies, replay
  dispatch, and source-page resume handling.
- Current dedupe/history integration tests and help/manual/website generated
  artifacts.

## Result

No actionable correctness, security, or scope findings remain.

The review checked that the shared runtime has no connector-name branch, that
the issue-label destination retains exactly its two declared actions, that the
replay exemption applies only to the two dedupe contracts, and that source
version identity excludes run-specific WAL metadata. `make lint`, the scoped
test suites, generated checks, and `connectorgen boundary` all pass.

## CI help-rendering follow-up

Manual inline `verify-work` and `code-review` follow-up found no actionable
correctness, security, or scope issue. The static ETL manual, rather than a
conditional mode list, had split the required history-mode description across
two output lines. The fix keeps that wording in the actual
`incremental_dedupe_history` entry, preserves the existing assertion, and
regenerates only the derived CLI manual and golden transcript. Direct built
binary help confirms the complete entry; the repeat website generator and the
full scoped gates pass.
