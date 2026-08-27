# Review — #4366 closed-schema composition foundation

## Frozen review record

- Review range: `origin/main...6c9b4d50d284344a919317956a456c6857c39f19`
- Reviewer: Codex local code review
- Scope: source-import admission, source projection, engine validation,
  record-schema promotion, commandrunner parity, generated GitHub projection,
  and Batch 1 reconciliation evidence.
- Route after PR open: `claude_auto` for the trusted-author, main-targeted PR;
  fallback `none` unless the automation fails or is skipped.

## Finding and disposition

| ID | Finding | Disposition | Fix | Evidence |
| --- | --- | --- | --- | --- |
| R1 | An untyped `oneOf`/`anyOf`/`allOf` wrapper could carry `properties`, `items`, or other object/array-only siblings. Projection would not retain those siblings, risking source-semantic loss if it were admitted. | Accepted | `6c9b4d50d284344a919317956a456c6857c39f19` rejects that wrapper in source admission and projection, retaining a source-cited `typed_input_schema` gap instead. | New importer and projection regressions plus final `go test -timeout 20m ./cmd/connectorgen -count=1` passed. |

## Fresh exact-SHA Codex re-review

Reviewed `6c9b4d50d284344a919317956a456c6857c39f19` after R1 was fixed.

- No further actionable correctness, source-provenance, pre-I/O validation,
  lane-promotion, generic-body, or generated-artifact findings.
- `git diff --check`, final full `cmd/connectorgen`, engine, commandrunner,
  source-import/check, validate, surface-sync, docs/website generation, and
  applicable direct-PR gates are recorded in `VERIFICATION.md`.
- The 608-row fixture remains `0` credential-bound runnable and `608`
  `missing_foundation`; the GitHub generated-schema correction changes no
  command availability or execution lane.

## External review status

Pending PR creation. After opening, inspect the actual Claude workflow and
comments for this commit range; record the PR URL, review SHA, coverage route,
and every disposition here before merge consideration.
