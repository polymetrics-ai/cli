# Foundation Post-Fix TDD Ledger r1

This ledger is bound to the frozen 46-finding canonical review in `POSTFIX-REVIEW.md`. A test is recorded as green only when it passes from the same commit as its production change; generated artifacts are part of the green state, not a later cleanup.

| Group | Finding set | Red contract | Green / regression evidence | State |
| --- | --- | --- | --- | --- |
| 1 | B01, B02, B03, B09, B12, W01 | `TestGitHubParityGenerationOrderIsCommutative`; `TestSourceProjectionGapOperationsCannotMasqueradeAsImplemented`; `TestGoogleAdsGeneratedPOSTReadsAcceptDeclaredNestedObjects`; v2 projection digest mutation; deleted route/parameter and semantic update/delete surface-sync cases. | `go test -timeout 20m ./cmd/connectorgen`; Node parity-order and combined-ledger checks; `connectorgen validate`; `surface-sync --check`; all six affected source IDs have installed coverage. | green; remote `d3bf5da0e6a4575628dd76dd94a7522220f9d3df` |
| 2 | B04-B08, W02 | `TestGraphQLOperationVariablesRequiresExactlyOnePaginationDirection`; `TestOperationDirectReadBackwardGraphQLPaginationUsesPreviousPageInfo`; `TestGraphQLIntUsesSigned32BitDomain`; `TestGeneratedGraphQLContractsClassifySecretInputsAndBoundedIdentitySelections`; `TestWebsiteFlagProjectionPreservesEverySafetyProperty`; `TestRenderCommandSurfaceCommandRendersSafetyConstraints`. | Focused generator, engine/schema, commandrunner preflight, website, guide/skills, generated GitHub artifacts, source-import, surface-sync, certification-candidate/sweep checks. | green; remote `0565f3fd6d152b38f2062aac5dd0df29170b6d4e` |
| 3 | B13-B14, B17, B19, B24 | Classified secrets versus ordinary IDs/headers; error-bearing direct/native result; status receipt; SQS success/error receipt. | connsdk, engine, commandrunner, native SQS, App, and CLI receipt suites. | green; remote `58c86d18bd27e55f334cea37f263dd4cdf7540ee` |
| 4 | B15-B16, B18, B21, B23, B25, W03-W04 | Hook sealed bytes/compound receipt; retry/redirect/cancel receipt; >2^53 CLI value; hostile cursor; SQS redirect; idempotency header; minLength witness. | engine, commandrunner, connsdk, native SQS, CLI, and structured-body regressions. | green; remote `b0eb22feb7f413d15f747b3f78d62c6c46e314b9` |
| 5 | B22, W05 | Existing destination collision/foreign file, error cleanup, and symlink race test each fail before publication. | binary output and `go test -race` multipart publication cohorts. | green; remote `2bddbf5387d323a0dbf074074cf43fa2d40b60b5` |
| 6 | B20, B26, B33, B36, W06-W07 | Stale/revoked authorization or stream owner reaches an effect; clone mutation leaks; indeterminate durable commit; expired park retries. | App, transport, coordination, Arrow/race, and auth fence regressions. | green; local atomic boundary pending commit |
| 7 | B27-B32 | Budget stop looks like EOF; self-certification; receipt-free readback; >2^53 cloning/comparison; one shared deadline. | App, synctransport, engine, and provider-readback behavior suites. | pending |
| 8 | B34-B38, W08 | Persisted terminal result is hidden; ambiguous finalization invents a run; CDC accepts swapped artifact; post-checkpoint error returns failure; declared route error disappears. | App, CLI, CDC/restart, transport, and state recovery suites. | pending |
| 9 | B10-B11 | Evidence/certification metadata for another or stale SHA is accepted. | Exact final-SHA evidence, matrix, candidate, and certification checks. | pending |

## Group 1 preserved test provenance

The recovery worktree contains uncommitted Group 1 production-shaped tests and generated fixtures. They are transferred as the reviewed recovery set without modifying their bytes, then run as the red/green evidence recorded above. The supplemental Node transport-order test is explicitly rerun before Group 1 commit because its ordering defect is the B01 manifestation.

## Group 1 full-CLI failure-set disposition (2026-08-21)

The first full `go test -timeout 20m ./internal/cli -count=1` run reached a
terminal **failure** after `792.193s`. The terminal capture was truncated by
the transport, so the four failures extracted from that output are frozen
below but are not represented as the entire Group 1 full-suite set. The later
exact-session terminal capture expands the frozen set immediately after this
table. A later green full run is evidence only after every row has its focused
red/green disposition. No row is treated as an environment-only exception.

| Exact failing test | Reproduced cause | Red evidence | Correction and focused green evidence |
| --- | --- | --- | --- |
| `TestSkillsGenerateMatchesTrackedSkills` | The source-derived GitHub/Google Ads and related connector command surfaces had changed while tracked `docs/skills/**` and matching connector manuals were stale. | The 792.193s full CLI output named this exact test; the tracked generated files differed from `pm skills generate`. | Regenerated only through `go run ./cmd/pm skills generate --dir docs/skills` and `go run ./cmd/pm docs generate --dir docs/cli`; the exact match test is included in the current session-tracked full rerun and must pass before this group closes. |
| `TestGoldenTranscripts/dynamic_connector_bare_json` | The generated GitHub command/help ordering changed, leaving the exact bare dynamic-connector JSON transcript stale. | The 792.193s full CLI output named this exact subtest. | Regenerated only with `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1`; the non-update exact transcript test is included in the current session-tracked full rerun and must pass before closure. |
| `TestReverseETLToGitHubCreatesPullRequestAfterApproval` | Projection replaced the declared GitHub compound-hook follow-up fields `labels` and `reviewers` with only the primary provider body fields, so planning rejected `record.labels` before any provider I/O. | The 792.193s full CLI output failed at `//labels: additional property not allowed`; `TestSourceProjectionPreservesDeclaredHookFollowupFieldsOutsideProviderBody` then failed before the projector repair. | Added generic declaration-owned `hook_fields`, requiring a closed schema, a registered hook, no duplicates, and no primary path/body overlap. `sourceProjectionRetainDeclaredHookFields` preserves only those named bounded fields outside `body_fields`; `TestSourceProjectionPreservesDeclaredHookFollowupFieldsOutsideProviderBody`, `TestValidateWriteHookFieldsRequireClosedSupplementalDeclarations`, and the exact reverse-ETL test now pass. |
| `TestYouTubeAnalyticsReportsDownloadRunsThroughBoundedBinaryExecutor` | The fixed binary operation exposed `{path}` but did not declare the provider resource identity; source sync reduced the command to no flags, so `--resource-name` was unknown. | The 792.193s full CLI output named this exact test and its unknown `--resource-name` failure; `TestDeriveCommandParameterFlagsUsesDeclaredCLIAliasForSafePathPlaceholder` then failed with generated `--path`. | Added a validated path-only declaration alias `cli_name`, mapping `--resource-name` to the closed `path.path` parameter with an explicit 4,096-byte cap. The alias unit test and the installed binary executor test now pass; the latter also proves a 4,097-byte value fails before another provider request. |

## Group 1 later exact-session full-CLI failure-set expansion (2026-08-21)

The non-overlapping session-tracked rerun returned terminal exit `1` after
`1201.193s`; it is red evidence, not a Group 1 closure. Its complete observed
failure set expands the frozen ledger as follows. No further full CLI run may
start until these rows, together with the preceding four, have focused
dispositions.

| Exact failing test | Reproduced cause | Red evidence | Correction and focused green evidence |
| --- | --- | --- | --- |
| `TestGitHubCommandSurfacePlansReverseETLCommand` | The immutable source declares the issue `title` through a scalar `oneOf` (`string` or `integer`). Projection retained that provider schema but generated a named `json` flag requiring JSON syntax, so the ordinary declared `--title Ship connector command plans` route failed before plan creation. | Exact-session full CLI output: `error: invalid JSON for --title: invalid character 'S' looking for beginning of value`; the focused installed test reproduced the same red before repair. | Added `allow_bare_string` only for an already-bounded, named reverse-ETL `json` flag whose concrete closed record schema has a declared string arm. `TestSourceProjectionStringUnionKeepsTextCLIAndProviderArms`, `TestValidateStructuredJSONRecordStringArmRequiresNamedDeclaredStringUnion`, and `TestRecordOverridesBareStringUnionRemainsBoundedAndRejectsMalformedContainers` pass; they retain all source union arms, reject malformed object/array JSON and 8-byte overflow, and add no raw body authority. Regenerated GitHub projection reports `cli=14`; exact installed `TestGitHubCommandSurfacePlansReverseETLCommand` passes. |
| `TestBahmniBareCommandGroupInvalidMultiPartPathIsNotHelp` | The full process exhausted its 20-minute suite budget while this test was active. Its source only creates a temporary project and asserts the invalid nested path yields usage output; it performs no provider/network action. The timeout stack showed unrelated generated-surface loading, not a Bahmni assertion failure. | Exact-session terminal output: `panic: test timed out after 20m0s`, active test `TestBahmniBareCommandGroupInvalidMultiPartPathIsNotHelp`. | Dispositioned as the prior full-suite environment/scheduling timeout, not a product failure: isolated exact runs passed three times (`2.571s`, `2.510s`, `2.490s` test elapsed; outer wall `6.15s` and `5.94s` for the repeated pair). The test still asserts the exact unknown-command and no-help/no-credential-resolution boundaries. No waiver is applied to the later combined full-suite gate. |

### Focused closure reruns

- `TestSkillsGenerateMatchesTrackedSkills` passed from the regenerated surface in `150.025s` (tracked terminal exit `0`).
- The remaining original failure tests passed serially with terminal exit `0`: `TestGoldenTranscripts/dynamic_connector_bare_json` (`2.467s`), `TestReverseETLToGitHubCreatesPullRequestAfterApproval` (`3.246s`), and `TestYouTubeAnalyticsReportsDownloadRunsThroughBoundedBinaryExecutor` (`4.168s`). The serial command reached each test only after its predecessor passed.
- `TestGitHubCommandSurfacePlansReverseETLCommand` passed after the named string-arm repair (`3.108s`); no full `internal/cli` rerun has been used to close Group 1.

## Group 1 atomic closure gates

All below commands ran from this exact pre-commit tree after the final help
renderer and generated artifact refresh. Group 1 intentionally uses focused
engine, generator, and installed-command evidence; the heavyweight combined
App/CLI/all-connectors suite is reserved for the final exact SHA.

- Red/green: `TestGitHubIssueCreateHelpDescribesDeclaredBareStringArm` failed while help printed `--title (json)` and passed after the renderer published `--title (json or string)`. The paired normal plan test also passed.
- Package gates: `go test -timeout 20m ./cmd/connectorgen -count=1` (`105.213s`), `./internal/connectors/engine` (`6.988s`), and `./internal/connectors/commandrunner` (`21.707s`) passed.
- Required idempotence: `TestSourceProjectionGapCreatesCommandFromExistingClosedActionVariant` passed twice with the exact `stats.CLI == 1` assertion intact; `TestSyncBundleDerivesRequiredPathFlagFromRESTParameter` and `TestSyncBundleProviderParameterProjectionIsIdempotent` passed. The following immutable source check then verified `1525` operations with no drift.
- Generated boundaries: source-import `--cache-dir .source-import-qualification-cache.OHUVcH --check`, `surface-sync --check`, GitHub certification candidates `--check`, sweep `--check` (`1616` rows; `1612` CLI commands), and `connectorgen validate` (`552` connectors, `0` findings) passed.
- Surface parity: `make github-parity-artifacts-check` passed all 16 Node tests and both generator checks; regenerated `docs/skills`, `docs/cli`, connector manuals, and `TestGoldenTranscripts` update completed. `TestSkillsGenerateMatchesTrackedSkills` then passed (`169.312s`), and the final focused installed CLI failure-set suite passed (`9.119s`).
- Build hygiene: `make docs-check`, `go build ./cmd/pm`, `go vet ./...`, and `git diff --check` passed.

## Group 1 source-lock cache qualification

- Red: `TestSourceImportCommandUsesExplicitVerifiedCacheRoot` initially failed with `unknown flag "--cache-dir"`; this demonstrated that an empty, qualification-owned cache could not be selected and that the earlier XDG-based observations did not establish a cold cache.
- Green: the command now accepts only an existing non-symlink `--cache-dir`, wraps the normal source fetcher at that root, and never changes the connector-owned URL, digest, byte count, or request authority. The focused cache suite passed: `TestSourceImportArtifactCacheColdSlowFetchWritesOnlyVerifiedBytes`, `TestSourceImportArtifactCacheHitVerifiesWithoutNetwork`, `TestSourceImportArtifactCacheRejectsCorruptionAndOnlyRecoversFromVerifiedFetch`, `TestSourceImportCommandUsesExplicitVerifiedCacheRoot`, and `TestSourceImportCommandContractAndMigrationDocumentation`.
- Live locked GitHub REST source (2026-08-21): a newly empty explicit root at `/Users/karthiksivadas/.treehouse/cli-83d592/57/cli/.source-import-qualification-cache.OHUVcH` completed `go run ./cmd/connectorgen source-import github --cache-dir <root> --check` in `real 9.01` seconds. It created exactly `/Users/karthiksivadas/.treehouse/cli-83d592/57/cli/.source-import-qualification-cache.OHUVcH/80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d.artifact`, with `12,920,264` bytes and SHA-256 `80850db290cde4eb487e0efb587cf27f305e77b6bef96933ed8a09b5169d5b1d`.
- Warm live check: the same exact root and immutable file completed the same command in `real 5.50` seconds; byte count and SHA-256 were rechecked and the artifact modification time remained `2026-08-21T10:55:15Z`. This temporary qualification root is inventoried here and must be removed recoverably before the Group 1 commit.

Correction: the cache root and recovery sentinel were deliberately retained outside
the index until Group 1 remote SHA `d3bf5da0e6a4575628dd76dd94a7522220f9d3df`
was verified, then moved recoverably to Trash. Neither was committed.

## Group 2 red-contract plan (2026-08-21)

- **GSD/manual fallback:** `scripts/gsd doctor`, the five required command
  prompts, and `go run ./cmd/agentcontractgen check` were resolved in this
  single-owner Herdr worktree. The frozen POSTFIX review fixes proceed inline;
  no specialist owns files or commits in this lane.
- **Skills:** `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-lint`, and
  `golang-documentation` were loaded. CLI help/manual/website parity is part
  of this group because the generated GraphQL flags and skills change.
- **B04/B05 red:** a fixed source-backed connection must reject neither/both
  `first`/`last` and cursor-without-direction before I/O; a `last` page must
  use `hasPreviousPage`/`startCursor`, reject malformed backward state, and
  never substitute an unrelated forward cursor.
- **B06/B07 red:** source-derived query scalar `invitationToken` must be
  env-only; classified `tempCloneToken` and `verificationToken` must not be
  selected; bounded `createIssue`/`addComment` selections retain source-owned
  IDs/URLs plus `clientMutationId`, while ordinary token-count and occurrence
  IDs remain unchanged. No caller document, selection, or raw-body channel is
  added.
- **B08 red:** GraphQL `Int` exact signed-32-bit boundaries must pass and
  adjacent values must fail in root, nested object, and list variables before
  any HTTP request.
- **W02 red:** every safety-relevant flag property (`env_only`, byte/item
  bounds, repeatability, allow-empty, format, required, values, maps-to, and
  numeric minimum) must survive website projection and generated skills/help
  must state env-only and active limits.

### Group 2 observed red evidence

- `go test -timeout 20m ./internal/connectors/engine -run 'Test(GraphQLOperationVariablesRequiresExactlyOnePaginationDirection|OperationDirectReadBackwardGraphQLPaginationUsesPreviousPageInfo|GraphQLIntUsesSigned32BitDomain)$' -count=1` failed as intended: neither direction and cursor-without-direction reached variable validation, mixed direction reported the older non-exact message, backward page state completed from `hasNextPage`, and schema compilation rejected the undeclared `minimum` keyword.
- `node --test scripts/tests/gen-github-graphql-parity.test.mjs` failed as intended with the complete frozen B06/B07 set: invitation token stayed inline; `tempCloneToken`/`verificationToken` remained selected in the five source-owned fixed documents; and `createIssue`/`addComment` omitted nested provider identity.
- `node --test website/scripts/cli-surface.test.mjs` failed as intended: website projection dropped `env_only`, `max_bytes`, `min_items`, and `max_items`.
- `go test -timeout 20m ./internal/connectors -run '^TestRenderCommandSurfaceCommandRendersSafetyConstraints$' -count=1` failed as intended: generated guide text rendered only `--input (required)` and omitted env-only, non-empty, item, and byte constraints.

### Group 2 green evidence

- **B04/B05:** `TestGraphQLOperationVariablesRequiresExactlyOnePaginationDirection`,
  `TestGraphQLOperationVariablesRejectsMixedPaginationDirections`, and
  `TestOperationDirectReadBackwardGraphQLPaginationUsesPreviousPageInfo` pass
  through the real engine. The fixed connection accepts exactly one of
  `first`/`last`, rejects neither, both, and cursor-without-direction before
  transport, and derives backward continuation solely from
  `hasPreviousPage`/`startCursor`.
- **B06/B07:** `TestGeneratedGraphQLContractsClassifySecretInputsAndBoundedIdentitySelections`
  passes. Source-declared query invitation tokens are `env_only` and map only
  to their exact root variable under a source-declared env/redaction policy;
  no raw document/body or undeclared variable form was added. Classified
  result tokens are absent from fixed selections. Mutation selections are
  source-derived, cycle/depth bounded (`3`), field-budget bounded (`64`), and
  retain `clientMutationId` plus provider IDs/numbers/URLs for `createIssue`
  and `addComment`.
- **B08:** `TestGraphQLIntUsesSigned32BitDomain` passes root, nested object,
  and list values at `[-2147483648, 2147483647]`, and adjacent values reject
  before I/O. `TestValidateFlagMaximumIsOptIn` proves the generated CLI flag
  maximum is independently enforced. The schema compiler now keeps exact
  numeric bounds rather than float-rounded substitutions.
- **W02:** `TestWebsiteFlagProjectionPreservesEverySafetyProperty` and
  `TestRenderCommandSurfaceCommandRendersSafetyConstraints` pass. Snake/camel
  projections retain env-only, byte/item/numeric bounds, repeatability,
  allow-empty/bare-string, type/values/mapping/format/requiredness; generated
  guide command lists render active safety qualifiers without hiding the
  command surface.
- **Static closure:** the red `TestValidate_CLISurfaceEnvOnlyFlagRequiresDeclaredSecretGraphQLContract`
  now proves the validator allows only a source-schema-declared scalar
  `graphql_query` variable with `input_mode=env`, `transform=none`, and exact
  `redact_fields`, while still refusing an omitted policy. The established
  typed-secret GraphQL mutation form remains unchanged.
- **Generation:** `connectorgen validate` reports `552 connector(s)` and zero
  findings; `surface-sync --check` reports zero drift; GitHub parity artifacts,
  source-drift, combined-ledger, certification-matrix, certification-candidates,
  and certification-sweep checks passed. The sweep remains exactly `1616`
  rows / `1612` CLI commands. `pm docs generate --dir docs/cli`,
  `pm skills generate --dir docs/skills`, and `pm docs validate --connectors-dir
  docs/connectors` regenerated/validated the tracked manual and skill surfaces.

## Group 3 red-contract plan (2026-08-21)

- **B13/B14:** `TestPublicReceiptProjectionMasksOnlyExactConfiguredAndDeclaredScalars` will drive quote, backslash, `<`, `>`, `&`, non-ASCII, short values (`id`, `token`, `0`, and one character), and encoded configured values through raw and decoded REST receipts. It must prove canonical public JSON has no concrete/proven encoding, while keys (`occurrence_id`, `trained_tokens`), header names (`WWW-Authenticate`), repeatable ordinary values, and the internal receipt remain unchanged. `TestGraphQLErrorMetadataDoesNotKeywordRedactOrdinaryProviderWords` will preserve the ordinary provider message `Unknown token type` and leave exact-value masking to the public result boundary.
- **B17:** `TestCommandRunnerPreservesLegacyPostProviderResultWithoutReceipt` and `TestAshbyOperationDirectReadPreservesEngineResultOnEnvelopeFailure` will require result-plus-error propagation for legacy, operation, navigation, and Ashby logical-envelope failures.
- **B19:** `TestOperationStatusCheckPreservesPostResponseFailureResult` will make a declared bounded-header validation error after a received HEAD response retain its operation/status/path/result; receipt and CLI tests then require a bounded complete 204/404/error envelope without body decoding.
- **B24:** `TestOperationDirectReadPreservesCompleteSQSReceiptOnSuccessAndProviderError` will table-drive ordinary 200 XML and terminal 4xx XML with repeated headers, raw byte count, decoded success body, and a received result on failure. Malformed XML, cap+1, and body-read errors share that same received-response path.
- **GSD/manual fallback:** the exact same single-owner, no-specialist fallback recorded for Group 2 applies. The red package set is connector output, engine, commandrunner, native Amazon SQS/Ashby, and CLI; no generated surface changes are expected for this behavioral group.

### Group 3 observed red evidence

- TestPublicReceiptProjectionMasksOnlyExactConfiguredScalars initially failed
  because public sanitization rewrote the provider-owned occurrence_id key to
  occurrence_[masked]. TestGraphQLErrorMetadataDoesNotKeywordRedactOrdinaryProviderWords
  initially replaced Unknown token type merely because it contained token.
- TestCommandRunnerPreservesLegacyPostProviderResultWithoutReceipt,
  TestOperationStatusCheckPreservesPostResponseFailureResult,
  TestOperationDirectReadPreservesCompleteSQSReceiptOnSuccessAndProviderError,
  and TestAshbyOperationDirectReadPreservesEngineResultOnEnvelopeFailure
  each initially observed a zero result after a received provider response or
  logical provider-envelope failure.
- The review-added TestPublicReceiptProjectionPreservesRawJSONBytesWhenNoMaskApplies
  failed with canonicalized/reordered JSON although the configured credential
  did not occur. Its opaque-byte companion proves a short configured value
  does not erase provider bytes that merely contain that substring.
- The review-added TestRunOmitsResultEnvelopeBeforeProviderResponse initially
  failed for legacy direct-read, status, and binary transport errors: each
  returned an empty result envelope even though no provider response existed.

### Group 3 green evidence

- B13/B14: engine receipts retain immutable raw headers/body; the public
  projection masks only exact configured or declared scalar values, keeps
  provider map/header identities intact, preserves unmodified JSON/base64
  bytes byte-for-byte, masks a real longer opaque credential without making a
  fabricated partial binary, and keeps ordinary GraphQL wording intact.
- B17/B19/B24: runner results retain legacy/direct/binary/status evidence
  with a non-nil error, the CLI emits the matching bounded result envelope
  even without a receipt, status HEAD retains raw metadata without decoding a
  body, and SQS/Ashby preserve complete provider receipts on ordinary,
  terminal, malformed, and logical-envelope paths.
  The runner now omits all three result forms when an executor error occurs
  before provider response evidence, preserving the fail-before-I/O boundary.
- Exact focused package gates from this final Group-3 tree passed: connectors
  (0.615s), engine (10.355s), commandrunner (21.832s), native Amazon SQS
  (1.245s), native Ashby (1.050s), and the focused App direct-write receipt
  suite (3.666s). The focused CLI receipt/envelope suite, including the
  no-receipt result-envelope regression, passed in 1.156s.
- go build ./cmd/pm and git diff --check passed. This behavioral group changes
  no connector declaration or generated surface; generator checks remain
  reserved for their owning source/surface groups and the final exact-SHA
  combined gate.

## Group 4 red-contract plan (2026-08-21)

- **GSD/manual fallback:** this remains the single-owner Herdr lane described
  for Groups 2 and 3. The frozen review is the plan authority; tests are added
  and observed red before their owning production correction. No outside
  worker, component branch, or recovery worktree is modified.
- **B15:** a handled GitHub/Ashby write must either execute a prepared, sealed
  ordered request plan or fail before I/O. Mutating the caller record while
  approval blocks must neither alter the wire bytes nor produce an unpreviewed
  effect. A compound failure must retain every attempted provider receipt in
  order and count only completed effects.
- **B16:** REST, GraphQL, form, and multipart operation writes must dispatch
  bytes/config/headers sealed at preview. Mutation of nested request/config/
  secret/digest state or a replaced multipart file after preview must result
  in the original material or a pre-I/O refusal; no live remarshal/reread is
  admitted.
- **B18:** refused 302/307 and terminal 429/503 followed by cancellation must
  retain the latest typed provider response with bounded metadata while never
  contacting a redirect target.
- **B21:** REST and GraphQL command values preserve exact numeric lexemes
  (including `9007199254740993`, `0.10000000000000001`, negative, and exponent
  forms) and compare exact declared minima before I/O.
- **B23/B25:** opaque pagination continuations and native SQS tokens reject
  unknown/duplicate fields, controls, malformed encoding, and capped-size
  expansion before authentication/request dispatch; a native SQS 302/307 must
  not contact the target or forward a session token.
- **W03/W04:** runtime-owned idempotency headers fail declaration validation
  rather than silently disappearing; minimum-witness generation produces a
  valid bounded string for a supported `minLength` schema.

### Group 4 observed red evidence

- **B15:** the frozen review established that `executeApprovedWrite` let
  `WriteHook.ExecuteWrite` choose physical requests after the preview, so the
  request set itself was absent from the approval digest even after receipt
  retention was repaired. Red-first
  `TestPreparedWriteHookSealsEveryPhysicalRequestAndRetainsTerminalReceipts`
  initially failed to build with the required plan fields/types absent
  (`PreparedRequest.Action`, `ResponseBinding`, and
  `PreparedWriteHookPlan`). The executable red requires two declaration-owned
  steps in preview order, a bounded first-receipt `id` binding for the second
  path, caller mutation after planning to remain off both wire bodies, and
  ordered 201/400 receipts to remain in the terminal result.
- **B16:** `TestPrepareOperationDirectWriteSealsNestedVariablesAndRuntimeMaps`
  initially failed with `prepared variables` carrying `mutated-name`; after
  sealing nested values it then failed with `prepared runtime config =
  map[tenant:mutated-tenant] / secret:mutated-secret`. Both failures occurred
  before I/O. The prepared JSON/form dispatch now consumes the digest-bound
  bytes, while multipart retains its immediate approved-digest revalidation.
- **B18:** `TestRequesterRetryTransportFailureRetainsEarlierProviderResponse`
  initially failed because `errors.As` found no `*HTTPError` after the first
  503 was followed by a transport failure. It now requires both the retained
  503 response/headers/body and the terminal transport cause. The companion
  stream redirect/backoff-cancellation tests cover the open-body branch.
- **B21:** `TestCoerceFlagValuePreservesExactNumericLexemes` observed
  `9007199254740993` as an `int`, `0.10000000000000001` as `float64(0.1)`, and
  `-1.25e-3` as `float64(-0.00125)`. Exact `json.Number` transport plus rational
  bound comparison replaces both conversion paths. The prior focused sweep
  exposed four stale integer/number expectations; they now assert the exact
  lexemes rather than reintroducing float coercion.
- **B23/B25:** the cursor red admitted `admin=true` on a same-origin link URL;
  the SQS red sent a CR/LF cursor to its endpoint. The redirect red delivered
  the session token to the redirected target. The added cursor/SQS tests prove
  pre-I/O refusal, no target contact, a retained redirect receipt, and the
  original `redirect refused` error identity together.
- **W03/W04:** `Idempotency-Key` and `X-Idempotency-Key` declaration validation
  initially returned nil; the minLength witness red returned `cannot prove a
  schema-valid string witness` for a closed `minLength:3` body. Both have
  focused green tests in engine.

### Group 4 focused green evidence to date

- `go test -timeout 20m ./internal/connectors/engine -run
  'Test(PrepareOperationDirectWriteSealsNestedVariablesAndRuntimeMaps|PreparedWriteHookSealsEveryPhysicalRequestAndRetainsTerminalReceipts|LegacyWriteHookClaimIsRefusedBeforeUnpreviewedTransport|OperationHeaderDeclarationsRejectRuntimeOwnedIdempotencyNames|StructuredRESTBodyMinimumWitnessHonorsMinLength|DirectReadCursorURLAdmitsOnlyBoundedDeclaredContinuationQuery)$' -count=1`
  passed.
- **B15 closure:** `PreparedWriteHook` now admits only named existing write
  declarations plus one bounded scalar JSON response field mapped to one
  earlier-step declared path field. The engine validates every selected action
  record/body, caps plans at eight physical requests per source record, seals
  the flattened action/binding/body/header set into the preview digest, and
  executes the same private ordered plan. A legacy `WriteHook` that claims an
  action fails before I/O. `TestPreparedWriteHookSealsEveryPhysicalRequestAndRetainsTerminalReceipts`
  passed against a real two-request server (create `201`, bound update `400`),
  retaining both complete provider receipts. `TestPreparedWritePlanEnumeratesEveryGitHubCompoundRequest`
  fixes all eight GitHub compound action variants and their exact declaration
  identities; `TestGitHubPreparedPlanExecutesOnlyPreviewedStepsAndRetainsTerminalReceipt`
  drives the real GitHub bundle through `engine.Write`, stops before an
  unplanned reviewers request after metadata `400`, and retains the create and
  terminal metadata receipts. Ashby's one-step native envelope route now uses
  the same plan/receipt boundary and its package suite passes.
- `go test -timeout 20m ./internal/connectors/connsdk -run
  'Test(RequesterRetryTransportFailureRetainsEarlierProviderResponse|RequesterMutationRetryCancellationRetainsLastResponse|DoStreamRetainsLastProviderEvidenceAcrossRedirectAndCancelledRetry|DoStreamDisableRetriesRejectsMutationRedirect)$' -count=1`
  passed.
- `go test -timeout 20m ./internal/connectors/commandrunner -run
  'Test(CoerceFlagValueNumber|CoerceFlagValuePreservesExactNumericLexemes|ValidateFlagNumericBoundsUseExactDeclarationLexemes|BuildWriteCommandPlansReopenAndPRSharedCommands|RecordOverridesBuildsExplicitNestedScalarFields|BuildOperationDirectWriteCommandUsesTypedInputsAndPlanLifecycle)$' -count=1`
  passed.
- `go test -timeout 20m ./internal/connectors/native/amazon-sqs -run
  'Test(ApprovedDestructiveWriteRefusesRedirectToUnapprovedTarget|OperationDirectReadRefusesSQSRedirectWithoutForwardingSessionToken|OperationDirectReadRefusesUnsafeSQSContinuationBeforeSigning)$' -count=1`
  passed.

### Group 4 atomic closure gates (2026-08-21)

- Focused behavioral gates passed from this exact uncommitted tree: engine
  (`0.993s`), GitHub prepared-plan (`1.052s`), Ashby native write (`1.047s`),
  connsdk retry/redirect (`0.335s`), commandrunner exact-number/write-plan
  (`1.035s`), SQS redirect/cursor (`0.935s`), and CLI structured-body/help
  (`1.160s`).
- Full affected package gates passed: `internal/connectors` (`0.618s`),
  `engine` (`8.931s`), `connsdk` (`3.463s`), `commandrunner` (`21.973s`),
  `hooks/github` (`5.031s`), `native/ashby` (`1.047s`), and
  `native/amazon-sqs` (`1.287s`).
- Transport/handler race gates passed: engine sealed-plan/cancellation
  (`2.132s`) and connsdk retry/stream boundary (`1.341s`) with `-race`.
- `go vet` over engine, connsdk, commandrunner, GitHub, Ashby, SQS, and CLI;
  `go build ./cmd/pm`; `connectorgen validate` (552 connectors, zero
  findings); `connectorgen surface-sync --check` (zero drift); and
  `git diff --check` all passed. No generated artifact changed in this group.

The B15 planner boundary and the frozen Group-4 package/generator gates are
closed and committed in remote-verified Group-4 SHA
`b0eb22feb7f413d15f747b3f78d62c6c46e314b9`; no recovery, credential, or
generated artifact was included.

## Group 5 red-contract plan (2026-08-21)

- **B22:** a binary download with `allow_overwrite=false` must leave no final
  name while its owned hidden temp is staging; a foreign final inserted before
  publication must survive byte- and inode-identically, and all owned temp
  entries must be removed after the link conflict. `TestBinaryDownloadNoOverwritePublicationIsCrashAndRaceSafe`
  is production-shaped: its reader blocks after writing staged bytes, letting
  the test observe the pre-publish directory and install the competing file.
- **W05:** the multipart escaping-symlink refusal must wait for handler
  completion before observing whether secret bytes reached the server. The
  handler has a request-completion channel and the test disables retries so the
  one observation has a single owner.

### Group 5 observed red evidence

- **B22:** before the state-machine correction,
  `TestBinaryDownloadNoOverwritePublicationIsCrashAndRaceSafe` failed at
  `final name exists before publication: <nil>`. The old `O_EXCL` reservation
  exposed a zero-byte final while the reader was blocked, so a crash could
  leave it and a later rename/cleanup could overwrite/delete a foreign file.
- **W05:** `go test -race -timeout 20m ./internal/connectors/connsdk -run
  '^TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation$'
  -count=1` reported a concrete data race between the test's read of
  `sawFile` and `uploadEcho`'s handler write. That result is retained as the
  red proof; the test was not treated as a reliable security assertion.

### Group 5 green evidence to date

- B22 now stages only an owned hidden temp, `Sync`s that file, uses
  `os.Root.Link` as the atomic no-replace final-name claim, removes only the
  owned temp, and syncs the containing `os.Root` directory. The focused test
  kills a helper process only after its first temp bytes are staged and proves
  no final name remains; its in-process race half preserves the pre-publish
  foreign sentinel byte-for-byte/on the same inode and removes the owned temp.
- W05's focused symlink test passes under `-race` after waiting on the handler
  completion channel before reading the observation.

### Group 5 atomic closure gates (2026-08-21)

- `go test -timeout 20m ./internal/connectors/engine -count=1` passed
  (`10.716s`); `go test -timeout 20m ./internal/connectors/connsdk -count=1`
  passed (`5.656s`).
- Race gates passed: the binary crash/no-replace test (`2.208s`) and the
  multipart escaping-symlink refusal repeated twenty times (`1.398s`) under
  `-race`.
- `go vet ./internal/connectors/engine ./internal/connectors/connsdk`,
  `go build ./cmd/pm`, and `git diff --check` passed. This group changes no
  generated declaration or CLI/manual surface.

## Group 6 red-contract evidence (2026-08-21)

- **B33:** `TestCloneRuntimeConfigDefensivelyCopiesCatalogNestedState` failed
  before the production correction. Mutating the request clone's catalog field,
  primary-key/cursor slices, raw schema, and discovery failure changed the
  caller-owned runtime; its first failure reported the source field renamed to
  `clone`. This establishes the nested aliasing boundary independently of
  ordinary record cloning. Arrow request cloning will use the same deep runtime
  clone and must create a fresh apply request for each segment.
- **B20:** `TestOrchestratorRevalidatesDestinationAuthorizationImmediatelyBeforeApply`
  failed with a nil run result after the test revoked authority in the completed
  warehouse-stage callback. The existing admission happened only before stage,
  and `ApplyDestination` still executed. This proves the live check was not at
  the external mutation boundary.
- **B36:** `TestRequesterRechecksRequestAdmissionBeforeRetry` initially failed
  to compile because the HTTP request boundary exposed no admission hook at
  all. The existing cohort wrapper checked only once around an operation, so a
  retry/page/send could not re-read a durable fence after the first request.
  `TestAuthCohortFenceIndeterminateCommitCancelsOldLocalEpoch` then failed at
  its one-second cancellation boundary: a post-rename directory-sync error had
  already persisted the fence, but `Fence` returned early without cancelling
  old local admissions.
- **B26:** `TestTransportTwoAppsFenceBeforeAnySideEffect` first failed under
  the former late-checkpoint model: the second App reached source/stage/apply
  before its state CAS could lose. The replacement holds the first owner at
  its pre-I/O source boundary and proves the contender has exactly zero
  source, stage, apply, and publication calls.
- **W06/W07:** the post-rename Create test initially left zero live timers
  despite one durable parked record, proving memory had returned on an
  indeterminate write without reloading it. The terminal authorization test
  initially failed to compile because parking had no closed
  `needs_reauthorization` outcome or cross-coordinator claim refusal.

## Group 6 focused green evidence (2026-08-21)

- **B20:** ordinary/full-overwrite and Arrow paths call the standing approval
  exactly once immediately before each apply/publication effect. The prior
  per-unit regression is now `TestOrchestratorRechecksAuthorizationImmediatelyBeforeEachApplyAndRefusesSecondEffect`: two pages stage, only the first provider apply occurs after the second final check revokes authority. The final-after-stage ordinary, full-overwrite, serial Arrow, and pipeline Arrow tests pass.
- **B26:** `TestTransportTwoAppsFenceBeforeAnySideEffect` and the ordinary,
  per-page, full-overwrite, and CDC/restart focused App cohort pass. A durable
  `ActiveWorkID`/monotonic fence is claimed before source I/O, renewed at each
  source boundary, and retired only by the matching terminal work owner;
  terminal/crashed owners are safely replaced while live owners have zero
  contender effects.
- **B33:** `TestCloneRuntimeConfigDefensivelyCopiesCatalogNestedState` passes.
  Runtime catalogs now deep-copy stream fields, keys, cursors, raw schemas and
  discovery failures; serial and pipeline Arrow calls receive fresh cloned
  request objects. The focused Arrow clone cohort passes under `-race`.
- **B36:** retry and redirect request-admission tests pass; auth-cohort
  indeterminate fence cancellation and the PostgreSQL pre-I/O admission test
  pass. HTTP sends, redirects, refreshes, PostgreSQL connects/queries, and
  transaction statements re-check the admitted durable epoch at their physical
  boundary while cleanup rollback remains available.
- **W06:** `TestRateParkingCoordinator_ReconcilesPostCommitCreateBeforeReturningUncertainty`, `TestRateParkingCoordinator_ReconcilesEachPostCommitResumeMutation`, and `TestRateParkingCoordinator_ReconcilesPostCommitRearmAndDelete` pass. They inject an indeterminate post-rename outcome after each of Create, Rearm, Claim, BeginResume, MarkResumeCompleted, Complete, and Delete, then assert the exact reloaded record set and one-or-zero correct timer outcome.
- **W07:** expired and revoked App authorization errors become a secret-free
  durable `needs_reauthorization` parking outcome. The memory and reopened
  file-store tests prove no retry timer or second-coordinator claim can revive
  it; its opaque scope remains blocked until explicit safe cancellation.

### Group 6 atomic closure gates

- `go test -timeout 20m ./internal/coordination -count=1` (`4.276s`),
  `./internal/connectors/connsdk` (`5.983s`),
  `./internal/connectors/native/postgres` (`1.471s`), and
  `./internal/synctransport` (`0.766s`) passed.
- Focused App work-fence, full-overwrite, CDC recovery, and terminal
  authorization cohort passed (`6.773s`); the scoped Arrow clone cohort passed
  under `go test -race` (`1.821s`).
- `go vet` over every Group 6 package, `go build ./cmd/pm`, and
  `git diff --check` passed. No declaration or generated-surface source
  changed, so the final full generator sweep remains reserved for the exact
  all-group SHA.

## Group 1 frozen GitHub mutation delta crosswalk

`874 -> 906` is held provisional until this crosswalk, the runner-bound proof,
and the generator suite are green. The independent test
`TestGitHubFoundationMutationDeltaHasUniqueClosedBoundedSourceCrosswalk`
compares the base candidate artifact at
`c9824b5837f487acaa2c2a39126d29cf401d7fb5` with the generated artifact,
requires exactly these 32 unique command paths, rejects duplicate write actions
or source identities, checks the `fixture_required_mutations`/`reverse_etl`/
`reverse_plan` cohort, and rechecks the immutable source URL, SHA-256, and
12,920,264-byte count. Every root has `additionalProperties: false`.

Bound legend: `S32` is a commandrunner-enforced 32,768-byte UTF-8 cap plus
schema `maxLength: 8192`; `S1120` is 1,120 bytes plus `maxLength: 280`; `J1M`
is the one-value commandrunner JSON cap of 1,048,576 bytes; `A256` is
`maxItems: 256`; `O256` is `maxProperties: 256` on the named dynamic object;
`C` is a recursively closed, bounded named JSON schema; `I`, `B`, and `E` are
the runner's parsed integer, parsed boolean, and finite enum forms. No row has
a raw body, method, path, or header channel.

| Unique command path | Immutable source identity | Write declaration | Method / path | additionalProperties | Certified cohort | Effective runner-enforced inputs |
| --- | --- | --- | --- | --- | --- | --- |
| `api agents set-selected-repos-for-org-secret` | `agents/set-selected-repos-for-org-secret` | `agents_set_selected_repos_for_org_secret` | `PUT /orgs/{org}/agents/secrets/{secret_name}/repositories` | false | fixture_required_mutations | org S32; secret_name S32; selected_repository_ids J1M+A256 |
| `api agents set-selected-repos-for-org-variable` | `agents/set-selected-repos-for-org-variable` | `agents_set_selected_repos_for_org_variable` | `PUT /orgs/{org}/agents/variables/{name}/repositories` | false | fixture_required_mutations | name S32; org S32; selected_repository_ids J1M+A256 |
| `api agents update-org-variable` | `agents/update-org-variable` | `agents_update_org_variable` | `PATCH /orgs/{org}/agents/variables/{name}` | false | fixture_required_mutations | name S32; org S32; selected_repository_ids J1M+A256; value S32; visibility E |
| `api code-scanning update-alert` | `code-scanning/update-alert` | `update_code_scanning_alert` | `PATCH /repos/{owner}/{repo}/code-scanning/alerts/{alert_number}` | false | fixture_required_mutations | alert_number I; assignees J1M+A256; create_request B; dismissed_comment S1120; dismissed_reason E; state E |
| `api dependabot update-alert` | `dependabot/update-alert` | `update_dependabot_alert` | `PATCH /repos/{owner}/{repo}/dependabot/alerts/{alert_number}` | false | fixture_required_mutations | agent_assignment J1M+C; alert_number I; assignees J1M+A256; dismissed_comment S1120; dismissed_reason E; state E |
| `api git create-ref` | `git/create-ref` | `create_ref` | `POST /repos/{owner}/{repo}/git/refs` | false | fixture_required_mutations | ref S32; sha S32 |
| `api git update-ref` | `git/update-ref` | `update_ref` | `PATCH /repos/{owner}/{repo}/git/refs/{ref}` | false | fixture_required_mutations | force B; ref S32; sha S32 |
| `api issues add-assignees` | `issues/add-assignees` | `add_issue_assignees` | `POST /repos/{owner}/{repo}/issues/{issue_number}/assignees` | false | fixture_required_mutations | assignees J1M+A256; issue_number I |
| `api issues add-labels` | `issues/add-labels` | `add_issue_labels` | `POST /repos/{owner}/{repo}/issues/{issue_number}/labels` | false | fixture_required_mutations | issue_number I; labels J1M+A256 |
| `api issues create-milestone` | `issues/create-milestone` | `create_milestone` | `POST /repos/{owner}/{repo}/milestones` | false | fixture_required_mutations | description S32; due_on S32; state E; title S32 |
| `api issues remove-assignees` | `issues/remove-assignees` | `remove_issue_assignees` | `DELETE /repos/{owner}/{repo}/issues/{issue_number}/assignees` | false | fixture_required_mutations | assignees J1M+A256; issue_number I |
| `api issues set-labels` | `issues/set-labels` | `set_issue_labels` | `PUT /repos/{owner}/{repo}/issues/{issue_number}/labels` | false | fixture_required_mutations | issue_number I; labels J1M+A256 |
| `api issues update-comment` | `issues/update-comment` | `update_issue_comment` | `PATCH /repos/{owner}/{repo}/issues/comments/{comment_id}` | false | fixture_required_mutations | body S32; comment_id I |
| `api issues update-milestone` | `issues/update-milestone` | `update_milestone` | `PATCH /repos/{owner}/{repo}/milestones/{milestone_number}` | false | fixture_required_mutations | description S32; due_on S32; milestone_number I; state E; title S32 |
| `api pulls create-review-comment` | `pulls/create-review-comment` | `create_review_comment` | `POST /repos/{owner}/{repo}/pulls/{pull_number}/comments` | false | fixture_required_mutations | body S32; commit_id S32; in_reply_to I; line I; path S32; position I; pull_number I; side E; start_line I; start_side E; subject_type E |
| `api pulls dismiss-review` | `pulls/dismiss-review` | `dismiss_pull_request_review` | `PUT /repos/{owner}/{repo}/pulls/{pull_number}/reviews/{review_id}/dismissals` | false | fixture_required_mutations | event E; message S32; pull_number I; review_id I |
| `api pulls request-reviewers` | `pulls/request-reviewers` | `request_reviewers` | `POST /repos/{owner}/{repo}/pulls/{pull_number}/requested_reviewers` | false | fixture_required_mutations | pull_number I; reviewers J1M+A256; team_reviewers J1M+A256 |
| `api pulls submit-review` | `pulls/submit-review` | `submit_pull_request_review` | `POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews/{review_id}/events` | false | fixture_required_mutations | body S32; event E; pull_number I; review_id I |
| `api pulls update-review-comment` | `pulls/update-review-comment` | `update_review_comment` | `PATCH /repos/{owner}/{repo}/pulls/comments/{comment_id}` | false | fixture_required_mutations | body S32; comment_id I |
| `api repos add-collaborator` | `repos/add-collaborator` | `add_collaborator` | `PUT /repos/{owner}/{repo}/collaborators/{username}` | false | fixture_required_mutations | permission S32; username S32 |
| `api repos create-commit-comment` | `repos/create-commit-comment` | `create_commit_comment` | `POST /repos/{owner}/{repo}/commits/{commit_sha}/comments` | false | fixture_required_mutations | body S32; commit_sha S32; line I; path S32; position I |
| `api repos create-deployment` | `repos/create-deployment` | `create_deployment` | `POST /repos/{owner}/{repo}/deployments` | false | fixture_required_mutations | auto_merge B; description S32; environment S32; payload J1M+O256+maxLength 1,048,576; production_environment B; ref S32; required_contexts J1M+A256; task S32; transient_environment B |
| `api repos create-or-update-environment` | `repos/create-or-update-environment` | `create_or_update_environment` | `PUT /repos/{owner}/{repo}/environments/{environment_name}` | false | fixture_required_mutations | deployment_branch_policy J1M+C; environment_name S32; prevent_self_review B; reviewers J1M+A256; wait_timer I |
| `api repos create-or-update-file-contents` | `repos/create-or-update-file-contents` | `create_or_update_file` | `PUT /repos/{owner}/{repo}/contents/{path}` | false | fixture_required_mutations | author J1M+C; branch S32; committer J1M+C; content S32; message S32; path S32; sha S32 |
| `api repos create-webhook` | `repos/create-webhook` | `create_webhook` | `POST /repos/{owner}/{repo}/hooks` | false | fixture_required_mutations | active B; config J1M+O256; events J1M+A256; name S32 |
| `api repos delete-file` | `repos/delete-file` | `delete_file` | `DELETE /repos/{owner}/{repo}/contents/{path}` | false | fixture_required_mutations | author J1M+C; branch S32; committer J1M+C; message S32; path S32; sha S32 |
| `api repos merge` | `repos/merge` | `merge_branch` | `POST /repos/{owner}/{repo}/merges` | false | fixture_required_mutations | base S32; commit_message S32; head S32 |
| `api repos replace-all-topics` | `repos/replace-all-topics` | `replace_repo_topics` | `PUT /repos/{owner}/{repo}/topics` | false | fixture_required_mutations | names J1M+A256 |
| `api repos update-commit-comment` | `repos/update-commit-comment` | `update_commit_comment` | `PATCH /repos/{owner}/{repo}/comments/{comment_id}` | false | fixture_required_mutations | body S32; comment_id I |
| `api repos update-release-asset` | `repos/update-release-asset` | `update_release_asset` | `PATCH /repos/{owner}/{repo}/releases/assets/{asset_id}` | false | fixture_required_mutations | asset_id I; label S32; name S32; state S32 |
| `api repos update-webhook` | `repos/update-webhook` | `update_webhook` | `PATCH /repos/{owner}/{repo}/hooks/{hook_id}` | false | fixture_required_mutations | active B; add_events J1M+A256; config J1M+O256; events J1M+A256; hook_id I; remove_events J1M+A256 |
| `api secret-scanning update-alert` | `secret-scanning/update-alert` | `update_secret_scanning_alert` | `PATCH /repos/{owner}/{repo}/secret-scanning/alerts/{alert_number}` | false | fixture_required_mutations | alert_number I; assignee S32; resolution E; resolution_comment S32; state E; validity E |
