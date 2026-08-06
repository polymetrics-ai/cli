# VERIFICATION — issues #3745 and #3746

Status: automated verification passed; inline GSD verify-work and code review completed.

## Checklist

- [x] RED tests executed before production implementation and actual failures captured in TDD ledger.
- [x] Descriptor types are closed; `CDCReader` does not satisfy the new capability-provider interface.
- [x] `changefeed.json` is optional and absent descriptors are non-capable.
- [x] PostgreSQL descriptor is evidence-backed unsupported with no executor claim.
- [x] Implemented catalog membership requires an implemented descriptor and matching executor.
- [x] `pm connectors catalog --capability cdc --json` excludes PostgreSQL.
- [x] `pm connectors inspect postgres --json` explains unsupported status safely.
- [x] No other connector is classified or modified.
- [x] No new dependency, credentials, provider calls, generic protocol escape hatch, or redaction path.
- [x] CLI help/manual/website work is explicitly deferred to #3748; no unrelated docs surface changes.
- [x] Focused tests, build, vet, applicable independent verify gates, boundary check, and diff check pass.
- [x] GSD verify and code-review prompts are executed inline and review dispositions recorded.

## Automated evidence

Focused regression and contract tests passed:

```text
go test ./internal/connectors
go test ./internal/connectors/engine
go test ./internal/connectors/native/postgres
go test ./internal/cli
go vet ./internal/connectors ./internal/connectors/engine ./internal/connectors/native/postgres ./internal/cli
go build ./cmd/pm
```

The no-credential runtime probes passed:

```text
go run ./cmd/pm connectors catalog --capability cdc --json
go run ./cmd/pm connectors inspect postgres --json
go run ./cmd/pm connectors
go run ./cmd/pm help connectors
go run ./cmd/pm connectors --help
```

The CDC catalog returned `count: 0`, and PostgreSQL inspect returned its top-level
`changefeed` descriptor with `status: unsupported`, `mechanism: logical_replication`,
the PostgreSQL documentation source, and its concise implementation reason.

Independent repository gates passed: `make tidy-check`, `make lint`, `make docs-check`,
`make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`,
`make connectorgen-surface-sync`, `make connector-boundary`, and
`make release-workflow-check`. `git diff --check` passed.

`go test ./...` and `make verify` were deliberately not run as monoliths because the
repository's per-command-timeout guidance assigns the 550-connector full suite to CI.

## GSD execution and review

- `scripts/gsd prompt verify-work issue-3745-3746-changefeed-discovery-r1` was generated and
  executed as the required inline Pi-adapter fallback. `SUMMARY.md` records three deterministic,
  automated deliverables and `issue-3745-3746-changefeed-discovery-r1-UAT.md` records the passed
  UAT verdict.
- `scripts/gsd prompt code-review issue-3745-3746-changefeed-discovery-r1` was generated and
  executed inline. `REVIEW.md` records the review scope and disposition. The review found two
  warning-level contract gaps, added their red tests, and resolved them before this verification.

## Scope and parity disposition

Only the optional descriptor contract, its loader/projection path, PostgreSQL's one
unsupported declaration, and focused tests changed. No other connector has been classified.
No command/flag/help text, docs, website data, command-runner, generator, conformance fixture,
dependency, credential, provider call, generic protocol surface, or redaction/masking path was
added. Human manual/help/website rendering and generated-surface enforcement remain owned by
#3748 and #3749 respectively.
