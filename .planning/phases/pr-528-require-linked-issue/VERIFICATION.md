# Verification checklist - PR 528 require-linked-issue repair

- [x] Existing submitted head preserved: `777e630e620e6931f2493aade1c61a843608d94c`.
- [x] Local branch restored from detached checkout to `fm/cli-release-and-connector-issues-r1`.
- [x] Red evidence reproduced with `cmd/prissueguard` against live PR #528 body.
- [x] PR #528 body contains `Refs #67`.
- [x] `cmd/prissueguard` passes against the edited live PR body.
- [x] `go test ./internal/coordination/issueguard ./cmd/prissueguard` passes.
- [x] Planning trace prepared for commit and push to `fm/cli-release-and-connector-issues-r1`.

## Out of scope

- Full `make verify`; this repair changes no Go production code and targets the PR-body guard.
- No release, tag, or website deploy.
