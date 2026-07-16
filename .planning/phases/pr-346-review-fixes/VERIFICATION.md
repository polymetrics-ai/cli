# Verification Checklist

- [x] Targeted bookmark collision test is red before the migration/query fix.
- [x] Targeted rejected-delete test is red before provider recovery changes.
- [x] Targeted hydration test is red with an immediate session response.
- [x] Targeted tests pass after implementation.
- [x] Website typecheck passes.
- [x] Website unit tests pass: 10 files, 66 tests.
- [x] Website end-to-end tests pass against disposable PostgreSQL: 25 tests, one worker.
- [x] Website production build passes: 1,121 generated pages.
- [x] Podman production image builds without Better Auth default-secret errors.
- [x] The production image runs on port 3100 with ignored runtime env.
- [x] Blog, docs, GitHub OAuth initiation, comments, bookmarks, and migrations pass.
- [x] Production-container browser smoke passes: 12 tests.
- [x] `make verify` passes, including 547 connector definitions and 0 lint findings.
- [ ] Parent and stacked PR current-head CI/review status is rechecked after pushing fixes.
