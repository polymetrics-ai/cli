# SUMMARY — Wave 1 parallel batch 1 (Workflow connector factory)

Status: **completed** (GO). 6/6 connectors green; `make verify` green (independently verified).

## What ran
A Workflow (`wf_e618a486-872`) fanned out **6 universal-loop agents in parallel**, one per connector,
each writing ONLY its own `internal/connectors/<name>/` dir with strict TDD (httptest, red-first),
then a **convergence agent** ran `go run ./cmd/registrygen` + `gofmt` + `make verify`, quarantining
any failure. ~542k subagent tokens, 7 agents, ~7 min wall-clock.

## Enabler built first (made safe parallelism possible)
`cmd/registrygen/main.go` — derives `internal/connectors/registryset/registry_gen.go` by scanning
connector dirs, so parallel agents never touch the shared registry file (the collision that would
otherwise corrupt a naive fan-out). Each connector self-registers via `init()→RegisterFactory`.

## Connectors delivered (all green, on connsdk)
| Connector | Caps | Notes |
|---|---|---|
| slack | read | Bearer; ok:false-aware cursor pagination; users/channels/messages |
| hubspot | read+write | Bearer; after-cursor; contacts/companies/deals/tickets; create/update_contact (allow-listed) |
| notion | read | Bearer + Notion-Version header; POST /search cursor; databases/pages/users |
| jira | read | Basic(email,api_token); startAt/maxResults; issues/projects/users |
| sendgrid | read | Bearer; marketing lists/segments/contacts/bounces |
| postgres | read | pgx (already in go.mod); snapshot + cursor-incremental; **CDC = documented stub** (pglogrepl gated); fixture mode for credential-free tests |

## Verification (independent of the workflow)
- `make verify` exit 0; all 6 per-connector test suites pass; registrygen idempotent (8 imports:
  github, stripe + the 6 new). No `_quarantine/`.
- go.mod unchanged (no new deps; **pglogrepl absent** → CDC correctly stubbed).
- `pm connectors list` → 8 per-system connectors registered; `inspect <name> --json` → kind Connector,
  capabilities as above.

## Boundary / deferred
- **Catalog not flipped**: `source-slack`/`source-hubspot`/… remain `planned_native_port` in
  catalog_data.json (connectors are reachable by bare name via the registry regardless). A deliberate
  batched catalog-enablement pass (with the conformance-count assertion update) is a follow-up.
- Postgres full logical-replication **CDC** needs `pglogrepl` (dependency human-gate) — staged.
- Read-only connectors have no reverse-ETL writes yet (added where the API has sensible mutations).

## Significance
Proves the **parallel connector factory**: the same Workflow scales to large batches (concurrency
cap ~16, up to 4096 specs/run). HTTP connectors fan out freely (no deps); DB-CDC/cloud/file families
run as smaller gated batches.
