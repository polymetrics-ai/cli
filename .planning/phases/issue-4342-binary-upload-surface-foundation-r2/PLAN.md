# Plan — issue #4342 binary upload CLI and certification foundation

## GSD command path

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, `code-review`: resolved to the pinned official command sources.
- `scripts/gsd prompt discuss-phase issue-4342-binary-upload-surface-foundation-r2`: generated and executed inline as recorded in `CONTEXT.md`.
- `scripts/gsd prompt plan-phase issue-4342-binary-upload-surface-foundation-r2 --tdd`: generated and executed inline as recorded in `CONTEXT.md`.

## Implementation slices

1. **Red — closed command admission.** Add commandrunner/unit tests for `binary_upload` to prove it accepts only an implemented declared write whose body is binary, base64, or multipart with a file part; prove no generic dispatch path exists and bad bindings fail before a provider call.
2. **Green — runtime and public surface.** Add the intent to schema/model/runtime validation, route it through the existing write-command planner, bind GitHub's declared asset action, and update help/manually rendered command flags. Keep source field, cap, digest, media policy, and confirmation declaration-owned.
3. **Red — truthful certification.** Add report and stage tests distinguishing a transfer from refusal. Assert response evidence, byte count, digest, read-back/cleanup state, and failure on omitted cleanup or altered bytes.
4. **Green — candidate, stage, and sweep.** Add a dedicated upload candidate contract and stage; report download/upload independently; add the `binary_upload` sweep class and separate operation-evidence classification. Preserve the `file_upload` no-executor guard.
5. **Refactor and generated projections.** Regenerate/verify connector manual, skill, and website data using project generators. Update conventions only where the newly public intent changes author guidance.
6. **Verify/review.** Run GSD verify-work generated prompt inline, resolve any gaps with the required gap workflow, then run GSD code-review generated prompt inline and record dispositions.

## CLI help/manual/website parity

- [ ] `pm github releases assets upload --help` lists the declared upload source flag and approval behavior.
- [ ] The same command's help resolves without a credential or project when invoked with `--help`.
- [ ] Existing bare connector namespace behavior remains unchanged and tests remain green.
- [ ] `docs/connectors/github/{MANUAL,SKILL}.md` and website generated CLI surface use `binary_upload` and its closed safety rule.
- [ ] No `docs/cli` generic upload command is added because no generic command exists.
- [ ] Generated documentation and website drift checks are run.

## Verification plan

Use `GOFLAGS='-p=3'` and one heavy suite at a time. The targeted package set is `internal/connectors/engine`, `internal/connectors/commandrunner`, `internal/cli`, `internal/connectors/certify`, and `cmd/connectorgen`, plus docs/generator checks. The full test suite and `make verify` are not run as one process in this memory-bound worktree; every omitted local gate will be recorded with its exact reason before PR creation.
