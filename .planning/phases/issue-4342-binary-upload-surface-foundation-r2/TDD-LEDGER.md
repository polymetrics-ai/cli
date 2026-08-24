# TDD ledger — issue #4342 binary upload CLI and certification foundation

## Red

- [x] `commandrunner`: a new `binary_upload` intent cannot resolve unless it references an implemented write action with `body_type` `binary_upload`, `base64_upload`, or a multipart file part; arbitrary write paths and JSON/form actions fail before I/O.
- [x] `internal/cli`/`app`: a GitHub-style binary upload command creates a plan and does not directly dispatch; preview/approval binds source bytes and later file mutation prevents provider I/O (the latter execution protections remain covered by the pre-existing engine byte-transfer tests cited in the issue brief).
- [x] `certify`: refusal produces `blocked`/`not_live`, never an upload pass; transfer proof requires exact byte count/digest, provider response, independent read-back, and cleanup.
- [x] `connectorgen`: a `binary_upload` command projects to a dedicated binary-upload sweep/evidence class; `file_upload` stays rejected as executable.
- [x] **Review remediation, provider contract:** new red tests for F-4343-01 and F-4343-03
  refuse a GHES credential before public-host I/O and treat every non-`201` GitHub upload response
  as a retained failed receipt.
- [x] **Review remediation, App lifecycle/confidentiality:** a new in-process App test for
  F-4343-02/F-4343-05/F-4343-09 must execute plan → preview → approval → exact-byte `201` transfer,
  reject changed bytes before I/O, and prove its local-path sentinel is absent from outputs/state.
  It must fail when binary upload exposes an approval token before preview or fails to issue one
  after persisted preview, and must prove pre-preview execution has zero provider calls.
- [x] **Review remediation, media and user source root:** F-4343-04/F-4343-08 red tests
  refuse unenforceable media declarations before I/O and prove project-root relative source files
  execute without callers naming `.polymetrics`.
- [x] **Review remediation, generated guidance:** F-4343-07 guide-rendering test rejects an
  unsupported sibling note that denies the bounded executor its implemented sibling uses.

## Green

- [x] Production implementation for every red assertion.

### Public planner route execution record

- **Red:** `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner -run '^TestBuildWriteCommandPlansOnlyDeclaredBinaryUploadActions$' -count=1` failed as intended. A declared `binary_upload` command was blocked with `only implemented ETL stream commands are executable` before it could create the approval-bound plan.
- **Green:** The same command passed after admission, declaration-bound source validation, and plan construction were added. It covers raw binary, base64, and multipart required-file actions, then proves `Run` remains lifecycle-blocked instead of performing I/O.
- **Refactor:** The installed GitHub command has a separate behavioral test in `github_write_contract_test.go`; static validation and runtime reuse `engine.BinaryUploadSourcesForWriteAction` so they cannot promote different action shapes.

## Refactor

- [x] Reuse the existing write-command plan and engine binary payload preparation; no parallel raw upload/requester or generic CLI surface.
- [x] Keep report data typed and capability names explicit rather than inferring upload from command strings or a single `binary` cell.

## Deliberate break proof

- [x] Temporarily removed `binary_upload` planner admission and ran the public behavioral tests. `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner -run '^(TestBuildWriteCommandPlansOnlyDeclaredBinaryUploadActions|TestGitHubReleaseAssetUploadBuildsBoundBinaryPlan)$' -count=1` failed: the installed GitHub case and all raw/base64/multipart cases were rejected with `connector command is not a reverse ETL write command`.
- [x] Restored the admission with `apply_patch`; the exact command then passed (`ok polymetrics.ai/internal/connectors/commandrunner 2.440s`).
- [x] Temporarily removed the `binary_upload` branch in `planRequiresPersistedPreview`. `GOFLAGS='-p=3' go test -timeout 20m ./internal/app -run '^TestBinaryUploadConnectorCommandPersistsPreviewBeforeApproval$' -count=1` failed with `no longer matches its approved preview` where the test requires the stronger `must be previewed` pre-I/O refusal. Restored the branch and reran the test successfully.

## Commands/results

- Red: `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner -run '^TestBuildWriteCommandPlansOnlyDeclaredBinaryUploadActions$' -count=1` failed before implementation with `only implemented ETL stream commands are executable`.
- Green: `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/engine` passed; `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner` passed; `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/certify` passed; and `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen` passed after the final refactor.
- Review green: `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/engine -run '^TestGitHubReleaseAssetUpload_(InstalledCommandSendsExactBytes|RejectsEnterpriseCrossOriginBeforeIO|RequiresCreatedResponse|EnforcesDeclaredMediaPolicy|RejectsMissingChangedUnsafeOrOversizeFile|EmptyFileAndTerminalFailures)$' -count=1` passed (3.452s); the commandrunner admission set passed (4.075s); and the App lifecycle test passed (3.916s).
- Live proof: the real GitHub upload-host transfer, exact `201`, independent read-back, and provider cleanup are recorded in `LIVE-PROOF.md`. This supports the public path but does not manufacture a generated certification pass: the stage remains truthfully `not_live` until its own transfer/read-back/cleanup input contract exists.
