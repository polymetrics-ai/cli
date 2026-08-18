# #4063 Verification Checklist

**Status:** Local verification and no-mistakes green; inherited base-range whitespace gate needs disposition

## Required local gates

- [x] RED: go run ./cmd/connectorgen certification-matrix --check exited 1
      before the generator ran at exact starting HEAD.
- [x] go test -timeout 20m ./cmd/connectorgen -run
      '^(TestCertificationDiscoverFunctionKindsFromRuntimeSource|TestCertificationGeneratedArtifactDriftFails)$'
      -count=1
- [x] go test -timeout 20m ./cmd/connectorgen -count=1
- [x] go test -timeout 20m ./internal/flow ./internal/app ./internal/warehouse
      ./internal/cli -count=1
- [x] go run ./cmd/connectorgen certification-matrix --check
- [x] go vet ./cmd/connectorgen ./internal/flow ./internal/app
      ./internal/warehouse ./internal/cli
- [x] make lint
- [x] scripts/verify-gsd-workflow origin/feat/3988-github-certification
- [ ] git diff --check 5c92888c996319b41eec6e86ca99fcda4cb365f9
- [x] Semantic JSON comparison confirms exactly one flow-matrix scalar.
- [x] Inline manual-GSD verify-work records no gap; #4066's separately
      authorized final safety correction has its own verification record.
- [x] Inline manual-GSD code review dispositions every finding in
      `.planning/phases/issue-4066-flow-owner-alias-guard-r1/REVIEW.md`.
- [x] Existing Shepherd/no-mistakes run `01KZR72KN8ZKF8V8M5Q19FPJJ5` passed
      with push, PR, and CI skipped; it never discovered or retargeted a
      head-to-main PR.

## External delivery gates

- [ ] Existing #4060 is draft.
- [ ] Existing #4060 head is the pushed correction SHA.
- [ ] Existing #4060 base remains feat/3988-github-certification.
- [ ] Current replacement Verify is green.
- [ ] Current Snyk status is observed without rerunning or merging anything.

## Explicitly not applicable

No CLI command surface, help text, manual, website documentation, connector
definition, provider interaction, credential, or runtime behavior changes.

## Inherited base-range formatting finding

The required base-range command fails at exact starting head as well as the
current correction head, solely on seven Markdown hard-break whitespace lines
in the older #3897 planning records. They were introduced by
06fdfe2f94c8093ed9f504b210f1bd3f1ce01163, before this correction's exact
start. `git diff --check 002ddf674a447bf0872486aa979efdaa078f602c HEAD`
passes. The correction does not edit those pre-existing records; a scope
decision is required before changing them.
