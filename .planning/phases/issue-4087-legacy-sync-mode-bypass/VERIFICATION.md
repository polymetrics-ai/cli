# Verification: issue #4087

Status: complete (inline/manual GSD fallback)

## Acceptance checklist

- [x] Both aliases resolve to non-empty canonical typed contracts through normal and persisted-legacy parsing.
- [x] Both aliases return a typed pre-I/O `ModeNotExecutableError`, with no legacy source read.
- [x] `internal/synccontract/public_modes.go` single-sources the public mapping and generic capability projection without connector-specific behavior.
- [x] All closed canonical mode names preserve their existing parsed contract/admission behavior.
- [x] Runtime help, generated CLI docs, website docs, and the certification report agree that the aliases are typed admissions with pre-I/O refusal when no transport is admitted.
- [x] The connector-neutral static projection, generated manuals/catalog, skill surfaces, capability-flow matrices, and their checked generators agree; no certification report contract or version was changed in this review fix.
- [x] The effective runtime cursor projects either a schema cursor or a validated standalone incremental cursor; an explicitly conflicting pair remains separately rejected, and valid empty-cursor incremental declarations remain supported through their schema cursor.
- [x] Focused tests, formatting, vet, build, and individual repository gates pass.
- [x] Inline/manual GSD verify-work and code-review evidence is complete in `REVIEW.md` and `SUMMARY.md`.

## CI vulnerability repair — 2026-08-14

- [x] The supplied failing job was reproduced from its recorded Go 1.25.12 result; all seven reachable findings are standard-library advisories with Go 1.25.13 fixed versions.
- [x] `GOTOOLCHAIN=go1.25.13 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — passed: `No vulnerabilities found.`
- [x] `GOTOOLCHAIN=go1.25.13 go mod verify`; `GOTOOLCHAIN=go1.25.13 go mod tidy`; and `git diff --exit-code -- go.sum` — passed. The only module-file diff is the intended `toolchain go1.25.12` to `go1.25.13` update; `make tidy-check` is intentionally a clean-worktree assertion and therefore reports that pending expected diff before the outer commit phase.
- [x] `GOTOOLCHAIN=go1.25.13 go test -count=1 -timeout 20m ./internal/app -run '^(TestDedupedLegacyAliasesUseTypedContractsBeforeSourceIO|TestCanonicalSyncModesRetainParsedContracts)$'`; `GOTOOLCHAIN=go1.25.13 go test -count=1 -timeout 20m -run '^$' ./cmd/pm`; `go run ./cmd/agentcontractgen check`; `make release-workflow-check`; and `git diff --check` — passed.
- [x] `git rebase 2df18ee3a083fe507cbe1c07e0270e82c5ab0182` — completed with the branch already up to date. That commit is the exact merge base, so there were no conflict hunks to resolve at the supplied target.

## Certification report contract correction — 2026-08-14

- [x] The report remains schema version 1 and carries typed pre-I/O refusal evidence through the existing empty `data_source` encoding plus `reason`; no new report value, report type, or validation was introduced.
- [x] `go test -count=1 -timeout 20m ./internal/connectors/certify -run '^(TestSourceStagesAgainstSample|TestReportSaveAllowsEmptyPreIORefusalDataSource)$'` — passed.
- [x] `go test -count=1 -timeout 20m ./internal/cli -run '^(TestCertifyCLISingleConnectorTextMode|TestCertifyCLIReportFixtureRendersJSONAndMapsExitCodes|TestCertifyCLISingleConnectorSavesReport)$'` — passed.
- [x] `git diff --exit-code 2df18ee3a083fe507cbe1c07e0270e82c5ab0182 -- internal/connectors/certify/report.go docs/architecture/connector-architecture-v2-design.md docs/architecture/connector-certification-design.md` — passed.

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
- Review regeneration: `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` (`Generated docs in docs/cli and connector docs in docs/connectors`); `go run ./cmd/pm skills generate --dir docs/skills` (`Generated skills in docs/skills`); `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$'` (`ok polymetrics.ai/internal/cli 22.847s`); `npm run gen:docs` in `website/` (`Wrote 12 docs pages to lib/docs.generated.ts.`); and `go run ./cmd/connectorgen certification-matrix` (`generated certification matrices: connectors=556 capability_complete=0 certified=0`).
- Review drift checks: `go run ./cmd/pm docs validate --connectors-dir docs/connectors` (`Validated connector docs in docs/connectors`); `go run ./cmd/connectorgen surface-sync --check` (`552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)`); and `go run ./cmd/connectorgen certification-matrix --check` (`certification matrices are current: connectors=556 capability_complete=0 certified=0`).
- Review focused regression: `go test -count=1 -timeout 20m -run '^(TestParseSyncModeMatrix|TestDedupedLegacyAliasesUseTypedContractsBeforeSourceIO|TestCanonicalSyncModesRetainParsedContracts|TestPublicModesResolveCompatibilityNamesToClosedContracts|TestPublicModeCapabilitiesAndDefaultsUseMaterializingModes|TestDerivedSyncModesTruthTable|TestDerivedSyncModesNilSchemaIsNeitherCase|TestValidateConnectorDocsRejectsStaleIconMetadata|TestValidateConnectorDocsRejectsStaleGeneratedContent|TestDocsGenerateAndValidateConnectorDocs|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestSkillsGenerateMatchesTrackedSkills|TestGoldenTranscripts)$' ./internal/app ./internal/synccontract ./internal/connectors/engine ./internal/cli`; app, synccontract, and engine reported `ok` at `10.620s`, `0.445s`, and `1.885s`, and the explicit matching CLI invocation also exited 0.
- Review schema-cursor reconciliation: `go test -timeout 20m ./internal/connectors/engine -run '^TestDerivedSyncModesTruthTable$'`; `go test -timeout 20m ./cmd/connectorgen -run '^TestValidate_RejectsIncrementalCursorSchemaMismatch$'`; `go run ./cmd/connectorgen validate internal/connectors/defs/chargebee`; `go run ./cmd/pm docs validate --connectors-dir docs/connectors`; and `git diff --check` — passed.
- Review effective-cursor correction: `go test -timeout 20m ./internal/connectors/engine ./cmd/connectorgen -run 'Test(DerivedSyncModesTruthTable|ConnectorIncrementalCursorWithoutSchemaFieldProjectsEffectiveCursor|Validate_RejectsIncrementalCursorSchemaMismatch|Validate_AcceptsIncrementalCursorWithoutSchemaCursor)$'` — passed.
- CI recovery: `go test -count=1 -timeout 20m ./internal/app -run '^(TestQuerySQLScopesConnectionOwnedAndUnattributedViews|TestDedupedLegacyAliasesUseTypedContractsBeforeSourceIO|TestCanonicalSyncModesRetainParsedContracts)$'` and the full `go test -count=1 -timeout 20m ./internal/app` — passed (188.630s); `GOTOOLCHAIN=go1.26.6 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — `No vulnerabilities found`; `go mod verify` — passed.
