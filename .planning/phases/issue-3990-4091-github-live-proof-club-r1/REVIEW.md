---
status: clean
phase: issue-3990-4091-github-live-proof-club-r1
depth: standard
files_reviewed: 15
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
reviewer: inline-manual-fallback
---

# Code review — Issues 3990 and 4091: live GitHub certification proofs

## Scope

- `internal/connectors/certify/external_proof.go`
- `internal/connectors/certify/stages_glue.go`
- `internal/connectors/certify/stages_source.go`
- `internal/connectors/certify/stages_write.go`
- focused certification regressions
- `internal/cli/etl_transport.go`
- `internal/cli/etl_transport_test.go`
- `internal/cli/docs.go` and generated golden transcripts
- `internal/app/github_warehouse_transport_approval_test.go`
- checked-in CLI and website ETL documentation plus generated website data

The generated `code-review` prompt was executed inline because the canonical combined task is not a
numbered phase and this autonomous lane does not permit spawning the GSD reviewer role.

## Review checks

- Empty or unavailable live reads become explicit skips and cannot claim a successful round trip.
- Cleanup always previews before consuming its single-use approval, and delete cleanup carries the
  typed destructive confirmation required by the production CLI.
- External proof input remains globally bounded while applying its redirect/retry guard per HTTP
  method and target; refusal happens before artifact creation.
- Durable tokenless carriage remains restricted to a valid forward plan with explicit destructive
  confirmation. Cleanup still requires a new stdin token.
- Replay assertions require `AuthorizationTokenReplayError`, and live replay/401 evidence includes
  unchanged provider and checkpoint observations.
- Help, golden transcript, generated artifacts, and production-binary behavior agree. No credential,
  approval token, provider request target, or rendered rate scope is committed.

## Findings

None.
