---
phase: issue-4067-acknowledged-completion-rebase-r1
issue: 4067
status: r3_local_validation_review_passed_pending_loop_4
---

# #4067 R2 summary

Sol r2 rejected `3f84693bfbc128523a66e22653db7227fb9c0869` because a post-acknowledgement cancellation or source error could discard the durable acknowledgement witness. A later unrelated writer then made ordinary `failRun` reject the stale revision, returning `Run{}` while the durable run remained `running`. It also found an acknowledged completion rebase with a removed target run returned an untyped ordinary missing-run error.

The r2 correction preserves a result witness only after the real checkpoint acknowledgement. A separate finalizer then reads current state and terminalizes only the exact still-running run whose exact acknowledged stream remains present; it preserves every other entry and joins the original error with any persistence error. The completion-only missing target is now a typed revision conflict. No generic state refresh, checkpoint retry/overwrite, destination replay, #4046 behavior change, or provider/warehouse work was introduced.

TDD evidence includes a committed all-seven-mode behavioral RED for both durable symptoms, then green cancellation/source-error/missing-target/outcome tests. Review added direct changed/removed stream and terminal/removed run error-finalization coverage. Focused race, #4046/R7/R8, affected package, vet/build, individual repository gates, and read-only website checks are locally green; `make smoke-no-build` is intentionally omitted because it mutates a warehouse.

The phase uses the documented inline/manual GSD fallback because its named issue phase is absent from the numeric roadmap and this custody lane forbids role spawning. Manual execution, verification, UAT, and code review records are current in this directory. Fresh local-only loop 3/5 passed review, targeted persisted-state/race testing, documentation, and lint with push/PR/CI skipped. Its document agent added an unrelated warehouse-architecture rewrite; this scope-restoration follow-up cancels it while retaining the completed run in history. The correction budget is 3/5. Current help has no safe #4059-only delivery route, so Firstmate direction, exact-head CI, and independent Sol audit remain pending. No certification or merge readiness is implied.

## R3 stale second-page finalization

Sol r3 found the remaining mask: a page-one acknowledgement witness was non-empty, so a later typed page-two stale-writer conflict was routed through the ordinary acknowledged-error guard. A real concurrent winner therefore made that old witness fail before #4046's typed-conflict terminalizer ran, leaking a zero returned run and durable `running` loser.

The r3 correction is one typed-conflict-only early branch in `failAcknowledgedTransportRun`: it delegates `errTransportStreamStateConflict` directly to the established #4046 `failRun` path before consulting the old witness. The committed all-seven two-page RED reproduced the leak using real JSON persistence, two Apps, a real winner, unrelated state, and destination applies. GREEN proves a non-zero returned/reopened failed loser, typed conflict identity, winner/unrelated preservation, and exactly two loser applies without retry or replay. Normal and race repeats, the focused #4046/r2/R7/R8 suite, and full `internal/app` are locally green.

The added existing-connector gate is separate and intentionally limited: the real binary builds; GitHub definition/hook/preflight/CLI/harness tests and inspection pass; inspection reports GitHub and PostgreSQL source/destination transport roles `unsupported`, `certified: false`, and `COMMUNITY BUILD, UNCERTIFIED`. No approved credential/name or sanctioned secret-channel invocation was supplied, so no credentialed provider smoke was attempted or substituted. This branch does not provide production registration, a GitHub-to-DuckDB/Parquet round trip, certification, or #4015 wiring. Inline GSD review found no unresolved issue. The budget remains 3/5 until fresh local-only no-mistakes loop 4/5 is submitted with push/PR/CI/document skipped; no external delivery action is authorized afterward.
