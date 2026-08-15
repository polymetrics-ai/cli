# VERIFICATION — #3865/#3867 durable production wiring

Status: focused implementation evidence green; repository gates pending.

- [x] Auth state survives SIGKILL and a new CLI process refuses before I/O.
- [x] Dedicated credential test repairs the same cohort and advances epoch.
- [x] Fresh repaired ownership proceeds; stale and unrepaired ownership are typed refusals with zero I/O/checkpoint mutation.
- [x] Park/reset/committed checkpoint survive SIGKILL and reopen.
- [x] Due work resumes through `App.RunETL` without replaying acknowledged work.
- [x] Cross-process same-cohort/scope races have one winner and zero loser side effects.
- [x] Cancellation, connection death, empty/single/large, duplicate/out-of-order,
      schema drift, permission/auth refusal, interrupted resume, and replay rows pass.
- [x] Real PostgreSQL `databaseintegration` evidence passes with Docker/Colima.
- [ ] Focused normal/race tests, vet, build, and individual repository gates pass.
- [ ] Derived artifacts regenerated once and all drift checks pass with clean status.
- [ ] GSD verify-work and deep code-review complete with findings dispositioned.
- [ ] PR is open and API reports base `integration/4015-mvp-flat-r1`.
