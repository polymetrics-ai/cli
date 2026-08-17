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
| Product defect | 0 |
| Provider refusal or missing fixture | 63 |
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

### Rebase reconciliation

After the first rebase to integration `a96216d09`, generation still selected
the same 97 commands. The current base correctly marks ten REST path flags as
required, so the regenerated candidates include connector-owned fixture values
for their now-required flags. I reran the affected Advanced Security and
Codespaces cohorts (53 reads) with the rebased binary: all ten previously
reported path-flag product defects now reach GitHub and are provider/missing-
fixture non-passes. The bounded result is therefore **34 / 0 / 63**, not the
pre-rebase 34 / 10 / 53.

The final base `eba2658c5` changes only package-scoped `pm` test-binary reuse
and its phase evidence. It does not alter certification definitions, candidate
inputs, or runtime command code. I nevertheless regenerated candidates and
sweep twice on that head; both checks are byte-stable and the membership is
still 97. The live counts therefore remain current without falsely recording a
second execution of unchanged operations. There are no unresolved product
defects in this cohort; the current sweep keeps zero product defects explicit
rather than hiding them as a pass.

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
eligible_pending_live 122
+ fixture_required   1369
+ not_applicable       50
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
- `make verify` — PASS on final integration base `eba2658c5`; includes full `go test ./...`, with `internal/cli` PASS (892.463s), build, docs, smoke, lint, generated checks, boundary, canon, and release-target parity.

This direct-PR task used the documented inline/manual GSD fallback; its plan,
TDD red/green ledger, verification, live results, and final code review record
are committed under `.planning/phases/github-certification-candidate-generation-perishable-r1/`.
