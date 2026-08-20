# Review Convergence — Current Foundations Main Integration r1

## Canonical result

The immutable source reviewed is `9e5329f34e015e39160bb8e951452bbd071a698a`. Its underlying provisional code composite is `808896a28873c5f0479fa10e2f798da56f885b5e`, based on `114a67727f2ef60b132054091c73987be4118a9b`.

Canonical ledger: `.planning/phases/cli-current-foundations-main-integration-r1/REVIEW.md`.

**Disposition: merge blocked.** The three lens ledgers contain 37 source finding IDs. Four compound IDs contain two independently testable claims, and one issue is duplicated across mapping/runtime. The exact transformation is:

`41 atomic claims - 4 compound-ID regroupings - 1 cross-lens duplicate = 36 canonical findings`

The canonical set contains **27 BLOCKER** and **9 WARNING** findings. The full 41/41 lossless crosswalk, evidence, affected symbols, exact changes, behavioral tests, and six-surface impact are in the canonical ledger.

This remains one rollup checkpoint, not component-PR review. Do not push, open a pull request, merge, retarget, close, or mutate any component branch or pipeline worktree as part of this review.

## Exact ancestry to audit

| Foundation | Exact head | Provenance | Preserving merge |
| --- | --- | --- | --- |
| source import / #4306 | `19a32bd0bc08faf217be8f45b39841b5ff589a92` | published #4312, prior qualified intake | `223f7b126eb4039f5c940a3cf233b15d6a18eff6` |
| closed operation runtime / #4307 | `3c768cade6703426afd2272fbc01bfd60583e04f` | published #4311, prior qualified intake | `1cb9cdb31c2fc446fe1da0b176b7422f04e81111` |
| status/export / #4302 | `fe5b8e18788538c4fcce34969da7ff88a7fa66d6` | published #4308, Firstmate terminal qualification | `9e3cd99b7ebd2ebac2303ad8770e50fee85c92c6` |
| structured bodies / #4305 | `55ddb650aa5594ddd156b0939cb1df6027a31d56` | captain-authorized local pipeline `01M0DY0HM9HNZVNKJ2J9Z9SCG7` | `0eb98d3844da7b48d0ca27f51ba7deb46d8f5d1b` |
| declarative reverse ETL / #4303 | `e7f474375af969555efd82f684ad6d0b8a26cfc0` | captain-authorized local pipeline `01M0DYNQ9HSJBYS9YQ4MJR4JGR` | `808896a28873c5f0479fa10e2f798da56f885b5e` |

The local worktrees were read only. `data/cli-current-foundations-main-integration-r1/input-manifest.json` is the machine-readable provenance record. At synthesis intake, `HEAD` matched the frozen SHA, tracked source/index were clean, and only the three expected lens ledgers were untracked.

## Lens coverage and count reconciliation

| Lens | Scope | Source IDs | Atomic claims | Canonical IDs |
| --- | ---: | ---: | ---: | --- |
| Mapping | 25 files | 12 | 14 | `CF-B01`–`CF-B11`, `CF-W01` |
| Runtime | 30 files | 15 | 16 | `CF-B08`, `CF-B12`–`CF-B23`, `CF-W02`–`CF-W03` |
| Orchestration | 22 files | 10 | 11 | `CF-B24`–`CF-B27`, `CF-W04`–`CF-W09` |
| **Total** | **77** | **37** | **41** | **36 unique** |

`MAP-BL-01`, `MAP-BL-11`, `RT-B03`, and `ORCH-W09` are the four compound IDs. The atomic rows are 31 blocker claims plus 10 warning claims. Regrouping three compound blocker pairs and one compound warning pair yields 28 blocker + 9 warning source IDs. `MAP-BL-08` and `RT-B01` are the sole true cross-lens duplicate; de-duplicating them yields 27 blocker + 9 warning canonical findings.

Source ledgers:

- `.planning/phases/cli-current-foundations-main-integration-r1/reviews/MAPPING-REVIEW.md`
- `.planning/phases/cli-current-foundations-main-integration-r1/reviews/RUNTIME-REVIEW.md`
- `.planning/phases/cli-current-foundations-main-integration-r1/reviews/ORCHESTRATION-REVIEW.md`

## Canonical finding index

| ID | Classification | Converged issue |
| --- | --- | --- |
| `CF-B01` | BLOCKER | Real source lock cannot import and GraphQL is silently ignored. |
| `CF-B02` | BLOCKER | Imported descriptors are orphaned from bundle generation/validation. |
| `CF-B03` | BLOCKER | Reverse actions omit provider inputs, including required fields. |
| `CF-B04` | BLOCKER | Generated GraphQL results/pagination are narrowed. |
| `CF-B05` | BLOCKER | Installed GitHub binary upload is JSON to the wrong origin with no file/name. |
| `CF-B06` | BLOCKER | Reusable destination authorization can fail after write/checkpoint. |
| `CF-B07` | BLOCKER | Generic destination read-back never reads provider state. |
| `CF-B08` | BLOCKER | Public provider receipts disclose configured credentials. |
| `CF-B09` | BLOCKER | CLI discards persisted failed runs and receipts. |
| `CF-B10` | BLOCKER | No-response direct writes lose attempted operation identity. |
| `CF-B11` | BLOCKER | Direct-read/download outputs omit complete and error receipts. |
| `CF-B12` | BLOCKER | Accepted-status mismatch discards a direct-write receipt. |
| `CF-B13` | BLOCKER | Idempotent reverse writes can redirect outside approval and lose final failures. |
| `CF-B14` | BLOCKER | Buffered redirects can carry custom credentials cross-origin. |
| `CF-B15` | BLOCKER | Reverse-write previews disclose credentials/sensitive identifiers. |
| `CF-B16` | BLOCKER | REST/binary error formatting prints raw query/provider body data. |
| `CF-B17` | BLOCKER | Declarative writes rematerialize after approval. |
| `CF-B18` | BLOCKER | Empty JSON receipts fabricate body presence. |
| `CF-B19` | BLOCKER | Numeric enum validation collapses integers above 2^53. |
| `CF-B20` | BLOCKER | Singleton flags silently use the last occurrence. |
| `CF-B21` | BLOCKER | Caller path/query values have no encoded byte bound. |
| `CF-B22` | BLOCKER | REST direct reads admit undeclared query/body inputs. |
| `CF-B23` | BLOCKER | Name-based redaction deletes ordinary provider IDs/counters. |
| `CF-B24` | BLOCKER | Provider effects precede stale-writer ownership CAS. |
| `CF-B25` | BLOCKER | Readback failure invokes pre-publication abort after publish. |
| `CF-B26` | BLOCKER | Full-overwrite/Arrow provider output never reaches results. |
| `CF-B27` | BLOCKER | Declared delivery guarantees are never enforced. |
| `CF-W01` | WARNING | Generated connector skills use a five-name allowlist. |
| `CF-W02` | WARNING | Path interpolation is not revalidated after filters. |
| `CF-W03` | WARNING | Accepted GitHub secret transform has no executor. |
| `CF-W04` | WARNING | Transient rate-parking failures lose the retry timer. |
| `CF-W05` | WARNING | Same-scope parked runs can resume concurrently. |
| `CF-W06` | WARNING | Route selection collapses actionable preflight errors. |
| `CF-W07` | WARNING | Generated website data is deterministic but semantically stale. |
| `CF-W08` | WARNING | Planning/evidence state disagrees and overstates coverage/SHA. |
| `CF-W09` | WARNING | Rate-parking stores can persist unloadable state. |

## Six-surface disposition

| Surface | Result | Primary convergence work |
| --- | --- | --- |
| ETL | **BLOCKED** | Provider read-back, reusable authorization, pre-I/O fencing, delivery compatibility, safe complete receipts. |
| Reverse ETL | **BLOCKED** | Field-complete actions, immutable approved execution, safe retry/redirect, terminal output, ownership. |
| Direct read | **BLOCKED** | Complete GraphQL/source projection, closed REST bindings, complete masked receipts, safe errors. |
| Direct write | **BLOCKED** | Masked terminal receipts, no-response/status preservation, exact numeric/duplicate/bound validation. |
| Binary download | **BLOCKED on failure contract** | Success confinement/integrity passes; complete masked error receipt and safe diagnostics remain. |
| Binary upload | **BLOCKED / incorrectly implemented** | Replace advertised JSON action with bounded confined exact binary operation. |

## Provider output and secrets

Convergence requires a complete internal receipt and a credential-safe public projection. The engine retains exact provider truth privately. App/CLI output deep-clones it, preserves every ordinary field/value/ID/header/body representation, and masks only exact configured/selected credentials, proven encodings, and declaration-classified secret locations. Fields remain present with masking and presence/byte metadata. Names such as `token` are not secrets by themselves. Raw provider bodies and request queries are never rebuilt into printable error text.

## Source lock and generated surfaces

Required authority chain:

`strict REST+GraphQL source lock -> canonical descriptor -> operations/actions/API surface -> CLI/help/manual -> website/skills -> runtime preflight/certification`

Every provider operation/field must project exactly once or carry a typed source-bound gap. Source import, validation, drift checks, installed-command reachability, documentation parity, and skill generation must fail on omissions or stale derived state. Endpoint identity alone is not field-complete coverage.

## One ordered fix wave

1. Freeze red tests for all canonical findings.
2. Repair source authority/generation; regenerate mapping, CLI, docs, website, and skills.
3. Establish one complete internal receipt, credential-safe public projection, and terminal CLI error contract.
4. Close request construction: redirects/retries, previews, materialization, exact numbers, duplicate flags, encoded bounds, direct-read bindings, transforms.
5. Repair authorization/read-back/fencing/publication/delivery/rate-parking state machines.
6. Certify all six installed surfaces at one exact code/reviewed SHA and reconcile PLAN/TDD/VERIFICATION/evidence atomically.

## Required convergence findings

- Typed direct write accepts only declaration-owned path, query, structured body, and header bindings. It must reject malformed, unknown, oversized, duplicate, CR/LF, and cross-bound values before I/O; no raw HTTP method/path/header/body/action escape hatch is present.
- Terminal direct-write results preserve provider status, repeatable headers, exact text/binary body bytes, response presence, operation ID, and path, including terminal errors. Generated diagnostics must not copy response bodies or secrets.
- A fixed GraphQL operation parses JSON when its content type is JSON or omitted; a provider-declared text/binary response remains byte-exact rather than fabricated as a GraphQL envelope.
- `rest_status` keeps the final non-2xx HEAD response and typed declaration-owned headers after retries. Binary/text GET failures remain ordinary errors and do not create an output file.
- Persisted App and installed CLI reach multiple independently selectable reverse-ETL actions with plan, approval, apply, durable acknowledgement, provider result persistence, and provider readback; no connector-name branch narrows an action.
- Source-locked provider declarations remain lossless/bounded and reach the generated command/help surface; generated documentation is synchronized.
- Existing specialized GitHub, scalar, form, SCIM, multipart, and binary behavior remain covered rather than selected away in a conflict.

## Narrow contracts that passed

- Frozen SHA, source cleanliness, and recorded component ancestry at review intake.
- REST endpoint method/path identity coverage, but not source field completeness.
- Operation-backed direct-write closed mappings, approval binding, and direct-operation redirect/retry refusal.
- Recursively closed/bounded structured REST bodies and root-confined/digest-bound multipart files.
- Successful binary/text download confinement, atomicity, exact size, and SHA-256. Failure output remains blocked.
- Fixed GraphQL caller-authority closure, but not generated selection completeness.
- Atomic definition registration, local warehouse durability, per-unit approval ordering, durable local acknowledgement checks, cancellation behavior, and bounded Arrow credits.
- Deterministic website generation, but not semantic parity.
- Focused package tests/vet recorded by the three ledgers.

Historical focused checks in `data/cli-current-foundations-main-integration-r1/evidence-manifest.json` remain useful only for those narrow contracts. They are not release proof: the real checked-in GitHub `source-import` fails with `source grammar position byte limit exceeded`, `params-import --check` reports 211 drifted operations, and several existing tests explicitly encode unsafe credential/presence behavior.

## Out of scope for this provisional checkpoint

No full verifier, real-provider operation, credential, temporary declaration/download, PR, CI, or no-mistakes stage is claimed complete. The review did not modify production/test source, push, open or mutate a PR, merge, retarget, or close anything. Any later component commits are additive follow-ups and must not replace the exact reviewed heads above.
