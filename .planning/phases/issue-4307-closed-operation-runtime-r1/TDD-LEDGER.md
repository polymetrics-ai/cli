# TDD Ledger: Issue #4307

## Planned Red/Green Evidence

| Slice | Red | Green | Edge/regression proof |
| --- | --- | --- | --- |
| Declaration-owned headers | Synthetic operation declarations cannot represent a typed request header or command mapping. | One exact declared `header.<name>` map becomes a validated request header. | Two identities prove header/path/query/body and operation boundaries remain isolated. |
| Header hardening | Forbidden header variants can flow through request construction or invalid input reaches a provider double. | Canonical protected-header/name/value validation refuses before I/O. | Case/normalization, duplicate, unknown, CR/LF, malformed, byte, enum/pattern/length, requiredness, and runtime-owned failures leave a request counter at zero. |
| F4 kind admission | Valid synthetic status/text/binary/multipart declarations fail at an unregistered loader, validator, generator, or runner mirror. | Every named kind reaches exactly its existing typed executor through the installed command path. | Invalid/bare/unbounded declarations fail before I/O and cannot fall through to generic REST handling. |
| Download and text output | Existing primitive cannot prove exact media/redirect/cap/output atomicity through a declaration. | A bounded declared response produces only the expected final output. | Over-cap, wrong media/charset, redirect-policy, stream, unsafe-path, and replacement failure yield no partial-success file. |
| Upload and multipart | Caller data can select arbitrary parts/content types or evade declaration limits. | Declared parts, media, file/count/byte caps, approval, and destructive confirmation construct the exact request. | Unknown parts/raw bytes/unbounded file/changed approval digest fail before execute I/O. |
| Declared result preservation | Fixed-operation results drop unusual ordinary provider fields because a policy guesses they are irrelevant. | Every status/header/body/field element admitted by the exact declared response contract reaches the result unchanged. | Credential and transport-secret canaries retain their field name with the established explicit masking marker; the test proves no other scope/tier/risk filter exists. |
| Existing surfaces | New common paths can change current typed command semantics. | Existing operation families retain their present output and preflight behaviour. | GraphQL, scalar/form/SCIM, #4305 structured bodies, credential/auth redaction, and no-credential preflight tests stay green. |

## Actual Evidence

### 1. Typed request headers — red → green

- **Red:** Before the declaration fields and request channel existed,
  `go test -timeout 20m ./internal/connectors/engine -run 'TestOperationDirectRead(UsesOnlyDeclaredTypedHeaders|RejectsHeaderEscapeHatchesBeforeNetwork)' -count=1`
  failed to compile because `OperationParameter.Schema`,
  `OperationParameter.MaxBytes`, and `OperationDirectReadRequest.Headers` did
  not exist.
- **Green:** The same focused engine command passed after adding the closed
  declaration shape and shared request construction. The two synthetic
  identities prove their exact header reaches the provider while path/query/body
  stay separate.
- **Hardening:**
  `go test -timeout 20m ./internal/connectors/engine -run TestOperationDirectReadRejectsHeaderEscapeHatchesBeforeNetwork -count=1`
  passes with a zero-I/O provider counter for unknown, cross-operation,
  malformed, duplicate-case, protected/runtime-owned, CR/LF, byte, schema,
  enum, and missing-required inputs.

### 2. Bounded F4 response/result contract — red → green

- **Red:** The pre-change local result types had no bounded declaration-owned
  response header projection, no status/header fields on binary results, and no
  exact charset contract for text exports.
- **Green:**
  `go test -timeout 20m ./internal/connectors/engine -run 'Test(OperationDirectReadPreservesDeclaredResponseFieldsAndMasksKnownSecrets|OperationStatusCheckUsesDeclaredHEADWithoutJSONBody|OperationTextExport|OperationDirectWritePreviewsApprovesAndExecutesSingleFormRequest)' -count=1`
  passes. It proves complete ordinary body/status/header preservation,
  presence-preserving secret-header masking, header caps, text media/charset
  refusal with no file, and preview-digest rejection for a changed typed header.
- **Regression:** `go test -timeout 20m ./internal/connectors/engine -count=1`
  passed after the implementation.

### 3. Generated command mapping and App lifecycle — red → green

- **Red:** The operation command mapper admitted only path/query/body targets;
  `header.<name>` had no safe flag derivation or runtime request field.
- **Green:**
  `go test -timeout 20m ./cmd/connectorgen -run 'Test(ParamsImportImportsOnlyBoundedTypedHeaders|DeriveCommandParameterFlagsAddsExactHeaderMapping|CheckCLISurfaceOperationHeaderMappingsRequiresExactDeclaredHeader)' -count=1`
  and
  `go test -timeout 20m ./internal/connectors/commandrunner -run TestRunImplementedOperationDirectReadCommand -count=1`
  pass. They prove bounded source import, exact case-sensitive mapping, required
  generated help metadata, and installed-command request fidelity.
- **App proof:** `TestDirectWriteCommandPlanPreviewApprovalAndExecute` now
  asserts the plan persists the exact header, passes it through preview and
  execute, and observes it at the provider; the plan hash and engine preview
  include the header so changes invalidate approval.

### 4. Documentation and static gates

- `go run ./cmd/connectorgen validate internal/connectors/defs` passed: 552
  connector bundles, 0 findings.
- `go run ./cmd/connectorgen surface-sync --check internal/connectors/defs`
  passed without generating drift.
- `docs/migration/conventions.md` records declaration/adoption, generated-help,
  output-preservation, bounds, and masking rules; no existing provider command
  adopts a new header in this foundation slice, so no provider-specific manual
  or website page changes are appropriate.

The remaining full-package, build, vet, boundary, and `make verify` evidence is
recorded in `VERIFICATION.md` after those final gates complete.

### 5. Generic multipart fixture approval — red → green

- **Red:** Before this extension, a `fixtures/writes/<action>.json` record for
  a declared multipart file part reached the real preview with no
  `ApprovedPayloadSHA256` entry. The action therefore failed with a missing
  approved-payload-digest error, or a connector could only avoid that path by
  adding its own fixture bypass. The new conformance cases did not compile
  because the shared digest collector and fixture staging contract did not
  exist.
- **Green:**
  `go test -timeout 20m ./internal/connectors/conformance -run 'Test(WriteRequestShape_MultipartFixture|ApprovedFixtureWriteRequestMultipartPayloadDigest)' -count=1`
  passes. It stages a provider-neutral declared fixture payload, confirms the
  exact SHA-256 is present in the real fixture approval request, then proves
  successful replay, missing/oversized-asset refusal, changed-after-approval
  refusal, and stale-grant refusal with no provider capture request.
- **Regression:** `go test -timeout 20m ./internal/connectors/conformance -count=1`
  passes. The collector reuses the declaration's multipart part, project-root,
  regular-file, per-file and aggregate byte rules; it returns only opaque
  digests and never logs payload bytes. `docs/migration/conventions.md` now
  names the provider-neutral fixture-asset convention.
