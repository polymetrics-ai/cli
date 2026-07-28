# Verification checklist — CLI PM Broker profile/context foundation

## Scope and safety

- [x] Branch is `fm/cli-pm-broker-profile-context-r1` from `origin/integration/pm-broker-production-program`.
- [x] Diff does not touch the legacy vault, move credentials, add dependencies, create production resources, or add generic HTTP/SQL/shell/raw JSON escape hatches.
- [x] PM Broker live operations remain unsupported; fake-client integration is a TODO seam only if the package is absent.
- [x] Public auth registry/plugin SDK stability is not claimed.
- [x] `git diff --check` passes.
- [x] `fm-ensure-agents-md.sh .` reported a pre-existing tracked-real-file conflict between `AGENTS.md` and `CLAUDE.md`; per firstmate instruction, no agent-memory file was reconciled or edited in this PR.

## Focused behavior

- [x] `go test ./internal/pmbroker` passes.
- [x] `go test ./internal/config -run TestLoadBroker -count=1` passes.
- [x] `go test ./internal/cli -run 'TestPMBroker(Context|Organizations|Workspaces|Environments)' -count=1` passes.
- [x] `pm help context`, `pm context`, and `pm context --help` render context help.
- [x] `pm organizations`, `pm workspaces`, and `pm environments` render namespace help; `list`/`show` print cached safe metadata only.
- [x] Invalid context/runtime-mode actions return usage or validation errors.
- [x] JSON output uses versioned envelopes and contains no raw secrets.

## Docs/help/website parity

- [x] Embedded help updated for `context`, `organizations`, `workspaces`, `environments`, and broker config keys.
- [x] `docs/cli/context.md`, `docs/cli/organizations.md`, `docs/cli/workspaces.md`, `docs/cli/environments.md`, and `docs/cli/config.md` are updated/regenerated.
- [x] Website CLI reference/generated docs are updated and explicitly document the PM Broker CLI foundation scope.
- [x] Docs mention runtime modes `remote`, `local`, policy-bound `hybrid`, production remote default, and no local fallback for production writes/scheduled production jobs.

## Broader local gates

- [x] `gofmt -w cmd internal`
- [x] `go vet ./...`
- [x] `go test ./...`
- [x] `go build ./cmd/pm`
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- [x] `make verify`

`localVerificationPassed` is true. `verificationPassed` remains false until no-mistakes/PR checks pass.
