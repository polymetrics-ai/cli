# VERIFICATION — issue 3586 generated/unrelated connector remediation

## Required focused gates

```bash
go test ./internal/connectors/boundary ./cmd/connectorgen
rg -n "zendesk-support|google-ads|gong|hubspot|bitbucket|xero|bahmni" docs/connectors website internal/connectors/defs
```

## Additional local gates planned when feasible

```bash
gofmt -w cmd internal
go vet ./...
go build ./cmd/pm
make verify
```

## Connector/generated validation evidence

```bash
for p in internal/connectors/defs/zendesk-support internal/connectors/defs/google-ads internal/connectors/defs/gong; do
  go run ./cmd/connectorgen validate "$p"
done
```

## Website/docs tests

- No non-generated `website/**` edits planned. If this changes, run the repository website/docs test command(s) before PR handoff.

## Safety checks

- No secrets requested, printed, summarized, or stored.
- No credentialed connector checks.
- No new dependencies.
- No reverse ETL execution.
- No generic raw write tooling.
- No edits to parent issue/PR body, parent branch state, shared engine/runner/connectorgen, workflow/ruleset wiring, or PM/no-mistakes guidance.
- No push to `main`.

## Results

- PASS: Red evidence captured before remediation artifacts:
  - Ledger completeness validation failed with `ledger incomplete: 12 missing disposition(s)`.
  - `go run ./cmd/connectorgen ownership --help` failed with `connectorgen: unknown subcommand "ownership"` because #3581 target-aware guard is not integrated in this worker base.
- PASS: Ledger/proof artifacts created:
  - `REMEDIATION-LEDGER.md` covers 29 audited path references.
  - `ownership-guard-fixture.json` / `OWNERSHIP-GUARD-FIXTURE.md` cover representative reject/allow cases for #3532/#3535 generated/unrelated findings.
- PASS: Connector definition validation:
  - `go run ./cmd/connectorgen validate internal/connectors/defs/zendesk-support` -> `1 connector(s) checked, 0 findings`.
  - `go run ./cmd/connectorgen validate internal/connectors/defs/google-ads` -> `1 connector(s) checked, 0 findings`.
  - `go run ./cmd/connectorgen validate internal/connectors/defs/gong` -> `1 connector(s) checked, 0 findings`.
- PASS: `gofmt -w cmd internal` completed with no Go file diff.
- PASS: `go test ./internal/connectors/boundary ./cmd/connectorgen`.
- PASS: `rg -n "zendesk-support|google-ads|gong|hubspot|bitbucket|xero|bahmni" docs/connectors website internal/connectors/defs` completed; count check returned `31225` matches.
- PASS: `go vet ./...` passed on retry after one transient Go build-cache/std-package lookup failure.
- PASS: `go build ./cmd/pm`.
- PASS: `make verify`.
- Pending: no-mistakes scoped validation after commit/push.
