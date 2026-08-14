# Verification: issue #4087

Status: complete (inline/manual GSD fallback)

## Acceptance checklist

- [x] Both aliases resolve to non-empty canonical typed contracts through normal and persisted-legacy parsing.
- [x] Both aliases return a typed pre-I/O `ModeNotExecutableError`, with no legacy source read.
- [x] `internal/synccontract/public_modes.go` single-sources the public mapping and generic capability projection without connector-specific behavior.
- [x] All closed canonical mode names preserve their existing parsed contract/admission behavior.
- [x] Runtime help, generated CLI docs, website docs, and the certification report agree that the aliases are typed admissions with pre-I/O refusal when no transport is admitted.
- [x] The connector-neutral static projection, generated manuals/catalog, skill surfaces, capability-flow matrices, and their checked generators agree; no certification report contract or version was changed in this review fix.
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
- Review regeneration: `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` (`Generated docs in docs/cli and connector docs in docs/connectors`); `go run ./cmd/pm skills generate --dir docs/skills` (`Generated skills in docs/skills`); `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$'` (`ok polymetrics.ai/internal/cli 22.847s`); `npm run gen:docs` in `website/` (`Wrote 12 docs pages to lib/docs.generated.ts.`); and `go run ./cmd/connectorgen certification-matrix` (exit 0, no stdout).
- Review drift checks: `go run ./cmd/pm docs validate --connectors-dir docs/connectors` (`Validated connector docs in docs/connectors`); `go run ./cmd/connectorgen surface-sync --check` (`552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)`); and `go run ./cmd/connectorgen certification-matrix --check` (exit 0, no stdout).
- Review focused regression: `go test -count=1 -timeout 20m -run '^(TestParseSyncModeMatrix|TestDedupedLegacyAliasesUseTypedContractsBeforeSourceIO|TestCanonicalSyncModesRetainParsedContracts|TestPublicModesResolveCompatibilityNamesToClosedContracts|TestPublicModeCapabilitiesAndDefaultsUseMaterializingModes|TestDerivedSyncModesTruthTable|TestDerivedSyncModesNilSchemaIsNeitherCase|TestValidateConnectorDocsRejectsStaleIconMetadata|TestValidateConnectorDocsRejectsStaleGeneratedContent|TestDocsGenerateAndValidateConnectorDocs|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestSkillsGenerateMatchesTrackedSkills|TestGoldenTranscripts)$' ./internal/app ./internal/synccontract ./internal/connectors/engine ./internal/cli`; app, synccontract, and engine reported `ok` at `10.620s`, `0.445s`, and `1.885s`, and the explicit matching CLI invocation also exited 0.
