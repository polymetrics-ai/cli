# Context — CI toolchain and integration branch guard

## Locked decisions

- Scope is limited to the stale Go toolchain pin and the `branch-name` guard.
- `govulncheck` must be fixed by a supported Go patch upgrade, never a suppression, allow-list, or skipped check.
- The first candidate fixed patch is verified locally. `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` reports seven reachable standard-library vulnerabilities; each is fixed in Go 1.25.13. The same scan under Go 1.25.13 passes.
- All active workflow pins and `go.mod`'s `toolchain` directive must agree. Historical planning evidence is an audit record and is not rewritten.
- `.github/workflows/conventions.yml` owns the `branch-name` check. The new integration class must require a positive decimal issue number and retain the existing lower-case slug grammar; no named-branch exception and no non-blocking job are allowed.

## GSD execution fallback

`scripts/gsd doctor`, all required command-source resolutions, and `go run ./cmd/agentcontractgen check` passed. This issue-first CI maintenance slice is not a numbered roadmap phase, and this direct-PR worker cannot launch compatible isolated Pi workers. The generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts are therefore executed inline and recorded in this directory.

## Skills used

`golang-how-to`, `golang-continuous-integration`, `golang-security`, `golang-lint`, and `golang-testing`.

## Non-applicable parity

No CLI command, flag, help topic, manual, generated artifact, or website surface changes.
