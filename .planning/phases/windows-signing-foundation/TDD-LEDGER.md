# TDD Ledger — Windows signing foundation

## Rules

- Start with red tests or explicit validation artifacts before production edits.
- Do not call signing providers or require secrets in tests.
- Unsigned PR snapshots are validation artifacts only and are never releasable.

## Red / baseline evidence

| Slice | Red/baseline command | Expected red/baseline | Status |
|---|---|---|---|
| Current policy surface | `test -f docs/security/code-signing-policy.md` | Missing before docs slice | Pending |
| Version normalization | `go test ./build/windowsversion ./packaging/windows/winget` after adding tests before implementation | Failed: undefined `Version`, `NormalizeVersion`, `RenderRC` | Red captured |
| WinGet ID templates | `go test ./build/windowsversion ./packaging/windows/winget` after adding tests before templates | Failed: manifest templates missing | Red captured |
| Windows package CI | PR workflow dry-run on GitHub | Fails if unsigned build/package/install validation regresses | Pending |
| CI govulncheck | `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` | Failed with GO-2026-6061 in reachable `google.golang.org/grpc` v1.79.3; fixed in v1.82.1 | Red captured |
| PR issue guard | PR #559 `require-linked-issue` with narrative body references `issues #554 and #555` plus `reference #550/#554/#555` | Failed because only `Refs #N`/closing keywords and single issue tokens were recognized | Red captured from CI/PR metadata |
| Windows SDK tool discovery | PR #559 `unsigned-msi-snapshot` | Failed in `scripts/windows-versioninfo.ps1` because singleton PowerShell pipeline results do not expose `.Count` under `Set-StrictMode` | Red captured from CI |

## Green evidence

| Slice | Command | Expected | Result |
|---|---|---|---|
| Version generator | `go test ./build/windowsversion` | PASS | Passed via focused package run |
| WinGet templates | `go test ./packaging/windows/winget` | PASS | Passed via focused package run |
| WiX source guard | `go test ./packaging/windows` | PASS | Passed via focused package run |
| Go tests | `go test ./build/windowsversion ./packaging/windows ./packaging/windows/winget` | PASS | Passed |
| Formatting | `gofmt -w build/windowsversion packaging/windows` | no diff after gofmt | Passed |
| Broad Go gates | `go test ./...`, `go vet ./...`, `go build ./cmd/pm` | PASS or documented blocker | Passed locally |
| CI govulncheck fix | `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` | PASS | Passed after `google.golang.org/grpc` v1.82.1 update |
| PR issue guard fix | `go test -count=1 ./cmd/prissueguard ./internal/coordination/issueguard`; PR #559 title/body through `go run ./cmd/prissueguard` | PASS | Passed; actual PR body reports 3 linked issues |
| Windows SDK tool discovery fix | `go test -count=1 ./build/windowsversion ./packaging/windows ./packaging/windows/winget` plus source inspection | PASS or documented blocker | Passed focused Go tests; full PowerShell MSI build requires Windows runner |
| CI repair package set | `GOTOOLCHAIN=go1.25.12 go test -count=1 ./cmd/prissueguard ./internal/coordination/issueguard ./build/windowsversion ./packaging/windows ./packaging/windows/winget ./internal/runtimecheck ./internal/worker ./internal/connectors/native/postgres` | PASS | Passed |
| Bounded full suite rerun | `go test -timeout=4m ./...` | PASS or documented unrelated blocker | Timed out in existing broad connector/CLI tests while loading connector bundles; modified/fix-adjacent packages passed before timeout |
| PR-safe Windows package workflow | GitHub Actions `Windows Package Check` | PASS | Pending on PR CI |
| no-mistakes | `no-mistakes axi run --intent ...` | `checks-passed` | Pending |

## Deferred production-signing evidence

- SignPath signing request IDs, Authenticode verification, RFC 3161 timestamp verification, release upload, and WinGet external PR validation are intentionally deferred to provider-accepted follow-up work.
