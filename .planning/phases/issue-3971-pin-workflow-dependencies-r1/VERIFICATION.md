# Verification checklist: pin build dependencies

## Pre-edit research

- [x] Alert #135 remediation guidance read from GitHub REST API.
- [x] Alert location confirmed: `.github/workflows/verify.yml:86`.
- [x] All workflow `uses:` references, Dockerfile `FROM` lines, and literal workflow images inventoried locally.
- [x] Current action commits and image manifest-list digests captured in `TDD-LEDGER.md` before edits.

## Targeted gates

- [ ] Red: `./scripts/tests/pinned-build-dependencies.sh` fails against the baseline.
- [ ] Green: `./scripts/tests/pinned-build-dependencies.sh` passes after pinning.
- [ ] `make release-workflow-check` passes.
- [ ] Workflow YAML parser check passes for every `.github/workflows/*.yml` file.
- [ ] `git diff --check` passes.
- [ ] `go run ./cmd/agentcontractgen check` passes.
- [ ] `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check` pass individually where applicable.

## Remote follow-up (last, after local work)

- [ ] Search GitHub for an existing equivalent issue before creating a child of #3971.
- [ ] Create one focused child issue only if no duplicate exists.
- [ ] Re-read alert #135 and inventory the other four open alerts after GitHub REST quota permits.
- [ ] State whether alert #135 is closed only after a post-push analysis result proves it.

## Not applicable

- CLI help, manual, and website documentation parity: no CLI surface or user-facing documentation changes.
- Full `go test ./...`: not required for a workflow/image-only change; the task-specific and existing `make verify` component gates are the local evidence.
