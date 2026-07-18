# Phase 429 TDD Ledger

Issue: #429 — nativize credentials namespace.
Invocation session: `issue-429-pi-openai-codex-gpt-5.6-sol-high-20260718T143346Z`
Model: `openai-codex/gpt-5.6-sol`; thinking: `high`
Starting HEAD: `0f1ec1e89cdae761e9da06ab9906fcc641b38e0a`

## GSD and skills

Doctor/list passed; `scripts/gsd prompt plan-phase 429 --skip-research` generated a prompt and is being executed inline. `scripts/gsd prompt programming-loop init --phase 429 --dry-run` failed because the command is absent, so the manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-spf13-cobra`.

## Second bounded correction ledger

Session `issue-429-second-bounded-correction-pi-openai-20260718T170705Z`; profile `Sol/high`; exact correction start `fae7d599668637bea345fe76877dd75e31dd2ad8`. Rereview `/tmp/pm-397-rereview-429.log` accepted in full. GSD doctor/list passed and the plan prompt generated; the adapter still lacks `programming-loop`, so manual universal-loop fallback is active. Loaded `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-troubleshooting`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| S0 | Review/plan | Record all three findings, effect-boundary design, RED cases, broad shared-seam verification, and checkpoint sequence before production edits | Complete |
| S1 | RED | `go test ./internal/app -run 'Test(ResolvedLocalConnectorRelativePathUsesSelectedProjectRoot|LocalConnectorCheckRevalidatesPathAfterCredentialResolution)$' -count=1`; focused CLI source/redaction command | Failed as required: both relative paths missed the selected root; both denied post-resolution retargets reached external effects; leading-hyphen add exited 1. App `3.539s`, CLI `3.554s`, combined wall `13s` |
| S2 | GREEN | Runtime-only path normalization, explicit non-secret local-write policy, Warehouse/Outbox effect-boundary validation, bounded Cobra name carrier, and required actual state file | Pass: focused app `3.512s`, CLI `16.112s`, connectors `0.355s`, safety `0.340s`; all credentials `48.502s`; app/connectors/safety `26.557s`/`0.789s`/`0.454s` |
| S3 | Verify | Focused/repeated/race/app/connectors/CLI/base differential/full repository plus gofmt/vet/build/`make verify` | In progress: repeated app `14.917s`, CLI `79.908s`, connectors/safety `0.277s`/`0.411s`; focused race app `34.593s`, CLI `180.125s`, connectors/safety pass; focused vet pass |

Tests use temporary roots and synthetic inputs only. They must not print fixture values or contact external services. The corrected state helper now requires and decodes the actual state file; this test-only correction is exercised alongside the failing behavior cases. `verificationPassed` remains false until the complete declared gate exits 0.

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
