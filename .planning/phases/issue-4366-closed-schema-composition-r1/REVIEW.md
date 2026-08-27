# Review — #4366 closed-schema composition foundation

## Frozen review record

- Review range: `origin/main...6c9b4d50d284344a919317956a456c6857c39f19`
- Reviewer: Codex local code review
- Scope: source-import admission, source projection, engine validation,
  record-schema promotion, commandrunner parity, generated GitHub projection,
  and Batch 1 reconciliation evidence.
- Review route: a fresh independent Codex review of the exact post-policy-
  correction head is required before merge consideration. Do not wait for,
  request, or claim a Claude review.

## Finding and disposition

| ID | Finding | Disposition | Fix | Evidence |
| --- | --- | --- | --- | --- |
| R1 | An untyped `oneOf`/`anyOf`/`allOf` wrapper could carry `properties`, `items`, or other object/array-only siblings. Projection would not retain those siblings, risking source-semantic loss if it were admitted. | Accepted | `6c9b4d50d284344a919317956a456c6857c39f19` rejects that wrapper in source admission and projection, retaining a source-cited `typed_input_schema` gap instead. | New importer and projection regressions plus final `go test -timeout 20m ./cmd/connectorgen -count=1` passed. |
| R2 | The projection selected only the first declared type. A nullable composition wrapper expressed as `type: ["null", "object"]` therefore lost its object-level closure and required fields. | Accepted | `9fdf96dc4818c104cab5ac7455a79cfd18e575b6` converts every declared type kind and requires an explicit closed boundary for an object wrapper carrying composition. | New red/green regression proves `{}` is rejected and the declared object remains accepted; full `cmd/connectorgen` is recorded in `VERIFICATION.md`. |

## Prior exact-SHA Codex re-review

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

The original trusted-author automatic-review route is superseded. The PR stays
open while an independent Codex review covers the exact post-policy-correction
head. Record that SHA, the CI state, connector-shaped proof, remaining gap, and
every finding disposition in the PR before merge consideration. No Claude
review is awaited, requested, or claimed.
