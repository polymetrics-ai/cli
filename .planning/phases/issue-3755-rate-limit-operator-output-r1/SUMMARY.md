---
coverage:
  - id: D1
    description: A declared test-only bundle produces a persisted ETL rate-limit summary with declared policy identity, structural selection reason, provider budget, and separate latency.
    verification:
      - kind: integration
        ref: internal/app/rate_limit_output_test.go:TestRunETLPersistsDeclaredRateLimitSummaryFromTestBundle
        status: pass
      - kind: unit
        ref: internal/connectors/engine/rate_limit_runtime_test.go:TestRateLimitReportShowsDeclaredPolicyPacingAndProviderPushbackWithoutSecrets
        status: pass
    human_judgment: false
  - id: D2
    description: Human and JSON ETL output report undeclared honestly and never include a credential value.
    verification:
      - kind: integration
        ref: internal/cli/rate_limit_output_test.go:TestETLRunRateLimitOutputIsStructuredHumanReadableAndSecretFree
        status: pass
    human_judgment: false
  - id: D3
    description: Report output is bounded and cannot include a raw binding, scope projection, runtime subject, credential revision, or secret value.
    verification:
      - kind: unit
        ref: internal/connectors/rate_limit_report_test.go:TestRateLimitReportCoalescesLongRunsIntoBoundedPolicySummary
        status: pass
      - kind: unit
        ref: internal/connectors/engine/rate_limit_runtime_test.go:TestRateLimitReportShowsDeclaredPolicyPacingAndProviderPushbackWithoutSecrets
        status: pass
    human_judgment: false
---

# SUMMARY — issue #3755 rate-limit operator output

- Added a concurrency-safe, bounded `RateLimitSummary` to ETL runs. Each connector reports `declared`, `unknown`, `not_applicable`, or `undeclared`; missing `rate_limits.json` is explicitly `undeclared`, never unlimited.
- Declared policies expose only policy ID, subject kind, structural selection reason, and typed provider limit/remaining/reset facts. Local pacing, provider-429 retry waits, and ordinary request latency are aggregated separately.
- The requester and limiter report observation timing without changing selection, admission, reservation, reset, retry, or request behavior. Both buffered and streaming requester paths are covered.
- `pm etl run` and `pm etl status` render compact human lines; their existing JSON envelopes include the same `run.rate_limit` object.
- Added test-only declared-bundle ETL coverage, absent-declaration coverage, bounded long-run coverage, 429/retry coverage, and both output surfaces' secret-free regressions. No production `rate_limits.json` or dependency was added.
- Updated the ETL manual, generated CLI docs/golden transcript, website page, and generated website docs data.

All scoped tests, race coverage, build/vet, docs generation/checks, and non-monolithic repository gates passed. Full-suite `go test ./...` and `make verify` remain CI-owned under repository timeout guidance.
