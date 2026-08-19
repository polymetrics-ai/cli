# Issue 4015: Full-overwrite re-run correctness — Discussion Log

> Audit trail only. Decisions are captured in `CONTEXT.md`.

**Date:** 2026-08-18
**Mode:** `discuss-phase --auto`, inline/manual fallback

## Source refresh versus checkpoint resume

| Option | Description | Selected |
| --- | --- | --- |
| Source semantics own checkpoint eligibility | Full-refresh modes ignore prior positions; incremental modes resume | ✓ |
| Force all modes to re-read | Hides the overwrite symptom but breaks incremental contracts | |
| Special-case PostgreSQL overwrite | Repairs one route while leaving the shared defect | |

The user explicitly required the shared boundary and preservation of incremental skipping.

## Proof level

| Option | Description | Selected |
| --- | --- | --- |
| Exact binary counts plus independent destination query | Demonstrates the real state transition and stale-row removal | ✓ |
| Unit tests only | Does not prove the production data path | |
| Exit status only | Repeats the original false-success failure mode | |

## Scope

The task remains limited to full-refresh checkpoint eligibility, full-overwrite replacement proof, and incremental regression protection. No release-branch, PR #4250, runtime-daemon, or unrelated connector changes are included.

