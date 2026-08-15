# VERIFICATION — #3865/#3867 durable production wiring

Status: planned.

- [ ] Auth state survives SIGKILL and a new CLI process refuses before I/O.
- [ ] Dedicated credential test repairs the same cohort and advances epoch.
- [ ] Park/reset/committed checkpoint survive SIGKILL and reopen.
- [ ] Due work resumes through `App.RunETL` without replaying acknowledged work.
- [ ] Cross-process same-cohort/scope races have one winner and zero loser side effects.
- [ ] Cancellation, connection death, empty/single/large, duplicate/out-of-order,
      schema drift, permission/auth refusal, interrupted resume, and replay rows pass.
- [ ] Real PostgreSQL `databaseintegration` evidence passes with Docker/Colima.
- [ ] Focused normal/race tests, vet, build, and individual repository gates pass.
- [ ] Derived artifacts regenerated once and all drift checks pass with clean status.
- [ ] GSD verify-work and deep code-review complete with findings dispositioned.
- [ ] PR is open and API reports base `integration/4015-mvp-flat-r1`.
