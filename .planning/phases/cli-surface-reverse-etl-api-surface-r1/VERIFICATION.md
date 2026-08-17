# Verification — reverse-ETL API-surface derivation r1

## Checklist

- [x] New invariant demonstrably fails on the pre-fix condition (248 total: 214 GitHub, 34 Workday) and passes after regeneration.
- [x] Focused test proves generic endpoint-summary derivation and preserves alias-like human summaries.
- [x] GitHub has 214 recovered API-surface references and exactly 14 intentional empty aliases: `issue close`, `issue reopen`, `pr close`, `pr comment`, `pr lock`, `pr reopen`, `pr unlock`, `repo create`, `repo delete`, `repo archive`, `repo unarchive`, `secret set`, `secret delete`, and `cache delete`.
- [x] Generated sweep status buckets remain `1466 + 25 + 50 + 29 + 1 = 1571`, with zero delta per bucket.
- [x] Surface, certification sweep, website docs/data, and tracked-skills generators are byte-stable on their second relevant pass.
- [ ] `go test -timeout 20m ./cmd/connectorgen`, `go vet ./...`, `go build ./cmd/pm`, and applicable individual repository gates pass or an exact external/base blocker is recorded.
- [ ] Inline `verify-work` and code review are complete and recorded.

## Commands and results

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./cmd/connectorgen -run '^TestImplementedEndpointSummaryAlwaysHasAPISurface$' -count=1` | Red before the fix: 248 findings (214 GitHub and 34 Workday); green after generated synchronization. |
| `go test -timeout 20m ./cmd/connectorgen -run '^TestSyncBundleDerivesAPISurfaceFromEndpointSummary$' -count=1` | Red before the fix (`api surface fills = 0`); green after generic derivation. |
| `go test -timeout 20m ./cmd/connectorgen` | Pass. |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | Pass. |
| `go run ./cmd/connectorgen surface-sync --check` | Pass after a second surface-sync reports 0 fills and 0 corrections. |
| `go run ./cmd/connectorgen certification-sweep --connector github --check` | Pass; regenerated ledger has 1,571 commands. |
| `pnpm --dir website run gen:docs` ×2 | Both passes succeeded; `git diff -- website docs` SHA-256 stayed the empty-diff hash. |
| `pnpm --dir website run gen:website-data` ×2 | Both passes succeeded; `git diff -- website` SHA-256 stayed the empty-diff hash. |
| `go run ./cmd/pm skills generate --dir docs/skills --json` ×2 | Both passes succeeded; `git diff -- docs/skills` SHA-256 stayed the empty-diff hash. |
| `go run ./cmd/connectorgen certification-sweep --connector github` ×2 | Both generated passes produced identical diff SHA-256 `2299f51c275762bb5dd9e2ea4928c02a38b662780c09c4a839ddcc790a5ece91`. |
| `go run ./cmd/connectorgen surface-sync` ×2 | Both generated passes produced identical diff SHA-256 `ec7e049f5bdfb00e5aee817aa3c8078e797b42807491b9f189327305695e64e4`; second run reports 0 fields. |
