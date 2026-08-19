# Verification checklist — issue 4261

## Before push

- [x] `go test -timeout 20m ./cmd/connectorgen -run 'TestCertification'` — pass, 102.189s
- [x] `go run ./cmd/connectorgen certification-sweep --connector github --check`
- [x] `go run ./cmd/connectorgen certification-matrix --connector github --check` and global `--check`
- [x] authorized GitHub live proof and evidence validation — one read-only
  candidate, HTTP 200, redacted schema-v2 record, and scoped matrix check passed
- [x] complete changed-package tests — `go test -count=1 -timeout 20m ./cmd/connectorgen` pass, 90.215s; `go test -count=1 -timeout 20m ./internal/cli` pass, 527.280s
- [x] `git diff --check`
- [x] release-metadata and trace assertions are empty before commit; repeat with the final merge-base diff after commit
- [x] no deletion below `.planning/traces/` relative to `origin/main`
- [x] pre-push hook delegates to `make verify` through worktree-local
  `core.hooksPath=.githooks`
- [x] `make verify` — pass (full suite and every required gate)
- [x] inline review found no actionable issue; PR route is `claude_auto`

## PR read-back

- [ ] GitHub API reports PR base exactly `main`.
- [ ] Changed-file count is at most 30 and there are no broad deletions.
- [ ] Required CI checks are recorded with their conclusions; no merge is performed.
