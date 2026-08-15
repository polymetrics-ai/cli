# Issues #3990 and #4091 — verification checklist

**Status:** planned; live evidence pending.

- [ ] Fresh production `pm` binary identity recorded safely.
- [ ] Immutable GitHub live-lab boundary validated before credential/provider access.
- [ ] #3990 require-shared real GitHub proof records observable budget/provider outcomes.
- [ ] #4091 approved set-replace/keyed execution changes and independently reads back exact target state.
- [ ] #4091 replay, scope/consent/auth refusals assert unchanged provider and checkpoint state.
- [ ] Every captain edge case is live or individually documented as simulated/untestable.
- [ ] Final cleanup/restoration leaves zero unexpected GitHub residue.
- [ ] Evidence scan contains no secret, approval token, raw body, or rendered rate scope.
- [ ] Focused tests and non-full-suite repository gates pass.
- [ ] One-pass derived regeneration and all drift checks pass.
- [ ] Inline/manual `verify-work` and `code-review` complete.
- [ ] PR opened and API-reported base equals `integration/4015-mvp-flat-r1`.

## Production call chains to verify

- #3990: `cmd/pm/main.go` → `cli.Run` → shared rate registry configuration → certification/connector
  dispatch → GitHub policy admission/observation → shared coordinator.
- #4091: `cmd/pm/main.go` → `cli.Run` → `app.Open`/definition composition → `etl transport` /
  `etl run` dispatch → issue-label authorization gate → GitHub writer/read-back → acknowledged
  checkpoint.

## Known exclusions

- #4125 and #4158 are not verification failures for this task and will not be edited.
