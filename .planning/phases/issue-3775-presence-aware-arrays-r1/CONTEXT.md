# CONTEXT — issue #3775 presence-aware required string arrays

## Phase mapping

Parent issue #3775 integrates dependency-ordered slices #3778, #3780, #3781, and #3783 in one
branch and one eventual PR, as required by the assigned task. This is a shared command-runner
foundation, not a provider migration: there is no target connector and no connector bundle,
schema, capability, availability, docs, or generated-surface change in scope.

## Locked decisions

- A required flag has two independent conditions: raw CLI presence and materialized cardinality.
  The common validator must still reject an absent flag-map key and a present zero-length raw value
  slice before coercion.
- `string_array` coercion intentionally trims comma-separated/repeated values and drops blank
  entries. Therefore `--items ""` and blank-only CSV may materialize as literal `[]`.
- A materialized empty `[]string` is supplied for a required `string_array` when `min_items` is
  zero or omitted. Existing `min_items` and `max_items` enforcement in coercion remains the sole
  cardinality authority.
- Required scalar strings remain invalid when blank. `allow_empty` remains scalar-string-only;
  no schema field is added and `cmd/connectorgen/validate.go` is not changed.
- The real common validator must govern both operation-backed direct reads and reverse-ETL record
  construction. The reverse-ETL test stops at plan construction and verifies approval remains
  required; it never executes a write.
- The literal empty array is non-secret output/data. Do not redact, strip, mask, or replace it.
- The current task deliberately combines the four issue slices in one branch/PR rather than
  opening stacked child PRs. This task-specific topology overrides the normal parent/subissue
  branch fan-out; all issue references remain in the delivery record.

## Scope and ownership guard

- Owned production functions in `internal/connectors/commandrunner/runner.go`:
  `validateRequiredCommandFlags`, `commandValueEmpty`, `coerceFlagValue`,
  `operationDirectReadOverrides`, and `recordOverrides`.
- Owned focused tests: `internal/connectors/commandrunner/runner_test.go`.
- Do not edit the redaction-owned functions, `validateOperationDirectReadCommand`, connector
  bundles/docs, `cmd/connectorgen`, generated artifacts, or CLI parsing.
- No live provider, credential, or reverse-ETL execution is permitted. Tests use the existing
  in-memory fake connector only.

## GSD execution note

Adapter evidence: `scripts/gsd doctor` passed and `scripts/gsd sources` resolved
`discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`.
`go run ./cmd/agentcontractgen check` passed. The required GSD prompts are executed inline because
this task and the canonical delivery contract prohibit spawned roles. The issue tree already fixes
all behavior decisions, so the discuss phase records those decisions rather than reopening them.

## CLI help/docs/website parity disposition

This changes how already-declared `string_array` flags materialize at runtime; it adds no command,
flag declaration, help text, output format, manual topic, generated help, completion, or website
documentation. Runtime/help/docs/website artifact updates are intentionally not applicable. Tests
will prove the actual body/record representation and existing error shape instead.
