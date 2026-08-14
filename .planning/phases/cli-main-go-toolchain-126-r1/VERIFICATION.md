# Verification checklist — cli-main-go-toolchain-126-r1

Status: **local verification complete**.

- [x] `go.mod` retains `go 1.25.4` and sets only `toolchain go1.26.6`.
- [x] All named workflow setup-go pins use `go-version: '1.26.6'`.
- [x] `security.yml`'s explicit `GOTOOLCHAIN` uses `go1.26.6`.
- [x] `GOTOOLCHAIN=go1.26.6 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` passes:
  `No vulnerabilities found.`
- [x] `git diff --stat origin/main...HEAD` is restricted to the active toolchain pin plus new
  mandatory GSD/TDD evidence; no historical `.planning` file changed.
- [x] `scripts/verify-gsd-workflow origin/main` accepts the delivery evidence.

## Commands run

```text
GOTOOLCHAIN=go1.26.6 go version
GOTOOLCHAIN=go1.26.6 go run golang.org/x/vuln/cmd/govulncheck@latest ./...
scripts/verify-gsd-workflow origin/main
```
