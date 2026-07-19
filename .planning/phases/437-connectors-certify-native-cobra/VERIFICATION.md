# Phase 437 Verification

Invocation `issue-437-pi-sol-high-20260719T095145Z`; Sol/high; start `6c038bb4ab4a5497fca28a0cab42d0a7fa4eb22b`.

`verificationPassed`: true

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
