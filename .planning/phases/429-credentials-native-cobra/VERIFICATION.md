# Phase 429 Verification

Invocation session: `issue-429-pi-openai-codex-gpt-5.6-sol-high-20260718T143346Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `0f1ec1e89cdae761e9da06ab9906fcc641b38e0a`.

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [ ] Exact focused RED captured before production edits.
- [ ] Native credentials/add/list/inspect/test/remove/help tree; legacy wrapper removed.
- [ ] Typed repeated/bare current flags; spaced/assigned/unknown/extra positional forms compatible.
- [ ] Bare/text/JSON/long/short/positional help parity.
- [ ] Invalid action/global assigned boolean behavior preserved.
- [ ] Strict identifiers and existing write-path containment covered.
- [ ] Env/stdin-only secret intake and source-selection rules covered with controlled opaque fixtures.
- [ ] Stdout/stderr/state metadata redaction covered without logging fixture values.
- [ ] Leading invalid tokens cannot discover or execute later actions.
- [ ] Credentials action-tail help and literal `--` compatibility preserved.
- [ ] Only the credentials `parseFlags` call removed; dynamic/shared parser remains.

## Gates

- [ ] Focused credentials/router tests.
- [ ] Focused repeated tests.
- [ ] Focused `-race` credentials tests.
- [ ] Security/redaction/path/action-boundary tests.
- [ ] Golden transcript test and fixture diff review.
- [ ] Full `internal/cli/...`.
- [ ] Full repository tests in temporary roots where applicable.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify`.
- [ ] `git diff --check` and dependency/scope guards.

## CLI help/manual/website parity

- [ ] `pm help credentials`.
- [ ] Bare `pm credentials` exits 0 with contextual manual.
- [ ] `pm credentials --help`, `-h`, positional `help`, and JSON manual routes.
- [ ] Invalid actions remain usage errors.
- [ ] `docs/cli/credentials.md`: update or prove no change applicable with temp generation diff.
- [ ] `website/**`: update or prove no change applicable with generator diff.
- [ ] Generated/golden help artifacts: update or prove no change applicable.
- [ ] Completion/discovery seam tested; Phase 15 values remain deferred.
- [ ] Focused subcommand help/man churn remains deferred to Phase 19.

## Safety/scope/delivery

- [x] GSD doctor/list passed; missing programming-loop command and manual fallback recorded.
- [x] Required GSD, Go CLI/testing/error/security/safety/docs/Cobra skills loaded.
- [x] Parent draft PR #438 confirmed; this serialized worker remains local critical path.
- [ ] No real secret requested, read, printed, summarized, stored, or logged.
- [ ] No credentialed connector checks or external service calls.
- [ ] No interactive secret entry.
- [ ] No dependency or `go.mod`/`go.sum` delta.
- [ ] No connector-def, unrelated namespace, or broad generated churn.
- [ ] Planning, RED, GREEN, and final evidence checkpoints committed/pushed.
- [ ] No PR or external review requested.

Result: pending implementation and full verification.
