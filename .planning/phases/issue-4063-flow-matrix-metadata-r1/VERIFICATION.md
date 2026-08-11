# #4063 Verification Checklist

**Status:** Planned

## Required local gates

- [x] RED: go run ./cmd/connectorgen certification-matrix --check exited 1
      before the generator ran at exact starting HEAD.
- [ ] go test -timeout 20m ./cmd/connectorgen -run
      '^(TestCertificationDiscoverFunctionKindsFromRuntimeSource|TestCertificationGeneratedArtifactDriftFails)$'
      -count=1
- [ ] go test -timeout 20m ./cmd/connectorgen -count=1
- [ ] go test -timeout 20m ./internal/flow ./internal/app ./internal/warehouse
      ./internal/cli -count=1
- [ ] go run ./cmd/connectorgen certification-matrix --check
- [ ] go vet ./cmd/connectorgen ./internal/flow ./internal/app
      ./internal/warehouse ./internal/cli
- [ ] make lint
- [ ] scripts/verify-gsd-workflow origin/feat/3988-github-certification
- [ ] git diff --check 5c92888c996319b41eec6e86ca99fcda4cb365f9
- [ ] Semantic JSON comparison confirms exactly one flow-matrix scalar.
- [ ] Inline manual-GSD verify-work records no gap.
- [ ] Inline manual-GSD code review dispositions every finding.
- [ ] Existing Shepherd/no-mistakes path runs with --skip=push,pr,ci and
      never discovers or retargets a head-to-main PR.

## External delivery gates

- [ ] Existing #4060 is draft.
- [ ] Existing #4060 head is the pushed correction SHA.
- [ ] Existing #4060 base remains feat/3988-github-certification.
- [ ] Current replacement Verify is green.
- [ ] Current Snyk status is observed without rerunning or merging anything.

## Explicitly not applicable

No CLI command surface, help text, manual, website documentation, connector
definition, provider interaction, credential, or runtime behavior changes.
