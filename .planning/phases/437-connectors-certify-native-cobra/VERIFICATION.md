# Phase 437 Verification

Invocation `issue-437-pi-sol-high-20260719T095145Z`; Sol/high; start `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`.

`verificationPassed`: false

## TDD / behavior

- [x] Six artifacts predate tests and production edits.
- [x] Initial RED captured before production: native connectors constructors absent.
- [x] Two focused help RED checkpoints preceded final trailing/direct action help corrections.
- [x] Native connectors actions and nested certify action/current flags/operands.
- [x] Bare/text/JSON/topic/direct/positional/trailing help; invalid action usage.
- [x] Literal `--`, malformed/legal unknown, action/operand discovery and globals.
- [x] Certify exits 0/1/2/3 and one-envelope semantics.
- [x] Fresh-tree re-entrancy, bounded concurrency, cancellation, events, telemetry.
- [x] Credential values absent from output, reports, events, and telemetry tests.
- [x] Only namespace parser calls removed; dynamic connector `parseFlags` code/diff unchanged.

## Focused / broad gates

- [x] Focused native connectors/certify tests: final `3.989s`.
- [x] Focused repeated `-count=10`: final `34.833s`.
- [x] Focused `-race`: final `40.842s`.
- [x] Router/golden/certify/telemetry focus: final `111.919s`.
- [x] Full CLI: final `431.305s` through `make verify`.
- [x] Full certify package: final `337.280s` through `make verify`.
- [x] Certify concurrency/event race focus: `2.395s`.
- [x] Required local certify smoke: exit 0, `ConnectorCertification`, sample, pass; stderr empty.
- [x] Exact-start operational differential: 21/21 unchanged; contextual action help is the reviewed intentional change.
- [x] Connector validation: 547 checked, 0 findings.
- [x] `gofmt -w cmd internal` and clean diff check.
- [x] `go vet ./...`.
- [x] `go test ./...`.
- [x] `go build ./cmd/pm`.
- [x] Final `make verify` exit 0; docs/smoke/lint/connectors all green.

## Help/manual/website parity

- [x] `pm help connectors`, bare `pm connectors`, and `pm connectors --help` are byte-equal text manuals.
- [x] Direct and trailing connectors/certify action help is contextual and side-effect free.
- [x] Bare JSON is one `CommandManual` envelope for `connectors`.
- [x] Invalid action remains usage exit 2.
- [x] Certify examples, credential-reference safety, envelopes, and exit 0/1/2/3 are documented.
- [x] `docs/cli/connectors.md` regenerated from the canonical manual.
- [x] Website CLI reference mirrored and `website/lib/docs.generated.ts` regenerated.
- [x] Golden transcripts regenerated only for the reviewed connectors-manual change.
- [x] Docs generation/drift and website generation pass.
- [x] Website typecheck not applicable: existing `node_modules/.bin/tsc` is absent; no prohibited dependency install was attempted.
- [x] Completion metadata remains registered through the native tree with `NoFileComp`; Phase 15/19 broad churn remains deferred.

## Safety/scope/delivery

- [x] Correct isolated branch and exact start.
- [x] GSD adapter/manual fallback and required skills recorded.
- [x] Fixture/replay/local-only tests and smoke; no live credential check or external write.
- [x] No real credential value requested, printed, summarized, or stored.
- [x] No connector definitions, dependency files, or legacy dynamic parser changes.
- [x] No services, generic tools, destructive/admin/production action, or quality-gate reduction.
- [x] Planning, RED, GREEN, help-correction, and direct-help checkpoints committed/pushed.
- [x] No PR/review per user instruction.

## Accepted review correction checklist

Correction start: `0d1792cec3ea829ceb6228fc600b6dc7bbd90eee`; session `issue-437-review-correction-20260719T113319Z`.

- [x] Read and accept all five findings in `/tmp/pm-397-review-437.log`.
- [x] Reopen all phase artifacts before test or production edits.
- [x] RED: unsupported `record`/`replay`/production-write/rate-limit/budget/live-all-modes controls reject before runner invocation.
- [x] RED: single certify emits exactly one connector span and preserves connector-validation-before-options precedence.
- [x] RED: batch credential-file load precedes parallel parsing and preserves exact load/run error wrappers and bytes.
- [x] RED: only `--help`, `-h`, and intentional positional `help` render connectors manuals; false/assigned malformed/unknown short clusters do not.
- [x] RED: CLI and website docs separate pre-report usage/validation/runtime exits from completed report outcomes.
- [x] Focused differential and repeated/race tests: base/current 5/5 exact; race `29.046s`; ×10 `24.991s`.
- [x] Certify exits, redaction, and unsupported replay no-live/no-write runner test: package race `349.263s`; exit focus `21.618s`.
- [x] Local sample fixture smoke with temp root only: exit 0, `ConnectorCertification`, sample passed, stderr empty, planted value absent.
- [x] Runtime help, golden, generated CLI docs, website generation/parity: CLI docs/golden `24.275s`; website regeneration hash-stable.
- [x] Full CLI and certify packages: CLI `435.572s`; certify `338.846s`.
- [x] `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`: final verify exit 0, real `468.36s`.
- [x] `go run ./cmd/connectorgen validate` reports 547 checked, zero findings.
- [x] Final artifacts committed and pushed (`2987f21b`); no dependencies/services/credentials/PR/review.

## Correction result

Implementation head `a67d2ff9de84a2fabcd3b66097bf49518c1fa124`. Exact-base differential matched stdout, stderr, and exit codes in all five reviewed precedence/help cases. Unsupported replay and the other five controls return usage errors without any runtime call; no credential resolution, live stage, or write stage occurs. All verification used temporary roots, local sample behavior, synthetic planted redaction values, and existing replay fixtures only.

## Second accepted safety correction checklist

Second-correction start: `0d743e54e06c9e27e550eacce9be7899a9e23d19`; session `issue-437-second-safety-correction-20260719`.

- [x] Read and accept every P1/P2/P3 finding in `/tmp/pm-397-rereview-437.log`.
- [x] Reopen plan, TDD ledger, verification checklist, prompt snapshot, summary, and run state before tests or production edits.
- [x] Commit/push the planning checkpoint before RED tests (`aa39fd9d`).
- [x] RED effect-recorder tests expose that `--write=false` and `--skip=write` do not override credential-file `write: true`.
- [x] RED exposes configured credential-file sandbox/rate/budget/limit reaching batch/runner rather than failing closed.
- [x] RED exposes visible/accepted `--credential`, `--limit`, and `--modes`, unrestricted skip values, and mode-inapplicable controls running effects.
- [x] RED audit enumerates every declared certify flag by supported mode or explicit fail-closed expectation; GREEN mapping pending.
- [x] P3 stale certification architecture/PRD examples and claims removed; connector-help name claim made accurate.
- [x] Runtime help source, CLI docs, goldens, website docs, and generated data updated; final binary help checks remain in full verification.
- [x] Focused effect/no-op audit tests pass, repeated ×10, and under race (CLI `1.726s`, certify `2.535s`).
- [x] Full CLI (`440.910s`) and certify (`346.271s`) package suites preserve exits, redaction, dynamic dispatch, and valid base behavior.
- [ ] Local sample smoke uses only a temporary root and no credentials/services.
- [ ] Full CLI/certify/docs/website generators and drift checks pass.
- [ ] `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and `go run ./cmd/connectorgen validate` pass.
- [ ] Final artifacts committed and active issue branch pushed; no dependencies/services/credentials/PR/review.
