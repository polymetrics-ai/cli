Refs #4015

## Outcome

Adds a generic `connectorgen certification-candidates` projection for
implemented `direct_read` commands. It derives command tokens, required flags,
credential, JSON mode, and a produced `/response` object-or-array assertion
from `cli_surface.json`. Connector definition data supplies only auditable
cohort membership and fixture default values. No shared Go names a connector.

GitHub now declares the trial-perishable cohort literally in its own
`certification.json`: Advanced Security 50, Copilot 50, Enterprise 46, and
Codespaces 46 — **192** commands total. The generic projection generated the
**97 direct reads** among them (31 / 23 / 21 / 22); the remaining **95
reverse-ETL commands** (19 / 27 / 25 / 24) are explicitly deferred to the
separate mutation-fixture lifecycle. This PR does not claim all 192 and does
not close the 1,571-command certification gate.

Candidate provenance: 97 generated, 23 named manual overrides. Each retained
manual override asserts a fixture-specific semantic response value not present
in `cli_surface.json`; its exact command/reason list is in the committed
`LIVE-RESULTS.md`. A manual candidate can shadow only its exact command; there
is no unbounded escape hatch.

## Live direct-read proof

All 97 generated trial reads were executed serially by the disposable
certification identity against the approved fixture repository, with
environment-only credentials and produced-value assertions:

| Result | Count |
| --- | ---: |
| Produced-value pass (executed, not yet certified/published) | 34 |
| Product defect | 10 |
| Provider refusal or missing fixture | 53 |
| **Executed generated reads** | **97** |

The accepted-evidence importer is concurrently owned elsewhere, therefore the
34 passes are deliberately not recorded as `live_tested` or published here.
No GitHub mutation was performed; the local disposable credential project was
moved to Trash after the read run.

Code Security is treated correctly as live-runnable: `code-scanning
default-setup view` passed, while `code-scanning analyses view` returned the
provider message `no analysis found`. That is a missing-analysis fixture, not
an entitlement block. Provider policy and absent-resource responses are kept as
non-passes rather than reclassified as product failures.

### Product defects (10)

- `code-scanning analyses view-2` — `analysis_id` path flag optional
- `code-scanning autofix view` — `alert_number` path flag optional
- `code-scanning databases view-2` — `language` path flag optional
- `code-scanning instances view` — `alert_number` path flag optional
- `code-scanning repos view` — `codeql_variant_analysis_id` path flag optional
- `code-scanning sarifs view` — `sarif_id` path flag optional
- `code-scanning variant-analyses view` — `codeql_variant_analysis_id` path flag optional
- `dependabot secrets view-2` — `secret_name` path flag optional
- `secret-scanning locations view` — `alert_number` path flag optional
- `codespaces secrets view-2` — `secret_name` path flag optional

## Negative control

The known passing generated `copilot configuration view` candidate was
deliberately regenerated with `/response` typed as `string`. Its own stage went
red:

```
generated_direct_read_copilot_configuration_view: declared output at /response has the wrong type
```

Restoring `object_or_array`, regenerating both candidate and sweep artifacts,
and rerunning that same case made the stage pass again.

## Coverage arithmetic

The generated sweep remains exhaustive:

```
eligible_pending_live 109
+ fixture_required   1290
+ not_applicable       50
+ product_defect       92
+ provider_refused      1
+ schema_conformant    29
= total              1571
```

## Verification

- `go test -timeout 20m ./cmd/connectorgen` — PASS (117.973s)
- `go test -timeout 20m ./internal/connectors/certify` — PASS
- `go test -timeout 20m ./internal/connectors/engine` — PASS
- `make connectorgen-validate connectorgen-surface-sync connector-boundary` — PASS
- generated candidates and sweep each regenerated twice, then both `--check` commands passed (byte-stable)
- `make verify` — PASS; includes full `go test ./...`, with `internal/cli` PASS (816.971s), build, docs, smoke, lint, generated checks, boundary, canon, and release-target parity

This direct-PR task used the documented inline/manual GSD fallback; its plan,
TDD red/green ledger, verification, live results, and final code review record
are committed under `.planning/phases/github-certification-candidate-generation-perishable-r1/`.
