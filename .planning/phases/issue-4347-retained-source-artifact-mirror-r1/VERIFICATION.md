# Issue #4347 — verification

## Result

The retained-reader foundation is locally green. Historic-byte preservation is
attempted first; **Firstmate** separately authorized an explicit re-pin only
after the per-artifact response is classified as a real provider document. An
error response remains irrecoverable and is never re-pinned.

## Executed verification

- `gofmt -w cmd/connectorgen/sourceartifact.go cmd/connectorgen/sourceimport.go cmd/connectorgen/sourceimport_test.go` — passed.
- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceRetain|TestSourceImportRetainedArtifact|TestSourceImportVersion2|TestSourceImportRejectsSymlinkedSourcesDirectoryEvenInsideConnectorBundle|TestSourceImportCheckedInGitHubArtifactsAreRetainedAndLockVerified)$' -count=1` — passed in 1.239s. Covers retained machine-readable import without a provider, missing/mismatched/symlinked copies, manifest tampering, rendered reference bytes, zip bundle bytes, raw v2 GraphQL pins, and Elasticsearch/Zoom recovery fixtures.
- `go test -timeout 20m ./cmd/connectorgen -count=1` — final pass in 156.093s.
- `go test -timeout 20m ./internal/cli -count=1` — final pass in 431.272s.
- `go vet ./...`, `go build ./cmd/connectorgen`, and `go build ./cmd/pm` — passed.
- `go run ./cmd/connectorgen source-import --help` and `git diff --check` — passed.
- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceRetain|TestSourceImportRetainedArtifact|TestSourceImportVersion2|TestSourceImportRejectsSymlinkedSourcesDirectoryEvenInsideConnectorBundle|TestSourceImportCheckedInGitHubArtifactsAreRetainedAndLockVerified)$' -count=1` — passed in 1.264s after the final reader/materializer additions.
- `go run ./cmd/connectorgen source-retain github --retrieved-at 2026-08-24T07:02:03Z --license 'GitHub REST API Description: MIT; GitHub Docs GraphQL schema: CC-BY-4.0' --terms 'GitHub REST API Description LICENSE and GitHub Docs Terms of Service; attribution and notices retained in provenance.'` — passed; wrote two lock-verified raw artifacts totaling 14,471,636 bytes and the tracked provenance manifest.
- `go run ./cmd/connectorgen source-import github --check` — final pass in 8.907s: `github, 1525 operation(s), 0 inbound event(s) verified`. The normal import command instantiated only the retained reader.
- `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceImportPreservesFrozenGitHubArtifacts|TestSourceImportCheckedInGitHubArtifactsAreRetainedAndLockVerified|TestSourceRetain)' -count=1` — passed in 1.272s; the intentional GitHub lock re-pin has an updated frozen-byte snapshot.
- `make tidy-check`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, `make release-workflow-check`, and `make lint` — passed. The boundary report checked 317 files/553 connectors with zero findings; generator validation checked 553 connectors with zero findings, surface-sync filled/corrected zero fields, and release certification archive proof passed.
- Full `go test -timeout 20m ./...` / `make verify` were intentionally not run as one local command: repository instruction requires package-scoped checks in this per-command environment because the 550+ connector suite routinely exceeds the command window. CI remains the full-suite authority.

## Acceptance recovery evidence

- Before the write, `go run ./cmd/connectorgen source-import github --check` failed before any network path with `read retained artifact manifest: ... github-retained-artifacts.json: no such file or directory`. After the explicit `source-retain` maintenance command, the same hermetic check passes using the two tracked raw files.
- GitHub REST was preserved at `12,920,264` bytes / `80850d…d5b1d`. GitHub GraphQL was a Firstmate-authorized explicit re-pin from `1,546,421` / `c09aba…c249d` to the observed real GraphQL schema at `1,551,372` / `30347e…40c3d`; `REPIN-REPORT.json` records the old/new identities, HTTP 200 classification, retrieval date, and the complete pre-change tracked-lock Git blob.
- Read-only current-response measurement: all 27 Batch 8–10 nonzero source inputs plus GitHub are 454,667,903 bytes today versus 453,725,820 locked bytes. 351 Batch 8–10 documents differ from their pins; GitHub's GraphQL input also differs. They require the explicit Firstmate-authorized real-document classification and re-pin report before any retained-copy write.
- Zoom accounts: the historic dated URL returned HTTP 404 / 8,329 bytes. Elasticsearch: the historic URL returned HTTP 200 / 6,458,837 bytes versus the lock's 6,458,869 bytes. An exact SHA-256/length scan across every reachable Git blob found no matching historical copy for either artifact.
- The engine tests prove a correct retained copy would permit both cases to import without contacting the provider. They cannot truthfully claim recovery of the real historic Zoom bytes, because those bytes are not present. Elasticsearch is eligible only for Firstmate's documented re-pin path after real-document classification. `LANE-ADOPTION.md` gives each lane's next exact step; this branch does not import their unmerged connector files under Firstmate inbox 004.

## Post-PR main rollups and unavailable-source regression

- Merged `origin/main` at `1b893f348` and then `72fe0ba88` into this branch.
  The first merge had one `sourceimport_test.go` documentation-contract conflict;
  the resolved assertion preserves both retained-artifact and request-bound
  documentation requirements. No generated projection conflicted.
- Red: `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportReportsUnavailableSourceBeforeRequiringRetainedManifest$' -count=1` failed because an all-unavailable v3 Zoom lock stopped at a missing retained-artifact manifest instead of its immutable unavailable reason.
- Green: the same command passed in 1.282s after `source-import` bypassed the
  retained-reader constructor only for a lock with no actual provider artifact;
  the defensive fetcher cannot make a network request. Any lock with an
  artifact still requires the retained reader.
- `go test -timeout 20m ./cmd/connectorgen -count=1` — passed in 338.364s on
  the `72fe0ba88` rollup base.
- The first post-rollup `go run ./cmd/connectorgen source-import github --check`
  correctly found 18 derived CLI projection changes. Ran the canonical
  `go run ./cmd/connectorgen source-import github` generator (writes=0,
  cli=18), then re-ran `source-import github --check`,
  `make connectorgen-surface-sync`, and `make connectorgen-validate`; all
  passed (553 connectors, zero validation findings).
- `go vet ./...` and `make lint` — passed after the rollups and regeneration.
- The full `go test -timeout 20m ./internal/cli -count=1` then exposed only
  tracked generated-skill drift caused by those 18 regenerated GitHub CLI
  projections. Ran the canonical `go run ./cmd/pm skills generate --dir
  docs/skills --json`; it changed only `docs/skills/pm-github/SKILL.md`.
  `go test -timeout 20m ./internal/cli -run '^TestSkillsGenerateMatchesTrackedSkills$' -count=1`
  then passed in 6.215s.
- CI's `Website generated data` check then correctly found four stale website
  projections of those same GitHub availability changes. Ran the exact
  workflow-prescribed `pnpm run gen:website-data` in `website/`; it updated
  `content/docs/github-cli-surface.mdx`, `data/connectors.generated.json`,
  `lib/connectors.catalog.data.generated.json`, and `lib/docs.generated.ts`.

## CI-regression repair after canonical projection

- Red: `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceProjectionDoesNotBlockReadForUnusedOptionalNonScalarParameter$' -count=1` failed because an optional `has` parameter with a non-scalar serialization gap blocked its otherwise field-complete GET route. Current-main schema generation changed the reason wording; the runner can correctly make the request when the filter is omitted.
- Green: `go test -timeout 20m ./cmd/connectorgen -count=1` passed in 331.467s after source projection exempted every optional parameter schema gap, rather than only the prior `oneOf` wording. Required and non-parameter gaps remain blocking.
- Canonical regeneration: `go run ./cmd/connectorgen source-import github`, `go run ./cmd/connectorgen source-import github --check`, `make connectorgen-surface-sync`, and `make connectorgen-validate` passed. The 18 accidental partial GitHub projections returned to their executable source-derived state; `go run ./cmd/pm skills generate --dir docs/skills --json`, `go run ./cmd/pm docs generate --dir docs/cli`, `(cd website && pnpm run gen:website-data)`, and `make docs-check` regenerated matching documentation.
- CI regression proof: `go test -tags github_fixture_sweep -timeout 35m -count=1 -v ./internal/cli -run '^(TestPMBinaryProvesGitHubSharedAdmissionForGeneratedDirectReadFixture|TestPMBinaryExecutesGitHubGeneratedDirectReadCandidatesAgainstFixture|TestPMBinaryExecutesGitHubDisputedPartialVerdictsAgainstFixture|TestPMBinaryExecutesGitHubReleasedReadSurfaceAgainstFixture|TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle|TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip)$'` passed in 781.733s, including 97 generated candidates and all 633 released direct-read routes.
- `go test -timeout 20m ./internal/connectors/certify -run '^TestSurfaceInventoryForGitHubAccountsForAllReviewedEndpoints$' -count=1` passed in 1.733s. `go test -timeout 20m ./internal/cli -run '^TestSkillsGenerateMatchesTrackedSkills$' -count=1` passed in 6.561s. `git diff --check` passed.
- CI caught one remaining frozen source-lock embed snapshot: `TestProductionEmbedPreservesGithubSourceLockBytes` expected the old lock SHA-256 after the documented GraphQL re-pin. Red reproduced locally; after the expected digest changed from `281b1cf…8fb6` to committed-lock `79f6eaf…f2c8`, `go test -timeout 20m ./internal/connectors/defs -run '^TestProductionEmbedPreservesGithubSourceLockBytes$' -count=1` passed in 1.260s and the source-import regression subset passed in 1.404s.
- The next CI pass caught the only remaining generated projection of that re-pin: `.planning/phases/github-parity-extract-r1/GITHUB-COMBINED-OPERATION-LEDGER.json` retained the old GraphQL source hash. `node scripts/github-combined-operation-ledger.mjs --refresh` regenerated it from the unchanged lock without a provider request. `make github-parity-artifacts-check`, `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportPreservesFrozenGitHubArtifacts$' -count=1`, `go test -timeout 20m ./internal/connectors/defs -count=1`, and `go run ./cmd/connectorgen source-import github --check` all passed.
- Proactive post-ledger gates found the one remaining downstream projection: `go run ./cmd/connectorgen certification-subject --check` reported stale. Running the canonical no-network `go run ./cmd/connectorgen certification-subject` updated only `internal/connectors/certifications/current-subject.json`; its check and certification matrix/candidates/sweep checks all passed. `make connectorgen-operation-evidence`, `make connector-canon-check`, and `make release-workflow-check` also passed.

## Inline GSD lifecycle

- Resolved and applied inline due to the task's no-role-spawning contract: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `scripts/gsd prompt discuss-phase 4347`, `plan-phase 4347 --tdd`, `execute-phase 4347`, `verify-work 4347`, and `code-review 4347`.
- This artifact and `TDD-LEDGER.md` record the red/green/refactor evidence. No GSD role was spawned.
