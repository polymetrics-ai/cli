---
phase: cp16-generated-paths-references
reviewed: 2026-09-08T17:24:38Z
depth: deep
files_reviewed: 84
files_reviewed_list:
  - ".githooks/pre-commit"
  - ".github/workflows/release.yml"
  - ".github/workflows/verify.yml"
  - "AGENTS.md"
  - "Makefile"
  - "cmd/connectorgen/gen.go"
  - "cmd/connectorgen/source_foundation_proof_test.go"
  - "cmd/connectorgen/source_visibility.go"
  - "cmd/connectorgen/source_visibility_shape_165_test.go"
  - "cmd/connectorgen/source_visibility_test.go"
  - "cmd/connectorgen/testdata/flag-ownership-181.json"
  - "cmd/connectorgen/testdata/foundation-proof-inputs-153/README.md"
  - "cmd/connectorgen/testdata/foundation-proof-inputs-153/inputs.json"
  - "cmd/connectorgen/vnext_admission.go"
  - "cmd/connectorgen/vnext_command_paths_173_test.go"
  - "cmd/connectorgen/vnext_flag_env_185_test.go"
  - "cmd/connectorgen/vnext_flag_grammar_185_test.go"
  - "cmd/connectorgen/vnext_flag_integrity_182_test.go"
  - "cmd/connectorgen/vnext_flag_namespace_182_test.go"
  - "cmd/connectorgen/vnext_flag_ownership.go"
  - "cmd/connectorgen/vnext_flag_ownership_181_test.go"
  - "cmd/connectorgen/vnext_graph.go"
  - "cmd/connectorgen/vnext_lock.go"
  - "data/connector-canon/batch1-foundation-assessments.json"
  - "data/connector-canon/batch1-foundation-demand-register.json"
  - "data/connector-canon/batch1-source-lane-manifest.json"
  - "docs/connector-canon/SOURCE-LOCK-VNEXT.md"
  - "docs/connector-canon/foundations/catalog.json"
  - "docs/connectors/github/MANUAL.md"
  - "docs/connectors/github/SKILL.md"
  - "docs/connectors/gitlab/MANUAL.md"
  - "docs/connectors/gitlab/SKILL.md"
  - "docs/releasing.md"
  - "go.mod"
  - "go.sum"
  - "internal/app/app.go"
  - "internal/app/diagnostic_coordinates_173_test.go"
  - "internal/app/postgres_confirmation_173_test.go"
  - "internal/app/provider_alias_withholding_184_test.go"
  - "internal/app/source_visibility_test.go"
  - "internal/cli/asana_flag_grammar_186_test.go"
  - "internal/cli/cli.go"
  - "internal/cli/diagnostic_coordinates_173_test.go"
  - "internal/cli/generated_path_binary_173_test.go"
  - "internal/cli/github_transport_binary_test.go"
  - "internal/cli/help_carry_173_test.go"
  - "internal/cli/help_contracts_171_test.go"
  - "internal/cli/notion_flag_known_179_test.go"
  - "internal/cli/postgres_checkpoint_oracle_173_test.go"
  - "internal/cli/postgres_transport_binary_integration_test.go"
  - "internal/cli/production_registry_binary_173_test.go"
  - "internal/cli/production_registry_fixture_178_test.go"
  - "internal/cli/provider_cohort_wire_184_test.go"
  - "internal/cli/provider_config_plan_185_test.go"
  - "internal/cli/provider_env_carrier_185_test.go"
  - "internal/cli/provider_flag_consumers_184_test.go"
  - "internal/cli/provider_flag_help_185_test.go"
  - "internal/cli/provider_flag_lifecycle_184_test.go"
  - "internal/cli/provider_query_wire_184_test.go"
  - "internal/cli/source_visibility_shape_165_test.go"
  - "internal/cli/transport_physical_oracle_173_test.go"
  - "internal/connectors/command_flag_ownership.go"
  - "internal/connectors/command_flag_ownership_182_test.go"
  - "internal/connectors/defs/github/cli_surface.json"
  - "internal/connectors/defs/gitlab/cli_surface.json"
  - "internal/connectors/engine/bundle.go"
  - "internal/connectors/engine/command_endpoint.go"
  - "internal/connectors/engine/command_endpoint_diagnostics_173_test.go"
  - "internal/connectors/engine/command_surface_lookup_188_test.go"
  - "internal/connectors/engine/connector.go"
  - "internal/connectors/manifestindex/index_gen.go"
  - "internal/connectors/source_capabilities_gen.go"
  - "internal/connectors/source_visibility.go"
  - "internal/connectors/source_visibility_codec_174_test.go"
  - "internal/connectors/source_visibility_test.go"
  - "scripts/schema-test-requirements.txt"
  - "scripts/tests/release-production-layout.sh"
  - "scripts/tests/release-size-budget.sh"
  - "scripts/verify-release-assets.sh"
  - "scripts/verify-release-size-budget.sh"
  - "website/content/docs/github-cli-surface.mdx"
  - "website/data/connectors.generated.json"
  - "website/lib/connectors.catalog.data.generated.json"
  - "website/lib/docs.generated.ts"
findings:
  critical: 2
  warning: 1
  info: 0
  total: 3
status: issues_found
---
# CP16 complete independent code review173

## Narrative Findings (AI reviewer)

**Result: issues_found — 3 distinct required corrections: 2 BLOCKER, 1 WARNING.** This is the complete independent judgment on the bound candidate, including the retained Notion carry. No finding count was used as a discovery cutoff. The two inherited open171 corrections are closed by the evidence below. Historical Bitbucket membership and complete historical/final CI remain explicitly limited; they are not concealed by the actionable count.

This report is reviewer-authored substantive content for mechanical publication as REPORT.md. No structural pre-pass was supplied. No source fix, regeneration, commit, push, integration, provider-live request or additional review stage was performed. Captain156 scheduling is for Firstmate/canonical-owner reconciliation after this complete result; this review does not grant a main merge or certify unfinished connector chains.

## Binding, scope and method

- Exclusive run: cp16-final-review-173-20260908T100555Z.
- Reviewer: /root/cp16_final_review_173, native 01a081e8-da23-7fc1-b831-0b711bbdfc35, gsd-code-reviewer, actual gpt-6-astra/xhigh, fresh context, no children. Parent: canonical owner native 01a07bcc-454f-75a3-89f3-8bbb91572206. Firstmate192 independently verified the actual profile and parentage; the original launch tool's lack of UUID introspection remains recorded.
- Base/merge base: 147690998756202aa07302eba82a775a9e62f96a. Final head: f920dff403471b4fd477991cc6e6e92c66d7e8e7. Tree: 506977cbedbfecb821384bb07570695ffb91d6bb.
- Frozen worktree: /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli, branch fm/cli-top100-declaration-batch-r1-cp16. Private detached clone: /Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp16-final-review-173-20260908T100555Z/candidate.
- Task delivery: issue4344 within programme4325; existing PR https://github.com/polymetrics-ai/cli/pull/4294. The API-reported PR base is main; the authorized programme integration destination is fm/cli-top100-declaration-batch-r1. This corrects the early RUN header's conflation of those two bases. Reviewer owns no PR or integration.
- Full unchanged Firstmate173 prompt adopted: 18,825 bytes, SHA 256 aa38d744a7f7e91f5e58a5358fa6d57921da4d0cd220fe206cc4075e69444a4a. Actual binding, complete implementation173/originalCP16, complete preceding171 report/seal/carry ledger, all bound supplements through current 191, installed GSD review prompt, and mandatory policy/project/skill context were read. The new192 ownership mapping is integration-owner work, not another reviewer assignment.
- Initial identity verification at 16:47:09Z established exact head/tree/base, clean detached independent private Git directory, different resolved cwd, no symlinks back, and clean frozen tracked state before any probe. All 49 bound context/evidence reference pins and all 324 changed-file pins matched. Both copies' 15,803 tracked file bytes were compared. Neither copy has .codegraph, so CodeGraph was inapplicable.
- Deep review covers 84 primary changed source/configuration/generated/documentation files listed in frontmatter, all changed production paths and affected callees/public consumers, plus source/evidence membership and byte audits. The other changed paths are 230 content-addressed fixture payloads and ten phase artifacts. The complete retained fixture corpus has 232 payloads; all were content-hash checked. Fixture hashing is not presented as 232 independent behavioral tests.
- Required skills used: project caveman and connector-lane-build-order; installed GSD context; exhaustive-review/evidence policy; Go how-to, CLI, testing, error handling, security, safety, lint, design patterns, structs/interfaces, context, concurrency, database and documentation, with required routing, runtime integration and CLI/help/docs parity references. Role artifact restriction was honored: only private phase REVIEW.md is authored. No Read/Write file tool is exposed; available read-only command tooling and purpose-built patch creation of REVIEW.md substitute for those unavailable tool names.

## Required findings

### CR-173-01 — BLOCKER: admitted command segments disappear in the hand parser

**Impact:** medium. **Classification:** shared source-authoring/admission correctness defect exposed by this review.

**Files:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_graph.go:244; /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/connectors/commandrunner/runner.go:1317–1320; consumer /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/cli/parse.go:42–80.

**Accepted contract:** CP16-02/03 require emitted command paths to remain exact reachable positional paths through the real generator, hand parser and resolver. Identifier validity alone does not establish positional-token validity.

**Reached evidence:** reviewer probe173-P01 renders and publishes an implemented command whose explicit alias is “widgets --help”. Leased generation loading, engine.Load, exact binding, PreflightRequest and deterministic lock-render --check succeed. Its healthy “widgets fixed” control survives the actual parseGlobal/parseFlags route. The actual parser drops --help and --root, treats --json as global control, and consumes “--widgets fixed” as an option; all four bad paths nevertheless pass CommandPathSegments. The parser probe fails at the intended assertions, after valid setup. The unchanged renderer probe exits0.

**Cause and siblings:** canonical validation delegates to safety.ValidateIdentifier, which allows a leading double hyphen; the downstream hand parser interprets that prefix as option syntax. Any --prefixed segment at any position, and every availability admitted through this shared guard, has the same invariant. Preflight-only path oracles share the blind spot. This does not assert that a current member of the 13,856-command corpus failed: exact current corpus sweep credit remains intact.

**Fix and regression:** enforce the actual positional-token grammar in the shared command-path guard, rejecting segments beginning -- without changing the hand parser or introducing aliases. Preserve legal distinct explicit aliases, source IDs and path mappings. Retain the real renderer/publisher healthy control and actual hand-parser roundtrip tests for global, PM and arbitrary option prefixes, in first and later positions.

**Disposition/owner:** required shared CP16 correction, canonical shared owner. No fix performed. Probe source, both initial infrastructure failures, corrected toolchain runs and cleanup are retained under173-P01.

### CR-173-02 — BLOCKER: Notion's supplied provider limit is discarded

**Original ID:** CP16-FLAG-NOTION-179, counted once here. **Impact:** medium. **Classification:** retained preexisting optional-flag sibling exposed by the CP16 ownership/fleet work.

**Files:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/connectors/defs/notion/cli_surface.json:598–601; /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/cli/cli.go:1670–1675; shared classification in internal/connectors/command_flag_ownership.go.

**Accepted contract:** explicitly supplied provider inputs must survive the CLI into their declared wire fields. Firstmate181 explicitly retains this optional sibling for CP23; deferral does not constitute correction.

**Reached evidence:** the exact preserved separately tagged regression and raw cp16-notion-known-red-181-01 were read. The omitted-flag control reaches the missing-credential boundary. A direct commandrunner control sends limit=5. The same valid flags after actual CLI preparation reach local POST /v1/blocks/meeting_notes/query with an empty JSON object, losing body.limit and failing the supplied-value assertion. Four run events, two pass; final source behavior is unchanged. This is actual reached-wire evidence, not a required-only sweep inference.

**Cause and siblings:** the declaration uses provider name limit while the shared PM-control classification causes prepareConnectorCommandFlags to remove it. The 22 GitHub/GitLab field corrections are separately verified and do not close Notion. The tagged known-red test is honest evidence, not another defect.

**Fix and regression:** CP23's Notion authoring/corpus owner must assign the provider field a safe distinct alias while preserving body.limit, schema and metadata and PM's own control. Update generated execution/help/manual/website surfaces as applicable, and make this exact supplied/omitted/direct-runner wire regression green.

**Disposition/owner:** required CP23 source-backed migration. The canonical integration owner retains custody until the actual CP23 owner accepts it; no owner acceptance is invented.

### WR-173-01 — WARNING: the new CDC advance oracle accepts a backward LSN

**Impact:** medium. **Classification:** newly introduced test-reliability defect; no product rollback is established.

**File:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/cli/postgres_transport_binary_integration_test.go:2715–2725; faulty comparison at 2724–2725, callers at 303 and 332. The earlier journal citation2729–2740 was wrong and is explicitly superseded by this citation-only revision.

**Accepted contract:** CP16-07/08 and the critical-oracle safeguard require independently valid before→after checkpoint advancement at the post-barrier and resumed transaction witnesses.

**Reached evidence:**173-P02 executes the existing eight-case checkpoint oracle successfully, plus independent forward/unchanged/backward controls. Both backward assertions fail. Using the real native token representation, before primary LSN0/20 and after 0/10 with a one-second later CommittedAt returns true. The original opaque-token model also demonstrates the same inequality-only issue. Native postgres/cdc.go parses PostgreSQL LSNs and records transaction-end/commit positions; position inequality is not advancement.

**Cause and siblings:** postgresCheckpointAdvanced173 checks only unequal Position and a later wall timestamp after source/mechanism equality. Both callers therefore can certify a regressing checkpoint as “advanced.” Physical target rows, actual restart, lease exclusion and unchanged-state controls do not independently prove LSN monotonicity.

**Fix and regression:** decode and compare actual PostgreSQL LSNs monotonically, validate stable source/protocol checkpoint identity, and require a genuinely newer transaction-end LSN for these two separately inserted transaction witnesses. Include forward, unchanged, backward, malformed, and same-LSN-with-different-tie/time falsifiers. Rerun the helper falsifier and affected physical database witness after correction.

**Disposition/owner:** required shared reference-evidence correction. Positive physical database rows/restart evidence remains credited within its actual limits. P02 has no database startup: it selects tagged helper tests only.

## Ten original obligation dispositions

| Obligation | Independent disposition and evidence |
|---|---|
| CP16-01 — historical/current operation identity | **Reviewed current set; specifically blocked historical identity reconstruction.** Independently joined all 297 current Bitbucket source IDs to exact method/path/source pointers and the inventory. Keyed sets match; sorted array order is not treated as identity loss. All five current command joins match their exact source targets, with three implemented and two declared blocks. The historical50/28 set came from an uncommitted overlay whose complete members are unavailable. Neither five nor297 reconstructs its exact old→new membership. Recoverable parameter families are exercised, but the missing historical rows remain an explicit evidence limit. |
| CP16-02 — deterministic parser-valid generated paths | **Required correction CR-173-01.** Current aliases, duplicate/collision/brace/whitespace refusal, deterministic re-render, exact source operation and flag maps were traced. Current corpus exact reachability is credited. Independent repeated/adjacent/literal/hyphen/underscore legal shapes preserve distinct source targets. Shared admission still accepts option-prefixed segments that the real parser consumes. |
| CP16-03 — independent projector falsification | **Required correction CR-173-01; other tested binding invariants reviewed.** Real publication/leased loading/resolution and independently expected source/flag tuples were exercised, including omitted required invocation input and same-count wrong target. Retained-source projected omission is separately checked by P04. P03's authoritative-source omission interpretation is withdrawn, as detailed below; its original failed evidence is retained. Existing alias/provenance/source-reorder falsifiers reach admission rather than only string helpers. |
| CP16-04 — complete registry and built binary | **Reviewed for exact applicable reachability scope.** Independently reconstructed13,856 unique connector/path identities from current pinned execution JSON and reconciled every in-process run/pass and built-binary completed/recovery leaf. Missing, extra, failed and skipped applicable leaves: zero. Full in-process is one terminal run; original binary run remains interrupted, with explicit191 completed-leaf recovery. These required-argument/credential/approval/declared-block witnesses do not cover every optional supplied field or certify provider execution. Notion therefore remains CR-173-02. Unknown/invalid paths and fixture schema/control boundaries are separately tested. |
| CP16-05 — Bitbucket CP19 follow-up | **Reviewed concrete deferred scope and custody.** Tracked BITBUCKET-REACHABILITY.md/json retain all 297 exact source keys and current five joins; CP19 owns the full source/artifact/runtime-proof chain and final complete reachability after generation. Canonical programme owner retains custody pending Firstmate's actual CP19 worker binding. No CP19 worker is claimed launched/accepted and no complete Bitbucket chain is credited from CP16 shared tests. Historical50/28 uncertainty travels with this follow-up. |
| CP16-06 — full hermetic API reference | **Reviewed and credited within fixture scope.** Built-pm API physical and warehouse roundtrip receipts reach real App/registry, source read, DuckDB/WAL/Parquet, approved destination write, provider readback, persisted receipt/checkpoint, replay/refusal and cleanup frontiers. Exact rows and event order, not only exit0, were inspected. Physical carrier proves one source/target record and zero-residue cleanup; broader roundtrip intentionally reports zero_provider_residue=false and is not relabeled. Earlier compatible ack/failure contracts are reused with explicit source continuity. |
| CP16-07 — authorized local database reference | **Reviewed physical/reference scope; required evidence correction WR-173-01.** Actual disposable PostgreSQL full flow records1,001 exact source/target rows across1000+1 pages, physical WAL/Parquet and real CDC post-barrier/restart. Live lease refuses successor without state change; actual expiry precedes new approval/resume. History, full overwrite and unchanged-upsert mode witnesses execute real database operations. The checkpoint monotonicity claim exceeds its faulty oracle; no broader power-loss or customer database certification is made. |
| CP16-08 — structural storage/identity/order | **Reviewed runtime continuity and durable controls; required evidence correction WR-173-01.** Traced connection identity/AuthCohortKey→owned location/owner.json→append+fsync WAL→atomic single Parquet→manifest/receipt→ack/checkpoint and workset retirement/reopen. Independent row/byte/history/state comparisons and exact allowed receipt-owned three-file retirement protect unrelated work. Replay/schema-refusal and live-lease refusal retain before-state; restart is process-kill coverage. Backward-LSN oracle limitation remains isolated and explicit. |
| CP16-09 — honest evidence classes | **Reviewed.** Source records/Atlas/proof pins are authoring-only, selected generated safe metadata remains separate from ordinary execution, seven execution lanes remain distinct from sync modes. CP07/08 selection witnesses retain their original class. Current full hermetic references establish fixture execution; neither inventory, credential-boundary sweeps nor these fixtures certify private providers/customer DBs or all ten connector chains.4343 source identities and 30401 seven-lane cells remain unchanged. |
| CP16-10 — cumulative handoff, carries, checks | **Complete independent review with 3 required corrections and disclosed limitations.** All bound changed paths, all inherited rows, affected CP11–15 contracts, original failed/interrupted receipts, exact normal/race identity differences and final-source applicability were assessed. Scoped local checks are credited; a truncated historical Verify log and absent complete final CI result are not called global green. No parallel connector start or integration was performed by this reviewer. |

## Inherited and baseline carry matrix

Prior severities below preserve original impact labels. Closed findings are not counted again. Closures retain original reports, RED/GREEN receipts and causal links in the bound ledger; this report does not rewrite historical evidence.

| Original ID | Invariant / original impact | Current independent disposition |
|---|---|---|
| CR-158-01 | Admission-derived independent known obligations, low | Remains closed by independent164. Current source/known/demand comparison retains source identities and obligation authority; changed reporting of stale proof pins does not remove known obligations. No regression found. |
| WR-158-01 | Real two-operation same-file selector refusal, low | Remains closed164. Selection mechanism unchanged; current source inspection/semantic controls retain this boundary. No broad archive replay needed. |
| CR-164-01 | Strict selected schema/lineage shapes, low | Remains closed170. Compressed selected payloads still reach strict semantic validation; readable malformed fixtures, source-cell selection and public data-error controls remain. Encoding does not promote authoring or add fallback. |
| CR-170-01 | Definition-owned destructive confirmation, medium | Remains closed171. CP16 extends the existing closed managed-PostgreSQL target seam; persisted/fresh-App valid/tampered/missing/changed-target tests and actual DB grant/apply flow cover the affected path. Existing definition-owned transport policy is preserved. |
| CR-170-02 | Multipart allowed_media_types diagnostic coordinate, low | Remains closed171. Existing field-specific producer reason and private-cause/public-consumer chain preserved by diagnostic sibling coverage. |
| CR-170-03 | change_apply/non-change-capture diagnostic, low | Remains closed171. Producer-owned rejection coordinates/reason remain; changed target diagnostics add exact context without restoring generic replacement causes. |
| CR-170-04 | Static/dynamic polling-help semantics, low | Remains closed171. Current polling tests and actual help distinguish declarations from runtime eligibility; no obsolete-wording requirement restored. Alias BD-CP15-169-01 stays attached. |
| CR-170-05 | Real source-origin restriction and ordering, low | Remains closed171. Prior project/vault/credential/provider order fixtures remain applicable; no source-proof reader becomes an execution gate. Alias BD-CP15-169-02 retains its derived34-file baseline limitation. |
| CR-170-06 | ETL mode/help contract, low | Alias of CR-171-02, closed with that invariant here; not a separate count. BD-CP15-169-03 is the same alias. |
| CR-170-07 | Loader Fatal/goroutine test reliability, low | Remains closed171. Loader synchronization code is unchanged; current cold/cache/race and bounded controls preserve coverage. No fatal-in-loader regression found. |
| CR-171-01 | Safe useful exact target diagnostic and complete private cause, low | **Closed by this review.** Changed engine validators identify exact method/path/property and retain useful producer reason through loader/App/CLI. Original semantic RED and green/race controls were read; healthy declarations, sibling failures, whole/joined cause preservation and actual JSON/text output with private sentinels are covered. A private error assertion alone was not accepted. |
| CR-171-02 | Every complete ETL mode section and truthful semantics, low | **Closed by this review.** Seven mode sections are independently specified, parsed and semantically checked. Omitted entire section, false meaning and duplicate-heading falsifiers reject; valid current semantics pass. Normal and race selections differ and were reconciled explicitly below. |
| BD-CP15-168-01 | Typed-destination confirmation baseline, medium | Remains merged into closed CR-170-01. Original derived34-file/current-dependency overlay was never complete historical-tree execution. |
| BD-CP15-169-01 | Polling-help baseline, low | Remains merged into closed CR-170-04; current static/dynamic contract controls preserve the original behavioral distinction. The original limited overlay evidence is retained. |
| BD-CP15-169-02 | Source-origin/project/vault baseline, low | Remains merged into closed CR-170-05; no provider send or secret exposure was established by that original baseline. Current declared-restriction/order controls remain applicable. |
| BD-CP15-169-03 | ETL compatibility-name/mode wording baseline, low | Alias of CR-170-06 and CR-171-02; closes with the complete seven-mode semantic oracle here, without another actionable count. |
| CI-174-01 | x/text vulnerability dependency | Current module pins use0.39.0; original update/tidy/govulncheck evidence is applicable and terminal. Credit is the recorded scan at its exact source/time, not a claim about future advisories or a complete final CI run. |
| CI-174-02 | Release size-only ceiling | Superseded by captain176: size_budget_requirement_removed_by_captain176; optimization_deferred. Tests now accept valid oversized artifacts and still reject integrity faults.177's51-input closure corrects the original omitted release-source capture; earlier receipt limitation remains. No removed ceiling is resurrected as a defect. |
| CI-175-01 | Observed Verify failures | All 25 observable original top-level failure groups have applicable148/148 local proof; schema dependencies and immutable historical proof fixture handling were inspected. Original61,508-byte Verify log is truncated, so unobserved groups and complete final CI remain unproven. This is a disclosed evidence limit, not a fabricated full-CI success or a new product defect. |
| CP16-FLAG-NOTION-179 | Optional supplied provider body.limit, medium | **Open; CR-173-02, counted once.** Canonical owner retains it for actual CP23 acceptance and source-backed repair. |
| Historical issue4344 50/28 | Exact absent overlay membership | Specifically blocked historical reconstruction; current 297 source set and CP19 full-chain follow-up retained. No silent loss or invented rename mapping. |

## Architecture and affected-surface map

| Entry / changed owner | Traced chain, side effects and downstream boundary |
|---|---|
| Schema4 source authoring; vnext_lock/graph/flag_ownership/admission | Immutable source→canonical operation with original source index/ID→safe explicit command/flag projection→exact authored-command/alias ledger invariant→execution render→engine.Load/manifest/commandrunner preflight→staged publication→leased read. Projection copies raw command metadata and only authorized alias Name values; no authoring/runtime second reader. Admission precedes publication. CR-173-01 identifies the remaining parser grammar mismatch. |
| Connector invocation; cli.go/control registry | Actual parseGlobal/parseFlags→prepare flags/config/env→registry/App→commandrunner required/enum/schema/binding checks→credential/approval boundary→closed executor. Shared PM classification preserves the old16-control filter. Narrow exact shared integer limit tuple is retained only for approved ETL/direct-read stream semantics. Provider config/plan/env aliases, approval withholding and exact request fields were traced. Notion's retained sibling is CR-173-02. |
| Diagnostic producers; bundle/command_endpoint | Specific malformed declaration→producer-owned property/reason→wrapped whole cause→loader/App/CLI classification→safe JSON/plain output, before provider execution. Healthy owners reach their intended branch; compound companion errors and secret sentinels remain private. No new joined-error producer or changed collision-retry policy. |
| Command surface projection; engine connector.go | Bundle operations→invocation-local first exact-ID map→existing operation flag/parameter synthesis→fresh caller-owned output. Duplicate first-match, missing/nonREST, reorder, same-name changed bundle, caps, output mutation and concurrency controls preserve semantics. No persistent cache/invalidator or new lifetime. |
| Selected source metadata; gen/source_visibility/manifest | Deterministic gzip generation→generated encoded/decoded length and SHA→selected bounded decode through EOF/CRC→single-member/trailing/hash/strict JSON/semantic validation→safe result. Read-only stream closure occurs after integrity check. Runtime does not open Atlas/proof/source locks; malformed selected metadata remains separate from ordinary execution. |
| App managed-target confirmation | Persisted plan→fresh load/sealed target/confirmation validation→existing closed definition-owned confirmation policy→grant→actual apply. Source change is a closed managed-PostgreSQL seam; missing/tampered seals and changed targets cannot borrow authority. No generic SQL/HTTP write. |
| API and PostgreSQL transport references | Real App/registry/binary→source page→connection-owned warehouse stage/WAL/Parquet→preview/approval→destination physical request/DB transaction→readback→receipt/ack/checkpoint. Reopen checks structural owner, manifest and hashes; retries/replay retire only explicitly committed receipt-owned work. Live lease survives killed process; exact expiry is observed, not accelerated. WR-173-01 limits checkpoint-advance certification. |
| Help/docs/website and release/check tooling | Provider aliases propagate through help/bare/inspect/manual/skills and both website projections. Four final generated manual/skill files are alias-only deltas; remaining1,533 generated docs unchanged. Timeout policy changes Make/Verify/hook budgets without production timeout/lease changes. Release removes only disallowed size ceilings, retaining content/digest/archive integrity. |

## Contract-to-evidence and actual fault frontiers

| Contract | Healthy reached phase / independent observable | Fault frontier, owner and allowed transition | Judgment |
|---|---|---|---|
| Exact command/source/path binding | Real renderer publication/leased bundle, exact source operation ID/method/path and required flag; actual parser healthy control | Collision/brace/whitespace/mapping loss must fail before publication or runtime provider I/O; independent projected mutation preserves original source | CR-173-01 remains; P03 interpretation and P04 distinction preserved |
| Supplied provider fields / PM separation |22 frozen GitHub/GitLab fields;21 actual approved local writes and one actual query; exact returned bodies, EnvOnly and help/control consumers | Withheld approval/changed provider input/PM config-plan controls must not be confused; exact Name-only migration preserves wire targets | Corrected cohort credited; Notion open |
| Diagnostic privacy and useful reason | Valid target owner reaches semantic validation; exact safe property/rule observable in text and JSON, complete cause inspectable internally | Invalid method/path and sibling producers reject before send; joined private companion and secret sentinel remain undisclosed | CR-171-01 closed |
| Closed destructive approval | Persisted plan reloaded into fresh App, valid sealed PostgreSQL target receives grant and physical execution | Missing/tampered/changed target/confirmation fail before grant/apply; no broader ordinary action capability | Current shared policy preserved |
| Physical row/byte integrity | API1 row; DB1001 independently seeded IDs1..1001 with sequence=id*10 and known label values; WAL/Parquet hashes, records, page1000+1 and target contents | Same-count wrong row, omitted record, changed hashes/files must fail; exact before snapshots precede later cleanup/retirement | Physical witnesses and falsifiers credited |
| Committed workset retirement | Retained receipt identifies committed work; independent before directory/state inventory | Only receipt-owned committed three-file workset may retire; wrong receipt, unrelated mutation or omitted comparison fail | Credited; no “all state unchanged” shortcut |
| CDC lease/restart/checkpoint | Actual SIGKILL leaves durable nonterminal lease; live successor refused unchanged; actual~1m55s expiry then new approval/resume and target1→2 | Before/after lease, files/state/target are retained; no TTL manipulation, no customer database, no whole-system power-loss claim | Actual restart/exclusion credited; backward LSN invalidates advance oracle only |
| Compressed selected metadata | Valid deterministic decoded payload and expected semantic member set | Actual readable malformed/truncated/CRC/trailing/second-member/oversized/cancel fixtures reach decode frontier; no fallback/partial output | Correctness/integrity credited; size-ceiling removal is independent |
| Fresh command projection | Exact first-match lookup, fresh returned values and existing cap rules | Duplicate/reorder/missing/nonREST/output mutation/concurrent calls and same-name new bundle | Source-equivalent optimization credited; original timeout is not fabricated correctness RED |
| Seven-mode help | Independently specified complete sections/meanings, actual CLI help | Remove each section, replace its semantics, duplicate header; healthy alternate casing remains valid | CR-171-02 closed |

## Evidence identity, source applicability and limits

All 161 terminal receipt/command/raw-output byte/hash pins were independently checked against final191 index. Recomputed run/pass events, output lengths and terminal exits matched, including 37 failed terminal attempts. The original interrupted binary 188 command/output pins matched; no terminal receipt or exit was invented. Original source/prompt/fixture/generation and witness records retain their authorship. Raw failure, compile/setup failure, fixture correction, limited probe and withdrawn interpretation remain distinct.

The complete member set was independently reconstructed from the actual current cli_surface.json inputs, not trusted from aggregate reports. Its sorted newline-delimited identity SHA 256 is 9f8113fe445fd52b17423f36fb849e5f7234ac286891276b3c1b823ad9f15d9f. In-process188 has exactly 13,856 run/pass leaves. Original binary 188 has 8,832 started and 6,997 completed leaves, with 1,835 started without pass;21 terminal recovery shards execute8,165 completed leaves, including 1,306 explicit overlaps. The union is exactly 13,856 unique members;15,162 observed passes are not15,162 distinct commands. The retained binary's SHA 256 is 024c4eaa03a58f3620769a97c1fc17dbd9d9bed8f2affd15ecb047087e537504. The recovery reuses that exact binary and compatible source/configuration; it does not convert the interrupted parent into a successful full run.

Independent source applicability audit covered215,846 captured references across18 relevant receipts, hashing11,334 unique physical snapshots with no missing/hash mismatch. Full in-process/build/recovery pin sets contain12,035 captured inputs; the sole final captured-input delta is the authoring-only foundation demand report. Later hook/planning/manual changes are separately bound rather than silently equated to tested snapshots. All 324 final changed-path pins and 49 bound references establish the final candidate independently of earlier command HEAD values.

For the four full physical/mode reference receipts, changed captured inputs were individually classified. Warehouse/transport/native PostgreSQL/authentication identity/receipt/lease/App execution and module inputs relevant to those flows are unchanged. The cli.go control-enum replacement preserves the old16 PM-control names; engine's first-ID map preserves exact old lookup semantics; source_visibility's explicit ignored Close result follows completed integrity-through-EOF. GitHub/GitLab alias changes are outside the selected reference transport route. Authoring/proof/Atlas/website changes do not execute there. The added TestTransportWarehouseFiles173 is independently covered by later fixture-oracles-race 179 and is not retroactively counted in the earlier25-event oracle run. These are specific continuity reasons, not a byte-identical-current-tree replay claim.

The API physical receipt173-02 reaches GET source(page size100), POST destination, GET readback(page size100), DELETE204 and DELETE404; actual physical warehouse record/hash and zero-residue assertions were read. The warehouse roundtrip173-01 reaches actual query/flow records1, comments2, persisted checkpoint/receipt and replay/unapproved/auth refusal; auth refusal leaves checkpoint unchanged. Both raw runs execute tests rather than replay a successful test cache, despite their original argv lacking -count=1; later continuity runs explicitly use -count=1. Their exact argv are retained. The original fixture setup and destructive-confirmation failures are not mislabeled as successful physical flow.

Database physical173-03 is an actual -count=1 databaseintegration run,189.81s, not a selected/no-start harness witness. Mode173-01 executes three database tests in74.72s: dedupe history, full overwrite [1,2,3]→[2,3,4] with changed2/removed1, and unchanged upsert reporting0 read/load. Physical restart/schema-refusal/replay assertions and receipt-owned retirement were inspected from setup through return/cleanup. Early harness capacity/configuration failures and wrong test retirement/projection assumptions remain historical setup/oracle failures. No new live database test was run by this reviewer.

Exact normal/race accounting was compared rather than inferred from totals:

| Selection | Actual membership comparison |
|---|---|
| Cumulative focused173 |152 events each, exact same normal/race set, all pass |
| Source CLI174 |49 each, exact same set, all pass |
| Carry CLI |13 normal /18 race are different selections. Normal-only contains ETLHelpListsAllSyncModes and HelpContractFalsifiers171 family; race-only contains EveryModeHelpContractFalsified173 plus seven mode children. Separate normal carry-controls173 covers all seven; no false13=18 claim |
| Carry engine/App |36 normal /37 race; race additionally selects foundation_gap/bound |
| Projection188 | Normal has 18 correctness passes plus a benchmark event; race has 19 tests including deferred-foundation-gap. A benchmark lacking a test-pass event is not a missing test |
| Remaining continuity | Cold/cache race 14; actual CLI89; engine+runner2448; PostgreSQL carry2304; checkpoint/physical/retirement oracle race 25 and additive fixture-oracle race 23 retain their separate identities and overlaps |
| CI175 observed failures |148 passing local events cover exactly 25 observed original top-level groups; they are neither the complete truncated historical log nor complete final CI |

All 553 generated manifest entries were independently decoded from their actual Go literals. Ten cohort source-capability payloads have exact declared decoded lengths/digests and unchanged source/seven-lane member identities. Current cohort totals remain4343 source keys and 30401 cells. GitHub and GitLab execution JSON comparisons against base show respectively 7 and 15 changes solely to flag Name; all other fields and their1,612/1,351 command memberships are unchanged. The 232 immutable fixture payloads match content-addressed filenames; restoring them in t.TempDir enables historical proof tests without making stale current proofs valid or runtime-authoritative.

The terminal local evidence includes applicable scoped vet/new-only lint, patched-dependency scan, tidy/gofmt, source-lanes/demands checks, agent contract,553-definition validation, canonical rendering/publication, boundary scan323files/553connectors/0findings with six existing exceptions, final generated docs191-03, smoke and actionlint. Source-demand191-01 and docs191-02 failed before their respective authoring/generated refreshes; those failures remain. Docs191-03 then passes with exact four generated alias-only files; website188 input continuity remains applicable. Neither a completed439-second boundary scanner nor selected package tests establish global go test ./..., global lint or final GitHub Verify. The original Verify34214648721/job102023437321 log is truncated mid-group. Complete final CI is specifically unproven by this handoff.

## Reviewer-designed probe record

All probes were designed substantively by this reviewer and executed mechanically by the canonical owner only inside the verified private clone. Each journal subdirectory preserves exact source pins, command argv/cwd/environment, raw JSONL output, terminal receipt and hash-checked added-file removal. Production and existing tests were never edited. Writable HOME/TMPDIR/GOCACHE/GOTMPDIR locations are journal-owned; the existing Go module cache is explicitly referenced, with network module lookup disabled. Installed exact Go1.26.6, GOTOOLCHAIN=local and explicit GOROOT avoid implicit toolchain selection. No provider/customer/DB connection is used by these probes.

| Probe | Result and interpretation |
|---|---|
|173-P01 | Initial renderer/parser attempts fail before compilation because isolated HOME plus GOSUMDB=off blocks automatic toolchain checksum verification: infrastructure only. Unchanged pinned sources rerun using installed local toolchain: renderer exit0,29.1584s; actual parser exit1,10.0264s. Source hashes960c734f251f6e2d2fd0aee63c07cafee25f0d267e6a58007b1d5618436a4050 and 4acd8d6be9bd0580134d5e9f7bbe703df3a53375a89f92dfacb8241b5dca4cfc. Corrected raw hashes8c826763ae94531d0a0aafdb0d60c1068b546f1d33d8daa48766439f307fcf37 /727180d61f70b1bf1f0e64f5c0d69176a510dd457b518a90fc40c0c710dffa6f. Supports CR-173-01 |
|173-P02 | Existing eight-falsifier oracle passes, forward/unchanged controls pass, two backward assertions fail. Exit1,5.539484s. Source SHA 2a01754cf4a768a4c5908247c061c37f98517c14cb6022ed5fa951583a9cbda1; raw 17,660bytes SHA b66680595d4a4c0afe56654ffad9ed66b095c70ed8f397ab486374fc0eca82af. Supports WR-173-01 only, no product rollback claim |
|173-P03 | All five source-derived literal/repeated/adjacent/hyphen/underscore exact render/binding/flag and deterministic-check controls pass; same-count wrong path target is rejected. The omitted subtest fails because it removes flags from authoritative source before canonicalization. That does not establish projector erasure: engine operation_parameters.go:139–146 explicitly permits configuration path values. The defect inference is withdrawn, retaining source/raw unchanged. Exit1,5.869739s; source SHA 12ad4c09acdf3faf5fab555e7214177f08b695f142a0dda5b04d8cb78dd97acd; raw 7,874bytes SHA 5cedaccd9ad8c563997093e66a6ff46536ac05302535bb0c7c1d991b5ad82d19 |
|173-P04 | All five real-render source/path controls, healthy canonical control and both projected corruptions pass. Each mutation starts from a fresh descriptor, asserts the sole expected widget-id→path.widget_id required binding, preserves AuthoredCommand, and changes only the projected command/graph. Admission rejects omitted flag and same-count path.other with the exact authored-binding invariant. Exit 0, 5.808293s; source 6,913 bytes SHA 256 a453bb3fed6251da075381718cbd82b0f6baadc2104827edaabdbd7c8c8929e0; raw 9,511 bytes SHA 256 de29666c958587a3da9261b7a3c84e4ffeefcab5d49a597d43cd6bba3052d180. P03's failed inference is disconfirmed rather than counted |

No rejected candidate interpretation is counted as a finding. P03's five legal shapes demonstrate exact emitted binding and preflight; P01 separately reaches the actual hand parser. Existing actual built-parser generation tests and complete current registry leaves provide public consumer coverage. These independent facts are not collapsed into a claim that P03 itself calls the private CLI parser.

## Ten lenses

| Lens | Disposition with specific evidence |
|---|---|
| Architecture/data flow | **Complete.** Single schema4 authoring→execution JSON→existing manifest/registry→hand parser/App/commandrunner→closed executor preserved; map above traces changed seams. No connector-name dispatch, second executor reader/registry or generic write escape added |
| Happy/bad/edge | **Complete within accepted scope; findings recorded.** Legal distinct/repeated/adjacent source targets, absent/empty/whitespace/braces/duplicates/unknown/malformed aliases and exact required flags, semantic sibling diagnostics, config/env, gzip truncation/caps and mode refusals traced. CR-173-01 exposes a concrete grammar edge |
| State/concurrency | **Complete runtime review; evidence limitation WR-173-01.** Fresh projection/copy/concurrency, publication lease/admission boundary and reference state ownership/reopen/retirement traced. Actual killed-process lease persists; precise expiry/refusal/resume witnessed. No total power-loss or distributed-atomicity claim |
| Secret taint | **Complete for affected paths.** Actual public text/JSON diagnostic sentinel checks retain private cause; alias config/env/provider values keep approved ownership and output redaction boundaries. No secrets requested, read out or stored by review. No new private/provider log channel |
| Retry/rate/resume/idempotency | **Complete affected continuity; WR-173-01 limits one oracle.** Runtime physical-send accounting/retry/idempotency/strict ambiguity/ack/lease implementation unchanged. Actual replay/refusal/lease expiry/new approval and restart records support the selected guarantees; no exactly-once or provider-wide-limit inference |
| Output integrity | **Complete.** Independent physical row/hash/member comparisons, gzip EOF/CRC/single member/size/hash checks, full diagnostic text/JSON, generated alias/docs equality and archive integrity retain boundaries. Existing unchanged short-write contracts are reused; no changed writer introduces a new unchecked public write |
| Declaration reachability/closed surface | **Complete with CR-173-01/02.** Exact 13,856 current membership accepted at declared boundaries, unknown paths refused; proof/Atlas metadata does not gate ordinary runtime. Admission allows option-like aliases and Notion loses optional value, both retained |
| CLI/App parity | **Complete with findings.** Actual hand parser/App/built binary, flag wire/lifecycle/help, persisted approval and generated docs/website inspected; resolver mocks were not used as public execution proof |
| Provider semantics | **Complete for retained declarations/hermetic fixtures; specifically blocked beyond them.** Exact source endpoints/parameters and fixture effects reviewed; no provider-live or customer DB certification is available or authorized. Historical Bitbucket50/28 member mapping remains unavailable; CP19 current 297 chain retained |
| Tests/evidence | **Complete audit; limits explicit.** All 161 terminal and original interrupted receipts/pins, exact normal/race identities, independent oracle falsification, current-source impacts, known-red and withdrawals retained. WR-173-01 and historical/final CI limitations prevent an unqualified green statement |

## Five safeguards

| Safeguard | Application and result |
|---|---|
| Complete contract-to-evidence table | All ten CP16 rows and every original carry/baseline row are individually dispositioned above; runtime/public owner and observable frontier are explicit |
| Helper internals, first side effect through return | Admission precedes publication; config/env preparation precedes App; diagnostic rejection precedes provider send; decode reads EOF/CRC before success; approval precedes grant/apply; WAL/materialization/receipt and allowed work retirement/reopen were traced with cleanup and real healthy setup |
| Compound causes and actual public consumers | Complete producer cause survives wrapped/joined errors while useful safe rule coordinates reach JSON/text; private sentinels stay private. Old166/171 producer/public obligations remain distinct from internal errors.Is checks |
| Independently retained members/bytes/state/history | Current execution JSON independently defines13,856 members; exact 297 source joins, source/fixture/receipt hashes, seeded1001 rows, before-state/file inventories and receipt-owned permitted retirement are compared without equal-count substitution |
| Falsified critical oracles at actual phases | Parser option-token failure and backward-LSN acceptance were independently reached. Legal controls, same-count wrong target, projected omission, physical wrong/omitted rows/files/hashes, wrong receipt, whole missing help section and malformed compressed readable payloads challenge the relevant oracles. P03's wrong source-mutation inference and infrastructure-only P01 attempts are preserved rather than promoted to RED |

## Final custody and return

Final pre-report audit at 2026-09-08T17:23:07.915586Z rechecked all 373 bound reference/changed-file pins (49+324): no mismatch. All 15,803 tracked files remain byte-identical between frozen and private copies; both remain at the bound head/tree. Private status was completely clean after P04 cleanup, detached with its own .git. Frozen tracked state is clean. The bound protected untracked inventory remains exactly 19,519 paths, with no additions or removals. This is an inventory comparison, not a claim of before/after byte hashes for every protected untracked file: the binding did not supply those full byte pins. No reviewer action wrote those files or ran a probe/build/generator in the frozen worktree.

All seven reviewer-probe invocations are terminal: P01's two infrastructure failures and two corrected-toolchain runs, plus P02/P03/P04. All added probe test files were hash-checked and removed; original probe source/command/output/receipt/cleanup bytes remain in the journal. The reviewer subtree inventory contains only this running reviewer and no child. All reviewer read/audit tool calls are terminal; artifact creation and its read-only verification are the final operations. The only permitted private-copy difference at return is this phase REVIEW.md. The canonical owner must record the actual native completion after the final return; this report does not prematurely assert that the reviewer native process has ended.

The final seal must bind this report's exact bytes together with the existing immutable candidate binding, all three individual finding records including the WR citation revision, RUN history including P03's withdrawal, all P01–P04 source/command/raw/receipt/cleanup files, the initial/final source state and actual native/profile/parentage. Mechanical publication is authorized only for exact reviewer-authored substance.

The complete final actionable set is **CR-173-01, CR-173-02 (CP16-FLAG-NOTION-179), WR-173-01: 3 distinct required corrections, 2 BLOCKER and 1 WARNING**. CR-171-01 and CR-171-02 close here with their original causal aliases retained. Historical membership/CI limits and current fixture-only certification limits remain part of this result. Canonical owner must publish this entire report and findings without dropping warnings, seal exact report/source/probe/receipt bytes and record actual native terminal completion after this reviewer returns. There is no active reviewer child or additional review stage, and this reviewer supplies no integration action or approval.

