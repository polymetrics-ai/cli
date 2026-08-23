# TDD ledger — additive v3 multi-document source locks

## Task Delivery Header

- Issue: Refs #4318 — feat(connectorgen): add additive v3 multi-document source locks
- Base branch: main (`6410fe59c7ed9017dbe3f830f4361d4d015cd8e9` confirmed an ancestor of `origin/main`)
- Merges into: main
- Delivery: Direct PR open against `main` with local verification and API base confirmation.
- Working branch: fm/cli-multidoc-source-lock-foundation-r1
- Task: Deliver only the shared versioned v3 multi-document source-lock foundation and preserve all v2/GitHub bytes.
- Verification: Red/Green tests, full `make verify`, and frozen GitHub hashes.

## Red/Green ledger

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| V3 importer/provenance | `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportVersion3ImportsDocumentOwnedProvenanceAndDuplicateProviderIDs$'` failed: strict legacy decoding rejects aggregate `rest.openapi` as an array (`cannot unmarshal array ... into ... string`). Missing/drift/query cases share the same structural failure. | V3 fixtures import through the real parser/fetcher with two documents and duplicate provider `operationId` values, but retain distinct locked source IDs plus document/published provenance. Missing/digest-drift artifacts fail closed; published citations are never fetched; `TestSourceImportVersion3SynchronizesDuplicateArtifactDigests` proves one fetch for one digest. | green |
| V3 projection/surface-sync | `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceProjectionAcceptsVersion3DocumentProvenance$'` failed at the same v3 strict-decode boundary before schema-3 projection could run. | Source projection accepts the matching v3 lock/descriptor pair and rejects independent changes to provider ID, artifact URL, document ID, published URL/capture/hash/bytes/adapter, form, or version. Surface-sync accepts schema 3. | green |
| GitHub frozen parity | Pending: install immutable byte/hash assertions before Green completion. | `TestSourceImportPreservesFrozenGitHubArtifacts` verifies all three frozen byte count/SHA-256 pairs. | green |
| YAML scalar mapping-key normalization | `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceImportNormalizesScalarYAMLMappingKeys$'` failed: an unquoted OpenAPI response code reached the strict YAML guard as `YAML mapping key at /paths/~1items/get/responses must be a string`. | The real source-lock/import path accepts unquoted `200:` with the same operation contract as `'200':`; `200:` plus `'200':` still fails as one normalized duplicate, and a compound mapping key remains rejected. Standard JSON scalar tags (`!!int`, `!!float`, `!!bool`, `!!null`) use canonical JSON spelling; YAML-only/custom tags remain rejected. | green |

## Constraints

- No live Zoom retrieval: fixtures are hermetic because the provider's Next.js build-ID URLs are transient and may return 404.
- Published citation queries are accepted only under the bounded non-fetched policy in `PLAN.md`; retrieval URLs remain queryless.
- Do not migrate GitHub, change any GitHub definition artifact, or alter its JavaScript generator lane.

## Cycle 1 — v3 document-owned source import

- **Red:** schema-v3 locks use `rest.retrieval`, aggregate `rest.openapi: []`, and `rest.source_documents`; the legacy `sourceImportREST` decoder accepts only one embedded artifact and rejects the array before an importer can fetch or attribute a document.
- **Green target:** exact schema dispatch admits only v1/v2 legacy or v3 document models; the real importer fetches only each document artifact, verifies its bytes/digest, uses the locked global source ID despite duplicate provider operation IDs, and serializes document/published provenance.
- **Refactor constraint:** leave the v1/v2 path, checked-in GitHub lock, descriptor, and combined ledger byte-for-byte untouched.
- **Green:** focused v3 importer/projection/cache/hash tests passed, then `go test -count=1 -timeout 20m ./cmd/connectorgen` passed in `441.543s`.

## Cycle 2 — YAML scalar mapping keys

- **Red:** a pinned Docker Hub-style OpenAPI YAML document with an unquoted `200:` response key was rejected before import because the YAML-to-JSON bridge accepted only `!!str` mapping keys.
- **Green target:** JSON-representable scalar YAML keys normalize into the exact JSON object member string, and duplicate detection happens after that normalization.
- **Green:** `TestSourceImportNormalizesScalarYAMLMappingKeys` passes through the source-lock/fetch/import path; its duplicate and non-scalar assertions keep the strictness boundary closed.
