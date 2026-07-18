# Phase 429 TDD Ledger

Issue: #429 — nativize credentials namespace.
Invocation session: `issue-429-pi-openai-codex-gpt-5.6-sol-high-20260718T143346Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `0f1ec1e89cdae761e9da06ab9906fcc641b38e0a`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 429 --skip-research` generated a prompt and is being executed inline. `scripts/gsd prompt programming-loop init --phase 429 --dry-run` failed because the command is absent, so the manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-spf13-cobra`.

## Targeted parser-order correction ledger

Session `issue-429-targeted-parser-order-correction-pi-openai-20260718T212111Z`; exact correction start `4870f53b028613fbc3632a404e9a382419d87f8a`. Rereview `/tmp/pm-397-rereview6-429.log` accepted. GSD doctor/list passed; the adapter still lacks `programming-loop`, so the manual universal-loop fallback is active. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-troubleshooting`, `golang-error-handling`, `golang-security`, and `golang-safety`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| P0 | Review/plan | Record the accepted parser-order finding, four known-flag lifecycle matrix, first-token-before-tail design, adversarial preservation gates, differential, and checkpoint sequence before production edits | Complete |
| P1 | RED | `go test ./internal/cli -run 'TestCredentials(KnownAddFlagNamesPreserveLifecycleAndTailFlags|RawInternalNameCarrierFailsClosed|LeadingInvalidNameTokensCannotDiscoverLaterNames)$' -count=1` | Failed as required in `18.545s` (wall `21.75s`): all four known-flag name adds captured an invalid normalized `name=ignored` token; raw-carrier and invalid action/name ownership guards stayed green |
| P2 | GREEN | Capture/remove the required add name before StringArray normalization; normalize only the remaining tail | Pass: focused lifecycle/raw-carrier/invalid-ownership gate `28.076s` (wall `30.55s`); all four names complete lifecycle with later spaced connector/config flags and ignored positionals |
| P3 | Verify | Focused/adversarial/repeated/race CLI, exact parent-base/start/head differential, full CLI, gofmt/vet/build/diff/scope/dependency guards | Pass: focused `79.397s`; repeated ×5 `176.825s`; race `392.438s`; differential base/head `12/12`, start add rejections `4`, base-seeded start/head `8/8`, exact output pairs `8`; full CLI `340.707s`; remaining gates clean |

Tests use temporary roots and config-only credentials. They did not read, print, summarize, or store private values or contact services.

### Targeted parser-order final evidence

- RED: focused package failed in `18.545s` (wall `21.75s`), with exactly the four known-flag-name add failures and all preservation guards green.
- GREEN: focused lifecycle/raw-carrier/invalid-ownership passed in `28.076s` (wall `30.55s`).
- Stability: broader focused/adversarial passed in `79.397s`; exact adversarial matrix repeated five times in `176.825s`; focused race passed in `392.438s`.
- Differential: exact parent base and final head each passed 12 add/inspect/remove operations; exact correction start rejected all four adds; base-seeded start/head each passed eight inspect/remove operations; eight base/head add/remove output pairs matched exactly. The first two differential harness attempts failed only in temporary Python setup (missing destination directory, then an unsupported Python 3.10 tar filter argument); the corrected harness passed in `73.03s`.
- Full CLI: `go test -timeout 20m ./internal/cli/... -count=1` passed in `340.707s` (wall `343.08s`).
- `gofmt -w cmd internal`, clean worktree diff, `go vet ./...` (`2.31s`), readonly build (`3.85s`), start-range diff/scope/dependency guards, and runtime help parity passed; help routes were byte-identical at 1252 bytes and invalid action exited 2.
- Checkpoints: `adb0eab9` planning, `7f0357a9` RED, `9e87a007` GREEN; final evidence checkpoint prepared for push.

## Compatibility correction ledger

Session `issue-429-compatibility-correction-pi-openai-20260718T202616Z`; exact correction start `7497483de2187b3117c32b9cafb3db54ebac792f`. Rereview `/tmp/pm-397-rereview5-429.log` accepted. GSD doctor/list passed; the adapter still lacks `programming-loop`, so the manual universal-loop fallback is active. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-troubleshooting`, `golang-error-handling`, `golang-security`, and `golang-safety`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| K0 | Review/plan | Record the accepted compatibility finding, complete safety-valid name matrix, ordinary-validation design, adversarial preservation gates, differential, and checkpoint sequence before production edits | Complete |
| K1 | RED | `go test ./internal/cli -run 'TestCredentials(SafetyValidPrivateNamesSupportAddInspectRemove|RawInternalNameCarrierFailsClosed|LeadingInvalidNameTokensCannotDiscoverLaterNames)$' -count=1` | Failed as required in `23.030s`: all 14 safety-valid short/double-hyphen add cases were rejected by private validation; raw-carrier rejection and invalid action/name ownership guards stayed green |
| K2 | GREEN | Remove the private validator and validate privately carried names through `validateCredentialIdentifier(..., "credential")`; rerun compatibility plus raw-carrier and invalid ownership guards | Pass: focused compatibility/adversarial gate `56.416s`; all 14 names complete add/inspect/remove and first-token ownership, while raw-carrier and invalid ownership guards remain green |
| K3 | Verify | Focused/adversarial/repeated/race CLI, exact parent-base/start/head differential, full CLI, help parity, gofmt/vet/build/diff/scope/dependency guards | Pass: repeated ×5 `352.467s`; aggregate race timed out at 600s, then compatibility/adversarial partitions passed in `457.137s`/`262.781s`; differential 14 names and 42 lifecycle ops per base/head with 42 start rejections; full CLI `333.259s`; remaining gates clean |

Tests use temporary roots and config-only credentials. They did not read, print, summarize, or store private values or contact services.

### Compatibility correction final evidence

- RED: all 14 compatibility subtests failed only at private validation in `23.030s`; raw-carrier and invalid action/name ownership guards stayed green.
- GREEN: focused compatibility/adversarial `56.416s`; repeated ×5 `352.467s`.
- Race: the aggregate command exceeded 600 seconds without a test failure; exact compatibility and adversarial partitions then passed in `457.137s` and `262.781s`.
- Differential: original parent base and final head each passed 42 add/inspect/remove operations for 14 names; correction start rejected all 42 corresponding operations; 28 base/head add/remove stdout/stderr pairs were byte-identical and inspect metadata named the owned first token.
- Full CLI: `go test -timeout 20m ./internal/cli/... -count=1` passed in `333.259s`.
- Runtime help topic/bare/long-help were byte-identical at 1252 bytes with empty stderr; invalid action exited 2.
- `gofmt -w cmd internal`, clean format diff, `go vet ./...`, `go build ./cmd/pm`, readonly module graph, start-range `git diff --check`, scope, dependency, and clean-worktree gates passed. An initial shell-only scope assertion used the wrong lexical filename order; the corrected exact file-list assertion passed.
- Checkpoints: `d8aec609` planning, `a3752713` RED, `199b802c` GREEN; final evidence checkpoint prepared for push.

## Final bounded correction ledger

Session `issue-429-final-bounded-correction-pi-openai-20260718T194756Z`; exact correction start `80246e42f508f685d281fecbcc3735eadcf271a9`. Rereview `/tmp/pm-397-rereview4-429.log` accepted. GSD doctor/list passed; the adapter still lacks `programming-loop`, so the manual universal-loop fallback is active. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| X0 | Review/plan | Record both accepted findings, first-token ownership and early-validation design, overwrite raw-temp cleanup lifetime, RED cases, differential, and verification/checkpoint sequence before production edits | Complete |
| X1 | RED | `go test ./internal/cli -run 'TestCredentials(AddLeadingHyphenNameParsesLaterSourceFlagsAndIgnoresExtraPositionals|RawInternalNameCarrierFailsClosed|LeadingInvalidNameTokensCannotDiscoverLaterNames)$' -count=1`; `go test ./internal/app -run '^TestWarehouseMaterializationRejectsFinalFileSymlinkEscape$' -count=1` | Failed as required: leading-hyphen add exited 1 while raw-carrier/no-discovery guards stayed green (CLI package `23.649s`); overwrite truncate and truncate-create left the opened raw temp (app package `3.576s`) |
| X2 | GREEN | Carry every leading-hyphen first name privately, validate it before action execution, and register overwrite cleanup immediately after raw-temp open | Pass: focused CLI/app/safety green (`19.187s`/`3.362s`/`0.473s`); repeated ×5 CLI/app green (`97.954s`/`16.976s`); race CLI/app/localwrite green (`214.993s`/`34.800s`/`1.406s`) |
| X3 | Verify | Focused/repeated/race CLI/app/localwrite, exact-base differential, broader relevant packages, gofmt/vet/build/diff/scope/dependency guards | Pass: connector local-write ×5/race `0.471s`/`1.364s`; differential base/start/head `0/1/0` and base/head byte-identical; full relevant CLI/app/safety/connectors `289.768s`/`30.339s`/`0.320s`/`0.474s`; help 3/3 exact, invalid 2; gofmt/vet/build/diff/scope/dependency clean |

Tests use temporary roots, config-only credentials, synthetic records, and non-sensitive sentinel bytes. They did not print private fixture content or contact services.

### Final correction evidence

- RED: CLI `23.649s`, with the compatibility case exiting 1 and raw/no-discovery guards green; app `3.576s`, with both non-deduped overwrite final-open failures leaving raw temps.
- GREEN: focused CLI/app/safety `19.187s`/`3.362s`/`0.473s`; repeated ×5 CLI/app `97.954s`/`16.976s`; race CLI/app/localwrite `214.993s`/`34.800s`/`1.406s`; connector write ×5/race `0.471s`/`1.364s`.
- Differential: exact parent base `0f1ec1e8` exit 0, correction start `80246e42` exit 1, final head exit 0; base/head stdout and stderr byte-identical, 46/0 bytes; resulting `-legacy` metadata inspect succeeded without private values.
- Full relevant packages: CLI `289.768s`, app `30.339s`, safety `0.320s`, connectors `0.474s`.
- Gofmt, `go vet ./...`, `go build ./cmd/pm`, start-range/current diff checks, unchanged dependency graph, and clean connector-def/docs/website scope passed.
- Runtime help topic/bare/long-help were byte-identical at 1252 bytes with empty stderr; invalid action exited 2.
- Checkpoints: `a717fb6d` planning, `fc49483d` RED, `74e8cffe` GREEN; final evidence checkpoint prepared for push.

## Fourth bounded correction ledger

Session `issue-429-fourth-bounded-correction-pi-openai-20260718T185126Z`; exact correction start `0d70335f37456f42432b3c502860f7b43231ed98`. Rereview `/tmp/pm-397-rereview3-429.log` accepted. GSD doctor/list passed; the adapter still lacks `programming-loop`, so the manual universal-loop fallback is active. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| F0 | Review/plan | Record the final-file symlink escape, effect-time `os.Root` design, temp-only RED matrix, shared-seam gates, and commit/push checkpoints before production edits | Complete |
| F1 | RED | `go test ./internal/connectors -run 'Test(Warehouse|Outbox)WriteRejectsFinalFileSymlinkEscape$' -count=1`; `go test ./internal/app -run '^TestWarehouseMaterializationRejectsFinalFileSymlinkEscape$' -count=1` | Failed as required: all 5 connector and all 6 app append/truncate/create cases followed final-file links outside the root; connectors `0.317s` (wall `0.97s`), app `3.359s` (wall `6.31s`) |
| F2 | GREEN | Add `safety.LocalWriteFS` over Go 1.25 `os.Root`; route Warehouse/Outbox Check+Write and app raw/final mkdir/open/read/remove/rename effects through one held scope | Pass: focused safety/connectors/app `7.73s` wall; repeated ×5 `18.12s`; focused race passed (app `33.985s`); broader safety/connectors/app `32.24s` wall; explicit external, nil-policy, modes, in-root symlink/nonexisting, and rename-replacement coverage green |
| F3 | Verify | Focused/repeated/race safety/connectors/app/CLI; broader packages; full repository; gofmt/vet/build/`make verify`; scope/dependency guards | Pass: comprehensive focused ×5 `42.12s`; race `84.54s`; broader packages `350.01s`; full repo `347.88s`; vet `3.22s`; build `1.81s`; `make verify` `374.34s`, lint 0, connector validation 547/0 |

Tests use only `t.TempDir`, synthetic records, and non-sensitive sentinel bytes. They did not display fixture contents or contact services.

### Fourth correction final evidence

- RED: connectors package `0.317s` and app package `3.359s`; all 11 append/truncate/create final-link cases reached external targets before production edits.
- GREEN: held Go 1.25 `os.Root` scope covers every relevant directory/final open/read/remove/rename effect; comprehensive focused ×5 passed in `42.12s`, and focused `-race` passed in `84.54s`.
- Broader safety/connectors/app/CLI passed in `350.01s` (app `33.320s`, CLI `300.061s`, certify `341.868s`).
- Full repository passed in `347.88s` (app `32.492s`, CLI `296.796s`, certify `341.340s`).
- `gofmt -w cmd internal`, `go vet ./...` (`3.22s`), and Go 1.25.12 build (`1.81s`) passed.
- `GOTOOLCHAIN=go1.25.12 make verify` passed in `374.34s`; lint 0, connector validation 547/0, docs validation and dependency-free local smoke passed.
- Scope/dependency guards: readonly module graph; no `go.mod`, `go.sum`, connector definition, checked-in CLI docs, or website delta; final local write files contain no ordinary `os.OpenFile`/`os.Rename` effects.
- Checkpoints: `810a35d6` planning, `f6a15a83` RED, `bc13b768` GREEN; final evidence checkpoint prepared for push.

## Third bounded correction ledger

Session `issue-429-third-bounded-correction-pi-openai-20260718T180016Z`; exact correction start `6158cdc92d5df01cbaa577ceeb5a870ddcb8f685`. Rereview `/tmp/pm-397-rereview2-429.log` accepted. GSD doctor/list passed; the adapter still lacks `programming-loop`, so manual universal-loop fallback is active. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| T0 | Review/plan | Record the raw hidden-carrier ownership bypass, 12-case action/form RED matrix, inaccessible state design, safety constraints, and verification/checkpoint sequence before production edits | Complete |
| T1 | RED | `go test ./internal/cli -run '^TestCredentialsRawInternalNameCarrierFailsClosed$' -count=1` | Failed as required in `11.651s`: 9/12 cases violated the contract; assigned/spaced forms overrode all four actions and exited 0, while add/bare returned runtime code 3 instead of usage 2 |
| T2 | GREEN | Remove hidden pflag; use private command-context state; reject raw carrier before parsing; preserve leading-hyphen names and globals | Pass: focused/adversarial `34.099s`, raw matrix ×5 `56.733s`, focused race `273.254s` |
| T3 | Verify | Focused/repeated/race/adversarial/base differential/full CLI plus gofmt/vet/build/diff | Pass: normal exact-base differential 7/7 byte-equal; all 12 head raw cases exit 2; full CLI `332.836s`; gofmt/vet/build/diff/scope clean |

Tests use temporary roots and synthetic metadata/config only. They must not print fixture values or contact external services.

### Third correction final evidence

- RED: `11.651s`; 9/12 contract violations before production edits.
- GREEN/adversarial: `34.099s`; raw matrix repeated five times in `56.733s`; focused race `273.254s`.
- Exact `6158cdc9` differential: seven preserved help/error/ordinary/leading-name/global cases matched exit/stdout/stderr byte-for-byte; all 12 current raw carrier cases exited usage 2.
- Full CLI: `go test -timeout 20m ./internal/cli/... -count=1` passed in `332.836s`.
- `gofmt -w cmd internal`, `go vet ./...`, `go build ./cmd/pm`, start-range/current `git diff --check`, scope, dependency, and clean-worktree gates passed.
- No private value output, real credential, service, dependency, PR, or external review.

## Second bounded correction ledger

Session `issue-429-second-bounded-correction-pi-openai-20260718T170705Z`; profile `Sol/high`; exact correction start `fae7d599668637bea345fe76877dd75e31dd2ad8`. Rereview `/tmp/pm-397-rereview-429.log` accepted in full. GSD doctor/list passed and the plan prompt generated; the adapter still lacks `programming-loop`, so manual universal-loop fallback is active. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-troubleshooting`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| S0 | Review/plan | Record all three findings, effect-boundary design, RED cases, broad shared-seam verification, and checkpoint sequence before production edits | Complete |
| S1 | RED | `go test ./internal/app -run 'Test(ResolvedLocalConnectorRelativePathUsesSelectedProjectRoot|LocalConnectorCheckRevalidatesPathAfterCredentialResolution)$' -count=1`; focused CLI source/redaction command | Failed as required: both relative paths missed the selected root; both denied post-resolution retargets reached external effects; leading-hyphen add exited 1. App `3.539s`, CLI `3.554s`, combined wall `13s` |
| S2 | GREEN | Runtime-only path normalization, explicit non-secret local-write policy, Warehouse/Outbox effect-boundary validation, bounded Cobra name carrier, and required actual state file | Pass: focused app `3.512s`, CLI `16.112s`, connectors `0.355s`, safety `0.340s`; all credentials `48.502s`; app/connectors/safety `26.557s`/`0.789s`/`0.454s` |
| S3 | Verify | Focused/repeated/race/app/connectors/CLI/base differential/full repository plus gofmt/vet/build/`make verify` | Complete: full app `25.808s`, CLI `280.439s`, connectors certify `337.004s`; exact-start differential preserved 5/5 and corrected leading-name base/head exits 1/0; full repo app `27.976s`, CLI `285.504s`, certify `340.518s`; gofmt/vet/build and `make verify` pass |

Tests use temporary roots and synthetic inputs only. They did not print fixture values or contact external services. The corrected state helper requires and decodes the actual state file. Full verification passed, including local smoke through the existing reverse ETL plan → preview → approval → execute flow; no external runtime service was used.

### Second correction final evidence

- Focused GREEN: app `3.512s`, CLI `16.112s`, connectors `0.355s`, safety `0.340s`.
- Stability: repeated app `14.917s`, CLI `79.908s`, connectors/safety `0.277s`/`0.411s`; focused race app `34.593s`, CLI `180.125s`, connectors/safety pass.
- Broader packages: all credentials `48.502s`; app `25.808s`; CLI `280.439s`; connectors tree passed with certify `337.004s`.
- Exact `fae7d599` differential: five preserved help/error/ordinary-add cases matched exit/stdout/stderr; corrected leading-hyphen add changed from base exit 1 to head exit 0.
- Full repository: app `27.976s`, CLI `285.504s`, certify `340.518s`; all packages passed.
- `make verify`: pass; docs validation and local smoke passed, lint reported 0 issues, connector validation checked 547 definitions with 0 findings.
- Built help: topic/bare/long-help exits 0, stderr empty, 1252 identical bytes, SHA-256 `dbe21fc0c6a594046611ad51c5a4119f8c82aa8a30879425bbb74485ea6fd949`.
- No dependency, checked-in docs/website, private fixture output, external service, real credential, PR, or review.

## Bounded independent review correction ledger

Session `issue-429-bounded-security-compat-correction-pi-openai-codex-gpt-5.6-sol-high-20260718T155702Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact correction start `758b059bbeb54032dbcd1b9a2a540ca83058861b`. Review `/tmp/pm-397-review-429.log` accepted in full. GSD doctor/list passed; the adapter still lacks the documented programming-loop command, so manual universal-loop fallback remains active. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-lint`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| R0 | Review/plan | Record HIGH symlink escape, MEDIUM legacy-name stranding, LOW help-tail incompatibility and verification plan before production edits | Complete |
| R1 | RED | `go test ./internal/cli -run 'TestCredentials(TestRejectsSymlinkEscapeBeforeLocalConnectorEffects|LegacyValidateIdentifierNamesRemainInspectableAndRemovable|NamespaceHelpIgnoresTrailingUnknownFlags)$' -count=1` | Failed as required in `6.546s`: warehouse/outbox denied cases created external effects; `_legacy`/`.legacy` inspect rejected; long/short namespace help tails exited 2 |
| R2 | GREEN | Reusable realpath/nearest-existing-ancestor containment in `ValidateLocalWritePath`, revalidation in `resolveCredential` immediately before runtime use, restored credential-name compatibility, namespace-help tail normalization | Pass: focused safety+CLI `8.257s`; full credentials `46.463s`; safety+app `23.300s`; focused race safety+CLI `109.283s` |
| R3 | Verify | Focused/repeated/race/security/base differential/full CLI/path tests plus gofmt/vet/full tests/build/`make verify` | Pass: repeated correction `49.385s`; repeated safety `0.560s`; full CLI `282.493s`; full repo CLI `284.380s`/certify `340.838s`; exact base help 2/2; gofmt/vet/build/`make verify` pass |

Tests must not use, inspect, print, summarize, or store secret content and must not contact services. The path test checks only filesystem existence outside the project.

### Bounded correction final evidence

- Focused RED: failed in `6.546s`; both denied local connectors created external directories, both requested legacy names were stranded, and both namespace help forms exited 2.
- Focused GREEN: safety+CLI `8.257s`; all credentials `46.463s`; safety+app `23.300s`; focused race safety+CLI `109.283s`.
- Stability: correction tests five times `49.385s`; safety identifier/path tests twenty times `0.560s`.
- Full CLI: `go test ./internal/cli/... -count=1` passed in `282.493s`.
- Exact differential: current long/short trailing-unknown help exits 0, emits 0 stderr bytes, hashes to `dbe21fc0c6a594046611ad51c5a4119f8c82aa8a30879425bbb74485ea6fd949`, and byte-matches `0f1ec1e8`; correction start `758b059b` exits 2.
- Full repository: `go test -timeout 20m ./...` passed (CLI `284.380s`, certify `340.838s`); gofmt, `go vet ./...`, `go build ./cmd/pm`, and `make verify` passed; lint 0, 547 connector definitions/0 findings.
- Runtime help: help topic, bare namespace, and trailing-unknown long help exit 0 and are byte-equal. Golden transcripts and generated manuals pass in `6.146s`.
- Scope: no `go.mod`, `go.sum`, connector definition, checked-in CLI docs, or website delta. No real credential, secret material, service, dependency, PR, or external review.

## Prior local security review correction ledger

Session `issue-429-action-name-boundary-fix-pi-openai-codex-gpt-5.6-sol-high-20260718T151236Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact correction start `36b2e388d78aea5e79dac63b10f6310d25002198`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| C0 | Review/plan | Identify post-action positional-name discovery bypass; update phase artifacts before correction production edit | Complete |
| C1 | RED | `go test ./internal/cli -run '^TestCredentialsLeadingInvalidNameTokensCannotDiscoverLaterNames$' -count=1` | Failed before correction: 8/10 assigned-unknown, short, assigned-help-like, and literal add/remove cases executed later names; package `7.940s` |
| C2 | GREEN | Insert required-name boundary and strict credential/connector leading-character validation | Pass: correction+strict names `18.166s`; full focused `40.299s` |
| C3 | Verify | Focused/repeated/race, exact differential, full CLI, formatting/static/build/scope | Pass: full CLI `275.269s`; differential 28/28; docs/website clean; gofmt/vet/build/`make verify` pass |

The correction tests use config-only temporary credentials and no secret source. Bare unknown cases already failed closed; the eight failing cases prove Cobra consumed an invalid first name token and discovered a later name.

GREEN inserts a literal boundary immediately after exact name-taking actions when the legacy first-name token begins with `-`. The child receives that token as the name and rejects it before connector/source lookup; Cobra cannot consume it and discover a later name. Credential and connector names now require an ASCII alphanumeric first character after the existing identifier/path-traversal checks. Ten add/remove boundary cases pass, full focused tests pass, repeated correction tests pass five times, focused race passes, and goldens remain unchanged.

## Planned red / green / refactor log

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create all six issue-local phase artifacts before test or production edits | Complete |
| 1 | RED | Add focused credentials tree/operation/help/security tests; run `go test ./internal/cli -run 'Credentials|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1` | Failed as expected before production edits: `undefined: newCredentialsCobraCommand`; package build failed |
| 2 | GREEN | Native tree, typed flags/handler, controlled input, action boundary, strict validation | Pass: focused credentials/router `25.475s`; focused race subset `111.267s` |
| 3 | Refactor | Focused/repeated/race/security/router/golden/full CLI and exact legacy differential | Pass; final full CLI `275.269s`, preserved differential 28/28 exact |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` | Pass; final `make verify` CLI `278.385s`, certify `342.715s`, lint 0, 547 connectors/0 findings |
| 5 | Parity/delivery | Built help/list/error checks, docs/website/generated diff, scope/dependency guards, commit/push | Pass; final artifact commit/push pending |

## Planned RED coverage

- `credentials` is absent from legacy wrappers and native Cobra owns `add`, `list`, `inspect`, `test`, `remove`, and hidden positional `help`.
- `add` declares `stringArray` flags `connector`, `from-env`, `value-stdin`, and `config`, all with legacy bare-value `true`, repeated semantics, unknown tolerance, and no-file completion seam.
- Add/list/remove preserve text and JSON output, spaced/assigned/bare/repeated flags, extra positionals, unknown flags, and fresh-tree re-entrancy.
- Bare/text/JSON/long/short/positional help remains canonical; action-tail help and literal `--` retain legacy behavior.
- Invalid actions and invalid assigned global booleans retain usage/validation categories; valid assigned global booleans remain effective.
- Credential, connector, secret-field, environment-variable, and config-key names reject control/path-traversal input before persistence or environment/stdin reads.
- Warehouse/outbox paths cannot escape the temporary root without explicit existing opt-in; allowed local and file-source paths remain valid.
- `--from-env` supports repeated mappings, detects malformed/missing/empty named sources, and never prints values. `--value-stdin` reads only controlled Cobra input, trims only trailing CR/LF, and final repeated field selection remains compatible. Config-only credentials remain valid; no interactive input path exists.
- Opaque synthetic env/stdin fixtures are absent from stdout, stderr, and state metadata after success and error paths; tests never log fixture content.
- Leading unknown, short, assigned help-like, and literal-boundary tokens cannot discover or execute later add/remove actions; temporary state remains unchanged.

## Exact RED

Captured after the complete focused test-only edit and before any production edit:

```text
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/credentials_cli_test.go:22:9: undefined: newCredentialsCobraCommand
FAIL\tpolymetrics.ai/internal/cli [build failed]
FAIL
```

The missing native constructor is intentional. The same test-only checkpoint specifies native ownership/flags, list/add/remove, help/global behavior, strict identifiers/path containment, controlled env/stdin source handling, redaction, legacy action-tail compatibility, and fail-closed action discovery. No test executed and no stdin/environment credential fixture was read during this RED build failure.

## Focused GREEN

```text
$ go test ./internal/cli -run 'Credentials|CobraRouterShellBuildsFreshHiddenWrapperTree' -count=1
ok  \tpolymetrics.ai/internal/cli\t25.475s

$ go test -race ./internal/cli -run 'CredentialsCommandIsNative|CredentialsSecretSources|CredentialsOutputsAndErrors|CredentialsLeadingInvalid' -count=1
ok  \tpolymetrics.ai/internal/cli\t111.267s
```

Native Cobra now owns the complete credentials subtree. Current flags are typed `StringArray` values with bare compatibility, stdin comes only from `cmd.InOrStdin`, invalid action heads are bounded before Cobra discovery, and all identifier/config/path checks run before named environment or stdin values are read. Opaque fixture tests pass without emitting fixture content. The golden transcript test passes and a 28-case start-vs-head differential matches exact exit/stdout/stderr for preserved help, list, add flag forms, unknown/extra/tail/literal inputs, invalid heads, and globals.

## Final GREEN / refactor evidence

- Focused credentials/router: pass (`40.299s` after correction).
- Repeated action-name boundary: pass five times (`62.622s`).
- Focused race/security boundary: pass (`248.367s`).
- Golden transcripts: pass (`5.602s` final); generated CLI manual parity test also passed (`6.437s` before correction, unchanged by correction).
- Full `internal/cli/...`: pass (`275.269s` final).
- Exact preserved start differential: 28/28 exit/stdout/stderr matches after correction.
- Built binary in a temporary root: help topic/bare/long byte-equal; JSON manual/list kinds present; invalid action exit 2.
- Temporary CLI/connector docs generation and validation: pass; `docs/cli` byte diff clean.
- Website `gen:docs`: 11 pages generated; tracked website diff clean.
- `gofmt -w cmd internal`, `go vet ./...`, `go build ./cmd/pm`, `git diff --check`: pass.
- Full `go test -timeout 20m ./...`: pass before correction; final `make verify` reran the full suite after correction (CLI `278.385s`, certify `342.715s`), docs validation, local smoke, lint (`0 issues`), and connector validation (547/0).
- Scope/dependency guards: no `go.mod`, `go.sum`, connector definition, checked-in CLI docs, website, or golden delta.

No focused or full command printed or logged opaque credential fixture content. No external connector, optional service, or credentialed external check ran.
