# PLAN — Issue #60 GitHub binary transfer

## Task contract

- Primary issue: #60 — add a small but real binary/file transfer slice for GitHub CLI parity.
- Parent issue/PR: #44 / PR #49, base branch `feat/44-github-cli-parity`.
- Worker branch: `feat/60-github-binary-transfer`.
- Allowed write scope: `.planning/phases/github-binary-transfer/**`, `internal/connectors/engine/**`, `internal/connectors/commandrunner/**`, `internal/connectors/defs/github/{operations.json,cli_surface.json,api_surface.json,docs.md}`.

## Slice boundary

Deliver this coherent slice:

1. Engine support for executing `binary_download` operations declared in `operations.json`.
2. Command runner dispatch from implemented `cli_surface.json` operation commands into that engine support.
3. Safe output path policy for disk writes: explicit destination/directory, relative-safe paths, no traversal/control chars/home expansion/root/system locations, no symlink/device/dir targets, no overwrite unless operation policy and request both allow it.
4. Size limit enforcement from operation metadata and request max bytes.
5. JSON manifest result containing connector, operation/source command, resolved API path/status, written path, byte count, content type, checksum metadata when available, and overwrite flag.
6. GitHub metadata for release asset download and repo archive download commands/operations.
7. Documentation note that `file_upload`/upload-like paths are deferred because safe remote binary writes need reverse-ETL approval semantics and overwrite/clobber policy.

Explicitly defer:

- Archive extraction.
- Executing downloaded content.
- File upload/release upload execution.
- Broad external output roots or symlink following.

## TDD plan

### Red slice A — commandrunner and engine binary download tests

Add failing tests before production edits for:

- Unsafe output path rejection (`../`, control chars, `~`, `/`, system roots, path separators in remote filename).
- Overwrite denial when a file already exists and overwrite is not explicitly permitted.
- Size-limit failure when response exceeds operation/request cap.
- Successful binary download manifest with path, byte count, content type, checksum metadata, source operation, command, and no stdout payload.
- HTTP error handling propagates a safe, redacted error.

### Red slice B — GitHub command metadata wiring

Add failing tests for:

- `release download` dispatches to operation `github.release.download_assets` with required `asset-id` and output destination flags.
- `repo archive` dispatches to operation `github.repo.archive` with format/ref/destination flags.
- Embedded GitHub bundle contains both operations and cli_surface commands.

### Green implementation

- Introduce narrow engine types/interfaces for binary download execution (not generic HTTP writes).
- Extend commandrunner `Request`/`Result` for binary destination policy and manifest return.
- Resolve operation by command metadata and validate it is `binary_download`.
- Reuse existing auth/base-url/runtime construction and endpoint path interpolation safeguards.
- Write through a temp file in the destination directory, verify byte count before final rename, reject unsafe targets, and compute SHA-256.
- Update GitHub `operations.json`, `cli_surface.json`, `api_surface.json`, and `docs.md` only for release asset/repo archive and deferred upload notes.

### Refactor/checkpoint

- Keep helpers small and local to `engine`/`commandrunner` unless a broader safety package change is explicitly approved (not in this worker scope).
- Run `gofmt` after Go edits.
- Focused verification before commit/push.

## Verification checklist

Focused:

```bash
go test ./internal/connectors/engine -run 'TestBinaryDownload|TestBundleLoadEmbeddedGitHubOperations'
go test ./internal/connectors/commandrunner -run 'TestRunBinary|TestRunGitHubBinary'
go run ./cmd/connectorgen validate internal/connectors/defs/github
go test ./internal/connectors/commandrunner ./internal/connectors/engine
go build ./cmd/pm
```

Broader when green:

```bash
go test ./internal/connectors/...
make verify
```

## Commit/push checkpoints

1. Plan artifacts committed/pushed after initial planning.
2. Red tests committed only if useful; otherwise record red evidence in TDD ledger before implementation.
3. Green binary transfer implementation + metadata committed/pushed after focused verification.
4. Review-fix commits, if automated review finds actionable issues.

## Safety notes

- No secrets in fixtures/logs/output.
- No downloaded bytes printed to stdout by default; only JSON manifests are returned.
- No generic shell, generic raw HTTP write, archive extraction, or execution of downloaded content.
- Upload/file write operations remain deferred until reverse-ETL plan/preview/approval/execute policy is designed.
