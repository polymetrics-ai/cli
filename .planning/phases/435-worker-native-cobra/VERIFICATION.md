# Phase 435 Verification

Invocation: `issue-435-pi-sol-high-20260719T064417Z`; profile `Sol`; thinking `high`; exact start `14c02d295065c3bf33c65eaac5f8d36642798f81`.

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [x] Focused RED captured before production edits (five intentionally undefined native worker symbols).
- [ ] Hidden native worker status/serve/help tree; legacy wrapper removed.
- [ ] No local worker flags; current globals/operands preserved.
- [ ] Bare/text/JSON/long/short/positional/trailing help is contextual and side-effect free.
- [ ] Literal `--`, malformed/legal unknowns, invalid actions, and no action discovery bypass.
- [ ] Fake status/serve routing and no service start on help/invalid/config errors.
- [ ] Context propagation and cancellation.
- [ ] Config precedence and unrelated-value nondisclosure.
- [ ] Typed RLM-workflow-only worker; no generic runner.

## Gates

- [ ] Focused native worker/router tests.
- [ ] Focused tests repeated.
- [ ] Focused race tests.
- [ ] Existing worker fake/workflow tests requiring no service.
- [ ] Router and golden transcript focus; worker fixture deltas reviewed.
- [ ] Exact-start parser/output differential.
- [ ] Full `go test -timeout 15m ./internal/cli/... -count=1`.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test -timeout 20m ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify` dependency-free/default-only.
- [ ] `git diff --check`; no dependency/connector-definition/unrelated delta.

## CLI help/manual/website parity

- [ ] `pm help worker` direct hidden topic behavior.
- [ ] Bare `pm worker` contextual help and exit 0.
- [ ] `pm worker --help`, `-h`, positional help, trailing action help, and JSON manual routes.
- [ ] Invalid actions remain usage errors and start no service.
- [ ] `docs/cli/worker.md` generated parity or explicit not-applicable reason.
- [ ] Website CLI-reference/architecture worker docs checked/generated.
- [ ] Generated/golden help artifacts checked; any worker fixture delta reviewed.
- [ ] Worker remains hidden from root discovery/completion listing.
- [ ] Phase 16 dashboard and Phase 19 broad help/man work remain deferred.

## Safety/scope/delivery

- [x] Exact branch/start and parent draft PR confirmed.
- [x] GSD doctor/list passed; unavailable programming-loop/manual fallback recorded.
- [x] Required CLI/testing/error/security/safety/context/concurrency/docs/Cobra skills loaded.
- [ ] Invocation-local fake status/serve only; no Temporal dial/worker or listener.
- [ ] No Podman, database, runtime service, credential, secret/config-canary disclosure, or generic runner.
- [ ] No dependencies, unrelated namespaces, or broad generated churn.
- [ ] Planning, RED, GREEN/refactor, and final checkpoints committed/pushed.
- [x] No PR/review will be created per user instruction.

Result: pending; `verificationPassed=false` until full `make verify` exits 0.
