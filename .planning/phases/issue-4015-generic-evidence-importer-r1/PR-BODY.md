Refs #4015

Related finding: #4211

## Summary

- Adds `connectorgen certification-evidence report`: a connector-agnostic
  importer from a completed certification report plus its redacted external
  proof into accepted evidence records. Connector-owned `evidence_import`
  bindings supply the provider and matrix identities; shared Go contains no
  connector-name branch or boundary allowlist change.
- Replaces the database transport importer's hard-coded provider with
  definition-owned database proof metadata. The twelve existing PostgreSQL
  transport records and the two later CDC records are byte-identical to
  `db494bc8f`.
- Proves the same report importer with GitHub and a second, independently
  declared Xero shape. It validates all bindings before opening any record,
  binds the proof to the report's fresh-child run ID, and requires the report's
  actual passing `full_parity` stage before an accepted record can be written.

## Honest GitHub scope

GitHub publishes **0** live-tested capability cells in this PR. This is
intentional and honest: no GitHub accepted-evidence records are committed.

The disposable GitHub App read-only runs did execute and pass all **25**
currently `eligible_pending_live` commands: **23 REST reads** and **2 GraphQL
queries**. Their reports do not contain a passing `full_parity` stage. The
accepted-record contract currently has only `credential_scope: full_parity`,
so importing them would claim a credential scope that the report did not
verify. The importer therefore refuses those reports before opening a record.

This is the separate #4211 finding, not a claim that the 1,571-command gate
has closed. Current GitHub ledger counts are:

| Classification | Count |
| --- | ---: |
| Published/certified in this PR | 0 |
| Executed and passed, withheld pending #4211 | 25 |
| Provider-refused | 1 |
| Fixture-required | 1,466 |
| Schema-conformant (not live) | 29 |
| Not applicable | 50 |
| Product defects | 0 |
| Declared total | 1,571 |

The one provider refusal is `actions fork-pr-contributor-approval view`:
GitHub returned HTTP 422 because that setting does not apply to the disposable
fixture repository. No write wave ran for this PR, so no provider resource was
created and no external cleanup residue exists.

The #4211 scope finding also applies to existing PostgreSQL records: the 12
transport records originated in a real `--full --write` built-binary run with
independent read-back, but not an explicit `--full-parity` run; its writer
stamped the full-parity credential field unconditionally. The two later CDC
records use the same unconditional writer. Their real execution evidence is
not in dispute; the record-level credential-scope assertion is unverified.

## Negative controls

- The importer test plants one known value in a request authorization header,
  URL query, and response body; emitted evidence contains no value and its
  response body consists only of fingerprint markers.
- A synthetic secret-shaped value planted in a valid certification definition
  made `connectorgen validate` fail with `[secret_literal]`; restoring it made
  the same check pass.
- A capability with no accepted evidence remains not live-tested in the
  GitHub matrix test.
- Deliberately changing PostgreSQL CDC evidence from `capability:cdc` to
  `capability:read` regenerated its own cell as
  `live_tested=false, records=0`; restoration regenerated it as
  `live_tested=true, records=1`.

## Verification

- `make verify` — pass, including the complete `go test -timeout 20m ./...` suite and all generator, boundary, docs, lint, release, and certification gates.
- `go test -timeout 20m ./cmd/connectorgen -count=1`
- `go test -timeout 20m ./internal/connectors/certify -run 'Test(Write|Read)ExternalProof' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs/github`
- `go run ./cmd/connectorgen certification-sweep . --connector github --check`
- `go run ./cmd/connectorgen certification-matrix --connector postgres`
- `go run ./cmd/connectorgen certification-matrix --connector github`
- `git diff --exit-code db494bc8f -- internal/connectors/certifications/evidence`
- `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors` twice
- `pnpm run gen:docs` twice and `pnpm run gen:website-data` twice from `website/`

The repository's `security/snyk` check is CI-only here and has the known
identical base-branch failure stated in the delivery brief; this PR does not
change or mask it. GSD ran as the recorded inline/manual fallback because this
direct PR has no roadmap phase; the required Go skills were loaded. Claude
automatic review is the post-open review route.
