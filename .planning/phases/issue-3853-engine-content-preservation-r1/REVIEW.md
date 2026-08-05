# REVIEW — issue #3853 engine content preservation

Mode: inline `code-review` fallback. The canonical single-worker task contract prohibits spawning
the workflow's reviewer role.

## Scope reviewed

- `internal/connectors/engine/write.go` and its reversed preview coverage in `write_test.go`
- `internal/connectors/engine/direct_read.go`, `binary_read.go`, and their `httptest` coverage
- `internal/cli/docs.go`, generated `docs/cli/reverse.md`, golden transcripts, and the reverse-ETL
  website documentation
- This phase's plan, TDD, verification, and coverage artifacts

## Findings

No actionable correctness, security, or quality findings.

### Correctness

- The preview uses the unchanged runtime config/secrets and record only for interpolation; it does
  not mutate the caller record. Existing test coverage retains top-level and nested `redact_fields`
  declarations to prove declaration compatibility without preview substitution.
- The direct-read helper uses typed `connsdk.HTTPError` fields that have already been bounded by the
  transport. It preserves URL/query/body diagnostics even though `HTTPError.Error()` applies a
  presentation policy; non-HTTP errors keep their original text.
- Error-map classification and hint ordering remain exactly in the existing call sites. Binary
  failure handling still returns before opening a destination file, covered by the retained cleanup
  assertion.

### Safety and security

- Response size, timeout, redirect, filesystem confinement, output-policy, and credential-storage
  safeguards are unchanged. This issue removes only the engine masking that caused the approval
  preview and error diagnostics to omit content.
- No generic write tool, capability claim, command declaration, bundle rewrite, provider call,
  credential, or reverse-ETL execution was introduced.

### Scope

- The diff does not modify #3771-owned command-runner functions, the #3852 enum, `connsdk`,
  successful-output redaction, binary result-record redaction, or generic source-table masking.
- The real runtime/declaration checks passed: connector validation found 0 findings across 550
  bundles and surface sync found 0 drift, so the change cannot mint an unsupported implemented
  command.

## Dispositions

None required; the inline review found no actionable item.
