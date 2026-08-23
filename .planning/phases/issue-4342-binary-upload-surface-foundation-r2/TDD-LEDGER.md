# TDD ledger — issue #4342 binary upload CLI and certification foundation

## Red

- [x] `commandrunner`: a new `binary_upload` intent cannot resolve unless it references an implemented write action with `body_type` `binary_upload`, `base64_upload`, or a multipart file part; arbitrary write paths and JSON/form actions fail before I/O.
- [x] `internal/cli`/`app`: a GitHub-style binary upload command creates a plan and does not directly dispatch; preview/approval binds source bytes and later file mutation prevents provider I/O (the latter execution protections remain covered by the pre-existing engine byte-transfer tests cited in the issue brief).
- [x] `certify`: refusal produces `blocked`/`not_live`, never an upload pass; transfer proof requires exact byte count/digest, provider response, independent read-back, and cleanup.
- [x] `connectorgen`: a `binary_upload` command projects to a dedicated binary-upload sweep/evidence class; `file_upload` stays rejected as executable.

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

## Commands/results

- Red: `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner -run '^TestBuildWriteCommandPlansOnlyDeclaredBinaryUploadActions$' -count=1` failed before implementation with `only implemented ETL stream commands are executable`.
- Green: `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/engine` passed; `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner` passed; `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/certify` passed; and `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen` passed after the final refactor.
