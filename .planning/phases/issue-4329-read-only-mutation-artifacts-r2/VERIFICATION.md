# Verification — issue #4329, r2

## Status

Local verification is complete for the current-main integration. A fresh
independent audit of the exact integration SHA and required remote CI remain
pending after push.

## Completed evidence

- **Red:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestSourceProjectionWriteDisabledLockedSourcesRetainMutationArtifacts|TestSourceProjectionWriteDisabledMutationArtifactsPreserveExecutableDeletes)$'` failed before production code with undefined `sourceProjectionApplyWriteDisabledMutationArtifacts`.
- **Green:** `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestSourceProjectionWriteDisabledLockedSourcesRetainMutationArtifacts|TestSourceProjectionWriteDisabledMutationArtifactsPreserveExecutableDeletes|TestSourceProjectionWriteCapableBundlesDoNotAutoDeferMutations|TestSourceProjectionWriteDisabledMutationArtifactsRequireProviderCitation|TestSourceProjectionWriteDisabledMutationArtifactsRetainGraphQLMutations|TestSourceImportCommandDerivesWriteDisabledMutationArtifacts)$'` passed (1.093s).

## Source-lock regression vectors

- Byte-identical read-only fixtures carry the preserved Sentry and Vercel full
  `operation-source-lock.json` inventories. The test parses 223 Sentry and
  400 Vercel operations, respectively, and derives every test operation from
  that retained source data rather than a hand-written citation.
- The source locks retain Sentry SHA
  `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435`
  (3868570 bytes) and Vercel SHA
  `74cb7ff3dc0b89cc344b13ac9c6d5f1d9b7d7a9356cfd6b5a779da51fd43da28`
  (10463249 bytes).
- The test retains all 103/237 classified Sentry/Vercel mutations as exact
  cited artifacts, and separately validates the real
  `listOrganizationProjects`/`createOrganizationDashboard` and
  `getProjects`/`deleteStorageStoresBlobById` pairs with a declared GET route.
  Projection leaves `writes.json` and `cli_surface.json` byte-identical.
- Current-main acceptance additionally loads the exact source-lock deletes
  `deleteOrganizationDashboard` at
  `/api/0/organizations/{organization_id_or_slug}/dashboards/{dashboard_id}/`
  and `deleteStorageStoresBlobById` at `/storage/stores/blob/{id}`. Each has
  a declaration-owned destructive action plus `availability: implemented`,
  `intent: reverse_etl` CLI route. With `metadata.capabilities.write=false`,
  each remains action-first (zero automatic artifact count) and returns zero
  executable-coverage findings.

## Local command results

- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestSourceProjectionIssue4329SourceLocksRetainMutationInventoryAndReadSurface|TestSourceProjectionWriteDisabledMutationArtifactsPreserveExecutableDeletes|TestSourceProjectionWriteDisabledMutationArtifactsPreserveSourceLockedExecutableDeletes)$'` — PASS (1.118s).
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./cmd/connectorgen` — PASS (154.269s) after merging `origin/main` at `1324c52bab0b224ed8958858af7676b8b8e191b4` with no conflicts.
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./internal/connectors/engine` — PASS (14.192s).
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./internal/connectors/commandrunner` — PASS (24.871s); `make connector-runtime-preflight` — PASS (8.737s).
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./internal/cli` — PASS (438.166s).
- `GOFLAGS=-p=3 go vet ./...`; `GOFLAGS=-p=3 go build ./cmd/pm`; `make tidy-check`; `make lint`; `make docs-check-no-build`; and `make smoke-no-build` — PASS.
- `make agent-contract-check`; `make connectorgen-validate`; `make connectorgen-surface-sync`; `make connectorgen-declaration-admission`; `make connectorgen-operation-evidence`; `make github-parity-artifacts-check`; `make connectorgen-certification-matrix`; `make connectorgen-certification-candidates`; `make connectorgen-certification-sweep`; `make connector-boundary`; `make connector-canon-check`; and `make release-workflow-check` — PASS. The generator validated 553 connectors, source surface sync made zero corrections, declaration admission found zero findings, operation evidence is current at 1,525 rows, and the boundary scan found zero findings.
- `git diff --check` and `gofmt -d` over changed Go files — PASS (no output).

## Built-binary command boundary

This shared foundation adds **zero** connector commands and changes no existing
implemented command. Therefore it has no newly implemented command to probe.
The rebuilt `./pm help sentry` and `./pm help vercel` both return `help topic
... not found` on this branch: those source-locked connector bundles are
intentionally outside this shared-foundation PR. No user command is claimed
usable without a credential-boundary probe. The real downstream gap is
materializing the source-locked Sentry/Vercel read bundles on their owned
source-bound read/execution work; this PR only removes the shared validator
blocker that previously prevented that work. That named downstream foundation
is the only remaining gap to a built-binary credential-boundary proof.

## Remaining remote gates

- Push the exact green SHA to the existing PR and confirm its API-reported
  `base.ref=main`.
- Obtain the Captain-requested fresh, separate Codex audit of the exact pushed
  integration SHA, then disposition any
  findings before requesting a merge.
- Wait for every required PR check to pass. No merge is performed by this task.
