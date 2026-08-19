# VERIFICATION — Issue #3898: direct-read page context

## Goal-backward check

**Does a direct read still report a completeness it cannot prove?** No.

## Evidence

Fixture-backed runs of the real `pm` binary (local `127.0.0.1` fixture serving
120 records, provider default page 30):

```
github  pulls files view              -> records: 100, complete: false, next_number: 2
github  pulls files view --page 2     -> records:  20, complete: true          (100 + 20 = 120)
gong    logs list                     -> records:  30, complete: false, next_cursor: 30
gong    logs list --page-cursor 30/60/90 -> 30 + 30 + 30, final complete: true (120 total)
notion  comment list                  -> records: 100, complete: false, next_cursor: 100
```

gong returns 30 per page because its bundle declares no `size_param`. That is
correct derived behaviour, and it is now reported explicitly instead of being
silent — raising it is a one-line bundle declaration, not engine work.

Human-readable path: stdout stays clean JSON, the notice goes to stderr —
`note: page 1 of a paged result (100 records); more remain — rerun with --page 2`.

## Gates

```
gofmt -w changed Go files                  clean
go test -timeout 20m ./cmd/connectorgen   ok
go test -timeout 20m ./internal/connectors/engine  ok
go test -timeout 20m ./internal/connectors/defs/crisp  ok
go vet ./...                               clean
go build ./cmd/pm                          clean
git diff --check                           clean
go test -timeout 20m ./internal/cli        ok (full suite, 584.394s)

Focused acceptance regressions:
  cmd/connectorgen cursor/parameter suite  ok
  internal/cli GitHub wire/page-context     ok
  Crisp --per-page wire/content             ok
  engine caller-size/completeness suite     ok
  connectorgen surface-sync --check         551 scanned, 0 changes

npm --prefix website run gen:catalog        regenerated 551 connector records
npm --prefix website run gen:connectors     regenerated connector type data
npm --prefix website run gen:docs           current
node website/scripts/cli-surface.test.mjs   7/7 pass
npm --prefix website run typecheck           not run: tsc is not installed in this worktree
```

## Inline GSD verify-work and code-review fallback

`scripts/gsd sources` and `scripts/gsd prompt` were resolved for both
`verify-work` and `code-review`. This session cannot invoke the project-local
Pi workers, and the canonical issue contract forbids substituting spawned
planner/reviewer roles, so the documented inline fallback is used and recorded
here.

Verify-work accepted the automated fixture counts, server-observed parameter
assertions, generated-help checks, full CLI suite, and the captain-authorized
live PM/`gh-axi` evidence above. No human-judgment-only product assertion
remains: the one intentional product boundary (no local clone) is explicitly
declared and the archive redirect failure is explicitly recorded as a separate
gap rather than treated as a success.

Manual code review covered the cursor cleaner's scope, the legacy no-operation
branch, ambiguous `after`/`before` semantics, page-size preservation, the
fixture count assertions, and generated artifacts. It found and corrected one
coverage gap during review: legacy Gong `logs list` returned before the
operation metadata gate, so a new red/green regression moves the cursor cleanup
before that gate. Final disposition: no unresolved critical or warning finding.

## Known, stated plainly

- The original 30-record loss was **shared**, not GitHub-only: real fixture
  command runs across GitHub, Gong, and Notion established that every direct
  read used the shared one-page executor. Direct reads now return the declared
  page and explicit context (`records`, `size`, `has_more`, next address, and
  `complete`), so a caller can distinguish a complete small result from a
  bounded page with more records.
- The strict mutation transport had a separate ALPN defect: it forced HTTP/1
  while its TLS configuration advertised `h2`. Removing only the forced
  protocol restored honest negotiation; fresh connections, closed requests,
  no request replay, zero requester retries, and redirect refusal stay intact.
- Captain-authorized GitHub validation is now green. Through PM reverse ETL in
  the dedicated private repository, issue creation, issue comment, creation of
  the disposable deletion target, and deletion all completed; each requested
  state transition was independently checked with read-only `gh-axi`. A PM
  GitHub ETL run then read five issue rows, including the created issue.
- The parameter contract is proved on actual returned fixture rows, not exit
  status: invalid enums and missing required path parameters make zero
  requests; `since` and `--per-page 37` reach the wire; raw opaque cursor flags
  are absent from **all implemented direct reads** (including legacy Gong); and
  100 + 20 returned rows expose incomplete then complete page context and total
  120.
- GitHub local clone is a deliberate unsupported-local boundary, not a hidden
  shell escape hatch: `repo clone` blocks before dispatch with
  `intent=local_workflow`, `availability=unsupported_local`. Supporting it
  would require a separately designed, validated local-git/filesystem
  operation; no implementation was added here.
- Content-path audit: PM `repo read-file` returned HTTP 200 metadata and
  intentionally redacted file bytes/URL; PM release-asset list and individual
  asset metadata calls returned HTTP 200. The declared release-to-disk command
  is also intentionally unsupported-local. The declared tarball and zipball
  binary commands currently fail live because GitHub redirects to
  `codeload.github.com`, which is not allowlisted by the binary downloader.
  That is a separate, unmodified gap recorded for captain decision; no archive
  file was written.
- PR #3902's body was corrected on 2026-08-08 under the captain's explicit
  authorization for the owner PR: it now states the shared
  13-connector/362-command scope, the explicit page-context user behavior, the
  completed PM live-write validation, and the binary-path gap. The PR branch
  was then refreshed against `main`; no GitHub content outside the dedicated
  private test repository was changed.

## Post-main-refresh CI recovery

The refreshed-head Verify run exposed two pre-Parquet, branch-owned reverse
plan fixtures in `internal/app/reverse_confirmation_test.go`. Their JSONL
tables were correctly refused by the new warehouse contract; the product code
was not changed. The fixture setup now calls the existing
`seedWarehouseTableRows` helper, which uses `warehouse.WriteTable` to create
the same Parquet format the runtime reads. The two focused tests, the full
`internal/app` package, and `go vet ./internal/app` are green locally. A
whole-test-tree scan found no other ordinary legacy JSONL warehouse-table
fixture; the remaining references are deliberate refusal coverage. Inline
manual code review confirmed the diff changes only those fixtures and the
required GSD/TDD evidence. GitHub Verify is rerun from the committed fix.
