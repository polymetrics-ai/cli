# TDD ledger — source-lock operation import

## Task Delivery Header

- Issue: Refs #4306 — feat(connectorgen): import hash-locked provider operation contracts
- Base branch: main (verified from the GitHub pull-request API for #4312)
- Merges into: main
- Delivery: CodeQL allocation-overflow findings are fixed in the source-import foundation, focused local regression checks pass, and the worktree is returned to the outer CI executor without pushing or controlling its pipeline.
- Working branch: fm/cli-source-lock-import-r1
- Task: Remove the overflow-prone allocation-capacity sum reported by CodeQL in `cmd/connectorgen/sourceimport.go` without weakening the source-import contract or accepting an oversized document.
- Verification: Run the existing source-import reference-sibling regression, the `cmd/connectorgen` package tests, `go vet ./cmd/connectorgen`, and `git diff --check`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A reference-sibling overlay does not combine two map lengths to preallocate its result | live | The implementation preallocates only from an already representable single map length; the outer CI executor owns the subsequent CodeQL re-analysis. |
| Valid bounded source-import documents retain their canonical descriptor behavior | live | Existing source-import golden tests assert the exact canonical descriptor output for the synthetic Alpha and Beta fixtures. |
| The fix is scoped to the reported CodeQL root cause | live | `git diff --check` and the focused package tests exercise only the allocation guard and its regression coverage. |

## Cycle 16 — CodeQL allocation-size closure

- **Red:** CodeQL reported two high-severity `size computation for allocation may overflow` alerts at `cmd/connectorgen/sourceimport.go:4375`; the preallocation size directly adds two map lengths.
- **Green:** the overlay preallocates from `len(target)` only, then lets normal map growth accommodate reference siblings; it no longer evaluates the overflow-prone aggregate capacity.
- **Refactor constraint:** preserve the existing importer limits and valid descriptor ordering; do not broaden source inputs or alter connector definitions.
- **Verification:** `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceImportPreservesLiteralReferenceFieldsAndReferenceSiblings$' -count=1` passed; `go test -timeout 20m ./cmd/connectorgen` passed; `go vet ./cmd/connectorgen` and `git diff --check` passed.
- **CI-phase fallback:** this is an assigned active no-mistakes CI phase. The outer executor retains pipeline, PR, push, and lifecycle control; required `golang-how-to`, security, safety, error-handling, lint, testing, and no-mistakes guidance is applied inline.

## Cycle 1 — closed retrieval and integrity

- **Red:** `go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImport'` initially failed to compile because no `sourceImportLock`, fetch boundary, descriptor, limits, or importer existed.
- **Green:** two checked-in source locks now fetch only their exact fixed fixture URL; mismatched bytes/digest and artifact-size violations return a source-lock-refresh/bound error before command output is created. Connector lock lookup is fixed to `defs/<connector>/sources/<connector>-operation-source-lock.json`.
- **Refactor:** one internal fetch interface keeps fixture retrieval explicit while the production fetcher permits only the lock's HTTPS URL, follows no redirects, sends no credentials, and bounds its read.
- **Lint repair:** the first full `make verify` run reached lint after all tests/smoke checks and reported an unchecked HTTP response-body close. The fetcher now returns read/close failures with context; it does not discard cleanup errors.

## Cycle 2 — canonical provider descriptors

- **Red:** alpha/beta fixture tests required a non-existent descriptor contract and deterministic canonical serializer.
- **Green:** OpenAPI 3 YAML and Swagger 2 JSON fixtures yield byte-stable sorted descriptors with distinct path/query/header/body sections, source URL/digest/location, pagination/auth/byte metadata, empty provider-operation ID plus derived source ID, and JSON/binary/status/text classification. Resolved response declarations retain both `200` and `403` shapes plus `rare_paid_result` and `access_token` fields.
- **Refactor:** response declarations are retained wholesale; classification is a separate additive field so later runtime redaction cannot become an import-time field filter.

## Cycle 3 — bounded reference and request contracts

- **Red:** isolated fixture tests required named failure modes for references, request contracts, identity, callbacks, source drift, and bounds.
- **Green:** external, unresolved, cyclic, over-depth, and over-count references; duplicate identities; unbounded/dynamic request bodies; unsupported XML encoding; oversized artifact/schema; and operation-count excess all fail before descriptor emission; source-locked callback and webhook routes are retained as explicitly merge-blocked inbound events.
- **Refactor:** the local JSON Pointer resolver owns bounded ref traversal and canonical map copying, avoiding a second schema walker with different safety rules.

## Cycle 4 — closed adoption command and documentation

- **Red:** `source-import` was an unknown command and migration conventions had no adoption contract.
- **Green:** help documents connector/defs/output/check only; it exposes none of `--url`, `--method`, `--path`, `--header`, `--body`, or `--credential`. Command/check-mode tests prove owned-lock resolution and descriptor drift detection; migration guidance records lock refresh and preservation constraints.
- **Refactor:** standard and connector-qualified `--help` both render the same closed command contract. Inline GSD code review found no remaining correctness, safety, or documentation findings.

## Cycle 5 — review hardening parity

- **Red:** review cases exposed duplicate JSON/YAML object members, grammar-blind `$ref` resolution, silently dropped inbound route metadata, lossy parameter serialization, dynamic references, semantic bound errors, reference amplification, and mixed success/error media collapse.
- **Green:** the parser now reports duplicate-member JSON Pointers; scoped OpenAPI resolvers preserve literal `$ref` fields and legal 3.1 siblings; inbound events, route servers, extensions, and parameter wire metadata are canonical merge-blocked evidence; dynamic references fail with a named foundation; finite enums and actual finite bounds are evaluated semantically; aggregate descriptor bytes are bounded; response status/media variants remain independent.
- **Refactor:** document-level source-import results keep operation descriptors compatible while making inbound events and extensions explicit, and a shared compact-size budget bounds every retained result before append.

## Cycle 6 — grammar-complete source form hardening

- **Red:** follow-up review cases exposed unresolved Link/Example references, lossy content-only parameter contracts, unvalidated JSON Schema `prefixItems`, version-form collapse, missing Swagger route binding, cross-kind source-ID collisions, non-string YAML keys, and pre-budget schema reference amplification.
- **Green:** focused source-import coverage now resolves and type-checks Link/Example references, rejects malformed parameter/schema/version/YAML forms, preserves exact source form plus Swagger root/operation route evidence, rejects cross-kind identities, and returns typed schema-expansion limits before retaining amplified children.
- **Refactor:** the importer now carries one exact document-form model through parsing, reference rules, provenance, parameter handling, and route binding; the resolver owns grammar-specific target checks and incremental per-object/document expansion charging.

## Cycle 7 — semantic enum and grammar-position closure

- **Red:** review follow-up found finite object/array enums rejected despite bounding the complete input space, and found non-OpenAPI fields plus Encoding Header references outside the grammar-scoped resolver boundary.
- **Green:** finite composite enums now retain schema-shape validation while relaxing only inherited cardinality bounds; OpenAPI-only reference positions are form-scoped, and Encoding Header references fail closed before descriptor creation.
- **Refactor:** bounded-schema validation now carries finite-enum context through child schemas, while resolver traversal follows the parsed document form rather than field names alone.

## Cycle 8 — exact schema grammar and expansion closure

- **Red:** review found non-string operation IDs silently converted to derived IDs, omitted OpenAPI 3.1 schema-bearing positions, descendant component references rejected by category shorthand, OpenAPI response-field loss, float-rounded contradictory bounds, and inline schema expansion bypassing depth/node budgets.
- **Green:** focused source-import cases now cover typed operation IDs, literal OpenAPI response fields, nested component-schema pointers, `contentSchema`/`unevaluatedItems`, exact large-number bounds, OpenAPI 3.1 finite `const`/union/closed-tuple forms, and inline depth/node limits.
- **Refactor:** a version-scoped schema grammar owns child traversal and pre-retention limits, while exact rationals own numeric-bound comparisons.

## Cycle 9 — whole-document grammar and allocation closure

- **Red:** review cases found wrong-kind nested pointers, unused invalid components, rounded YAML numerics, late deep-schema materialization, cross-version schema keywords, and repeated response expansion.
- **Green:** version-scoped grammar indexing now preflights every supported reference and schema position before descriptors; YAML numbers retain exact JSON-number lexemes; structural schema limits precede cloning; source-form keyword shapes are enforced; and response expansion reserves aggregate bytes before append.
- **Refactor:** one grammar index now supplies both expected reference kinds and whole-document preflight, keeping literal `$ref` values outside reference positions untouched.

## Cycle 10 — extension-context and inbound expansion closure

- **Red:** review cases exposed legal x-prefixed reusable-component names being skipped as extensions, non-request Media Type encodings surviving resolution, inbound response amplification, and unbounded grammar-index position staging.
- **Green:** focused source-import cases cover x-prefixed reusable component references and unused external refs, non-request encoding rejection, webhook/callback response expansion limits, and grammar-position caps before component ordering.
- **Refactor:** extension handling is now explicit per grammar owner, while response and grammar-index accounting are shared pre-retention boundaries.

## Cycle 11 — response-child, target, and discovery closure

- **Red:** review follow-up found nested referenced response children escaping aggregate reservation, operation and inbound-event limits checked after construction, unvalidated Link operation targets, dropped request Media Type encoding metadata, and extension keys bypassing grammar-index capacity before ordering.
- **Green:** source-import coverage now reserves resolved response-child expansion before cloning, reserves outbound and inbound count slots at discovery, validates unique in-artifact Link targets, retains request encoding with a merge-blocking foundation gap, and charges extension positions before sorting.
- **Refactor:** the response estimator follows reference-bearing response grammar children without materializing them; one shared import count budget and grammar-position ledger bound all discovery paths.
- **Review fallback:** this was an assigned active no-mistakes review phase, so the outer executor retains pipeline control; required `golang-how-to`, security, safety, error-handling, lint, and testing guidance was applied inline.

## Cycle 12 — local-link, form-media, and inbound-retention closure

- **Red:** review findings r50–r53 identified percent-encoded local Link targets being misresolved, form-media defaults being rejected or non-form encodings accepted, and request/inbound expansion being retained before aggregate limits applied.
- **Green:** local Link fragments now decode exactly once before JSON Pointer resolution and require one reachable operation occurrence; form request media retains implicit or explicit encoding semantics behind a named merge block; request media and inbound event expansions reserve resolved children before cloning.
- **Refactor:** reachability is indexed separately from grammar positions, while shared retained-expansion budgets protect request media and inbound declarations without changing response preservation.
- **Review fallback:** this assigned no-mistakes review phase applied `golang-how-to`, security, safety, error-handling, lint, and testing guidance inline; outer-pipeline commands remain intentionally uninvoked.

## Cycle 13 — non-response reference retention closure

- **Red:** review found that Path Item, Example, Header, Link, and Security Scheme aliases could clone large resolved targets repeatedly before any aggregate budget ran.
- **Green:** the shared reference resolver reserves fully expanded non-response reference targets before cloning; focused aliases and ordinary reusable Path Item tests prove deterministic early rejection.
- **Refactor:** existing schema, response, request-media, and inbound budgets remain their respective owners, while a single reference-target budget covers the previously unowned resolver kinds.
- **Review fallback:** this assigned no-mistakes review phase applied `golang-how-to`, security, safety, error-handling, lint, and testing guidance inline; outer-pipeline commands remain intentionally uninvoked.

## Cycle 14 — callback and canonical-fragment closure

- **Red:** review found Callback aliases were not charged at the shared pre-clone boundary, and ordinary local `$ref` fragments bypassed URI-fragment normalization.
- **Green:** callback chains now reserve every fully expanded hop before recursive merge cloning; direct references normalize once before cycle, pointer, and grammar-index handling, with encoded, malformed, double-encoded, and canonical-cycle cases covered.
- **Refactor:** one canonical local-reference boundary now serves generic `$ref` and Link operation targets while reachability reuses its canonical stack identity.
- **Review fallback:** this assigned no-mistakes review phase applied `golang-how-to`, security, safety, error-handling, lint, and testing guidance inline; outer-pipeline commands remain intentionally uninvoked.

## Cycle 15 — literal-percent reference identity

- **Red:** the canonical-fragment guard treated a once-decoded literal `%HH` segment as a second decode attempt, rejecting legal nested-schema references.
- **Green:** local fragments decode once into canonical pointer identities; literal-percent targets resolve through grammar indexing while malformed escapes and unresolved second-decode attempts still fail, and canonical aliases retain cycle detection.
- **Refactor:** the shared reference normalizer is the only URI-percent decoder; all downstream pointer, index, and cycle paths consume its canonical result unchanged.
- **Review fallback:** this assigned no-mistakes review phase applied `golang-how-to`, security, safety, error-handling, lint, and testing guidance inline; outer-pipeline commands remain intentionally uninvoked.
