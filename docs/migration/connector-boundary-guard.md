# Connector boundary guard

`connectorgen boundary` is a developer guard for connector-definition ownership. It scans shared production Go for connector-specific policy that should live in `internal/connectors/defs/<connector>/`, `internal/connectors/hooks/<connector>/`, or `internal/connectors/native/<connector>/`.

```bash
go run ./cmd/connectorgen boundary .
go run ./cmd/connectorgen boundary . --json
go run ./cmd/connectorgen boundary . --base origin/main
make connector-boundary
```

`--base <ref>` limits the primary scan to Go files changed from that git ref, including untracked
files, while exception-ledger contracts are still checked against the whole tree.

## Connector implementation ownership

`connectorgen ownership` validates changed paths for connector implementation PRs. It requires one
connector target, either from a machine-readable scope file or inferred from changed connector-owned
paths so label/tag omission cannot skip the check.

```bash
go run ./cmd/connectorgen ownership . --base origin/main --scope-file connector-scope.json
go run ./cmd/connectorgen ownership . --base origin/main --json --scope-file connector-scope.json
```

Scope file contract:

```json
{
  "api_version": "polymetrics.ai/v1",
  "kind": "ConnectorImplementationScope",
  "connectors": ["target-slug"]
}
```

The `connectors` array must contain exactly one connector slug. The validator allows target
`internal/connectors/defs/<slug>/`, target hooks/native/legacy connector paths, target generated
connector docs/icons/manual outputs, connector-owned test files that follow the
`<slug>_..._test.go` naming convention directly inside `cmd/connectorgen/` or
`internal/connectors/engine/` (hyphens in the slug become underscores, so `google-ads` owns
`google_ads_..._test.go`), and a narrow set of shared generated indexes/goldens (including
`docs/cli/connectors.md`, `docs/cli/reverse.md`,
`docs/connectors/catalog/all-connectors.{json,md}`, and `website/lib/docs.generated.ts`; that
allowlist is literal paths only, so other `docs/cli/` pages stay rejected as shared docs).

Every `.planning/phases/**` path is ignored rather than matched against the target, so a lane
commits the GSD plan/TDD/verification artifacts `AGENTS.md` requires without registering its phase
directory with the guard.

The validator rejects shared runtime/tooling, unrelated connectors, unrelated generated
docs/website churn, and guardrail exception/config edits — the guard's own files under
`internal/connectors/boundary/`, `cmd/connectorgen/ownership.go`,
`cmd/connectorgen/ownership_test.go`, `cmd/connectorgen/boundary.go`, this doc, and required-check
workflow files. This optional single-connector guard is not the cohort-delivery admission rule.

`ownership` therefore only applies to optional single-connector implementation lanes. A bounded
cohort PR that deliberately spans connectors or edits a named shared foundation is instead admitted
by its immutable source-lock ledger, per-connector ownership/path matrix, changed-path review, and
Foundation Atlas disposition. A non-zero exit from `ownership` for that cohort means this optional
single-lane guard does not classify the PR; it is not a cohort-delivery defect. That is also why it
is wired into neither `make verify` nor any workflow, unlike
`connectorgen boundary` (`make connector-boundary`, `.github/workflows/connector-boundary.yml`),
which applies to every PR. Run `ownership` by hand on connector lanes; do not add it to a blanket
gate until it can recognise a foundation PR.

## Output and exit status

- Exit `0`: clean boundary report.
- Exit `1`: policy violations or exception-ledger contract failures.
- Exit `2`: invalid invocation or scanner configuration.

JSON output uses the `ConnectorBoundaryReport` envelope with stable arrays for `findings`, `warnings`, and applied `exceptions`.

## Exception ledger

The baseline ledger is `internal/connectors/boundary/exceptions.json`. Each row must bind an exact `rule`, `connector`, `path`, `match`, `reason`, `migration_issue_url`, `owner`, `expires_on`, and `max_matches`.

The guard fails when an exception expires, stops matching, or matches more findings than `max_matches`. Free-form `approved_by` prose is ignored and does not count as approval.

## Review disposition

When the guard fails, prefer moving behavior into connector definitions. Add an exception only for current residue with a public follow-up issue, an owner, an expiry, and a bounded match count. This guard does not need connector credentials or live provider calls.
