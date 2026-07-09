# Summary: Chatwoot Help Renderer And Docs Parity

Status: implemented and locally verified.

- Created #150 GSD/TDD plan and verification checklist before production edits.
- Added Chatwoot runtime manual coverage for command-surface rendering.
- Regenerated checked-in Chatwoot connector manual/skill so `COMMAND SURFACE` is present.
- Regenerated website connector data so `/docs/connectors/chatwoot/data.json` exposes `cliSurface`.
- Updated website CLI reference copy to mention Chatwoot command metadata.
- Fixed command-surface flag punctuation and empty stream/write description trailing whitespace in guide rendering.
- Scope stayed docs/help/website only; no runtime `pm chatwoot ...` dispatch or new writes/reads were added.
- Verification passed: targeted Go tests, website unit/typecheck/build, docs validation, connectorgen validation, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, and `make verify`.
