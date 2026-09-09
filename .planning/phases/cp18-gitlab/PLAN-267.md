# CP18 GitLab repair plan — 267

## Task Delivery Header

- Issue: Refs #4384, #4394, and #4413 — GitLab source, generated-artifact, and runtime-proof chain.
- Base branch: `fm/cli-top100-declaration-batch-r1`.
- Merges into: `fm/cli-top100-declaration-batch-r1` → `main` (only by a separate captain decision).
- Delivery: GitLab-only, coherent, verified increments are committed and normally integrated/pushed to the existing R1 / PR #4294; completion additionally requires independent Terra/xhigh review and honest remaining-gap dispositions.
- Working branch: `fm/cli-batch1-cp18-gitlab`.
- Task: Complete the CP18 source-to-runtime proof chain without a second runtime reader, connector-specific bypass, provider-live access, or unrelated connector work.
- Verification: Existing retained REDs plus fresh targeted `-count=1` production-path tests; admitted `lock-render` / `--check`; targeted compiler, engine, commandrunner, App, warehouse, CLI, and generated-output checks; an immutable candidate review.

## 267 baseline and workflow

- Implementation runs inline because the canonical single-worker contract forbids role spawning in this runtime. This is the documented manual fallback for `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`; independent review remains Firstmate-routed Terra/xhigh work.
- `scripts/gsd doctor`, all five `scripts/gsd sources` commands, all five generated prompts, and `go run ./cmd/agentcontractgen check` were run before production edits. The adapter reports one unrelated missing historical issue prompt; the requested commands resolve and the contract check passes.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.
- Source-lock and Atlas rules were read. The Atlas presently establishes reuse of `warehouse.stage-etl.v1` and `warehouse.reverse-etl.v1`; any further shared change requires the named owner/symbol/proof classification before it is proposed.
- Baseline is `84f6d456cb165a726cc4d7489d8c9f4c1fe8822c` / tree `d3162e16235119660545feeb4cdccd6dfd87a938`. Its preserved untracked state is 437 files: lane proof test SHA-256 `4fdb83ec1abcb624e66320208647e2f3c285e052212ef1224c9a1188d0917180`; local authoring tree SHA-256 `896d5f435cfb54a1735ea20ed1e414db36cb8367c3d142a6a2d6ea7b68a26a71`.
- Latest authorised integration worktree is clean at `da0a41267b1808246dc33d37ab2d7fd570d233c7`; it is revalidated immediately before any normal R1 publication.

## Obligation ledger

| Group | Current established fact | Planned Green evidence | Atlas classification |
| --- | --- | --- | --- |
| G1 | 1,344 retained failures: 200 `per_page`, 2 `max_results`, and 1 `limit` declaration joins are refused before transport; possible `scope` joins are masked. | Generator emits only reviewed input fields; production CLI/App requests carry exact method/path/query/auth; undeclared inputs remain pre-I/O refusals. | Reuse existing source compiler/input validation; classify any gap before shared change. |
| G2 | Four source numeric fields lose inclusive bounds; 8 direct failures plus 6 polluted snapshot failures. | Below/above reject before plan or HTTP; both endpoints and interior values reach the declared frontier in isolated state. | Reuse existing constraints if declarations render them; otherwise constrained compiler extension. |
| G3 | 82 DELETE operations yield 164 failures: completed-send disconnects produce five sends where tests expect one; current `idempotent=true` allows retry. | Per-policy source decision, direct and saved send counts, one-time approval consumption, receipt/ack and no-blind-replay assertions. | Reuse typed approval/reverse ETL; classify retry policy from source rather than disabling retries globally. |
| G4 | Six operations / 14 retained cases serialize source numeric JSON arms as strings. | Exact number, string, bigint, and supported null wire bytes agree across direct, plan/fingerprint, and saved paths. | Classify source-generated typed-arm seam before shared code. |
| G5 | Group-member valid source semantics are absent from fixture; date prose/schema conflict is recorded. | Independent fake-provider assertions reject wrong role/value/type/count and premature durable state. | Connector fixtures only unless an existing schema constraint is insufficient. |
| G6 | 37 operations, 67 variants, and 49 parameters need full structured-query source matrix accounting. | Exact style/explode, null/empty/omit, direct/saved and continuation identities; no inferred bracket/JSON dialect. | Existing structured-query foundation first; constrained extension only for a demonstrated unexpressible source contract. |
| G7 | HEAD reads, bodyless remote-mirror POST, binary download/upload, and four Conan POST reads remain outside the stopped batch. | Production-path command discovery, exact headers/body/media/bytes/auth and bounded size; correct read/write intent and approval boundary. | Existing REST/binary/typed encoder seams first; no generic escape. |
| G8 | MLflow, source projection/classification, lane applicability, 553 unstarted writes, and one interrupted leaf remain incomplete. | Mode-specific MLflow preservation/no-materialization proof, source class 1752+2 joins, actual seven-lane dispositions, and explicit selected-test membership. | `warehouse.stage-etl.v1` / `warehouse.reverse-etl.v1` reused; shared compiler/proof-catalog changes only after source-bound seam review. |

## Execution order

1. Reconcile this frozen baseline with R1 only at the stopped batch boundary; preserve original RED records, helper-custody 259, and all untracked material.
2. Establish narrowly selected, source-bound REDs for G1/G2 first and apply declaration-only repairs where the current foundations suffice.
3. Run the matching fresh Green tests, then grouped affected checks. Do not rerun an unchanged broad discovery sweep.
4. Repeat for G3–G8, classifying shared needs as reuse / constrained extension / actual gap with exact Atlas owner and proof test references.
5. At each all-green coherent increment, commit; transfer only complete dependencies into the clean integration worktree, validate there, push normally, and verify its remote SHA.
6. Freeze the final GitLab candidate, request Firstmate's independent Terra/xhigh review, disposition findings, and run the prescribed GSD gap loop if verification finds gaps.

## Firstmate277 G3 admission-closure increment

The preserved G3 candidate is `deleteApiV4AdminCiVariablesKey`, for its direct-write and reverse-ETL commands.  The previous source-lane result is retained diagnostic evidence, not an admission result: it reports a source/canonical/execution coordinate mismatch plus the three open proof-contract deficits below.  This increment closes those deficits through the canonical source-lock producer and production runtime; it does not start another broad G3 batch or G4 work.

| Track | Red first | Green acceptance | Foundation classification |
| --- | --- | --- | --- |
| Coordinate producer | A focused source-lane check observes divergent canonical and execution generation identities after a rendered GitLab source lock. | The isolated canonical producer derives matching identities; the source-lane check accepts the candidate's exact canonical/action coordinates without hand-edited output. | Reuse `authoring.source-lock-vnext.v1` producer seam; constrained compiler repair only if the canonical producer, not the annotation, owns the mismatch. |
| Input bound provenance | The retained source declares `key` as string without `maxLength`, while the request schema adds `maxLength:8192`, yielding `target_schema_unverified`. | Provider source fact and existing PM transport/safety policy remain separate and observable; the declared policy bound is accepted only through a reusable, typed compatibility rule with a narrow negative control. | Constrained extension of `authoring.source-lock-vnext.v1` schema-projection compatibility, if existing policy is insufficient. |
| Typed success response | The source declares a 200 JSON `APIEntitiesCiVariable`; the write action exposes only status/body capture, yielding `target_response_contract_unverified`. | The typed action declaration drives production direct-write and saved-write consumers; success is decoded/validated and source-relevant error/malformed responses are not represented as successful typed parity. | Constrained extension through the typed write-action response seam; no status-only shortcut. |
| Review catalog | The normal source-lane producer receives an empty review catalog, so a current executable proof cannot be admitted. | A canonical external catalog reader accepts a valid test-scoped review and rejects missing, stale, and coordinate-mismatched reviews. Production contains no self-authored approval; the independent reviewer supplies the frozen-candidate review file. | Reuse `authoring.source-lock-vnext.v1` proof-catalog seam with an external-input adapter. |

The completion artifact will contain two precise proof proposals, one for direct write and one for reverse ETL. Each proposal binds the target action, selected test leaf, portable raw receipt, retained source identity, changed implementation inputs, coverage, and stated limitations. The independent reviewer output path and strict document format are part of the committed candidate so Firstmate can route an immutable review without inferring a design.

## Firstmate282 source-lane correction

The post-transfer `source-lanes` run preserved the intended source → canonical →
execution identity for `deleteApiV4AdminCiVariablesKey`; it no longer reports
either generation mismatch. Its two remaining reference errors occur because
the authoring-only checker (1) classifies JSON Schema's `$schema` document
meta-keyword as an unknown instance-use edge and (2) rechecks a request-side
path mapping independently for the companion response-schema binding to the
same typed write. The retained target has the correct `record.key` route
coordinate and exact source-derived request/response schemas. `proof_unavailable`
remains an honest, separate review/proof deficit.

This is a constrained extension of Atlas
`authoring.source-lock-vnext.v1`, specifically
`sourceLaneCheckedProjectionNode`, `sourceLaneCheckPresent`,
`sourceLaneRouteContract`, and `sourceLaneProjectionContract`. It reuses the
existing source-lanes projection seam: the change may recognize JSON Schema
document metadata and aggregate exact same-target route/parameter mappings
across request and response roles, but may not change runtime execution,
command availability, lane implementation, or proof admission. The red test
must show both failures against a typed write with a source response; its
negative control must still reject a missing request-side path mapping.

Firstmate283's report delta establishes the causal classification: the
`target_schema_unverified` result predates response-role work and is the
pre-existing JSON-Schema-document metadata gap. The parameter and path errors
first appear when the response binding is added; that binding correctly has no
second copy of the request-side route map, so the errors were fix-created
role-local checking behavior, not an annotation mistake. The focused red test
captures both. Its green result accepts the two role-specific bindings while
the missing-request-map control still refuses. The exact post-correction
source-lanes report is valid with zero errors; it retains only the two
independently reviewed-proof deficits and therefore does not promote either
lane to implemented.

## Guardrails

- No provider credentials, provider-live checks, whole-repository suite, new dependencies, source-lock runtime reader, generic HTTP/SQL/shell write surface, raw cursor, fake implemented lane, or hand-edited generated execution/index/proof output.
- CLI help/manual/website parity is tracked for each changed command surface. Source-only/generated-surface changes will explicitly record the relevant parity result; no unrelated website work begins.
- At or below the instructed effective 5% quota threshold, start no further development or test batch; preserve evidence, stop owned runners, publish only already verified increments, and write a pause receipt.

## Firstmate285 first-admission freeze

The only active closure is the `gitlab/primary/deleteApiV4AdminCiVariablesKey`
source operation for `direct_write` and `reverse_etl`; the recovered fifteen
unstarted G3 leaves remain untouched. Two author-owned proof records bind the
current request and typed-response write projections, portable four-leaf C2
receipt, selected provider-response leaf, and exact changed code/test/module
inputs. The reverse record additionally declares the warehouse-table →
one-record-limit → key mapping → write path and unchanged-table boundary.

The original proof validator rejected the canonical source-lane
`sha256:<digest>` generation coordinate despite emitting it itself. The focused
GitLab test captures that RED and preserves bare-digest and malformed-prefix
controls. Its constrained shared repair accepts either closed spelling but
does not normalize target identity; `sourceLaneTargetRefEqual` remains exact.
The resulting source-lane report is valid and intentionally leaves both rows
`mapped_unproven` until an independent reviewer supplies exact matching records
through `data/connector-canon/batch1-source-lane-proof-reviews.json`.
