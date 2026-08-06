# UAT — issue #3755 rate-limit operator output

Automated acceptance is sufficient for this deterministic CLI/runtime slice; no live provider or browser judgment is required.

- PASS — `TestRunETLPersistsDeclaredRateLimitSummaryFromTestBundle` runs the declared fixture through `App.RunETL`, verifies the persisted structured result, and verifies the human-line rendering.
- PASS — `TestRateLimitReportShowsDeclaredPolicyPacingAndProviderPushbackWithoutSecrets` proves local pacing, a 429, its honoured retry wait, provider remaining budget, and prohibited identity-value absence on both output representations.
- PASS — `TestETLRunRateLimitOutputIsStructuredHumanReadableAndSecretFree` verifies the actual CLI's text and JSON ETL output uses `undeclared` rather than implying unlimited traffic.
- PASS — `TestRateLimitReportCoalescesLongRunsIntoBoundedPolicySummary` proves fixed-size policy rows and scalar aggregates under repeated requests.
