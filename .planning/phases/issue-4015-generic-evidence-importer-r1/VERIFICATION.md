# Verification Checklist — Generic Certification Evidence Importer

- [ ] Certified diagnosis rechecked against current integration-base code.
- [ ] Focused `./cmd/connectorgen` tests exercise redaction, generic import, missing evidence, broken evidence, and definition-derived second shape.
- [ ] Existing PostgreSQL accepted evidence is byte-identical.
- [ ] GitHub matrix has a non-zero, honest, exact live-tested count and the PR scope ledger states what remains.
- [ ] Deliberate broken-evidence command result captured, followed by restoration result.
- [ ] `gofmt`, `go vet`, `connectorgen validate`, `surface-sync --check`, `connectorgen boundary`, docs/generated checks, and applicable lint pass.
- [ ] GSD inline `verify-work` and `code-review` fallback results recorded.
- [ ] PR API base read-back equals `integration/4015-mvp-flat-r1`.
