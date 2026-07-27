# TDD Ledger — Windows signing foundation

## Rules

- Start with red tests or explicit validation artifacts before production edits.
- Do not call signing providers or require secrets in tests.
- Unsigned PR snapshots are validation artifacts only and are never releasable.

## Red / baseline evidence

| Slice | Red/baseline command | Expected red/baseline | Status |
|---|---|---|---|
| Current policy surface | `test -f docs/security/code-signing-policy.md` | Missing before docs slice | Pending |
| Version normalization | `go test ./build/windowsversion` after adding tests before implementation | Fails until generator is implemented | Pending |
| WinGet ID templates | `go test ./packaging/windows/winget` after adding tests before templates | Fails until templates exist | Pending |
| Windows package CI | PR workflow dry-run on GitHub | Fails if unsigned build/package/install validation regresses | Pending |

## Green evidence

| Slice | Command | Expected | Result |
|---|---|---|---|
| Version generator | `go test ./build/windowsversion` | PASS | Pending |
| WinGet templates | `go test ./packaging/windows/winget` | PASS | Pending |
| Go tests | `go test ./build/windowsversion ./packaging/windows/winget` | PASS | Pending |
| Formatting | `gofmt -w build/windowsversion packaging/windows/winget` | no diff after gofmt | Pending |
| Broad Go gates | `go test ./...`, `go vet ./...`, `go build ./cmd/pm` | PASS or documented blocker | Pending |
| PR-safe Windows package workflow | GitHub Actions `Windows Package Check` | PASS | Pending |
| no-mistakes | `no-mistakes axi run --intent ...` | `checks-passed` | Pending |

## Deferred production-signing evidence

- SignPath signing request IDs, Authenticode verification, RFC 3161 timestamp verification, release upload, and WinGet external PR validation are intentionally deferred to provider-accepted follow-up work.
