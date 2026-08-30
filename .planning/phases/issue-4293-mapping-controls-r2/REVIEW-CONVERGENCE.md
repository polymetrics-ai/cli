# Review convergence — Issue #4293 mapping controls R2 (Codex-only)

## Review custody and exception

- **Reviewer:** one independent Codex reviewer.
- **Codex-only exception:** the captain expressly excluded Claude for this review. The default Claude audit in `firstmate-exhaustive-review` is therefore not invoked. This ledger preserves the exception and records the complete Codex evidence set instead; it is not a claim that a Claude audit occurred.
- **Immutable code review SHA:** `82445ed8dc445bbf0b34b3a9423588f4e9a4b0fa` (`codex/4293-mapping-controls-r1`).
- **Baseline mapping-control SHA:** `27608b31ed0f3b138fe6218188ca02a084b4d8eb` (`fix/4293-source-operation-multilane-manifest-r1`).
- **Merge base:** `27608b31ed0f3b138fe6218188ca02a084b4d8eb`.
- **Review branch:** `codex/4293-mapping-controls-r1-codex-review-r1`, isolated from the implementation branch.
- **Custody:** source and runtime files remain read-only throughout discovery. Only this review ledger may be added after discovery is frozen.

## Task delivery header

- **Issue:** Refs #4293 — Batch R1 source-operation multi-lane manifest and validator.
- **Base branch:** `fix/4293-source-operation-multilane-manifest-r1` at `27608b31ed0f3b138fe6218188ca02a084b4d8eb`.
- **Merges into:** review artifact branch only; no merge is authorized.
- **Delivery:** a Codex-only canonical ledger that names the exact reviewed SHA, evidence, findings, and a `PASS` or `BLOCK` verdict; an issue #4293 comment may link the committed ledger.
- **Working branch:** `codex/4293-mapping-controls-r1-codex-review-r1`.
- **Task:** independently review the Batch R1 mapping-control repair without modifying mappings, source locks, connector definitions, certification, runtime behavior, or generated connector outputs.
- **Verification:** focused `cmd/connectorgen`/engine checks, cohort `--check`, schema and changed-path inspection, and `git diff --check`; no broad `go test ./...`.

## Frozen change surface

| Area | Files / symbols | Review focus |
| --- | --- | --- |
| Cohort anchor | `data/connector-canon/batch1-source-operation-mapping-cohort.json`, `cmd/connectorgen/sourceoperationmappingcohort.go` | Ten fixed locks, exact source membership/count/digest, canonical matrix references, no duplicated rows. |
| Mapping validation | `cmd/connectorgen/sourceoperationmapping.go`, `cmd/connectorgen/sourceoperationmapping_test.go`, mapping schemas | Mutation lanes, source-node-bound facts, ETL/binary/sync applicability, negative boundaries. |
| CLI/check-only wiring | `cmd/connectorgen/main.go` | Reachability and no provider/runtime execution. |
| Embedded schema loading | `internal/connectors/engine/bundle.go`, `internal/connectors/engine/metaschemas.go` | Schema registration only; no execution policy change. |
| Planning evidence | `.planning/phases/issue-4293-mapping-controls-r2/*` | Contract and claimed command results match the reviewed source. |

## Mandatory-lens status

| Lens | Status | Evidence / disposition |
| --- | --- | --- |
| Architecture and data flow | complete | `run` dispatches the check-only cohort command (`cmd/connectorgen/main.go:85-88`) to `sourceOperationMappingCohortPathCheck` (`sourceoperationmappingcohort.go:102-207`): schema → exact ten-lock membership → raw lock digest/count/source-ID digest → aggregate source-ID digest. It references connector-local matrix paths only (`:35-38`, `:248-254`). The source-mapping checker separately validates source-node-bound facts and lane cells. |
| Happy, bad, and edge behavior | complete | Focused tests cover malformed/digest/count/membership cases, missing mutation lanes, locked DELETE, GraphQL mutation roots, a source-cited non-mutating POST, missing record shape, wrong-node citations, and ETL/media/event contradictions (`sourceoperationmapping_test.go:189-456`). `MAP-R1-004` records the uncovered documented-MIME edge. |
| State machine / concurrency | not applicable | The reviewed path is synchronous, local file/schema validation; it introduces no goroutine, durable store, callback, lease, retry state, or connector execution state. |
| Security / secret taint | complete | The new paths accept local manifest paths, read source-lock bytes, and emit deterministic findings. The change surface contains no credential, authorization header, provider client, transport, or provider-I/O modification. |
| Retry / rate-limit / resume / idempotency | not applicable | This is authoring-time validation only; no request, retry, rate-limit, checkpoint, or write/replay path is introduced. |
| Output integrity | complete | Findings are sorted by connector/file/message (`sourceoperationmappingcohort.go:197-206`); the CLI reports checked connector/source counts and findings (`:84-91`). Focused cohort check reports exactly `10 connector(s), 4341 source operation(s), 0 finding(s)`. |
| Declaration reachability / closed surface | complete | The cohort command is reachable from root help and requires `<manifest> --check`; source locks must be canonical connector-owned regular files, and matrix inputs must be canonical connector-local paths. No generic connector path is accepted. |
| CLI / App parity | not applicable | The expected `connectorgen` check-only CLI wiring is present; no App, persisted App state, connector executor, or runtime surface changed. |
| Provider semantics | complete with finding | REST PUT/PATCH/DELETE and GraphQL `Mutation.` roots require direct-write/reverse-ETL dispositions; GET/HEAD require non-mutation; POST remains fact-driven. Exact source-node citations are enforced. `MAP-R1-004` identifies an unsound binary-media classifier. |
| Tests / evidence | complete with finding | Focused mapping, cohort, engine, schema, vet, and diff checks pass. Existing fixtures cover only `application/octet-stream`; no regression covers Vercel's source-documented `application/gzip`, which is required before a clean verdict. |

## Contract reconciliation

| Required control | Review result | Evidence |
| --- | --- | --- |
| Exact ten-lock / 4,341-operation denominator with digest and membership enforcement | pass | The tracked cohort declares the ten fixed connectors, `source_operation_count: 4341`, aggregate source-ID SHA-256 `790858ed64c5dfbf290048496417e8b6320348d3fdd4bde2315f3765e38ea67f`, and ten lock-specific byte/count/ID digests. The checker enforces exact membership, per-lock byte and sorted-ID digests, aggregate digest, and deterministic findings (`sourceoperationmappingcohort.go:126-206`). |
| REST mutations and GraphQL mutation roots require both `direct_write` and `reverse_etl`; non-mutating POST remains negative boundary | pass | `sourceOperationMappingMutationFindings` recognizes GraphQL `Mutation.` and REST PUT/PATCH/DELETE identities (`sourceoperationmapping.go:535-570`). It deliberately leaves POST dependent on a cited mutation fact. Focused tests cover missing write lanes, DELETE, GraphQL, mutation mismatch, and non-mutating POST (`sourceoperationmapping_test.go:302-387`). |
| Record shape plus ETL/binary/sync applicability cite source nodes and reject contradictions | block | Required fact citations are bound to the exact locked source URL and location (`sourceoperationmapping.go:443-470`) and ETL/sync contradiction checks are present. Binary applicability, however, is incorrectly limited by an unsourced MIME allow-list; see `MAP-R1-004`. |
| Connector-local matrix paths are referenced, not duplicated | pass | The cohort manifest contains only a canonical `matrix_path` per lock and no lane cells or operation rows. The checker validates path shape only (`sourceoperationmappingcohort.go:35-38,248-254`), preserving connector-local ownership. The path is intentionally a reference label at this mapping-control stage; it is not an in-scope physical matrix admission check. |
| No source membership hidden; no connector/runtime behavior changed | pass | The cohort re-reads every pinned lock before evaluating any mapping input, and reports all source IDs through count/digest reconciliation. The target diff changes planning, schemas, `connectorgen`, and engine schema registration only; it changes no source lock, connector definition, executor, transport, credential, certification, or provider-I/O path. |

## Finding ledger

| ID | Severity | Exists at review SHA | Reachability / violated invariant | Evidence | Proposed regression and disposition |
| --- | --- | --- | --- | --- | --- |
| MAP-R1-004 | High — BLOCK | yes, at `82445ed8dc445bbf0b34b3a9423588f4e9a4b0fa` | A source-cited binary-upload cell must be admissible when the exact locked provider operation documents binary request media. The validator must not reject documented binary media solely because it is absent from a small tool-owned MIME allow-list. This path is reached by a connector-local source-lane matrix entry for `vercel.rest.writeSessionFiles` with `media.request: ["application/gzip"]`, `binary_upload: applicable`, and its exact source-node citation. | `sourceOperationMappingApplicabilityFindings` delegates binary upload/download evidence to `sourceOperationMappingBinaryMedia` (`cmd/connectorgen/sourceoperationmapping.go:523-525`). That helper recognizes only PDF, octet-stream, ZIP, image, audio, and video (`:581-589`), so it returns false for `application/gzip`. The pinned Vercel source lock identifies `vercel.rest.writeSessionFiles` at `paths["/v2/sandboxes/sessions/{sessionId}/fs/write"].post` and explicitly requires a gzipped tarball with `Content-Type: application/gzip` (`internal/connectors/defs/vercel/sources/vercel-operation-source-lock.json:170705-170714`). The mapping schema itself permits arbitrary non-empty media strings, confirming the contradiction is in the validator rather than schema admission. | Add a focused regression using that exact Vercel source operation/citation and `application/gzip`; it must accept an applicable `binary_upload` disposition and cell. Replace or extend the classifier so documented source-backed binary request/response media is not falsely rejected by this closed list, while retaining rejection of genuinely unsupported/un-cited claims. Do not alter source locks, connector definitions, runtime, or certification as part of this repair. |

No other blocker was found in the requested mapping-control surface. The absent physical connector-local matrices are not a finding: the issue's preceding control review explicitly establishes their paths as reference-only labels for the in-flight connector-local matrix tracks, and this target correctly avoids copying their rows into the cohort.

## Command ledger

| Command | Result | Purpose |
| --- | --- | --- |
| `go test -timeout 20m ./cmd/connectorgen -run '^(TestBatch1SourceOperationMappingCohortCheckAcceptsTrackedDenominator\|TestBatch1SourceOperationMappingCohortCheckRejectsDigestCountAndMembershipDefects\|TestSourceOperationMappingCheckRequiresSourceBackedMutationWriteCells\|TestSourceOperationMappingCheckRequiresWriteCellsForLockedDelete\|TestSourceOperationMappingCheckRetainsSourceCitedNonMutatingPOSTBoundary\|TestSourceOperationMappingCheckRequiresGraphQLMutationWriteCells\|TestSourceOperationMappingCheckRejectsGraphQLMutationFactMismatch\|TestSourceOperationMappingCheckRejectsUncitedAndContradictoryApplicabilityFacts)$' -count=1` | pass (exit 0) | Focused cohort, mutation, citation, and applicability controls only. |
| `go run ./cmd/connectorgen source-operation-mapping-cohort data/connector-canon/batch1-source-operation-mapping-cohort.json --check` | pass — `10 connector(s), 4341 source operation(s), 0 finding(s)` | Executes exact pinned denominator/digest validation. |
| `go test -timeout 20m ./internal/connectors/engine -count=1` | pass — `ok polymetrics.ai/internal/connectors/engine 15.990s` | Verifies embedded cohort/mapping schema registration. |
| `go vet ./cmd/connectorgen ./internal/connectors/engine` | pass (exit 0) | Focused static Go check for touched packages. |
| `jq empty data/connector-canon/batch1-source-operation-mapping-cohort.json internal/connectors/engine/schema/source_operation_mapping.schema.json internal/connectors/engine/schema/source_operation_mapping_cohort.schema.json` | pass (exit 0) | Validates changed JSON artifact/schema syntax. |
| `git diff --check 27608b31ed0f3b138fe6218188ca02a084b4d8eb 82445ed8dc445bbf0b34b3a9423588f4e9a4b0fa` | pass (exit 0) | Checks immutable implementation diff whitespace/integrity. |
| Static source inspection: `nl -ba cmd/connectorgen/sourceoperationmapping.go | sed -n '487,589p'`; `nl -ba internal/connectors/defs/vercel/sources/vercel-operation-source-lock.json | sed -n '170700,170725p'` | finding reproduced by direct code/source comparison | Establishes `MAP-R1-004` without modifying source or constructing a synthetic project artifact. |

No broad `go test ./...` was run.

## Final verdict

**BLOCK** for immutable implementation SHA `82445ed8dc445bbf0b34b3a9423588f4e9a4b0fa`.

The ten-lock/4,341 denominator, mutation controls, exact source-node citation controls, matrix-reference ownership, and no-runtime-change boundary are implemented and pass the focused checks. `MAP-R1-004` remains a high-severity mapping-control false negative: a provider-documented binary `application/gzip` upload cannot be marked applicable because the validator's closed MIME allow-list rejects it. The next implementation SHA requires a fresh review after the focused regression and repair.

This review adds only this ledger. It does not modify mapping code, source locks, connector definitions, runtime, certification, generated output, or merge state.
