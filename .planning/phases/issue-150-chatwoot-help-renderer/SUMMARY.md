# Summary: Chatwoot Help Renderer And Docs Parity

Status: merged to parent branch.

- Created #150 GSD/TDD plan and verification checklist before production edits.
- Added Chatwoot runtime manual coverage for command-surface rendering.
- Regenerated checked-in Chatwoot connector manual/skill so `COMMAND SURFACE` is present.
- Regenerated website connector data so `/docs/connectors/chatwoot/data.json` exposes `cliSurface`.
- Updated website CLI reference copy to mention Chatwoot command metadata.
- Fixed command-surface flag punctuation and empty stream/write description trailing whitespace in guide rendering.
- Scope stayed docs/help/website only; no runtime `pm chatwoot ...` dispatch or new writes/reads were added.
- Verification passed: targeted Go tests, website unit/typecheck/build, docs validation, connectorgen validation, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify`.
- PR #240 remote checks passed and was squash-merged into parent commit `80db5020b297f1323f94c0c965f4a80ab6b08eb3`.
- CodeRabbit skipped PR #240 because the base branch was non-default; parent PR #223 manual CodeRabbit review was requested after integration and replied `Review finished` with no inline findings returned by GitHub API.
