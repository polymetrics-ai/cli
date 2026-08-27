# Plan — #4366 closed-schema composition foundation

## Objective

Replace the source projection's composition gap-only behavior with a shared closed typed-schema conversion, backed by the engine compiler and pre-I/O request validation. Preserve source mapping; do not promote an operation merely because its request schema becomes representable.

## Slice 1 — red: contract and safety boundaries

1. Add failing source-import/projection tests for nested local-reference `oneOf`/`anyOf`/`allOf`, nullable scalar unions, discriminated object alternatives, compatible and contradictory intersections, duplicate/ambiguous alternatives, and malformed/external/cyclic references.
2. Add failing engine schema tests that prove exact `oneOf`, `anyOf`, and `allOf` selection, and no-match/multi-match/contradiction rejection.
3. Add a failing deferred-command test with a no-I/O transport spy: a source-cited composition operation with another absent foundation remains `missing_foundation`; no operation/write/stream/binary binding is generated.
4. Add the failing 608-row manifest reconciliation, preserving `record_key`, connector, lane, source URL/SHA/location, source method/path, canonical target, intended command path, and composition disposition.

Red: run the new test names before production code and record their expected failures in `traces/red-composition.txt`.

## Slice 2 — green: common closed composition representation

1. Introduce the smallest source-projection conversion that recursively preserves closed schemas, named composition arms, nullable source-backed scalar types, required fields, discriminator metadata where supported, and local-reference closure from the existing resolver.
2. Extend the engine typed schema compiler/validator with separate `oneOf`, `anyOf`, and `allOf` semantics. Reject non-object arms, zero arms, duplicate equivalent arms, impossible intersections, open/dynamic object fallbacks, and unrecognised keywords.
3. Thread composition failure as an exact source contract gap with source path context; keep all affected descriptors' provenance and connector-relative route unchanged.

Green: focused importer/projection and engine tests pass. Record the commands and result in `TDD-LEDGER.md`.

## Slice 3 — admission/evidence and Batch 1 reconciliation

1. Ensure source projection, declaration admission, and operation evidence retain every source operation and its exact foundation disposition.
2. Allow an operation to be runnable only through the existing lane-specific admission and real commandrunner preflight; do not auto-create a runtime declaration.
3. Reconcile all 608 Batch 1 composition rows. For each non-promoted record, preserve the exact source mapping and one `missing_foundation` disposition. Record the promoted and retained counts.
4. Add representative source-locked regression cases from Bitbucket, CircleCI, GitLab, Jira, and Vercel.

## Slice 4 — refactor, generated artifacts, and delivery

1. Run gofmt and targeted tests. Run source-import/check, connector validation, and `surface-sync --check` for every touched connector, plus generated artifact checks.
2. Build `pm`. For every promoted command, initialise an isolated project and assert the exact credential boundary with no provider I/O. If no command is promoted, record `0` with the admission reason rather than inventing a proof.
3. Run applicable direct-PR checks, review routing, frozen finding disposition, and a fresh exact-SHA Codex re-review; update verification and PR body.

## Non-goals

- No provider source fetch, provider credential, live request, generic route/body, permissive object, flattened union, inferred field, or broad generated connector rewrite.
- No execution-lane claim from HTTP method or schema composition alone.
- No changes to the active Stripe descriptor, deferred-visibility, Sentry route override, or connector materialization lanes except conflict-safe rebase after a committed checkpoint.
