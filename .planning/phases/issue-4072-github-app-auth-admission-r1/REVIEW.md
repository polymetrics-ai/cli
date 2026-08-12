---
status: clean
phase: "issue-4072-github-app-auth-admission-r1"
depth: deep
files_reviewed: 5
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
reviewer: inline_manual
base: da8a8ff07aaf00e5c7965cd4d1d3c7252017d785
green_head: 3f83bf3afc6efa0ebc323e385e4345f588a41db1
---

# Deep code review — Issue #4072

`scripts/gsd prompt code-review issue-4072-github-app-auth-admission-r1 --depth=deep --files=internal/connectors/engine/auth.go,internal/connectors/engine/hooks.go,internal/connectors/engine/read.go,internal/connectors/hooks/github/hooks.go,internal/connectors/hooks/github/hooks_test.go`
was resolved. The canonical single-worker contract forbids spawning a reviewer,
so this is the required inline manual fallback.

## Scope reviewed

- `internal/connectors/engine/auth.go`
- `internal/connectors/engine/hooks.go`
- `internal/connectors/engine/read.go`
- `internal/connectors/hooks/github/hooks.go`
- `internal/connectors/hooks/github/hooks_test.go`

## Review findings

No source, security, concurrency, error-handling, or test-isolation findings
remain in the reviewed change.

- Runtime construction resolves the rate boundary before custom authentication,
  so `require_shared` fails before a network-capable GitHub App hook can run.
- The auth hook receives only an engine-owned declared-route capability; it
  cannot obtain the raw coordinator, runtime, or an arbitrary URL writer.
- GitHub's token exchange admits declared `POST /app/installations/{installation_id}/access_tokens` while the requester sends the interpolated, escaped installation path. The requester owns exactly one Decide/Finish lifecycle for the physical request.
- JWT-bearing headers are copied into a single requester clone. The
  secret-blind transport/coordinator fakes do not retain headers, body, JWT,
  private key, or minted token; error paths retain the existing redaction
  boundary.
- Existing bearer behavior, GitHub write-hook admission, ordinary declared
  requests, process-local admission, and the GraphQL exclusion remain covered.
- Tests do not run in parallel and restore the test transport with cleanup;
  the bounded race matrix passes.

## Evidence

```text
GOMAXPROCS=2 go test -p 1 ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1
GOMAXPROCS=2 go test -p 1 -race ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1
go vet ./internal/connectors/engine ./internal/connectors/hooks/github
```

All listed commands passed on 2026-08-12. Full repository validation and the
delivery/review service routes remain intentionally deferred to the shared
#4071 validation gate.
