# Verification Checklist — Generic Certification Evidence Importer

- [x] Certified diagnosis rechecked against current integration-base code.
- [x] Focused `./cmd/connectorgen` tests exercise redaction, generic import, missing evidence, broken evidence, and definition-derived second shape.
- [x] Existing PostgreSQL accepted evidence is byte-identical: the 12 transport records and 2 CDC records have no diff from `db494bc8f`.
- [x] GitHub matrix remains at zero published capabilities. The PR names #4211 as the reason while recording that all 25 currently eligible commands executed successfully (23 REST and 2 GraphQL); no unverified record is committed.
- [x] Deliberate broken-evidence command result captured, followed by restoration: PostgreSQL CDC went red (`live_tested=false, records=0`) after its record key was broken, then green (`true, 1`) after restoration.
- [x] Secret-scanner negative control: a synthetic secret-shaped certification string made `connectorgen validate` fail with `[secret_literal]`; it was removed and the same command returned zero findings.
- [x] `make verify` passed: formatting, tidy, vet, the complete `go test -timeout 20m ./...` suite, build, docs, smoke, lint, agent contract, generator validation/surface checks, generated GitHub parity artifacts, certification matrix/sweep, boundary, connector canon, and release workflow checks.
- [x] Targeted evidence checks passed: `go test -timeout 20m ./cmd/connectorgen -run 'TestCertificationEvidence(Report|Postgres)' -count=1`; `go test -timeout 20m ./internal/connectors/certify -run 'Test(ReadExternalProof|WriteExternalProof)' -count=1`; `go test -timeout 20m ./internal/connectors/engine -count=1`; and `go test -timeout 20m ./internal/connectors/certify -count=1`.
- [x] Generated documentation was regenerated twice and byte-stable: `go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors`, plus `pnpm run gen:docs` and `pnpm run gen:website-data` from `website/`.
- [x] GSD inline `verify-work` and `code-review` fallback completed: no actionable implementation, accounting, path-safety, or secret-redaction finding remains. The external review route is `claude_auto` after the direct PR opens.
- [x] PR API base read-back equals `integration/4015-mvp-flat-r1`: `gh-axi pr list -R polymetrics-ai/cli --state open --base integration/4015-mvp-flat-r1 --head fm/cli-generic-evidence-importer-r1` returned only PR #4216; `gh-axi pr view 4216 --full` returned the complete issue-linked body.

## Local limitation

- `security/snyk` is not a local Make target in this checkout. Its failure is
  the known identical base-branch CI condition supplied by the delivery brief;
  this change neither invokes nor masks it.
