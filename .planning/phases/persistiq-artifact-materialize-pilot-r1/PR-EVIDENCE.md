# Generator capability PR evidence

**Date:** 2026-08-08
**Issue:** #3958
**Scope:** generator policy and staged pilot evidence only. No eligible-392
connector surface was fetched into production, generated under
`internal/connectors/defs`, or certified.

The multi-source contract applied in this update is recorded in
`CAPTAIN-ORDER-multisource-mapping.md` and requires authoritative fallbacks,
bounded official-source traversal, normalized per-operation provenance,
explicit disagreements, and visible disposition gaps.

## Capability delivered

`connectorgen batch materialize` now:

- accepts provider-owned OpenAPI/Swagger, OpenAPI-fragment, Postman, and
  official-reference inputs;
- parses every OpenAPI 3.x or Swagger 2.0 path operation, including OpenAPI
  3.1 top-level webhooks and safe external Swagger path-item references;
- uses the ledger's optional `provider_reference_url` as an authoritative
  fallback when the primary artifact cannot parse or is narrower than its
  measured ledger inventory;
- traverses official HTML/Markdown reference indexes to linked same-provider
  machine sources and explicit method/path pages, with 64-document and 64 MiB
  bounds, HTTPS/public-destination checks, credential-shaped query rejection,
  connector-scoped SHA-named caching, and resumable reads;
- normalizes method/path operations and keeps per-operation source URL, source
  kind, version, retrieval date, SHA-256, coordinate, and alternatives in
  `api_surface.json` provenance; operation catalog source URLs follow the
  operation's primary cited source;
- merges/deduplicates sources canonically while preserving disagreements in
  `provenance.alternatives`;
- retains surface operations absent from a narrower artifact with the exact
  `present-in-surface-absent-from-artifact` discrepancy marker;
- emits `availability=not_implemented` plus a machine-checkable
  `named_dependency=<slug>` for every command the runtime cannot execute;
  only runtime-preflightable commands are `implemented`; and
- keeps repository-wide static/runtime/reachability gates in the single final
  `batch gate` over staged results rather than paying the 551-connector sweep
  during each materialization.

The implementation does not use an AI or prose guess as parity evidence. HTML
extraction accepts explicit method/path evidence and provider-linked machine
sources only; ambiguous text remains an unknown inventory rather than an
invented endpoint.

## Red/green evidence

The prior red-first failures are preserved in
[`TDD-LEDGER.md`](TDD-LEDGER.md): complete inventory and named dependency
support failed before the policy implementation; the shape extension then
failed to compile before webhook and external-reference seams existed. Green
tests now include:

```text
go test -timeout 20m ./cmd/connectorgen -count=1
ok   polymetrics.ai/cmd/connectorgen
```

Coverage includes top-level webhooks, local and external path-item refs,
reference cycles/sibling safety, Markdown and HTML official-source traversal,
Postman nesting/normalization/deduplication, multi-source provenance
alternatives, text artifact-cache inputs, exact discrepancies, strict source
version checks, and runtime named-dependency validation.

## Existing 551-bundle regression

No production connector bundle changed. The unchanged embedded corpus remains
the regression baseline:

| Gate | Result | Evidence |
|---|---|---|
| `connectorgen validate internal/connectors/defs --json` | 551 connectors, 0 findings, 0 warnings | [`all-551-validate-rerun.json`](pr-evidence-2026-08-08/all-551-validate-rerun.json) |
| `connectorgen surface-sync internal/connectors/defs --check` | 551 scanned, no drift | [`all-551-surface-sync-rerun.txt`](pr-evidence-2026-08-08/all-551-surface-sync-rerun.txt) |
| `TestEveryImplementedCommandPassesRuntimePreflight` | pass | [`all-551-runtime-preflight-rerun.txt`](pr-evidence-2026-08-08/all-551-runtime-preflight-rerun.txt) |
| Focused generator/engine/runner tests | pass | [`focused-tests-rerun.txt`](pr-evidence-2026-08-08/focused-tests-rerun.txt) |
| `go vet` and `go build ./cmd/pm` | pass | command evidence in phase verification |

The new fields are optional and absent from the existing connector bundles;
the existing corpus has zero `not_implemented` commands before this
capability is consumed.

## PersistIQ pilot rerun

The persisted one-connector pilot remains green under the complete-inventory
policy:

| Measure | Result |
|---|---:|
| Artifact operations mapped | 21 |
| ETL / direct_read / reverse_etl / direct_write / binary_download / unclassified | 11 / 1 / 7 / 2 / 0 / 0 |
| Implemented | 21 |
| Named dependency | 3 |
| Flagged discrepancy | 3 |
| Real-binary paths reachable | 24/24 |
| Failed candidates | 0 |

The artifact was OpenAPI 3.0.1, 47,796 bytes, SHA-256
`0bf3e1ecbfbf6215360b5bb8f9d4fda816df4e1872470a00b529fb3e8b80946f`.
`GET /v1/mailboxes`, `/v1/activities`, and `/v1/accounts` remain visible with
the exact discrepancy reason. PersistIQ timings and operation-level evidence
are under [`rerun-2026-08-08/`](rerun-2026-08-08/).

## Multi-source generalization pilots

All generated outputs are staged evidence only. The final numbers are:

| Connector | Shape | Mapped | Implemented | Named dependency | Discrepancy | Reachable | Failed |
|---|---|---:|---:|---:|---:|---:|---:|
| watchmode | read-only OpenAPI 3.0.3 | 23 | 13 | 32 | 22 | 45/45 | 0 |
| docuseal | OpenAPI 3.1.0 + 11 webhooks | 34 | 9 | 25 | 0 | 34/34 | 0 |
| float | Swagger 2.0 + external refs | 102 | 5 | 99 | 2 | 104/104 | 0 |
| copper | Postman fallback | 77 | 5 | 77 | 5 | not applicable: legacy scaffold | 0 |

Per-bucket counts and every normalized operation/provenance record are in
[`generalization-validation-2026-08-08/`](generalization-validation-2026-08-08/):

- Watchmode: 0 ETL / 0 reverse ETL / 23 direct read / 0 direct write / 0
  binary / 0 unclassified.
- DocuSeal: 4 ETL / 6 reverse ETL / 3 direct read / 10 direct write / 0
  binary / 11 unclassified webhook operations.
- Float: 5 ETL / 0 reverse ETL / 42 direct read / 55 direct write / 0 binary /
  0 unclassified.
- Copper: 0 ETL / 0 reverse ETL / 29 direct read / 48 direct write / 0
  binary / 0 unclassified.

The combined staged gate included all four, dropped none, checked 32
implemented commands, and saw 265 declared operation rows. The three required
real-binary sweeps reached 183/183 command paths with zero failures. Copper's
static output is useful fallback evidence, but its current native connector
does not expose the generated command surface, so it is not presented as a
runtime pass.

Final rerun wall-clock slices: materialize Watchmode 6.07s, DocuSeal 1.75s,
Float 0.94s, Copper 0.99s; combined validate 1.02s, surface-sync derive
0.89s, surface-sync check 1.05s, batch gate 1.17s; staged binary build
12.27s; binary reachability Watchmode 105.88s, DocuSeal 79.62s, Float
251.73s. These are evidence timings, not provider latency or certification.

## Certification and delivery boundary

No credentials were read, requested, printed, or stored. No provider operation
was exercised. **Implemented, not certified, never exercised against the
provider.** This PR is a generator capability change only; PR #3957 remains
unmerged, the 392 production generation follows separately, and the
seven-connector consolidation is explicitly deferred until firstmate confirms
the generator has landed.

Required skills and manual GSD fallback evidence remain recorded in the phase
artifacts. PR body linkage is `Closes #3958`.
