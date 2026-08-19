# VERIFICATION — issue #3863 secret-free credential coordination identity

Status: passed by inline/manual verification on 2026-08-06.

## Checklist

- [x] A linked compatible pair shares an opaque auth cohort without vault participation in identity derivation. Focused app tests cover creation-time and explicit links; the internal app test removes its temporary vault entry before direct identity derivation.
- [x] A rate key requires explicit policy/kind/non-secret subject; same binding with different subject differs. The typed identity tests prove subject/kind separation and reject empty or unsupported declarations.
- [x] Unlinked credentials remain isolated; mismatched profiles cannot link. The focused app/CLI tests cover compatible and incompatible cross-connector cases.
- [x] No raw binding, secret, revision, or raw rate subject appears in identity JSON, errors, CLI JSON, or runtime state. Tests read the protected binding only to prove it is absent from inspection output; runtime identity retains opaque projections only.
- [x] Legacy credential metadata migrates safely and rotation remains approval-only. The migration test removes the binding, salt, and declaration fields before reopening; approval revision remains a distinct value.
- [x] Runtime config carries opaque identity only; #3754/#3865/#3867 seams are present but unimplemented. The new auth/rate projection types prevent accidental cross-use; no registry, requester, fence, or parking behavior was added.
- [x] CLI runtime help/manual/website/generated docs and bare namespace behavior match the actual commands. `pm help credentials`, `pm credentials`, and `pm credentials --help` returned the updated manual successfully; golden transcripts and generated docs/data were refreshed.
- [x] Targeted tests, package suites, formatting/static checks, repository gates, GSD verify, and code review are recorded below.

## Automated evidence

Passed after the final production change:

- `go test ./internal/connectors -run '^TestCoordinationIdentity' -count=1`
- `go test ./internal/app -run '^TestCredentialCoordination' -count=1`
- `go test ./internal/cli -run '^TestCredentialsCoordination' -count=1`
- `go test ./internal/connectors -count=1`
- `go test ./internal/app -count=1`
- `go test ./internal/cli -count=1`
- `go test ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1`
- `go build ./cmd/pm`; `go vet ./internal/connectors`, `./internal/app`, and `./internal/cli`
- `make tidy-check`; `make lint`; `golangci-lint run ./internal/connectors`
- `make docs-check-no-build`; `pnpm run gen:docs`; `pnpm run typecheck`
- `make agent-contract-check`; `make connectorgen-validate`; `make connectorgen-surface-sync`; `make connector-boundary`; `make release-workflow-check`
- `git diff --check`

`make smoke-no-build` is intentionally not run: its fixture injects a credential value and executes a
reverse-ETL write, both excluded by this identity-only issue. Per repository guidance, the timeout-prone
`go test ./...` and `make verify` monoliths remain CI-owned.

## Base freshness

The branch was rebased cleanly onto `origin/main` at `4d77ef3ed` and the scoped package suites,
command-runner preflight, build, vet, lint, docs/website checks, and repository gates above were
rerun afterward. A final `git fetch origin main` confirmed that `origin/main` remains an ancestor
of this branch.

## GSD and review evidence

- Inline/manual `execute-phase` prompt completed; the canonical contract and runtime do not permit
  role spawning in this worker.
- Inline/manual `verify-work` prompt completed with automated coverage in `UAT.md`.
- Inline/manual standard-depth code review is clean: `REVIEW.md`. The same no-spawn fallback is
  recorded there; no external PR review is requested before the firstmate shipping gate.

## Intentionally excluded

- Rate registry/admission/shared coordinator (#3754), requester policy (#3752/#3753), auth fence
  (#3865), persisted parking (#3867), transport dispatch (#3864), live provider checks, and
  reverse-ETL execution.
