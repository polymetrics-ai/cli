# Plan — issue #4329 source-cited read-only mutation artifacts, r2

## Task Delivery Header

- Issue: Closes #4329 — allow source-cited read-only connectors with mutation artifacts
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, with the API-reported base equal to `main`, requested evidence committed, fresh independent audit recorded, and required CI checks green. No merge is performed.
- Working branch: fix/4329-read-only-mutation-artifacts
- Task: Derive exact source-cited non-executable mutation artifacts during source import for an intentionally write-disabled connector, while preserving usable source-backed read/ETL commands and refusing to conceal a real executable write action.
- Verification: TDD focused source-import/projection/coverage tests; engine and commandrunner tests; serial generator/validation gates; `go build ./cmd/pm`; credential-free built-binary probes for affected implemented commands when available from this branch; `git diff --check`; PR CI and audit.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A write-disabled source bundle retains each real provider mutation as a cited non-executable artifact | live | A retained-source import fixture emits `runtime.non_executable_mutation`, its exact source ID/method/path/citation, and the named source-cited foundation gap. |
| Read/ETL commands still materialize and validate | live | The same source-backed fixture retains its implemented GET/stream route while projection and executable coverage report no findings. |
| Sentry and Vercel remain representative source vectors | live | Distinct Sentry and Vercel source-shaped fixtures cover their locked source identities and mutation methods without connector-name/count shortcuts. |
| A real action is never suppressed | live | Complete delete/reverse-ETL action and implemented command controls remain executable and reject/avoid the automatic artifact. |
| No generic or fabricated write surface appears | live | Projection bytes for writes/CLI remain unchanged and no generated action, command, request schema, transport, or partial status is asserted. |

## TDD execution slices

1. **Red:** Add source-import/projection/coverage tests for a write-disabled locked source with a supported read plus mutation, using separate Sentry and Vercel-shaped cases. Record the pre-implementation failure.
2. **Green:** Add the smallest source-import annotation helper. It may only generate the existing exact non-executable mutation artifact when metadata declares no write capability; retain the named foundation/citation contract.
3. **Safety:** Add controls for complete delete/reverse-ETL actions, implemented incomplete claims, write-capable bundles, GraphQL mutations, and byte stability of writes/CLI output.
4. **Proof:** Run focused `cmd/connectorgen`, engine, and commandrunner tests; source import/check/validate/surface sync gates; build the binary and probe every newly implemented command if any; then run the serial applicable gate set and `git diff --check`.
5. **Review:** Execute inline `verify-work` and code review. Push the exact green SHA, open a main-targeted PR, read its API base, obtain an independent Codex audit of that SHA, and wait for all required CI checks. Do not merge.

## Skills and CLI parity

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, and
`golang-documentation`.

This changes the developer-facing `connectorgen` import/validation contract,
not a new `pm` command, flag, help topic, manual, or website page. Generated
surface/docs checks are applicable; user help/manual/website changes are not
unless the implementation materially changes the embedded `pm` surface.

## Scope guard

Only shared `cmd/connectorgen` import/projection/coverage behavior, its tests,
and this issue's evidence may change. The preserved Batch 1 Sentry/Vercel
worktree, source-lock bytes, connector declarations, write actions, and active
#4351/#4356 repair worktrees are not modified.
