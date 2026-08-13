# #4077 — execution record

## Status

The GREEN implementation passes the focused regression. The next step is broader package and race verification.

## Next action

Run focused race and package tests, then complete the verification/review phase artifacts.

## RED outcome

`json.RawMessage` and `map[string]string` were still returned by the `cloneRecordValue` default path
and therefore shared source-owned storage. A `map[string]int` also crossed each boundary without
rejection. The failure is behavioral, not a compile/setup failure.

## GREEN implementation

`cloneRecordValue` now enumerates the closed JSON-like scalar and mutable-value contract. It explicitly
copies `json.RawMessage`, `[]byte`, `map[string]string`, `map[string]any`, `[]any`, and
`[]connectors.Record`; every unknown value returns an error that the orchestrator wraps before it crosses
the source→stage or stage→destination boundary.
