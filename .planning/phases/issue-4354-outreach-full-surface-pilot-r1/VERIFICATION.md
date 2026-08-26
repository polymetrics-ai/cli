# Verification checklist — #4354

- [ ] Candidate-to-main delta contains only seven Outreach artifact paths before any local correction.
- [ ] Source lock, declaration disposition, API surface and CLI surface reconcile on every captured operation.
- [ ] Six-lane counts/evidence are recorded, including explicit zero/not-applicable binary lanes when warranted.
- [ ] Existing source-evidence validation result is captured; any schema-v3 block identifies PR #4350 and is not bypassed.
- [ ] Focused red/green tests are recorded with observable assertions.
- [ ] `pm` builds and representative valid ETL, direct-write/reverse-ETL, and delete paths reach `missing --credential` through fixture transport.
- [ ] Wrong source identity/method/path is rejected pre-request.
- [ ] Connector validation, surface sync, certification/evidence and scoped tests pass or are truthfully reported as blocked.
- [ ] CLI help/manual/website parity is verified or marked inapplicable with evidence.
- [ ] `git diff --check` and independent clean-worktree usability pass.
- [ ] PR base API readback equals `main`; PR has a Conventional Commit title and complete delivery record.
