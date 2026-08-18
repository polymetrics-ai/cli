---
status: complete
phase: issue-4072-github-app-auth-admission-r1
source: SUMMARY.md
updated: 2026-08-14
---

# UAT: GitHub App shared-rate admission

This internal engine change has no user-facing CLI or UI checkpoint. The
autonomous lane therefore uses its live observable integration proof as the
testable acceptance surface; no product judgment was required.

| Test | Expected observable result | Result |
|---|---|---|
| Missing/unreachable shared coordinator | GitHub token transport records zero sends and the typed unavailable error is recoverable | pass |
| Real shared coordinator | Two processes sharing budget one yield one mint, one exhausted-budget timeout, and one fixture POST | pass |

## Summary

total: 2

passed: 2

issues: 0

pending: 0
