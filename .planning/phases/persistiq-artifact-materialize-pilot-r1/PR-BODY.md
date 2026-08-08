## Intent

Generator capability change for complete, honest provider-artifact mapping.
Closes #3958.

This PR does not generate the eligible 392 connector surfaces. It ships the
generator policy and staged validation evidence that the later sweep depends
on; PR #3957 must land before any production sweep begins.

## What changed

- Every operation in a valid OpenAPI 3.x or Swagger 2.0 artifact is retained
  in the normalized parity surface, including OpenAPI 3.1 top-level webhooks
  and externally referenced Swagger path items.
- The ledger's `provider_reference_url` is an authoritative fallback when a
  primary artifact fails to parse or is narrower than the measured inventory.
  Provider Postman collections and official HTML/Markdown reference indexes
  are supported source shapes.
- Official source traversal is bounded to 64 documents and 64 MiB, uses
  HTTPS/public-destination and credential-shaped-query guards, caches by
  connector and URL SHA-256, and preserves resumability.
- Every normalized operation carries source URL, source kind, version,
  retrieval date, SHA-256, exact coordinate, and preserved authoritative
  alternatives. `operations.json` uses the operation's primary source URL.
- Surface operations absent from a narrower artifact remain visible with the
  exact `present-in-surface-absent-from-artifact` marker.
- Commands are `implemented` only when runtime preflight can execute them;
  otherwise they remain visible as `not_implemented` with a named machine-
  checkable dependency. Webhooks use
  `engine.webhook_receiver_executor` until that runtime exists.
- Materialization stays cheap; the repository-wide gates run once over staged
  results in `batch gate`, not once per connector.

## Verification

- `go test -timeout 20m ./cmd/connectorgen -count=1` — pass.
- Existing 551-bundle `connectorgen validate` — 551 checked, 0 findings,
  0 warnings.
- Existing 551-bundle `surface-sync --check` — no drift.
- `TestEveryImplementedCommandPassesRuntimePreflight` — pass.
- Four staged source-shape bundles — validate 0 findings, surface-sync no
  drift, batch gate 4 included / 0 dropped / 32 implemented commands checked.
- Real binary command reachability — Watchmode 45/45, DocuSeal 34/34, and
  Float 104/104, each with bare namespace success and zero unknown commands.

## Pilot evidence

| Connector | Artifact | Mapped | ETL / reverse ETL / direct read / direct write / binary / unknown | Implemented | Named dependency | Discrepancy | Reachable |
|---|---|---:|---|---:|---:|---:|---:|
| PersistIQ | OpenAPI 3.0.1, 47,796 bytes | 21 | 11 / 7 / 1 / 2 / 0 / 0 | 21 | 3 | 3 | 24/24 |
| Watchmode | OpenAPI 3.0.3, 101,353 bytes | 23 | 0 / 0 / 23 / 0 / 0 / 0 | 13 | 32 | 22 | 45/45 |
| DocuSeal | OpenAPI 3.1.0, 192,929 bytes | 34 | 4 / 6 / 3 / 10 / 0 / 11 | 9 | 25 | 0 | 34/34 |
| Float | Swagger 2.0, 8,634 bytes | 102 | 5 / 0 / 42 / 55 / 0 / 0 | 5 | 99 | 2 | 104/104 |
| Copper | Postman, 1,334,523 bytes | 77 | 0 / 0 / 29 / 48 / 0 / 0 | 5 | 77 | 5 | staged static only; legacy native scaffold |

Artifact SHA-256 values, full operation maps, provenance, reports, and
reachability TSVs are in
`.planning/phases/persistiq-artifact-materialize-pilot-r1/generalization-validation-2026-08-08/`.

Final rerun wall-clock slices: PersistIQ's previously recorded pilot total is
52.92s. For the generalization rerun, materialization took Watchmode 6.07s,
DocuSeal 1.75s, Float 0.94s, and Copper 0.99s. The one combined staged gate
took validate 1.02s, surface-sync derive 0.89s, surface-sync check 1.05s,
and batch gate 1.17s. The real binary build took 12.27s; command reachability
took Watchmode 105.88s, DocuSeal 79.62s, and Float 251.73s. These are static
help-path checks only.

## Certification and safety

No credentials were read or stored. No provider operation was exercised.
**Implemented, not certified, never exercised against the provider.**
Certification is withheld for PersistIQ and all staged pilots. The generated
pilot bundles are evidence only; they are not production connector surfaces.

Required GSD/manual-fallback evidence is under
`.planning/phases/persistiq-artifact-materialize-pilot-r1/`. The eligible 392
and seven-connector consolidation are explicitly deferred until this PR is
merged and firstmate authorizes the next phase.
