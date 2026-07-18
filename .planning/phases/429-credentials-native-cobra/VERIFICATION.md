# Phase 429 Verification

Invocation session: `issue-429-pi-openai-codex-gpt-5.6-sol-high-20260718T143346Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `0f1ec1e89cdae761e9da06ab9906fcc641b38e0a`.

## Compatibility correction checklist

Invocation `issue-429-compatibility-correction-pi-openai-20260718T202616Z`; exact start `7497483de2187b3117c32b9cafb3db54ebac792f`.

- [x] Read `/tmp/pm-397-rereview5-429.log` and accept the MEDIUM safety-valid legacy-name compatibility finding.
- [x] Run GSD doctor/list, record unavailable programming-loop/manual fallback, load required skills/policies, and update issue artifacts before production edits.
- [x] RED: 14 `safety.ValidateIdentifier`-valid short leading-hyphen and double-hyphen representatives cover allowed suffix classes across config-only add/inspect/remove; all were rejected by private validation at exact correction start (`23.030s`).
- [x] RED preservation gate: raw internal-carrier rejection and invalid action/name ownership remained green; the only failures were the 14 compatibility subtests.
- [x] GREEN: private name carriage uses ordinary credential identifier validation with no private length/double-hyphen restrictions and no action-discovery changes; focused compatibility/adversarial gate passed in `56.416s`.
- [x] Focused (`56.416s`), adversarial, repeated ×5 (`352.467s`), split race (`457.137s`/`262.781s`), exact parent-base/start/head differential, and full CLI (`333.259s`) pass without private data output. The aggregate race command timed out at 600 seconds without a failure before exact partitions passed.
- [x] Differential: 14 names; parent base and head each passed 42 lifecycle operations; exact correction start rejected 42 add/inspect/remove operations; 28 base/head add/remove output pairs matched exactly.
- [x] Runtime help/manual/website/generated parity remains unchanged: topic/bare/long-help are byte-identical at 1252 bytes with empty stderr, invalid action exits 2, and checked-in docs/website scope is clean.
- [x] `gofmt -w cmd internal`, clean format diff, `go vet ./...`, `go build ./cmd/pm`, readonly module graph, current/start diff, scope, dependency, and clean-worktree gates pass. The first shell-only scope assertion had a lexical-order typo; the corrected exact assertion passed.
- [x] Planning, RED, and GREEN checkpoints committed and pushed; final evidence checkpoint prepared for push. No service, dependency, PR, or review.

Result: pass; `verificationPassed=true` for the full user-declared compatibility-correction gate set.

## Final bounded correction checklist

Invocation `issue-429-final-bounded-correction-pi-openai-20260718T194756Z`; exact start `80246e42f508f685d281fecbcc3735eadcf271a9`.

- [x] Read `/tmp/pm-397-rereview4-429.log` and accept the MEDIUM leading-hyphen compatibility and LOW overwrite-temp cleanup findings.
- [x] Run GSD doctor/list, record missing programming-loop/manual fallback, load required skills/policies, and update issue plan/TDD/verification artifacts before production edits.
- [x] RED: leading-hyphen add with later connector/source/global flags and an ignored extra positional exited 1 at current head, while the focused raw-carrier and no-later-name-discovery guards remained green (CLI package `23.649s`); exact-base success remains the differential target.
- [x] RED: overwrite final-temp open rejection left the already-open raw temp in truncate and truncate-create cases (app package `3.576s`).
- [x] GREEN: every leading-hyphen first required-name token is privately owned and validated before action execution; later flags/ignored positionals work, raw carrier remains usage-rejected, and later names are never discovered. Focused and repeated ×5 CLI passed (`19.187s`/`97.954s`).
- [x] GREEN: overwrite cleanup is active immediately after raw-temp open and removes it when final-temp open fails. Focused and repeated ×5 app passed (`3.362s`/`16.976s`).
- [x] Focused/repeated/race CLI/app/localwrite and exact-start differential pass without private data output. Race CLI/app/localwrite passed (`214.993s`/`34.800s`/`1.406s`); connector writes ×5/race passed (`0.471s`/`1.364s`); exact parent-base/start/head exits were `0/1/0`, with base/head stdout and stderr byte-identical.
- [x] Broader relevant packages passed: CLI `289.768s`, app `30.339s`, safety `0.320s`, connectors `0.474s`; gofmt, `go vet ./...`, `go build ./cmd/pm`, diff/scope/dependency guards passed.
- [x] Runtime help/manual/website/generated surfaces remain unchanged or parity-verified: topic/bare/long-help were byte-identical at 1252 bytes with empty stderr, invalid action exited 2, and checked-in docs/website scope is clean.
- [x] Planning, RED, and GREEN checkpoints committed and pushed; final evidence checkpoint prepared for push. No service, dependency, PR, or external review.

Result: pass; `verificationPassed=true` for the full user-declared final-correction gate set.

## Fourth bounded correction checklist

Invocation `issue-429-fourth-bounded-correction-pi-openai-20260718T185126Z`; exact start `0d70335f37456f42432b3c502860f7b43231ed98`.

- [x] Read `/tmp/pm-397-rereview3-429.log` and accept the HIGH final-file symlink finding.
- [x] Run GSD doctor/list, record missing programming-loop/manual fallback, load required skills/policies, inspect Go 1.25 `os.Root`, and update issue artifacts before production edits.
- [x] RED: temp-only Warehouse append/truncate/create final-link cases failed before production edits; all 3 followed the link (`internal/connectors`, package `0.317s`) without printing target contents.
- [x] RED: temp-only Outbox append/create final-link cases failed before production edits; both followed the link (same focused connectors run) without printing target contents.
- [x] RED: temp-only app append/truncate/create and deduped truncate/create materialization final-link cases failed before production edits; all 6 followed the link (`internal/app`, package `3.359s`) without printing target contents.
- [x] GREEN: confined local writes use one held Go 1.25 `os.Root` for all relevant Warehouse/Outbox and app directory, open/read, cleanup, and rename effects; focused tests passed in `7.73s` wall and repeated ×5 in `18.12s`.
- [x] GREEN: explicit `allow_external_path=true`, nil-policy compatibility, `0700`/`0600` modes, append/overwrite semantics, in-root relative symlinks, nonexisting paths, and final-symlink rename replacement remain supported; focused race passed (app `33.985s`).
- [x] Focused/repeated/race safety/connectors/app/CLI passed (×5 `42.12s`; race `84.54s`); broader package gate passed in `350.01s`.
- [x] Full repository passed in `347.88s`; `gofmt -w cmd internal`, `go vet ./...` (`3.22s`), Go 1.25.12 build (`1.81s`), and `make verify` (`374.34s`) passed.
- [x] Dependency/scope/help-doc parity guards pass: readonly unchanged module graph, no connector-def/docs/website delta, full CLI plus local smoke green; no CLI surface change required help/manual/website edits.
- [x] No private data display, external service, dependency, PR, or external review. Existing `make verify` local smoke followed plan → preview → approval → execute.
- [x] Planning, RED, and GREEN checkpoints committed and pushed; final evidence checkpoint prepared for push.

Result: pass; `verificationPassed=true` for the full user-declared fourth-correction gate set.

## Third bounded correction checklist

Invocation `issue-429-third-bounded-correction-pi-openai-20260718T180016Z`; exact start `6158cdc92d5df01cbaa577ceeb5a870ddcb8f685`.

- [x] Read `/tmp/pm-397-rereview2-429.log` and accept the MEDIUM raw internal-carrier ownership finding.
- [x] Run GSD doctor/list, record missing programming-loop/manual fallback, load required skills/policies, and update issue artifacts before production edits.
- [x] RED: the 12-case `add|inspect|test|remove` × assigned/bare/spaced matrix failed before production edits (`11.651s`): 9 contract violations, including successful assigned/spaced ownership overrides for every action and add/bare runtime code 3 instead of usage 2.
- [x] GREEN: all raw internal-carrier spellings return usage/fail closed; no positional ownership override, wrong record access/removal, or synthetic value output occurs (`34.099s`; repeated ×5 `56.733s`).
- [x] GREEN: safety-valid leading-hyphen credential names and normal current/global flags remain supported without a user-addressable pflag; focused race passed in `273.254s`.
- [x] Focused, repeated, race, adversarial, exact-start differential, full CLI, gofmt, vet, build, and diff/scope guards pass. Full CLI: `332.836s`; normal differential byte-identical 7/7; all 12 current raw cases exit 2.
- [x] Runtime help/manual/website/generated surfaces remain unchanged or parity-verified by exact help differential and clean checked-in docs/website scope.
- [x] Planning, RED, GREEN, and final evidence checkpoints committed and pushed; no private data display, real credentials, services, dependencies, PR, or review.

Result: pass; `verificationPassed=true` for the user-declared third-correction gate set.

## Second bounded correction checklist

Invocation `issue-429-second-bounded-correction-pi-openai-20260718T170705Z`; profile `Sol/high`; exact start `fae7d599668637bea345fe76877dd75e31dd2ad8`.

- [x] Read `/tmp/pm-397-rereview-429.log` and accept all three HIGH/MEDIUM/LOW findings.
- [x] Run GSD doctor/list, generate the phase plan prompt, record missing programming-loop/manual fallback, and update issue artifacts before production edits.
- [x] RED: selected-root relative warehouse/outbox path effects missed the selected root (`internal/app`, `3.539s`).
- [x] RED: both denied post-resolution retarget cases reached external effects while explicit opt-in remained allowed (`internal/app`, same run).
- [x] RED: safety-valid leading-hyphen add with later connector/source flags exited 1 (`internal/cli`, `3.554s`); existing no-discovery cases remain in the focused suite.
- [x] Test helper correction: state-redaction helper now requires and parses `.polymetrics/state/state.json` and is exercised by the focused CLI run.
- [x] GREEN: runtime-only path normalization and explicit non-secret effect policy pass focused app/connectors/safety tests; nil-policy direct connector compatibility remains green.
- [x] GREEN: bounded Cobra name carrier preserves the first token, parses later connector/source/global flags, and keeps later-name discovery tests green.
- [x] Focused, repeated, race, app, connectors, CLI, exact-start differential, full repository, gofmt, vet, build, and `make verify` pass.
- [x] Runtime help/manual/website/golden surfaces are unchanged or parity-verified: topic/bare/long-help are byte-identical; no checked-in docs/website delta.
- [x] Plan, RED, and GREEN checkpoints committed and pushed; final evidence checkpoint prepared for push. No real credentials, private fixture output, external services, dependencies, PR, or review.

Result: pass; `verificationPassed=true`. Full repository tests passed (app `27.976s`, CLI `285.504s`, certify `340.518s`), and `make verify` passed with lint 0 and connector validation 547/0.

## Bounded review correction checklist

Invocation `issue-429-bounded-security-compat-correction-pi-openai-codex-gpt-5.6-sol-high-20260718T155702Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `758b059bbeb54032dbcd1b9a2a540ca83058861b`.

- [x] Read `/tmp/pm-397-review-429.log` and accept all HIGH/MEDIUM/LOW findings.
- [x] Update PLAN/TDD-LEDGER/VERIFICATION/RUN-STATE before production edits.
- [x] Focused RED proves symlink escape, legacy leading `_`/`.` inspect/remove regression, and namespace help-tail regression (`6.546s`, all three findings reproduced).
- [x] Symlink-resolved warehouse/outbox escape fails before any external filesystem effect without opt-in.
- [x] Explicit `allow_external_path=true` semantics and platform behavior remain supported; symlink tests skip only when the platform cannot create symlinks.
- [x] Legacy names accepted by `safety.ValidateIdentifier`, including leading `_`, `.`, and `-`, remain inspectable/removable.
- [x] `credentials --help` and `-h` ignore unknown trailing flags like exact base.
- [x] Focused, repeated, race, security, exact base differential, full CLI, safety path, gofmt, vet, full tests, build, and `make verify` pass.
- [x] No secret material, real credentials/services, dependencies, PR, or external review.

## TDD and behavior

- [x] Six phase artifacts created before test or production edits.
- [x] Exact focused RED captured before production edits (`undefined: newCredentialsCobraCommand`).
- [x] Native credentials/add/list/inspect/test/remove/help tree; legacy wrapper removed.
- [x] Typed repeated/bare current flags; spaced/assigned/unknown/extra positional forms compatible.
- [x] Bare/text/JSON/long/short/positional help parity in focused tests.
- [x] Invalid action/global assigned boolean behavior preserved in focused tests.
- [x] Strict identifiers and existing write-path containment covered.
- [x] Env/stdin-only secret intake and source-selection rules covered with controlled opaque fixtures.
- [x] Stdout/stderr/state/project-file redaction covered without logging fixture values.
- [x] Leading invalid namespace/action-name tokens cannot discover or execute later actions or names.
- [x] Credentials action-tail help and literal `--` compatibility preserved.
- [x] Only the credentials `parseFlags` call removed; dynamic/shared parser remains.

## Gates

- [x] Focused credentials/router tests (`25.475s`).
- [x] Focused repeated action-name boundary tests (`-count=5`, `62.622s`).
- [x] Focused `-race` credentials subset (`111.267s`).
- [x] Security/redaction/path/action-boundary tests.
- [x] Golden transcript test (`5.513s`) and fixture unchanged.
- [x] Full `internal/cli/...` (`275.269s` final).
- [x] Full repository tests (`go test -timeout 20m ./...`; final rerun inside `make verify`).
- [x] `gofmt -w cmd internal`.
- [x] `go vet ./...`.
- [x] `go build ./cmd/pm`.
- [x] `make verify` (full tests, docs, local smoke, lint 0, connector validation 547/0).
- [x] `git diff --check` and dependency/scope guards.

## CLI help/manual/website parity

- [x] `pm help credentials`.
- [x] Bare `pm credentials` exits 0 with contextual manual.
- [x] `pm credentials --help`, `-h`, positional `help`, and JSON manual routes.
- [x] Invalid actions remain usage errors (built binary exit 2).
- [x] `docs/cli/credentials.md`: no update applicable; temp generated byte diff clean.
- [x] `website/**`: no update applicable; `gen:docs` wrote 11 pages and tracked diff stayed clean.
- [x] Generated/golden help artifacts: no update applicable; goldens and generated manual test pass unchanged.
- [x] Completion/discovery seam tested; Phase 15 values remain deferred.
- [x] Focused subcommand help/man churn remains deferred to Phase 19.

## Safety/scope/delivery

- [x] GSD doctor/list passed; missing programming-loop command and manual fallback recorded.
- [x] Required GSD, Go CLI/testing/error/security/safety/docs/Cobra skills loaded.
- [x] Parent draft PR #438 confirmed; this serialized worker remains local critical path.
- [x] No real secret requested, read, printed, summarized, stored, or logged.
- [x] No credentialed external connector checks or external service calls.
- [x] No interactive secret entry.
- [x] No dependency or `go.mod`/`go.sum` delta.
- [x] No connector-def, unrelated namespace, or broad generated churn.
- [x] Planning, initial RED/GREEN, correction RED/GREEN, and final evidence checkpoints included in the delivery sequence.
- [x] No PR or external review requested.
- [x] `scripts/gsd prompt verify-work 429` generated 7161 bytes and was executed inline.
- [x] `scripts/gsd prompt code-review 429` generated 6027 bytes; local review found and fixed the action-name boundary issue, then found no remaining actionable item.

Result: all applicable original verification passed at implementation head `92284dd2e55e250031389ce3673a9a6909253341`. The accepted bounded review correction also passed all declared local gates at implementation head `7970896ca7f75a6976a2a6d2d3621c45bd3338f1`; exact correction start was `758b059bbeb54032dbcd1b9a2a540ca83058861b`.
