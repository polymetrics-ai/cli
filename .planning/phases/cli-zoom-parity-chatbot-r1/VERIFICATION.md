# Verification Checklist — Zoom Chatbot parity, R1

## Lifecycle

- [x] GSD command provenance was resolved with `scripts/gsd sources`.
- [x] Required skills and canonical connector/CLI references are recorded in `PLAN.md`.
- [x] Live provider artifact URL, retrieval date, HTTP result, byte count, SHA-256, and exact four-operation audit are recorded before RED.
- [x] RED failure is captured before production changes and committed/pushed (`17b4dfdf4`, `f2776f23f`, `19325b93b`).
- [x] Reusable OAuth Basic and closed typed JSON-object foundations are red/green tested, separately committed, and pushed (`c3038e29c`, `68dc984fe`).
- [x] The separate no-body direct-write foundation is RED/Green tested and separately committed/pushed (`b81cefb78`, `acbf7405c`).
- [x] The separate direct-write path-error redaction foundation is RED/Green tested and separately committed/pushed (`c9c89c707`, `070432f40`).
- [x] Connector declaration and generated output are staged for this coherent slice commit and will be pushed with it.
- [x] Inline verify-work and code-review evidence are complete.

## Source parity

- [x] All four Chatbot ledger rows are executable direct writes with exactly one disposition.
- [x] Zero Zoom rows are `unsafe_or_disallowed`.
- [x] Each command accepts only documented typed path/body members; no paging flags are invented.
- [x] Client-credentials Basic exchange and API Bearer application are declared and tested without raw secret output.
- [x] Link Unfurls treats HTTP `204 No Content` as a successful action.

## Runtime/docs checks

- [x] Focused engine/connsdk/commandrunner/app/Zoom/conformance/certify tests, vet, and lint pass.
- [x] A fresh binary passes base/group/command help and every isolated Chatbot plan lifecycle.
- [x] The isolated fixture proves token request Basic auth, action request Bearer auth, exact method/path/body/status, and redaction.
- [x] Surface sync/reconciliation/validation, endpoint-ledger scope, docs/website validation, and CLI golden checks pass.

## Inline verify-work record

The official GSD command sources were re-resolved before verification. The provider category is
not runtime-registered and the parent contract forbids role spawning, so this is the required
manual inline `verify-work` record rather than an absent automated worker.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
ok      polymetrics.ai/internal/connectors/defs/zoom

$ go test -count=1 -timeout 20m ./cmd/connectorgen ./internal/connectors/connsdk ./internal/connectors/commandrunner ./internal/app
ok      polymetrics.ai/cmd/connectorgen
ok      polymetrics.ai/internal/connectors/connsdk
ok      polymetrics.ai/internal/connectors/commandrunner
ok      polymetrics.ai/internal/app

$ go test -count=1 -timeout 20m ./internal/connectors/conformance ./internal/connectors/certify ./internal/connectors/boundary ./internal/cli
ok      polymetrics.ai/internal/connectors/conformance
ok      polymetrics.ai/internal/connectors/certify
ok      polymetrics.ai/internal/connectors/boundary
ok      polymetrics.ai/internal/cli

$ go vet ./...
$ make lint
0 issues.
$ make tidy-check agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check
... all gates passed
$ make smoke-no-build
smoke ok
```

The current checked declaration also passed `./pm docs validate --connectors-dir docs/connectors`,
`npm run typecheck` from `website/`, and `git diff --check`. `go run ./cmd/connectorgen
surface-sync --check` and scoped reconciliation were clean after generation. The docs generator
was run for the whole repository, then every unrelated generated connector file was restored as a
whole file; retained generated changes are only `docs/connectors/zoom/{MANUAL,SKILL}.md` and the
Zoom record in the two website connector catalogs.

Fresh compiled-binary reachability (all exit 0):

```text
$ go build ./cmd/pm
$ ./pm help zoom
$ ./pm zoom
$ ./pm zoom chatbot
$ ./pm zoom chatbot messages send --help
$ ./pm zoom chatbot messages edit --help
$ ./pm zoom chatbot messages delete --help
$ ./pm zoom chatbot link-unfurls create --help
```

## Inline code-review record

The manual review covered the exact operation/CLI mappings, closed request schemas, output policies,
redaction, operation-scoped Basic-to-Bearer auth, destructive confirmation, generated artifacts,
and the endpoint-ledger diff. It independently verified four—and only four—Chatbot
`covered_by.direct_write` rows and a zero `unsafe_or_disallowed` count. Canonically sorted JSON
hashes of the website connector catalogs with the Zoom record removed matched `HEAD`, proving their
generated diffs contain no unrelated connector changes.

One actionable review finding was resolved before this record: direct-write errors could retain a
declared path value inside a URL. Its separate red/green foundation is `c9c89c707` → `070432f40`.
The full engine and focused connector suites passed after that fix. No unresolved code-review
finding remains in this slice.
