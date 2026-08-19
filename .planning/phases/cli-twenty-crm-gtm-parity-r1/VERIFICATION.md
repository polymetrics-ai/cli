# Verification checklist — Twenty CRM

- [x] GSD command resolution and inline fallback recorded.
- [ ] Recovery assessment precedes bundle import.
- [x] Targeted Twenty test red/green evidence recorded in `TDD-LEDGER.md`.
- [x] Focused structural generator and conformance checks: `go run ./cmd/connectorgen validate internal/connectors/defs/twenty`, `go run ./cmd/connectorgen surface-sync --check`, and `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/twenty'`.
- [x] Structural generator, runtime preflight, conformance, boundary, docs/golden,
  lint, build, and verification results recorded with exact commands.
- [x] Historical built-binary live read, pagination, create, update, round-trip,
  delete, and post-delete absence evidence is retained without secret material.
- [x] PR #4298 was retargeted through the GitHub API to
  `fm/cli-reverse-etl-destination-r1`.
- [x] Typed-action eligibility red/green evidence is recorded in
  `TDD-LEDGER.md` and covers every action in `writes.json`.
- [ ] Final #4304 refresh: fetch and normally merge the updated foundation
  without rewriting history; prove its refreshed SHA is an ancestor; bind all
  55 required eligible actions (56 record-shaped bindings total); and exercise
  persisted App/CLI per-action dispatch before final push.
- [ ] Fresh dedicated-instance certification after that merge: built-binary
  authenticated list/get with pagination, lane-owned create/read-back/update/
  delete/verified cleanup, ETL source, reverse-ETL plan/apply/acknowledgement
  plus independent API read-back, and binary transfer only if documented.

## Historical live proof — superseded for final acceptance

The disposable published-Compose stack used the pinned image
`twentycrm/twenty@sha256:cb80b05bc2619a88a3a83293f45f2be495a55ac77a90946fa1f7d85f0b7fde24`.
The stack's `/healthz` returned 200. GraphQL `__schema` introspection was
disabled, so the current image's supported `workspace:seed:dev --light` command
created the disposable workspace instead; no UI and no captain-owned instance
were used.

Firstmate's former direct-stream decision avoided a Keychain truncation
observation. This historical proof is not reused for the current acceptance
gate. The refreshed proof must source the registered `pm-twenty`
Keychain/secret-store reference only at stdin injection, never from argv,
output, a file, a prompt, commit, or evidence. The immediate built-binary
`pm twenty companies list --credential twenty-live --limit 3 --json` returned
three real seeded records, proving the whole stored credential authenticated.

The actual built binary then passed these bounded live probes (all identifier
and token values withheld):

- `companies list --limit 1000` returned 600 records, proving cursor pagination
  crossed a provider page boundary. A separate `people list --limit 1000`
  returned 1,000 records.
- The 28 declared ETL object commands each completed a bounded `list --limit 1`;
  where seeded data existed, a following operation-backed `get --id` completed.
- A company created through the plan → approval-stdin → run flow was readable
  with its exact input name, then updated through the same flow and readable
  with the changed name.
- The same record was deleted through plan → preview → approval-stdin →
  `--confirm destructive`; the run reported one staged, one succeeded, zero
  failed. It was absent from the following list and its get returned provider
  404.

This is real-data runtime proof only; it does not widen the repository's
separate certification allowlist or claim `CERTIFIED` status for Twenty.

## Current dedicated-runtime discovery — blocked

Captain authorized a read-only local-runtime and protected-tree discovery on
2026-08-20. Docker reported two stopped Polymetrics PostgreSQL integration
containers, zero Compose projects, and no Twenty container. Its local image
store contains `twentycrm/twenty:latest` and `twentycrm/twenty-app-dev:latest`,
but neither image is deployed. Podman is installed but unavailable; both local
machines are stopped. The preserved `.twenty-live.SRsCLr/` tree contains one
non-executable JSON file and no compose, shell, environment, secret, credential,
stdin-handoff, or deployment marker. Its contents and all environment/inspect
values were intentionally not printed.

Therefore no currently running dedicated Twenty identity, isolation proof, or
non-exporting credential handoff is available. Do not create, start, inspect
with environment output, or repurpose any other local container until a
dedicated disposable runtime and an execution-time stdin/FD/approved-secret-store
credential reference are available. The refreshed #4304 head is also still
`c6f03c937`; it remains an ancestor, but it does not provide the required
persisted multi-action selection.

## Local verification results

All of these completed successfully unless explicitly marked as a documented
red-to-green checkpoint:

```text
go test -timeout 20m ./internal/connectors/defs/twenty
go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/twenty'
go test -timeout 20m ./internal/connectors/commandrunner -run 'TestEveryImplementedCommandPassesRuntimePreflight'
go test -timeout 20m ./internal/connectors
go test -timeout 20m ./cmd/iconregistrygen
go vet ./...
go build -o .twenty-live.SRsCLr/pm ./cmd/pm
go run ./cmd/connectorgen validate internal/connectors/defs/twenty
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen certification-sweep --connector twenty
go run ./cmd/connectorgen certification-sweep --connector twenty --check
make fmt
make tidy-check
make docs-check-no-build
make smoke-no-build
make lint
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-runtime-preflight
make connector-canon-check
make connector-boundary
make github-parity-artifacts-check
make connectorgen-certification-matrix
make connectorgen-certification-candidates
make connectorgen-certification-sweep
make release-workflow-check
scripts/verify-gsd-workflow
git diff --check
```

The expected initial red checkpoint was
`go test -timeout 20m ./internal/cli`: catalog count 556 and root-help golden
fixtures did not yet include Twenty. After regenerating exactly the nine
root-help transcripts from the binary and changing the catalog expectation to
557, the following green results were recorded:

```text
go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1
# PASS (17.729s)
go test -timeout 20m ./internal/cli
# PASS (1184.608s)
```

The subsequent GitHub `Verify` run `32274541836` supplied one further red
projection check: `internal/connectors/bundleregistry` loaded 553 definitions
but its historical assertion still expected 552. The fixed focused check is:

```text
go test -timeout 20m ./internal/connectors/bundleregistry
# PASS
```

After rebasing the committed branch onto the latest `origin/main`, the final
pre-push run passed again:

```text
go test -timeout 20m ./internal/cli
# PASS (596.926s)
go test -timeout 20m ./internal/connectors/defs/twenty
go run ./cmd/connectorgen validate internal/connectors/defs/twenty
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen certification-sweep --connector twenty --check
make docs-check-no-build
make smoke-no-build
make lint
make connector-boundary
go build ./cmd/pm
git diff --check
# PASS; boundary: 553 connectors loaded, zero findings
```

`make verify` was not invoked as a monolithic command because this repository's
agent guidance explicitly forbids it under a per-command timeout: it embeds
the whole suite and a timeout is indistinguishable from a hang. Its named
non-suite entry points above were run individually, and the full
`internal/cli` package was run with the repository-required 20-minute timeout.
The remaining repository-wide suite is CI coverage.

The post-build real-instance authentication check was run from the disposable
proof project:

```text
../pm twenty companies list --credential twenty-live --limit 3 --json | jq ...
# {"status":"passed","record_count":3}
```

The credential value, record values, record identifiers, approval material,
and endpoint configuration are intentionally absent from this evidence.

## Pull request

Created non-draft PR [#4298](https://github.com/polymetrics-ai/cli/pull/4298)
with `Refs #277`. The reconciliation API read-back returned base
`fm/cli-reverse-etl-destination-r1` and head
`fm/cli-twenty-crm-gtm-parity-r1`. A final API read-back is still required
after the fresh #4304 merge and before the next push; the generic
destination's persisted App/CLI dispatch is not yet claimed as deployable.
