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

## Current-main integration — inbox 001

- Authorization: Captain message `001.msg`, processed 2026-08-27.
- Target: merge `origin/main` at
  `1324c52bab0b224ed8958858af7676b8b8e191b4` into this PR branch without a
  force push. This is not a merge of the PR into `main`.
- Reason: #4351 is now in main and must be integrated with the shared
  write-disabled mutation-artifact foundation before review.
- Inline GSD fallback: the canonical single-worker contract forbids lifecycle
  role spawning, so discuss/plan/execute/verify/review are recorded in this
  phase evidence and executed inline.
- Required proof: run the full preserved Sentry/Vercel source-lock inventory
  classification and real selected read/mutation coverage after integration;
  also rerun the executable delete/reverse-ETL control. Report every merge
  conflict, use no credential, and obtain a fresh separate Codex audit of the
  pushed integration SHA.

## Audit repair M1 — inbox 002

- Audit target: PR #4357 head
  `e6418bb1dc50001eefc438d4fb8c62e441de25c9`; independent Codex audit found
  that a missing nested `capabilities.write` JSON member decoded as Go `false`
  and incorrectly admitted the automatic artifact.
- Delivery: repair the existing main-targeted PR with a normal push only. Do
  not alter connector source locks, generated connector definitions, user
  commands, or the preserved worker worktrees; do not merge.
- Inline GSD fallback: the canonical single-worker contract forbids lifecycle
  role spawning. The resolved `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review` prompts are executed and
  recorded inline.
- Red: add an adversarial missing-member test that demonstrates the former
  zero-value admission through actual metadata decoding and source-artifact
  classification. It must observe no automatic artifact and a source coverage
  failure after the repair, while explicit `write:false` remains green.
- Green: retain the declaration-presence bit alongside the decoded boolean;
  automatic artifact derivation and automatic-artifact validation require the
  explicit bit plus `write:false`. The metadata schema rejects the missing
  member on full bundle loads. `write:true`, provider citation, complete action,
  and implemented-action precedence remain unchanged.
- Proof: focused engine/source-import/source-projection tests; full
  `cmd/connectorgen`, engine, commandrunner, and CLI suites; generator and
  applicable repository gates; `go build ./cmd/pm`; source-lock Sentry/Vercel
  inventory plus real delete/reverse-ETL controls; `git diff --check`. No new
  implemented Sentry/Vercel `pm` command exists, so their eventual
  credential-boundary proof remains downstream. The existing source-cited
  GitHub delete control is the connector-shaped binary witness for the shared
  command foundation. Obtain a fresh separate Codex audit of the exact pushed
  repair SHA.

## Current-main evidence — inbox 003

- `git fetch origin main` resolved `origin/main` to
  `1324c52bab0b224ed8958858af7676b8b8e191b4`. It was already the merge base
  of repair head `36ac3b0e8c424efda8a73345932af57138526a49`, so integration is
  a clean no-op: no force, reset, stash, conflict resolution, or discarded
  work was needed.
- Re-ran the exact M1/Sentry/Vercel/delete/citation/complete-action/
  certification focused suite across connectorgen, engine, and commandrunner;
  all passed. Rebuilt `./pm`, initialized a fresh project without a credential,
  and ran the existing source-cited GitHub `label delete --name bug` control.
  It exited 1 at `missing --credential`, proving dispatch reaches the
  credential boundary without a provider request. It is not a claim that this
  foundation materialized a Sentry or Vercel user command.
- The temporary 12 KB proof project was moved to Trash (recoverable). Push the
  final evidence checkpoint normally, then obtain a fresh independent audit of
  that resulting exact head. Do not merge the PR.
