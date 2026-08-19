---
status: complete
phase: issue-3852-output-policy-declaration-r1
source: inline automated verification (no SUMMARY.md; task authorizes autonomous execution)
started: 2026-08-05T22:32:44Z
updated: 2026-08-05T22:32:44Z
---

## Current Test

[testing complete]

## Tests

### 1. Declare and execute complete direct-write JSON

expected: A bundle with an implemented `direct_write` command, matching POST `api_surface`,
and `output_policy: "json"` loads; the existing executor returns the complete decoded JSON body.
result: pass
evidence: `go test ./internal/connectors/engine -run '^TestBundleLoadAcceptsRuntimeSupportedNonRedactingDirectWriteJSONPolicy$' -count=1`
human_judgment: false

### 2. Keep declarable and runtime policy sets synchronized

expected: The CLI-surface output-policy enum is exactly the direct-read/direct-write runtime union
plus the legacy binary-download compatibility policy, with no duplicate entries.
result: pass
evidence: `go test ./internal/connectors/commandrunner -run '^TestCLISurfaceOutputPolicyEnumMatchesRuntimePolicySets$' -count=1`
human_judgment: false

### 3. Preserve fleet compatibility

expected: All existing connector bundles still validate against the reconciled schema.
result: pass
evidence: `go run ./cmd/connectorgen validate` — 550 connectors checked, 0 findings
human_judgment: false

## Summary

total: 3
passed: 3
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

None.

## Execution Note

`verify-work` was completed inline because the issue contract forbids role spawning and this
non-UI schema/runtime foundation has no judgment-dependent acceptance criterion. The task directs
autonomous work; all three deliverables have directly observable automated evidence.
