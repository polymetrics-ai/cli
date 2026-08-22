# Foundation 0.3.0 release-candidate r1 review record

source_sha: 1d83dd9ab82dbdaf0f19f0f4e0f28446f1c95d91

This is final code-and-generated implementation SHA **I**. Its schema-3 evidence closure is deliberately the next commit: an evidence file cannot truthfully name the Git object that contains itself.

## Intake and composition crosswalk

| Item | Disposition and current witness |
| --- | --- |
| Original 38 blocker categories | Historical and repaired. The authoritative audits trace each category to committed repairs; none was reimplemented here. |
| Core / reverse / public inputs | Core `041d2ec…` was non-force published after predecessor-ancestry proof; reverse `487ec14…` is preserved by `50e90fa…`; public `7fdef00…` is preserved by `e18f437…` and was non-force published after proof. Each remote SHA was read back with `git ls-remote`. |
| Already-inherited components | Source import `19a32bd…`, closed operation `3c768ca…`, structured body `66b4c3c…`, and reverse destination `d814875…` are ancestors; none was remerged or cherry-picked. |
| RC-05 current-code witness | The earlier exact gate found a stale certification-sweep expectation and a public-predecessor source-lock digest; the narrow test-only correction added no runtime authority. |
| RC-06 Recurly confirmation | At old closure `7c3d856…`, unchanged `TestBinaryDownloadCommandsExecute` failed: `invoice pdf get` rejected `--invoice-id` and `export files get` rejected `--export-date`. Direct witness `d3bf5da0` had removed five still-required path flags: `account-id`, `invoice-id`, two `subscription-id` bindings, and `export-date`. |
| RC-06 minimal repair | Each affected Recurly operation declares its exact required `path` parameter in `operations.json`; `connectorgen surface-sync` generated the five flags and derived manual/skill output. No runtime branch, mapping expansion, provider call, or test weakening occurred. |

## Conflict-invariant preservation

| Seam | Preserved invariant | Evidence |
| --- | --- | --- |
| Core + reverse: App, Postgres transport, sync transport | Core empty-publication, authentication, deadline, and acknowledgement fences coexist with reverse action-owned receipt/readback binding. | Exact-I App, Postgres race, synctransport, commandrunner, and connector cohorts pass. |
| Core + public: registry, engine direct-read/headers, SQS, output | Exact configured-secret masking remains without corrupting ordinary keys, identifiers, raw receipt bytes, or declaration-owned parameter authority. | Exact-I engine, connector, commandrunner, SQS, App, and CLI cohorts pass. |

## FND-B09 and FND-W01

- **FND-B09 — reproduced and closed:** inherited schema-2 evidence failed strict decode because `component_inputs` was an object rather than the required typed list. Old closure `7c3d856…` is superseded after the confirmed Recurly correction. This schema-3 closure binds I, five typed component identities, current subject fingerprint, and exact artifact digests.
- **FND-W01 — deferred / non-reproducing:** current `sourceimport` flow and hermetic `TestSourceImport.*` evidence were inspected. No deterministic externally observable publication-order defect was established, so no speculative publication rewrite was made.

## Behavioral proof, provider proof, and certification claim

| Command or surface | Local behavioral / fixture proof | Current real-provider proof | Certification claim |
| --- | --- | --- | --- |
| Foundation seams | Exact-I App, CLI, engine, commandrunner, sync transport, SQS, Postgres-race, smoke, and output cohorts pass. | No approved non-secret credential reference was supplied, inferred, or probed. | Not a connector certification claim. |
| Recurly `invoice pdf get`, `export files get` | Unchanged local-fixture binary execution test passes; initialized-project invocations accept `--invoice-id`/`--export-date` and stop at `missing --credential`. The other three restored path flags are generator-derived from their required declarations. | No Recurly provider execution. | **Implemented but uncertified.** Only local behavioral/fixture proof. |
| GitHub source-mapped surface | Declaration validation, source/surface synchronization, candidate/sweep, and binary-subject checks pass. | No current provider evidence; retained historical 422 is not new live proof. | Implemented/partial as declared, **uncertified**. |

The certification matrix reports `connectors=3 capability_complete=0 certified=0`. No command is certified from source mapping, fixture proof, or historical provider output.

## Exact-I commands and results

| Command | Result |
| --- | --- |
| Recurly package before repair | RED at `7c3d856…`: unchanged binary execution test rejected the two supplied path flags. |
| `go test -v -count=1 ./internal/connectors/defs/recurly` after repair | PASS; unchanged binary execution witness passes. |
| `connectorgen surface-sync [--check]`; Recurly validation | PASS; generator filled five required flags, then reported zero drift and zero Recurly findings. |
| Initialized-project `pm recurly invoice pdf get --invoice-id …` and `pm recurly export files get --export-date …` | PASS for CLI reachability: each parsed the flag and reached only `missing --credential`; no provider I/O. |
| `go test -timeout 20m ./cmd/connectorgen -count=1` | PASS (151.700s). |
| `go test -timeout 20m ./internal/app -count=1` | PASS (264.726s). |
| `go test -timeout 20m ./internal/cli -count=1` | PASS (822.864s). |
| Engine/connectors/commandrunner/synctransport/SQS cohort | PASS. |
| `go test -race -timeout 20m ./internal/connectors/native/postgres -count=1` | PASS. |
| Formatting, diff, tidy, vet, and build | PASS. |
| `make tidy-check lint docs-check-no-build smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check` | PASS. |
| Certification subject/matrix/candidates/sweep | PASS; matrix remains `certified=0`; no live certification performed. |

## Merge recommendation

The local Foundation implementation and targeted exact-I release gates are green. Recurly is no longer a product blocker after the narrow generated-declaration repair, but is explicitly not certified. Do not merge this PR from this record: remaining gates are the strict evidence gate on the new closure, pushed-head/PR identity read-back, fresh automated review, and GitHub CI. PR #4314 remains unmerged against `fm/cli-current-foundations-main-integration-r1`.
