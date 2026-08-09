# PLAN: 426 connector artifact sweep

## Authority and execution mode

Scope authority is `/Users/karthiksivadas/karthik-agent-workspace/data/cli-mass-artifact-materialize-r1/CAPTAIN-ORDER-fast-426-terra-20260809.md`. It supersedes the earlier pilot/defer and broad generator-review orders: the current target is a static, complete 426-connector JSON surface sweep, not a further generic-generator initiative.

This concise record is the required inline/manual GSD fallback. `scripts/gsd doctor`, the sources for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`, and `go run ./cmd/agentcontractgen check` passed. The phase identifier is intentionally non-numeric and uses `PLAN.md`, so the generated role workflows cannot safely bind their numeric phase filenames; the one-worker captain lane therefore records its Red/Green evidence here and in `TDD-LEDGER.md` instead of spawning incompatible roles. The task is executed by one worker without spawned GSD roles.

Every reconciled target retains one provider-cited `rate_limits.json`. The existing engine schema and loader remain the foundation; this slice adds the missing production embed, deterministic declaration producer, target-only schema/load accounting, and ledger totals without adding runtime pacing behavior.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-documentation`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

## Slices

1. Build a deterministic 426-target ledger by reconciling `main`, this recovered branch, the four pilots, and the seven-connector extraction branch. Preserve exact source/provenance and one accounting state per target; primary-source drops enter the explicit retry queue rather than becoming terminal.
2. Recover only intended seven-connector bundle artifacts absent from this worktree; do not import its unrelated shared foundation work.
3. Materialize the 221 recovered retry targets from official machine sources or bounded official reference traversals. Retain complete JSON even when execution is unavailable; mark that state `foundation_pending`. A cache-only reference pass may not fetch an uncached secondary source; it preserves the retry route instead.
4. Materialize exactly one `rate_limits.json` for every 426 target. Declare only provider-cited enforceable policies; otherwise retain a precise official-source publication gap as `unknown`. Embed and schema-load the files, and include declared/unknown/not-applicable/file totals in the deterministic ledger.
5. Run only parse/schema/load/discovery and reachable-command checks where the foundation exists. Keep exact ledger counts current after each deterministic retry slice. Run the one final no-mistakes pass only after every target has materialized JSON or a documented, fully exhausted official-source block.

## Boundaries

No credentialed or live-provider checks, certification, per-connector review, runtime-foundation work, speculative generator redesign, or changes outside connector bundles unless one concrete extraction defect requires the smallest evidence-backed generator fix. The rate-limit slice additionally permits the first `defs.FS` embed update plus task-scoped producer/reconciler/test code needed to prove the 426-file contract; it does not change engine pacing semantics.

## Pipeliner dynamic OpenAPI server base

- Red: B054 fetched Pipeliner's official OpenAPI from the provider-linked HTTPS S3 object endpoint but rejected the documented request base because it includes `/spaces/{space_id}`. The source operation inventory could not be extracted or classified.
- Green: `TestParseBatchOpenAPIArtifactPrefixesServerBasePathWithPathVariable` admits only well-formed OpenAPI-style variable names in an otherwise ordinary request path. `TestMaterializeAPISurfaceUsesExistingDynamicServerBaseSuffix` preserves the exact existing connector-relative `{space_id}` binding rather than recasting it as an artifact-absent discrepancy. The focused parser/materializer regressions pass, and B055 validates and materializes Pipeliner's complete cited 1,510-operation artifact.

## Cached-reference static materialization

- Red: B043's 37 cached primary sources invoked the generic reference resolver, which opened hundreds of linked-document connections before producing a report. That defeats the captain's bounded, source-first retry pass and prevents a direct official root from being evaluated independently of unrelated provider navigation.
- Green: an explicit `--cached-references-only` materialization mode admits only already-cached secondary official documents. A cache miss fails the candidate immediately and records its normal retry outcome; it never performs a network request. The root artifact remains provider-cited and hash-proven, so direct OpenAPI/Swagger/Postman or complete HTML roots can materialize without a recursive crawl.

## Feed-link extraction repair

- Red: Countercyclical's provider-owned HTML root includes its own `/rss.xml` alternate feed. Its URL shares the `endpoint` marker used to admit possible reference documents, so the generic traversal selected the non-operation feed and stopped on its provider 404 before it could finish cached API pages.
- Green: feed suffixes are treated like other static non-operation assets. `TestParseBatchHTMLReferenceSkipsFeedCandidates` proves traversal fetches only the official machine/reference document, while the static-asset and cache-only guard regressions remain green.

## Rendered Markdown operation extraction

- Red: Countercyclical's provider-owned GitBook Markdown pages render requests as `` `GET` `/investments` `` inside markup. The raw-token matcher saw neither a whitespace-separated method/path pair nor an endpoint, so a fully cached official reference traversal returned no operation inventory.
- Green: the extractor scans visible HTML text in addition to raw request lines and accepts code-formatted method/path pairs. `TestParseBatchHTMLReferenceRecognizesMarkdownCodeOperations` proves the exact form; Countercyclical's staged 12-operation artifact validates and its promoted bundle loads through `pm connectors inspect countercyclical --json`.

## Final deterministic reconciliation

- Red: retained retry reports alone express an unfinished route, so the reconciler could not safely turn a provider-source drop into a terminal result even after every bounded official source had been exhausted.
- Green: B056 records one explicit `genuinely_blocked` outcome for each of the remaining 195 targets, with `official_source_exhausted`, a retained official source route, reason, and evidence. `TestFinalBlockedAttemptRequiresExplicitTerminalOutcome` rejects malformed terminal outcomes and a later retry, while the reconciler proves the exact final partition: `231 materialized + 195 genuinely_blocked + 0 retry_pending = 426`; it also conserves `24 reachable`, `217 foundation_pending`, and `3 declared + 422 unknown + 1 not_applicable` rate-limit declarations. Faker is recorded as a valid v2 zero-operation bundle because its target has no external HTTP/API surface.

## Derived runtime endpoint ledger

- Green: `connectorgen surface-sync` regenerated only the embedded runtime operation-endpoint ledger from the now-complete bundle set. The immediate `surface-sync --check` scan reports no remaining bundle-field or ledger drift, so the generated index remains executable rather than a hand-authored metadata change.

## V2 inventory regression alignment

- Red: the complete `cmd/connectorgen` suite still asserted Gorgias and Notion's superseded v1 operation-ledger shape after their cited v2 source inventories were materialized. In particular, it looked for blocked-operation source URLs in the v1 field instead of endpoint-local v2 provenance, and omitted Notion's separately blocked generic `POST /v1/search` contract from its qualified-binding row total.
- Green: the two assertions now require v2 plus endpoint-local citations. Notion retains 51 unique provider operations in 55 classified rows: its two stream-specific search arms and one generic blocked search contract are distinct truthful bundle views of the same provider request. The focused tests and complete `go test -timeout 20m -count=1 ./cmd/connectorgen` suite pass.
