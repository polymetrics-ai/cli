# Verification — reverse-ETL API-surface derivation r1

## Checklist

- [x] New invariant demonstrably fails on the pre-fix endpoint-like condition, then joins canonical API-surface evidence to recover the 214 true GitHub defects without guessing 34 punctuated Workday strings; the final invariant passes after regeneration.
- [x] Focused test proves generic endpoint-summary derivation and preserves alias-like human summaries.
- [x] GitHub has 214 recovered API-surface references and exactly 14 intentional empty aliases: `issue close`, `issue reopen`, `pr close`, `pr comment`, `pr lock`, `pr reopen`, `pr unlock`, `repo create`, `repo delete`, `repo archive`, `repo unarchive`, `secret set`, `secret delete`, and `cache delete`.
- [x] Generated sweep status buckets remain `1466 + 25 + 50 + 29 + 1 = 1571`, with zero delta per bucket.
- [x] Surface, certification sweep, website docs/data, and tracked-skills generators are byte-stable on their second relevant pass.
- [x] `go test -timeout 20m ./cmd/connectorgen`, `go test -timeout 20m ./internal/cli`, `go vet ./...`, `go build ./cmd/pm`, and all applicable individual repository gates pass.
- [x] Inline `verify-work` is complete. Local code review found no actionable issue; PR-open Claude automatic review is the selected external route.

## Commands and results

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./cmd/connectorgen -run '^TestImplementedEndpointSummaryAlwaysHasAPISurface$' -count=1` | Initial broad red: 248 endpoint-like strings (214 GitHub plus 34 punctuated Workday summaries); final canonical-endpoint invariant passes after generated synchronization. |
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
| `go test -timeout 20m ./internal/cli` | Pass. |
| `gofmt -w cmd internal && git diff --exit-code -- cmd internal` | Pass; formatting left no diff. |
| `go vet ./...` | Pass. |
| `go build ./cmd/pm` | Pass. |
| `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build` | Pass. |
| `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync` | Pass. |
| `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`, `make connectorgen-certification-sweep` | Pass. The matrix reported `connectors=2 capability_complete=0 certified=0`; the GitHub sweep reported 1,571 commands. |
| `make connector-boundary`, `make connector-runtime-preflight`, `make connector-canon-check`, `make release-workflow-check` | Pass. Boundary reported 290 checked shared files, 552 loaded connectors, and no findings. |
| `./pm help github >/dev/null && ./pm github >/dev/null` | Pass; no CLI help or behavior changed. |
| `git diff --check origin/integration/4015-mvp-flat-r1...HEAD` | Pass. |

`make verify` was deliberately decomposed into its repository-prescribed individual gates because the repository instructions warn that the aggregate suite exceeds the per-command timeout. No gate was skipped; the changed package and `internal/cli` were run separately with the required 20-minute timeout.
