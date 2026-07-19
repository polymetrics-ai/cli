# Phase 431 Verification

Invocation session: `issue-431-pi-openai-codex-gpt-5.6-sol-high-20260719T010451Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `0b03361e3ec5082d54c416a31715851f71e845fa`.

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [x] Focused RED captured before production edits (`undefined: newReverseCobraCommand`).
- [x] Native reverse list/plan/preview/approval-bearing run/status/help tree; legacy wrapper removed.
- [x] All current flags typed with repeated/bare/assigned compatibility.
- [x] Bare/text/JSON/long/short/positional help parity.
- [x] Trailing help, literal `--`, strict first-operand ownership, and no carrier override.
- [x] Unknown flags/actions, global assigned booleans, and no action-discovery bypass.
- [x] Exact exit taxonomy and one-envelope stdout/stderr behavior in focused tests.
- [x] Approval value absent from JSON, diagnostics, errors, and local logs in focused tests.
- [x] Typed confirmation and strict plan → preview → approval → execute ordering with no pre-gate writer invocation.
- [x] Only reverse `parseFlags` call sites removed; dynamic connector parser remains.

## Gates

- [x] Focused native reverse/router tests (`28.527s`; broader reverse/router safety suite `62.562s`).
- [x] Focused repeated tests (`-count=5`: `143.255s`).
- [x] Focused race tests (`-race`: CLI `322.315s`; reverse app `29.405s`).
- [x] Existing reverse app/CLI safety tests (`62.562s`; app `4.615s`; smoke-order `0.390s`).
- [x] Router and golden transcript tests (`7.754s`); fixture unchanged.
- [x] Exact-start binary differential: 21/21 exit/stdout/stderr transcripts match.
- [x] Full `go test ./internal/cli/...` (`396.560s`).
- [x] Full reverse/app packages (app `29.601s`).
- [x] `gofmt -w cmd internal`.
- [x] `go vet ./...` (`3.104s`).
- [x] `go test -timeout 20m ./...` (`6m44.294s`; CLI `399.524s`, certify `340.870s`).
- [x] `go build ./cmd/pm` (`1.865s`).
- [x] Established ordered `make verify` gate (`6m56.086s`, CLI `388.922s`, lint 0, connectors 547/0), including its existing local smoke only.
- [x] `git diff --check`; no dependency/unrelated/connector-definition delta.

## CLI help/manual/website parity

- [x] `pm help reverse`.
- [x] Bare `pm reverse` exits 0 with contextual manual.
- [x] `pm reverse --help`, `-h`, positional `help`, and JSON manual routes.
- [x] Invalid action remains usage error, including trailing help.
- [x] `docs/cli/reverse.md` generated parity checked; no update applicable because canonical manual bytes are unchanged.
- [x] `website/content/docs/reverse-etl.mdx` generation/parity checked; no update applicable because public behavior/content is unchanged.
- [x] Generated/golden help artifacts checked and unchanged.
- [x] Completion discovery seam present; Phase 15 values remain deferred.
- [x] Focused per-subcommand help/man churn remains deferred to Phase 19.

## Safety/scope/delivery

- [x] Exact branch/start and parent draft PR confirmed.
- [x] GSD doctor/list passed; unavailable programming-loop/manual fallback recorded.
- [x] Required Go CLI/testing/error/security/safety/context/docs/Cobra skills loaded.
- [x] Local fakes and temporary state only; no external write or credentialed check.
- [x] No token values in artifacts, logs, JSON, diagnostics, or final handoff.
- [x] No optional services, dependencies, unrelated namespaces, or broad generated churn.
- [x] Planning, RED, and GREEN checkpoints committed/pushed; final evidence prepared for commit/push.
- [x] No PR/review created per user instruction.

Original result: pass at implementation head `f5aeafb7bb7a6702077382e98acb790d3865073f`.

## Parser compatibility correction checklist

Correction start: `c8f5b9e97a2f71f25cdb362af0055c1c31dc8420`.

- [x] Review log read; exact clean start/upstream confirmed.
- [x] GSD doctor/list run; unavailable programming-loop/manual fallback recorded.
- [x] Correction plan, TDD ledger, verification checklist, prompt snapshot, summary, and run state reopened before production edits.
- [x] RED differential tests cover every reverse action and malformed legacy-accepted `--=x`, `---x`, plus representative variants (50 combinations).
- [x] RED proves pflag rejects every malformed form with usage exit 2 before baseline action behavior (list 0, plan validation 3, preview/run/status missing-object 1).
- [x] Committed differential includes unchanged-state, empty-outbox, and no-approval-output assertions; RED stopped at the expected outcome mismatch, and all effect guards executed and passed after GREEN.
- [x] Production normalization changes only pflag-invalid reverse-tail names beginning with `=` or an extra `-` after `--`.
- [x] Unit and workflow tests preserve ordinary known flags, legal unknown flags, first operands, approval/confirmation ordering, and action outcomes.
- [x] Focused correction (`26.289s`) and complete reverse native (`63.513s`) suites pass.
- [x] Focused race gate passes (`295.302s`).
- [x] 324/324 exact-start parser differential matches exit/stdout/stderr; state hashes and outbox listings remain unchanged.
- [x] Full CLI suite passes (`417.589s`).
- [x] No approval field/value or human approval line appears in differential output, focused output, artifacts, logs, or diagnostics.
- [x] `gofmt`, `go vet ./...`, `go build ./cmd/pm`, `git diff --check`, scope, and dependency checks pass.
- [x] Help/manual/docs/website changes not applicable: no public command/flag/help/output change.
- [x] No external write/service/credential/dependency/PR/review is used.
- [x] RED commit `c98e4dad` and GREEN commit `bbe9bb9c` pushed; this terminal artifact checkpoint completes requested delivery on push.

Correction result: pass at implementation head `bbe9bb9c`; `verificationPassed=true` because every user-declared bounded correction gate exited 0. No full repository or `make verify` rerun was declared for this review correction; the original implementation's full repository and `make verify` gates remain recorded above.
