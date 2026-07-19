# Phase 436 Verification

Invocation: `issue-436-pi-sol-high-20260719T074902Z`; profile `Sol`; thinking `high`; exact start `eec03373dcc581c7f5c3331fe63287519b317f53`.

`verificationPassed`: false

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [ ] Focused RED captured before production edits.
- [ ] Hidden native extract command; legacy wrapper/parser call removed only for extract.
- [ ] Every current extract flag and legacy repeated/bare/assigned/space/operand behavior covered.
- [ ] Bare/text/JSON/long/short/topic/positional/trailing help is contextual and side-effect free.
- [ ] Literal `--`, malformed/legal unknowns, invalid actions/operands, and no later discovery covered.
- [ ] Globals, output envelopes, error taxonomy, and context preserved.
- [ ] Dependency-free simple/RLM routes use injected or hermetic fakes only.
- [ ] Input/output traversal containment and rooted final-link safety pass under temp roots.
- [ ] External sentinel files stay unchanged; valid in-root operations still work.

## Focused and broad gates

- [ ] Focused native extract tests.
- [ ] Focused extract tests repeated.
- [ ] Focused extract/RLM/safety race tests.
- [ ] Router and golden transcript tests.
- [ ] Exact-start parser/output differential for preserved and intentional-help cases.
- [ ] Full CLI/extract/RLM/safety tests.
- [ ] `gofmt -w cmd internal` and clean diff check.
- [ ] `go vet ./...`.
- [ ] `go test ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `make verify` dependency-free/default-only.
- [ ] No dependency, connector-definition, unrelated namespace, or broad generated delta.

## CLI help/manual/website parity

- [ ] `pm help extract` resolves the hidden topic.
- [ ] Bare `pm extract` prints contextual help and exits 0.
- [ ] `pm extract --help`, `-h`, positional help, trailing help, and JSON help are accurate.
- [ ] Invalid actions/operands remain usage errors and start no effect.
- [ ] Generated `docs/cli/extract.md` matches the canonical manual.
- [ ] Website agent-guide/CLI-reference/architecture references remain accurate.
- [ ] Generated docs/website checks produce only reviewed applicable deltas.
- [ ] Golden extract help changes are reviewed.
- [ ] Extract remains hidden from root discovery/completion listing.

## Safety/scope/delivery

- [x] Exact branch/start confirmed.
- [x] GSD doctor/list and plan-phase passed; unavailable programming-loop/manual fallback recorded.
- [x] Required CLI/testing/error/security/safety/path/docs/Cobra skills loaded.
- [ ] Temp-root tests use only synthetic local records and sentinel files.
- [ ] No broad input/output paths or final-link external effects.
- [ ] No external files/services, model, Temporal, Podman, worker, listener, database service, credentials, or connector checks.
- [ ] No dependencies, generic shell/HTTP/SQL write surface, destructive/admin action, production action, or quality-gate reduction.
- [ ] Planning, RED, GREEN/refactor, and terminal checkpoints committed/pushed.
- [ ] No PR/review created per user instruction.
