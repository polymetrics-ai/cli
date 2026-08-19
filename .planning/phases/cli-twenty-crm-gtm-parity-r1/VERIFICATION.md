# Verification checklist — Twenty CRM

- [x] GSD command resolution and inline fallback recorded.
- [ ] Recovery assessment precedes bundle import.
- [x] Targeted Twenty test red/green evidence recorded in `TDD-LEDGER.md`.
- [x] Focused structural generator and conformance checks: `go run ./cmd/connectorgen validate internal/connectors/defs/twenty`, `go run ./cmd/connectorgen surface-sync --check`, and `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/twenty'`.
- [x] Structural generator, runtime preflight, conformance, boundary, docs/golden,
  lint, build, and verification results recorded with exact commands.
- [x] Built-binary live read, pagination, create, update, round-trip, delete,
  and post-delete absence evidence recorded without secret material.
- [ ] PR base is read back through GitHub API and equals `main` (after PR creation).

## Live proof — passed

The disposable published-Compose stack used the pinned image
`twentycrm/twenty@sha256:cb80b05bc2619a88a3a83293f45f2be495a55ac77a90946fa1f7d85f0b7fde24`.
The stack's `/healthz` returned 200. GraphQL `__schema` introspection was
disabled, so the current image's supported `workspace:seed:dev --light` command
created the disposable workspace instead; no UI and no captain-owned instance
were used.

Firstmate's decision removed Keychain from this proof after the red 128-byte
truncation observation. The current image's seeded API-key emitter was piped
directly to `pm credentials add twenty-live --connector twenty --value-stdin
api_key`. The pipe boundary measured 444 bytes; no key entered argv, a file, a
prompt, output, commit, or this evidence. The immediate built-binary
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
