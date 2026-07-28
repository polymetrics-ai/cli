# Verification checklist — CLI PM Broker profile/context foundation

## Scope and safety

- [ ] Branch is `fm/cli-pm-broker-profile-context-r1` from `origin/integration/pm-broker-production-program`.
- [ ] Diff does not touch the legacy vault, move credentials, add dependencies, create production resources, or add generic HTTP/SQL/shell/raw JSON escape hatches.
- [ ] PM Broker live operations remain unsupported; fake-client integration is a TODO seam only if the package is absent.
- [ ] Public auth registry/plugin SDK stability is not claimed.
- [ ] `git diff --check` passes.

## Focused behavior

- [ ] `go test ./internal/pmbroker` passes.
- [ ] `go test ./internal/config -run TestLoadBroker -count=1` passes.
- [ ] `go test ./internal/cli -run 'TestPMBroker(Context|Organizations|Workspaces|Environments)' -count=1` passes.
- [ ] `pm help context`, `pm context`, and `pm context --help` render context help.
- [ ] `pm organizations`, `pm workspaces`, and `pm environments` render namespace help; `list`/`show` print cached safe metadata only.
- [ ] Invalid context/runtime-mode actions return usage or validation errors.
- [ ] JSON output uses versioned envelopes and contains no raw secrets.

## Docs/help/website parity

- [ ] Embedded help updated for `context`, `organizations`, `workspaces`, `environments`, and broker config keys.
- [ ] `docs/cli/context.md`, `docs/cli/organizations.md`, `docs/cli/workspaces.md`, `docs/cli/environments.md`, and `docs/cli/config.md` are updated/regenerated.
- [ ] Website CLI reference/generated docs are updated or explicitly link to canonical docs.
- [ ] Docs mention runtime modes `remote`, `local`, policy-bound `hybrid`, production remote default, and no local fallback for production writes/scheduled production jobs.

## Broader local gates

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify` if practical in this lane before no-mistakes.

`verificationPassed` remains false until the declared local and PR checks pass.
