---
coverage:
  - id: D1
    description: Required zero-minimum string arrays distinguish explicit presence from omission.
    verification:
      - kind: unit
        ref: internal/connectors/commandrunner/runner_test.go:TestValidateRequiredCommandFlagsPreservesStringArrayPresence
        status: pass
    human_judgment: false
  - id: D2
    description: Operation direct-read bodies retain an explicit empty array.
    verification:
      - kind: integration
        ref: internal/connectors/commandrunner/runner_test.go:TestRunOperationDirectReadPreservesExplicitEmptyRequiredStringArray
        status: pass
    human_judgment: false
  - id: D3
    description: Reverse-ETL plans retain an explicit empty array while keeping approval and avoiding execution.
    verification:
      - kind: integration
        ref: internal/connectors/commandrunner/runner_test.go:TestBuildWriteCommandPreservesExplicitEmptyRequiredStringArray
        status: pass
    human_judgment: false
---

# SUMMARY — issue #3775 presence-aware required string arrays

Implemented parent #3775 and its ordered child scope #3778, #3780, #3781, and #3783 in the shared
command runner.

- Required flags still reject an absent raw map key or present zero-length raw value slice.
- An explicit blank / blank-only CSV for a required zero-minimum `string_array` is now supplied,
  materializes as a non-nil empty `[]string{}`, and serializes as literal `[]`.
- Required scalar blanks, `min_items`, and `max_items` behavior remain unchanged.
- Both the operation direct-read body path and reverse-ETL plan record path preserve `[]`; the
  reverse-ETL test retains approval-required semantics and proves no provider write occurs.
- No provider bundle/schema/capability/output-policy/redaction path, credential, live request,
  retry, or reverse-ETL execution was introduced.

GSD discuss/plan/execute/verify/code-review prompts were resolved and executed inline under the
required no-spawn fallback. Full evidence is in `TDD-LEDGER.md`, `VERIFICATION.md`, and `REVIEW.md`.
