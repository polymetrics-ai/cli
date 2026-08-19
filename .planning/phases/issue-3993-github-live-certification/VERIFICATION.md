# Verification checklist — Issue #3993

Status: historical harness slice remains verified; fresh current-SHA recovery
is **needs-decision**, not a GitHub certification, as of 2026-08-11.

- [x] Deterministic harness tests show that the supplied `Polymetrics-Cert` boundary controls both emitted cases and classification.
- [x] Deterministic harness tests show a common barrier release; write cases require an independent read-back before any PM child can start.
- [x] Write validation rejects every direct or `--config` owner/repository override before the credential-inspection child can start.
- [x] The two artifact self-checks pass after regeneration.
- [x] A real `pm` binary builds locally without secrets.
- [x] The App installation credential is used without disclosure; the revoked fine-grained token is untouched.
- [x] The whole applicable surface has a complete/failed/unavailable tally and failures are grouped by actual cause and quota bucket.
- [x] No provider mutation was dispatched; final repository read-back and temporary-run cleanup prove no created-resource residue remains.
- [x] GitHub → Parquet warehouse → DuckDB inbound flow is independently proven.
- [x] The outbound workflow refusal is attributed to #3994/#3992, with no duplicate action-path implementation.
- [x] Targeted local checks and required non-full-suite verification gates pass.
- [x] Inline/manual GSD `verify-work` is recorded in `UAT.md`; the adapter has
  no roadmap phase for issue 3993 and the dispatch prohibits role spawning.
- [x] Inline/manual standard code review is recorded in `REVIEW.md`; manual
  Shepherd-compatible evidence in `SHEPHERD-COMPAT.md` explicitly is not an
  automatic #3995 approval.

## Commands run

```text
node --test scripts/tests/github-live-cases.test.mjs scripts/tests/github-live-proof-sweep.test.mjs scripts/tests/github-live-lab.test.mjs
node scripts/github-live-proof-sweep.mjs --self-test
node scripts/github-live-lab-manifest.mjs --check
node scripts/github-live-bootstrap-probes.mjs --check
node --test --test-name-pattern='rejects a write case that overrides' scripts/tests/github-live-proof-sweep.test.mjs
go build ./cmd/pm
go run ./cmd/connectorgen surface-sync --check
make connector-runtime-preflight
make connector-canon-check
go run ./cmd/agentcontractgen check
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
node --test --test-name-pattern='retires the legacy rate proof without current certification semantics|rejects an unsafe bootstrap manifest before inventory hashing|rejects unsafe supplied bootstrap inventories before drift validation' scripts/tests/github-live-rate-limit-proof.test.mjs scripts/tests/github-live-lab.test.mjs
node --test scripts/tests/github-live-cases.test.mjs scripts/tests/github-live-lab.test.mjs scripts/tests/github-live-proof-sweep.test.mjs scripts/tests/github-live-rate-limit-proof.test.mjs
```

The named synthetic command intentionally returned three RED failures before
the production edits. The final focused command passed all 50 tests without a
credential, provider command, or network test.

## Credentialed result and evidence handling

The App-authenticated fresh binary was exercised only against the immutable
`Polymetrics-Cert` boundary. The current canonical classifier records 1,521
commands: 182 attemptable, 1,339 blocked, and 169 / 470 direct reads. Its static
movement from the separately named historical terminal measurement is -483 / +483.
No provider command was run for that correction.

The historical full one-barrier release records 665 attempted operations: 0
proven, 665 terminal-bound failures, and 856 untestable rows. No HTTP status was
available for the failures. Rate snapshots immediately around that historical
sweep measured REST core 15,000 → 14,997 and GraphQL 5,000 → 5,000.

The App token itself was not printed or committed.  A single returned-record
control and a 16-way returned-record control establish that the App path worked;
the all-at-once 665-child result is a local harness scaling finding, not a
synthetic authentication result or a #3990 policy implementation.

The inbound flow was planned, previewed, run, and then queried through DuckDB:
1 record read, 1 record written, 1 row returned from the connection-scoped
Parquet table.  The temporary project (including local credential state and
warehouse files) was deleted after its sanitized evidence was checked.

## Verification interpretation

The 0/665 barrier result is verified as a complete, truthful measurement, not
as a provider-certification success. The isolated returned-record controls and
rate snapshots rule out treating it as a revoked credential or a #3990 quota
policy result. Follow-up is explicitly limited to #3990 admission, cleanup-safe
fixtures, and outbound #3994/#3992 foundations.

## Final correction loop

The completed correction sequence is `1/#4020 -> 2/#4022 -> 3/#4027 ->
4/#4039 -> 5/#4050`. Round 1 closed the direct/config write-scope escape
before process launch; round 3 made pre-barrier inputs fail closed; round 4
bound current artifacts to the canonical classifier; and round 5 retired the
legacy rate proof's current-certification claim and guarded bootstrap artifact
input/output. The three named round-5 RED tests were observed failing before
production edits, then the focused 50-test GREEN suite passed. This completes
**5 of 5** permitted correction loops without altering barrier concurrency,
rate admission, or the measured historical 0/665/856 result.

## Fresh current-SHA recovery — lineage 1/5

This section is the only current-SHA evidence for the captain-authorized fresh
line. It does not promote any historical provider result, classifier output,
fixture, or successful exit to a certification claim.

### Build, process, and static evidence

- Local commit at build time: `aeaec4dd18a44db573f69ece9e71b1f80ccd5ea6`.
- Built binary: `./pm`, SHA-256
  `c7f9f9c7d070d8c0f6ead322735c5146c33d47fce92a42dc083c8ab6ef554fd4`.
- `go build -o ./pm ./cmd/pm`, `./pm help connectors`, and
  `./pm connectors certify --help` all exited successfully.
- `cmd/pm/main.go` registers `certify.SetCLIRunFunc(cli.Run)` once before the
  outer `cli.Run`; `internal/connectors/certify/cliharness.go` invokes that
  function in-process. This proves the available legacy harness execution
  model is one built `pm` process, not the external Node child-process runner.
- Current source hashes: `cli_surface.json`
  `de7c51e24083ed0e886be691b9441efa8f3049fce41bbf9eb8cdcdb67be92c81`,
  `certification.json`
  `029261737a4ff74483173f74e06829622ecde6777f1a086a8e149e598d7e5a34`,
  and `rate_limits.json`
  `b0748da57a150d31ab944222a40a9ab7b54836cfb741405eea12c5a08f485198`.
- The current surface declares 1,521 implemented commands (639 direct reads,
  279 direct writes, 577 reverse-ETL actions). The checked manifest partitions
  all 1,521 rows into 900 run-owned repository, 308 run-owned organization,
  33 App-installation, and 280 feature/entitlement cases.

### Focused gates

- Fresh Node RED was observed before the behavior change; the corresponding
  GREEN proof-sweep suite passed 15/15 and the combined GitHub focused suite
  passed 51/51.
- `node scripts/github-live-lab-manifest.mjs --check` passed with the current
  1,521-row cohort counts; `node scripts/github-live-bootstrap-probes.mjs
  --check` passed with 308 organization and 33 App-marketplace cases.
- `go test -timeout 20m ./internal/connectors/certify` passed; `go test
  -timeout 20m ./internal/cli` passed in 157.243 seconds.
- The final focused GitHub suite passed 51/51. `go test -race -timeout 20m
  ./internal/connectors/certify` passed in 55.547 seconds; `go run
  ./cmd/connectorgen certification-matrix --check` and `go run
  ./cmd/agentcontractgen check` also passed.

### Honest execution assessment

`pm connectors certify github --full` is a useful in-process legacy harness,
not the #3993 all-command certification executor. It can exercise catalog
streams through GitHub → connection-owned Parquet/DuckDB materialization and
capture-backed destination modes; it has two definition-owned direct-read
candidates and one binary-download candidate. Its flow is capture-backed
source → warehouse/query only. Its schedule stage uses an ephemeral crontab
sentinel and does not prove a persisted schedule firing. It runs only the first
safe write pairing lifecycle; the `--full` sweep records remaining pairings as
`untested`, not executed. Therefore it cannot establish #3993's full direct
provider read/pagination/rate, all-write, warehouse-to-GitHub, authored-action
flow, or real-schedule acceptance.

The existing external Node sweep is now mechanically fail-closed: its only
accepted blocker evidence identifies `external_pm_per_operation`; a
`credentialed_live` record must identify `built_pm_in_process`, and the runner
rejects before credential/provider dispatch. `single_barrier_release` remains
topology only, never shared-rate-coordinator evidence.

### Live-safety decision and current result

After the required notice, `working: GitHub focused gates green; requesting
heavy validation window`, credential inventory found no approved full-parity
GitHub App secret source in this isolated environment and no initialized local
project credential state. No secret value was read, printed, fingerprinted, or
persisted. No `Polymetrics-Cert` immutable run-owned boundary configuration was
available to verify. Result: **needs-decision**.

No provider request, mutation, plan, preview, approval, execute, read-back,
or cleanup ran. Accordingly there are no provider resource IDs, no cleanup or
empty-residue result, no current rate-limit observation, no Parquet/DuckDB row
count or hash, no connection-owned table, and no flow identity to claim. The
declared App-installation policy (installation scope, 5,000 requests/hour and
900 points/60 seconds) is static policy only, not observed admission evidence.

The independent unavailable slices remain explicitly blocked: warehouse-to-
GitHub action flow by #3994/#4059, and real persisted schedule firing by #3992.
PR #4060 has not been rebased or copied; #4059 transport code has not been
copied; parked #4062 was not used.

## GSD verify-work — inline `3993 --auto`

`scripts/gsd prompt verify-work 3993 --auto` was generated and the project
adapter confirmed that the issue phase exists. The inline verifier recorded two
automated passes (execution-model provenance and built-binary focused gates) and
one third-party safety block (approved App credential plus immutable boundary).
No gap plan was created: the block is authority/configuration outside this
branch, not a code defect. The GSD advance command recomputed aggregate
`STATE.md` counts but reported that it could not parse this issue-backed legacy
phase position; progress, metric, and session mutations returned no-op. The
phase-specific durable lifecycle state remains in `3993-01-SUMMARY.md`, this
checklist, and `UAT.md`.

## No-mistakes — local-only stacked-PR disposition

`no-mistakes axi` was inspected on 2026-08-11 without starting a pipeline.
Its available `axi run` mode accepts only an intent, generic skipped steps, and
an auto-approval switch, while its normal lifecycle owns push, PR, and CI
state. It has no mode that can prove it will preserve the existing draft
stacked PR #4061, its required `feat/3988-github-certification` base, and the
captain-owned parent integration.

The required no-mistakes stacked-certification resolution report therefore
applies its direct-existing-PR exception: no pipeline run was started, no
no-mistakes push/PR/CI action occurred, and all local review, test, lint,
generator, documentation, and contract gates remain recorded above. This is a
safety disposition, not a claim that no-mistakes completed a delivery run.
