# TDD Ledger: Batch R1 vNext source-lock cutover

## Planned evidence

| Slice | Red characterization | Green contract | Refactor/verification |
| --- | --- | --- | --- |
| Runtime dependency | Reference connectors require embedded authoring/certification/admission material. | GitHub, GitLab, and Asana load and reach credential/approval preflight from execution JSON alone. | Audit embedded/runtime reads and run focused plus fleet tests. |
| Connector-local invalidity | Global ledgers or one bundle error can suppress the fleet. | Malformed required execution JSON rejects that connector without hiding healthy connectors. | Assert stable typed diagnostics and deterministic discovery. |
| Canonical rendering | Existing source-lock paths do not own a single canonical all-lane projection. | One vNext model renders byte-stable existing execution JSON through shared schema refs. | Check every rendered file and reject stale output. |
| Lane semantics | Retention/certification state can hide documented commands and source operations cannot express all lanes canonically. | Direct, binary, ETL, reverse ETL, sync, and explicit-empty lanes are surfaced without provider switches. | Run the same all-lane contract for every Batch R1 connector. |

## Actual evidence

### 2026-09-06 — CP11 Group 1 F-04 safe snapshot oracle

- **RED (exact original fixture):** [GROUP1-EVIDENCE.md](GROUP1-EVIDENCE.md) records the bounded `7e014d00` child matrix. The old snapshot blocks on regular→FIFO and mixes retained A metadata with B bytes for regular→symlink and directory A→B substitutions. The fixture report names source/overlay hashes, command, child outputs, kill/reap bounds, and its unrun F-01 overlay limit.
- **GREEN:** the active 17.659 s matrix records regular-file identity/bytes from one nonblocking/no-follow descriptor, verified directory descriptors for recursion, refusal of FIFO/symlink/directory B without B bytes, a nested positive, and semantic interrupted-authority/FIFO caller preservation. The exact tested diff identity and output witnesses are in `GROUP1-EVIDENCE.md`; this does not accept Group 2 F-01.

### 2026-09-06 — CP11 Group 1 F-08 bounded child ownership

- **RED (exact original fixture):** [GROUP1-EVIDENCE.md](GROUP1-EVIDENCE.md) records a process that follows the old `Start` then readiness-failure-before-cleanup order. Its exact sleeper PID remains live after the inner test exits. The outer fixture, not the original unarmed process, performs PID-specific `SIGKILL` then `ESRCH` absence observation; it does not directly `Wait` for or claim to reap the orphaned grandchild. Its `SIGKILL` is explicitly excluded from real-signal success evidence.
- **GREEN:** every active child becomes direct-`Wait`/cleanup-owned immediately after `Start`; FIFO and connector-directory descriptor observed lock-contention readiness are bounded; normal/cleanup waits are bounded; real OS `SIGINT` proves cancellation/no-success-output/state preservation/retry; and withheld FIFO readiness proves direct owned-child `Wait` cleanup separately from signal success. The exact matrix/output is in `GROUP1-EVIDENCE.md`.

### 2026-09-06 — CP11 Group 2 F-01 recorded-capture identity

- **RED (exact original fixture):** [GROUP2-EVIDENCE.md](GROUP2-EVIDENCE.md) records all three actual dependent opens against `7e014d00`: validation and candidate falsely accept B after A, and mutating reopen moves exact CURRENT bytes into B before its late identity error. The record preserves the two incomplete fixture setups and the final 5.256 s failure separately.
- **GREEN contract (in progress):** every recorded capture-dependent enumeration, candidate lookup, no-replace move, and sync must use a descriptor checked against `capture.Identity`; B must be rejected before mutation while A/B identities and public controls remain intact, and a valid capture must preserve normal recovery/reacquisition behavior.

### 2026-09-02 — G0 direct-parent delivery amendment

- Authority and route: #4325 comment `5500153864` mandates the issue lifecycle and one-writer, ordinary fast-forward delivery to `fm/cli-top100-declaration-batch-r1`; #4294 comment `5500165004` is the authoritative routing correction. The attempted PR-body replacement failed before mutation on deprecated `projectCards`, so no PR-body mutation is claimed.
- Immutable baseline: `git fetch origin fm/cli-top100-declaration-batch-r1`; `git rev-parse HEAD`; and `git rev-parse origin/fm/cli-top100-declaration-batch-r1` each returned `d260b725ce6f53403961d7af1ef48ea6651cdd66`. `git merge-base --is-ancestor HEAD origin/fm/cli-top100-declaration-batch-r1` succeeded, and `git merge-base origin/main origin/fm/cli-top100-declaration-batch-r1` returned `813f457a925f7ee3fe3bea101a43e445992c8552`. The fixed #4325 denominator is 4,341 primary retained source operations.
- Local-state inspection: `git ls-tree -r --name-only HEAD -- internal/connectors/certifications`, `git ls-files --stage -- internal/connectors/certifications`, and `git log --all --full-history -- internal/connectors/certifications/.fingerprint-salt` returned no tracked item. The opaque untracked `.fingerprint-salt` resides below the legacy certification tree deleted by `0b214b79eeb871238ce8454cd7b896e71e2746a7`; it is local-only, unowned by the parent branch, not ignored, unstaged, unread, and unmodified. Its disposition is preserve in place and exclude from G0/N1, not recover, delete, certify, or admit.
- Green condition: the G0 evidence-only change is reviewed with `git diff --check`, committed without the local residue, normally pushed to the existing parent branch, and its remote SHA is read back before N1 starts. No production or test behavior changes belong to G0.

### 2026-09-02 — #4423 N1 executable proof baseline (planned)

- Scope: restore truthful executable proof without a runtime behavior change. The fixed denominator/base above remains immutable.
- Red: preserve the stale `defs` compile/API failure, demonstrate that each named reference/Atlas proof selector cannot silently match zero tests, and add a reference-lock characterization that fails on the frozen architectural defect.
- Green: repair only test/API drift and proof selectors; make the reference-lock characterization read authoring inputs, render the complete expected set in memory, and fail a frozen architectural defect without source-lock re-pin, execution-manifest rewrite, generated connector change, runtime change, credential use, or provider I/O.
- Required evidence: focused `defs`, `connectorgen`, `engine`, and named proof commands with `-count=1 -timeout 20m`; explicit test-count assertion; `git diff --check`; and a fresh exact-SHA review range. The green commit carries `Refs #4423`.
- Actual RED: `go test -count=1 -timeout 20m ./internal/connectors/defs -run '^TestRuntimeEmbedContainsExecutionJSONOnly$'` exited 1 because `defs_test.go` referenced removed `engine.Bundle` fields `Docs`, `Surface`, and `Fixtures`. The two named Atlas selectors in `./cmd/connectorgen` and `./internal/connectors/engine` both exited 0 with `no tests to run`; the explicit `go test -list '^<name>$' | grep -Fx '<name>'` test-count assertion exited 1 for each. These are retained as the honest baseline, not a pushed failing commit.
- Actual GREEN: only `defs_test.go`, `vnext_lock_test.go`, and `vnext_execution_bundle_test.go` changed. The repaired defs test asserts current execution fields; the Foundation Atlas names now select real test functions; the reference characterization reads the three committed source locks, renders each complete execution set twice in memory, byte-compares it with committed execution JSON, and proves the closed-set comparator rejects an unrendered artifact. No runtime, source lock, generated execution JSON, credential, or provider-I/O path changed.
- Green commands: the three exact named proof commands passed with `-count=1 -timeout 20m`; each `go test -list '^<name>$' | awk` assertion counted exactly one selected top-level test. `go test -count=1 -timeout 20m ./cmd/connectorgen ./internal/connectors/defs` passed. `go test -count=1 -timeout 20m ./internal/connectors/engine` is not claimed green: `TestOperationRoutesFailClosedBeforeProviderIO` fails on its route-diagnostic expectation, while the N1 diff in that package is only the named-test rename. The named engine proof remains green; this unrelated broader-suite failure is a remaining gate for later work.
- Go-task skill record: the repository-required route was loaded before the N1 test work and re-read before review/push after Captain policy inbox `006`: `go-engineering` with `references/fundamentals.md`, `references/production.md`, and `references/agentic-etl.md`; `golang-how-to`; `golang-design-patterns`; `golang-structs-interfaces`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-testing`; `golang-context`; `golang-concurrency`; and `golang-cli`. No substitution was needed. N1 adds no production interface, error, credential, context, goroutine, or CLI behavior; the review applies those skills as test-only guardrails for concrete fields, checked errors, bounded local file reads, no secret/provider path, no mutable global state, and no concurrent work.

### 2026-09-01 — inherited ledger reconciliation

- The inherited `Red: pending` / `Green: pending` record contradicts the branch handoff's claimed reference-cohort green checkpoint. It is retained as history, not accepted as evidence for this continuation.
- Baseline: clean isolated continuation branch `fm/cli-batch1-vnext-cutover-r2` at `0b214b79eeb871238ce8454cd7b896e71e2746a7`, with that SHA proven reachable from `origin/fm/cli-top100-declaration-batch-r1`.
- Manual GSD fallback: the adapter resolves every required lifecycle command and the canonical contract check passes, but its generated commands cannot execute this pre-existing named phase because `.planning/ROADMAP.md` is absent. The inline artifacts in this directory carry the required lifecycle evidence.

### 2026-09-01 — cleanup slice A: native fixture bypass

- Red target: a native connector supplied only `config.mode=fixture` must not report a successful check or emit canned records; it must continue through normal credential/config validation before provider I/O.
- Red command: `go test -timeout 20m ./internal/connectors/native/alpha-vantage -run '^TestFixtureModeNoLongerBypassesCredentialBoundary$'`.
- Green contract: the same invocation returns the connector's missing-credential error with no provider request, and the production implementation contains no fixture-mode branch.
- Follow-on residual proof: scan production native/hook/engine/connectorgen code, definitions, generated docs/skills, and website sources for a fixture, importer, certification, retention, compatibility, feature-flag, or second-executor execution/admission path. Any retained mention must be connector-local provider provenance only and is recorded by path and reason in `VERIFICATION.md`.

### 2026-09-01 — cohort migration template

- Red: before each named connector migration, `lock-render <connector> --check` fails against the newly written lock or the connector lacks a usable declaration/credential-boundary witness.
- Green: `lock-render <connector>` produces byte-identical execution JSON; all seven lanes are explicit; malformed execution JSON rejects locally; and an isolated, credential-free command reaches the ordinary missing-credential or approval boundary without provider I/O.
- Connector sequence: Bitbucket, CircleCI, Docker Hub, Jira, Notion, Sentry, Stripe, Vercel. Each receives its own red/green entry, review record, commit SHA, and normal push evidence.

### 2026-09-02 — S1A accepted correction map

- Independent exact-range N1 review is **BLOCK** at `c5bf5c5d544e85dcca5eac3ebed45ba78ad7fb33`; its narrow N1 proofs are real, but it is not a green S1A unlock. The accepted D1/D2 decisions require API connectors to retain rendered engine executors, forbid API native overwrite/delegating-hook and public fixture/origin routes, and preserve PostgreSQL, MySQL, and DynamoDB native database registrations and behavior unchanged.
- RED gates before production deletion: `TestFixtureModeNoLongerBypassesCredentialBoundary` is listed exactly once and fails because fixture mode currently bypasses credentials; `TestEveryImplementedCommandHasProductionRuntimeSurface` fails rather than skipping an erased command surface; `TestRegistryRejectsDuplicateConnectorNames` fails on silent overwrite; and `TestAlphaVantageRejectsUntrustedBaseURLBeforeSecretSend` fails when an untrusted origin receives a credential-bearing request.
- GREEN contract: the four API gates pass after API bundles retain their rendered executors, every implemented command has a runtime surface, duplicate executor identity is rejected, fixture/origin selectors are removed, and the rejected origin receives no request. PostgreSQL, MySQL, and DynamoDB remain native. The Atlas records open R1-05/R1-13 gaps until the approved D2 publication transaction and strict source-lock decoding are implemented.
- Actual RED: every required selector counted exactly once with `go test -list ... | sed ... | wc -l`, returning `1` for `TestFixtureModeNoLongerBypassesCredentialBoundary`, `TestAlphaVantageRejectsUntrustedBaseURLBeforeSecretSend`, `TestEveryImplementedCommandHasProductionRuntimeSurface`, and `TestRegistryRejectsDuplicateConnectorNames`.
- Actual RED: `go test -count=1 -timeout 20m ./internal/connectors/native/alpha-vantage -run '^(TestFixtureModeNoLongerBypassesCredentialBoundary|TestAlphaVantageRejectsUntrustedBaseURLBeforeSecretSend)$'` exited `1`: fixture `Check` returned `<nil>` instead of the missing-credential error, and the untrusted test origin received one request.
- Actual RED: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^(TestEveryImplementedCommandHasProductionRuntimeSurface|TestRegistryRejectsDuplicateConnectorNames)$'` exited `1`: `google-calendar` had no production command surface because `nativeset.definitionConnector` overwrote the rendered executor, and `Registry.Register` silently accepted a duplicate name.
- Actual GREEN: `go test -count=1 -timeout 20m ./internal/connectors/native/alpha-vantage -run '^(TestFixtureModeNoLongerBypassesCredentialBoundary|TestAlphaVantageRejectsUntrustedBaseURLBeforeSecretSend)$'` passed after removing the fixture branch and rejecting `base_url` overrides before requester construction.
- Actual GREEN: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^(TestEveryImplementedCommandHasProductionRuntimeSurface|TestRegistryRejectsDuplicateConnectorNames|TestProtectedNativeDatabasesRemainRegistered)$'` passed after removing API native registration, retaining the three protected native database registrations, and rejecting duplicate registry names.
- Protected-database RED/GREEN: `TestProtectedNativeDatabasesRemainRegistered` first failed because DynamoDB was replaced by `*engine.Connector`; it then passed after the factory partition restored DynamoDB, MySQL, and PostgreSQL as native registry entries. This test is the non-regression guard for Captain's API-only S1A scope.
- Amazon SQS bounded source-lock RED: `go run ./cmd/connectorgen lock-render amazon-sqs --check` exited `1` because `source.lock.json` was absent. This replaces neither the earlier commandrunner RED nor the D2 stale-bundle gap.
- Amazon SQS bounded source-lock GREEN: schema-4 `source.lock.json` records provider citations for DeleteQueue and PurgeQueue, declares every lane unsupported, and renders only two `docs_only` commands with `unsupported_with_provider_evidence`. `lock-render amazon-sqs` and `--check` passed; `TestAmazonSQSDispositionBundleLoads` passed; implemented-command preflight passed; and the built CLI displayed both commands then rejected `queue delete` at the policy boundary before credential resolution or provider I/O.
- The Amazon bundle inspection still exposes legacy streams and write actions left by per-file rendering. That is the already-recorded R1-05 open-set publication gap, not a claim that those stale artifacts are source-lock-authoritative or executable through the two disposition commands.

### 2026-09-02 — full API generic-engine migration

- Scope: 28 deleted API hook/native executor families listed in `PLAN.md`; Amazon SQS remains the captain-approved non-generic disposition, and DynamoDB/MySQL/PostgreSQL remain behaviorally unchanged native databases.
- RED per family: show the missing generic-engine check/read/ETL route, response mapping, command preflight, or production-registry witness with a deterministic local fake server; count each focused selector exactly once. A legacy package, fixture selector, caller-provided origin, or unbound hook is never a substitute for RED evidence.
- GREEN per family: one schema-4 source lock renders byte-identical execution JSON; all seven lanes are explicit; the production registry exposes the declared engine connector; check/read/ETL send the provider-declared request shape only to the local fake server and reach ordinary credential/approval boundaries without provider I/O; no deleted hook/native path or caller origin remains.
- Shared GREEN: a global Atlas selector resolver validates every affected owner, guarantee, selector, and proof record; generated CLI docs, skills, catalog, and website output are current; the completed migration receives a fresh exact-SHA independent OMP review before any publication.
- Alpha production RED: `TestProductionAlphaVantageRejectsFixtureAndCallerOriginBeforeIO` is selected from `bundleregistry.New()` exactly once and fails before any migration because the rendered legacy bundle sends ten retrying requests for `mode=fixture` or caller `base_url` rather than rejecting either input locally. Its transport spy records count only; no provider request or secret value is observed.
- Alpha production GREEN: a schema-4 lock renders five fixed-provider ETL streams with API-key query auth, keyed-object response extraction, literal dotted provider-field mapping, and strict unknown-config refusal. `TestProductionAlphaVantageRejectsFixtureAndCallerOriginBeforeIO`, `TestProductionAlphaVantageGenericCheckAndRead`, and `TestResolveRecordPathValueUsesLiteralDottedKey` pass; the alpha native executor and its direct-only test are deleted after those production-registry proofs.
- CloudTrail SigV4 RED/GREEN: the former native-only signer had no Atlas owner or generic auth mode. Captain authorized the constrained fixed CloudTrail/us-east-1 extension; `TestCloudTrailSigV4AuthenticatorDeterministicVector`, `TestReadCursorPaginationMovesTokenIntoDeclaredBody`, and `TestProductionCloudTrailGenericCheckAndRead` pass with local-only fixed-signature, POST-body NextToken, and production-registry check/read witnesses. The schema-4 CloudTrail lock renders and checks byte-current with no caller origin or fixture selector.
- Babelforce production RED: `TestProductionBabelforceGenericCheckAndRead` fails through `bundleregistry.New()` because the rendered bundle still resolves its caller-controlled legacy placeholder origin before it can prove the fixed dual-header route or records response mapping.
- Babelforce production GREEN: a schema-4 lock renders five fixed declared REST streams with strict regional host enum, dual-header auth, page-number ETL pagination, and passthrough provider records. `TestProductionBabelforceGenericCheckAndRead` and `lock-render babelforce --check` pass through the production registry and local transport spy.
- Basecamp production GREEN: a schema-4 lock renders fixed account-bound Basecamp routes, declared bearer/refresh authentication, Link-header pagination, and three passthrough ETL streams. `TestProductionBasecampGenericCheckAndRead` and `lock-render basecamp --check` pass through the production registry and local transport spy.
- Bunny production GREEN: a schema-4 lock renders subdomain-pattern-bounded GraphQL host selection, bearer auth, five named GraphQL cursor connections, and provider node/pageInfo mappings. `TestProductionBunnyGenericCheckAndRead` and `lock-render bunny-inc --check` pass through the production registry and local transport spy.
- Canny production GREEN: a schema-4 lock renders fixed Canny REST form requests with form-body API-key containment, offset ETL pagination, and five declared collection response mappings. `TestProductionCannyGenericCheckAndRead` and `lock-render canny --check` pass through the production registry and local transport spy.
- Copper production GREEN: a schema-4 lock renders fixed Copper search routes, declared three-header auth, JSON check bodies, body-carried page-number ETL pagination, and five search response mappings. `TestProductionCopperGenericCheckAndRead` and `lock-render copper --check` pass through the production registry and local transport spy.
- Dixa production GREEN: a schema-4 lock renders fixed export routes, bearer auth, declared export bounds, and four conversation projections. `TestProductionDixaGenericCheckAndRead` and `lock-render dixa --check` pass through the production registry and local transport spy.
- FastBill production GREEN: a schema-4 lock renders fixed SERVICE-envelope API requests, Basic auth, body-carried offset pagination, and five response-envelope streams. `TestProductionFastbillGenericCheckAndRead` and `lock-render fastbill --check` pass through the production registry and local transport spy.
- Feishu planned RED: a returned production-registry connector must reject caller-origin and fixture configuration before any I/O, then perform only the declared tenant-token JSON exchange, bounded Bitable tables check, and Bitable records route with the exchanged bearer token. The existing bundle instead allows `base_url`/`mode` and invokes `__legacy_hook`; the focused proof establishes this mismatch without provider I/O.
- Feishu actual RED: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionFeishuGenericCheckAndRead$'` exited `1`; its local transport spy received `example.invalid`, proving the execution bundle still used its caller-controlled legacy placeholder rather than a declared Feishu host.
- Feishu GREEN: schema-4 `source.lock.json` declares all seven lanes and the two-host enum; it renders an execution-only bundle with fixed Bitable paths, cursor pagination, and a source-selected declared-route tenant auth extension. The generic and fixture/origin production proofs each counted exactly once; both passed with `-count=1 -timeout 20m`, and `go run ./cmd/connectorgen lock-render feishu --check` passed. No provider endpoint or credential store was used.
- FreeAgent actual RED: `TestProductionFreeAgentGenericCheckAndRead` first failed through the returned production registry because the former bundle selected `example.invalid`; after the auth descriptor field was introduced, `TestOAuth2RefreshTokenBasicClientAuthentication` also failed because the refresh encoder omitted HTTP Basic client authentication.
- FreeAgent GREEN: schema-4 `source.lock.json` declares all seven lanes, fixed FreeAgent v2 routes, page/per_page pagination, and the `client_authentication: basic` refresh-token binding. The generic encoder sends client credentials only in HTTP Basic and rejects an unsupported authentication form before any token request. The two FreeAgent production selectors each counted once and passed with `-count=1 -timeout 20m`; `go run ./cmd/connectorgen lock-render free-agent-connector --check` and the focused connsdk tests passed without provider I/O.
- Freightview actual RED: `TestProductionFreightviewGenericCheckAndRead` failed through the returned production registry because the legacy bundle resolved the caller-controlled `example.invalid` placeholder instead of the declared v2.0 provider host.
- Freightview GREEN: schema-4 `source.lock.json` declares all seven lanes, fixed client-credentials authentication, bounded continuation-token pagination, and declared shipment-to-quote/tracking fan-out. The generic production proof exercises the token exchange, check, root read, and stamped subresource fan-out; it and the rejected-origin/fixture proof each counted once and passed with `-count=1 -timeout 20m`. `go run ./cmd/connectorgen lock-render freightview --check` passed without provider I/O.
- Google Analytics Data actual RED: `TestProductionGoogleAnalyticsDataGenericCheckAndRead` failed because the former bundle sent its check to `https://example.invalid/__legacy_hook/check`. The focused response-header tests then failed because rows retained positional arrays and accepted unknown headers, mismatched counts, and missing values.
- Google Analytics Data GREEN: Captain-approved `runtime.response-header-projection.v1` resolves only declared response headers to equal-position row values before any page emits. The schema-4 GA4 lock declares all seven lanes, five fixed `runReport` streams, declared bearer auth, body-offset pagination, and the exact dimension/metric header sets. Focused foundation and production-registry check/read/origin proofs pass with local-only transport; `go run ./cmd/connectorgen lock-render google-analytics-data-api --check` passes.
- Google Classroom RED/GREEN: the production registry initially used `example.invalid` and `__legacy_hook`. A schema-4 lock now selects fixed Google OAuth refresh-token authentication, cursor paging, and declared course-root fan-out for teachers, students, course work, and announcements. `TestProductionGoogleClassroomGenericCheckAndRead` passed through the production registry with a local transport after the RED, and `lock-render google-classroom --check` is byte-current.

- Google PageSpeed Insights actual RED: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionGooglePageSpeedGenericCheckAndRead$'` exited 1 because the production registry routed `Check` through `https://example.invalid/__legacy_hook/check`, rather than the declared provider route. The local transport rejected the incorrect host before any provider I/O.
- Google PageSpeed Insights GREEN: a schema-4 `source.lock.json` now declares the fixed PageSpeed origin, optional declared API-key query auth, one ETL stream, fixed category repetitions, the bounded HTTPS URL × mobile-or-desktop strategy fan-out, and a hard twenty-request budget. `go run ./cmd/connectorgen lock-render google-pagespeed-insights`, its `--check` variant, and the focused production-registry proof passed. The registry proof observed one bounded fixed-origin check plus four deterministic local URL-and-strategy reads, emitted four flattened records, and observed no caller-selected origin, hook, or fixture route.
- PageSpeed fan-out foundation GREEN: `TestReadCartesianConfigFanOutRunsDeclaredProductDeterministically` proves source-order URL × strategy requests with repeated declared categories and record stamps; `TestReadCartesianConfigFanOutRejectsBudgetBeforeProviderIO` proves an over-budget product makes no request. Both passed with `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestReadCartesianConfigFanOutRunsDeclaredProductDeterministically|TestReadCartesianConfigFanOutRejectsBudgetBeforeProviderIO)$'`.
- Atlas global-selector RED/GREEN: the first `TestFoundationAtlasSelectorsResolve` run found stale owner/proof references and missing proof files in existing Atlas entries. The resolver now parses every declared owner symbol and proof-test selector, rejects empty guarantees/selectors, and the obsolete records were corrected to current executable proofs. `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestFoundationAtlasSelectorsResolve$'` passed.

- Less Annoying CRM actual RED: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionLessAnnoyingCRMGenericCheckAndRead$'` exited 1 because the production registry sent the check to `https://example.invalid/__legacy_hook/check` instead of the fixed v2 RPC origin.
- Less Annoying CRM GREEN: the schema-4 lock declares fixed v2 POST `GetUsers`, `GetContacts`, `GetTasks`, `GetNotes`, and `GetEvents` operations, raw API-key header authentication, page-number request parameters, and all seven lanes. `lock-render less-annoying-crm --check` and the focused production-registry proof passed with a local transport: it observed the bounded `GetUsers` check plus the paged `GetContacts` read, raw record projection, and no caller-origin, fixture, hook, credential-store, or provider-I/O path.

- Lokalise actual RED: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionLokaliseGenericCheckAndRead$'` exited 1 because the legacy bundle selected `example.invalid/__legacy_hook/check`.
- Lokalise GREEN: a schema-4 lock declares the fixed API v2 project routes, `X-Api-Token` authentication, page/limit pagination, five ETL streams, and all seven lane states. Rendering, `lock-render lokalise --check`, and the production registry proof passed with a local fixed-origin transport that observed the bounded languages check and paged language read without a caller origin, fixture, hook, credential-store, or provider-I/O route.

- Mendeley actual RED: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionMendeleyGenericCheckAndRead$'` exited 1 because the production registry still selected `example.invalid`; the local transport received no fixed Mendeley origin.
- Mendeley Foundation Atlas RED/GREEN `[key=mendeley-per-stream-accept]`: Captain approved `runtime.per-stream-headers.v1` for source-declared static vendor `Accept` media types only. The new schema-4 lock declares each Mendeley resource’s exact media type, OAuth refresh-token auth, Link-header pagination, and all seven lane states. Focused loader rejection/merge tests, the four-request token/check/token/read production witness, `lock-render mendeley --check`, and no-hook execution passed; dynamic, non-vendor, and protected header routes remain unavailable before provider I/O.

- Rootly actual RED: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionRootlyGenericCheckAndRead$'` exited 1 because the legacy bundle routed the check through `example.invalid/__legacy_hook/check`.
- Rootly GREEN: a schema-4 lock declares fixed JSON:API incidents, services, and users routes; bearer authentication; bounded next-link pagination; JSON:API attribute projection; and all seven lane states. Rendering, `lock-render rootly --check`, and the local production-registry witness passed without a caller origin, fixture, hook, credential-store, or provider-I/O route.

- My Hours actual RED: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionMyHoursGenericCheckAndRead$'` exited 1 because the legacy bundle selected `example.invalid`; the local transport received no fixed provider host.
- My Hours password-token foundation GREEN `[key=my-hours-password-token-auth]`: `runtime.declared-password-token.v1` binds only a fixed HTTPS JSON `{email,password}` login route to a top-level `accessToken` bearer. Focused secret-containment/static-route tests and the Atlas selector resolver passed; no source lock or runtime uses a generic password grant, caller token URL, refresh, replay, custom auth hook, or native reader.
- My Hours date-window foundation GREEN `[key=my-hours-time-window-fanout]`: `runtime.date-window-fanout.v1` schedules only source-declared contiguous UTC `DateFrom`/`DateTo` windows. Its focused tests prove deterministic no-gap/no-overlap scheduling and pre-I/O cap refusal. The schema-4 My Hours lock now declares all five ETL streams, the fixed token exchange, static API-version header, and a 600-window time_logs cap. Rendering, `lock-render my-hours --check`, focused production-registry check/clients/time_logs witness, Atlas selector resolver, and definition validation all passed.

- SafetyCulture actual RED/GREEN: `TestProductionSafetyCultureGenericCheckAndRead` first rejected the legacy `example.invalid/__legacy_hook/check` route, then passed after a schema-4 lock declared fixed audits, templates, and users routes, bearer authentication, bounded next-link pagination, and all seven lane states. `lock-render safetyculture --check` passed with the local production-registry audits witness.

- Pocket actual RED/GREEN: `TestProductionPocketGenericCheckAndRead` first exposed the legacy route and then caught an invalid fallback projection before the schema-4 coalesce form was corrected. The final fixed-origin POST witness proves source-declared request credentials, bounded count/offset, keyed-object item ID stamping, and title/URL fallback projection. `lock-render pocket --check` passed.

- Mode GREEN: schema-4 lock renders fixed workspace-scoped HAL+JSON streams with static `Accept`, Basic authentication, and bounded next-link pagination. The production-registry check/spaces witness and `lock-render mode --check` passed through the local fixed-origin transport.

- Mercado Ads GREEN: schema-4 lock preserves six advertiser/metric schemas, fixed product bindings, OAuth refresh-token authentication, API-version header, declared campaign bindings, and date-range inputs. The local production registry proved token/check/token/brand-advertiser request shape; `lock-render mercado-ads --check` and full definition validation passed.

- Ashby RED/GREEN: the initial production check used the retained `/apiKey.info` route, failing the fixed `POST /candidate.list` witness. The source lock check was corrected to the declared candidates path; the production registry then proved bounded check/read behavior, Basic authentication, versioned media header, one candidate emission, and preservation of all 178 command bindings. `lock-render ashby --check` passed.

### 2026-09-02 — S1A 28-family reconciliation

- The denominator is exactly the 28 API families named in `PLAN.md`: Apify Dataset, Ashby, AWS CloudTrail, Babelforce, Basecamp, Bunny Inc, Canny, Copper, Dixa, FastBill, Feishu, Free Agent, Freightview, Google Analytics Data, Google Classroom, Google PageSpeed, Less Annoying CRM, Lokalise, Mendeley, Mercado Ads, Metabase, Mode, My Hours, Pocket, PrestaShop, Rootly, SafetyCulture, and Yahoo Finance.
- Counted GREEN only after execution-only production-registry check/read witnesses and deterministic `lock-render --check`: Apify Dataset, Ashby, AWS CloudTrail, Babelforce, Basecamp, Bunny Inc, Canny, Copper, Dixa, FastBill, Feishu, Free Agent, Freightview, Google Analytics Data, Google Classroom, Google PageSpeed, Less Annoying CRM, Lokalise, Mendeley, Mercado Ads, Metabase, Mode, My Hours, Pocket, PrestaShop, Rootly, SafetyCulture, and Yahoo Finance — **28/28**.
- Alpha Vantage and Amazon SQS are explicitly outside this 28-family denominator; Amazon remains the captain-approved all-lanes-unsupported disposition. DynamoDB, MySQL, and PostgreSQL remain protected native database registrations.
- Historical gap statements for Metabase, PrestaShop, and Yahoo Finance above describe the pre-approval state only. Their matching approved foundations and executable RED/GREEN evidence are recorded below; no hook, native executor, raw response, caller origin, capability demotion, provider credential, or provider request was used.

### 2026-09-03 — tenant-origin and declared-session completion

- Tenant-origin RED: `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestTenantOriginOverridesDefaultOperationRoute$'` exited 1 because `resolveOperationRoute` returned `https://unused.invalid` instead of the configured loopback tenant `/api` origin. This was the read-path defect left after the initial runtime-only tenant resolver work.
- Tenant-origin GREEN: `resolveOperationRoute` now uses the same validated `TenantOriginSpec` before building the stream runtime. `TestTenantOriginOverridesDefaultOperationRoute`, `TestResolveTenantOriginPermitsOnlyDeclaredLoopbackHTTP`, `TestProductionPrestaShopTenantOriginRejectsLegacyConfigBeforeIO`, and `TestProductionMetabaseDeclaredSessionCheckAndRead` pass with local transport only. `runtime.tenant-origin.v1` records the owner, strict HTTPS/loopback boundary, route parity, and proof selectors in the Foundation Atlas.
- Metabase RED: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionMetabaseDeclaredSessionCheckAndRead$'` exited 1 because the legacy bundle routed its check to `https://example.invalid/__legacy_hook/check`. No request reached a provider.
- Metabase GREEN: the schema-4 `metabase/source.lock.json` declares `instance_api_url` through the tenant-origin selector, fixed `/api/session` authentication, source-declared cards/dashboards/collections/databases/users routes, and all seven lanes. The generic `declared_session` mode sends only `{username,password}` to its declared session exchange, extracts only top-level `id`, disables automatic exchange retries, caches only per runtime, and applies `X-Metabase-Session` to data requests. The production witness proves both password-session exchange and an existing `session_token` header route, with no password on data requests. `runtime.declared-session.v1` records its closed auth contract and proofs.

### 2026-09-03 — Yahoo Finance array-zip completion

- Array-zip RED: `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestArrayZipProjection'` initially failed to compile because the bounded declaration type and executor did not exist. The Yahoo production RED then exited 1 when its legacy check sent `GET https://query1.finance.yahoo.com/__legacy_hook/check`.
- Array-zip GREEN: `runtime.array-zip-projection.v1` accepts only source-declared scalar paths and equal-length array paths; missing/non-array/mismatched inputs fail before row emission. `TestArrayZipProjectionCombinesDeclaredArrays` and `TestArrayZipProjectionRejectsMismatchedArrays` pass.
- Yahoo Finance GREEN: the schema-4 `yahoo-finance-price/source.lock.json` fixes the chart origin and `/v8/finance/chart/{{ config.symbol }}` route, declares all seven lanes, and zips `chart.result[].meta` plus aligned timestamp, quote, and adjusted-close arrays into OHLCV rows. `TestProductionYahooFinancePriceArrayZipCheckAndRead` passes through the production registry and observes exactly one local check plus one local read with two emitted OHLCV records. `lock-render yahoo-finance-price --check` and full definition validation pass without credentials or provider I/O.

### 2026-09-03 — final validation remediation

- Agent inventory RED: `go test -count=1 -timeout 20m ./internal/agentcontract -run '^TestCheckProjectionsIgnoresNestedCacheClaudeAgents$'` exited 1 because recursive project-root traversal treated `.cache/preserved-baseline/.claude/agents/pm-delivery-worker.md` as a second project definition.
- Agent inventory GREEN: inventory now walks only `.claude/agents` beneath the project root; it still rejects duplicate and unexpected definitions under that real root directory, while the nested-cache regression passes. `go run ./cmd/agentcontractgen check` now passes in this preserved worktree without altering `.cache`.
- Claude guidance GREEN: `CLAUDE.md` was reconciled to the exact canonical two-line `@AGENTS.md` pointer. `/Users/karthiksivadas/karthik-agent-workspace/bin/fm-ensure-agents-md.sh .` now reports the existing AGENTS.md unchanged with its canonical pointer; AGENTS.md itself was not edited.

### 2026-09-03 — exact-SHA review remediation

- R1 Ashby RED: `TestProductionAshbyRejectsCallerOriginAndFixtureBeforeIO` failed because `base_url` and `mode` remained accepted configuration fields; caller input could reach the generic requester. GREEN removes both fields, fixes the source-lock origin to `https://api.ashbyhq.com`, declares strict unknown-config rejection, and proves hostile origin/fixture inputs make zero request before credential or transport work.
- R2 PrestaShop RED: `TestProductionPrestaShopGenericCheckAndRead` failed because the rendered tenant bundle still selected `auth:none` and legacy stream hooks. GREEN declares Basic `access_key` authentication, real `/customers`, `/orders`, `/products`, `/addresses`, and `/carts` resource routes, JSON/full-resource query shape, offset/limit windows, nested collection extraction, and `date_upd` client filtering from `start_date`. The local production witness proves one Basic-authenticated check and one read with no hook route.
- R3 tenant-origin RED: `TestResolveTenantOriginRejectsUnexpectedPathPrefix` failed because a configured `/admin` prefix became `/admin/api`. GREEN permits only empty/root or the exact normalized declared append path; encoded paths, arbitrary prefixes, and suffixes fail before I/O.
- R4 Yahoo RED: `TestProductionYahooFinancePriceRejectsChartErrorBeforeEmit` initially failed to compile because no typed declared response error existed. GREEN adds the source-selected `response_error` object, checks it before record extraction, returns `*DeclaredResponseError` for `chart.error.description`, and proves zero OHLCV records emit.
- R5 generated surfaces GREEN: `./pm docs generate --dir docs/cli`, `./pm skills generate --dir docs/skills --json`, and `npm --prefix website run gen:website-data` completed. `npm --prefix website run test:scripts` passed 34/34, docs validation passed, and Ashby/PrestaShop manuals, skills, catalog, and complete website connector data are current.

### 2026-09-03 — second exact-SHA review remediation

- RR1 Ashby RED: the two-page local witness first failed because every rendered stream omitted a JSON body. GREEN declares a fixed POST JSON `limit`, cursor-in-body `nextCursor`/`moreDataAvailable` pagination, success=false envelope rejection, and source-backed body bindings for all 71 ETL stream commands. The witness proves 100 records plus a second page, no query cursor, no repeated cursor send, and zero records after a failure envelope.
- RR2 PrestaShop RED: the former split `offset`/`limit` path could not send provider-conforming `limit=offset,count`. `runtime.offset-count-pagination.v1` now emits exactly `0,100` then `100,100`, stopping on the short second page. PrestaShop uses resource-keyed collection paths and `start_date` client filtering; its two-page witness proves the pre-start record does not emit.
- RR3 shared configuration RED: the hostile Ashby direct operation accepted removed `base_url` configuration and reached the local transport. Every network-capable `engine.Connector` method now calls the shared configuration validator before auth admission or requester construction; the direct-operation witness observes zero request.
- RR4 docs RED: all 21 migrated API `docs.md` inputs plus Ashby retained obsolete connector-managed ownership and/or stale request facts. `TestMigratedAPIDocsDescribeRenderedStreams` now compares every rendered check, stream route, records path, body, pagination, and response-error fact against these docs, and rejects the stale owner language and Ashby `base_url`/`mode`.
- Declared-session response regressions: malformed JSON and a missing session `id` both fail after exactly one exchange, set no session header, and do not replay.

### 2026-09-03 — final exact-SHA review remediation

- F1 Ashby RED/GREEN: the blanket body/cursor shape omitted operation-specific fields and typed values. All 71 streams now declare JSON bodies from their own command facts; list endpoints alone declare cursor bodies, info/synchronous endpoints never invent cursors, typed optional values are coerced before encoding, and success=false envelopes stop before emit. The two-page test proves 100+1 records, no query cursor, and no failure-envelope records.
- F2 PrestaShop RED/GREEN: client filtering could lose records against capped pages. Streams now send `filter[date_upd]` plus `sort=[date_upd_ASC]` and the provider's `limit=offset,count` windows. The two-page witness proves exact filter/sort/combined queries and 101 server-filtered records.
- F3 CloudTrail RED/GREEN: each stream now carries declared `StartTime` from `start_date`; a cap with a known cursor raises `ReadBudgetStoppedError` with an opaque continuation rather than returning success. The production witness proves StartTime on both body-cursor pages.
- F4/F5/F6: password-token redirects stop at the 307/308 response before any destination request; date-window budget counting now iterates calendar windows and rejects an extreme range before I/O; offset-count declarations fail at both load and paginator boundaries when parameter or size is malformed.
- F7/F8: Ashby fixture-only metadata and generated claims are removed under semantic docs parity. The Atlas now carries a fixed CloudTrail SigV4 contract, response-envelope selectors, offset-count ownership, real owner files, and proof membership checks; the deleted nativeset owner path is repaired.

### 2026-09-03 — final architecture authority remediation

- F1: all 71 Ashby streams now derive their closed JSON bodies from operation command facts; list endpoints alone carry typed body cursors, non-list operations never receive invented cursors, success=false is typed and pre-emission, and repeated cursors fail before a third send.
- F2/F3: PrestaShop emits its documented server filter/sort and combined page values with no client-scan truncation. CloudTrail writes StartTime to every stream body and returns an opaque budget continuation at a known cap rather than success.
- F4/F5/F6: token redirects are non-following; date-window limits count calendar windows without duration saturation; malformed offset-count specs are rejected at both load and paginator boundaries.
- F7/F8: Ashby fixture-only wording is absent from source and all generated surfaces under semantic parity. The Atlas proves literal owner files, declared owner membership, repaired native owner paths, CloudTrail SigV4 selection, declared-session malformed-response coverage, response envelope, and offset-count foundations.

### 2026-09-03 — terminal review correction B3

- Foundation Atlas discovery: **constrained extension** of the existing generic direct-execution pagination contract. Canny needs a declared `POST` form-body `skip`/`limit` window and an authoritative `hasMore` stop; no connector branch, hook, caller query channel, native executor, or new foundation is introduced. The shared owner is `internal/connectors/engine/{bundle.go,paginate.go,read.go}`.
- RED: `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestNewPaginatorOffsetLimitHonorsDeclaredStopPath$'` requested offset `4` after a full second page declared `hasMore:false`; `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionCannyGenericCheckAndRead$'` observed `limit=100&skip=0` in the read URL instead of the declared form.
- GREEN: Canny’s schema-4 lock selects body-bound `skip`/`limit` and `hasMore`; the generic offset paginator stops on a declared falsy stop path even for a full page. The focused engine and production-registry fake-transport tests pass, read exactly two 100-record pages with form-only pagination, and `lock-render canny --check` is deterministic.

### 2026-09-03 — terminal review correction B4

- Foundation Atlas discovery: **constrained extension** of the existing engine pagination contract. The engine-owned opaque durable continuation must resume the exact generated provider cursor, offset window, or same-origin next URL; it does not admit caller pagination state, replay an acknowledged page, or add a connector-specific reader.
- RED: `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestReadContinuationResumesExactProviderCursor$'` requested the empty cursor twice before requesting `page-two`, proving that the prior `skipped_pages` continuation replayed acknowledged provider state.
- GREEN: `engine_pagination_v2` binds the definition digest and persists only the next engine-generated request state. Resumable paginators validate declared cursor/window/query state and reapply URL origin guards. `TestReadContinuationResumesExactProviderCursor`, `TestReadContinuationResumesExactProviderOffset`, `TestReadContinuationResumesExactProviderURL`, and the durable transport continuation proof pass with only the initial then exact continuation request observed.

### 2026-09-03 — terminal review correction B5

- Foundation Atlas discovery: **constrained extension** of `authoring.source-lock-vnext.v1`. Immutable schema-4 authoring requires one complete JSON document with no duplicate object members before canonicalization or any generated-file replacement; this is authoring admission only and never a runtime reader.
- RED: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestDecodeVNextSourceLockRejectsTrailingAndDuplicateJSON$'` did not compile because the strict source-lock decoder was absent.
- GREEN: `decodeVNextSourceLock` walks every object before typed decode, rejects trailing values and duplicate root or nested keys, then applies unknown-field rejection. `TestRunLockRenderRejectsNonCanonicalSourceBeforeWriting` proves a rejected source leaves an existing generated artifact byte-for-byte untouched; the decoder tests and `lock-render canny --check` pass.

### 2026-09-03 — terminal review correction B6

- RED: the strengthened `TestS1ASourceParentContractMatrixMatchesExecution` failed at `ashby.stream.hiring_team_role_list`: its matrix omitted the source-declared `namesOnly:true` body fact. The strengthened docs parity test also found every base-pagination fact and Canny’s form body absent or mislabeled.
- GREEN: the matrix now rejects unknown/trailing/duplicate JSON, has the exact 28-family/operation/stream bidirectional denominator, and compares lock and execution auth plus every recorded stream method, path, body, body type, required field list, records, incremental, and response-error fact. Ashby records and documents `namesOnly=true`. All 28 docs now state their inherited default pagination, while Canny documents typed form-body `skip`/`limit` windows and `hasMore=false` termination. The matrix/docs selector passes.

### 2026-09-03 — terminal corrections B3–B6 verification

- Required skill route: `connector-lane-build-order` and `go-engineering` were loaded. The repository-named `golang-*` skills and `firstmate-exhaustive-review` remain unavailable in this session; the recorded `go-engineering` substitute and the existing inline/manual-GSD fallback remain authoritative.
- Red: the B3–B6 RED commands and their observable failures are recorded immediately above, before their production changes. They were not re-created by mutating the preserved green worktree.
- Green: `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^(TestNewPaginatorOffsetLimitHonorsDeclaredStopPath|TestReadContinuationResumesExactProviderCursor|TestReadContinuationResumesExactProviderOffset|TestReadContinuationResumesExactProviderURL)$'`, `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestProductionCannyGenericCheckAndRead$'`, and `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestDecodeVNextSourceLockRejectsTrailingAndDuplicateJSON|TestRunLockRenderRejectsNonCanonicalSourceBeforeWriting|TestS1ASourceParentContractMatrixMatchesExecution|TestMigratedAPIDocsDescribeRenderedStreams)$'` all passed.
- Broader Green: complete `cmd/connectorgen`, `internal/connectors/defs`, `internal/connectors/bundleregistry`, `internal/connectors/commandrunner`, and `internal/connectors/engine` suites passed with `-count=1 -timeout 20m`.
- Baseline only: complete `internal/app` and `internal/cli` suites retain the already-recorded typed-destination approval/I-O and polling-help/source-origin failures; CLI also logs its expected local Redis connection refusals. These paths are outside B3–B6 and were not changed.
- Render and documentation checks passed: `go build ./cmd/pm`; `lock-render --check` for GitHub, GitLab, Asana, and Canny; full `connectorgen validate internal/connectors/defs` (553 connectors, 0 findings); and `./pm docs validate --connectors-dir docs/connectors`.
- Audit: `git diff --check` passed. The complete 57-path tracked diff, including full generated website JSON artifacts, was scanned programmatically for private keys, GitHub tokens, OpenAI tokens, and AWS access keys with no matches. The protected local certification residue was not read, modified, or staged. Free disk was 213 GiB.

### 2026-09-03 — #4294 parent gate: CodeQL allocation correction

- Scope: parent-gate remediation only; CP02 remains unstarted. Foundation Atlas classification is **reuse** of the existing `runtime.direct-execution.v1` query binding contract; no connector, executor, source-lock, or CLI surface changes are needed.
- Red: GitHub Advanced Security CodeQL check run `100636096077` on candidate `ab5785b6d426e9db3423a43248588cd167923b38` reports one high-severity potential allocation-overflow alert at `internal/connectors/engine/direct_read.go:685`. `git blame` identifies the summed map-capacity expression as `fed381e132`, within the parent PR range after merge base `813f457a925f7ee3fe3bea101a43e445992c8552`; it is candidate-caused, not stale or external.
- Green plan: retain the fixed declared-query capacity only, so untrusted requested-field cardinality cannot participate in a preallocation sum; add a direct function proof that declared fixed and caller-requested typed values both survive the merge. Run the selector and the complete engine suite before refreshing the parent check.
- Green (local): `go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectReadQueryMergesDeclaredAndRequestedValues$'`, the complete engine suite, and `go vet ./internal/connectors/engine` passed. `git diff --check` passed. The new test proves both source-declared and typed caller query values remain present; the capacity now excludes caller-controlled cardinality. Remote CodeQL/verify/connector-boundary/package-check confirmation remains the parent gate.

### 2026-09-03 — CP02: closed sync execution contracts

- Scope: first half of #4424/S1A only. `synccontract` now owns closed progression, apply, object, binding, key, delete, and budget axes; `syncplan` owns immutable exactly-one result shapes; `syncrun` owns the additive version-one run-transition record. No resolver, app/CLI, connector, warehouse, credential, executor, filesystem, network, SQL, or clock wiring is included.
- Baseline gap: the prior mode-only contract could not distinguish snapshot replacement from cursor event append, history dedupe with composite keys/history-close, or change-capture tombstone apply; its persisted checkpoint envelope is unchanged.
- Green: four valid controls, every unknown closed discriminant, invalid budget, exactly-one result selection, canonical JSON round trips, and valid/invalid additive run transitions pass through pure validators. `go test -count=1 -timeout 20m ./internal/synccontract ./internal/syncplan ./internal/syncrun`, `go vet` on those packages, and `git diff --check` pass.

### 2026-09-03 — CP03: deterministic pure resolver

- Scope: `syncplan.Resolve` consumes only immutable plan axes and caller budgets. It returns a sealed reduced-budget plan or a stable typed incompatibility; it constructs no executor, credential, provider, database, warehouse, clock, or filesystem resource.
- Green: focused resolver proof asserts budget reduction, widening refusal, and incompatible snapshot/apply classification. `go test -count=1 -timeout 20m ./internal/syncplan`, `go vet ./internal/syncplan`, and `git diff --check` passed.

### 2026-09-03 — CP05: bounded lazy manifest store/cache

- Task Delivery Header: Refs #4424 — S1A API generic-engine migration; base `main` through existing PR #4294; delivery is one ordinary fast-forward commit to `origin/fm/cli-top100-declaration-batch-r1`; working branch `fm/cli-batch1-vnext-cutover-r2`.
- Scope: a new in-memory `manifeststore` consumes CP04's immutable `manifestindex.Index` and only loads indexed execution manifests on demand. It must bound retained entries and bytes, share one concurrent load without allowing one cancelled caller to cancel another, and cancel a loader with no remaining waiters. It adds no runtime reader, execution route, network behavior, provider I/O, credential path, source-lock use, or connector-specific branch.
- Foundation Atlas classification: reuse of `authoring.source-lock-vnext.v1`'s deterministic execution-manifest boundary; this is a local immutable-artifact cache, not a new connector runtime foundation.
- Required skill route: `connector-lane-build-order`, `go-engineering` advanced concurrency guidance, and `tdd` were loaded. The repository-named `golang-*` skills and `firstmate-exhaustive-review` are unavailable in this session; `go-engineering` is the recorded substitute and the existing inline/manual-GSD fallback remains authoritative.
- Red: `go test -count=1 -timeout 20m ./internal/connectors/manifeststore` failed to compile because public `New` and `Limits` do not yet exist. The focused behavior tests specify count and byte bounds, shared loading after a first caller cancels, cache reuse, cancellation of an abandoned load, and a fresh retry.
- Green: `manifeststore.New` requires positive entry and byte limits plus a loader; `Load` rejects unknown indexes before I/O, shares an owned-context loader per connector, returns private byte copies, maintains LRU limits, and cancels only flights with no remaining waiter. `go test -count=1 -timeout 20m -race ./internal/connectors/manifeststore`, `go test -count=1 -timeout 20m ./internal/connectors/manifestindex ./internal/connectors/manifeststore`, and `go vet ./internal/connectors/manifestindex ./internal/connectors/manifeststore` passed. The test suite observes entry and byte eviction through required second loader calls, oversize no-retention, unindexed refusal before loading, caller-mutation isolation, one loader shared across a cancelled first and live second waiter, cache reuse, and cancellation followed by a fresh retry.
- Broad check (deferred baseline): `go test -count=1 -timeout 20m ./internal/connectors/...` ran. CP05 packages passed; unchanged failures are `TestScanFailsClosedWhenConnectorMetadataCannotLoad/invalid_cli_surface`, `TestReviewMutationQueryControls`, `TestReviewDeleteRetryEvidenceIsOperationScoped`, `TestReviewStreamContractsAreProviderShaped`, and `TestRecoveredLegacyStreamMetadataPreserved` (the latter four require absent Recurly fixtures/research). Firstmate instruction `103.msg` defers these exact failures to final integrated Batch 1 verification. A whole-tree search finds `manifeststore` only in its new package and tests, so CP05 has no call path to them.

### 2026-09-03 — CP06: explicit construction and sealed compatibility

- Task Delivery Header: Refs #4425 — A1 manifest-selected executor registry; base `main` through existing PR #4294; delivery is one ordinary fast-forward checkpoint to `origin/fm/cli-top100-declaration-batch-r1`; working branch `fm/cli-batch1-vnext-cutover-r2`.
- Required skill route: `connector-lane-build-order`, `go-engineering`, `tdd`, and the Firstmate workspace copies of `golang-how-to`, `golang-structs-interfaces`, `golang-design-patterns`, `golang-safety`, `golang-concurrency`, `golang-context`, `golang-testing`, `golang-security`, `golang-error-handling`, `golang-data-structures`, and `firstmate-exhaustive-review`. CodeGraph is absent and Go LSP is unavailable, so the named built-in fallback caller inventory is the required reference record.
- Preflight base and preserved state: local/remote/API parent `cb33f2ef2584c7307f3c88b869c474796698b0fb`; 20 tracked CP06 candidate paths plus protected local residue predate this reconciliation and remain untouched. `go test -list` confirmed the factory controls and the actual App controls `TestDefaultRegistryContainsOnlyBuiltins` and `TestOpenConstructsTheExplicitProductionRegistry`; the formerly recorded `TestDefaultRegistryDoesNotUseProcessGlobalBuilder` selector selected zero tests and is not evidence.
- RED sequence after READY: (1) generated compact index list/lookup and digest-bound acquire prove zero-decode, one same-digest decode, bounded eviction, and exact identity; (2) duplicate/empty/unknown/multiple executor and extension IDs prove typed pre-construction refusal with zero constructor/I-O counters; (3) App/direct CLI prove one held generation/digest/executor, metadata zero-decode, and invalid selection before file/vault/approval/auth/request work; (4) exact 49 explicit hook factories and 3+4 native/compatibility inventories prove deterministic construction independent of import/registration order.
- Smallest GREEN sequence: first correct CP04/CP05's production bridge and generator-owned closed selection index without adding an execution reader; second route App and CLI through that construction; third replace every hook init/global registry path with explicit generated factories and split native database from four-row compatibility. Every slice starts with its own listed failing behavioral test and ends with its matching focused GREEN command.
- Disposition: READY for the first P1/P2 RED only. No preserved factory, App, CLI, generator, or Atlas change is a CP06 GREEN claim until it conforms to the reconciled `PLAN.md` allowlist and its own observable RED/GREEN evidence.
- Red [CP06-P2 closed executor vocabulary]: `go test -count=1 -timeout 20m ./internal/connectors/manifestindex -run '^TestIndexRejectsExecutorOutsideClosedVocabulary$'` exited 1. The behavioral assertion observed `manifestindex.New` accept `arbitrary_executor.v1`; no constructor or external I/O was involved. The test also requires an error carrying that rejected identity, so a generic validation string cannot satisfy the GREEN contract.
- Green [CP06-P2 closed executor vocabulary]: `manifestindex.New` now admits only the eight sealed executor identities and returns a typed error carrying the rejected value for every other identity. `go test -count=1 -timeout 20m ./internal/connectors/manifestindex ./internal/connectors/manifeststore` passed; existing store fixtures now use the explicit generic engine identity rather than an unrestricted placeholder.
- Red [CP06-P1 generated index owner]: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestGenWritesDeterministicManifestIndex$'` exited 1 because `connectorgen gen` emitted no `internal/connectors/manifestindex/index_gen.go`. The temporary execution-only `metadata.json` fixture supplied no source lock, credential, or provider input; the missing deterministic generated artifact is the observable failure.
- Green [CP06-P1 generated index owner]: `connectorgen gen` now derives `manifestindex/index_gen.go` only from the rendered runtime JSON file set, emits a stable generation/digest and one closed executor identity per entry, and rejects unsupported database driver identities. `go test -count=1 -timeout 20m ./cmd/connectorgen` and `go test -count=1 -timeout 20m ./internal/connectors/manifestindex` passed. The generated projection contains 553 closed entries, including generic GitHub, typed PostgreSQL, and the exact sealed legacy rows; it does not embed or read a source lock.
- Red [CP06-P5 App construction ordering]: `go test -count=1 -timeout 20m ./internal/app -run '^TestOpenWithRegistryBuildsRegistryBeforeProjectIO$'` exited 1. With a missing `.polymetrics` directory, `openWithRegistry` returned from `os.Stat` before invoking its supplied construction closure. The test observes the exact pre-I/O ordering defect without opening a vault, credential, or provider.
- Green [CP06-P5 App construction ordering]: `openWithRegistry` now invokes and validates the explicit registry construction closure before normalizing the root or statting/opening project state. `go test -count=1 -timeout 20m ./internal/app -run '^TestOpenWithRegistryBuildsRegistryBeforeProjectIO$'` passed; the missing-project witness observed construction and no vault/credential/provider path.
- Red [CP06-P4 native/compatibility partition]: `go test -count=1 -timeout 20m ./internal/connectors/native/nativeset -run '^TestCompatibilityAdapterExcludesNativeDatabases$'` exited 1 because `NewCompatibilityAdapter().Construct("dynamodb")` returned a native database connector. This is a reachable name-selected compatibility fallback and violates the required disjoint 3+4 inventory.
- Red [CP06-P2 protected native projection]: after the first generated-index cutover, `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^(TestNewLoadsDeclarativeBundlesWithProtectedNativeDatabases|TestProtectedNativeDatabasesRemainRegistered)$'` exited 1: generated `dynamodb` and `mysql` entries selected `api_engine.v1`, so the returned connectors were `*engine.Connector` rather than protected native adapters. These two retained native definitions lack `database.json`; their exact generated native IDs must therefore be a sealed generation-time inventory, never a runtime name fallback.
- Green [CP06-P2 protected native projection]: `connectorgen gen` now emits the sealed `native_database/dynamodb.v1`, `native_database/mysql.v1`, and `native_database/postgres.v1` entries before runtime construction. The focused generated-index and protected-registry selectors passed; runtime selection consumes those IDs and no longer chooses a database adapter by connector name.
- Green [CP06-P4 native/compatibility partition]: `nativeset` now exposes separate executor-keyed three-row database and four-row compatibility adapters. The compatibility adapter rejects all native IDs, and `bundleregistry` consumes generated executor IDs before choosing either adapter. Focused `manifestindex`, `nativeset`, and `bundleregistry` controls passed.
- Red [CP06-P4 global hook-route closure]: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestHookPackagesDoNotRegisterGlobalHooks$'` exited 1 at `internal/connectors/hooks/akeneo/hooks.go`; the same direct scan covers the full generated hook cohort and proves that `engine.RegisterHooks` remains an import-time execution route.
- Green [CP06-P4 global hook-route closure]: every one of the 49 hook packages now relies on its explicit factory; `connectorgen gen` validates that contract and emits the sorted `hookset.Factories` table without blank imports. `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestGenWritesDeterministicManifestIndex|TestHookPackagesDoNotRegisterGlobalHooks)$'`, `go test -count=1 -timeout 20m ./internal/connectors/hooks/...`, and affected engine/registry/Twenty/Zoom controls passed. The whole Recurly package still has the recorded missing-fixture/research baseline failures; its affected binary witness passed independently.
- Red [CP06-P3 one App/CLI construction]: `go test -count=1 -timeout 20m ./internal/cli -run '^TestConnectorCommandInputDefectsFailBeforeWithApp$'` exited 1 after an initialized fixture supplied an explicit `input-ordering` registry and credential. Direct-command preflight accepted that registry, but `App.Open` rebuilt the production registry and failed `connector "input-ordering" not found`; the two paths selected different construction instances.
- Green [CP06-P3 one App/CLI construction]: default direct command execution now passes its already-preflighted registry into `app.OpenWithRegistry`/`OpenForReverseExecutionWithRegistry`; injected opener tests remain an explicit test seam. The valid fixture reaches its declared post-credential execution boundary instead of losing the connector to a second production registry. `go test -count=1 -timeout 20m ./internal/cli -run '^TestConnectorCommandInputDefectsFailBeforeWithApp$'` passed.
- Red [CP06-P1 bounded manifest charge]: `go test -count=1 -timeout 20m ./internal/connectors/manifestindex -run '^TestIndexRejectsMissingByteCharge$'` exited 1 because an indexed manifest carried no bounded byte charge. A count-only cache cannot enforce the CP05 byte ceiling before allocating or retaining a bundle.
- Green [CP06-P1 bounded manifest charge]: the generated index now stores the conservative sum of every runtime-embedded execution JSON byte, and `manifestindex.New` rejects zero/negative charges before a store can retain the entry. `go test -count=1 -timeout 20m ./internal/connectors/manifestindex ./internal/connectors/manifeststore` and the Foundation Atlas selector check passed.

### 2026-09-03 — CP06 retrofitted impact disposition

- Firstmate inbox `109.msg` pauses CP06 production work. The prior CP06 RED/GREEN entries are preserved as historical observations, not a checkpoint claim: the current `bundleregistry` package cannot compile because `TestLoadDefinitionsCachesEmbeddedBundleSnapshot` still calls deleted `loadDefinitions`.
- Revalidated identity: local/remote/API parent is `cb33f2ef2584c7307f3c88b869c474796698b0fb`, PR #4294 remains `fm/cli-top100-declaration-batch-r1` to `main`, and frozen `0b214b79eeb871238ce8454cd7b896e71e2746a7` remains reachable. The current diff has 119 tracked paths; untracked local cache/certification residue was not read or changed.

| Impact lens | Current observable evidence | TDD implication |
| --- | --- | --- |
| Architecture/data flow | `Construction.BuildRegistry` acquires every generated entry. | RED must observe fleet decode; GREEN must observe zero-decode list and one selected acquisition. |
| Callers | CLI `appRegistry`, `app.Open`, and issue-label identity each construct production state. | RED/GREEN must cover one shared App/CLI construction root, not just a direct-command injection seam. |
| Interfaces/configuration | Index/handle omit safe metadata and held generation/digest identity. | Add a behavioral identity/mismatch test before changing the bridge. |
| Generated/docs surfaces | Generator emits only executor identity; Atlas names deleted/eager symbols; docs/skills call eager registry. | Add deterministic generation and surface-parity tests before claiming generated output current. |
| Compatibility | Three database and four compatibility IDs are split, but their registry package is unbuildable. | Add exact 3+4 factory coverage and stale-test replacement before factory GREEN. |
| Security | Only closure-before-`os.Stat` is observed. | Add invalid-selection zero filesystem/vault/approval/auth/request/constructor counters. |
| Concurrency/bounds | `BundleStore` lacks cancellation/retry and generation/digest production witnesses. | Add those tests before accepting its production use. |
| CLI/App reachability | Direct commands share a supplied registry, but metadata still fleet-builds it. | Prove list/lookup/decode counts and implemented/partial/unknown boundaries. |
| Provider semantics | No source lock or rendered execution JSON changed. | Keep provider paths out of CP06 tests and code. |
| Tests/evidence | Manifest index/store race tests and 49 hook packages pass; bundleregistry list/test compile fails. | No CP06 full GREEN checkpoint exists. |

| Prerequisite | Disposition |
| --- | --- |
| CP06-P1 | `blocked`: production bridge eagerly decodes and cannot prove exact held identity. |
| CP06-P2 | `blocked`: generated executor rows are closed but complete factory/extension validation is not proven. |
| CP06-P3 | `blocked`: App and CLI retain duplicate/eager production construction paths. |
| CP06-P4 | `blocked`: explicit factories replaced globals, but exact closed inventory and stale registry/Atlas proof migration remain. |
| CP06-P5 | `blocked`: construction-before-`os.Stat` is covered, but invalid-selection zero-I/O ordering is not. |

- Firstmate released `[key=cp06-impact-preflight]` through inbox `110.msg`; the historical blocked disposition above is superseded for the exact CP06 P1–P5 scope only.
- Actual RED [CP06-P1 lazy construction]: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestConstructionBuildsLazyRegistryWithoutDecodingList$'` exited 1 after the controlled two-entry construction decoded both bundles during `BuildRegistry`.
- Actual GREEN [CP06-P1]: generated entry metadata and command summaries build a lazy registry with zero list decodes; a named `Get` decodes one bundle. `go test -count=1 -timeout 20m -race ./internal/connectors/bundleregistry ./internal/connectors/manifestindex ./internal/connectors/manifeststore` passed.
- Actual RED/GREEN [CP06-P2]: `TestConstructionRejectsUnknownExtensionBeforeLoading` first observed a nil error for `hook/unknown.v1`, then passed after closed generated executor/extension validation was installed before loader/factory construction.
- Actual GREEN [CP06-P4]: `TestGeneratedHooksetConstructsRepresentativeHooks` proves 49 unique generated extension IDs with exact connector ownership; `TestHookPackagesDoNotRegisterGlobalHooks` remains the historical import-time-registration RED/GREEN guard. The Atlas owner/proof references now resolve in `go test -count=1 -timeout 20m ./cmd/connectorgen`.
- Actual RED/GREEN [CP06-P3]: App open initially resolved two lazy metadata entries through eager transport composition; root command help initially resolved one. `TestOpenWithLazyRegistryDoesNotResolveMetadataEntries`, `TestOpenLazilyRegistersDefinitionOwnedProductionTransports`, and `TestDynamicConnectorCommandsUseLazyMetadata` now pass.
- Actual GREEN [CP06-P5]: `TestOpenWithInvalidRegistryStopsBeforeProjectStat` observes one invalid construction and zero project-stat calls; the existing historical P5 RED/green closure-order proof remains retained.

### 2026-09-03 — CP06 rate-limit preservation overlay

- New direct controls from Firstmate inbox `111.msg`: RL-03 rate-file digest/byte charge; RL-04 `{connector,generation,digest}` store/flight/held-handle identity; RL-06 selected API/native/compat construction without a second definition load; RL-07 zero-I/O invalid selection; RL-09 same-scope GitHub process-budget reuse.
- Actual GREEN [Harness A]: `TestGeneratedManifestIndexDigestIncludesRateLimits` changes/removes only a temporary `rate_limits.json` and observes a changed/restored generated index.
- Actual RED/GREEN [Harness C]: `TestBundleStoreSeparatesConnectorGenerations` first reused the same-connector cache (`loader calls = 1`, want `2`). `AcquireEntry` now rejects generation/digest mismatch before its loader, keys cache/flights by the three-part identity, and holds exact selected generations. `TestBundleStoreRejectsGenerationOrDigestMismatch`, `TestBundleHandlePinsGenerationAfterCacheRelease`, and the store race suite pass.
- Actual GREEN [Harness B]: `TestLazyConstructionSharesGitHubRateAdmission` creates two production GitHub selections against only a local fake 429/reset; the shared public-IP scope sends once before reset while a second IP sends independently. It also observes selected process-local coordination.
- Actual GREEN [RL-06]: protected native and sealed compatibility factories now receive the selected `engine.Bundle`; focused native/nativeset suites passed. No GitHub/PostgreSQL rate declaration, source lock, provider resource, credential, rate service/daemon, shared-policy activation, certification route, or external review changed.

- Actual RED/GREEN [P1 metadata parity]: generated GitHub metadata first exposed `Catalog:false`; `TestGeneratedMetadataPreservesCatalogProjection` now passes with generated catalog capability and PostgreSQL native CDC parity.
- Actual RED/GREEN [P2 builtin collision]: a generated `sample` entry first built a registry with no refusal. `TestConstructionRejectsBuiltinIdentityBeforeLoading` now passes: reserved builtin identities fail before loader/factory work.
- Actual GREEN [generated ownership]: the initial whole-tree boundary scan reported handwritten generator literal branches and unclassified `manifestindex/index_gen.go`. Native/compat manifest selections now live in `nativeset`, runtime executor syntax is generic, and the generated-index classifier/ownership tests pass; the whole-tree boundary scan returns zero findings.

### 2026-09-03 — CP07 GitHub reference selection

- Base: `843a32de5f927b1235cc00883fa0c5e0f5ea8c5b`, the normally pushed CP06 parent checkpoint.
- Manual RED disposition: CP06 already generated GitHub's `api_engine.v1` / `hook/github.v1` entry, so no truthful fresh failing state remains. A test asserting another selection would be false. The CP07 witness is therefore a characterization-to-non-regression proof, recorded as the repository-approved manual-TDD fallback rather than a fabricated RED.
- GREEN contract: the production registry returns GitHub as `*engine.Connector`, its entry selects exactly `api_engine.v1` and `hook/github.v1`, and `RateLimitCoordinationOf` remains declared process-local. A native/compatibility return, missing/extra extension, or lost rate bundle fails the witness.

- Actual GREEN: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestGitHubReferenceUsesManifestSelectedAPIEngine$'` passed. It proves generated GitHub `api_engine.v1`/`hook/github.v1`, returned `*engine.Connector`, and declared process-local coordination from the production registry.

### 2026-09-03 — CP08 PostgreSQL reference selection

- Base: `c267f6ccb6988c6d0132f264e963c6701b8134f1`, the normally pushed CP07 reference checkpoint.
- Manual RED disposition: CP06 already generated PostgreSQL `native_database/postgres.v1` with no extension. A fresh opposite-state test would be fabricated; record the proof-only fallback explicitly.
- GREEN contract: generated PostgreSQL selection is exactly native database; production registry returns `nativepostgres.Connector`; no compatibility/API fallback or extension applies; rate coordination is absent because the selected rate declaration is explicitly not applicable.

- Actual GREEN: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestPostgresReferenceUsesManifestSelectedNativeDatabase$'` passed. It proves exact PostgreSQL native selection/no extension, `nativepostgres.Connector`, and no false rate coordination claim.

### 2026-09-03 — A1 consolidated correction wave

- Review BLOCK: A1-01 resolver admits `incremental_upsert + append` and lacks the required identity/durability axes; A1-02 store validates the requested index entry rather than loaded bytes; A1-03 normal `cli.Run` opens a second registry due `len(candidates)` ownership inference.
- Required RED/GREEN is defined in `PLAN.md`'s correction matrix. Corrections are one dependency-ordered wave; preserve RL-03/04/06/07/09, GitHub/PostgreSQL proofs, 49 hooks, 3+4 inventory, and all zero-I/O controls.

- A1-01 review RED: exact SHA `62e04650f` accepted `ModeIncrementalUpsert` with `DestinationApplyAppend`; review command/source evidence is `data/cli-a1-62e-phase-review-r1/report.md:69-100`. The newly added `TestResolveRejectsModeApplyContradiction` is the regression oracle and passed after the closed mode matrix.
- A1-01 GREEN: `ExecutionAxes` now carries retry/idempotency/receipt/acknowledgement/checkpoint; `Plan` carries source/target binding refs plus generation/artifact/executor/foundation/evidence digests; source/destination executor roles are exactly one each. Full seven-mode table tests, real-axis C3/C4 results, canonical JSON identity round-trip, and `FuzzResolveNeverExecutesContradictoryModeAxes` pass without runtime construction or I/O.
- A1-02 review RED: exact SHA `62e04650f` loader returned `*engine.Bundle` without loaded identity, allowing same-name bytes B to be cached as entry A; source evidence is the exact review report A1-02. The new `TestBundleStoreRejectsSameNameLoadedIdentityMismatch` covers generation, digest, and byte-charge mismatches.
- A1-02 GREEN: shared `manifestidentity.ForFS` calculates the closed execution identity for generator and `engine.Load`; `Bundle.Identity` and `manifeststore.LoadedBundle.Identity` both must equal the requested connector/generation/digest/charge before cache insertion, handle return, engine/native/compat factory, or rate resolver. The same-name mismatch constructor regression observes zero factory calls.
- A1-03 review RED: exact SHA `62e04650f` passed registry A into direct preflight then used default `app.Open` registry B because `len(candidates)` controlled ownership; source evidence is review report A1-03.
- A1-03 GREEN: `cli.Run` constructs one registry once and carries it in explicit production openers; default App/reverse/root manual/connectors/docs/skills routes reuse it. Test-only openers are an explicit mode. `TestRunRootRouterReusesPreflightRegistryForNormalHelpAndReverse` proves normal, dynamic help, and approval-token reverse-plan paths construct exactly one selected connector and never fall back to another registry. `TestSkillsGenerateMatchesTrackedSkills` also confirms a single lazy traversal remains under the existing time budget.

- Consolidated current GREEN: `go test -count=1 -timeout 20m ./internal/synccontract ./internal/syncplan ./internal/syncrun` passed; the five-second resolver fuzz completed 1,376,676 executions with no new contradictory executable; identity/store/index/registry race, normal router/skills race, retained rate/reference/hook/native, engine, generator, definitions, and commandrunner-preflight controls passed. No provider endpoint or credential was used.

### 2026-09-04 — A1 entry-capacity review correction

- Review BLOCK intake: Firstmate instruction `118.msg` identifies a distinct-key concurrency defect at exact SHA `701a0b45175f308400c938322fd1634a28efdaef`. Current `reserveLocked` constrains `len(cache)` and byte reservations but does not account for a pending flight as one entry, allowing two pending identities when `Bytes` permits both.
- Foundation Atlas: constrained extension of `definitions.bundle-loader.v1`; the changed guarantee is in-flight entry accounting and the catalog owner/proof inventory is updated in the same correction. No new foundation, source lock, connector capability, rate declaration, or CLI surface is planned.
- RED planned: `TestBundleStoreReservesEntryCapacityAcrossDistinctFlights` creates two identities, `Limits{Entries: 1, Bytes: 2}`, and deterministic loader barriers. It must fail before production changes by observing B start while A is pending, or by observing cached plus reserved count exceed one. It then exercises A cancellation while its loader remains blocked, A completion, B retry/completion, and A retry after B release.
- GREEN planned: pair a per-distinct-flight count reservation with the existing byte reservation under `BundleStore.mu`; every terminal load path releases both exactly once. Same-identity waiters do not consume a second slot; canceled last waiters retain the slot until the loader exits.
- Refactor planned: name/account the count reservation next to the byte reservation, retain error types and lock ownership, and add no public API or background work.
- Required commands: exact RED selector; exact GREEN selector with `-count=1 -timeout 20m`; `go test -count=1 -timeout 20m -race ./internal/connectors/manifeststore`; affected construction/identity selectors; scoped vet; Atlas validation; `go run ./cmd/agentcontractgen check`; `git diff --check`; frozen local self-review. All results remain pending until observed.
- Actual RED: `go test -count=1 -timeout 20m ./internal/connectors/manifeststore -run '^TestBundleStoreReservesEntryCapacityAcrossDistinctFlights$'` failed at `bundle_store_test.go:389-390`: while alpha was barrier-held, `Acquire(bravo)` returned `nil` instead of `ErrBundleCapacity`, and the one-byte fixture observed retained plus reserved entries `2`, not at most `1`. This reproduces the review finding without provider, credential, or rate activity.
- Actual GREEN: the exact regression passed after count reservation was paired with byte reservation in the existing lock scope. `go test -count=1 -timeout 20m -race ./internal/connectors/manifeststore` passed; `go test -count=1 -timeout 20m ./internal/connectors/manifestidentity ./internal/connectors/manifestindex ./internal/connectors/bundleregistry` passed; `go vet ./internal/connectors/manifeststore ./internal/connectors/bundleregistry` was clean; Atlas JSON/unique-ID checks and `TestFoundationAtlasSelectorsResolve` passed; `go build ./cmd/pm` and `go run ./cmd/agentcontractgen check` passed. No provider, credential, or rate activity occurred.
- Independent exact-SHA review PASS: Firstmate instruction `120.msg` accepted `c761e7e6f2d042c7560ab0c520dc9aa182110e6e` with zero blockers. The A1-04 RED/GREEN evidence and frozen local review are accepted for normal parent publication; the reviewed candidate is not amended.

### 2026-09-04 — CP09 strict source parsing and canonical graph

- Base: A1-04 review evidence is committed and normally published through `988dd16c3d206a28d3e7b16f8a0d805c4163f7ca`; Firstmate instruction `120.msg` authorizes this first #4426/N2 checkpoint without a reset or rebase.
- Discovery correction: the impact report names a removed `cmd/connectorgen/sourceimport.go` decoder. In the current parent, `decodeStrictJSON` in `cmd/connectorgen/vnext_lock.go` is the sole strict JSON implementation and already serves both schema-4 locks and the S1A source-parent matrix. CP09 extends that one decoder/canonicalizer rather than resurrecting a legacy parser or adding a second reader.
- RED planned: `TestRunLockRenderRejectsInvalidCanonicalGraphBeforeWriting` supplies compact schema-4 locks mutated one fact at a time: unknown raw stream field, wrong request/record role, unsupported body encoder, normalized command-path alias collision, missing record binding, and non-object source identity. Every case must return a stable source JSON-pointer diagnostic while a pre-existing generated-file sentinel stays byte-identical. `TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering` reverses same-rank source operations and requires equal closed rendered bytes/digest plus retained source IDs.
- GREEN planned: parse only once; build typed canonical operation/stream/write/operation/command nodes while retaining provider, CLI, and operation source facts; bind roles only to structurally compatible lane nodes; canonicalize tie ordering by immutable source identity; reject normalized alias duplication; and run the existing static engine bundle loader over an in-memory rendered graph before any writer. `engine.Load` is structural validation only and never constructs an `engine.Connector`, credential, transport, provider request, resolver, or preflight path. CP10 remains responsible for semantic source-to-execution/manifest/resolver/preflight admission.
- Refactor boundary: source-route matching is a CP10 semantic join and remains out of scope. No source lock or generated execution artifact is changed, no provider source is fetched/re-pinned, no runtime executor or second parser is introduced, and no semantic join is pre-implemented. The existing `authoring.source-lock-vnext.v1` Atlas entry will be updated together with its proof list because the graph contract changes.
- Required GREEN commands: exact CP09 test selectors, complete `./cmd/connectorgen` test package, read-only `lock-render <reference> --check` controls, `connectorgen validate internal/connectors/defs`, Atlas selector proof, scoped vet/build, `agentcontractgen check`, `git diff --check`, then inline/manual `verify-work` and `code-review`. Actual RED/GREEN results are recorded only after execution.
- Actual RED: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestRunLockRenderRejectsInvalidCanonicalGraphBeforeWriting|TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering)$'` exited 1. Every malformed case returned `runLockRender() = 0` and replaced the sentinel (`unknown stream execution field`, wrong request role, unsupported `body_type`, normalized command alias, missing record binding, and source-route mismatch). Reversing same-rank stream operations changed `streams.json` bytes. No provider, credential, runtime executor, or source artifact publication occurred; the temporary test roots were removed by the test framework.

- Scope correction after RED: a source route mismatch is provider-fact-to-execution semantic admission and is therefore CP10 work. The final structural sentinel case is a non-object `source` value; valid direct-operation request/response references are explicitly admitted by `TestVNextCanonicalGraphAllowsDirectOperationSchemaRoles`.
- Actual GREEN: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestRunLockRenderRejectsInvalidCanonicalGraphBeforeWriting|TestVNextCanonicalGraphAllowsDirectOperationSchemaRoles|TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering|TestEverySourceLockBuildsCanonicalGraph)$'` passed. It proves strict no-write rejection, direct-operation role admission without a premature semantic join, invariant rendered bytes/digests under irrelevant source-operation ordering, retained source facts, and graph admission for every committed schema-4 lock.
- Actual GREEN: `go test -count=1 -timeout 20m ./cmd/connectorgen` passed. It includes the Foundation Atlas owner/proof inventory check and all source-lock renderer, strict-decoding, graph, and matrix controls.
- Actual GREEN: read-only `go run ./cmd/connectorgen lock-render asana --check`, `... github --check`, and `... gitlab --check` each reported `vNext execution bundle is current`; `go run ./cmd/connectorgen validate internal/connectors/defs` reported `553 connector(s) checked, 0 findings`.
- Actual GREEN: `jq empty docs/connector-canon/foundations/catalog.schema.json docs/connector-canon/foundations/catalog.json` and the unique-ID assertion passed; `make docs-check`, `go vet ./cmd/connectorgen`, `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, and final `git diff --check` passed. Atlas revision 30 records the constrained `authoring.source-lock-vnext.v1` graph extension and the source-lock canon documents structural role placement without claiming CP10 joins.
- GSD: `scripts/gsd prompt execute-phase batch-r1-vnext-cutover`, `... verify-work ...`, and `... code-review ...` resolved and were executed inline under the recorded unavailable-worker/manual fallback. No provider, credential, runtime executor, source-lock publication, or rendered execution publication occurred; preserved `.cache` and certification residue were not staged or changed.
- Delivery: CP09 was committed as `d11277378abe556323226e3f6998ce3caf6033dc` and normally pushed from the working branch to `origin/fm/cli-top100-declaration-batch-r1`; GitHub PR #4294 reports that exact head with `base=main` and `draft=true`. This is a parent checkpoint only; no source-lock/rendered execution artifact was published and CP10 remains unstarted.

### 2026-09-04 — CP10 semantic source-execution admission

- Authority/base: Firstmate instruction `122.msg` accepts CP09 and starts the serial second #4426/N2 checkpoint from normally published parent `85c28e70e4c8f811ea342a1f1054e09759cde1c1`. CP10 remains authoring-only/in-memory: no connector lock, rendered bundle, global-output, provider, credential, or publication change.
- Required skill/lifecycle route: `go-engineering`, `tdd`, `connector-lane-build-order`, and `connector-migration-exact-sha-review` are loaded; `golang-how-to` is unavailable. Every GSD lifecycle command resolves through `scripts/gsd sources`; the custom phase lacks the adapter ROADMAP and retains the known missing `issue-122-rebootstrap.md` doctor blocker, so discussion/plan/execute/verify/review prompts are recorded and fulfilled inline without fabricating a worker.
- RED planned: introduce a compact two-operation GraphQL fixture that shares `/graphql` but binds one command to the other operation, plus one-fact staged mutations for missing/duplicate/cross-connector/stale identity, wrong schema/binding, invalid rate declaration, and unsatisfied supplied resolver facts. Each refusal must name immutable source ID and JSON field path while `runLockRender` preserves a pre-existing output sentinel.
- GREEN planned: retain one canonical graph and build a deterministic in-memory staged generation. It must use `engine.Load`, `engine.New`, `commandrunner.Preflight`, `manifestindex.New`, `nativeset.ManifestSelections`, and `syncplan.Resolve`; provenance binds schemas, streams, writes, operations, commands, optional rate JSON, manifest/index identity, and explicitly supplied sync/Atlas facts without inventing a missing endpoint.
- Required commands: exact RED selector; exact GREEN selectors with `-count=1 -timeout 20m`; complete `./cmd/connectorgen`; read-only reference `lock-render --check` controls; `connectorgen validate`; Atlas JSON/selector and docs checks; scoped vet/build; `agentcontractgen check`; `git diff --check`; inline/manual verify and code review. Actual results are recorded only after observation.
- Actual RED: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextSemanticAdmissionRejectsSwappedGraphQLOperationBeforeWriting$'` exited 1. A source node for `source:widgets.first` declared fixed `FirstWidgets` while its command was made internally valid for the separate `widgets.second`/`SecondWidgets` operation at the same physical `/graphql` endpoint. Current `runLockRender` exited 0 and replaced the output sentinel with a six-file bundle. The loader therefore cannot prove source-operation identity by itself; no provider, credential, or live I/O occurred.
- Actual GREEN: after adding the in-memory semantic-admission stage, the same exact selector passed: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextSemanticAdmissionRejectsSwappedGraphQLOperationBeforeWriting$'` → `ok`. The stage rejects the cross-source GraphQL operation binding at `source operation "source:widgets.first" field /operations/0/commands/0/operation` before `runLockRender` can write the sentinel.
- Actual GREEN: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextSemanticAdmission(RejectsSwappedGraphQLOperationBeforeWriting|StagesRateIdentityAndManifestIndex|RejectsMalformedRateBeforeWriting|RejectsCrossConnectorManifestAndResolvesSuppliedSync|RunsResolverAndPreflightBeforeWriting)$'` → `ok`. This proves deterministic in-memory provenance and manifest-index inputs; rate addition/change/removal identity behavior; malformed rate no-write refusal; exact source/digest/executor sync admission through `syncplan.Resolve`; and resolver/preflight route and flag-binding rejection before replacement.
- Actual GREEN: `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (all source-lock canonicalization, semantic admission, Foundation Atlas selector, deterministic-render, and generator tests).

### 2026-09-04 — N2 boundary-review correction

- Authority/base: Firstmate instruction `124.msg` accepts the independent N2 exact-SHA BLOCK at `56ec3d9d7dc1d726203b0ef0c03ddec3209b8dde` as three bounded CP09/CP10 corrections. CP11 remains prohibited. The staged path remains in-memory only: no source lock, rendered/global output, provider/credential I/O, certification, filesystem staging, activation, or publication behavior changes.
- Foundation/lifecycle/skills: this is a **constrained_extension** of `authoring.source-lock-vnext.v1`. Existing owner authorities are `engine.Load`, `engine.New`, `hookset.Factories`, `nativeset.ManifestSelections`, `manifestindex.New`, commandrunner preflight, and syncplan resolve. `go-engineering`, `tdd`, `diagnose`, `connector-lane-build-order`, and `connector-migration-exact-sha-review` are loaded; `golang-how-to` and `golang-testing` are unavailable. All GSD command sources and discuss/plan prompts resolve; the established missing ROADMAP/`issue-122-rebootstrap.md` blocker requires documented inline/manual execution.
- RED planned — effective schemas: a temp source lock changes only a request or response reference to a second valid shared schema while its write/stream retains the first actual runtime schema. `runLockRender` must currently replace a sentinel; GREEN must refuse at the authored `schema_refs` field before replacement.
- RED planned — hook manifest: the GitHub source lock's staged entry currently differs from the closed generated production entry because `Extension` is empty. GREEN must produce exact full-entry/index equality including `hook/github.v1` and use the selected hook in a local engine/preflight path without provider or credential work.
- RED planned — provenance: reverse the existing same-rank two-operation fixture. Current rendered bytes/digest remain equal while staged provenance differs because it retains authored operation indexes. GREEN must make full staged provenance equal and retain authored source-array locations in a separate malformed-input diagnostic.
- Required GREEN commands: exact selector count/list and each RED/GREEN selector; full `./cmd/connectorgen`; reference check-only renders; definition validation; Atlas JSON/selector and docs checks; scoped vet/build; agent-contract; final diff check; inline/manual verify and code review. Actual results are recorded only after observation.

- Actual RED — effective schemas: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextSemanticAdmissionRejectsSwappedEffectiveSchemaBeforeWriting$'` exited 1. Both a request reference swapped to a second valid registry schema and a response reference swapped to a second valid registry schema reached `runLockRender() = 0` and replaced the output sentinel (`7` and `5` generated files respectively). Current admission proves only registry existence, not the actual loaded write/stream schema.

- Actual GREEN — effective schemas: after binding request roles to the loaded write/direct body schemas and response roles to the loaded stream schema, the exact selector passed: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextSemanticAdmissionRejectsSwappedEffectiveSchemaBeforeWriting$'` → `ok`. Both swapped existing schema references now fail before output replacement at their authored `schema_refs` pointers.

- Actual RED — GitHub hook manifest: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextSemanticAdmissionStagesGitHubProductionHookEntry$'` exited 1. The real GitHub lock passed schema admission, then refused the supplied closed production entry because `manifest extension "hook/github.v1" does not match staged ""`. No provider, credential, or output write occurred.

- Actual GREEN — GitHub hook manifest: selection now uses the closed native inventory plus `hookset.Factories`, constructs the selected hook before `engine.New`, and carries `Extension` into the staged entry/index. The real-lock selector passed: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextSemanticAdmissionStagesGitHubProductionHookEntry$'` → `ok` (including exact generated GitHub entry/index and a non-nil selected `hook/github.v1` with `ConnectorName() == "github"`). No provider or credential path was exercised.

- Actual RED — canonical provenance: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering$'` exited 1. The existing semantic-equality fixture retained identical rendered output/digest, but staged provenance changed from the original author array indexes (`stream:gadgets` at `/operations/1/...`, `stream:widgets` at `/operations/0/...`) under a semantic reorder.

- Actual GREEN — canonical provenance: after assigning source-ID-sorted canonical positions and using them only in staged provenance, the exact selector passed: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering$'` → `ok`. Reordered equivalent locks now have identical full staged provenance; the intentionally malformed reordered lock still reports the original author-visible `/operations/0/schema_refs/record` diagnostic.

- Actual GREEN — direct-role boundary and regression: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextSemanticAdmissionRejectsSwappedEffectiveSchemaBeforeWriting|TestVNextSemanticAdmissionStagesGitHubProductionHookEntry|TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering|TestVNextSemanticAdmissionRejectsUnboundDirectOperationSchemaRoles)$'` → `ok`; the direct-only request/response test proves that a structurally valid role with no typed effective runtime schema cannot become a provenance-only admission. Full `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 51.013s.
- Actual GREEN — canon/Atlas: catalog revision `32` parsed and retained unique IDs; `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestFoundationAtlasSelectorsResolve$'` → `ok`; `make docs-check` built `pm` and validated connector docs. `SOURCE-LOCK-VNEXT.md` and the Atlas now state the effective request/response joins, closed hook extension staging, canonical provenance identity, and direct-role rejection boundary.

### 2026-09-04 — CP11 B1 transactional publication impact-ready

- Authority/base: Firstmate instruction `125.msg` records a fresh N2 correction-review PASS for immutable `36e4d980de0d51d92fe74a68306845643596a6cb` and authorizes #4427/B1 only from that exact head. CP12 remains prohibited. The refreshed ten-row map in `PLAN.md` is `READY` before the first production test or publisher edit.
- Scope: CP11 consumes only CP10's already-admitted `vNextStagedGeneration`. It may add a connector-local generation publisher/resolver, journal, integrity metadata, and lock-render wiring over temporary roots. It must not materialize any real connector/source lock, modify `defs.FS`, App/CLI runtime routing, connector data, source locks, rendered execution JSON, provider/credential/database path, `.cache`, or certification residue.
- RED planned — closed publication: a current per-file write permits an optional `rate_limits.json` to survive when the next closed set omits it; a reader can combine old index metadata with new execution bytes. A temporary-root test must first demonstrate this torn/stale state, then require one `CURRENT`-resolved generation to contain exact execution, manifest, provenance, Atlas-reference, compact-index, proof, and integrity files with no extra executable member.
- RED planned — durability/recovery: inject an error immediately before and after every staged-file fsync, stage-directory fsync, journal write/fsync, pointer temp write/fsync/rename, pointer-parent fsync, active validation, commit record, and prune. Restart/recover after each cut point must yield a valid old or new complete set only; failed stage preserves old `CURRENT`, while failed active validation restores or durably completes one valid pointer.
- RED planned — concurrency/pruning: barrier-held old readers and two marker-distinct publishers must prove that a writer cannot interleave, a reader observes one pointer-bound bundle/index pair, and stale generation pruning skips the old generation until its handle releases. Use barriers/channels and temporary roots; no sleep, provider, credential, source-lock materialization, or cross-connector transaction.
- GREEN planned: reuse CP10 semantic admission as the sole pre-stage validator, write the exact artifact inventory beneath same-parent staging, fsync every durable boundary, atomically replace only `CURRENT`, recover journal state, and prune only generation directories proven unheld. The resolver accepts only `CURRENT`; it never falls back to flat artifacts.
- Refactor boundary: consolidate only publication primitives after every state-machine test is green. Preserve source-lock canonicalization, execution renderer, `manifestidentity` digest algorithm, runtime embedded snapshot, and existing command arguments unless a test proves an in-allowlist contract change is necessary.
- Required commands: list exact selectors before RED/GREEN; run focused publisher/lock-render tests including `-race`, affected `./cmd/connectorgen`, source-lock in-memory parity (not real publication), docs/Atlas checks, scoped vet/build, agent contract, diff check, secret scan, inline/manual execute/verify/review prompts, and normal PR read-back. Record only observed results below.


### 2026-09-04 — CP11 B1 transactional publication actual RED/GREEN

- Inline lifecycle: `scripts/gsd doctor` reported only the established missing
  canonical issue prompt. `sources` and generated prompts for
  `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
  `code-review` all resolved. With no adapter roadmap or compatible isolated
  worker/reviewer, the command procedures are executed inline and recorded in
  the plan, verification, and review ledgers.
- Actual RED — orphan before journal: `go test -count=1 -timeout 20m
  ./cmd/connectorgen -run '^TestVNextGenerationPublisherRecoversOrphanBeforeJournalCommit$'`
  initially failed because `Recover()` retained two generation entries after a
  crash before durable journal creation. GREEN removes only the abandoned
  staged publisher generation and retains the exact old complete marker.
- Actual RED — failed active validation: `go test -count=1 -timeout 20m
  ./cmd/connectorgen -run '^TestVNextGenerationPublisherRollsBackFailedActiveValidationWithoutOrphan$'`
  initially failed because post-switch validation left a rejected generation
  beside old `CURRENT`. GREEN restores/removes the pointer as applicable,
  removes the rejected validated generation, fsyncs its parent, and then clears
  the journal.
- Actual RED — repeat publish recovery: `go test -count=1 -timeout 20m
  ./cmd/connectorgen -run '^TestVNextGenerationPublisherRepublishingActiveSetRemainsRecoverable$'`
  failed with `journal old and new generation are equal` after an injected
  repeated-set crash. GREEN treats an already exact active generation as a
  no-op before writing a journal or `CURRENT`.
- Actual RED — digest framing: `go test -count=1 -timeout 20m
  ./cmd/connectorgen -run '^TestVNextPublicationGenerationIDDelimitsArtifactNamesAndBytes$'`
  demonstrated two distinct artifact maps sharing the delimiter-based
  generation input. GREEN length-prefixes every artifact name and payload
  before SHA-256 input.
- Actual RED — unsafe paths: `go test -count=1 -timeout 20m
  ./cmd/connectorgen -run '^TestVNextGenerationPublisherRejectsUnsafeArtifactPathsBeforeWriting$'`
  accepted `schemas/.hidden.json` and created a root for a NUL-bearing name.
  GREEN rejects hidden path components, separators/traversal, reserved names,
  backslashes, and control characters before creating the generation root.
- Actual RED — unowned pruning: `go test -count=1 -timeout 20m
  ./cmd/connectorgen -run '^TestVNextGenerationPublisherRefusesToPruneUnownedGeneration$'`
  deleted a generation-shaped directory containing a regular lease and
  author-owned sentinel. GREEN requires the candidate tree's own integrity and
  closed-tree validation before it is publisher-owned enough to prune.
- Actual RED — symlinked generation root: `go test -count=1 -timeout 20m
  ./cmd/connectorgen -run '^TestVNextGenerationPublisherRefusesSymlinkedGenerationRootWithoutDeletingTarget$'`
  followed `generations/` to an author-owned target and deleted its `.stage-*`
  content during recovery. GREEN permits removal only beneath a non-symlink
  publisher-owned generation root verified by `Lstat` before recovery, staging,
  checking, opening, or pruning.
- Actual GREEN — closed set/check/read-only security: the publisher tests
  cover omitted `rate_limits.json`, exact `CURRENT` byte comparison, journal
  refusal without recovery, source-lock/flat-artifact preservation,
  symlinked-current/stale-stage/generation-root refusal, bounded control-file
  reads, and no removal of an unowned directory. The staged real lock-render
  test proves physical semantic revalidation through the existing
  loader/selection/preflight without provider or credential I/O.
- Actual GREEN — fault/concurrency: `go test -race -count=1 -timeout 20m
  ./cmd/connectorgen -run '^(TestVNextGenerationPublisher.*|TestVNextPublicationGenerationIDDelimitsArtifactNamesAndBytes|TestRunLockRenderPublishesOnlyClosedGeneration)$'`
  → `ok`. Every declared fsync/rename/validation/commit/prune cut point recovers
  to an exact old or new complete generation; the barrier test rejects writer
  interleave and mixed metadata/index reads; held leases defer stale prune.

### 2026-09-04 — CP11 exact-review correction plan (instruction 126)

- Authority/base: Firstmate accepted the independent CP11 exact-SHA BLOCK at
  `c4f0bc3728dda318ea3d01f78de7aa299b6135cb` as five bounded corrections.
  Only temporary publication roots are allowed; CP12, actual connector/source
  lock materialization, runtime routing, provider/credential I/O, certification,
  cross-connector transactions, and author-owned deletion remain prohibited.
- RED 1 — descriptor confinement: make a valid connector entry or its
  `.connectorgen.lock` a symlink to an external sentinel. Probe every public
  publisher operation and `runLockRender` for a missing uniform pre-access
  refusal; GREEN rejects the link via rooted no-follow handling before target
  mutation.
- RED 2 — stage ownership: add `.stage-author-owned` with a sentinel and no
  valid typed marker. `Recover`, `Publish`, `Open`, and `Prune` must currently
  remove it; GREEN preserves it and refuses the operation, while a
  publisher-created/synced marker still permits stale-stage recovery.
- RED 3 — closed-tree identity: copy a valid generation under another valid
  `g-*` basename and rewrite self-consistent pointer/integrity identity; add an
  unexpected empty directory or nonempty lease. Current validation accepts
  these shapes; GREEN recomputes the framed artifact address and requires exact
  directory/lease invariants on every check/open/recovery/prune path.
- RED 4 — control documents: inject oversized or duplicate-member `CURRENT`,
  journal, integrity root, and integrity-file records. Current integrity
  validation can allocate unbounded bytes and duplicate fields use
  last-member-wins; GREEN uses one bounded rooted reader plus strict duplicate
  rejection and leaves durable state unchanged.
- RED 5 — cancellation: hold the actual connector lock from another descriptor
  and cancel each waiting public publisher/CLI operation. Current blocking
  `Flock` cannot return; GREEN returns the supplied context error without a
  goroutine leak or state change, then succeeds after release.
- Refactor condition: factor rooted control/file/lock helpers only after each
  observable behavior is green; do not widen the publisher into a runtime
  resolver or add a second source-lock reader.

### 2026-09-04 — CP11 exact-review correction actual RED/GREEN

- Base discipline: all correction evidence started from the instruction-126
  authorization at `c4f0bc3728dda318ea3d01f78de7aa299b6135cb`; no CP12,
  connector/source-lock materialization, provider, credential, or runtime-route
  work was performed.
- RED 1 observed: before rooted connector handling,
  `TestVNextGenerationPublisherRefusesSymlinkedConnectorRootWithoutTouchingTarget`
  reported `Publish`, `Recover`, and `Prune` as successful against an external
  symlink target; `Check` failed only for a missing lock and `Open` only for no
  active generation. The test therefore exposed the missing boundary rather
  than a safe refusal. The lock-specific fixture was added after that root
  correction; its pre-change behavior is the independent exact-review finding,
  not a retroactively claimed local RED run.
- GREEN 1: `os.Root`-relative connector/lock handling rejects root and lock
  symlinks before access; source-lock/control reads and control replacement/
  removal are rooted. The external-sentinel root, lock, and lock-render tests
  pass.
- RED 2 observed: absent or malformed `.stage-author-owned` markers let
  `Recover`, `Publish`, `Open`, and `Prune` return nil and remove the sentinel
  stage. GREEN writes and fsyncs a typed `{version,connector,generation,stage}`
  marker, preserves/refuses unproven stages, and still recovers a valid
  publisher-owned stage.
- RED 3 observed: every operation in
  `TestVNextGenerationPublisherRejectsSelfConsistentRenamedGeneration` accepted
  a renamed tree whose pointer and integrity had been rewritten consistently;
  `Check` also accepted an unexpected directory and a nonempty lease. GREEN
  streams the sorted artifact bytes through the publication framing on every
  validation, requires the selected generation name, exact directories/members,
  and an empty lease.
- RED 4 observed: duplicate members in `CURRENT`, `JOURNAL`, integrity root,
  and an integrity file entry all let `Recover` return nil. An integrity payload
  over the limit was fully read and failed later as invalid JSON instead of at
  the bound. GREEN routes the controls through one bounded rooted reader and
  `decodeStrictJSON`; all duplicate and oversized cases refuse deterministically.
- RED 5 manual fallback: the authorized base used blocking
  `syscall.Flock` without a context path. Executing cancellation against that
  blocking call would not return, so the source-level red is recorded rather
  than faking a timed run. GREEN uses nonblocking retry with the supplied
  context; the publisher and real `runLockRenderContext` contention tests return
  `context.DeadlineExceeded`, preserve `CURRENT`/absence of `JOURNAL`, and retry
  after release without a retained lock handle.
- GREEN command:
  `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSymlinkedConnectorRootWithoutTouchingTarget|TestVNextGenerationPublisherRefusesSymlinkedLockWithoutTouchingTarget|TestRunLockRenderRefusesSymlinkedConnectorRootWithoutTouchingTarget|TestVNextGenerationPublisherPreservesUnownedStageAcrossMutations|TestVNextGenerationPublisherRecoversDurablyOwnedStage|TestVNextGenerationPublisherRejectsSelfConsistentRenamedGeneration|TestVNextGenerationPublisherRejectsUnexpectedDirectoryAndNonemptyLease|TestVNextGenerationPublisherRejectsDuplicatePublicationControlMembers|TestVNextGenerationPublisherRejectsOversizedIntegrityControl|TestVNextGenerationPublisherContextCancelsContendedLockWithoutStateChange|TestRunLockRenderContextCancelsContendedPublicationAndRetries)$'`
  → `ok`.
- Broader changed-package GREEN:
  `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (59.779s).

### 2026-09-04 — CP11 correction-review repair TDD plan (instruction 127)

- Authority/base: Firstmate accepts the review block as five further bounded
  repairs from immutable
  `4fedb3875cbe7071799aed0e9b6ce1e34257f95e`. CP12, checked-in source-lock
  materialization, rendered outputs, runtime routing, provider/credential I/O,
  certification, cross-connector transactions, author-owned deletion, `.cache`,
  and certification residue remain prohibited.
- RED 1 — descriptor lifetime/no-follow: use a publisher mutation hook to
  rename the verified connector and replace its original pathname with an
  external tree containing a valid-looking owned stage and sentinel. Existing
  absolute-path cleanup can delete that external sentinel. GREEN retains
  connector/generations descriptors and expresses all actions through atomic
  no-follow descriptor-relative operations.
- RED 2 — component boundary: scan a real fixture containing
  `vNextCanonicalCommand`; the compact `cal-com` identifier alias crosses
  `Canonical`/`Command` and currently produces a false policy finding. GREEN
  rejects that accidental join while retaining a real `calCom...Policy`
  finding.
- RED 3 — cancellation acquisition race: coordinate cancellation with release
  of the competing connector lock after the nonblocking attempt begins.
  Existing code can acquire and mutate after cancellation. GREEN rechecks
  context, unlocks, reaches no mutation hook, preserves pointer/journal/tree,
  and permits a later retry.
- RED 4 — journal order: assert a durable prepared journal exists before final
  rename and inject faults at prepared-journal-before-rename and
  final-rename-before-CURRENT. Existing code activates the final directory
  first. GREEN follows the binding canon and recovers both cuts to a valid
  old/new complete selection.
- RED 5 — Atlas proof mapping: mutate an Atlas guarantee mapping to remove its
  positive or negative proof. Existing selector validation accepts a nonempty
  unrelated proof list. GREEN requires exactly one known mapping per guarantee
  and resolving positive/negative test symbols.
- Refactor condition: finish each red/green vertical slice before extracting the
  descriptor helper. Do not use a pathname fallback, generic filesystem route,
  runtime reader, or connector-specific bypass.

### 2026-09-04 — CP11 correction-review repair actual RED/GREEN (instructions 127–129)

- Scope: correction starts from immutable
  `4fedb3875cbe7071799aed0e9b6ce1e34257f95e`; CP12, checked-in connector
  materialization, provider/credential/database I/O, runtime routing,
  `.cache/`, and certification residue remain untouched.
- RED 1 observed: replacing `<defs>/acme` after operation setup redirected
  absolute-path Publish and Recover cleanup into an external sentinel tree; the
  prior Prune path could delete that sentinel while returning success. GREEN
  `TestVNextGenerationPublisherRetainsDescriptorsAcrossConnectorReplacement`
  retains the original inode across Publish, Recover, and Prune. The
  lock-render-specific
  `TestRunLockRenderRetainsSourceDescriptorAcrossConnectorReplacement` proves
  that its source reread and publish use that same retained connector
  descriptor.
- RED 2 observed: `TestScanMatchesCompactIdentifierAliasesOnlyAtComponentBoundaries`
  found a `cal-com` policy match inside `vNextCanonicalCommand`. GREEN preserves
  `calComOutputPolicy` while rejecting that interior accidental join.
- RED 3 review defect: acquisition could win after cancellation and enter
  caller mutation. GREEN
  `TestVNextGenerationPublisherCancelsAfterLockAcquireBeforeMutationAndRetries`
  uses the post-acquire barrier, observes `context.Canceled`, no staging hook,
  unchanged CURRENT/JOURNAL/generation tree, and a successful retry.
- RED 4 review defect: final rename preceded the durable prepared journal.
  GREEN `TestVNextGenerationPublisherWritesPreparedJournalBeforeFinalRename`
  observes a prepared journal and owned stage before final generation
  visibility; `TestVNextGenerationPublisherRecoversPreparedJournalAfterFinalRename`
  observes old CURRENT and restores the old complete generation after the
  renamed/pre-CURRENT cut.
- RED 5 review defect: Atlas selector validation accepted publication
  guarantees without one resolving positive/refusal pair. GREEN
  `TestVNextPublicationProofContractRejectsOmittedGuarantee` rejects the
  omission, while `TestFoundationAtlasSelectorsResolve` checks the real
  source-lock publication record. Per instruction 128, this enforcement is
  source-lock-only rather than a catalog-wide migration.
- Additional admission regression: a source change after lock acquisition is
  refused before creating a generation directory by
  `TestRunLockRenderRejectsSourceMutationBeforeGenerationCreation`; it prevents
  a correction-induced no-write-before-admission regression.
- GREEN commands: focused source/Atlas tests → `ok` (1.357s);
  `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (61.431s);
  `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (325.701s);
  `go test -count=1 -timeout 20m ./internal/connectors/boundary -run
  '^TestScanMatchesCompactIdentifierAliasesOnlyAtComponentBoundaries$'` →
  `ok` (0.411s); and the real boundary scanner returned
  `"outcome":"clean"` for 553 connectors and 280 checked files.

### 2026-09-04 — CP11 final repair plan (instructions 130–131)

- Authority/base: Firstmate instruction `130.msg` BLOCKs immutable
  `f7a325aec3594635acbd27e39099640283ca3663` and authorizes only F1–F4;
  instruction `131.msg` required full reading, explicit status acknowledgement,
  and archival of `130.msg` before further work. CP12 remains prohibited.
- Red 1: deterministic hooks replace the fsynced temporary entry for each
  `CURRENT` and `JOURNAL` atomic control write. The test must observe that the
  old control bytes, original temporary inode, and replacement remain intact
  after refusal.
- Red 2: stage, generation/lease, rollback, and control-cleanup barriers
  replace the exact named entry after marker/integrity/lease/control validation.
  Recover, Prune, and rollback must refuse rather than remove either object.
- Red 3: a held first operation loses the visible lock pathname at its
  post-acquisition barrier; a second operation must refuse or wait rather than
  enter a separate `Flock` domain, and a restored bound lock must retry.
- Red 4: a real implemented command is made preflight-invalid only in the
  physical staged bundle after semantic admission. Its refusal must include the
  staged preflight boundary; Atlas mappings must use the valid-stage and
  unowned-generation tests and reject multi-guarantee mappings.
- Green: add only publication-local descriptor identity binding and proof
  mappings under `authoring.source-lock-vnext.v1`; then run focused selectors,
  full/race `cmd/connectorgen`, source-lock/Atlas checks, static/docs checks,
  and read-only review. No provider, credential, database, or checked-in
  materialization path is permitted.

### 2026-09-04 — CP11 final repair actual RED/GREEN (instructions 130–131)

- **F1 RED:** `TestVNextGenerationPublisherRefusesReplacedAtomicControlTemporary` replaced each fsynced temporary `CURRENT`/`JOURNAL` entry and observed `<nil>` before the fix. **GREEN:** it now passes with the original temporary descriptor held through `renameBound`; prior control bytes, moved original, and replacement survive refusal.
- **F2 RED:** the three post-validation stage, generation/lease, and committed-control cleanup tests observed `<nil>` from Recover, Prune, and Publish (1.494s). **GREEN:** retained directory/control descriptors plus identity checks make all three pass (1.525s); `TestVNextGenerationPublisherRefusesReplacedRollbackGenerationCleanup` separately passes the rollback path (1.234s).
- **F3 RED:** temporarily allowing an existing lock to create a missing anchor made `TestVNextPublicationOpenLockRefusesExistingLockWithoutAnchor` fail with a successful open. **GREEN:** existing unanchored state refuses; `TestVNextGenerationPublisherRefusesReplacedLockAfterAcquisition` covers the held first inode, removed anchor, visible replacement, second-operation refusal, restored original hard-link pair, unchanged closed transaction, and successful serial retry.
- **F4 RED:** temporarily restoring predecessor mapping behavior made `TestVNextPublicationProofContractRejectsCompoundMapping` fail because a two-guarantee mapping was accepted. **GREEN:** the source-lock-only validator now requires exactly one guarantee. `TestVNextGenerationPublisherPhysicallyPreflightsImplementedStagedCommand` and `TestVNextGenerationPublisherRefusesPhysicallyStagedCommandPreflight` prove actual physical `commandrunner.Preflight` success/refusal for an implemented command after semantic admission.
- Focused GREEN commands so far: F1 (1.219s), F2 cleanup trio (1.525s), F2 rollback (1.234s), F3 replacement (1.220s), F3 missing anchor (1.005s), F4 mapping/physical witnesses (1.268s), exact lease/Atlas witnesses (1.711s), and `make connectorgen-vnext-locks` (31.096s). Final normal/race package, vet, definition, docs, contract, tidy, and review evidence is recorded in `VERIFICATION.md`.

### 2026-09-04 — CP11 exact-review continuation plan (instruction 132)

- Authority/base: Firstmate BLOCKed `d3661661dbd1646376e0fbae6d73ab658532a153` and authorizes only the report's F1–F4 repairs over that candidate. The report and `132.msg` were read in full before this plan. CP12, source-lock/materialized-output changes, provider/credential/database I/O, `.cache`, and certification residue remain out of scope.
- RED 1: a new post-identity/pre-rename fault point replaces a private control source for `CURRENT` and `JOURNAL`; the old control, moved original, and replacement must all survive the refusal.
- GREEN 1: actual rename source is a retained private-directory child, and its identity is rechecked at the exact late test boundary before descriptor-relative `renameat`.
- RED 2: late barriers replace a marker-proven stage, integrity-proven generation, lease-only member, rollback generation, `CURRENT`, and `JOURNAL` after validation but before quarantine movement. Existing pathname removal must expose the check-then-unlink/tree gap.
- GREEN 2: move cleanup candidates into a private quarantine directory, prove moved root plus bound child identities, restore a mismatched moved replacement, then delete only the verified quarantined tree.
- RED 3: while the first publisher holds its original lock, install a matching `.connectorgen.lock` plus anchor replacement pair and start a second publisher. The old file-pair lock domain permits independent acquisition.
- GREEN 3: publication `Flock` is held on a separately owned descriptor for the retained connector directory inode; the sibling pair cannot affect the domain. Verify no second mutation, normal first completion, and recovery/retry.
- RED 4: Atlas mappings that cite in-memory closed-set, unsynced marker, whole-generation-only, Journal-only, or missing-anchor-only tests overclaim their listed physical/durable behavior.
- GREEN 4: map every source-lock-vNext publication guarantee to one actual behavior-appropriate positive/refusal witness, including physical closed publication, durable marker, lease-only, `CURRENT`, and matched-pair cases.

### 2026-09-04 — CP11 exact-review continuation actual RED/GREEN (instruction 132)

- **F1 RED:** an intentionally exposed post-identity/pre-rename boundary left the predecessor's root-path source vulnerable: a replacement changed the prior `CURRENT` or `JOURNAL` control. **GREEN:** `TestVNextGenerationPublisherRefusesLateReplacedAtomicControlTemporary` now drives the exact late boundary for both controls and proves old control bytes, original private child, and replacement remain after identity refusal.
- **F2 RED:** temporary late-bound probes showed stage cleanup lost its bound-path proof after movement; generation, lease, `CURRENT`, and `JOURNAL` removal accepted a substituted object; rollback surfaced active validation rather than the identity defect. **GREEN:** `TestVNextGenerationPublisherRefusesLateReplacedValidatedStageCleanup`, `...ValidatedGenerationCleanup`, `...GenerationLeaseCleanup`, `...RollbackGenerationCleanup`, and `...ControlCleanup` all pass after private quarantine movement and revalidation.
- **F3 RED:** `TestVNextGenerationPublisherSerializesMatchedLockAnchorReplacement` initially observed the replacement sibling pair admit a second publisher while the first operation remained active. **GREEN:** the retained connector-directory inode is the sole Flock domain; the test now blocks the second mutation, completes the first publication, then recovers and retries serially.
- **F4 manual RED:** review of the old mappings showed an in-memory comparator, unsynced fixture marker, whole-generation-only witness, Journal-only witness, and missing-anchor scenario claimed stronger physical/durable behavior than executed. **GREEN:** `TestVNextGenerationPublisherCheckRefusesPhysicalClosedSetMutation`, `TestVNextGenerationPublisherRecoversPublisherWrittenDurableStage`, `TestVNextGenerationPublisherRefusesPruneWithInvalidLease`, and `TestFoundationAtlasSelectorsResolve` pass with one truthful mapping per `authoring.source-lock-vnext.v1` guarantee.
- Complete normal and race package GREEN: `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (61.981s), and `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (329.199s). These runs include the exact new witnesses and preserve the declared no-provider/no-credential boundary.

### 2026-09-04 — CP11 final F1 continuation plan (instruction 133)

- Authority/base: Firstmate BLOCKed `958a07a778fba6264d1aec567efa5d8c853eefa2` and authorizes only F1 plus its dependent source-lock-vNext Atlas refusal witness. F2/F3 are accepted; CP12, provider/credential/database I/O, `.cache`, certification residue, reset, rebase, and broad Atlas work remain prohibited.
- RED: a fault barrier after the final source identity check and immediately before namespace installation swaps the private child for `CURRENT` and `JOURNAL`. The predecessor uses that mutable pathname in `renameat`, returns an identity error only after overwriting the prior control, and therefore fails the preservation assertion.
- GREEN: retain a verified private backup of the prior control, install, verify the target inode, hard-link any mismatched installed replacement into quarantine, and atomically restore the prior control. Assert prior bytes, moved original temporary, and replacement for both controls.
- F4 GREEN: map only the F1 publication-guarantee refusal to the final-boundary test; `TestFoundationAtlasSelectorsResolve` must resolve the revised single mapping.

### 2026-09-04 — CP11 final F1 continuation actual RED/GREEN (instruction 133)

- **F1 RED:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherRefusesFinalReplacedAtomicControlTemporary$'` failed for `CURRENT` and `JOURNAL`. The exact post-final-validation barrier moved the original private `control` aside, installed `"final unrelated replacement"` at its path, and the predecessor's `renameat` replaced each old public control before discovering the wrong installed inode.
- **F1 GREEN:** `writeAtomicLocked` first hard-links a verified prior control into a retained private quarantine. After the final source validation/fault barrier, it installs, verifies the target inode, hard-links the mismatched installed replacement into quarantine, and atomically restores the prior inode. `TestVNextGenerationPublisherRefusesFinalReplacedAtomicControlTemporary` passed in 1.568s for both controls and asserts prior bytes/inode, moved original temporary, and quarantined replacement.
- **F4 GREEN:** `TestFoundationAtlasSelectorsResolve` passed in the 5.657s focused suite after replacing only the F1 mapping negative witness with the final-boundary test.

- **Complete GREEN:** final `go test -count=1 -timeout 20m ./cmd/connectorgen` passed in 65.113s and final `go test -race -count=1 -timeout 20m ./cmd/connectorgen` passed in 328.540s after the current F1 implementation and witness were in place.

### 2026-09-04 — CP11 durable F1 transaction/cut matrix (instruction 134, before production edits)

| Path | Durable phase/cut | Public control allowed at crash | Required restart resolution before parsing `JOURNAL`/`CURRENT` |
| --- | --- | --- | --- |
| Existing `CURRENT` or `JOURNAL` | private prior backup created, before its directory sync | prior | Ignore the incomplete private copy; ordinary control remains authoritative. |
| Existing `CURRENT` or `JOURNAL` | prior backup fsynced; typed repair authority `prepared` fsynced in connector namespace, before exposure | prior | Read repair authority first, prove backup/prior identities, retain or restore prior, then clear authority durably. |
| Existing `CURRENT` or `JOURNAL` | post-install, before target namespace sync or installed-identity decision | prior, intended candidate, or substitution | Pending authority makes the public name non-authoritative; retain any non-prior object for forensics and restore the recorded prior. |
| Existing `CURRENT` or `JOURNAL` | intended target namespace synced, before authority clear | intended candidate | Pending authority still resolves deterministically to the recorded prior; old-or-new complete publication recovery remains valid only after authority resolution. |
| Existing `CURRENT` or `JOURNAL` | mismatch replacement retained and repair authority updated/fsynced, before restore | substitution | Never parse the substitute; preserve its bound identity in repair storage, restore the prior, sync the public namespace, then clear authority. |
| Existing `CURRENT` or `JOURNAL` | restored public control synced, before authority clear | prior | Verify the recorded prior public inode, preserve forensic replacement, then clear authority and sync its removal. |
| Existing `CURRENT` or `JOURNAL` | authority removal synced after durable resolution | prior or intended candidate | No repair authority remains; ordinary journal/current recovery may parse only the durable resolved control. |
| First-publication/no-prior `CURRENT` or `JOURNAL` | authority `prepared`, install, mismatch, retention, restore, and clear cuts | absent, intended candidate, or substitution | Pending authority makes the target non-authoritative; retain any installed object, restore the valid absent state, sync it, then clear authority before ordinary recovery. |

- Repair authority contract: a strict bounded typed record under the connector control namespace binds version, target (`CURRENT` or `JOURNAL`), phase, repair-directory name/identity, intended candidate identity, optional prior identity, and optional observed replacement identity. It is persisted with the verified prior backup before the source can be exposed; recovery loads it before ordinary control decoding.
- Phases: `prepared` means backup/authority is durable and public target is untrusted; `replacement_retained` means the observed substitute is durably retained; `restored` means the prior or no-prior public namespace is fsynced. Authority removal follows only a durable resolved namespace.
- RED: current synchronous code has no authority record or cut hooks. A restart after a source-substituted install can parse a durable substitute or malformed `JOURNAL`/`CURRENT`; the current test only observes a normal handler return.
- GREEN: each matrix cut uses test-only fault instrumentation, a fresh publisher instance, and real local filesystem state to prove prior valid control or valid absent first-publication state; it also proves an observed substitute is retained and no repair authority remains after recovery.

### 2026-09-04 — CP11 durable F1 actual RED/GREEN (instruction 134)

- **Lifecycle/manual fallback:** the resolved `execute-phase`, `verify-work`, and `code-review` prompts require an inline/manual fallback because Firstmate prohibits role spawning and no compatible authorized worker/reviewer session exists. No GSD role was spawned. `golang-how-to` remains unavailable; loaded skills are recorded in `PLAN.md`.
- **F1 RED:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherRecoversInterruptedReplacedControlBeforeDecode$'` failed for both controls. The fault after raw substituted installation left the public control as `"interrupted substituted control"`; fresh recovery decoded it directly and failed with `decode CURRENT` or `decode publication journal`.
- **F1 GREEN:** a strict bounded `.connectorgen-control-repair.json` record is fsynced only after the private prior hard link and quarantine directory are fsynced. It binds the target, intended inode, optional prior, quarantine inode/name, repair phase, and an observed replacement when one is retained. `recoverLocked` invokes repair recovery before any ordinary `JOURNAL` or `CURRENT` read.
- **Durable-cut GREEN:** `TestVNextGenerationPublisherRecoversEveryControlRepairDurableCut` exercises backup, prepared-authority, raw-install, installed-namespace-sync, and cleared-authority cuts for existing and no-prior `CURRENT`/`JOURNAL`. `TestVNextGenerationPublisherRecoversEveryRetainedReplacementCut` additionally exercises replacement-link sync, replacement-authority sync, public-restoration sync, restored-authority sync, and clear sync for the same four paths. Each uses a fresh publisher, checks pending/absent authority at the cut, then proves valid prior/first-publication resolution and cleared authority.
- **Substitute GREEN:** `TestVNextGenerationPublisherRecoversInterruptedReplacedControlBeforeDecode` retains the malformed substitute as the typed `replacement` member and succeeds only because restart recovery resolves authority before ordinary decoding. Replacement-cut tests use syntactically valid but nonexistent-generation `CURRENT`/`JOURNAL` substitutes so a syntactically valid substitute cannot become selection authority either.
- **F4 GREEN:** the changed source-lock-vNext guarantee is mapped only to `TestVNextGenerationPublisherRecoversEveryControlRepairDurableCut` and `TestVNextGenerationPublisherRecoversInterruptedReplacedControlBeforeDecode`, both restart-capable durable witnesses. The former synchronous final-boundary test remains registered as a regression but is no longer the mapping witness.

### 2026-09-04 — CP11 monotonic-authority redesign pre-inspection route (instruction 136)

- **Skills before Go inspection:** `go-engineering` plus fundamentals, advanced concurrency/strict serialization, and production safety/security/error/testing references; `diagnose`; and `tdd`. Required `golang-how-to` was attempted and is unavailable. Routing and GSD adapter references were read; no role was spawned because Firstmate prohibits it.
- **RED plan:** prove with the existing temporary-root publisher that a cooperative writer lock does not stop a same-permission external `rename`/`unlink` from changing a post-exposure authority pathname. The control test must restart a fresh publisher and distinguish a verified immutable authority/phase chain from a missing, replaced, or redirected authority.
- **GREEN plan:** one vertical slice first establishes immutable prepared authority before exposure and recovers it before public control decoding; successive slices establish create-only phase persistence, source/target substitution, no-prior state, replacement retention, cleanup identity, and cooperating-writer serialization.
- **Counterfactual:** if the earlier workspace/connection-ID implementation has no concrete private namespace allocation/identity invariant reusable under descriptor confinement, do not generalize from it or add manual-unlock behavior; retain the smallest connector-local immutable protocol proven by the divergence tests.

### 2026-09-04 — CP11 monotonic-authority investigation and RED (instruction 136)

- **Exact workflow route:** `scripts/gsd doctor` returned exit 1 solely for the known missing `.gsd/prompts/issue-122-rebootstrap.md`; all other adapter checks passed. `scripts/gsd sources {discuss-phase,plan-phase,execute-phase,verify-work,code-review}` resolved the same project-local registry, lock, and official command document. The corresponding prompts were generated. Firstmate forbids their role spawning, so the authorized path is inline/manual execution with this ledger as the lifecycle evidence; it does not replace the required external exact-SHA review.
- **Design facts:** `writeAtomicLocked` creates the old repair record before final control installation; `recoverLocked` consults repair before ordinary controls. In candidate `f36b5d0a`, however, `updateControlRepairLocked` atomically replaces `.connectorgen-control-repair.json` after public exposure. `acquireOperation`/`vNextPublicationAssertLockBound` bind the cooperative lock to the retained connector directory only.
- **Reusable identity facts:** `internal/app.(*App).migrateWarehouseIdentity`, `allocateUniqueIdentity`, `warehouse.LocationFor`, and `warehouse.Location.EnsureOwnership` implement opaque unique structural ownership and reject identity mismatch instead of rewriting it. Commit `fcff76a7305fe469c3903c33a89bc47912852ac6` carries the cross-connection loss RED and connection-isolation GREEN; its direct proofs are `TestSecondConnectionDoesNotDestroyFirstConnectionRows`, `TestConnectionIdentityIsOpaqueAndNotDerivedFromNameOrCredential`, `TestSafePathPartRejectsRatherThanRewrites`, and `TestEnsureOwnershipRefusesAnotherConnectionsDirectory`.
- **RED, observed:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherPrivatePreparedAuthoritySurvivesPublicControlSubstitution$'` failed `CURRENT` and `JOURNAL` source/target rename cases while holding the real publisher lock. Every failure reported `private prepared recovery authorities = 0, want one`; `os.Rename` succeeded, proving that the cooperative lock did not prohibit the external same-permission mutation.
- **Selected GREEN criterion:** use a dedicated connector-local private transaction directory with the same random `Mkdirat(..., O_EXCL)` descriptor-allocation pattern as `vNextPublicationCreateQuarantine`, but without reusing that quarantine object. Create immutable `prepared.json`, then create-only identity/digest-predecessor-bound phases. The latest verified phase—not a rewrite of a root authority pathname—drives recovery before public decode. A phase, prepared record, namespace, source, or target identity mismatch must fail before public control trust.

### 2026-09-04 — CP11 monotonic-authority implementation RED/GREEN (instruction 136)

- **Additional RED, observed during review:** after the initial private-authority implementation, `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherRefusesSubstitutedPreparedAuthorityBeforePublicControlExposure$'` failed with `prepared-authority substitution installed a public control before refusal`. The new test renamed/recreated `prepared.json` at the real final-source barrier; predecessor code detected it only during post-install resolution. This is a new executable RED, not a relabeling of an earlier test.
- **Immediate GREEN:** `writeAtomicLocked` now reasserts the prepared transaction and immutable authority identity after the final source barrier and immediately before `CURRENT`/`JOURNAL` installation. The same command passed: the original public control inode remains installed, the moved original authority remains retained, and fresh recovery refuses the private transaction before ordinary `CURRENT` decoding.

- **Additional durable-clear RED, observed during review:** after adding the post-prepared-authority-sync crash cut to the existing normal and retained-replacement matrices, `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRecoversEveryControlRepairDurableCut|TestVNextGenerationPublisherRecoversEveryRetainedReplacementCut)$'` failed all eight new cases: normal writes returned `<nil>` and retained substitutions returned the prior mismatch refusal instead of the injected crash. The predecessor deleted phase/backup material before the sole prepared authority and had no durable cut at authority retirement.
- **Durable-clear GREEN:** `clearControlRepairLocked` now deletes and fsyncs the prepared authority first, then exposes the crash cut before it removes only private phase/backup garbage. A crash after that sync has a verified public old/new-or-absent control and no pending authority; a crash before it retains prepared plus every required restoration member. The same matrix passed in 12.632s.

### 2026-09-04 — CP11 F1 Design B linearization plan (instruction 139, before production edits)

- **Exact subject:** immutable candidate `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8`, immediate parent `f36b5d0a275ed27fd5f4da242ba192e43f8066d5`, report `/Users/karthiksivadas/karthik-agent-workspace/data/cli-batch1-cp11-f1-linearization-research-r1/report.md`. The defect is no linearization point between mutable public `CURRENT`/`JOURNAL` and later retirement of sole private prepared authority.
- **Mask/symptom:** cooperative `Flock` writers hide direct pathname interference; old hooks run before a later validation; and public-only `--check` masks a no-prior pending `JOURNAL` behind old valid `CURRENT`. Candidate can destroy a late public inode or permit authority-free third-state public trust.
- **RED A:** against candidate, real descriptor-relative operations plus a direct lock-ignoring actor replace/unlink `CURRENT` and `JOURNAL` after final install/restore/no-prior identity validation and before plain public `renameat`/`unlinkat`. Assert substitute destruction or missing forensic identity, not only error status.
- **RED B:** after final public validation and before candidate `prepared.json` removal/fsync, mutate the public target, restart a fresh publisher, and invoke real `lock-render --check`; prove authority-free public trust, including no-prior `JOURNAL` plus old valid `CURRENT` false success.
- **GREEN:** protected mode retains one terminal head per target. A durable successor binds a predecessor terminal; strict identity/digest-bound capture-intent, captured, selected, and terminal records append with `O_EXCL`; no-replace capture preserves actual public occupant; create-only link selects presence; absence never invokes public unlink; four attempts terminalize `retry_required` with typed error and retained evidence.
- **Matrix:** `CURRENT`/`JOURNAL`; prior present/absent; install/restore/logical absence; rename/unlink/second+third replacement; every capture/selection/terminal/successor-cleanup cut; fresh recovery; read-only real check snapshot; malformed graph; unsupported/collision no-replace error; cleanup substitution; cooperating writers; race detector.

### 2026-09-04 — CP11 F1 Design B candidate RED observations

- **Post-terminal replacement RED:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherCheckRefusesTerminalAuthorityFreeReplacement|TestVNextGenerationPublisherCheckRefusesPendingPrivateAuthority)$'` failed `TestVNextGenerationPublisherCheckRefusesTerminalAuthorityFreeReplacement`: a direct actor renamed and recreated byte-valid `CURRENT` after candidate final validation but before prepared-authority retirement; fresh `Check()` returned `<nil>`, accepting an authority-free third inode.
- **Pending no-prior JOURNAL RED:** the same command failed `TestVNextGenerationPublisherCheckRefusesPendingPrivateAuthority`: an injected crash after durable private no-prior `JOURNAL` preparation left old valid `CURRENT`; fresh candidate `Check()` returned `<nil>`. This is the production-check masking defect, not a synthetic checker failure.

### 2026-09-04 — CP11 F1 Design B actual RED/GREEN (instruction 139)

- **Red:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherCheckRefusesTerminalAuthorityFreeReplacement|TestVNextGenerationPublisherCheckRefusesPendingPrivateAuthority)$'` failed against `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8`: `Check` accepted a valid terminal-authority-free replacement and accepted an old `CURRENT` while a no-prior private `JOURNAL` repair was pending.
- **Green — terminal authority:** `TestVNextGenerationPublisherTerminalAuthorityTransitionMatrix` covers `CURRENT` and `JOURNAL`, prior-present and prior-absent installs/restores/removals, including first `CURRENT`, first and replacement `JOURNAL`, and logical absence. `TestVNextGenerationPublisherRecoversEveryTerminalAuthorityDurableCut` restarts fresh publishers at all eleven reachable pre/post-durability cuts for both targets and both prior states.
- **Green — adversarial namespace behavior:** real descriptor-relative tests prove no-replace error handling, source absence, late capture, removal-without-public-unlink, A/B/C repeated substitutions, malformed/fork/gap refusal, private transaction/prepared/phase/capture/predecessor replacement refusal, terminal divergence recovery, retained successor/predecessor authority, and cooperating-writer serialization.
- **Green — read-only production path:** `TestRunLockRenderCheckReadsAuthorizedTerminalAuthorityWithoutWriting`, `TestRunLockRenderCheckRefusesPendingPrivateAuthorityWithoutWriting`, and `TestRunLockRenderCheckRefusesDivergentTerminalAuthorityWithoutWriting` exercise actual `lock-render --check`, assert the tree is unchanged, and require no success stdout for pending/divergent authority.
- **Review correction:** the first Design B reader scanned private state then read a public control without revalidating private identities. `TestVNextGenerationPublisherCheckRevalidatesPrivateAuthorityBeforePublicDecode` now substitutes the private transaction at the real post-scan boundary and proves shared `Check` refuses before malformed `CURRENT` decoding.
- **Race correction:** an initial race run exposed a test assertion that overfit whether the graph scanner or predecessor validator observed a substituted predecessor first. The test now requires either private-authority refusal form and forbids public decoding; focused `-race` and the final full race suite pass.

### 2026-09-05 — CP11 Astra B-01/B-02 correction TDD plan

- Authority/base: Firstmate authorizes one local CP11 correction wave from frozen candidate `8214bd91403ce620773b61caf674faa540ee1701`; B-01 and B-02 are the complete blocking ledger in `data/cli-batch1-cp11-astra-review-r2/report.md`. CP12 remains prohibited. No source lock, rendered execution artifact, runtime route, provider, credential, database, `.cache`, certification residue, release action, reset, rebase, or push is in scope.
- Test instrumentation is limited to the existing nil-in-production `vNextPublicationHooks` fault boundary mechanism. It is needed because the two reported syscall/durability seams have no existing deterministic pause; it changes no production behavior with `At == nil` and exercises the real descriptor-relative implementation rather than a helper substitute.
- RED B-01 cleanup restoration: create A, move it to a retained path, replace its public name with regular B at the existing post-identity/pre-quarantine seam, then after B is quarantined and the production restore path has observed absence create regular C at the new immediately-before-restore seam. The candidate's plain restore rename must overwrite C, so the test fails its assertions that C remains public, B remains quarantined, A remains reachable, and an identity/conflict refusal is returned. Run the same assertion through stale-stage recovery, stale-generation pruning, and active-validation rollback.
- RED B-01 activation: at `before_stage_rename`, create an empty final-generation destination at the exact content-addressed target. The candidate's plain activation rename replaces it, so the test fails its assertions that the collision directory and staged directory both remain reachable and activation refuses.
- RED B-02 bootstrap: inject an interruption after base prepared authority plus both directory syncs but before base terminal append, once for `CURRENT` and once after a terminal `CURRENT` while `JOURNAL` is pending. The candidate fresh `Recover` must fail at markerless pending authority. Before that recovery, actual `lock-render --check` must remain nonzero, emit no success output, and leave the tree byte-identical.
- GREEN B-01: route restoration and activation through `renameNoReplaceFrom`; preserve collision errors, C at the public name, and B in private quarantine. There is no precheck-based substitute, overwrite fallback, or optional cleanup.
- GREEN B-02: under the existing exclusive lock, recognize only a phase-empty, predecessor-free, strict/anchor-validated base state whose prior logically equals intended. Append and fsync its committed terminal phase, re-scan the complete graph, create any missing base head, and write the marker only when both terminal heads exist. Then ordinary authority-first reconciliation handles public divergence. Missing prepared authority, malformed JSON, successors, forks, cycles, gaps, non-base phases, unequal prior/intended, and invalid anchors remain refusals.
- Falsifying controls: a B-01 candidate test that leaves C public without code change disproves the restore-clobber finding; a B-02 candidate test whose strict malformed/missing-prepared cases recover disproves the proposed narrow resume predicate and blocks GREEN. Neither control is silently reclassified as a pass.
- Planned selector before RED: `go test -list 'TestVNextGenerationPublisher(RefusesSecond|RefusesFinalGeneration|ResumesInterrupted)' ./cmd/connectorgen`.
- Planned focused RED/GREEN command: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore|TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation)$'`.
- Planned continuity selector: existing malformed topology, terminal authority matrix, real `lock-render --check`, late cleanup, and durable-cut witnesses run with the new selectors; package/race and canon/Atlas/static checks are recorded only after observed execution.


### 2026-09-05 — CP11 Astra B-01/B-02 actual RED

- **Red:** after adding only nil-in-production `vNextPublicationHooks` seams at the reported restore and post-durable-base-prepared cuts, `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore|TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation)$'` exited 1 in 3.250s.
- **B-01 RED:** real stale-stage recovery, stale-generation prune, and failed-active-validation rollback each moved validated A aside, quarantined regular B, then installed regular C after the production absence observation. Candidate plain restore replaced C: every subtest observed public bytes `"second public replacement"` instead of C's `"third public replacement"`. The real `before_stage_rename` collision control returned `<nil>`, proving plain activation replaced the late empty destination.
- **B-02 RED:** the new post-`prepared.json`/transaction-sync/connector-sync pre-terminal seam interrupted both bootstrap cuts: CURRENT as the first base and JOURNAL after a terminal CURRENT. In both cases actual `lock-render --check` first returned nonzero with zero success stdout and an unchanged tree snapshot; fresh ordinary `Recover` then refused the strict durable state with `publication control authority marker is missing from pending repair`.
- **Green boundary:** no B-01 destination primitive or bootstrap recovery behavior has changed yet. Green must make the same command pass while retaining C public/B quarantined/A reachable, retaining both activation inodes, terminalizing only the strict phase-empty base state, and preserving all existing malformed/missing-prepared/successor/topology refusals.

### 2026-09-05 — CP11 Astra B-01/B-02 actual GREEN

- **Green:** `restoreQuarantinedLocked` and `activateStageLocked` now call the existing descriptor-relative `renameNoReplaceFrom` primitive. There is no post-check overwrite fallback. A restoration collision leaves C public and B in the private quarantine; `refuseQuarantinedReplacementLocked` wraps both the identity refusal and the typed no-replace cause, so the witnesses require `errors.Is(err, fs.ErrExist)`.
- **B-01 GREEN:** `TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore` passes its real stale-stage recovery, stale-generation prune, and failed-active-validation rollback subtests. Each proves moved A, regular B, and regular C retain their expected directory/file identity and bytes. `TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision` passes with a typed collision while retaining both staged and final-destination directory inodes.
- **B-02 GREEN:** `ensureControlAuthorityLocked` re-scans under the existing exclusive lock. It resumes exactly one phase-empty/no-predecessor base record with equal prior/intended logical state only after private identity validation, appends its committed terminal, then re-scans to create a missing base head and marker. The CURRENT-first and JOURNAL-second fresh-publisher witnesses now terminalize both heads, preserve recorded absence, and finish a normal retry; their pre-recovery `lock-render --check` snapshot remains non-mutating and nonzero.
- **Atlas/canon Green:** the old one-substitution proof mappings were not exact enough for post-absence C or pre-terminal bootstrap recovery. `SOURCE-LOCK-VNEXT.md` and the existing `authoring.source-lock-vnext.v1` record now bind only the changed no-replace final-activation, A/B/C cleanup, and strict base-recovery guarantees to the new registered physical/durable witnesses. No new foundation or runtime path was added.
- **Observed focused Green:** the three correction witnesses plus `TestFoundationAtlasSelectorsResolve` passed in 5.461s; the continuity selector passed in 47.655s; the planned focused race selector passed in 36.321s. At this checkpoint the full package/race and static/canon gates were intentionally deferred to the final verification section below.

### 2026-09-05 — CP11 Astra B-01/B-02 final local verification

- **Complete package Green:** `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 157.507s; `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 462.893s.
- **Authoring/canon Green:** `make connectorgen-vnext-locks` → `ok`; `make connector-canon-check` → `connector canon check: ok`; `go run ./cmd/connectorgen validate internal/connectors/defs` checked 553 connectors with 0 findings; catalog JSON parsed and the Atlas selector proof passed.
- **Static/smoke Green:** `go vet ./cmd/connectorgen`, `go build -o /dev/null ./cmd/connectorgen`, `go mod tidy -diff`, `go run ./cmd/agentcontractgen check`, `git diff --check`, and `go run ./cmd/connectorgen --help` passed. `make docs-check` validated `docs/connectors`; its generated local `pm` binary was removed afterward.
- **Boundary scan:** the explicit changed-path literal-secret scan returned no matches. An earlier broad identifier scan found only the local directory-name variable `token`, not a value or credential. `.cache/` and `internal/connectors/certifications/` were never read, modified, staged, or scanned.
- **Delivery state:** freeze one local normal commit, record its exact SHA in Firstmate status, and request the required independent Astra review. Do not push or start CP12 pending disposition.

### 2026-09-05 — CP11-R3-01 audited FIFO reader correction

- **Finding and causal boundary:** independent publication and bootstrap Astra reviews plus an independent merged-ledger audit agree that `vNextPublicationDirectory.openRegular` opens a FIFO with blocking `O_RDONLY` before checking `Stat().Mode().IsRegular()`. The same final-member exposure also exists at `vNextPublicationDirectoryFS.Open`, which passes raw `openFile` to semantic-admission `io/fs` consumers. This is one Medium root with multiple callers; B-01 and B-02 are not reopened, and L-01 remains excluded.
- **Red:** before any production change, add only temporary-root subprocess instrumentation. A no-writer FIFO at the actual stage marker/public control/final delegated admission member must make the child fail to finish inside its bounded parent deadline. The parent asserts the candidate did not return a refusal or success output and that no fixture reader consumed FIFO bytes. The RED record must name the child selector, timeout, exact observed block, and caller reached; a plain unit test that risks blocking the suite is invalid evidence.
- **Green:** add the minimum descriptor-local read contract: no-follow, nonblocking open; same-opened-descriptor allowed-type validation; raw `io/fs` exposure only after that validation, with a directory retained only where nested schema traversal needs it. Retain existing regular control/artifact behavior, no-create read semantics, O_EXCL writer behavior, lease locking, `fs.ErrNotExist` and symlink diagnostics, and byte-limit validation.
- **Green assertions:** each former child returns promptly with a nonregular-member refusal and no success line; FIFO metadata and active generation/control identities remain unchanged; a subsequent operation acquires the publication lock; valid regular members and nested schema directories still load; missing/symlink/closed-tree paths remain rejected. The semantic-admission test performs a finite regular-to-FIFO substitution after enumeration and before the actual adapter `Open`, proving the final descriptor boundary rather than a pathname precheck.
- **Refactor stop rule:** no new helper package, runtime path, generic filesystem interface, public flag, dependency, daemon, provider operation, or CP12 work. Any discovery that a platform cannot provide the required nonblocking descriptor semantics is a new architectural decision for Firstmate, not an overwrite/fallback.

### 2026-09-05 — CP11-R3-01 actual RED

- **Command:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextPublicationFIFOReaderRefusesBeforeBlockingOpen$'` exited 1 in 2.932s.
- **Observed stale-stage path:** the subprocess published a valid temporary generation, replaced only the stale stage ownership marker with a no-writer FIFO, and entered production `Recover`. The parent deadline expired after 1.00s with only `=== RUN` from the child; it never reached the expected regular-file refusal. This is the real `removeStagesLocked` → stage-owner read path, not a helper substitute.
- **Observed admission adapter path:** the subprocess enumerated a regular staged filesystem through `vNextPublicationStageFS`, replaced the enumerated `spec.json` with a no-writer FIFO, then called the real delegated `fs.ReadFile`. The parent deadline likewise expired after 1.01s before a refusal. This demonstrates the `vNextPublicationDirectoryFS.Open` bypass after enumeration.
- **Interpretation:** both failures establish the same `Openat(O_RDONLY)`-before-type-check root. No FIFO bytes were read, no provider/credential/database path was reached, and the parent killed only its disposable child process. GREEN must make the exact selector return promptly with descriptor-validated nonregular-file errors and preserve the FIFO.

### 2026-09-05 — CP11-R3-01 focused GREEN

- **Implementation:** `openRegular` now adds `O_NONBLOCK` to its existing no-follow descriptor open before `file.Stat()` checks for a regular file. `vNextPublicationDirectoryFS.Open` no longer exposes raw `openFile`: its new descriptor-local member open is no-follow/nonblocking and returns only a regular file or a real directory, preserving nested schema traversal while refusing FIFO/device members before `io/fs` can read them. Create/exclusive writers and lease call sites are unchanged.
- **Command:** after `gofmt`, the same focused selector `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextPublicationFIFOReaderRefusesBeforeBlockingOpen$'` passed in 2.129s.
- **Observed GREEN:** all four bounded children returned promptly. Stale-stage `Recover` and public `CURRENT` recovery return a regular-file refusal while leaving the FIFO visible to `Lstat` and releasing the exclusive operation lock. Public `JOURNAL` through the real `lock-render --check` returns code 1 with empty success output and releases the lock; restoring the original inode (or absent journal) permits a succeeding check. The staged filesystem enumerates and reads a valid nested schema directory, then rejects an enumerated-to-FIFO `spec.json` at the delegated adapter open without consuming FIFO bytes.
- **Execution workflow record:** the exact authority for the inline execution is `.agents/agentic-delivery/references/gsd-pi-adapter.md`, **Agent requirements**, paragraphs beginning “Non-interactive or non-Pi runners must use `scripts/gsd prompt <command>`” and “Execute generated prompts inline when compatible isolated runtime agents are unavailable or the canonical single-worker contract forbids spawning roles.” `node scripts/gsd prompt execute-phase 4427` was generated before this GREEN record and its TDD execution obligation was performed inline in this ledger. The custom phase has no `.planning/ROADMAP.md` for the Pi executor and no compatible project-local Pi worker; this is not a claim that `/gsd-execute-phase` ran automatically. `node scripts/gsd prompt verify-work 4427` and `node scripts/gsd prompt code-review 4427` are generated inputs only at this point; their separately recorded inline verification and fresh independent review remain required after the remaining local gates.
- **Continuity GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextPublicationFIFOReaderRefusesBeforeBlockingOpen|TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore|TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation|TestRunLockRenderCheckReadsAuthorizedTerminalAuthorityWithoutWriting|TestRunLockRenderCheckRefusesPendingPrivateAuthorityWithoutWriting|TestRunLockRenderCheckRefusesDivergentTerminalAuthorityWithoutWriting|TestVNextGenerationPublisherBootstrapsRetainedTerminalAuthorities)$'` passed in 6.280s.
- **Race harness correction and GREEN:** the first race run showed each child test passed in 0.38–0.49s but race-runtime process shutdown exceeded the parent’s one-second deadline. This did not alter the product outcome; the parent timeout alone was widened to five seconds so a defect remains bounded and a passing race child can exit. The same race selector then passed in 12.619s, and the normal selector rerun passed in 6.278s. No production retry, timeout, or behavior changed.
- **Complete changed-package GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen` passed in 129.210s. `go test -race -count=1 -timeout 20m ./cmd/connectorgen` passed in 407.953s. These runs include the full CP11 publication, cleanup, authority, durability, check-only, and new FIFO matrix; they do not represent provider-live, database, Linux filesystem-runtime, or full-repository certification.

## 2026-09-06 — CP11 R3-02–07 coordinated repair TDD ledger (pre-production)

The immutable candidate is `3a455877cdd9686ba6f04341960a3c31196909bd`; `c6194254560ff874ac63e69a6c80dfe9ab06b5e2` is ledger-only. The following is planned RED/GREEN work, not an executed result.

| Finding | RED before production change | GREEN and controls | Status |
| --- | --- | --- | --- |
| R3-02 | Fresh child creates valid retained authority history, verifies effective `RLIMIT_NOFILE` after runtime setup, and demonstrates real `EMFILE` through Check/Open/read/release/Recover/Prune/publish; prepared-hook partial error leaks no state. | Same fixture/limit succeeds while authority identities/history survive; transaction/prepared/phase/capture/anchor/predecessor substitution and lock-reacquisition controls remain. | Planned |
| R3-06 | Nil-production barrier after final validation moves A and installs regular B before actual Open, demonstrating candidate B bytes. | Handle reads A bytes or refuses; A/B identities and closed membership are asserted. | Planned |
| R3-07 | Hold actual A reader; new publication plus moved A lease/replacement B lets real Prune and recovery/publish cleanup delete A or break its read. | Held A/bytes survive cleanup, B is not clobbered, and ordinary stale cleanup resumes after release/restoration. | Planned |
| R3-03 | Built-main SIGINT/SIGTERM/repeat subprocesses show global interception for non-consuming paths. | Non-consuming command has normal signal termination; existing held-lock render cancellation exit 1 without success/mutation and post-unlock retry are preservation controls, not a new RED absent a separately demonstrated defect. | Planned |
| R3-04 | Configured lint output lists all 31 introduced CP11 diagnostics; error-handling behavior obtains focused failure proof. | Zero introduced diagnostics without suppression/config weakening or hidden durability errors. | Planned |
| R3-05 | A FIFO read-trapping fixture proves snapshot code does not open/read the FIFO. | No-follow identities and permitted regular data prove real refusal/JOURNAL check and adapter-substitution preservation; wording names only acquired locks. | Planned evidence-only |

Retained controls: R3-01 nonblocking descriptor admission; B-01 A/B/C and activation no-clobber; B-02 strictly valid interrupted bootstrap and non-mutating malformed/private refusal. The four pre-range `internal/cli` failures and 20 pre-range lint diagnostics are not CP11 green evidence.

### 2026-09-06 — Group A actual RED against `3a455877cdd9686ba6f04341960a3c31196909bd`

- `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherOpenKeepsValidatedDirectory$'` exited 1 in 4.063s: after the nil-production post-validation barrier moved validated A and installed regular B, actual `Open` returned B's `metadata.json` (`{"marker":"replacement-B"}`), not A's bytes.
- `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherHeldGenerationUsesStableCleanupLock$'` exited 1 in 6.606s: real Prune, Publish cleanup, and crash/recover cleanup each removed held A after its empty companion lease was moved outside scanned generations and B was installed at the original lease path.
- `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherBoundsAuthorityScanDescriptors$'` exited 1 in 99.621s. Each fresh child first built valid retained history, then verified `RLIMIT_NOFILE=96`; Check, Open, Recover, Prune, and Publish respectively failed in 18.13–21.72s with `open publication control predecessor terminal: too many open files`. The parent and child process limits were isolated, and this is a real authority scan rather than a malformed fixture.
- These are the completed Group-A RED observations. The following production edits repair the shared descriptor/identity root; no Group-B signal classification is implied by this record.

### 2026-09-06 — CP11-R3-07-LATE-LEASE preserve/fix RED/GREEN

- **Adjudicated sibling and scope:** Firstmate accepted the independent Astra/xhigh preserve/fix judgment recorded at `data/cli-batch1-pi-takeover/cp11-late-lease-astra-xhigh-c6194254560ff874ac63e69a6c80dfe9ab06b5e2-21e53fd1a36d4cb5a523c7af70e0f0d64f72fe3a94d86b2896b1d631db5e868f.md` (report SHA-256 `87dffb471993c91a4c9946d218409e4ef328f6ff1f697a991e15c5d0847d1ade`). The uncommitted directory-lock repair had removed the old lease-member binding, its nil-production late barrier, and the mapped negative test. Stable generation-directory reader/cleanup locking remains required; it is not a substitute for post-validation member identity preservation.
- **Artifact history:** the audited complete state before this repair is HEAD `c6194254560ff874ac63e69a6c80dfe9ab06b5e2` plus eight-path diff `21e53fd1a36d4cb5a523c7af70e0f0d64f72fe3a94d86b2896b1d631db5e868f`; the two-file `68fe9dd363225003b8ec2081380a3b8b35afb1cda9014d3d3106ae824c27ca21` projection records only the test/hook/binding deletion. The independent disposition itself is committed before production as `a6da870a257539e9007bdc684c8471488143be17` with all source/test repair files kept unstaged.
- **RED (barrier restoration):** restoring only the mapped test and fault constant while the cleanup call still supplied no binding initially failed in `1.57s` because the barrier was not wired; that was instrumentation incompleteness, not the defect. Restoring the inert production hook argument while keeping the binding nil produced the real RED: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherRefusesLateReplacedGenerationLeaseCleanup$'` exited 1 in `2.485s`. Both nonempty and empty regular B replacements reached the post-member-validation/pre-quarantine barrier, `Prune()` returned nil, and the failure observed moved A, B, and the stale generation all absent. This is actual destructive object loss, not a hook-hit-only failure.
- **GREEN (independent member guard):** `removeGenerationLocked` retains its directory-descriptor exclusive Flock, then opens the existing `.lease` no-follow/regular descriptor without flocking it, captures its same-opened-descriptor identity, retains that one bounded descriptor through the attempt, and passes the binding to the existing pre/post-quarantine checks. Its close error now participates in the named result. The same selector passed in `2.479s` for both empty and nonempty B, proving an identity refusal and A/B inode/type/byte preservation with unchanged stale-generation root, active generation, and `CURRENT` inode/bytes.
- **No-clobber sibling GREEN:** `TestVNextGenerationPublisherLateLeaseReplacementRetainsPublicGenerationCollision` passed in `3.155s`. After the lease binding mismatch moved G into quarantine, its actual before-restore hook creates public C. The real `renameNoReplaceFrom` path returns a typed `fs.ErrExist` chain while retaining C public and the original G/A/B tree quarantined, with original/replacement inodes and bytes plus active generation/CURRENT identity asserted.
- **Caller matrix GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesLateReplacedGenerationLeaseCleanup|TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers|TestVNextGenerationPublisherLateLeaseReplacementRetainsPublicGenerationCollision|TestVNextGenerationPublisherRefusesLateReplacedRollbackGenerationLeaseCleanup)$'` passed in `7.245s`. The public matrix exercises explicit Prune, no-journal Recover, Publish initial recovery, Open transitive recovery, and committed/new-selected journal recovery; the rollback fixture separately exercises rejected-generation cleanup after failed active validation. Each late actor reaches the common post-binding/pre-quarantine sink and observes identity refusal without deleting A or B. Read-only Check remains intentionally outside this destructive matrix.
- **Atlas/canon:** `SOURCE-LOCK-VNEXT.md` and `authoring.source-lock-vnext.v1` now state distinct contracts: directory-inode locking protects reader lifetime, while the descriptor-bound lease member protects late cleanup integrity. The existing mapped negative proof name remains registered; it is not replaced by the held-reader test. Selector/canon execution remains an explicit final-wave gate.
- **Race GREEN:** `go test -race -count=1 -timeout 20m ./cmd/connectorgen` passed in `676.216s`. This is the complete changed-package race corpus after the independent lease-member repair; it remains local/static evidence only, not provider-live or database certification.

### 2026-09-06 — CP11 R3-02–07 final local verification

- **Behavioral GREEN:** the broader CP11 selector (Atlas proof lookup; descriptor scan/open/held-generation, late-member no-clobber/caller/rollback, signal, and FIFO witnesses) passed in `121.026s`; the complete normal changed package `go test -count=1 -timeout 20m ./cmd/connectorgen` passed in `253.295s`; and the complete race package passed in `676.216s`. These results cover the local CP11 repair matrix, not a provider-live, credential, database, full-repository, or final-programme claim.
- **Canonical/admission/docs GREEN:** `make connectorgen-vnext-locks` passed; `make connector-canon-check` reported `connector canon check: ok`; `go run ./cmd/connectorgen validate internal/connectors/defs` checked `553` connectors with `0 findings`; and `make docs-check` built `pm` then validated `docs/connectors`.
- **Static GREEN:** `gofmt -d` on every changed Go file, `go vet ./cmd/connectorgen`, `go build -o /dev/null ./cmd/connectorgen`, `go run ./cmd/agentcontractgen check`, `go mod tidy -diff`, and `git diff --check` passed. Configured `golangci-lint run ./cmd/connectorgen/...` still reports exactly `15` inherited pre-range diagnostics (one existing `staticcheck` and fourteen existing unused declarations); it reports no R3-02–07 location.
- **Boundary GREEN:** the single retained interactive `make connector-boundary` capture exited `0` with `ConnectorBoundaryReport.outcome=clean`, `checked_files=284`, `connectors_loaded=553`, `findings=[]`, `warnings=[]`, and six existing documented exceptions. A prior noninteractive boundary invocation had already produced its JSON, but its terminal capture collapsed before the final exit; the one interactive repeat exists solely to retain this definitive exit/result and did not follow a code change or unresolved finding.
- **Next gate:** inspect and freeze this coherent local repair as one code-bearing commit, then obtain the required fresh independent Astra/xhigh exact-SHA full-range review. No push, CP12 work, or acceptance follows these local checks.

## 2026-09-06 — CP11 F-01–F-08 coordinated TDD ledger

This ledger supersedes no earlier historical result.  Its RED/GREEN entries are
written before the coordinated source/test wave; no planned row below is an
executed result until the exact command, output, exit, duration, and tested
state are appended.

| Group | Red: observable candidate failure | Green: observable repaired behavior | State |
| --- | --- | --- | --- |
| F-04 | Bounded child schedules replacement after snapshot classification; current oracle can mix A metadata/B bytes, follow symlink, or block/read FIFO. | Opened descriptor identity/type is checked before bytes; regular→FIFO/symlink/directory swap either preserves A or refuses boundedly. | RED/GREEN recorded; package/full review pending |
| F-08 | Bounded held-lock/withheld-readiness child exposes unbounded Wait, sleep race, or child cleanup omission without a broad package hang. | Cleanup/reap/lock release runs from every early exit; deterministic contention readiness, actual signal exit, preserved state, and retry pass. | RED/GREEN recorded; package/full review pending |
| F-01 | Each real capture validator/candidate/mutating reopen admits or moves into B after A→B swap. | Checked-open identity refuses before dependent public mutation and valid continuity/reacquisition remains. | RED/GREEN recorded; package/full review pending |
| F-02 | Repeated nested cleanup substitution grows descriptor count/owned open descriptor; open Fstat error controls early exit. | Count/ownership remains bounded without finalizers; A/B/root/quarantine outcomes match recursive policy. | RED/GREEN recorded; package/full review pending |
| F-03 | Real writable close-injection after Write/Sync reports success or loses secondary typed/anchor cleanup outcome. | Error exposes close cause, preserves meaningful primary/secondary causes and actual durable recovery state/no success output. | RED/GREEN recorded; package/full review pending |
| F-05/F-06 | Equal-byte distinct inode mutation evades old assertions; cut decoding demonstrates mislabeled prepared JOURNAL and omitted restart/final prune. | Assertions see A/B/C/control identity before restore; all named caller/cut rows decode and preserve/refuse/recover correctly. | proof control/matrix GREEN recorded |
| F-07 | Historical canonical record says trace unavailable. | Corrections record recovered 12.383s RED, edit, 15.417s GREEN, hashes and uncommitted-tree limit. | canonical correction recorded |

Manual GSD execution record: `scripts/gsd doctor`, lifecycle `sources`,
`prompt discuss-phase batch-r1-vnext-cutover`,
`prompt plan-phase batch-r1-vnext-cutover --tdd`, and
`go run ./cmd/agentcontractgen check` executed before this ledger. The custom
phase lacks a ROADMAP-compatible Pi runner and Firstmate requires one Terra
writer, so the project adapter's permitted inline fallback is used; it does not
permit skipping RED/GREEN, verification, or the fresh Astra review.

### 2026-09-06 — CP11 Group 2 F-01/F-02/F-03 executed evidence

- **F-01 RED:** the immutable `7e014d00` fixture's final command exited 1 / `FAIL polymetrics.ai/cmd/connectorgen 5.256s`; it records separate validation, candidate, and mutating reopens accepting/moving into replacement B after A. Its two preceding setup failures remain historical setup evidence only. The exact archive/overlays and A/B/CURRENT observations are in [GROUP2-EVIDENCE.md](GROUP2-EVIDENCE.md).
- **F-02 RED:** the exact original-tree loop exited 1 / `FAIL polymetrics.ai/cmd/connectorgen 1.073s`. It repeated 24 descriptor-relative identity substitutions, preserved distinct A/B bytes, and grew numeric process descriptors from 7 to 31 with GC disabled. The earlier Darwin `/dev/fd` fixture setup error did not reach the behavior under test and is not credited.
- **F-03 RED:** the exact original-tree `Publish` command exited 1 / `FAIL polymetrics.ai/cmd/connectorgen 1.365s`. Its narrow callback closed each real temporary once and returned a sentinel only afterward; old behavior performed three write/sync/transition completions and nevertheless returned success, with readable CURRENT and normal JOURNAL removal.
- **F-01/F-02/F-03 GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextPublicationDurableCaptureRefusesReplacementForCurrentAndJournal|TestVNextPublicationRemoveTreeBoundsChildDescriptorsOnReplacement|TestVNextPublicationRemoveTreeClosesChildOnFstatError|TestVNextPublicationPublishReturnsWritableCloseErrorAndFreshRecovery|TestVNextPublicationWriteAtomicPreservesPrimaryAndCloseCauses|TestVNextPublicationAtomicCloseFailuresFollowTheirDurableCuts|TestVNextPublicationRollbackRetainsValidationAndCloseCauses|TestVNextPublicationRecoveryCallersReportRestoreCurrentCloseError|TestVNextPublicationFailedRepairCreationCleansLinkedPredecessorAnchor|TestRunLockRenderReportsWritableCloseFailureWithoutSuccessOutput)$' -v` exited 0 / `ok polymetrics.ai/cmd/connectorgen 12.024s`.
- The GREEN matrix covers six genuine private pending capture paths (CURRENT/JOURNAL × validation/candidate/mutating actual open), records replacement A/B identity/type/empty-directory state and public bytes before restoration, then proves fresh recovery/lock reacquisition/retry. It also covers F-02 identity and fstat early exits without descriptor growth; F-03 close-error reporting for prepared/current/committed journal cuts, active rollback, Recover/Open/Prune and Publish-initial-recovery restoration, CLI no-success output, joined primary/secondary errors, and predecessor-anchor cleanup. Tested source remains uncommitted; its exact identities/diff hash and detailed observations are in [GROUP2-EVIDENCE.md](GROUP2-EVIDENCE.md).

### 2026-09-06 — CP11 F-07 recovered historical RED/edit/GREEN

- **Superseded history:** the older unavailable-trace row remains historical; it is not a claim that current green code was rerun as a RED. The authoritative recovery is `/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-F07-signal-provenance-recovery-7294373166db75466e2c92269f7887f51ceaddc6.md`.
- **Actual RED:** worker record 3224/event 3223/physical line 3224, SHA-256 `65acae633258da13e38c4a9e0a64d532009bc2555b35ffc681a31b8d8828a14f`, ran `gofmt -w cmd/connectorgen/vnext_publication_test.go && go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestConnectorgenMainPreservesNonConsumingSignalTermination$'`. Exit was 1, parent/subtests `interrupt`, `terminate`, and `repeated-interrupt` failed to terminate `validate`, and raw package result was `FAIL polymetrics.ai/cmd/connectorgen 12.383s`.
- **Actual edit and GREEN:** record 3230/event 3229/physical line 3230, SHA-256 `9e4d444101f045b996d3c19ebe5e4d61ecd4abc2b2ab256b7d583bd2d88a3aa9`, is the actual `main.go` signal-routing edit. Record 3238/event 3237/physical line 3238, SHA-256 `b0501c5a52f0fa821bd27907990bae074199e1bfb2563cb4cab3ea9f8c11fcff`, ran the combined non-consuming/lock-render selector to exit 0 / `ok polymetrics.ai/cmd/connectorgen 15.417s`.
- **Limit:** the RED worktree was uncommitted. The records bind the task/session/command/output/edit chronology but yield no full tree SHA; `7294373166db75466e2c92269f7887f51ceaddc6` is later reviewed code, not the failing snapshot. F-08 is independent and this correction claims neither forced-second-signal handling nor mid-transaction cancellation.

### 2026-09-06 — CP11 Group 3 F-05/F-06 and F-03 public-recovery proof

- **F-05 proof control:** Open now compares returned descriptor A to displaced A and B by dev/inode/type plus metadata bytes. Held Prune/Publish/Recover compare A/B before restoration; the deliberately equal-byte, different-inode lease B passes only because identity is asserted. Public-caller and immediate-rollback fixtures gain the same A/B observations; the existing A/B/C collision remains retained. This is oracle strengthening, not a fabricated production RED.
- **F-06 matrix:** prepared/new-selected is explicitly decoded at `AfterCommitSync`; true committed/new-selected is interrupted at `BeforePrune`; successful Publish final-prune starts old-only and retains new CURRENT/committed JOURNAL on late refusal; fresh old-selected/rejected-new recovery covers separate empty/nonempty B variants after `AfterStageRename`; immediate rollback covers both variants. Existing explicit Prune, no-journal Recover/Open, Publish initial recovery, owned-stage cleanup, and read-only Check remain separately named executable rows.
- **F-03 caller completion:** `TestVNextPublicationRecoveryCallersReportRestoreCurrentCloseError` now covers Recover, Open, Prune, and Publish entering a durable recovery. The focused caller execution passed in 4.093s; the full Group 3 command below executes it again with the matrix.
- **Focused GREEN:** [GROUP3-EVIDENCE.md](GROUP3-EVIDENCE.md) records the exact eleven-selector command, exit 0, and package duration 16.675s. It also records actual controls, identities, intentional selection advance, fixture-only restorations, and remaining whole-package/race/static/review limits.

### 2026-09-06 — CP11 coordinated current validation boundary

- The two F-02 test-parent close findings from current-only lint were fixed by checking the cleanup errors. The focused F-02 selector passed in 1.274s and `golangci-lint run --new-from-rev=HEAD ./cmd/connectorgen/...` returned `0 issues.`; this is an introduced-debt result, not a claim that old global lint debt disappeared.
- Final current source execution then passed `go test -count=1 -timeout 20m ./cmd/connectorgen` in 263.677s and `go test -race -count=1 -timeout 20m ./cmd/connectorgen` in 691.666s. This is the repeated current package/race boundary after all F-01–F-08 source/test work, not the earlier partial-wave results.
- Current formatter diff, `go vet ./cmd/connectorgen`, `go build ./cmd/connectorgen`, `go build ./cmd/pm`, `go mod tidy -diff`, `go run ./cmd/agentcontractgen check`, and `git diff --check` passed. Earlier in the same coherent wave, source-lock test gate, canon check, 553-definition validation, runtime preflight, docs check, clean whole-tree boundary report, and release workflow check passed; detailed commands/results are retained above and in `GROUP2-EVIDENCE.md`/`GROUP3-EVIDENCE.md`.
- Next: inspect only intended paths, commit the coherent local code/test/canonical-evidence wave, bind exact SHA/tree, and freeze source for the Firstmate-prescribed independent Astra/xhigh review and Luna/low mechanical index. CP11 remains unaccepted; no push, CP12, no-mistakes, or provider action is authorized.

## 2026-09-06 — CP11 e77 seven-finding TDD ledger (steer 051)

| Group | Audited IDs | Required pre-edit RED or negative control | GREEN acceptance boundary | State |
| --- | --- | --- | --- | --- |
| 1 | F-04-R, F-08-R | **Executed:** F-04's bounded original-behaviour control observed FIFO block and A/B symlink/directory mixing; preserved without rerun in `GROUP1-F04-ORIGINAL-CONTROL.md`. No separately executed pre-edit F-08 pre-flock control was found. The recorded pre-flock gate ran after its test-only readiness/signal instrumentation and is repair-harness proof only; it must not be labelled a pre-edit RED. | **GREEN 2026-09-06:** retained descriptor yields coherent A identity/bytes across FIFO/symlink/directory B; exact retained-directory `EWOULDBLOCK` acknowledgement precedes real-main SIGINT/SIGTERM; direct waits, no-success/state and retry hold. Normal/race focused selector exit 0 / package 9.198s/15.924s in `GROUP1-EVIDENCE.md`. | Complete before Group 2 |
| 2 | F-03-A, F-03-B, F-03-C | **Executed RED:** `GROUP2-F03-ORIGINAL-CONTROLS.md` records the 2.687s bounded actual controls: post-record failure strands prepared-only authority; real definitions/connector close and staged-file primary/completion paths flatten a typed cause, while missing-control open plus parent close loses the absence cause; temporary and quarantine A→empty-B allocations delete B. | Complete identity-proven preparation is retained or removed coherently; meaningful typed causes survive public consumers including absence handling; only owned A is cleaned and B/A state survives empty/nonempty directory cases. | Dynamic matrix executed at `54746816`; fresh review still pending |
| 3 | F-02-P, F-05/F-06-P | Public nested quarantine replacement/opened-child failure lacks complete resource proof; caller/cut matrix lacks raw control/root/private-authority identity and required B variants. | Nested public Recover/Prune preserves accounted residue/ownership and bounded descriptors; all named cuts assert allowed controls, root/content/private graph before restoration, accurate variants, no-clobber and fresh retry. | Authorized for execution after Group 2 |

The exact GSD prompts for issue 4427 were generated/read. `doctor` has the documented unrelated missing issue-122 prompt; compatible Pi role execution is unavailable under the single-writer contract, so this ledger records the permitted inline fallback. The table is the plan-time state; later exact command/output/duration and test identity addenda supersede its execution state.

### 2026-09-06 — CP11 Group 2 intended-behaviour RED (steer 055)

- `GROUP2-F03-REPAIR-EVIDENCE.md` records the desired-assertion selector after
  the immutable original-control checkpoint and before Group 2 production
  edits. The initial missing-argument compile failure is excluded. The actual
  2.567s selector exited 1: prepared-only authority, both empty replacement-B
  deletions, and all three exercised compound-cause losses failed their intended
  contract. No desired-assertion source hash was recorded at execution time;
  the earlier immutable original-control source/output remains separately
  bound. Steer 056 later reconstructed the target test bytes exactly relative
  to the stated FileChange/formatter record set (snapshot SHA-256
  `041f44816b2bd103a9b133dea196d6a68c243e64c6751375944ca2d5feb3a228`), but
  that is not an execution-time or full working-tree identity. This is a
  chronology limit, not a GREEN claim.

- **GREEN:** the same F-03-A/B/C selector subsequently exited 0 / `ok
  polymetrics.ai/cmd/connectorgen 4.389s`. It covers coherent post-record
  recovery/Check/retry, typed compound producer/consumer errors including pure
  absence, and empty/nonempty replacement B through public Publish/Prune. The
  complete source sibling audit and its intentional read-only teardown limits
  are recorded in `GROUP2-F03-REPAIR-EVIDENCE.md`. The source-bound receipt at
  physical `10897` separately records the then-uncommitted intermediate full
  package PASS (`271.387s`; wall `273.672594250s`; raw record SHA-256
  `4fd7390eef0b432f8f5f983d3924f9360bde2993d51586fa48bc6658634e0255`). It
  does not certify later edits; final three-group package/race/static and
  independent review gates remain pending.

### 2026-09-06 — CP11 F-03 steer-058 resource-backed GREEN completion

This is added post-repair proof, not a rewrite of the earlier RED. The current
test source is `cmd/connectorgen/vnext_publication_group2_original_test.go`.
It uses only the inert direct-descriptor seams
`vNextPublicationControlRecordHooks`, record/directory-sync points in
`vnext_publication.go` and `vnext_publication_repair.go`, the raw
opened-file-close seam in `vnext_publication_dir.go`, and read-control
completion/close seams in `vnext_publication.go`.

- **F-03-A GREEN:** `TestCP11F03ARepairPreparationFrontierMatrix` proved the
  before-record, actual short-write, actual Sync+injected completion, actual
  Close+injected completion, post-record, transaction-Sync, and connector-Sync
  cuts for JOURNAL absent→present, CURRENT present→present, and JOURNAL
  present→logical-absence. It records prepared/phase/anchor/public identity,
  nonmutating pending Check, fresh recovery/Check/retry. Base interruption is
  absent/absent in `TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation`
  and both CURRENT/JOURNAL present/present in
  `TestCP11F03ARepairBasePresentPresentAuthorityRecovers`.
- **F-03-B GREEN:** `TestCP11F03BRepairCompoundCausesRemainInspectable`
  follows all remaining real producer/consumer pairs (open/parent Close,
  parent/opened-file Close, read/Close/pure absence, marker/prepared/phase
  writers, predecessor, stage, and capture) and checks joined causal identity.
- **F-03-C GREEN:** the four `TestCP11F03CRepair*` selectors cover Publish
  temporary, CURRENT/JOURNAL transition temporary, generation quarantine, and
  stage quarantine. Each preserves moved A and empty/nonempty B identity/type
  and exact B bytes/residue before any fixture-only cleanup, then has its
  specified fresh recovery/retry path.

Commands/results: F-03-A normal matrix `ok 20.804s`; three state-class race
selectors `ok 10.200s`, `10.089s`, and `11.507s`; base normal/race `ok
2.617s`/`4.095s`; B/C normal/race `ok 8.442s`/`11.749s`. The preserved earlier
physical `11540/11549/11563/11577` receipts remain associated with their
earlier two-class source state. See `GROUP2-F03-REPAIR-EVIDENCE.md` for full
selectors, source-state limits, and all assertions. This closes the F-03
matrix, not CP11 or the final Group 2 review/static boundary.

### 2026-09-06 — Group 2 bind; Group 3 proof-only TDD record (steer 061)

- **Bound Group 2 GREEN:** exact checkpoint `54746816735a964d0177a7a64646d29561f08180`
  contains the F-03 matrix implementation/test/evidence result stated above.
  The interim `10897` and four `11540/11549/11563/11577` receipts remain
  correctly limited to their earlier source states. F-03-C's B is a replaced
  directory (empty or with foreign bytes), not a regular file.
- **F-02-P proof control:** a new public Recover/Prune fixture must traverse
  actual root quarantine into nested owned stage/generation removal, replace a
  nested child after its identity observation and before open, and fault the
  real opened child identity. It must observe A/B identity/type/bytes at
  public and quarantine paths before restoration, account for root/quarantine/
  child/lock descriptors without GC/finalizer credit, state permitted earlier
  sibling removal and residue, then prove fixture-only restoration and bounded
  fresh recovery/retry. This is an oracle-completeness test; a passing old
  production path is not a fabricated RED.
- **F-05/F-06-P proof control:** each mandatory caller/cut must retain
  descriptor-safe raw control bytes/type/inode plus decoded state, actual
  selected/rejected root content, and real transaction/prepared/phase/anchor
  identity before any restoration. Initial Publish normally creates private
  transaction authority; prior documentation asserting otherwise is corrected.
  Legitimate new controls/phase/capture state are compared as real-cut
  advancement, not against an obsolete pre-Publish inode. Each unheld
  destructive generation row requires empty and nonempty replacement **regular
  `.lease` files** B; held-reader rows retain an empty lease-file B for
  contention. Group 2 F-03-C's distinct temporary/quarantine B is a
  replacement directory.

### 2026-09-06 — CP11 Group 3 F-02-P/F-05/F-06-P executed proof

- **F-02-P:** `TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership`
  is a serial, bounded-no-GC public-path matrix: Recover-owned-stage, Prune
  stale-generation, and Publish-final-prune generation each execute nested
  post-identity/pre-open directory A→B replacement and an actual opened-child
  identity failure. Before only fixture-owned reconstruction it asserts public
  root absence, retained quarantine-candidate root identity, nested A/B
  identity/type/bytes or retained child bytes, observable partial residue, and
  no descriptor growth. It does not claim automatic recursive rollback; fresh
  recovery follows the assertion. The historical direct-child RED remains the
  behavioral red. This is the new public proof GREEN, not a new production RED.
- **F-05/F-06-P:** `vNextPublicationAssertDurableCutWitnessForTest` uses
  descriptor-bound CURRENT/JOURNAL witnesses and scans the actual retained
  authority graph. It asserts raw regular-control bytes/inode, decoded heads,
  selected/rejected/stale stable-tree content, marker, transaction identity,
  prepared record, all phases and validated anchors. The named public matrix
  invokes the ten destructive caller/cut obligations plus read-only Check.
  `TestVNextPublicationUnheldDurableRowsRetainEmptyLeaseReplacementB` adds the
  empty half of every unheld schedule; existing named controls are the paired
  nonempty cases. These B values are the real regular `.lease` member, not a
  directory substitute.
- **No false RED:** the prior Group 3 sentence denying ordinary private
  authority is corrected in `GROUP3-EVIDENCE.md`; first normal Publish retains
  the marker and real repair transactions. The added tests can pass existing
  repaired production, which is the intended proof-only result.
- **Focused GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run
  '^(TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership|TestVNextPublicationUnheldDurableRowsRetainEmptyLeaseReplacementB|TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers|TestVNextGenerationPublisherCheckIsReadOnly|TestVNextPublication(CommittedJournalNewSelectedRecoveryRejectsLateLeaseReplacement|SuccessfulPublishFinalPruneRejectsLateLeaseReplacement|FreshRejectedNewRecoveryRejectsLateLeaseReplacement|ImmediateRollbackRejectsLateLeaseReplacementIdentityVariants))$'` passed, package `30.270s`. The same exact selector with `-race` passed,
  package `42.177s`.

### 2026-09-06 — CP11 Group 3 final observation boundary (steer 063)

- **Object kinds corrected:** F-03-C's B is a replacement directory; F-05/F-06
  B is the actual regular `.lease` member; F-02-P's nested A→B is a directory.
  The eight preserved changing-source result records are retained at the
  Firstmate receipt path named in `GROUP3-EVIDENCE.md`; no wording-only rerun
  is relabelled as an execution result.
- **F-02-P GREEN strengthening:** the public Recover-owned-stage/Prune/final
  Publish nested matrix now reads the actual stage-owner marker, observes raw
  heads and real private authority before fixture reconstruction, and asserts
  no lsof descriptor remains below the unique root after each fresh recovery.
  GC remains disabled and the independent numeric descriptor count remains
  bounded. The extended normal/race selector, including owned-stage cleanup,
  passed `32.789s`/`45.628s`.
- **Final current-source GREEN:** full `cmd/connectorgen` normal/race passed
  `319.928s`/`783.834s`; formatter/diff/vet/build/tidy/agent-contract/new-code
  lint all passed as recorded in `GROUP3-EVIDENCE.md`. This is an execution
  checkpoint, not CP11 acceptance; an independent exact-SHA review remains
  required.

### 2026-09-06 — CP11 current gate ledger, artifact-only bind (Firstmate 064)

The frozen behavioral candidate is
`7481d1770a21cc95869fd10bf281f632af48c089` / tree
`a2e583336ffa8ad86a0de95110259342bfa6dab0`; this accounting append does not
change source/test behavior. Fresh recorded GREEN is: source lock target
`make connectorgen-vnext-locks` (`206.821s`); canon target
`make connector-canon-check`; definition target `make connectorgen-validate`
(553/0); `TestFoundationAtlasSelectorsResolve` (`1.326s`); docs target
`make docs-check`; and release target `make release-workflow-check`. Exact
runtime preflight exited 0 but its Go package output was **cached**, so it is
not upgraded to fresh execution evidence. The package normal/race `319.928s` /
`783.834s` receipts remain separate exact-candidate checks.

The literal `make connector-boundary` invocation is a pending ledger cell:
the parallel wrapper lost its session/result/output before the process ended,
and no duplicate was authorized. It is neither GREEN nor a waived check. See
the immutable-result record (SHA-256
`a9bcdf60d0fb4945e096216727a39344ae87816e18abde8b6bdb71022e2bc908`) at
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-7481d177-current-gate-report-064.md`.
Fresh independent exact-SHA review remains the required next acceptance gate.
