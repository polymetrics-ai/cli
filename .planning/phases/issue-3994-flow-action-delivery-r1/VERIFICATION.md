# Verification checklist — Issue #3994

- [x] Targeted flow, app, and CLI action tests are green: scoped full suites passed.
- [x] Changed scope proves zero provider sends: `TestConnectorFlowActionRunnerScopeDriftStopsBeforeTargetRequest` observes no target validation/write/read-back events and zero receipts. The #4132 foundation independently covers revocation, expiry, and token replay.
- [x] Receipt/read-back/checkpoint ordering is asserted: the app fake observes zero receipts during write/read-back, then the CLI composition test observes a successful checkpoint after the receipt.
- [x] Production composition contains no `HTTPActionRunner` or `DestURL` route: the app/CLI production-path search is clean and the composition test observes connector mutation/read-back.
- [x] `pm flow`, `pm help flow`, and `pm flow --help` all returned contextual help successfully; `docs/cli/flow.md`, golden transcripts, website source, generated docs, docs validation, and website typecheck were updated/checked.
- [x] GSD execute-phase, verify-work, and code-review were completed as the documented inline/manual fallback; review evidence is in `REVIEW.md` and coverage is in `SUMMARY.md`.
- [ ] PR has the exact integration base verified through the API after opening.

## External live evidence

Deferred to the captain-runbook procedure on #3994: this worktree has no
authorized credentialed provider connection, and no credential was requested
or exposed. The PR must retain this as a pre-merge human-gated proof; the
hermetic checks above establish the same observable call ordering but do not
claim a live provider result.
