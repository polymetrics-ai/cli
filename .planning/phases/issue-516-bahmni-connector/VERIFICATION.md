# Verification — Bahmni-docker connector (issue #516)

## Gates run locally

| Gate | Result |
| --- | --- |
| `go run ./cmd/connectorgen validate internal/connectors/defs` | 548 connector(s) checked, **0 findings** |
| `go test ./internal/connectors/conformance -run 'TestConformance/bahmni'` | **PASS** (10/10 static, real fixture replay) |
| `./pm docs validate --connectors-dir docs/connectors` | **PASS** |
| `gofmt -l cmd internal` | clean |
| `go vet ./...` | clean |
| `go build ./cmd/pm` | ok |
| `go test -timeout 20m ./...` | green |

`internal/connectors/certify` is a pre-existing slow package that needs the repo's declared
`-timeout 20m` (the default 10m trips it). It never reads this bundle — no certify test enumerates
the defs fleet, and a full run logs zero `bahmni` references — so this connector does not affect it.

## CLI / help / docs / website parity

- `pm bahmni` (bare namespace) renders the connector manual and exits 0.
- `pm connectors inspect bahmni` and `--json` render identity, capabilities, config,
  12 ETL streams with primary keys, and 9 reverse-ETL actions with endpoint, required fields, and
  risk text. `password` renders as `(secret)`; no secret value is read or printed.
- `pm connectors catalog --json`, `--capability write`, and `connectors list --all` all include it.
- `docs/connectors/bahmni/{MANUAL.md,SKILL.md}` generated; `docs/connectors/README.md` and
  `docs/connectors/catalog/all-connectors.{json,md}` carry the new entry.
- `docs/cli/connectors.md` CATALOG text updated to 552/548.
- Website catalogs regenerated (`connectors.generated.json`, `connectors.generated.ts`,
  `connectors.catalog.generated.ts`, `connectors.catalog.data.generated.json`).

## Safety checklist

- No secret value is requested, printed, stored, or summarized; `password` is `x-secret`.
- No generic raw HTTP write, raw JSON body, arbitrary OpenMRS method/path/body escape hatch, shell
  write, or SQL write is exposed.
- Every non-excluded mutation is a named reverse-ETL action; reverse ETL stays
  plan → preview → approval → execute, with `confirm: destructive` on the clinical/destructive ones.
- Blocked endpoints carry recorded evidence: binary download (`binary_read`), patient purge
  (`destructive_action`), OpenMRS system administration (`admin_reverse_etl`).
- No new dependencies.

## Automated-review follow-up findings

An automated review pass raised six findings against this branch. The captain authorized the four
real defects for direct follow-up, and `.planning/phases/issue-516-bahmni-rename-parity-followup/`
records the fixes and validation evidence:

1. `document-path-not-redacted` — fixed by using `document_file_path`, which matches the existing
   `file_path` reverse-plan redaction marker.
2. `offset-paginator-on-array-endpoints` — fixed by setting per-stream `pagination: {type: none}`
   for the four root-array streams (`drug_orders`, `lab_results`, `appointments`, `diagnoses`).
3. `session-check-always-passes` — fixed by replacing `/ws/rest/v1/session` with authenticated
   `GET /ws/rest/v1/provider?v=default&limit=1` for the connector health check fixture/config.
4. `drug-orders-group-name-mismatch` — fixed by making group/command metadata consistently use
   `drug_orders`.

Product decisions still deferred to the maintainer:

5. `phi-redaction-unbacked` (error) — the bundle's text promises PHI redaction that no runtime code
   performs. The engine redacts only secret-shaped keys (`engine.shouldRedactJSONField`) and
   `sensitive_policy.redact_fields` has no consumer outside `validateSensitivePolicy`, so OpenMRS
   PHI carriers (`display`, `identifier`, `person`, addresses, `codedAnswer`, `value`) are not
   redacted. Either soften the claims to what the engine does, or add a real engine-level PHI
   policy (which would affect all 548 connectors and belongs in its own issue).
6. `diagnoses-nullable-primary-key` (info) — `x-primary-key: ["existingObs"]` is nullable and not
   required; Bahmni returns `null` for diagnoses not yet backed by an obs. Metadata-only impact
   today.

## Human gates

- Merge to `main`.
- Live Bahmni credentials and any live clinical write against a non-local deployment.
- Patient-document/attachment payload tests beyond fixtures.
