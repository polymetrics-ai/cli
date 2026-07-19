# Phase 434 Verification

Invocation: `issue-434-pi-sol-high-20260719T053630Z`; profile `Sol`; thinking `high`; exact start `2ac457a163cbd7bc9a3708da88b03d375ec5e952`.

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [x] Focused RED captured before production edits (four intentionally undefined native RLM symbols).
- [ ] Native RLM run/help tree; legacy wrapper removed.
- [ ] All current local flags typed with repeated/bare/assigned/space behavior preserved.
- [ ] Deterministic/fixture/model/agent routing verified with injected fakes only.
- [ ] Bare/text/JSON/long/short/positional/trailing help parity.
- [ ] Literal `--`, malformed/legal unknowns, invalid actions, no action/operand discovery bypass, and globals.
- [ ] Exact error taxonomy, stdout/stderr, one-envelope JSON, request non-leakage, and no generic runner.
- [ ] Only RLM parser/dispatcher removed; dynamic connector parser remains.

## Gates

- [ ] Focused native RLM/router tests.
- [ ] Focused tests repeated.
- [ ] Focused race tests.
- [ ] Existing RLM analyzers/spec/fixture/model/agent tests.
- [ ] Worker fake/workflow tests that require no service.
- [ ] Router and golden transcript focus; fixture unchanged or explicitly reviewed.
- [ ] Exact-start parser/output differential.
- [ ] Full `go test ./internal/cli/...`.
- [ ] `gofmt -w cmd internal`.
- [ ] `go vet ./...`.
- [ ] `go test -timeout 20m ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify` dependency-free.
- [ ] `git diff --check`; no dependency/connector-definition/unrelated delta.

## CLI help/manual/website parity

- [ ] `pm help rlm`.
- [ ] Bare `pm rlm` exits 0 with contextual manual.
- [ ] `pm rlm --help`, `-h`, positional help, and JSON manual routes.
- [ ] Invalid action remains a usage error.
- [ ] `docs/cli/rlm.md` generated parity; update or mark unchanged with evidence.
- [ ] Website CLI-reference/architecture RLM docs checked/generated; update or mark unchanged.
- [ ] Generated/golden help artifacts checked.
- [ ] Completion discovery seam present; Phase 15 values remain deferred.
- [ ] Phase 16 RLM viewer and Phase 19 focused help/man work remain deferred.

## Safety/scope/delivery

- [x] Exact branch/start and parent draft PR confirmed.
- [x] GSD doctor/list passed; unavailable programming-loop/manual fallback recorded.
- [x] Required CLI/testing/error/security/safety/context/concurrency/docs/Cobra skills loaded.
- [ ] Temp specs/warehouses and injected analyzers/hermetic fake runner only.
- [ ] No model, Temporal, Podman, worker service, credential, secret/request leakage, or generic runner.
- [ ] No optional services, dependencies, unrelated namespaces, or broad generated churn.
- [ ] Planning, RED, GREEN/refactor, and final checkpoints committed/pushed.
- [x] No PR/review planned per user instruction.

Result: pending.
