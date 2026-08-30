# TDD ledger — CircleCI semantic lane repair R2

| Stage | Command / probe | Expected assertion | Result |
| --- | --- | --- | --- |
| Intake | `go run ./cmd/agentcontractgen check` | Canonical delivery contract is current | pass |
| Red | `go test ./internal/connectors/defs/circleci -run TestCircleCISourceWebhookRegistrationIsTheOnlySyncCandidate -count=1` before the two sync-cell corrections | Fails because `createWebhook` had `sync_transport=not_applicable` despite its source-proven delivery contract | pass: observed the expected failure |
| Green | focused semantic test set after matrix correction | Exact source semantics, counts, backlinks, and lanes satisfy the contract | pass |
| Edge: false positive | in-memory source mutations | `limit`, array, request-only paging input, and non-string cursor cannot promote ETL | pass |
| Edge: missing continuation | retained source reconciliation | A response cursor without provider request/link continuation is non-ETL | pass for six retained cases |
| Edge: sync | matrix/source mutation | Paging cannot become sync; actual outbound registration must retain a named source-backed missing-foundation cell | pass |
| Edge: mutation | source semantic read/mutation probes and full matrix validation | Every source mutation needs dual direct-write/reverse-ETL mapping, while bounded semantic reads do not | pass |
| Refactor | `gofmt`, JSON/count checks | Deterministic format and 111-row/777-cell retained denominator | pass |
