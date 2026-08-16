# Verification checklist — GitHub live read coverage r1

- [x] Scope ledger reconciles all 1,571 declared commands and all 37 declared streams into explicit buckets; see `SCOPE.md`.
- [x] Early escalation is recorded: the current generic runner lacks a declaration-owned produced-value assertion and requires a foundation split before any live result can be called pass.
- [x] Candidate schema/configuration validation passes: `go run ./cmd/connectorgen validate internal/connectors/defs` and surface-sync `--check` are green.
- [x] Targeted Go certification and engine tests pass.
- [x] Scratch post-schema read break demonstrably fails; source was restored and revalidated before the final live run.
- [x] Compiled binary performed a live GitHub `--direct-read-only` certification: 23/23 declaration-owned candidates passed result-kind and output assertions; the artifact guard accepted the safe summary and temporary run root was discarded.
- [x] Rate-limit responses are honored and resumability is bound to a safe matching checkpoint: final run recorded 72 bounded admission events, candidates run serially, and checkpoint tests prove matching-only resumes are labelled rather than re-executed.
- [x] `go vet ./...`, scoped tests other than the consumer package, lint, docs/generator checks, generated-file checks, `connectorgen boundary`, surface sync, and the other `make verify` sub-gates are green.
- [ ] Before push, rebase after PostgreSQL PR #4199 lands, regenerate its affected matrix through `connectorgen`, then rerun `go test -timeout 20m ./cmd/connectorgen`. The current consumer failure reproduces unchanged at merge-base `404536538`, so it is pre-existing rather than caused by this branch.
- [x] CLI help/manual/website parity: `internal/cli/docs.go`, `docs/cli/connectors.md`, and `website/content/docs/cli-reference.mdx` document `--direct-read-only`; `./pm help connectors`, `./pm connectors`, and `./pm connectors certify --help` are green; `make docs-check` is green.
- [x] Inline code review found no unresolved actionable issue: reviewed checkpoint secrecy/fingerprint binding, no-value assertion errors, direct-only stage isolation, candidate boundaries, report wording, and fail-closed artifact rendering.
- [ ] PR base is API-read as `integration/4015-mvp-flat-r1`.
