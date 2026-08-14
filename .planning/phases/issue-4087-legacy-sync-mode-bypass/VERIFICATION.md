# Verification: issue #4087

Status: complete (inline/manual GSD fallback)

## Acceptance checklist

- [x] Both aliases resolve to non-empty canonical typed contracts through normal and persisted-legacy parsing.
- [x] Both aliases return a typed pre-I/O `ModeNotExecutableError`, with no legacy source read.
- [x] The mapping is single-sourced and connector-neutral.
- [x] All closed canonical mode names preserve their existing parsed contract/admission behavior.
- [x] Runtime help, generated CLI docs, website docs, and the certification report agree that the aliases are typed admissions with pre-I/O refusal when no transport is admitted.
- [x] Focused tests, formatting, vet, build, and individual repository gates pass.
- [x] Inline/manual GSD verify-work and code-review evidence is complete in `REVIEW.md` and `SUMMARY.md`.

## Commands passed

- `go test -count=1 -timeout 20m ./internal/app`
- `go test -count=1 -timeout 20m ./internal/cli`
- `go test -timeout 20m ./internal/connectors/certify`
- `go test -timeout 20m ./internal/synccontract ./internal/synctransport`
- `go vet ./...`; `go build ./cmd/pm`; `gofmt -w` for changed Go files; `git diff --check`
- `pm help etl`, `pm etl`, and `pm etl --help` show the typed-admission compatibility wording.
- `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors`; `make docs-check`
- `npm run gen:website-data`, `npm run typecheck`, `npm run lint` (pre-existing warnings only), and `npm run build` in `website/`
- `make tidy-check`, `make lint`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connector-boundary`, `make connector-canon-check`, and `make release-workflow-check`
- `scripts/verify-gsd-workflow`
