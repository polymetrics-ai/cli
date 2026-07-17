# PROMPTS — Issue #404

## Kickoff snapshot

Task: execute issue #404 under parent #397 on branch `feat/404-redacted-slog` from base `20475ddf`, scoped to redacted stdlib slog foundation, per-run JSONL routing/retention, vault.Get redaction registry, Temporal structured logger bridge, focused tests, and issue-local planning artifacts.

Required command path attempted:

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 404 --skip-research
scripts/gsd prompt programming-loop init --phase 404 --dry-run
```

Downstream artifact: `.planning/phases/404-redacted-slog/PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `RUN-STATE.json`.

Verification result: initial implementation pushed; PR #455 review-fix implemented; requested non-extended gates passed; `verificationPassed=false` because extended full CLI race remains coordinator-pending; `programming-loop` adapter command missing, manual GSD fallback recorded.

## Manual-GSD fallback prompt in effect

Follow `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` manually: plan before coding, capture red tests before production edits, implement minimal green slices, run focused and full gates, commit/push coherent green checkpoints, keep issue-local artifacts current, and stop for human gates.

## Review-fix kickoff snapshot

Task: security/adversarial review-fix PR #455 issue #404 at `1cf8673b23b4b1b2b7aa82eaae49784d0d73586b`. No merge/deps/TTY/perf/OTel/parent edits. Accepted blockers: slog group recursion, typed URL redaction, invocation-scoped error redaction, run-log symlink/retention hardening, Temporal probe lifetime, run correlation, registry hardening, and single-line diagnostics.

Downstream artifact: PLAN/TDD/VERIFICATION/RUN-STATE/SUMMARY updated for review-fix before production edits.

Verification result: red tests captured, fixes implemented, requested gates passed; `verificationPassed=false`; extended full CLI race explicitly pending coordinator.

## Second security review-fix kickoff snapshot

Task: second security review-fix PR #455 issue #404 at `e27647806b44d40c09bccc1199e290c3054db452`. No merge/deps/TTY/perf/OTel/parent edits. Accepted findings: generic/raw URL fail-closed redaction, context-aware bounded Temporal dials and RLM probe ordering, slog group semantics, worker serve ready/start output seam, fail-closed `Any`, process-wide active log leases/retention, scoped/global registry hardening plus dynamic key/group caps, and bounded encoded-variant disposition.

Downstream artifact: PLAN/TDD/VERIFICATION/RUN-STATE/SUMMARY updated before test and production edits.

Verification result: requested second-review gates passed after outage recovery; extended full CLI race remains coordinator-owned/deferred and was not run by this worker.
