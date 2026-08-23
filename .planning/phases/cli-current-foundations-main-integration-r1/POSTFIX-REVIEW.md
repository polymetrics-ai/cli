---
phase: cli-current-foundations-main-integration-r1
reviewed: 2026-08-21T03:27:40Z
depth: deep
source_sha: 8a8a866ff6d5282c28bda12acceed8a624218f01
diff_base: e62ae21d428f0d27225f9bff564dc2cd797f6b65
source_ledgers:
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/POSTFIX-MAPPING-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/POSTFIX-RUNTIME-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/POSTFIX-ORCHESTRATION-REVIEW.md
prior_closure_ledgers:
  - .planning/phases/cli-current-foundations-main-integration-r1/REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/REVIEW-FIX.md
files_reviewed: 5
files_reviewed_list:
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/POSTFIX-MAPPING-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/POSTFIX-RUNTIME-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/reviews/POSTFIX-ORCHESTRATION-REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/REVIEW.md
  - .planning/phases/cli-current-foundations-main-integration-r1/REVIEW-FIX.md
raw_finding_ids: 47
raw_atomic_claims: 73
compound_raw_ids: 24
intra_id_regrouping_reductions: 26
cross_lens_duplicate_merges: 1
findings:
  critical: 38
  warning: 8
  info: 0
  total: 46
status: issues_found
verdict: blockers
---

# Foundation Post-Fix Canonical Code Review

**Frozen source:** `8a8a866ff6d5282c28bda12acceed8a624218f01`

**Depth:** deep

**Status:** issues_found — merge blocked

## Summary

The three post-fix ledgers are losslessly converged into **46 canonical findings: 38 BLOCKER and 8 WARNING**. Their 47 raw IDs expand into 73 independently testable atomic claims. Twenty-four compound raw IDs account for 26 intra-ID regrouping reductions, restoring 47 logical source findings. Exactly one cross-lens pair is materially duplicate: runtime `PFR-BL-10` and orchestration `ORCH-PF-B12` both prove the same `streamBinaryDownloadToRoot` publication defect and require the same no-replace/owned-cleanup publication primitive. The arithmetic is therefore **47 raw IDs -> 73 atoms -> 47 regrouped source findings -> 46 canonical findings**.

No other similarly themed claims were merged. GraphQL secret classification and public receipt sanitization, GraphQL `Int` bounds and CLI numeric fidelity, early destination authorization and cross-process auth fencing, ETL page budgets and direct-read cursors, and the four numeric-fidelity failures all traverse different code paths and need different fixes.

The prior `REVIEW-FIX.md` says all 36 original findings were fixed. This ledger marks each remaining issue as a **claimed-closure regression/residual**, a **newly exposed gap**, or a **mixed** finding. A prior narrow behavior can remain closed while a different path is blocked; the classification does not erase that distinction.

## Non-Negotiable End-to-End Contracts

### Ordinary provider truth and secret-only masking

The internal receipt must retain complete provider truth: operation and attempt identity, ordered repeatable headers, status, body presence/bytes/raw encoding, decoded value, GraphQL data/errors/extensions, provider resource IDs, and occurrence IDs. The public result is a deep, non-mutating projection that masks only exact configured secret scalars, their proven syntax encodings, and explicitly source-classified secret value locations. It must never rename or remove JSON keys or header names, apply token/secret keyword heuristics, round numbers, or alter an ordinary value such as `graphql-occurrence-9007199254740993`.

### Complete API reachability and source-driven writes

Provider source truth must drive REST and GraphQL operations, reverse-ETL actions, direct-read/write mappings, binary transports, CLI flags, help, website, skills, and certification. A source identity with a typed gap may remain visible but cannot be `availability=implemented` or count as covered. Generic reverse ETL must remain endpoint/action driven; no GitHub- or connector-name dispatch is an acceptable repair.

The frozen audit confirms the narrow generic reverse path is currently declaration-driven: `App.PlanReverseETL` resolves endpoint and `WriteValidator` at `internal/app/app.go:1871-1915`, preview uses `DryRunWriter` at `:2292-2352`, and execute selects `Writer`/`OperationDirectWriter` at `:2870-2933`. Remediation must preserve this closure while repairing incomplete source-derived actions and receipt/read-back semantics.

### Durable delivery

Approval, authorization, ownership, prepared bytes, provider effects, read-back, checkpoint, and terminal publication form one state machine. Durable CAS outcomes and exact receipts must determine recovery. A successful external effect must not be replayed or relabeled failed because cleanup, marker publication, or directory sync later fails.

## Critical Issues

### PF-CF-B01 — GitHub generated parity artifacts are stale and order-dependent

- **Severity / source / state:** BLOCKER; `PFM-B01`; claimed-closure residual of `CF-B02` and the prior evidence closure.
- **Proven manifestations:** parity generation is not order-commutative, leaving the combined operation ledger with 11 stale classifications/progress rows; both arise from the same noncanonical artifact-set generation boundary and are one atomic claim in the crosswalk.
- **Exact evidence:** `Makefile:108-115,158-163`; `scripts/gen-github-graphql-parity.mjs:665-676`; `scripts/tests/github-combined-operation-ledger.test.mjs:259-272`. At the frozen SHA, `node scripts/gen-github-graphql-parity.mjs --check` and `node scripts/github-combined-operation-ledger.mjs --check` exit 1; the nodes row says `fixed_typename_projection` instead of `fixed_projection_only`.
- **End-to-end path:** source lock -> generated REST/GraphQL operation and command cohorts -> bundle files -> combined ledger -> `github-parity-artifacts-check` -> release verification.
- **Concrete flow:** generate nine new REST commands before GraphQL versus after GraphQL; the same source produces different command ordering and a different ledger/progress total.
- **Six-surface impact:** ETL runtime not disproved; reverse ETL, direct read, direct write, and binary upload have stale ledger rows; binary download shares the red release gate.
- **Production risk:** required release evidence is red and cannot identify the executable artifact set deterministically.
- **Minimal fix:** impose one canonical total order independent of generation sequence, regenerate CLI and combined-ledger artifacts, and keep operation/API content byte-exact.
- **Red-first tests:** `TestGitHubParityGenerationOrderIsCommutative`; generate both orderings in separate temporary trees and require byte-identical operations/CLI/API/ledger output, correct nodes state, both `--check` commands, and `make github-parity-artifacts-check` green.

### PF-CF-B02 — Gap-tagged source operations masquerade as implemented API coverage

- **Severity / source / state:** BLOCKER; `PFM-B02`; claimed-closure residual of `CF-B02`, `CF-B03`, and `CF-B22`.
- **Exact evidence:** `cmd/connectorgen/sourceprojection.go:94-109,613-616,688-720` skips gap-tagged mutations in both projection and coverage. GitHub `repos/update` has provider body fields at `github-operation-descriptor.json:594164-594238` and a gap at `:598837-598845`, but `writes.json:2535-2547` and `cli_surface.json:3385-3399` expose an empty action. `/advisories` query inputs are at descriptor `:45076-45163`, gaps at `:46074-46087`, and an implemented zero-flag command at CLI `:11713-11726`. Cohort: 351 gap-tagged operations, including 36 implemented reverse writes and one implemented direct read with zero inputs.
- **End-to-end path:** provider source -> descriptor `runtime.gaps` -> projector skips identity -> empty authored action survives -> `surface-sync` derives no flags -> coverage validator skips the same identity -> commandrunner sends an incomplete request.
- **Concrete flow:** `pm github repos update --repo2 ...` cannot express any legal update field; list-global-advisories cannot express `affects`, `cwes`, `direction`, or `ecosystem`, yet both are advertised implemented.
- **Six-surface impact:** reverse ETL and direct read blocked; direct write coverage can be falsely affirmative; no ordinary ETL or current binary gap was proved.
- **Production risk:** commands reach provider I/O with invalid or semantically empty requests while certification says complete.
- **Minimal fix:** count every source identity in completeness; represent bounded variants or typed alternatives; keep unresolved identities unavailable with a source-bound gap rather than deleting them from accounting.
- **Red-first tests:** `TestSourceProjectionGapOperationsCannotMasqueradeAsImplemented`; pin `repos/update` and `/advisories`; add/change/delete one path, query, body, and header field and require validation/`--check` failure before network I/O until exact typed mappings exist.

### PF-CF-B03 — Google Ads POST reads close legal nested request objects as empty

- **Severity / source / state:** BLOCKER; `PFM-B03`; newly exposed gap.
- **Exact evidence:** `internal/connectors/defs/google-ads/operations.json:322-403` declares `customers.generate.keyword.ideas`; six legal branches are closed empty objects at `:355-386`. `engine.compileOperationDirectReadBodySchema` closes them at `internal/connectors/engine/direct_read.go:847-925`; `operationReadBody` validates at `:755-816`. Audit: 31 closed-empty nested nodes across 14 implemented POST reads.
- **End-to-end path:** Discovery schema -> generated `body_schema` -> scalar-only command flags -> commandrunner overrides -> closed engine schema -> provider POST.
- **Concrete flow:** a non-empty `keywordSeed` is the legal provider input for `generateKeywordIdeas`, but local validation rejects it and no flag can represent it.
- **Six-surface impact:** direct read blocked for 14 POST operations; ETL, reverse ETL, direct write, and binary surfaces unaffected.
- **Production risk:** implemented provider reads are unreachable for normal valid inputs.
- **Minimal fix:** resolve all nested Discovery references into finite closed child schemas, preserve requiredness/unions, generate JSON or dotted flags, and enforce the exact seed one-of.
- **Red-first tests:** `TestGoogleAdsGeneratedPOSTReadsAcceptDeclaredNestedObjects` for `keywordSeed` and an alternate seed; add a cohort walk that rejects reachable closed-empty objects unless explicitly annotated provider-empty.

### PF-CF-B04 — Paginated GraphQL commands accept no pagination direction

- **Severity / source / state:** BLOCKER; `PFM-B04`; claimed-closure residual of `CF-B04`.
- **Exact evidence:** `scripts/gen-github-graphql-parity.mjs:207-230,364-386`; generated search schema `operations.json:9558-9602`; flags `cli_surface.json:41032-41069`; runtime `internal/connectors/engine/graphql_operation.go:563-619`. Current tests cover flag presence and mixed directions but not zero direction.
- **End-to-end path:** locked GraphQL arguments -> optional `first`/`last` schema and flags -> commandrunner body -> `graphQLOperationVariables` -> fixed GitHub connection query.
- **Concrete flow:** invoke a generated connection query with required business arguments but neither `first` nor `last`; local preflight passes and GitHub rejects the resolver call.
- **Six-surface impact:** direct read GraphQL connections blocked; the other five surfaces unaffected.
- **Production risk:** provider-invalid requests are approved as locally valid.
- **Minimal fix:** enforce exactly one of `first` or `last` before I/O, or define one documented deterministic direction/default across schema, help, and runtime; cursors remain opaque.
- **Red-first tests:** `TestGraphQLOperationVariablesRequiresExactlyOnePaginationDirection`: neither and both fail pre-I/O, first-only and last-only pass, cursor-without-direction fails; sweep every generated paginated command for the same contract.

### PF-CF-B05 — Backward GraphQL traversal uses forward `pageInfo`

- **Severity / source / state:** BLOCKER; `PFM-B05`; claimed-closure residual of `CF-B04`.
- **Exact evidence:** generation declares both directions at `scripts/gen-github-graphql-parity.mjs:409-417`; runtime selects `before` for `last` at `internal/connectors/engine/graphql_operation.go:573-596`, drops direction at `:688-695`, and always reads `hasNextPage/endCursor` at `:938-988`.
- **End-to-end path:** `--last --page-cursor` -> variables use `before` -> provider returns `hasPreviousPage/startCursor` -> directionless page parser reads forward fields -> CLI pagination state.
- **Concrete flow:** response has `hasPreviousPage=true`, `startCursor=previous-cursor`, `hasNextPage=false`, and an unrelated `endCursor`; result incorrectly reports complete with no usable continuation.
- **Six-surface impact:** direct read backward GraphQL pagination truncates/misnavigates; other surfaces unaffected.
- **Production risk:** silent partial data with a success result.
- **Minimal fix:** carry the validated direction into page parsing and select previous/start for backward, next/end for forward.
- **Red-first tests:** `TestOperationDirectReadBackwardGraphQLPaginationUsesPreviousPageInfo`, including terminal and malformed cases and an unrelated forward cursor.

### PF-CF-B06 — GraphQL secret input and result locations are not source-classified

- **Severity / source / state:** BLOCKER; `PFM-B06`; claimed-closure residual of `CF-B08`/`CF-B23`.
- **Atomic claims:** query arguments such as `invitationToken` are allowed inline; provider-issued result secrets such as `tempCloneToken` and `verificationToken` are selected and published without a response policy.
- **Exact evidence:** generator classifies only mutations at `scripts/gen-github-graphql-parity.mjs:306-335,436-463`; source query argument at lock `:11111-11124`; generated query/CLI at `operations.json:8994-9018` and `cli_surface.json:40544-40564`; selected `tempCloneToken` at operations `:9458-9468`; `verificationToken` mutation at `:19143-19181` and CLI without response policy at `:45024-45046`; engine response masking depends on `SensitivePolicy.ResponseSecretField` at `internal/connectors/engine/direct_write.go:249-254`.
- **End-to-end path:** GraphQL schema argument/result -> generator omits input/output sensitivity -> CLI accepts argv or fixed document selects token -> engine sees no declared secret location -> App persists/prints result.
- **Concrete flow:** invitation token enters shell history; `regenerateVerifiableDomainToken` returns a new verification token that `json_redacted` publishes.
- **Six-surface impact:** direct read input/output, reverse ETL, and direct write blocked; ETL and binary surfaces unaffected.
- **Production risk:** credential disclosure at command history and public-result boundaries.
- **Minimal fix:** derive reviewed source-owned sensitive input/output paths for both Query and Mutation; require env-only secret inputs and mask/withhold only exact classified result values while preserving ordinary token-like values and occurrence IDs.
- **Red-first tests:** enterprise invitation rejects inline and accepts env input; repository `tempCloneToken` and domain `verificationToken` are masked; issue IDs, `clientMutationId`, token-count fields, and `graphql-occurrence-9007199254740993` remain exact.

### PF-CF-B07 — Generated GraphQL mutations omit created/updated provider resources

- **Severity / source / state:** BLOCKER; `PFM-B07`; newly exposed GraphQL output gap; prior ordinary receipt/occurrence-ID preservation remains narrowly closed.
- **Exact evidence:** `outputSelection` selects scalar leaves only at `scripts/gen-github-graphql-parity.mjs:258-273`; source `createIssue` payload at lock `:13498-13518,44752-44770`; generated document `operations.json:13432-13444`; engine returns only selected data at `internal/connectors/engine/direct_write.go:212-234`. Audit: 254/274 mutation payloads contain omitted nested object/interface/union fields.
- **End-to-end path:** GraphQL payload type -> shallow selection generator -> fixed mutation document -> provider returns only selected scalars -> operation result -> reconciliation.
- **Concrete flow:** `createIssue` reports success but cannot return the created issue ID, number, or URL.
- **Six-surface impact:** reverse ETL and direct write cannot reconcile most GraphQL resources; ETL/direct read/binary unaffected.
- **Production risk:** successful writes lose authoritative resource identity, making retry/reconciliation unsafe.
- **Minimal fix:** generate bounded schema-derived nested selections with depth/field ceilings and the secret policy from `PF-CF-B06`; include stable identity/status/URL fields and scalar status/client mutation ID.
- **Red-first tests:** `TestGeneratedGraphQLMutationDocumentsPreserveResourceIdentity` for `createIssue` and `addComment`; assert resource IDs/URLs and occurrence IDs exact, classified secrets masked, and limits enforced.

### PF-CF-B08 — GraphQL `Int` accepts values outside the signed 32-bit domain

- **Severity / source / state:** BLOCKER; `PFM-B08`; newly exposed gap distinct from prior exact-enum and current CLI-fidelity findings.
- **Exact evidence:** `scalarSchema` maps `Int` to unbounded integer at `scripts/gen-github-graphql-parity.mjs:160-173`; generated `requiredApprovingReviewCount` at lock `:21100-21105` and operations `:12633-12635`; collection-only bound checker `internal/connectors/engine/graphql_operation.go:439-468`; schema lacks numeric range fields at `internal/connectors/engine/schema.go:87-116`.
- **End-to-end path:** GraphQL source `Int` -> generic integer schema/flag -> native coercion -> schema validation -> JSON variables -> provider coercion.
- **Concrete flow:** `2147483648` passes local validation on a 64-bit build and fails only at GitHub; the same occurs inside nested input objects/lists.
- **Six-surface impact:** GraphQL direct read, reverse ETL, and direct write blocked; ETL/binary unaffected.
- **Production risk:** approved plans contain provider-invalid values and fail after local validation.
- **Minimal fix:** encode exact min/max or walk the GraphQL type/value tree before I/O, including nested lists/input objects.
- **Red-first tests:** `TestGraphQLIntUsesSigned32BitDomain`: both bounds pass; adjacent values fail pre-server at root, nested object, and list locations.

### PF-CF-B09 — Embedded v2 GraphQL projection is not authenticated by its advertised digest

- **Severity / source / state:** BLOCKER; `PFM-B09`; claimed-closure residual of `CF-B01`.
- **Exact evidence:** REST import hashes fetched bytes at `cmd/connectorgen/sourceimport.go:650-660`; v2 GraphQL path bypasses fetch/digest at `:716-740` and copies unrelated external digest metadata at `:748-780`; permissive test `cmd/connectorgen/sourceimport_test.go:125-147`; validator compares copied metadata at `cmd/connectorgen/sourceprojection.go:645-685`.
- **End-to-end path:** embedded lock projection -> import bypass -> descriptor inherits unrelated digest -> validator compares copied value -> generator trusts unauthenticated type system.
- **Concrete flow:** edit one embedded GraphQL signature/type/count but retain the external source SHA; import and validation still pass.
- **Six-surface impact:** GraphQL direct read, reverse ETL, and direct write provenance blocked; REST ETL and binary source locks remain authenticated.
- **Production risk:** generated executable API can drift from the source identity certification claims.
- **Minimal fix:** hash canonical embedded query/mutation/type-system bytes with explicit `projection_sha256/projection_bytes`, or fetch and verify an immutable artifact; never copy an unrelated digest.
- **Red-first tests:** `TestSourceImportVersion2RejectsEmbeddedGraphQLProjectionDigestDrift`; mutate a signature, type field, and root count; define deterministic reorder behavior.

### PF-CF-B10 — Foundation evidence gate accepts evidence from another implementation snapshot

- **Severity / source / state:** BLOCKER; `PFM-B10`; claimed-closure residual of `CF-W08`.
- **Exact evidence:** manifest fields at `cmd/connectorgen/evidencegate.go:15-29`; validation checks formatting/reviewed SHA only at `:76-149`; manifest `data/cli-current-foundations-main-integration-r1/evidence-manifest.json:1-7` records code/base/component identities that do not equal the frozen HEAD/base, yet the real gate passes.
- **End-to-end path:** checked evidence manifest -> `runEvidenceGate` -> partial SHA validation -> affirmative Foundation status for current checkout.
- **Concrete flow:** keep test evidence from `808896a.../114a677...`, review a later `8a8a866.../e62ae21...` tree, and receive a passing evidence gate.
- **Six-surface impact:** all six surfaces can be certified by tests run against different code.
- **Production risk:** release certification is not evidence of the shipped artifact.
- **Minimal fix:** bind exact clean implementation HEAD, diff base, component SHAs, and preserving merge to the reviewed graph; keep `reviewed_sha` as review identity, not code identity.
- **Red-first tests:** mutate code SHA, base SHA, every component SHA, and preserving merge independently; each must fail; regenerated exact frozen identities pass.

### PF-CF-B11 — Certification matrices treat historical accepted evidence as current forever

- **Severity / source / state:** BLOCKER; `PFM-B11`; claimed-closure residual of `CF-W08`.
- **Exact evidence:** accepted evidence model `cmd/connectorgen/certificationmatrix.go:159-183`; proof contains binary fingerprints at `certificationproof.go:45-59`; validation/matching at `certificationmatrix.go:1394-1435,1508-1575` checks identity shape, not current implementation/source/config subject. Audit: 985 GitHub records span 13 PM binary digests.
- **End-to-end path:** historical evidence JSON -> shape validation -> identity-only match -> `LiveTested/LiveEvidence` -> capability/workflow/sync cell -> generated certification status.
- **Concrete flow:** change an operation mapping or PM binary while retaining old connector/operation evidence identity; the cell remains live-tested.
- **Six-surface impact:** every surface can inherit stale affirmative certification.
- **Production risk:** changed executable behavior ships under historical proof.
- **Minimal fix:** add a deterministic subject fingerprint over binary/build, declarations, source/projection digest, command mapping, relevant config, and proof protocol; preserve nonmatches as historical/stale.
- **Red-first tests:** `TestCertificationEvidenceBecomesStaleWhenSubjectChanges`; independently change operation field, CLI `maps_to`, source digest, and PM binary digest; each clears live status until new evidence arrives.

### PF-CF-B12 — `params-import` fails open for deleted routes and unresolved parameter references

- **Severity / source / state:** BLOCKER; `PFM-B12`; claimed-closure residual of `CF-B02`.
- **Atomic claims:** absent provider path/method silently skips; unresolved OpenAPI parameter reference silently skips.
- **Exact evidence:** zero-change success at `cmd/connectorgen/paramsimport.go:64-82`; absent route `:207-217`; unresolved references `:465-488`.
- **End-to-end path:** provider OpenAPI -> route/ref lookup -> missing input omitted -> partial/zero parameter set -> `surface-sync` -> stale CLI/API -> `--check` success.
- **Concrete flow:** delete a provider method or break a component `$ref`; generated artifacts remain unchanged and the check exits zero.
- **Six-surface impact:** ETL/reverse ETL may retain stale assumptions; direct read/write can silently lose path/query/header inputs; no current binary-specific occurrence was proved.
- **Production risk:** provider API deletion/rename silently leaves an executable but invalid local surface.
- **Minimal fix:** fail every executable identity with absent method/path or unresolved ref; permit only a source-identity-bound, reviewed, expiring exception and write nothing during `--check`.
- **Red-first tests:** `TestParamsImportRejectsDeletedProviderRoute` and `TestParamsImportRejectsUnresolvedParameterRef`, covering component and legacy refs with exact source diagnostics and byte-stable output.

### PF-CF-B13 — Public receipts and GraphQL error metadata can disclose concrete or declared secrets

- **Severity / source / state:** BLOCKER; `PFR-BL-01`; claimed-closure residual of `CF-B08`, `CF-B11`, and `CF-B23`.
- **Atomic claims:** raw REST/write receipt bodies evade literal masking through JSON escaping and declared response locations are not applied to the receipt; GraphQL retained error/extension metadata never receives concrete credentials or declared secret paths.
- **Exact evidence:** `internal/connectors/engine/response_receipt.go:30-47`; direct-read convenience-only redaction at `internal/connectors/engine/direct_read.go:143-194,490-532`; literal variants at `internal/connectors/connectors.go:980-1007,1094-1119,1192-1212`; GraphQL receipt/error metadata at `internal/connectors/engine/graphql_operation.go:665-710,797-832`; declared Zendesk token fields `internal/connectors/defs/zendesk-support/operations.json:3461-3475`; public emission `internal/cli/cli.go:1415-1438`.
- **End-to-end path:** provider response -> complete receipt constructed before declaration policy -> convenience body redacted but receipt raw/decoded body or GraphQL metadata remains unsafe -> commandrunner/App -> persistence/CLI.
- **Concrete flow:** configured secret `pa"ss` appears as JSON `pa\\\"ss`, or inside a GraphQL partial-error message with no token keyword; literal replacement misses it and output serializes it.
- **Six-surface impact:** unsafe shared receipt boundary affects ETL, reverse ETL, direct read, direct write, binary download, and binary upload.
- **Production risk:** credential/PII disclosure in stdout and durable run artifacts.
- **Minimal fix:** keep an immutable internal receipt; at the single public boundary decode JSON with `UseNumber`, traverse exact configured/declared secret value locations, mask values without removing keys, and canonicalize raw public JSON; use explicit withheld/encoded handling for invalid/binary bytes; sanitize GraphQL errors/extensions with the same authority.
- **Red-first tests:** table-drive quote, backslash, `<`, `>`, `&`, non-ASCII, and base64-sensitive secrets through REST raw/decoded bodies, GraphQL partial errors/extensions, declared Zendesk fields, and binary/error receipts; serialized public output contains no concrete/proven encoding, ordinary fields and occurrence IDs remain exact, internal receipt is unchanged.

### PF-CF-B14 — Public sanitizers corrupt ordinary provider keys, IDs, headers, and messages

- **Severity / source / state:** BLOCKER; `PFR-BL-02`; claimed-closure regression of `CF-B23`.
- **Atomic claims:** substring replacement with a concrete short secret rewrites keys/header names/ordinary values; keyword-only GraphQL/header heuristics mask provider text without any concrete secret.
- **Exact evidence:** key/header mutation at `internal/connectors/connectors.go:1019-1031,1058-1082,1208-1212`; header-name heuristic `internal/connectors/engine/operation_headers.go:21-44`; GraphQL message heuristic `internal/connectors/engine/graphql_operation.go:817-832`; current preservation test `internal/connectors/commandrunner/content_preservation_test.go:193-213` lacks short-secret collision coverage.
- **End-to-end path:** complete provider result -> public sanitizer recursively rewrites map keys, header names, strings, and matching numbers -> App/CLI publishes altered provider truth.
- **Concrete flow:** configured secret `id` changes `occurrence_id` to `occurrence_[masked]`; an ordinary “Unknown token type” error is replaced solely because it contains `token`.
- **Six-surface impact:** all six surfaces can lose record/attempt identity, repeatable headers, diagnostics, or receipt truth.
- **Production risk:** audit, reconciliation, retry, and downstream parsers operate on fabricated output.
- **Minimal fix:** never mutate keys or header names; mask only exact scalar values authorized by configured material or explicit declaration, preserve field/name presence, and remove keyword-only redaction.
- **Red-first tests:** configured secrets `id`, `token`, `0`, and one-character values; return `occurrence_id`, `trained_tokens`, `token_type`, `WWW-Authenticate`, duplicate headers, large IDs, and “Unknown token type”; only a scalar exactly equal to a proven secret may change.

### PF-CF-B15 — `WriteHook` bypasses sealed approval and drops compound provider receipts

- **Severity / source / state:** BLOCKER; `PFR-BL-03`; mixed: claimed-closure residual of `CF-B17`, plus newly exposed compound-receipt loss.
- **Atomic claims:** hooks execute the original mutable record and a request sequence not represented by the approved `PreparedWrite`; handled hooks append counts but no response receipt per attempted provider mutation.
- **Exact evidence:** incomplete hook guard `internal/connectors/engine/write_prepare.go:35-55`; destructive classifier `engine/write_gate.go:26-31`; hook-first execution and late equality check `engine/write.go:317-427`; GitHub handled actions `internal/connectors/hooks/github/hooks.go:257-286`, compound close `:369-395`, PR create/update `:432-498`; Ashby delegate `internal/connectors/native/ashby/engine_delegate.go:75-93`.
- **End-to-end path:** App/CLI declarative preview -> digest one `PreparedRequest` -> approval -> `executeApprovedWrite` invokes hook with caller-owned record -> hook emits one or more different requests -> responses discarded -> count-only terminal result.
- **Concrete flow:** approve a GitHub close-resource preview, mutate the original record while approval blocks, then the hook may send a comment plus PATCH derived from the changed record and return neither response.
- **Six-surface impact:** hook-backed ETL destinations, reverse ETL, and direct writes blocked; future binary upload/compound hooks cannot use this seam safely; reads/download unaffected.
- **Production risk:** approval authorizes neither actual bytes nor effect count/order; partial commits cannot be reconciled.
- **Minimal fix:** replace execute override with a prepare hook returning an exact ordered list of sealed requests plus projector, or fail closed for every classifier-handled action until supported; execute pinned material only and return a bounded receipt for every attempted request including the terminal failure.
- **Red-first tests:** for all GitHub handled actions and representative Ashby create/delete, mutate caller input during approval; assert exact previewed sequence/bytes or zero I/O; success returns every response/ID, failure at N retains receipts 1..N, counts reflect committed effects, public output is secret-safe.

### PF-CF-B16 — Operation direct writes re-materialize aliased state instead of executing sealed bytes

- **Severity / source / state:** BLOCKER; `PFR-BL-04`; claimed-closure residual of `CF-B17`.
- **Atomic claims:** REST/GraphQL execution re-marshals aliased object/config/secret state instead of sealed bytes; multipart execution re-reads a replaceable file and caller-owned approved-digest map.
- **Exact evidence:** live prepared state `internal/connectors/engine/direct_write.go:38-59`; sealed REST bytes plus retained object `:1578-1620`; sealed GraphQL bytes plus retained payload `:1699-1738`; execution re-marshals at `:138-163` through `internal/connectors/connsdk/http.go:625-634`; shallow defaults `internal/connectors/engine/read.go:521-537`.
- **End-to-end path:** prepare/digest -> approval blocks -> caller mutates nested input/config/secrets/digest map or swaps upload file -> execution resolves/re-marshals/re-reads live state -> wire request differs from preview.
- **Concrete flow:** replace a multipart file and update the aliased approved SHA map after approval; upload passes the changed bytes although the signed preview did not change.
- **Six-surface impact:** operation-backed ETL destinations, reverse ETL, direct write, and binary upload blocked; reads/download unaffected.
- **Production risk:** approval and audit digest do not bind the actual credential, endpoint, body, or file.
- **Minimal fix:** deep-copy all security-relevant runtime/maps/forms at preparation, execute the exact body bytes through a bounded byte method, and bind/reverify stable file identity plus digest immediately before send without caller-owned state.
- **Red-first tests:** pause approval and mutate nested input, Config, Secrets, approved digest map, and file; REST JSON, GraphQL, form, and multipart must send exact original material or refuse pre-I/O; run under `-race`.

### PF-CF-B17 — Direct-read results are erased at commandrunner and native adapter boundaries

- **Severity / source / state:** BLOCKER; `PFR-BL-05`; claimed-closure residual of `CF-B11` on distinct command/native paths.
- **Atomic claims:** commandrunner zeroes legacy and post-navigation error results; Ashby zeroes an engine result on logical envelope failure.
- **Exact evidence:** engine returns result-plus-error at `internal/connectors/engine/direct_read.go:480-521`; runner erasure `internal/connectors/commandrunner/runner.go:569-615,667-698`; CLI can emit only retained receipts at `internal/cli/cli.go:1344-1355`; Ashby erasure `internal/connectors/native/ashby/engine_delegate.go:111-119`.
- **End-to-end path:** provider response -> engine populated result/receipt plus error -> command/native wrapper returns zero result -> CLI receives only error -> no terminal result envelope.
- **Concrete flow:** a 4xx, malformed JSON, pagination/navigation failure, or Ashby `{success:false}` has a received response but stdout contains no receipt.
- **Six-surface impact:** direct read blocked; ETL source diagnostics can be lost through command/native paths; binary runner must retain analogous results; write/upload unaffected.
- **Production risk:** provider status, request IDs, partial data, and pagination evidence vanish from nonzero outcomes.
- **Minimal fix:** preserve `Result` for every post-provider/post-validation error whether or not receipt is currently nonnil; wrappers return `result, err`; CLI emits one masked envelope before the categorized error.
- **Red-first tests:** legacy 4xx/malformed/pagination/navigation, operation navigation, and Ashby false/invalid envelopes; exactly one complete terminal JSON envelope plus nonzero exit, no duplicate body or credential on stderr.

### PF-CF-B18 — Redirect/retry/cancellation transitions discard the last provider response

- **Severity / source / state:** BLOCKER; `PFR-BL-06`; claimed-closure residual of `CF-B13` beyond the repaired declarative mutation path.
- **Atomic claims:** stream transport loses response on redirect refusal/backoff cancellation; buffered transport overwrites a captured 429/503 with later cancellation/transport failure.
- **Exact evidence:** stream loss `internal/connectors/connsdk/stream.go:102-107,137-161`; buffered loss `internal/connectors/connsdk/http.go:1186-1189,1223-1243,1296-1301`; binary receipt can only use returned stream/typed HTTP error at `internal/connectors/engine/binary_read.go:162-200`.
- **End-to-end path:** attempt receives 3xx/429/503 -> policy/backoff/next attempt -> cancellation or transport error -> latest provider response discarded -> engine/CLI has no receipt or rate evidence.
- **Concrete flow:** 503 with `Retry-After`, then context cancellation during sleep, returns only `context canceled`; `errors.As` no longer finds the provider HTTP error.
- **Six-surface impact:** shared HTTP transitions affect ETL, reverse ETL, direct read/write, binary download, and binary upload.
- **Production risk:** real provider outcomes become “no response,” breaking audit, retry parking, and diagnosis.
- **Minimal fix:** retain latest bounded response/typed HTTP error across attempts; join it with terminal cause; retain refused 3xx without following; close superseded bodies deterministically.
- **Red-first tests:** 302/307 refusal with untouched target and retained 3xx; 503 backoff cancellation; 429/503 then reset/cancel; `errors.As` retains HTTP error and CLI emits one nonzero safe envelope.

### PF-CF-B19 — Status-check operations lack complete success/error receipts

- **Severity / source / state:** BLOCKER; `PFR-BL-07`; newly exposed API-reachability gap within the shared receipt contract.
- **Atomic claims:** success result models only status/body bytes/admitted headers; transport/read/header errors zero the result and commandrunner/CLI erase every received response.
- **Exact evidence:** result type `internal/connectors/connectors.go:835-846`; engine zero-result errors `internal/connectors/engine/status_check.go:26-79`; runner erasure `internal/connectors/commandrunner/runner.go:2393-2408`; CLI success/failure coverage `internal/cli/cli.go:1344-1355,1392-1413`.
- **End-to-end path:** declared HEAD/status command -> requester receives response -> narrow projection or zero error result -> CLI success-only status view/generic error.
- **Concrete flow:** HEAD receives 404 or a redirect refusal with duplicate trace headers; no complete provider receipt survives.
- **Six-surface impact:** status-based ETL/preflight and direct-read `rest_status` commands blocked; write/binary surfaces not directly implicated.
- **Production risk:** a reachable operation kind violates complete provider-output and single-envelope contracts.
- **Minimal fix:** add shared receipt to status result, populate from every received/typed-error response, preserve result-plus-error through runner/CLI, and keep HEAD body policy explicit/bounded.
- **Red-first tests:** 204, final 404, redirect refusal, retry cancellation, and read error; retain complete repeatable headers/status/body metadata, one safe terminal envelope, correct nonzero classification, no body decoding.

### PF-CF-B20 — Destination authorization is stale by the time side effects occur

- **Severity / source / state:** BLOCKER; `PFR-BL-08`; newly exposed authorization-boundary residual related to, but distinct from, ownership CAS and auth-cohort fencing.
- **Exact evidence:** contract requires revalidation at `internal/synctransport/types.go:396-443`; ordinary early admission and later apply `internal/synctransport/orchestrator.go:167-243`; full overwrite `:396-450`; Arrow serial `arrow_fast_path_controller.go:96-151`; pipeline `arrow_fast_path_pipeline.go:202-287`; live callback `internal/app/declarative_typed_destination_approval.go:192-215`; native PostgreSQL/Arrow gaps `native/postgres/transport_destination.go:119-139`, `arrow_full_overwrite_transport.go:82-114`.
- **End-to-end path:** durable authorization -> early orchestrator admission -> blocking stage/reopen/queue/transform/provision -> authorization expires/revokes -> adapter mutates/publishes without a final live check.
- **Concrete flow:** pause after warehouse reopen but before PostgreSQL apply, revoke authorization, resume, and observe the database mutation.
- **Six-surface impact:** ETL and transport-backed reverse ETL blocked; binary upload is affected when used as a destination; standalone read/write/download paths are separate.
- **Production risk:** revoked or expired authority permits external mutation.
- **Minimal fix:** carry a non-serializable live checker into every adapter/session and invoke it immediately before each actual mutation/publish after blocking preparation; early admission may remain only as an optimization.
- **Red-first tests:** ordinary, full-overwrite, Arrow serial/pipeline: block at final pre-I/O edge, revoke/expire, assert zero effect/publish/checkpoint; revoke between segments and retain earlier acknowledged effects without replay.

### PF-CF-B21 — CLI numeric coercion rounds provider values and minimums

- **Severity / source / state:** BLOCKER; `PFR-BL-09`; claimed-closure residual of `CF-B19` at a different boundary.
- **Atomic claims:** input lexemes are coerced through `ParseFloat`/platform `Atoi`; command minimum metadata and comparison use `float64` even when the schema later supports exact numbers.
- **Exact evidence:** exact engine comparison `internal/connectors/engine/schema.go:604-678`; float minimum metadata `internal/connectors/command_surface.go:29-53`, `engine/bundle.go:1051-1069`; float coercion/comparison `internal/connectors/commandrunner/runner.go:1433-1461,2089-2124`; shipped number flags `internal/connectors/defs/github/cli_surface.json:41095-41100`.
- **End-to-end path:** CLI numeric token -> float/platform coercion and minimum check -> mapped request -> exact engine sees already-rounded value -> approval/wire bytes bind wrong number.
- **Concrete flow:** `9007199254740993` becomes `9007199254740992`, or `0.10000000000000001` rounds before preview.
- **Six-surface impact:** typed numeric parameters affect ETL, reverse ETL, direct read/write, binary download, and binary upload metadata.
- **Production risk:** wrong identifiers/amounts are authorized and sent; exact bounds admit/reject incorrectly.
- **Minimal fix:** carry declaration and input lexemes as `json.Number`, compile exact rational minima, compare before mapping, and use explicit integer domains independent of host architecture.
- **Red-first tests:** real commands with `0.10000000000000001`, adjacent >2^53 integers, negative/exponent forms, exact min/min-epsilon; preview and wire preserve value on 32-/64-bit builds for REST and GraphQL.

### PF-CF-B22 — Binary no-overwrite publication can expose, overwrite, delete, or lose the final file

- **Severity / sources / state:** BLOCKER; merged `PFR-BL-10` + `ORCH-PF-B12`; newly exposed data-loss gap that invalidates the prior “atomic download” passing assertion.
- **Atomic claims:** visible zero-byte reservation can survive a crash; success rename replaces a competing file and failure cleanup deletes it; containing directory is not synced after publication.
- **Exact evidence:** `internal/connectors/engine/binary_read.go:461-485` creates the visible final reservation, `:487-500,519-535` removes the final name on multiple failures, and `:530-537` uses replacing rename without a containing-directory sync.
- **Exact merge rationale:** both ledgers cite `internal/connectors/engine/binary_read.go:461-537`, the same reservation/temp/rename state machine, same foreign-file overwrite/delete failure, and same atomic no-replace plus owned-cleanup fix. Orchestration adds crash and directory-durability edges but no distinct implementation boundary.
- **End-to-end path:** binary command -> reserve final name with `O_EXCL` -> stream/fsync temp -> another process replaces placeholder or process crashes -> ordinary rename/cleanup -> foreign/zero/lost final file.
- **Concrete flow:** remove the zero-byte reservation and create a sentinel while download streams; success overwrites it, while an injected later error removes it. A crash can leave only the empty reservation.
- **Six-surface impact:** binary download blocked/data loss; binary upload must not copy this staging pattern; ETL/reverse/direct read/write unaffected.
- **Production risk:** local user data can be overwritten/deleted; success may not survive crash.
- **Minimal fix:** keep final absent, write/fsync owned hidden temp, publish through an atomic no-replace primitive under `os.Root`, remove only owned temp/inode entries, and fsync the containing directory.
- **Red-first tests:** `TestBinaryDownload_NoOverwritePublicationIsCrashAndRaceSafe`: subprocess death leaves no final; foreign sentinel inserted pre-publish survives success/error paths with same inode/bytes; temp cleanup and directory-sync completion are asserted; no outside-root touch.

### PF-CF-B23 — Page cursors bypass field authority and encoded-size admission

- **Severity / source / state:** BLOCKER; `PFR-BL-11`; claimed-closure residual of `CF-B21`/`CF-B22`.
- **Atomic claims:** same-origin next/link URLs can add undeclared query controls; all cursor forms bypass dangerous-character and encoded-byte caps, including native SQS.
- **Exact evidence:** CLI accepts cursor unchecked `internal/cli/cli.go:1740-1766`; runner removes page flags before validation `internal/connectors/commandrunner/runner.go:560-570`; insertion `internal/connectors/engine/direct_read_paginate.go:599-609`; URL admission allows arbitrary query `:466-501`; SQS form insertion `internal/connectors/native/amazon-sqs/direct_read.go:76-100`.
- **End-to-end path:** CLI cursor -> typed validation bypass -> paginator/native request inserts opaque token or URL -> authenticated request carries undeclared/unbounded query/form bytes.
- **Concrete flow:** same-origin/same-path continuation URL appends `?admin=true`, or an oversized/control-bearing SQS token reaches SigV4 signing.
- **Six-surface impact:** ETL/direct read pagination blocked; other writes/binary surfaces unaffected unless they reuse continuation channel.
- **Production risk:** closed request authority can be bypassed after page one; control/oversized input reaches signing/logging/provider I/O.
- **Minimal fix:** preferably issue a signed opaque continuation binding connector/operation/allowed query/size; otherwise admit only paginator-owned keys and enforce control, duplicate, collision, UTF-8, percent/form-encoded caps before auth.
- **Red-first tests:** replay a legitimate emitted cursor; reject unknown query, duplicate paginator keys, oversize UTF-8/percent expansion, CR/LF/controls, and oversized SQS token before any request.

### PF-CF-B24 — Native Amazon SQS direct reads never produce complete provider receipts

- **Severity / source / state:** BLOCKER; `PFR-BL-12`; newly exposed native-path residual of `CF-B11`.
- **Atomic claims:** successful native SQS results omit operation/headers/receipt; post-I/O, XML, oversize, and read failures discard the received response entirely.
- **Exact evidence:** production factory `internal/connectors/native/nativeset/factories.go:20-29`; zero/sparse results `internal/connectors/native/amazon-sqs/direct_read.go:19-49`; response type and loss paths `connection.go:166-169,201-238`; shared receipt contract `internal/connectors/connectors.go:474-513`.
- **End-to-end path:** installed SQS command -> native SigV4/XML executor -> provider response -> native adapter discards metadata/error response -> commandrunner/CLI cannot publish receipt.
- **Concrete flow:** AWS returns a 4xx XML error with `x-amzn-requestid`; output has neither status/header/body nor operation identity.
- **Six-surface impact:** direct read and SQS ETL observability blocked; SQS reverse/write should share the corrected model; direct write/binary unaffected currently.
- **Production risk:** provider-specific audit/retry contract hole and unauditable AWS failures.
- **Minimal fix:** extend native response with cloned headers/raw/presence/byte metadata; return it with post-I/O errors; construct/sanitize shared receipt and preserve result-plus-error through CLI.
- **Red-first tests:** each installed SQS read class with 200, AWS 4xx XML, malformed XML, cap+1, and injected read error; exact duplicate headers/request ID/raw+decoded bounded body and operation survive; one nonzero safe CLI envelope.

### PF-CF-B25 — Native Amazon SQS can forward a session credential across origins

- **Severity / source / state:** BLOCKER; `PFR-BL-13`; claimed-closure residual of `CF-B14` outside declarative HTTP.
- **Exact evidence:** SQS adds/signs `X-Amz-Security-Token` at `internal/connectors/native/amazon-sqs/connection.go:201-223`; `transportpolicy.HTTPClient` no-ops for non-destructive context at `internal/connectors/transportpolicy/policy.go:22-41`; direct read is logically non-destructive, so default redirect behavior can copy the custom session-token header cross-origin.
- **End-to-end path:** SQS direct read -> SigV4/auth headers -> ambient client -> origin returns 302/307 -> default client follows to attacker origin with custom token -> native receipt path also loses redirect history.
- **Concrete flow:** endpoint redirects to a second server; signature may be invalid for the new host, but the reusable temporary session token reaches it.
- **Six-surface impact:** SQS ETL/direct read blocked; reverse SQS writes need the same uniform policy; other direct/binary surfaces use separate clients.
- **Production risk:** temporary AWS credential exfiltration.
- **Minimal fix:** clone client per SQS request and refuse all redirects by default; if an explicit same-origin policy is ever admitted, strip every signing/auth header on origin change and retain the first 3xx receipt without contacting target.
- **Red-first tests:** production `New()` connector against 302 and 307 origin->second server with session credentials; target receives no request/auth header, first 3xx survives in one safe nonzero envelope, same-origin non-redirect success remains green.

### PF-CF-B26 — Stream ownership is claimed only after provider/warehouse side effects

- **Severity / source / state:** BLOCKER; `ORCH-PF-B01`; explicit claimed-closure regression of `CF-B24`—the cited fix commit does not add a pre-I/O claim.
- **Atomic claims:** ordinary/full-overwrite/Arrow paths perform effects before App CAS; CDC publishes WAL/table artifacts before state CAS.
- **Exact evidence:** snapshot without claim `internal/app/transport_dispatch.go:235-253`; only CAS `:282-313`; ordinary/full-overwrite effects first `internal/synctransport/orchestrator.go:237-283,435-497`; Arrow `arrow_fast_path_controller.go:145-216`, `arrow_fast_path_pipeline.go:125-158,277-283`; CDC `internal/app/change_capture_dispatch.go:109-253`. Tests require loser effects at `internal/app/transport_dispatch_test.go:1056-1124,1383-1392`.
- **End-to-end path:** two processes load same stream generation -> both stage/apply/publish -> winner commits checkpoint -> loser fails late CAS -> provider/local warehouse may contain loser state while App names winner.
- **Concrete flow:** stale full-overwrite process publishes older data after winner, then loses only local checkpoint CAS.
- **Six-surface impact:** ETL blocked across ordinary/full/Arrow/CDC; transport-backed reverse/direct writes and future binary-upload destinations affected; standalone reads/download unaffected.
- **Production risk:** duplicate or out-of-order provider effects and state/provider divergence.
- **Minimal fix:** acquire a durable connection+stream lease/fence with monotonic generation and stable work identity before source/stage/provider I/O; require effects to reject stale fences and checkpoint to retire the same fence; recovery reconciles receipts before replay.
- **Red-first tests:** `TestTransport_TwoAppsFenceBeforeAnySideEffect`, table-driven for append/upsert, full overwrite, serial/pipeline Arrow, CDC, expiry/crash/renewal loss/takeover; loser has zero effect or provider atomically rejects its stale fence.

### PF-CF-B27 — Pagination budget is misreported as source exhaustion

- **Severity / source / state:** BLOCKER; `ORCH-PF-B02`; newly exposed delivery gap.
- **Atomic claims:** full overwrite publishes a capped prefix as complete; append/incremental modes cannot persist/resume beyond the cap and replay the prefix.
- **Exact evidence:** GitHub modes `internal/connectors/defs/github/sync_transport.json:3-55`; default one page `internal/app/issue_label_warehouse_transport.go:1171-1254`; cap break without incomplete state `internal/connectors/engine/read.go:307-326,424-433`; current test encodes first-page success `internal/app/transport_composition_test.go:2442-2519`; full-overwrite publication `internal/synctransport/orchestrator.go:461-497`.
- **End-to-end path:** source declaration -> default/max page budget -> engine silently breaks -> source returns nil EOF-equivalent -> orchestrator publishes/commits truncated data.
- **Concrete flow:** three-page GitHub collection with omitted `transport_max_pages` publishes page one as a successful full overwrite.
- **Six-surface impact:** ETL data loss and reverse destinations fed a truncated set; direct-read pagination API is separate; direct write/binary unaffected.
- **Production risk:** silent deletion of destination rows and incremental livelock/replay.
- **Minimal fix:** return typed `exhausted` versus `budget_stopped` with opaque continuation; forbid full-overwrite publish without proven exhaustion; persist/resume bounded continuation for incremental modes or refuse caps pre-I/O.
- **Red-first tests:** `TestDeclarativeTransport_PageBudgetIsNotEOF` with omitted/1/2/unlimited caps, full-overwrite zero publish on budget stop, incremental page N+1 restart and exactly-once records, restart between runs.

### PF-CF-B28 — Connector declarations self-certify keyed/idempotent delivery

- **Severity / source / state:** BLOCKER; `ORCH-PF-B03`; claimed-closure residual of `CF-B27`.
- **Exact evidence:** App derives conformance references from candidate declarations `internal/app/issue_label_warehouse_transport.go:78-139`; verifier accepts those values `internal/app/transport_composition.go:12-31`, `internal/synctransport/definition_composition.go:46-83`; destination checks only declared `keyed` plus route `issue_label_warehouse_transport.go:418-477`; production-shaped non-idempotent POST test is admitted at `transport_composition_test.go:405-489`; engine only disables in-request retry `internal/connectors/engine/write.go:645-674`.
- **End-to-end path:** connector claims keyed -> verifier authority generated from same claim -> registry admits -> POST effect succeeds -> read-back/checkpoint fails -> whole run replay creates another object.
- **Concrete flow:** synthetic create POST with no idempotency header invents its own conformance ID and is accepted as keyed.
- **Six-surface impact:** declarative ETL/reverse destinations blocked; direct standalone write does not establish replay safety; binary upload cannot be promoted as keyed without proof.
- **Production risk:** duplicate provider objects under a falsely advertised delivery guarantee.
- **Minimal fix:** independent immutable evidence keyed by executor and definition/action digest; require stable preview/workset idempotency key or independently certified intrinsic mechanism; reject unsupported claimed-keyed actions before source I/O.
- **Red-first tests:** `TestDeclarativeDestination_ClaimedKeyedWithoutIndependentProofIsRejected`; positive evidence+header case retries after provider success with identical key and one mutation; action/evidence drift invalidates admission.

### PF-CF-B29 — Provider read-back ignores the mutation receipt and scans an unrelated prefix

- **Severity / source / state:** BLOCKER; `ORCH-PF-B04`; claimed-closure residual of `CF-B07`.
- **Atomic claims:** engine never consumes `ReadBackRequest.Receipt`; verifier scans an unfiltered collection prefix bounded by `MaxRecords`.
- **Exact evidence:** App passes public acknowledgement output as receipt `internal/app/issue_label_warehouse_transport.go:330-340`; engine ignores it and calls unfiltered `Read` `internal/connectors/engine/connector.go:242-272`; cap abort `:260-264`; acknowledgement is already public/sanitized `issue_label_warehouse_transport.go:248-267`.
- **End-to-end path:** mutation -> public output attached to acknowledgement -> read-back policy -> engine ignores locator -> collection starts at page one -> expected row beyond cap not found -> checkpoint withheld -> mutation replayed.
- **Concrete flow:** collection exceeds `max_records`; newly written row is on page two, so every verification scans only page one despite a receipt containing its identity.
- **Six-surface impact:** generic ETL/reverse destination durability blocked; underlying direct read works but is not receipt-bound; standalone direct write/binary unaffected.
- **Production risk:** successful mutations replay forever or are “verified” without causal linkage.
- **Minimal fix:** retain a complete private mutation locator distinct from public output; declaration projects it into a point query or bounded pagination that ends only after all expected identities; prove batch/readback capacity at preflight.
- **Red-first tests:** `TestDeclarativeDestination_ReadBackUsesInternalReceiptLocator`; expected row beyond page one, missing/foreign/changed receipt prevents checkpoint, eventual-consistency retries reuse exact locator, public output masks secret while private locator stays complete.

### PF-CF-B30 — Declarative source cloning rounds integers above 2^53

- **Severity / source / state:** BLOCKER; `ORCH-PF-B05`; newly exposed numeric-fidelity gap.
- **Exact evidence:** clone uses JSON marshal + default unmarshal at `internal/app/issue_label_warehouse_transport.go:1574-1583`; used before staging at `:1147-1164,1211-1230`; provider decoder preserves `json.Number` at `internal/connectors/connsdk/extract.go:10-18`; PostgreSQL emits `int64` at `internal/connectors/native/postgres/reader.go:99-113`.
- **End-to-end path:** provider/native record -> declarative callback clone -> default JSON numbers become float64 -> stage/map/approval/write/checkpoint.
- **Concrete flow:** `json.Number("9007199254740993")` becomes a rounded float before destination authorization.
- **Six-surface impact:** ETL and reverse writes fed by declarative sources blocked; direct read receipts preserve numbers; standalone write/binary unaffected.
- **Production risk:** provider IDs/amounts change before delivery and checkpoint identity.
- **Minimal fix:** use a typed recursive clone or `UseNumber` normalization; preserve `json.Number`, signed/unsigned integers, nested arrays/maps; reject unsupported mutable types.
- **Red-first tests:** `TestDeclarativeTransportClone_PreservesLargeNumbers` from REST and PostgreSQL through stage, destination/read-back/checkpoint, including nested/boundary values and caller non-mutation.

### PF-CF-B31 — Provider read-back compares numeric Go representations instead of values

- **Severity / source / state:** BLOCKER; `ORCH-PF-B06`; newly exposed numeric-fidelity gap distinct from cloning and CLI coercion.
- **Atomic claims:** expected-field matcher uses `reflect.DeepEqual`; identity hash encodes raw Go numeric types.
- **Exact evidence:** comparison `internal/app/issue_label_warehouse_transport.go:361-395`; identity/hash `:398-415`, `internal/app/util.go:307-312`; provider JSON uses `json.Number` `internal/connectors/engine/read.go:1463-1474`; native source can use `int64` `internal/connectors/native/postgres/reader.go:99-113`.
- **End-to-end path:** native/staged numeric -> provider write -> read-back JSON number -> raw-type equality/hash mismatch -> checkpoint denied -> write replay.
- **Concrete flow:** `int64(42)` and `json.Number("42")` are semantically equal but compare/hash differently.
- **Six-surface impact:** ETL/reverse destination verification blocked; direct read returns truth but does not compare; standalone write/binary unaffected.
- **Production risk:** correct successful mutations replay; identity mismatches above 2^53 cannot be repaired safely through float.
- **Minimal fix:** schema-aware canonical JSON numeric comparison and identity encoding at arbitrary precision, with explicit `42` versus `42.0` semantics; never coerce strings/booleans.
- **Red-first tests:** `TestDeclarativeReadBack_NumericSemanticEquality`: int64/json.Number equality, >2^53 exactness, unequal 43 rejection, explicit 42/42.0 policy for expected and identity fields.

### PF-CF-B32 — Apply/publish and read-back consume one shared deadline

- **Severity / source / state:** BLOCKER; `ORCH-PF-B07`; newly exposed provider-readback gap.
- **Exact evidence:** ordinary shared context `internal/synctransport/orchestrator.go:237-273`; full overwrite `:471-485`; Arrow serial/pipeline `arrow_fast_path_controller.go:184-199`, `arrow_fast_path_pipeline.go:125-140`; child readback timeout `internal/app/issue_label_warehouse_transport.go:330-358`; unit contract `internal/synctransport/types.go:505-508`.
- **End-to-end path:** successful apply consumes most deadline -> read-back inherits residual/canceled parent -> verification fails -> checkpoint withheld -> effect replay/reconciliation.
- **Concrete flow:** 40ms successful apply under a 50ms unit deadline leaves insufficient time for a valid 20ms eventual-consistency read-back.
- **Six-surface impact:** ordinary/full/Arrow ETL and transport-backed reverse destinations blocked; standalone direct/binary deadlines separate.
- **Production risk:** systematic duplicate writes despite each phase fitting its intended bound.
- **Minimal fix:** cancel apply/publish context on return and create a fresh explicitly bounded read-back phase context; expose phase timing separately.
- **Red-first tests:** `TestTransport_ReadBackGetsIndependentUnitDeadline` across ordinary/full/serial/pipeline Arrow: 40ms apply +20ms readback succeeds with independent 50ms phases; >50ms apply fails.

### PF-CF-B33 — Runtime/catalog “clones” retain nested mutable state across executors and goroutines

- **Severity / source / state:** BLOCKER; `ORCH-PF-B08`; newly exposed isolation gap.
- **Atomic claims:** `cloneRuntimeConfig` shallow-copies nested catalog/discovery data; Arrow segment requests reuse original runtimes/binding across producer and consumer goroutines.
- **Exact evidence:** shallow clone `internal/synctransport/types.go:668-678`; mutable stream/catalog fields `internal/connectors/connectors.go:183-225`; Arrow request reuse `arrow_fast_path_controller.go:57-64,145-151`, `arrow_fast_path_pipeline.go:51-58,277-283`.
- **End-to-end path:** caller runtime/catalog -> shallow wrapper -> executor mutates nested schema/keys/failures/binding -> later source/readback/segment or concurrent goroutine observes mutation.
- **Concrete flow:** first pipeline segment changes `PrimaryKey` or raw schema; destination consumer and next segment share the change, with a race.
- **Six-surface impact:** ordinary and Arrow ETL plus reverse mappings blocked; standalone direct/binary calls do not use these wrappers.
- **Production risk:** unauthorized mapping/authorization drift, nondeterministic delivery, and data races.
- **Minimal fix:** deep-clone every stream slice/raw schema/discovery failure and create a fresh deeply cloned Arrow apply request per segment; define Arrow record ownership/release lifetime.
- **Red-first tests:** `TestTransportRequests_DefensiveCopyAllNestedState` mutates every nested field/config/secret/binding/segment across ordinary/full/serial/pipeline; originals and later calls unchanged; pipeline under `-race`.

### PF-CF-B34 — Ordinary ETL CLI drops persisted terminal runs and receipts

- **Severity / source / state:** BLOCKER; `ORCH-PF-B09`; claimed-closure residual of `CF-B09`.
- **Atomic claims:** ordinary `etl run` returns before emitting App run on execution error; runtime-ledger failure replaces an otherwise persisted App run; ambiguous App terminal-store outcome can return a zero run without exact reload.
- **Exact evidence:** early return `internal/cli/cli.go:750-758`; runtime-ledger suppression `:759-767`, `internal/cli/runtime_helpers.go:31-48`; correct approved pattern `internal/cli/etl_transport.go:495-519`; App creates run `internal/app/app.go:1381-1400`; terminal ambiguity `:1692-1700,3482-3535`.
- **End-to-end path:** CLI -> App persists running/terminal run -> execution or sidecar/save ambiguity -> App returns run+error or zero -> CLI emits generic error instead of one terminal run.
- **Concrete flow:** provider failure produces a persisted failed ETL run with receipt, but ordinary CLI stdout contains only a second generic error envelope; a runtime sidecar outage can hide a completed App run.
- **Six-surface impact:** ETL/CDC terminal contract blocked; reverse path is separately covered by `PF-CF-B35`; direct/binary unaffected.
- **Production risk:** operators lose durable run ID/status/receipt and cannot reconcile reported failure with state.
- **Minimal fix:** reconcile `state.CommitOutcome` by reloading exact run on may-have-committed; emit one nonempty terminal run then return an already-reported categorized error; runtime-sidecar failure annotates rather than replaces App truth.
- **Red-first tests:** `TestCLI_OrdinaryETLFailurePublishesOneTerminalRun` for failed, parked, CDC-failed, pre-rename no-commit, post-rename sync, and runtime recorder failure; one JSON object, exact persisted status/results, nonzero class, safe stderr.

### PF-CF-B35 — Reverse/direct-write CLI can publish a terminal run that was never persisted

- **Severity / source / state:** BLOCKER; `ORCH-PF-B10`; newly exposed finalization gap related to prior `CF-B09`.
- **Atomic claims:** definite pre-rename save failure is hidden behind provider error while fabricated run is returned; indeterminate post-rename outcome is not reloaded before publication.
- **Exact evidence:** terminal construction/update and wrong error order `internal/app/app.go:3149-3191`; CLI treats any nonempty run as persisted `internal/cli/cli.go:1831-1847,2179-2195`.
- **End-to-end path:** approved reverse/direct/binary write -> provider partial/error -> terminal save fails/ambiguous -> App returns in-memory run -> CLI publishes `ReverseRun` and suppresses persistence error.
- **Concrete flow:** provider returns partial response, pre-rename store write fails, stdout claims an inspectable run ID that does not exist.
- **Six-surface impact:** reverse ETL, generated direct writes, and binary upload blocked; ETL separate; reads/download unaffected.
- **Production risk:** false durable terminal claim and hidden plan/run uncertainty.
- **Minimal fix:** handle persistence outcome first; definite no-commit returns zero run plus joined errors and recovery state; may-have-committed reloads exact plan/run and publishes only durable terminal transition.
- **Red-first tests:** `TestReverseFinalization_DoesNotPublishUnpersistedRun`: provider partial/error + pre-rename failure yields one Error/no stored run/uncertain plan; post-rename sync reloads and emits exactly durable run if present.

### PF-CF-B36 — Durable authentication fencing does not stop already-admitted work across processes

- **Severity / source / state:** BLOCKER; `ORCH-PF-B11`; newly exposed cross-process/runtime boundary gap.
- **Atomic claims:** multi-request operations check durable health once and process-local membership cannot cancel another process; may-have-committed fence/repair CAS errors return before local old-epoch cancellation.
- **Exact evidence:** process-local ownership `internal/coordination/auth_cohort.go:121-129`; one-time execute check `:153-171`; per-send admission contract `:294-325`; engines wrap whole operations `internal/connectors/engine/connector.go:166-190,222-239,289-296,344-445`; PostgreSQL `connection.go:457`; ambiguous swap `internal/coordination/durable_store.go:97-117`, `internal/state/store.go:276-315`; early returns `auth_cohort.go:364-377,410-423,454-467`.
- **End-to-end path:** process A admits multi-page/multi-send work -> process B persists invalid fence/repair epoch -> A never rechecks before next send -> old credential continues; ambiguous committed fence may not cancel even same-process members.
- **Concrete flow:** A completes page one, B fences credential, A still issues page two with the invalid credential.
- **Six-surface impact:** ETL, reverse ETL, direct read/write, binary download, and binary upload all blocked at request/query/statement boundaries.
- **Production risk:** revoked/invalid credentials continue provider access or mutation after durable fence.
- **Minimal fix:** install epoch-bound admission at every HTTP page/retry/upload/download and database query/statement; typed CAS commit outcomes reload exact health and cancel old local epochs whenever persisted state advanced.
- **Red-first tests:** `TestAuthFence_StopsNextSendAcrossCoordinators` with two coordinators/store, page/send one then fence/repair, zero later sends; repeat post-rename sync ambiguity, local admissions, all six surfaces, PostgreSQL, epoch rollover.

### PF-CF-B37 — CDC recovery does not bind the restored receipt to exact warehouse artifacts

- **Severity / source / state:** BLOCKER; `ORCH-PF-B13`; newly exposed CDC data-loss gap.
- **Exact evidence:** source stage retains transaction key/count/digest `internal/connectors/database/transaction_stage.go:221-270`; recovery reduces to ID/sink/time `internal/connectors/native/postgres/cdc_v2.go:271-285`; App receipt ID uses connection+transaction only `internal/app/change_capture_dispatch.go:377-379`; restore checks only matching ID/sink and regular WAL/table paths `:256-287`.
- **End-to-end path:** durable PostgreSQL stage receipt -> crash/restart -> shared WAL/table paths merely exist -> App manufactures acknowledgement -> checkpoint/LSN advances -> source stage retires.
- **Concrete flow:** replace WAL and table with valid unrelated regular files before restart; restore accepts them as the staged transaction and acknowledges its LSN.
- **Six-surface impact:** ETL CDC blocked/data loss; other five surfaces unaffected.
- **Production risk:** permanent loss of unmaterialized transaction on recovery.
- **Minimal fix:** persist an atomic connection/stream/generation-owned warehouse receipt manifest with transaction key, record count, content digest, and WAL/final artifact digests; verify exact manifest/artifacts or enter typed reconciliation without LSN advance.
- **Red-first tests:** `TestCDCRecovery_ReceiptBindsExactWarehouseArtifacts`: replaced/truncated/stale-generation files prevent checkpoint; untouched exact artifacts restore without duplicate receive/write; typed recovery error.

### PF-CF-B38 — Post-checkpoint bookkeeping failures relabel delivered work as failed

- **Severity / source / state:** BLOCKER; `ORCH-PF-B14`; claimed-closure residual of `CF-B06`.
- **Atomic claims:** stage-receipt retirement can fail after checkpoint; declarative/managed approval markers can fail after committed checkpoint and output.
- **Exact evidence:** checkpoint before retirement `internal/synctransport/orchestrator.go:277-290,492-502,525-537`; post-commit marker work `internal/app/transport_dispatch.go:357-390`; marker stores `declarative_typed_destination_approval.go:346-368`, `postgres_transport_approval.go:283-308`; any error terminalized failed `internal/app/etl_mode_dispatch.go:66-75`, `internal/app/app.go:1612-1629,1710-1760`.
- **End-to-end path:** provider apply/read-back -> checkpoint CAS succeeds -> cleanup/marker save fails -> dispatcher returns error -> terminalizer writes failed although delivery/checkpoint are durable.
- **Concrete flow:** stage retirement fails after a successful checkpoint; operator sees failed, retry sees checkpoint and produces empty success, and neither run describes delivery.
- **Six-surface impact:** all staged/approved ETL and transport-backed reverse destinations blocked; standalone direct finalization is `PF-CF-B35`; read/download unaffected.
- **Production risk:** misleading delivery truth and unsafe operator replay/reconciliation decisions.
- **Minimal fix:** persist a distinct `delivered_reconciliation_required` phase after checkpoint; never downgrade/replay provider work; restart idempotently completes receipt retirement/markers from committed evidence, atomically with checkpoint where possible.
- **Red-first tests:** `TestTransport_PostCheckpointBookkeepingFailureRemainsDelivered` injects retire, declarative marker, and managed-target marker failures; one provider mutation, retained checkpoint/output, delivered-reconciliation status, restart repair with zero replay.

## Warnings

### PF-CF-W01 — `surface-sync` preserves stale provider-owned parameter contracts

- **Severity / source / state:** WARNING; `PFM-W01`; newly exposed robustness residual of source-projection closure.
- **Exact evidence:** add-only behavior `cmd/connectorgen/surfacesync.go:732-800`, especially existing-flag skip `:749-775`; enshrined by `cmd/connectorgen/paramsimport_test.go:371-406`; direct reads admit command-only query fields at `internal/connectors/engine/direct_read.go:656-689,718-727`.
- **End-to-end path:** provider parameter changes/deletes -> params import updates operation -> existing CLI flag survives unchanged -> runtime authorizes stale type/enum/location/requiredness/bounds.
- **Concrete flow:** provider changes an enum or deletes a parameter; `surface-sync --check` remains green with the old flag. Frozen GitHub audit found no current mismatch, so severity remains WARNING.
- **Six-surface impact:** reverse ETL/direct read/direct write may drift; no proved current ETL/binary mismatch.
- **Production risk:** future source changes silently create invalid or over-broad command input.
- **Minimal fix:** synchronize all provider-owned semantic properties exactly and remove deleted flags; preserve only authored prose or an explicitly reviewed source-bound narrowing.
- **Red-first tests:** change enum/type/location/requiredness/repeatability/max bytes and require exact update; delete a parameter and require removal or explicit exception; prose survives.

### PF-CF-W02 — Website and generated skills lose safety-critical flag metadata

- **Severity / source / state:** WARNING; `PFM-W02`; mixed residual of `CF-W01`/`CF-W07` and newly exposed metadata loss.
- **Atomic claims:** website mapper/types omit env-only, byte/item bounds, and repeatability; guide/skill renderer omits env-only and limits.
- **Exact evidence:** website mapping `website/scripts/lib/cli-surface.mjs:100-118`; type `website/lib/connectors.types.ts:37-47`; GitHub label max 8192 at CLI `:2364-2386` but missing website `connectors.generated.json:71645-71661`; env-only CLI `:43022-43035` missing website `:123341-123358`; skill renderer `internal/connectors/guide.go:304-317`; generated GitHub skill omissions around `:2689,3968`.
- **End-to-end path:** complete CLI flag -> lossy website model or guide renderer -> generated catalog/skill tells user inline secret input has no constraint.
- **Concrete flow:** create-migration-source appears safe to pass inline although source declares `env_only=true`; label delete hides its 8192-byte maximum.
- **Six-surface impact:** guidance for ETL, reverse ETL, direct read/write, and binary upload can be unsafe; binary download uses separate injected flags.
- **Production risk:** users expose secrets in argv/history or send rejected oversized input.
- **Minimal fix:** make website/skill models lossless for every semantic property (`env_only`, byte/numeric/item bounds, repeatable, allow-empty, format, required, values, maps-to) and render env-only as an explicit usage constraint.
- **Red-first tests:** semantic walk comparing every source CLI flag to website JSON and skill metadata; pin label `max_bytes=8192` and migration source `env_only=true`.

### PF-CF-W03 — Declaration-owned idempotency headers are previewed then silently stripped

- **Severity / source / state:** WARNING; `PFR-WR-01`; newly exposed latent request-drift gap.
- **Exact evidence:** header protection excludes `Idempotency-Key`/`X-Idempotency-Key` at `internal/connectors/operation_headers.go:22-56`; engine accepts declared bounded headers `internal/connectors/engine/operation_headers.go:46-108,201-309`; transport deletes them under retry disable `internal/connectors/connsdk/http.go:297-303`. No installed current declaration was proved.
- **End-to-end path:** future declaration -> CLI -> prepared/digested header -> execution disables retries -> transport strips header -> provider receives material different from preview.
- **Concrete flow:** a declared `Idempotency-Key` appears in approved request yet never reaches provider.
- **Six-surface impact:** latent ETL/reverse/direct write/binary upload drift; reads/download not current.
- **Production risk:** provider deduplication disappears while approval claims it is present.
- **Minimal fix:** reject these names as runtime-owned until safely supported, or preserve the exact preview-bound value while disabling replay independently; never silently mutate.
- **Red-first tests:** declaration of both names must fail pre-I/O or exact value must reach server once with redirects/retries disabled; aliases/duplicates remain rejected.

### PF-CF-W04 — Structured REST witness generation rejects supported `minLength` schemas

- **Severity / source / state:** WARNING; `PFR-WR-02`; newly exposed latent reachability gap.
- **Exact evidence:** witness gate `internal/connectors/engine/structured_rest_body.go:682-768`; string candidate ignores `minLength` at `:1785-1814`; shared schema supports it at `internal/connectors/engine/schema.go:315-339,434-447`. No installed generated occurrence was proved.
- **End-to-end path:** valid structured body schema -> minimum-witness compilation -> empty candidate violates minLength -> operation rejected before command execution.
- **Concrete flow:** required string with `minLength:1` and no pattern is locally declared unreachable although runtime validation accepts real non-empty values.
- **Six-surface impact:** latent reverse/direct structured writes, typed destinations, POST reads, and multipart metadata; binary download unaffected.
- **Production risk:** future valid provider operations require unsafe schema broadening/workarounds.
- **Minimal fix:** synthesize a bounded Unicode-code-point witness satisfying min/max, pattern, and format, then validate encoded JSON byte cap.
- **Red-first tests:** required minLength string plus nested object/array; min==max, Unicode, URI+minimum, pattern+minimum, impossible min>max; preview and wire exact.

### PF-CF-W05 — Multipart symlink-boundary security test races and can false-pass

- **Severity / source / state:** WARNING; `PFR-WR-03`; newly exposed test-reliability gap.
- **Exact evidence:** handler writes shared boolean at `internal/connectors/connsdk/multipart_bounds_test.go:24-45`; test reads it after requester may return but before handler completion at `:105-107`; focused `go test -race` reproduced the conflict.
- **End-to-end path:** request returns early -> test asserts unsynchronized false -> handler later observes leaked bytes and writes true -> security regression falsely appears green.
- **Concrete flow:** outside-root file can reach server after assertion already passed.
- **Six-surface impact:** direct-write/binary-upload security assurance degraded; reverse/ETL multipart destination confidence affected; reads/download unaffected.
- **Production risk:** a real filesystem-boundary regression can escape CI.
- **Minimal fix:** send handler observation through channel/atomic/lock and await handler/server completion before assertion; keep cancellation deterministic.
- **Red-first tests:** repeatedly run `go test -race -run TestRequesterDoMultipartRefusesEscapingSymlinkSwappedAfterValidation ./internal/connectors/connsdk`; prove handler completes and never sees sentinel bytes.

### PF-CF-W06 — File-backed rate-parking mutations ignore indeterminate commit outcomes

- **Severity / source / state:** WARNING; `ORCH-PF-W01`; newly exposed durability gap related to prior `CF-W04`.
- **Exact evidence:** mutation APIs return callback flags plus raw store error `internal/coordination/durable_store.go:194-253,269-306,317-373,401-437`; post-rename errors can mean committed `internal/state/store.go:276-315`; coordinator live-state divergence in `internal/coordination/rate_parking.go:541-578,650-673,800-816`.
- **End-to-end path:** durable transition renames successfully -> directory-sync/unlock error -> caller treats as no commit -> timers/maps diverge from disk until restart.
- **Concrete flow:** Park is durably created but no live timer is installed, so scope admission blocks indefinitely.
- **Six-surface impact:** shared rate parking can affect all six surfaces; primary current impact is ETL/reverse robustness.
- **Production risk:** wedged scopes, duplicate retries, or churn despite correct durable record.
- **Minimal fix:** typed commit outcomes for every mutation and exact reload/reconciliation on may-have-committed before timer/map/retry decisions.
- **Red-first tests:** `TestRateParking_IndeterminateMutationReconcilesLiveState` for Create, Rearm, Claim, MarkResumeCompleted, Complete, Delete; compare live coordinator to reopened store and exact timer cardinality.

### PF-CF-W07 — Expired/revoked authorization causes unbounded parked-run rearm

- **Severity / source / state:** WARNING; `ORCH-PF-W02`; newly exposed recovery gap.
- **Exact evidence:** 24h plan expiry `internal/app/write_approval.go:21-24`; authorization rejection `internal/app/authorization.go:248-282`, `declarative_typed_destination_approval.go:194-214`; provider reset may exceed it `internal/connectors/connsdk/http.go:1402-1416`; rearm retains original plan `internal/app/durable_coordination.go:178-279,338-347`; coordinator retries every resume error `internal/coordination/rate_parking.go:780-790`; no production Cancel caller.
- **End-to-end path:** provider parks beyond approval lifetime -> timer resumes old plan -> auth fails pre-I/O -> failed attempt remains retryable -> rearm repeats forever.
- **Concrete flow:** `Retry-After` lands after 24h plan expiry; failed attempts accumulate and provider scope never clears.
- **Six-surface impact:** ETL/reverse resume wedges; same parked scope can block direct read/write/download/upload admission.
- **Production risk:** permanent availability loss and unbounded failed-attempt records.
- **Minimal fix:** classify expiry/revocation as durable terminal `needs_reauthorization`, stop retry, expose exact-scope reapproval takeover or safe cancel/abandon.
- **Red-first tests:** `TestRateParking_ExpiredAuthorizationStopsAndCanRecover`: expiry and revocation each emit one needs-auth event, zero provider sends/retry growth, then reapproval or cancellation unblocks exact scope.

### PF-CF-W08 — Route selection hides declared-route preflight failures as declaration absence

- **Severity / source / state:** WARNING; `ORCH-PF-W03`; claimed-closure residual of `CF-W06`.
- **Exact evidence:** semantic filter returns declaration absent for unregistered/wrong-marker destination `internal/app/transport_dispatch.go:52-75`; registry has typed errors `internal/synctransport/registry.go:145-225`; dispatcher replaces with generic error `internal/app/etl_mode_dispatch.go:51-95`.
- **End-to-end path:** both endpoints declare transport -> early filter suppresses actual registry preflight -> generic no-route error instead of executor/binding/conformance reason.
- **Concrete flow:** destination executor is unregistered; user is told no transport was declared.
- **Six-surface impact:** ETL and reverse destination operability; direct/binary surfaces unaffected.
- **Production risk:** fail-closed behavior remains, but operators cannot repair configuration and may choose the wrong fallback.
- **Minimal fix:** once both declarations exist, always run registry preflight and return typed error; distinguish intentional `declared_but_not_selected` from true absence.
- **Red-first tests:** `TestETLRouteSelection_PreservesDeclaredPreflightError` for unregistered executor, wrong marker, missing binding/action, conformance rejection, zero I/O; retain two-sided-absent legacy case.

## One-Wave Remediation Plan

This is one merge-blocking wave. The groups below are safe atomic commits, ordered by dependency. A group may land on the remediation branch when its own red tests turn green, but the lane is **not mergeable** until the final once-per-wave gates pass and every BLOCKER is closed.

| Order | Atomic commit group | Findings | Dependency and exact scope |
|---:|---|---|---|
| 0 | Red-contract freeze | all | Add every named behavioral test first, prove it fails for the intended reason, and record the frozen failing command. No production or generated changes. |
| 1 | Source authority and complete REST/POST reachability | `B01-B03`, `B09`, `B12`, `W01` | Strict authenticated source/projection graph, fail-closed gaps/refs, nested Google Ads schemas, exact add/change/delete sync. Foundation for all generated/runtime contracts. |
| 2 | GraphQL and documentation projection | `B04-B08`, `W02` | Exact direction/page state, source-owned secret locations, nested mutation identity, Int32 domain, and lossless website/skills. Depends on group 1 source authority. |
| 3 | Complete receipt + secret-only public projection | `B13-B14`, `B17`, `B19`, `B24` | One immutable provider receipt and one value-only public masker across REST/GraphQL/status/native SQS/App/CLI. Depends on groups 1-2 for declared secret/result paths. |
| 4 | Sealed request and HTTP authority | `B15-B16`, `B18`, `B21`, `B23`, `B25`, `W03-W04` | Prepared hook sequences, exact bytes/config/file identity, response-retaining retry/redirect, exact numeric CLI input, cursor admission, SQS redirect policy, header/witness correctness. Depends on receipt types from group 3. |
| 5 | Crash/race-safe filesystem boundary | `B22`, `W05` | Atomic no-replace download publication, owned cleanup, directory fsync, synchronized multipart test. Independent of orchestration, but uses group 3 receipt output when errors are reported. |
| 6 | Durable coordination, auth, ownership, and cloning primitives | `B20`, `B26`, `B33`, `B36`, `W06-W07` | Typed commit outcomes, stream/auth fences at actual effect/send boundaries, deep request/runtime copies, and terminal reauthorization state. Must precede destination and terminal state-machine repair. |
| 7 | Source exhaustion and destination proof | `B27-B32` | Typed page-budget continuation, independent idempotency evidence, private receipt-targeted readback, precision-preserving clone/compare, independent phase deadlines. Depends on groups 1, 3, and 6. |
| 8 | Terminal durability and recovery | `B34-B38`, `W08` | Persisted-only CLI envelopes, may-have-committed reload, exact CDC artifact manifest, delivered-reconciliation state, exact route diagnostics. Depends on receipt, commit-outcome, ownership, and readback groups. |
| 9 | Evidence/certification regeneration | `B10-B11` plus artifact closure for `B01` | Bind evidence and certification subject fingerprints to the final clean SHA; regenerate all source, CLI, docs, website, skills, ledgers, matrices, candidates, and sweep artifacts once. Depends on every production group. |

### Gates run for every atomic group

1. Run the named red-first tests before production edits and retain the expected failure reason.
2. After the change, run `gofmt` on changed Go files, focused `go test -timeout 20m` for every affected package, and focused `go vet`.
3. Run `go test -race -timeout 20m` for any group touching goroutines, shared state, transport clients, filesystems, coordination, Arrow, or tests with handlers.
4. Generator groups run their package and relevant Node exact-output tests plus the affected generator `--check`; generated bytes must be committed with their owning generator change.
5. Confirm no unrelated tracked or untracked source/test/generated drift before the next group.

### Gates run once at the final unchanged remediation SHA

1. Verify exact HEAD/base and a clean worktree; run `git diff --check`.
2. Run all focused red-first tests together, then the complete package suite with repository-required `-timeout 20m`; run `go vet ./...` and the required race cohorts for connsdk, engine, commandrunner, synctransport, coordination, App, and native SQS/PostgreSQL paths.
3. Run `connectorgen validate`, `surface-sync --check`, source/params import checks, both GitHub parity checks, `github-parity-artifacts-check`, certification matrix/candidates/sweep checks, website/skills/docs exact-generation tests, and `agentcontractgen check`.
4. Exercise installed commands, not metadata alone: representative ETL, reverse plan/preview/approval/execute/readback, REST and GraphQL direct reads/writes, binary download/upload, native SQS, failure envelopes, pagination continuation, and CDC restart recovery.
5. Regenerate Foundation evidence and certification at that exact SHA/base/subject fingerprint, rerun their gates without source movement, and record ordinary provider/occurrence-ID preservation plus secret-only masking proof.

## Merge Verdict

**BLOCKED.** Thirty-eight blockers remain. The frozen lane must not be called clean, shippable, or mergeable until all groups and the once-per-wave gate set pass at one unchanged, evidence-bound SHA.

---

_Reviewed: 2026-08-21T03:27:40Z_

_Reviewer: gsd-code-reviewer convergence pass_

_Frozen diff: e62ae21d428f0d27225f9bff564dc2cd797f6b65..8a8a866ff6d5282c28bda12acceed8a624218f01_
