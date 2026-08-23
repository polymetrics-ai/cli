# UAT — issue #4318 additive v3 multi-document source locks

Status: passed by automated execution; no human judgment is required for the scoped foundation
properties.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| Multi-document v3 import | `TestSourceImportVersion3ImportsDocumentOwnedProvenanceAndDuplicateProviderIDs` runs the real lock parser and fetch/import path with two hermetic artifacts. | Pass — repeated provider `operationId` values retain distinct locked source identities and per-document published provenance. |
| Fail-closed source bytes | `TestSourceImportVersion3RejectsMissingOrDriftedDocument`. | Pass — a missing artifact and digest mismatch prevent import. |
| Query citation safety | `TestSourceImportVersion3RejectsUnsafePublishedCitationQuery` plus the valid fixture's fetch counter. | Pass — safe bounded citation queries are preserved but never fetched; credential-like queries are rejected. |
| v2/GitHub compatibility | `TestSourceImportPreservesFrozenGitHubArtifacts`, `source-import github --check`, and full `make verify`. | Pass — all three mandated artifact byte/hash values remain exact. |
| YAML response-code correctness | `TestSourceImportNormalizesScalarYAMLMappingKeys`. | Pass — `200:` imports like `'200':`, their coexistence is rejected, and a compound key remains rejected. |

The Docker Hub-specific source directory is absent from this base, so no live Docker Hub source-import
check is possible here. The reported parser defect is fully exercised through a hermetic source-lock
fixture without expanding this lane into capture or connector artifact work.
