# Verification — Issue 3985 connector canon

**Result:** passed on 2026-08-10.

## Automated/local checks

| Check | Result |
| --- | --- |
| `shasum -a 256 -c data/CANON-MANIFEST.sha256` | Pass; all current and archived source reports pinned. |
| `make connector-canon-check` | Pass. |
| `make connector-runtime-preflight` | Pass; real commandrunner sweep. |
| `make github-parity-artifacts-check` | Pass; 14 Node tests plus both generated artifact checks. |
| `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/commandrunner` | Pass. |
| `go test -timeout 20m ./internal/cli` | Pass. |
| `go vet ./...` | Pass. |
| `make tidy-check`, `make lint`, `make docs-check` | Pass. |
| `pnpm --dir website typecheck`, `pnpm --dir website test:scripts` | Pass. |
| `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync` | Pass. |
| `make connector-boundary`, `make release-workflow-check`, `make smoke-no-build` | Pass. |
| `git diff --check` | Pass before commits. |

## REST verification

- `gh-axi issue subissue list 3972` read back **12 of 12** children, including newly attached #3987.
- `gh-axi issue view 3972 --full`, `gh-axi issue view 3978 --full`, and `gh-axi issue view 3987 --full`
  confirmed the four warehouse routes, seven-mode truth, and #3987 → #3978 gate.
- `gh-axi issue view <child> --comments` confirmed each warehouse-route amendment. Active #3974 was
  read and deliberately left unchanged.

## UAT classification

No interactive product behavior was introduced. The deliverables are documentation, deterministic
guards, archive state, and GitHub planning metadata; all have direct automated or REST evidence in
`SUMMARY.md`'s `coverage` block. Live provider certification remains explicitly out of scope and
zero, so it is not silently passed as UAT.

## Full-suite/no-mistakes note

Per repository guidance, this worker did not run `go test ./...` or `make verify` as a single local
command because they commonly exceed command timeouts. CI/no-mistakes owns the aggregate run. This
branch has been pushed; the supervisor can run the required no-mistakes/PR gate next.
