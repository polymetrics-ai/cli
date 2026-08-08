# Generator generalization validation

**Date:** 2026-08-08
**Scope:** evidence-only validation of PR #3957. The Watchmode, DocuSeal,
Float, and Copper bundles are staged under this directory; none was added to
`internal/connectors/defs`. The eligible 392 production sweep was not started.

The source contract applied here is
`CAPTAIN-ORDER-multisource-mapping.md`: an artifact is a source, not a
complete-surface claim; authoritative references may supplement it; every
normalized operation keeps its source URL/kind/version/hash/date and exact
coordinate; disagreements remain visible; and traversal is bounded,
connector-scoped, cached, and resumable.

## Final result

The three required generalization shapes all pass. Copper is an additional
non-OpenAPI Postman fallback proof; its staged static bundle passes, but its
existing production connector is a legacy Tier-3 native scaffold with no
embedded `cli_surface.json`, so no honest real-binary reachability claim can
be made for Copper in this generator-only PR.

| Connector | Shape | Mapped | Implemented | Named dependency | Flagged discrepancy | Reachable | Failed | Result |
|---|---|---:|---:|---:|---:|---:|---:|---|
| watchmode | 23-read OpenAPI 3.0.3 | 23 | 13 | 32 | 22 | 45/45 | 0 | pass |
| docuseal | OpenAPI 3.1.0 with 11 top-level webhooks | 34 | 9 | 25 | 0 | 34/34 | 0 | pass |
| float | Swagger 2.0 with external path-item refs | 102 | 5 | 99 | 2 | 104/104 | 0 | pass |
| copper | provider Postman collection | 77 | 5 | 77 | 5 | not applicable: legacy native scaffold | 0 | static fallback proof |

The `reachable` column counts every generated command path, not only
implemented commands; each also passed the bare namespace check. The Copper
real-binary limitation is an existing runtime-foundation fact, not a silent
drop or a generator refusal.

## Artifact evidence

| Connector | Artifact URL | Verified kind/version | Bytes | SHA-256 |
|---|---|---|---:|---|
| watchmode | `https://api.watchmode.com/openapi.json` | OpenAPI 3.0.3 | 101,353 | `9e306e252b816d5ec68aa65473eab846e845ffc40e3cdeb4d9da9cadb05a7f48` |
| docuseal | `https://console.docuseal.com/openapi.yml` | OpenAPI 3.1.0 | 192,929 | `7ac10d1c39b335bce962b6de277d88aded8ce476518b83835c76ad80157e0e4b` |
| float | `https://developer.float.com/swagger-api-v3.yaml` | Swagger 2.0 | 8,634 | `d204eae066136386aea4ea955fb9d0d08ef9ca85eafabc2bb2d2cd30b8751211c` |
| copper | `https://developer.copper.com/download/copper_postman_collection.json` | Postman collection v2.1.0 | 1,334,523 | `6ec0df3a8fc0dff4f4dfa56dea2d0aa1e319e6f4cfdfe29d8ab670653a89a8be` |

The raw artifacts, manifests, generated bundles, reports, operation maps, and
reachability TSVs are all in this directory. The parser reports the complete
Postman inventory (77 unique method/path requests); it does not silently apply
the survey's semantic exclusions when mapping the artifact.

## Operation mapping

Each map is derived from the generated `api_surface.json` and lists every
artifact operation with its bucket, disposition, and normalized provenance:

- [`watchmode-operation-mapping-rerun-2.json`](reports/watchmode-operation-mapping-rerun-2.json): 23 = 0 ETL / 0 reverse ETL / 23 direct read / 0 direct write / 0 binary / 0 unclassified; 22 discrepancies.
- [`docuseal-operation-mapping-rerun-2.json`](reports/docuseal-operation-mapping-rerun-2.json): 34 = 4 ETL / 6 reverse ETL / 3 direct read / 10 direct write / 0 binary / 11 unclassified webhook operations; 0 discrepancies.
- [`float-operation-mapping-rerun-2.json`](reports/float-operation-mapping-rerun-2.json): 102 = 5 ETL / 0 reverse ETL / 42 direct read / 55 direct write / 0 binary / 0 unclassified; 2 discrepancies.
- [`copper-operation-mapping-rerun-2.json`](reports/copper-operation-mapping-rerun-2.json): 77 = 0 ETL / 0 reverse ETL / 29 direct read / 48 direct write / 0 binary / 0 unclassified; 5 discrepancies.

The 22 Watchmode, 2 Float, and 5 Copper discrepancy rows retain the exact
machine-readable marker `present-in-surface-absent-from-artifact`. No source
surface row was deleted because a fetched artifact was narrower.

DocuSeal's 11 OpenAPI 3.1 top-level webhooks are visible as `WEBHOOK`
operations with `webhooks[...].post` coordinates and the named dependency
`engine.webhook_receiver_executor`; they are not marked implemented and are
not refused. Float's external Swagger path items resolve through the bounded
reference resolver; each resolved operation cites the exact referenced URL
and JSON/YAML coordinate. When the same method/path appears in an
authoritative linked source, `provenance.alternatives` preserves that second
citation instead of overwriting it.

## Combined static gate

The four staged bundles were gated together after materialization, not one
repository-wide gate per connector:

| Gate | Result | Evidence |
|---|---|---|
| `connectorgen validate` | 4 connectors, 0 findings, 0 warnings | [`multi-source-validate-rerun-3.json`](reports/multi-source-validate-rerun-3.json) |
| `surface-sync` derive | 0 fields changed after materialization | [`multi-source-surface-sync-rerun-3.stdout`](reports/multi-source-surface-sync-rerun-3.stdout) |
| `surface-sync --check` | 4 scanned, no drift | [`multi-source-surface-sync-check-rerun-3.stdout`](reports/multi-source-surface-sync-check-rerun-3.stdout) |
| `batch gate` | 4 included, 0 dropped; 32 implemented commands preflighted | [`multi-source-gate-rerun-3.json`](reports/multi-source-gate-rerun-3.json) |

The combined gate saw 265 declared operation rows. It did not exercise any
provider API or resolve credentials.

## Wall-clock timings

These are observed final rerun slices, including the real CLI startup cost.
The original network fetches remain recorded by the artifact evidence and
their byte/hash records; the final rerun used those connector-scoped caches.

| Step | Watchmode | DocuSeal | Float | Copper |
|---|---:|---:|---:|---:|
| Materialize, parse, map, write bundle | 6.07s | 1.75s | 0.94s | 0.99s |
| Mapped operation count | 23 | 34 | 102 | 77 |

Shared final slices: build staged real binary 12.27s;
`connectorgen validate` 1.02s; surface-sync derive 0.89s;
`surface-sync --check` 1.05s; combined `batch gate` 1.17s. Real-binary
reachability sweeps were Watchmode 105.88s (45 commands), DocuSeal 79.62s
(34 commands), and Float 251.73s (104 commands), with zero non-zero exits in
each TSV. The command sweeps pass `--help` only and use no credentials or
provider network. Copper's staged bundle passes static gates but is excluded
from the binary overlay because the production native scaffold does not expose
the generated command surface.

## Certification and handoff

All four results are static/documentation-only. **Implemented, not certified,
never exercised against the provider.** Certification is withheld for every
connector, including the staged pilots. PR #3957 remains unmerged. The
eligible 392 production generation and the seven-connector consolidation are
deferred until firstmate confirms the generator PR has landed.
