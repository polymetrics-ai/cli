# Foundation 0.3.0 release-candidate r1 review record

source_sha: ad8bff65958a20924aff43d9aea3ea42301e9703

This is final code-and-generated implementation SHA **I3**. Its schema-3 evidence closure is deliberately the next commit: an evidence file cannot truthfully name the Git object that contains itself.

## Intake and composition crosswalk

| Item | Disposition and current witness |
| --- | --- |
| Original 38 blocker categories | Historical and repaired. The authoritative audits trace each category to committed repairs; none was reimplemented here. |
| Core / reverse / public inputs | Core `041d2ec…` was non-force published after predecessor-ancestry proof; reverse `487ec14…` is preserved by `50e90fa…`; public `7fdef00…` is preserved by `e18f437…` and was non-force published after proof. Each remote SHA was read back with `git ls-remote`. |
| Already-inherited components | Source import `19a32bd…`, closed operation `3c768ca…`, structured body `66b4c3c…`, and reverse destination `d814875…` are ancestors; none was remerged or cherry-picked. |
| RC-05 current-code witness | The earlier exact gate found a stale certification-sweep expectation and a public-predecessor source-lock digest; the narrow test-only correction added no runtime authority. |
| RC-06 Recurly confirmation | At old closure `7c3d856…`, unchanged `TestBinaryDownloadCommandsExecute` failed: `invoice pdf get` rejected `--invoice-id` and `export files get` rejected `--export-date`. Direct witness `d3bf5da0` had removed five still-required path flags: `account-id`, `invoice-id`, two `subscription-id` bindings, and `export-date`. |
| RC-06 minimal repair | Each affected Recurly operation declares its exact required `path` parameter in `operations.json`; `connectorgen surface-sync` generated the five flags and derived manual/skill output. No runtime branch, mapping expansion, provider call, or test weakening occurred. |
| RC-07 GitHub generated direct reads | At `596c90c…`, `TestDirectReadCandidatesForGitHub` reproduced 43 candidates instead of 120 (23 manual plus 97 generated). The narrow projection repair recognizes only existing `certification.json` generated-read cohort members; it restores 77 source-bound candidates plus the existing `codespace list` alias on the same endpoint, leaves unrelated partial commands untouched, and does not alter any expected count. |
| RC-07 behavioral witness | Before the hyphenated path fix, a freshly built `pm connectors certify github --direct-read-only` fixture run failed `enterprise-team-organizations get-assignments` before fixture I/O with unresolved `{enterprise-team}`. After repair, `TestPMBinaryExecutesGitHubDirectReadCandidatesAgainstFixture` executes all 97 generated stages through the fresh binary, requires a resolved authenticated HTTP request for each stage, and requires the generated object-or-array output assertion to pass. |
| RC-08 GitHub released read surface | At `a750e2c…`, the real inventory regression reported 854 covered REST endpoints where 1,224 were required: 1,225 endpoints still existed but 371 were blocked. The 370 added blocked GET routes shared the source-projection reason. `748865f1b` introduced the conservative downgrade; the count later oscillated as component merges interleaved. |
| RC-08 minimal repair and proof | The projection now restores each existing, field-complete, declaration-owned read route rather than only the certification cohort, and promotes an already-declared required source path flag (`repo read-file --path`) to required. No endpoint, route, ledger entry, or generated JSON was hand-authored. A fresh `pm` executes all 633 unique API-surface read targets against a fixture; every invocation emits its declared resolved request and meets its JSON/text/status-only/repository-content/binary output assertion. |
| RC-09 CodeQL confirmation | At evidence E `25b2f844…`, CodeQL directly reported three new PR alerts: high allocation-size overflow for `len(values)*4` in write redaction, and two useless assignments at parked-resume reconciliation and final HTTP non-2xx handling. All three findings have a direct current-code witness. |
| RC-09 minimal repair | The redaction literal list no longer derives capacity from input arithmetic; parked-resume reconciliation still runs for its durable side effect, and final non-2xx handling still returns its typed status error directly. No suppression, baseline, annotation, JavaScript, or `ledger.go` change was made. Focused redaction/reconciliation/terminal-receipt behavior, fresh-binary 633-command GitHub proof, full `cmd/connectorgen` (143.231s), and `go vet ./...` pass. Fresh CodeQL is pending externally. |
| RC-10 GitHub verdict correction | Verify at E2 failed two hard-coded assertions that fifteen already-implemented GitHub reads must be `partial`: eleven list commands plus `issue status`, `pr checks`, `ruleset check`, and `search prs`. The captain's artifact comparison proves the eleven lists are unchanged at `v0.2.1`, `origin/main`, and I3; they use valid API-surface bindings even without an `operation`/`stream`/`write` field. No declaration, generated surface, or runtime code changed. |
| RC-10 behavioral adjudication | `TestPMBinaryExecutesGitHubDisputedPartialVerdictsAgainstFixture` builds and runs a fresh `pm` for every disputed command. Each invocation initializes an isolated project with a fixture credential, reaches exactly one authenticated GET, matches the resolved request path to its declared API-surface template, and validates the emitted direct-read envelope and declared JSON output. All 15 pass (39.16s). That direct execution evidence—not inventory count—authorizes correcting the stale test expectations. |
| Candidate-output decoder | The reported non-JSON failure did not reproduce locally, but `runTransportPM` directly used `CombinedOutput` before JSON decoding. The candidate test now isolates stdout for JSON and reports independently redacted stderr. The full 97-stage fresh-binary candidate fixture proof passes after this change (70.30s); it remains behavioral fixture proof, not live certification. |
| CI website-data parity | Fresh CI directly failed because generator-owned website data was stale. A local `website` generator run reproduced it and changed only four generated website artifacts; its data now includes the required Recurly flags and current source-projection statuses. No provider/runtime declaration was altered. |

## Conflict-invariant preservation

| Seam | Preserved invariant | Evidence |
| --- | --- | --- |
| Core + reverse: App, Postgres transport, sync transport | Core empty-publication, authentication, deadline, and acknowledgement fences coexist with reverse action-owned receipt/readback binding. | Exact-I App, Postgres race, synctransport, commandrunner, and connector cohorts pass. |
| Core + public: registry, engine direct-read/headers, SQS, output | Exact configured-secret masking remains without corrupting ordinary keys, identifiers, raw receipt bytes, or declaration-owned parameter authority. | Exact-I engine, connector, commandrunner, SQS, App, and CLI cohorts pass. |

## FND-B09 and FND-W01

- **FND-B09 — reproduced and closed:** inherited schema-2 evidence failed strict decode because `component_inputs` was an object rather than the required typed list. Old closures `7c3d856…`, `133afe481…`, `596c90c…`, `25b2f844…`, and `8ed5ab93…` are superseded after the confirmed Recurly correction, CI-proven website-data drift, GitHub reachability regression, RC-09 CodeQL repair, and RC-10 behavioral verdict correction. This schema-3 closure binds I3, five typed component identities, current subject fingerprint, and exact artifact digests.
- **FND-W01 — deferred / non-reproducing:** current `sourceimport` flow and hermetic `TestSourceImport.*` evidence were inspected. No deterministic externally observable publication-order defect was established, so no speculative publication rewrite was made.

## Behavioral proof, provider proof, and certification claim

| Command or surface | Local behavioral / fixture proof | Current real-provider proof | Certification claim |
| --- | --- | --- | --- |
| Foundation seams | Exact-I App, CLI, engine, commandrunner, sync transport, SQS, Postgres-race, smoke, and output cohorts pass. | No approved non-secret credential reference was supplied, inferred, or probed. | Not a connector certification claim. |
| Recurly `invoice pdf get`, `export files get` | Unchanged local-fixture binary execution test passes; initialized-project invocations accept `--invoice-id`/`--export-date` and stop at `missing --credential`. The other three restored path flags are generator-derived from their required declarations. | No Recurly provider execution. | **Implemented but uncertified.** Only local behavioral/fixture proof. |
| GitHub generated direct reads | Fresh-binary fixture proof executes all 97 generated candidates through CLI parsing, commandrunner, engine, GitHub transport, declared auth, resolved path binding, and the declared object-or-array output assertion. Source import, surface sync, candidates, and sweep generation checks are current. | No current provider evidence; retained historical 422 is not new live proof. | **Implemented but uncertified.** This is behavioral fixture proof, not live certification. |
| GitHub released API read surface | Fresh-binary fixture proof executes 633 unique API-surface targets, a strict superset of the 370 restored endpoints. It asserts one resolved request per command, declared method, and declared JSON/text/status-only/repository-content/binary output behavior. | No approved GitHub credential reference was available or inferred. | **Implemented but uncertified.** The fixture proves executable local behavior only. |
| GitHub RC-10 disputed reads: `agent-task list`, `cache list`, `codespace list`, `gist list`, `gpg-key list`, `issue status`, `org list`, `pr checks`, `repo gitignore list`, `repo license list`, `ruleset check`, `search prs`, `secret list`, `ssh-key list`, `variable list` | Fresh-binary fixture proof executes every listed command. Each makes one authenticated GET with declared resolved path and a validated declared JSON result. | No approved GitHub credential reference was available or inferred. | **Implemented but uncertified.** All 15 are locally behaviorally proven; none is claimed live certified. |

The certification matrix reports `connectors=3 capability_complete=0 certified=0`. No command is certified from source mapping, fixture proof, or historical provider output.

## Exact-I commands and results

| Command | Result |
| --- | --- |
| Recurly package before repair | RED at `7c3d856…`: unchanged binary execution test rejected the two supplied path flags. |
| `go test -v -count=1 ./internal/connectors/defs/recurly` after repair | PASS; unchanged binary execution witness passes. |
| `connectorgen surface-sync [--check]`; Recurly validation | PASS; generator filled five required flags, then reported zero drift and zero Recurly findings. |
| Initialized-project `pm recurly invoice pdf get --invoice-id …` and `pm recurly export files get --export-date …` | PASS for CLI reachability: each parsed the flag and reached only `missing --credential`; no provider I/O. |
| `go test -timeout 20m -count=1 -v ./internal/cli -run '^TestPMBinaryExecutesGitHubDirectReadCandidatesAgainstFixture$'` | PASS (71.24s): fresh `pm` executes all 97 generated direct-read stages against a local fixture; each stage reports `ConnectorCommandDirectRead`, emits an authenticated resolved request, and passes the candidate's object-or-array output assertion. |
| `go test -timeout 20m -count=1 -v ./internal/cli -run '^TestPMBinaryExecutesGitHubReleasedReadSurfaceAgainstFixture$'` | PASS (190.44s): fresh `pm` executes all 633 released API-surface read targets with 633 fixture requests, declared method/path checks, and JSON/text/status-only/repository-content/binary output assertions. |
| RC-09 focused behavioral suites | PASS: write redaction preserves overlapping-mask behavior; parked rate-limit post-commit claim reconciliation preserves the durable retry; terminal requester retry retains its provider receipt and typed error. |
| `go test -timeout 20m -count=1 -v ./internal/cli -run '^TestPMBinaryExecutesGitHubDisputedPartialVerdictsAgainstFixture$'` | PASS (39.16s): a fresh `pm` executes all fifteen disputed commands against the isolated authenticated fixture; each request matches the declared GET path template and its direct-read result validates. |
| `go test -timeout 20m -count=1 -v ./internal/cli -run '^TestPMBinaryExecutesGitHubDirectReadCandidatesAgainstFixture$'` after decoder separation | PASS (70.30s): stdout JSON decodes independently of diagnostics; 97 generated direct-read stages execute through the fresh binary. |
| `go test -timeout 20m -count=1 ./internal/connectors/commandrunner` | PASS (21.816s). |
| `go test -timeout 20m -count=1 ./internal/cli` | PASS (1061.223s). |
| `go test -timeout 20m -count=1 ./cmd/connectorgen` after cache repair | PASS (143.786s) from I3. |
| `go vet ./...`; `go build ./cmd/pm`; `go run ./cmd/connectorgen surface-sync --check` | PASS; surface sync reports 552 connectors and zero corrected fields. |
| GitHub `source-import --check`, `surface-sync --check`, `certification-candidates --check`, generated `certification-sweep`, and direct-read candidate/cohort tests | PASS: 120 total direct-read candidates (23 manual, 97 generated); unchanged expectations. |
| Exact GitHub release-surface parity at I | PASS: `v0.2.1` and I both have `endpoints=1225`, `blocked=1`, and `direct_read_candidates=120`; the sole blocked identity is `POST /user/{user_id}/projectsV2/{project_number}/drafts`, duplicate of `POST /graphql (github.graphql.mutation.add-project-v2-draft-issue)`. |
| `go test -timeout 20m ./cmd/connectorgen -count=1` | PASS (151.700s). |
| `go test -timeout 20m ./internal/app -count=1` | PASS (264.726s). |
| `go test -timeout 20m ./internal/cli -count=1` | PASS (822.864s). |
| Engine/connectors/commandrunner/synctransport/SQS cohort | PASS. |
| `go test -race -timeout 20m ./internal/connectors/native/postgres -count=1` | PASS. |
| Formatting, diff, tidy, vet, and build | PASS. |
| `make tidy-check lint docs-check-no-build smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check` | PASS. |
| Certification subject/matrix/candidates/sweep | PASS; matrix remains `certified=0`; no live certification performed. |
| `cd website && npm run gen:website-data`; `make docs-check-no-build` | PASS and idempotent; CI-required website data projection is current. |

## Merge recommendation

The local Foundation implementation and the targeted GitHub candidate and RC-10 verdict gates are green. Recurly and GitHub reads are implemented but explicitly not live certified. Do not merge this PR from this record: remaining gates are the strict evidence gate on the replacement closure, fresh automated review, and GitHub CI. PR #4314 remains unmerged against `fm/cli-current-foundations-main-integration-r1`.
