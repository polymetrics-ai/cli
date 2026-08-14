# TDD ledger — cli-main-go-toolchain-126-r1

Manual-GSD fallback: this worker generated the required `scripts/gsd prompt` commands and records
the equivalent single-worker red/green evidence here because Pi is unavailable and role spawning is
forbidden.

| Slice | Red test/proof | Green assertion | Status |
| --- | --- | --- | --- |
| Toolchain pin inventory | Scoped search reports the old `1.25.12` pins. | All active target pins report `1.26.6`. | Green |
| Vulnerability check | Go 1.25.12 is the documented failing baseline for `govulncheck`. | `govulncheck ./...` succeeds under Go 1.26.6. | Green |

## Red evidence log

Before the edit, the scoped source inventory found `1.25.12` in every target workflow and
`toolchain go1.25.12` in `go.mod`; `go 1.25.4` was separately recorded as the immutable language
directive. Main's required `govulncheck` check was failing on that old toolchain for the six listed
standard-library advisories.

## Green evidence log

```text
$ GOTOOLCHAIN=go1.26.6 go version
go version go1.26.6 darwin/arm64

$ GOTOOLCHAIN=go1.26.6 go run golang.org/x/vuln/cmd/govulncheck@latest ./...
No vulnerabilities found.
```

The scoped post-change inventory confirms every target workflow `go-version` and the security
job's `GOTOOLCHAIN` use `1.26.6`, while `go.mod` retains `go 1.25.4` and changes only its
`toolchain` directive.
