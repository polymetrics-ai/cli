# Verification — issue #4329, r2

## Status

Local verification is complete. PR API base read-back, independent audit, and
required remote CI remain pending after the green commit is pushed.

## Completed evidence

- **Red:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestSourceProjectionWriteDisabledLockedSourcesRetainMutationArtifacts|TestSourceProjectionWriteDisabledMutationArtifactsPreserveExecutableDeletes)$'` failed before production code with undefined `sourceProjectionApplyWriteDisabledMutationArtifacts`.
- **Green:** `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestSourceProjectionWriteDisabledLockedSourcesRetainMutationArtifacts|TestSourceProjectionWriteDisabledMutationArtifactsPreserveExecutableDeletes|TestSourceProjectionWriteCapableBundlesDoNotAutoDeferMutations|TestSourceProjectionWriteDisabledMutationArtifactsRequireProviderCitation|TestSourceProjectionWriteDisabledMutationArtifactsRetainGraphQLMutations|TestSourceImportCommandDerivesWriteDisabledMutationArtifacts)$'` passed (1.093s).

## Source-lock regression vectors

- Sentry: `listOrganizationProjects` and `createOrganizationDashboard` retain
  the preserved source citation URL, SHA
  `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435`,
  byte count `3868570`, and exact OpenAPI locations.
- Vercel: `getProjects` and `deleteStorageStoresBlobById` retain the preserved
  source citation URL, SHA
  `74cb7ff3dc0b89cc344b13ac9c6d5f1d9b7d7a9356cfd6b5a779da51fd43da28`,
  byte count `10463249`, and exact OpenAPI locations.
- In both vectors a supported GET command remains valid while the mutation
  carries the exact source ID/method/path plus the named
  `source-cited-non-executable-mutation-foundation-r1` gap. Projection leaves
  `writes.json` and `cli_surface.json` byte-identical.

## Local command results

- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./cmd/connectorgen` — PASS (153.267s; final run).
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./internal/connectors/engine` — PASS (12.156s).
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./internal/connectors/commandrunner` — PASS (22.660s).
- `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./internal/cli` — PASS (441.520s).
- `go vet ./...` — PASS.
- `go build ./cmd/pm` — PASS.
- `make tidy-check lint agent-contract-check connectorgen-validate connectorgen-surface-sync connectorgen-operation-evidence connector-boundary connector-canon-check release-workflow-check` — PASS. The generator validated 553 connectors; surface sync corrected 0 fields; operation evidence is current (1525 rows).
- `make docs-check-no-build smoke-no-build` — PASS.
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
blocker that previously prevented that work.

## Remaining remote gates

- Push the exact green SHA, open a main-targeted PR, and confirm the API reports
  `base.ref=main`.
- Obtain an independent Codex audit of the exact SHA, then disposition any
  findings before requesting a merge.
- Wait for every required PR check to pass. No merge is performed by this task.
