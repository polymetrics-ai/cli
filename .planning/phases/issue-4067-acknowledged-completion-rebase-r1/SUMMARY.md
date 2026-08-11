---
phase: issue-4067-acknowledged-completion-rebase-r1
issue: 4067
status: r2_local_delivery_pending_no_mistakes
---

# #4067 R2 summary

Sol r2 rejected `3f84693bfbc128523a66e22653db7227fb9c0869` because a post-acknowledgement cancellation or source error could discard the durable acknowledgement witness. A later unrelated writer then made ordinary `failRun` reject the stale revision, returning `Run{}` while the durable run remained `running`. It also found an acknowledged completion rebase with a removed target run returned an untyped ordinary missing-run error.

The r2 correction preserves a result witness only after the real checkpoint acknowledgement. A separate finalizer then reads current state and terminalizes only the exact still-running run whose exact acknowledged stream remains present; it preserves every other entry and joins the original error with any persistence error. The completion-only missing target is now a typed revision conflict. No generic state refresh, checkpoint retry/overwrite, destination replay, #4046 behavior change, or provider/warehouse work was introduced.

TDD evidence includes a committed all-seven-mode behavioral RED for both durable symptoms, then green cancellation/source-error/missing-target/outcome tests. Review added direct changed/removed stream and terminal/removed run error-finalization coverage. Focused race, #4046/R7/R8, affected package, vet/build, individual repository gates, and read-only website checks are locally green; `make smoke-no-build` is intentionally omitted because it mutates a warehouse.

The phase uses the documented inline/manual GSD fallback because its named issue phase is absent from the numeric roadmap and this custody lane forbids role spawning. Manual execution, verification, UAT, and code review records are current in this directory. The correction budget has two completed no-mistakes runs; fresh loop 3/5, a safe #4059-only delivery route, exact-head CI, and independent Sol audit remain pending. No certification or merge readiness is implied.
