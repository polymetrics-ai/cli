# Issues #3990 and #4091 — verification checklist

**Status:** live proof and all local gates passed; commit and PR delivery pending.

- [x] Fresh production `pm` binary identity recorded safely.
- [x] Immutable GitHub live-lab boundary validated before credential/provider access.
- [x] #3990 require-shared real GitHub proof records observable budget/provider outcomes.
- [x] #4091 approved set-replace/keyed execution changes and independently reads back exact target state.
- [x] #4091 replay, scope/consent/auth refusals assert unchanged provider and checkpoint state.
- [x] Every captain edge case is live or individually documented as simulated/untestable.
- [x] Final cleanup/restoration leaves zero unexpected GitHub residue.
- [x] Evidence scan contains no secret, approval token, or rendered rate scope.
- [x] Focused tests and non-full-suite repository gates pass.
- [x] One-pass derived regeneration and all drift checks pass.
- [x] Inline/manual `verify-work` and `code-review` complete.
- [ ] PR opened and API-reported base equals `integration/4015-mvp-flat-r1`.

## Production call chains to verify

- #3990: `cmd/pm/main.go` → `cli.Run` → shared rate registry configuration → certification/connector
  dispatch → GitHub policy admission/observation → shared coordinator.
- #4091: `cmd/pm/main.go` → `cli.Run` → `app.Open`/definition composition → `etl transport` /
  `etl run` dispatch → issue-label authorization gate → GitHub writer/read-back → acknowledged
  checkpoint.

## Known exclusions

- #4125 and #4158 are not verification failures for this task and will not be edited.

## Local verification results

- `go test -timeout 20m ./internal/app/... -count=1` — pass (`784.631s`).
- `go test -timeout 20m ./internal/cli -count=1` — pass (`923.065s`).
- `go test -timeout 20m ./internal/connectors/certify -count=1` — pass.
- Focused transport, golden transcript, edge, coordination, engine, and durable authorization suites — pass.
- `go vet ./internal/connectors/certify ./internal/cli ./internal/app` and `make lint` — pass.
- Current production binary: docs validation, contextual help entry points, and `make smoke-no-build` — pass.
- One-pass connector/manual/skill/transcript/website regeneration followed by agent-contract,
  connector validation/surface/matrix/boundary, GitHub parity artifact, connector canon, pinned
  dependency, Homebrew notification, and release-target drift checks — pass.
- `git diff --check`, dependency tidy check, and tracked/untracked credential/rendered-scope scans — pass.
- Generated `verify-work` and `code-review` prompts executed through the documented inline/manual
  fallback; `UAT.md` is complete and `REVIEW.md` reports zero findings.
