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
- PARTIAL/BLOCKED: scoped no-mistakes validation.
  - `no-mistakes doctor`: pass.
  - `no-mistakes axi run --intent ...`: review/test/document/lint/push/pr completed with no findings and reported `outcome: checks-passed`, but the pipeline opened wrong-base PR #3592 to `main` instead of the required stacked sub-PR base.
  - Remediation: commented on and closed duplicate PR #3592; canonical sub-PR is #3589.
  - Current-code caveat: no-mistakes commit `1093a1f30` touched shared parent artifacts; forward commit `9a0b188c2` restored those out-of-scope changes. No-mistakes was not rerun after the scope-restoring commit to avoid another wrong-base duplicate PR.
- PASS: Final focused gate after scope restore: `go test ./internal/connectors/boundary ./cmd/connectorgen`.
