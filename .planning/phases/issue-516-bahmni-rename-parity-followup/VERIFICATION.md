# Verification — Bahmni rename/parity follow-up

## Custody

- Branch: `feat/bahmni-docker-connector`
- Starting head: `e18d58adf`
- Existing no-mistakes run for this branch: `01KYDEZMPV6W6FB7MHJVJXBRE7`, cancelled.
- PR #533 exists and is open/unmerged; this follow-up updates it rather than merging or replacing it.

## Local gates

| Gate | Result |
| --- | --- |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | PASS — 548 connector(s), 0 findings |
| `go test ./internal/connectors/conformance -run 'TestConformance/bahmni$' -count=1` | PASS |
| `./pm docs validate --connectors-dir docs/connectors` | PASS |
| `./pm connectors inspect bahmni --json` | PASS — connector resolves under new id |
| `./pm bahmni` | PASS — bare connector namespace renders manual and exits 0 |
| `./pm connectors inspect bahmni-docker --json` | Expected failure — old connector id no longer resolves |
| `go test ./cmd/connectorgen ./internal/connectors/bundleregistry ./internal/connectors/conformance ./internal/cli -count=1` | PASS |
| `go vet ./...` | PASS |
| `go test -timeout 20m ./...` | PASS |
| `go build ./cmd/pm` | PASS (required for regenerated embedded connector docs) |
| `website: npm run gen:website-data` | PASS |

## GitHub issue updates

Updated with `gh-axi issue edit`:

- #516 — Bahmni connector CLI feature parity parent roadmap
- #517 — Bahmni: CLI surface metadata (CLI parity)
- #518 — Bahmni: help renderer (CLI parity)
- #519 — Bahmni: stream runner (CLI parity)
- #520 — Bahmni: API surface inventory + exclusion ledger (CLI parity)
- #521 — Bahmni: direct read (CLI parity)
- #522 — Bahmni: bounded binary and blocked-operation policy
- #523 — Bahmni: typed reverse-ETL writes
- #524 — Bahmni: typed POST read-query operation execution
- #525 — Bahmni: schema-gated top-level JSON array request bodies
- #526 — Bahmni: bounded typed multipart upload support

Verification:

- `gh-axi issue view 516` still lists subissues #517–#526.
- `gh-axi issue subissue list 516` returns the same 10 subissues in order.
- Stale issue check reported no `Bahmni Docker` title/body wording and no `defs/bahmni-docker` path references.

## Parity artifact

- `docs/migration/issue-516-bahmni-operation-parity.md` records official sources, issue hierarchy,
  operation-family mapping, implementation/blocked/exclusion path, validation evidence, remaining
  executable gaps, and preserved captain-owned decisions.

## Captain-owned decisions preserved

1. `phi-redaction-unbacked`: engine-level PHI field redaction is not implemented by this connector-only follow-up. This follow-up only makes the local file-path field match current redaction markers and removes/softens false PHI-redaction claims. Broad clinical PHI field redaction remains undecided.
2. `diagnoses-nullable-primary-key`: `diagnoses.existingObs` nullable primary-key semantics remain undecided unless the captain separately chooses to drop or replace that primary key.

## PM/no-mistakes gate

Pending after commit on the exact candidate head.
