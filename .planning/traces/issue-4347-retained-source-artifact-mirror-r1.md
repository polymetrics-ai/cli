## Task Delivery Header

- Issue: Refs #4347 — retain pinned source artifacts for hermetic verification.
- Base branch: main.
- Merges into: main.
- Delivery: After the decision gate is explicitly closed, an issue-backed pull request is open against `main` with retained-byte verification, GSD/TDD evidence, and local checks recorded.
- Working branch: fm/cli-retained-source-artifact-mirror-r1.
- Task: Retain every provider artifact's exact pinned bytes and verify only the retained copy. Cover machine-readable, rendered-reference, and bundle sources while preserving existing valid pins and recording Elasticsearch/Zoom regression evidence.
- Verification: Focused `cmd/connectorgen` tests for happy, bad-input, and missing/mismatched/bundle edges; artifact identity checks; generated-file/snapshot checks; package tests and repository gates applicable to changed paths.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Retained bytes are used without a network call | live | Planned test will make a network fetch fail while retained bytes produce the expected import/projection. |
| All three source shapes are retained | live | Planned tests will import a spec, validate a rendered citation artifact, and unpack/verify a bundle. |
| Existing valid pins do not repin or weaken | live | Planned migration fixture will retain the original byte count and SHA-256 exactly. |
| A retained copy is independently proven identical | live | Planned test will reject a copy whose byte count or SHA-256 differs from the lock. |
| Elasticsearch and Zoom regressions are explicit | live | Planned hermetic fixtures will prove Elasticsearch drift recovery and Zoom 404 recovery or preservation of an irrecoverable state. |

## Decision Gate

Closed by Firstmate inbox 001 on 2026-08-24. Store tracked raw provider bytes, record provenance/terms as data without a new enforcement gate, and cover all 28 nonzero artifacts. The required skills are `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`. GSD execution is inline because the task contract prohibits role spawning.

## Decision Assessment — 2026-08-24

### Evidence

- `rg --files -uuu | rg '(^|/)sources/[^/]*operation-source-lock\\.json$'` finds exactly one checked-in source lock: GitHub. It declares two provider artifacts: REST `12,920,264` bytes and GraphQL `1,546,421` bytes, for an exact retained-raw addition of **14,466,685 bytes (13.80 MiB)**.
- `rg --files -uuu internal/connectors/defs | rg '/sources/[0-9a-f]{64}\\.(artifact|json|yaml|zip|gz)$'` finds zero retained, digest-addressed source artifacts. The current `github/sources/` files are a 3.42 MB lock and a 43.35 MB generated descriptor, not either locked provider input.
- `sourceImportArtifactCacheFetcher.FetchArtifact` first reads a local cache, then deletes invalid entries and calls the HTTP fetcher when it is absent or corrupt (`cmd/connectorgen/sourceimport.go:8300`). Its current failure text acknowledges this is a refresh path (`:8455`). A cache therefore cannot provide offline recovery.
- `internal/connectors/defs/elasticsearch/sources/` and `internal/connectors/defs/zoom/sources/` do not exist on current `main`. Zoom's ledger cites 35 Markdown module artifacts, including `https://developers.zoom.us/docs/api/accounts.md`; it preserves no bytes, SHA-256, or byte count. The current branch consequently cannot prove either the historic 404 bytes or their identity.
- The brief names 15 machine-readable and 12 rendered/bundle gaps. They have no current source locks, so their exact source byte totals are unknowable rather than estimable. The existing importer admits at most `256 MiB` raw artifacts per lock (`defaultSourceImportTotalArtifactBytes`): 27 new plus the current GitHub lock has a design-admission ceiling of **7.0 GiB**, not a forecast. This exceeds the brief's 3.8 GiB free-space observation. A fresh `df -k .` now reports 32.2 GiB free, but this is transient and does not make the worst case safe.

### Storage alternatives

| Location | Recovery and hermeticity | Footprint | Recommendation |
| --- | --- | --- | --- |
| **Tracked content-addressed artifact under each connector `sources/` directory** | A clone contains the bytes and can independently verify raw count/SHA-256 without provider or cache. Git history records provenance. | Exact current minimum: 13.80 MiB; unknown for the 27 unpinned gaps. Git compresses objects but worktrees materialize raw bytes. | **Recommended only for source artifacts with an affirmative redistribution clearance.** Keep a per-artifact license/provenance record and fail closed when missing. |
| Untracked adjacent file or user cache | Not recoverable by clone, branch change, or CI; deletion and a provider outage strand a lock again. | Low repository footprint. | Reject: it recreates the failure the task exists to remove. |
| Versioned external object store/Git LFS-style pointer | Can be content-addressed and separately immutable, but a fresh checkout or offline build needs another network service and credentials/availability policy. | Low worktree footprint; provider-sized remote storage. | Only an optional distribution cache. It cannot be the sole verification source under the no-network build requirement. |
| Compressed retained payload | Decompression plus raw digest can prove the original bytes, but it no longer stores those bytes verbatim and archive formats seldom shrink further. | Lower for JSON/Markdown; negligible for zip/gzip. | A possible future optimization only if the acceptance wording is relaxed; do not use it as the default mirror. |

### Licensing / redistribution gate

- **GitHub REST description:** its upstream repository explicitly declares MIT. A tracked snapshot is feasible if its license notice is retained. GitHub Docs content is CC-BY-4.0, so a copied documentation/schema artifact requires attribution and the license notice. Sources: https://github.com/github/rest-api-description and https://github.com/github/docs.
- **Elastic/Elasticsearch:** the exact historic artifact is absent, so its governing source and license cannot be established from current `main`. Elastic License 2.0 allows copying/distribution only subject to its no-managed-service, no-license-key-circumvention, and notice-preservation conditions (https://www.elastic.co/licensing/elastic-license). Do not mirror an Elastic artifact until the exact source path and its header/repository license are captured; no blanket clearance is justified.
- **Zoom:** treat a tracked raw documentation mirror as **restricted pending written clearance**. Zoom defines its APIs to include related documentation; its API Terms give only a limited, revocable license and prohibit creating/distributing derivative works except when expressly permitted, while reserving all other licenses (https://www.zoom.com/en/trust/legal/zoom-api-license-and-tou/?page=1). The public document currently resolves in this research environment, but its present availability cannot reconstruct the historic bytes and does not grant redistribution rights.

### Recommendation and requested decision

Adopt tracked, digest-addressed raw artifacts as the sole hermetic verification source **only for individually license-cleared inputs**; retain lock metadata plus raw size/SHA-256 and a separate provenance/license record. For Zoom, use an explicit unavailable/legal-blocked source state until written redistribution permission or an independently licensed published artifact exists. Do not use an untracked directory or cache as evidence.

## Firstmate retention resolution and measurement — 2026-08-24

Firstmate chose tracked retention for all provider bytes and requested no licensing enforcement gate; per-artifact source URL, retrieval timestamp, and known license/terms remain required data. The actual read-only measurement completed before any artifact write: **454,667,903 current response bytes** versus **453,725,820 locked bytes**, with **33,781,928 KiB** free (`df -k .`). The maximum known raw material is therefore approximately 433.6 MiB and does not threaten this worktree.

The measurement also proves why a live fetch cannot constitute a *preserving* migration: Auth0 is `5,655,204` current bytes versus a `5,655,921`-byte pin; GitHub GraphQL is `1,551,372` current bytes versus a `1,546,421`-byte pin. Across 917 Batch 8–10 source documents, 351 current response bodies differ in byte length from their locks.

## Firstmate re-pin amendment — 2026-08-24

**Firstmate**, not the captain, authorized the narrow amendment for the otherwise impossible historic-payload criterion: after an artifact-by-artifact check establishes that a fresh response is a real provider document (not an error, redirect, login wall, or similar response), the lock may explicitly re-pin to those exact current bytes and retain them. Every such row must visibly report connector/artifact, old and new byte counts and SHA-256 values, and retrieval date. An error response is never pin material: Zoom accounts remains irrecoverable at its observed HTTP 404/8,329-byte response.

## Firstmate scope ruling — 2026-08-24

Firstmate inbox 004 ruled that this foundation branch must not import connector source locks or retained artifacts from unmerged lanes. The deliverable is the hermetic mechanism plus retained artifacts for locks on current `main` (GitHub); Batch 8–10, Elasticsearch, and Zoom each adopt the mechanism and retain their own files in their own connector PRs. This preserves one-connector-per-lane delivery while leaving the mechanism fully tested for machine-readable, rendered-reference, and bundle shapes.
