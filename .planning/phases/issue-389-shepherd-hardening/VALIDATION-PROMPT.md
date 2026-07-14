# GPT-5.6 Sol Validation Prompt — Issue #389

Independently validate the exact current branch head and working tree after GPT-5.5 finishes issue
#389. Do not edit, commit, push, comment, or mutate GitHub. Read the implementation prompt, plan,
TDD ledger, verification checklist, issue-local GSD artifacts, and all changed Go/docs files.

Fail the review for any of the following:

- prompt advertises a GSD lifecycle tool excluded by the unit hard gate;
- issue/project/database identity can be reused by another issue;
- retries reset after process/database reopen or unsafe failures retry automatically;
- a signal can leave a child falsely running or nested activity has no <=15-second heartbeat;
- process exit alone can produce success without canonical advancement, expected artifact, exact
  head, model/high-thinking evidence, clean scope, current lease, and child closeout;
- `supervise` can dispatch a non-canonical/duplicate/concurrent unit, use generic auto, or cross the
  final parent-PR merge gate;
- logs include prompts, tool output, credentials, secrets, or chain-of-thought;
- tests assert implementation details while missing restart, stale-head, orphan, exhaustion, and
  final-gate behavior.

Run focused tests, `go test ./...`, `go test -race ./...`, `go vet ./...`,
`go build ./cmd/shepherd`, nested `make verify`, and root `go list ./...`. Return `PROCEED` only with
independent command evidence. Otherwise return `RETRY` with file/line findings and concrete fixes.
