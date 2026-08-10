# PLAN: 426 connector artifact sweep

## Authority and execution mode

Scope authority is `/Users/karthiksivadas/karthik-agent-workspace/data/cli-mass-artifact-materialize-r1/CAPTAIN-ORDER-fast-426-terra-20260809.md`. It supersedes the earlier pilot/defer and broad generator-review orders: the current target is a static, complete 426-connector JSON surface sweep, not a further generic-generator initiative.

This concise record is the required inline/manual GSD fallback. `scripts/gsd doctor`, the sources for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`, and `go run ./cmd/agentcontractgen check` passed. The phase identifier is intentionally non-numeric and uses `PLAN.md`, so the generated role workflows cannot safely bind their numeric phase filenames; the one-worker captain lane therefore records its Red/Green evidence here and in `TDD-LEDGER.md` instead of spawning incompatible roles. The task is executed by one worker without spawned GSD roles.

Every reconciled target retains one provider-cited `rate_limits.json`. The existing engine schema and loader remain the foundation; this slice adds the missing production embed, deterministic declaration producer, target-only schema/load accounting, and ledger totals without adding runtime pacing behavior.

Required skills loaded: `golang-how-to`, `golang-cli`, `golang-documentation`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

## Recovered pipeline checkpoint reconciliation (2026-08-10)

Before replay, the pipeline-only head was preserved at `origin/preserve/cli-mass-artifact-materialize-pipeline-20260810` (`858b72e0`). Its checkpoint commits `358d0632` and `318f3638` describe the same 194-target checkpoint and final 426-target ledger partition already represented by `ba950a91` and `620d509a` on the remote branch. Their conflicts are limited to an older copy of the reconciliation code and phase record; the retained branch carries the later rate-limit accounting, bounded reference parsing, and promoted-native command-surface repair. The final aggregate indexes are regenerated from the reconciled bundle set, while the preservation branch remains the durable source for the original pipeline history.

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

## Promoted-native command-surface forwarding

- Red: the materialized bundles for `apify-dataset`, `basecamp`, `copper`, `google-classroom`, `google-pagespeed-insights`, `metabase`, and `rootly` declare 25 implemented ETL commands, but `pm <connector> --help` exits 2 with `unknown command`. `bundleregistry.New` first constructs the hook-backed engine connector and then replaces it with the promoted legacy-native `definitionConnector`; the wrapper forwards only the bundle definition and configuration constraints, so it loses the bundle-owned `CommandSurface` required by the dynamic CLI.
- Green: `definitionConnector.CommandSurface` forwards `engine.Base.CommandSurface`, retaining the established native `Check`/`Read` implementation while exposing the already-declared JSON command surface. A table-driven CLI regression proves all seven bare connector namespaces and one implemented command-help route each render successfully and no longer hit the unknown-command path. This restores existing runtime help/discovery only; generated manuals, CLI docs, and website data have no independent authored text to update.

## PR #3957 CI-failure remediation (2026-08-10)

This is an in-scope remediation for issue #3958 / PR #3957. The current branch has a complete
426-target ledger; this slice repairs four CI gates without changing connector declarations or
answering the parked no-mistakes review gate. The captain fixed the order and supplied the
security disposition, so no product decision is pending.

The lifecycle is executed inline under the established manual GSD fallback: `discuss-phase`,
`plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were resolved through
`scripts/gsd`; the non-numeric existing phase and the single-worker constraint make generated
role workers incompatible. `scripts/gsd doctor` and `go run ./cmd/agentcontractgen check` are
green before implementation. Required skills used for this Go/CLI/security/test/website slice:
`golang-how-to`, `golang-cli`, `golang-error-handling`, `golang-lint`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-project-layout`, and `vercel-react-best-practices`. The repository-named design-only
website skills are unavailable in this runner; the website change is a focused test assertion,
not a UI design change.

1. **govulncheck first.** Red: `govulncheck ./...` stops before scanning because the two
   task-scoped scripts in one `scripts/` package both define `main`. Green: move each standalone
   program and its tests into its own `scripts/<command>/` package, update durable command
   references, then run the actual vulnerability scan so a passing check means packages were
   examined rather than merely compiled.
2. **Reference-link scheme allowlist.** Red: a `data:` candidate can pass the early selected-link
   filter even though later same-host and HTTPS gates reject it. Green: parse candidates at the
   filter boundary, permit only relative references or explicit `http`/`https` schemes, and add
   a regression proving `data:` is refused while a relative OpenAPI link remains eligible. This
   is defense in depth; the downstream same-host and HTTPS admissions currently prevent an
   exploit.
3. **Connector-boundary scanner collision.** Red: generic static asset literals `.rss` and
   `/rss.xml` are tokenized as the `rss` connector and incorrectly reported as shared
   connector policy. Green: make the scanner recognize asset suffix/filename literal context,
   retain detection for a standalone `"rss"` connector literal, and keep the asset filter text
   unchanged.
4. **Website search regression.** Red: `management_token` now has enough matching generated
   connector metadata in the 552-connector corpus that 100ms ranks 14th, outside the stale
   12-result test window. Green: use a connector-qualified query and assert the intended result
   is present, avoiding a corpus-size-dependent top-N expectation; run the focused Vitest test.

The verification order mirrors the requested fix order: focused script package tests and actual
`govulncheck ./...`; focused `cmd/connectorgen` tests; focused boundary package tests plus the
real connector-boundary gate; focused website test; then `surface-sync --check`, selected
non-test verify gates, and the PR checks triggered by the push. No live provider request,
credential, connector materialization, or no-mistakes gate response belongs to this remediation.

`make verify` exposed one mechanical consequence of making `./...` loadable: its tidy step now
correctly identifies the pre-existing production import of `golang.org/x/net/html` as direct.
The only module change retained is that direct/indirect classification in `go.mod`; no module
version, checksum, or dependency is added. It will be committed with this coherent green slice,
then Verify is rerun from the committed baseline so `tidy-check` can validate the exact result.

## Certify-timing budget calibration (2026-08-10)

The captain approved a narrow quality-gate calibration for PR #3957 after a controlled
main-versus-branch measurement. Scope is limited to `CERTIFY_TIMING_MAX_DURATION` in the
Makefile, the PR body, and this GSD/TDD evidence; no test topology, connector declaration,
runtime behavior, credentialed check, or no-mistakes gate changes. The 25 real harness calls and
92 real CLI invocations are unchanged on both revisions.

**Red:** the hosted Verify job reaches the real timing gate and fails at `421.290s` against the
old `210s` cap. A same-host, precompiled, `-count=1` two-run comparison measures 2,060 versus
13,534 endpoint-ledger entries (6.57x corpus growth) and 159.268s versus 295.117s mean test
elapsed time (1.85x). The work grows substantially less than the corpus, so this is not a
super-linear certify regression to optimize or hide.

**Green plan:** raise only the cap from 3m30s to 8m (+4m30s / 129%). The 8m limit leaves 58.710s
(13.9%) headroom over the observed hosted run while preserving the unchanged real-invocation
contract. Run `make certify-timing` from the edited tree, then push and require the hosted Verify
job to pass. Record the exact 6.57x/1.85x reasoning in the PR body so a future threshold change
has a reproducible decision record.

The lifecycle remains an inline/manual fallback: `scripts/gsd doctor` plus the resolved
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts
are used, while compatible isolated GSD roles are unavailable and this lane must not spawn agents.
Required skills used for this CI/test-performance slice are `golang-how-to`,
`golang-continuous-integration`, `golang-testing`, and `golang-benchmark`. CLI help/manual/website
parity is not applicable because no CLI surface, help, docs, or generated catalog changes.

## Existing command-surface preservation and redaction propagation (2026-08-10)

Captain scope is deliberately limited to the two named regression classes: preserve the established
spelling of flags on existing implemented reverse-ETL commands, and propagate every write action's
`redact_fields` into the corresponding generated command surface. Xero's independent direct-read
ledger inconsistency is not changed in this slice.

**Red:** `materializeCLISurface` regenerates `CLIFlag.Name` directly from `record_schema` field
paths (for example `receipt_handle`) and, for an existing implemented command, retains only its
path/source path/summary. That silently changes Amazon SQS's shipped `--receipt-handle` contract.
It also drops `WriteAction.RedactFields` while building the command, leaving Google Search
Console's `add site apply` without its action-declared `site_url` redaction. The hosted Verify
regressions reproduce both without credentials or provider I/O.

**Green plan:** add focused materializer regressions before the implementation. An existing
reverse-ETL command is an executable public contract, so materialization retains its complete
registered command (path, flags, optional fields, examples, approval metadata, and redactions) and
refreshes only the cited API-surface references. It must never re-derive a shipped flag spelling
from a record schema. The generated command's `redact_fields` is a deterministic union of the
retained command declarations and the action declaration. Regenerate only the affected Amazon SQS
and Google Search Console command surfaces, then scan all 552 bundles for (a) a
pre-materialization existing command whose flag spelling changed and (b) any action redaction
absent from its write command. The two counts are reported rather than hidden. The scan is
read-only and contains no live provider or credential use.

CLI parity applies to generated connector surfaces: verify bare namespace/help for Amazon SQS and
Google Search Console, generated command/website artifacts if the repository's generator reports
them, and `surface-sync --check`. Hand-authored docs are not applicable when the command path and
semantics are retained; the regression covers spelling, redaction propagation, and runtime
surface loading. This uses the required inline/manual GSD fallback after resolving
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`; compatible
isolated GSD roles are unavailable in this single-worker recovery lane. Required skills:
`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-documentation`, `golang-design-patterns`, and
`golang-structs-interfaces`.

## Full audited-corpus command-surface regeneration (2026-08-10)

**Superseding captain decision:** the earlier two-connector artifact scope is expanded to every
finding in the post-fix audit. Seven observed established-command flag regressions and seven
observed one-sided redaction gaps are not shippable: the former can break an already working
command and the latter can expose a field declared secret. The generated output is not patched
by hand.

**Red:** the 552-bundle audit reports six remaining connectors with 270 established flag spelling
changes across 196 reverse-ETL commands (Bitbucket, Chatwoot, Freshchat, Gorgias, Jira, and Lever
Hiring), plus six remaining connector families with action redactions absent from 146 generated
commands (Freshchat, Mailchimp, Recurly, Stripe, Xero, and Zendesk Support). The named Amazon SQS
and Google Search Console cases are already covered by the same generator regression.

**Green plan:** retain the full pre-sweep command contract only where a command existed before
materialization; use the current endpoint/operation inventory for its refreshed references; and
run the repaired materializer to regenerate every affected CLI surface. For a newly materialized
command, derive its flags normally but always union `WriteAction.RedactFields` into the command.
Regenerate the complete affected corpus through `connectorgen`, run the baseline-contract and
write-redaction audits across all 552 bundles, and require both counts to be zero. Then run
surface-sync, focused command/connector tests, validation, and the relevant verify gates before
pushing. The GSD lifecycle remains the documented inline/manual fallback; required skills remain
the Go CLI/testing/security set listed above.

## Verify CLI registry reuse after corpus growth (2026-08-10)

The hosted Verify log establishes a real post-sweep gate failure: `internal/cli` exceeded its
20-minute package timeout while the Bahmni command matrix repeatedly constructed the complete
552-bundle registry. The matrix is a useful end-to-end contract; weakening it or raising its
timeout would conceal the same production startup cost for every connector invocation.

**Red:** `TestBahmniDeclaredCommandMatrixIsRecognizedOrExplicitlyBlocked` reconstructs the
immutable embedded registry for every help and preflight invocation, and the hosted full package
times out with the test still active. The definitions are embedded and validated once per process,
so this repeated parsing has no semantic benefit.

**Green plan:** cache only the loaded immutable bundle slice inside `bundleregistry`, while still
constructing a fresh `connectors.Registry` and fresh engine connector wrappers for each caller.
The cache test must prove the loader runs once and that callers receive independent slice headers;
the existing registry and CLI matrix tests then prove native overlays and command behavior remain
unchanged. Run the focused registry/CLI tests, then the full CLI package and hosted Verify. This
is an inline/manual GSD fallback under the existing lifecycle and uses `golang-how-to`,
`golang-cli`, `golang-testing`, `golang-performance`, `golang-safety`, and
`golang-structs-interfaces`.

## Derived full-surface CLI expectations (2026-08-10)

The now-completing CLI package exposes two stale derived expectations rather than a runtime
failure. The exact four GitLab ETL streams remain executable, while the 1,741 other cited
provider endpoints are deliberately visible as `not_implemented` operation rows. The root manual
golden likewise predates the 552-connector catalog.

**Red:** `TestGitLabCommandSurfaceAdvertisesOnlyCitedReadCommands` asserts a total command count
of four despite `api_surface.json` containing 1,745 cited endpoints (four stream-backed plus
1,741 honest operation rows). `TestGoldenTranscripts` reports the old limited connector catalog
instead of the generated 552-connector root manual.

**Green plan:** retain exact assertions for the four executable GitLab streams and prove every
remaining command is a one-endpoint, operation-backed, `not_implemented` row; derive the total
from the API surface rather than hand-maintaining a magic count. Regenerate the golden transcript
with its documented test writer, inspect the generated diff, then rerun the focused tests and the
full CLI package. This follows the same inline/manual GSD fallback and uses `golang-cli`,
`golang-testing`, `golang-documentation`, and `golang-safety`.

## Connector-wide global-flag contract preservation (2026-08-10)

The regenerated whole-corpus comparison against `f96a47e80` revealed eleven additional existing
surfaces whose `global_flags` arrays had been removed by the same replacement path: Bitbucket,
Freshchat, GitLab, Gmail, Google Search Console, Help Scout, HubSpot, Jira, Lever Hiring, Stripe,
and Xero. These are public namespace/help contracts, including credential, output, paging, and
reverse-ETL controls.

**Red:** `materializeCLISurface` reconstructs a new surface without copying
`bundle.CLISurface.GlobalFlags`, so any materialization silently erases those accepted flags.

**Green plan:** retain a copied global-flag slice whenever a source CLI surface exists; extend the
focused materializer regression to assert it; regenerate each affected surface through
`connectorgen batch materialize` from the preserved pre-sweep source contract and current endpoint
inventory. A complete `f96a47e80` comparison must report zero global-flag differences alongside
the required zero spelling and redaction counts.

## Harvest conformance rate-limit fixture identity (2026-08-10)

**Red:** after `rate_limits.json` declares Harvest's provider-documented `general-account` policy,
the fixture driver builds `RuntimeConfig` without an opaque `CoordinationIdentity`. The production
rate-limit resolver correctly refuses to derive a scope from that zero value, so `check_fixture`
and every replayed read fail before an HTTP request.

**Green plan:** make `runtimeConfigForEngine` create a deterministic synthetic identity from
constant fixture-only public values. Add a focused conformance regression that loads the
Harvest bundle and proves both `check_fixture` and a fixture-backed read pass with the policy still
active. This changes neither real credential handling nor `rate_limits.json`; it only supplies the
non-secret test input the runtime contract already requires. Run the focused conformance package,
Harvest's dynamic checks, rate-limit engine tests, and then hosted Verify. This is the inline/manual
execution of resolved `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` prompts because compatible isolated GSD roles are unavailable. Skills: `golang-how-to`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, and `golang-testing`.
