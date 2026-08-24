# TDD Ledger: Issue #4305

## Planned red/green evidence

| Slice | Red | Green | Edge and regression |
| --- | --- | --- | --- |
| Declared nested body | An installed synthetic command refuses a declaration-backed nested body. | The exact declared route, query, headers, content type, and nested JSON reach the provider double. | Body values cannot populate route/query mappings. |
| Recursive validation | Input is either accepted as opaque raw JSON or rejected only after a request is formed. | The materializer recursively accepts only declared fields and types before request creation. | Unknown, required, malformed, over-depth, over-field, over-item, and over-byte cases have zero I/O. |
| Action reuse and approval | Typed actions have no shared structured-body construction or payload identity. | CLI and typed-action paths share the materializer and confirmation identity. | Separate actions cannot borrow fields; post-preview nested mutation is rejected before I/O. |
| Surface and contract | Generated command surface has no typed representation for declared nested inputs. | Help/schema present declared typed inputs and lack any raw-body bypass. | Scalar/form/SCIM/binary/specialized GitHub focused tests retain outcome. |
| Write query absence | `omit_when_absent` does not tolerate an absent `record.*` query value. | Only an object-form, source-locked query declaration can omit its own missing `record.*` value. | Required/undeclared/wrong-source/malformed values fail before I/O; config, secret, incremental, and explicit record values retain their established behavior. |

## Actual evidence

### 2026-08-20 — declaration-backed REST structured-body red checkpoint

- Red: `go test -timeout 20m ./internal/connectors/commandrunner -run TestBuildOperationDirectWriteCommandSupportsDeclaredStructuredRESTBody -count=1` failed. The production-shaped synthetic command carries a declared path field, query field, scalar body field, closed nested object, and bounded nested array. Runtime refused the first `json` body flag with: `structured JSON variables require a fixed GraphQL operation, got "rest_write"`.
- Observable contract: no caller-provided method, route, content type, or raw body exists in the fixture; the only missing capability is admitting the source-owned nested fields of the declared REST body.
- Next green slice: add an operation-owned structured-body preflight, use it before parsing the JSON flag, then materialize the value only into the declared body path. Engine tests will prove recursive schema and limit failures occur before transport I/O.

### 2026-08-20 — write-query record-absence red checkpoint

- Red: `go test -timeout 20m ./internal/connectors/engine -run 'Test(BuildWriteQueryOmitWhenAbsentScopesMissingRecordValuesToTheirDeclaredQuery|WriteActionRecordQueryRejectionsHappenBeforeProviderIO)' -count=1` failed in `missing optional record value is omitted` with `interpolate: unresolved key "optional" in record`.
- Observable contract: the existing shared query resolver deliberately omits only absent config, secret, or incremental values. A write query's exact object-form declaration must additionally, and only additionally, omit an absent `record.*` reference.
- Next green slice: keep the stream/check resolver unchanged; add a write-only resolver that recognizes typed unresolved record errors only under `omit_when_absent`, retaining all other pre-I/O failures.

### 2026-08-20 — declaration-backed REST structured-body green checkpoint

- Green: `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen ./internal/cli -run 'Test(OperationDirectWriteStructuredRESTBody|OperationStructuredJSONBodyPreflight|BuildOperationDirectWriteCommandSupportsDeclaredStructuredRESTBody|GitHubUserDraftCommandBuildsFixedGraphQLMutation|Validate_CLISurfaceDirectWriteStructuredRESTBody|StructuredRESTBodyCommandHelpAndManual|ConnectorsManualDocumentsConnectorArchitectureAndGithubExamples)' -count=1` passed.
- The runtime admits only a declared top-level object/array REST body field after validating its bounded closed schema, and `operationWriteBody` repeats that exact declaration validation for typed actions. The provider fixture proves exact route/query/header/content-type/body, action isolation, approval digest binding, and zero I/O for rejected inputs.
- Existing GraphQL preflight remains the direct-read behavior, and the focused specialized GitHub assertion remains green.

### 2026-08-20 — write-query record-absence green checkpoint

- Green: `go test -timeout 20m ./internal/connectors/engine -run 'Test(BuildWriteQueryOmitWhenAbsentScopesMissingRecordValuesToTheirDeclaredQuery|WriteActionRecordQueryRejectionsHappenBeforeProviderIO|WriteActionOptionalRecordQueryIsOmittedOrPreservedAtProvider|WriteActionQuery)' -count=1` passed.
- `buildWriteQuery` now uses a write-only resolver. It omits only an unresolved `record.*` reference whose exact object-form `QueryParam` sets `omit_when_absent`; plain/required entries, an undeclared record key attempting to populate a declared query, wrong sources, malformed values, and all other errors reject before provider I/O.
- Config, secrets, and incremental omission/default behavior stays in the shared resolver and is regression-tested unchanged. An explicit `record.*` value is transmitted through the declared parameter unchanged.

### 2026-08-21 — CI compiler and generated-website-data repair

- Red: the PR's `verify`, `connector-boundary`, `govulncheck`, native build, and package-binary jobs all stopped before execution with `internal/connectors/engine/read.go:960:53: not enough arguments in call to resolveExpr`. The focused local engine test reproduced the same compiler failure.
- Root cause: the secret-safe interpolation hardening extended `resolveExpr` with its `allowControlCharacters` argument, while `interpolateDeclaredHeader` retained the obsolete three-argument call.
- Green: the declaration-bound header path now supplies `false`, preserving its established unsafe-control-character rejection. `go test -timeout 20m ./internal/connectors/engine -run 'TestOperationDirectWriteStructuredRESTBody' -count=1` and the full engine package pass; `go build -trimpath -ldflags='-s -w' ./cmd/pm`, `go vet ./...`, `go run ./cmd/connectorgen boundary . --json`, and `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` also pass.
- Website generated data: `npm run gen:website-data` updates only `website/lib/docs.generated.ts`; a second generation is stable and website typechecking passes. The tracked payload now includes the existing declaration-bound structured-write documentation.

### 2026-08-21 — CI Verify follow-up red checkpoint

- Red: hosted Verify run `32430447912` failed two deterministic assertions. `TestDirectWriteCommandFailurePreservesErrorContent` expected a provider body containing `server-token` to appear in the public and persisted error, despite the direct-write safety policy deliberately returning the generic HTTP diagnostic for secret-bearing operations. The same run found the four connector-manual golden transcript entries stale after the corrected structured-write documentation changed `maps_to=body.<field>` to the declaration-accurate schema-path wording.
- Observable contract: a secret-bearing provider response must stay available only through the typed internal `*connsdk.HTTPError` cause while public/persisted diagnostics contain no secret; generated manual variants must exactly match the current connector help.
- Next green slice: correct the stale app assertion to prove redaction plus typed-cause preservation, then regenerate only `help_connectors`, `bare_connectors_manual`, `json_connectors_manual`, and `connectors_help_github_known_legacy_namespace_help_intercept` from the test harness.

### 2026-08-21 — CI Verify follow-up green checkpoint

- Green: `go test -timeout 20m ./internal/app -count=1` passed in 256.922s and `go test -timeout 20m ./internal/cli -count=1` passed in 561.504s. The app test now asserts the observable safe public/persisted diagnostic and verifies that the exact bounded provider response remains available solely through the typed cause.
- Green: the four generated transcript entries were refreshed with `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 POLYMETRICS_GOLDEN_TRANSCRIPT_NAMES=help_connectors,bare_connectors_manual,json_connectors_manual,connectors_help_github_known_legacy_namespace_help_intercept go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1`; the full subsequent CLI package run proves the tracked fixture is stable without update mode.
- Scope: this active no-mistakes CI-step repair did not invoke no-mistakes controls, a CI re-run, a push, or a PR mutation; those remain owned by the outer executor.

### 2026-08-23 — active CI-phase compiler and structured-write repair

- Manual-GSD fallback: this is an assigned no-mistakes CI phase, so the outer executor retains ownership of lifecycle controls, CI re-runs, pushes, and PR updates. This repair records its TDD evidence here without invoking those controls.
- Red: `go test -timeout 20m ./internal/connectors/engine -run '^$' -count=1` reproduced the hosted compile failure: the superseded `structured_body_plan_transform.go` redeclared the public structured-body plan transforms and `Connector`/`Base` methods already supplied by the canonical direct-write implementation. With those duplicates removed, the compiler exposed the branch's deleted declared-header validators, which direct-write preflight still requires.
- Red: the full engine package then exposed stale response-fixture assumptions and an actual source-schema gap: JSON GraphQL fixtures missing their response media type bypassed GraphQL application-error handling, and a fixed scalar array prefix shorter than its declared `minItems` could never be completed by a caller-owned array without overriding provider-owned data.
- Green: `go test -timeout 20m ./internal/connectors/engine -count=1` passes after retaining the one canonical structured-body implementation, restoring declaration-header validation, binding headers only after declaration-owned query/auth values, and rejecting the unmergeable fixed array container before I/O. GraphQL fixtures now explicitly model JSON responses, and the provider response is retained without a lossy re-marshal.
- Red: `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestBuildOperationDirectWriteCommandSupportsDeclaredStructuredRESTBody$' -count=1` then caught two generated-CLI regressions: an exact `json.Number` was incorrectly asserted as `float64`, and a malformed numeric body value was reflected into the public error text. The former would lose provider numeric precision if "fixed" by coercion; the latter could disclose a sensitive value supplied where a number is required.
- Green: `go test -timeout 20m ./internal/connectors/commandrunner -count=1` passes after preserving the exact `json.Number` contract and emitting the flag name plus expected numeric domain without echoing the rejected value.
- Green: final local CI-equivalent checks passed: `go build -trimpath -ldflags='-s -w' -o /dev/null ./cmd/pm`; `go vet ./...`; `go run ./cmd/connectorgen boundary . --json`; `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`; `go test -timeout 20m ./cmd/connectorgen -run '^$' -count=1`; and the complete tagged GitHub fresh-binary fixture command from `.github/workflows/verify.yml` (1265.704s).
- Scope: no production connector definition or generic HTTP/raw-body escape hatch was added. The repair remains confined to the declaration-bound engine, its regression fixtures, and this GSD/TDD evidence.

### 2026-08-23 — active CI-phase multipart integrity diagnostic repair

- Red: hosted Verify and `go test -timeout 20m -count=1 -run '^TestApprovedFixtureOperationRequestMultipartRejectsTamperAndStaleApprovalBeforeProviderIO$' -v ./internal/connectors/conformance` failed for the tampered multipart fixture with the generic `provider returned an HTTP error` diagnostic instead of `changed since approval`.
- Root cause: the credential-safe direct-write diagnostic branch treated every error from a credential-bound request as a provider error. The multipart snapshot's local, pre-I/O digest mismatch was therefore hidden despite the existing fixture's zero-provider-capture assertion.
- Green: only HTTP and URL-bearing transport errors retain the generic diagnostic; safe local errors pass through the existing literal-redaction path. The focused conformance regression, `go test -timeout 20m -count=1 ./internal/connectors/conformance`, and `go test -timeout 20m -count=1 ./internal/connectors/engine` pass. Credential-bound provider-error tests still prove provider echoes remain generic and secret-safe.
- Scope: this assigned CI phase did not invoke no-mistakes controls, rerun CI, push, or mutate the PR; those remain with the outer executor.

### 2026-08-24 — active CI-phase generated connector-manual fixture repair

- Manual-GSD fallback: this is an assigned no-mistakes CI phase, so the outer executor retains ownership of lifecycle controls, CI re-runs, pushes, and PR updates.
- Red: hosted Verify run `32695514116` and `go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1` failed only `json_connectors_manual` and `connectors_help_github_known_legacy_namespace_help_intercept`.
- Root cause: the structured-write manual's now-required closed schema-path wording changed the runtime output after those two generated transcript records were last refreshed. Both records render the same connector manual in different output modes.
- Green: regenerated only those two records with `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 POLYMETRICS_GOLDEN_TRANSCRIPT_NAMES=json_connectors_manual,connectors_help_github_known_legacy_namespace_help_intercept go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1`. Re-running `go test -timeout 20m ./internal/cli -run '^(TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals)$' -count=1` without update mode passed, as did `git diff --check`.
- Scope: only generated transcript records changed; no production connector definition, request behavior, raw-body escape hatch, no-mistakes control, CI re-run, push, or PR mutation was invoked.
