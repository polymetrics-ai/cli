# Phase 431 Verification

Invocation session: `issue-431-pi-openai-codex-gpt-5.6-sol-high-20260719T010451Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `0b03361e3ec5082d54c416a31715851f71e845fa`.

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [x] Focused RED captured before production edits (`undefined: newReverseCobraCommand`).
- [x] Native reverse list/plan/preview/approval-bearing run/status/help tree; legacy wrapper removed.
- [x] All current flags typed with repeated/bare/assigned compatibility.
- [ ] Bare/text/JSON/long/short/positional help parity.
- [x] Trailing help, literal `--`, strict first-operand ownership, and no carrier override.
- [x] Unknown flags/actions, global assigned booleans, and no action-discovery bypass.
- [x] Exact exit taxonomy and one-envelope stdout/stderr behavior in focused tests.
- [x] Approval value absent from JSON, diagnostics, errors, and local logs in focused tests.
- [x] Typed confirmation and strict plan → preview → approval → execute ordering with no pre-gate writer invocation.
- [x] Only reverse `parseFlags` call sites removed; dynamic connector parser remains.

## Gates

- [x] Focused native reverse/router tests (`28.527s`; broader reverse/router safety suite `62.562s`).
- [ ] Focused repeated tests (`-count=5`).
- [ ] Focused race tests (`-race`).
- [ ] Existing reverse app/CLI safety tests.
- [ ] Router and golden transcript tests; fixture unchanged or explicitly reviewed.
- [ ] Full `go test ./internal/cli/...`.
- [ ] Full reverse/app packages.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test -timeout 20m ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] Established ordered `make verify` gate, including its existing local smoke only.
- [ ] `git diff --check`; no dependency/unrelated/connector-definition delta.

## CLI help/manual/website parity

- [ ] `pm help reverse`.
- [ ] Bare `pm reverse` exits 0 with contextual manual.
- [ ] `pm reverse --help`, `-h`, positional `help`, and JSON manual routes.
- [ ] Invalid action remains usage error, including trailing help.
- [ ] `docs/cli/reverse.md` generated parity checked; update or explicit no-change rationale.
- [ ] `website/content/docs/reverse-etl.mdx` generation/parity checked; update or explicit no-change rationale.
- [ ] Generated/golden help artifacts checked.
- [ ] Completion discovery seam present; Phase 15 values remain deferred.
- [ ] Focused per-subcommand help/man churn remains deferred to Phase 19.

## Safety/scope/delivery

- [x] Exact branch/start and parent draft PR confirmed.
- [x] GSD doctor/list passed; unavailable programming-loop/manual fallback recorded.
- [x] Required Go CLI/testing/error/security/safety/context/docs/Cobra skills loaded.
- [ ] Local fakes and temporary state only; no external write or credentialed check.
- [ ] No token values in artifacts, logs, JSON, diagnostics, or final handoff.
- [ ] No optional services, dependencies, unrelated namespaces, or broad generated churn.
- [ ] Planning, RED, GREEN, and final evidence checkpoints committed/pushed.
- [x] No PR/review planned per user instruction.

Result: pending; `verificationPassed=false` until every declared gate, including `make verify`, exits 0.
