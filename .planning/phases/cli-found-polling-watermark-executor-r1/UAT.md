---
status: complete
phase: cli-found-polling-watermark-executor-r1
source: SUMMARY.md coverage block
started: 2026-08-06T00:00:00Z
updated: 2026-08-06T00:00:00Z
mode: automated-coverage
---

## Current Test

[testing complete]

## Tests

### 1. Fail-closed CDC promotion

expected: The test-only bundle is CDC-capable only through its full matching executor.
result: pass
evidence: SUMMARY.md D1 automated unit coverage

### 2. Closed declaration and checkpoint contract

expected: Missing checkpoint fields and unknown declared streams fail before promotion.
result: pass
evidence: SUMMARY.md D2 automated unit coverage

### 3. Tie and late-arrival correctness

expected: Ties replay, safety lag overlaps a committed timestamp, and initial reads do not silently skip history.
result: pass
evidence: SUMMARY.md D3 automated unit coverage

### 4. Honest delete observability

expected: Only declared soft deletes or deletion endpoints emit tombstones; hard deletes remain not available.
result: pass
evidence: SUMMARY.md D4 automated unit coverage

### 5. Durable, bounded execution

expected: Checkpoints follow downstream acknowledgement and the executor replays safely while honouring limits and cancellation.
result: pass
evidence: SUMMARY.md D5 automated unit coverage

## Summary

total: 5
passed: 5
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

none
