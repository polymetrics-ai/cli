# Phase 429 Summary

Status: normalization-order correction complete and verified from exact head `2013d361c949395f41bb6a30209d65cbec4a62c2`; MEDIUM spaced StringArray value finding closed; no private output, services, dependencies, PR, or review.

## Normalization-order correction

Planning, TDD, verification, prompt, summary, and run-state artifacts were reopened before test or production edits. Strict RED then failed all four known add-flag-shaped name cases in `18.206s` (wall `20.584s`) while raw-carrier and invalid action/name ownership protections stayed green; assertions inspected metadata structure and printed no values. The router now rejects raw carriers, privately captures/removes the first name without filtering, normalizes spaced StringArray values, and only then filters the legacy tail. Focused GREEN passed in `20.167s` (wall `22.347s`).

Focused/protection (`85.690s`), repeated ×5 (`100.438s`), race (`223.261s`), exact parent-base/start/head differential, full CLI (`343.447s`), help parity, gofmt, vet, build, diff, scope, and dependency gates passed. Parent base/head preserved metadata 4/4, exact start mismatched 4/4, and eight base/head output pairs matched exactly. Implementation head: `2df70e29c7e924eec1404eb0d208ff3a41320ed0`. All first-name, action-discovery, carrier, help/global, compatibility, and redaction protections remain intact.

## Targeted final parser-order correction

Planning was updated before test or production edits. Strict RED exercised safety-valid names equal to all known add flags (`--connector`, `--from-env`, `--value-stdin`, `--config`) with an immediately following ignored positional and later real spaced flags through add/inspect/remove. All four adds failed because StringArray normalization changed the required name into an invalid `name=ignored` token (`18.545s`, wall `21.75s`), while raw-carrier rejection and invalid action/name ownership stayed green. The implementation now captures/removes the required add name before StringArray space-value normalization and normalizes only the tail. Focused lifecycle/raw-carrier/invalid-ownership GREEN passed in `28.076s` (wall `30.55s`).

Focused/adversarial (`79.397s`), repeated ×5 (`176.825s`), race (`392.438s`), parent-base/start/head differential, full CLI (`340.707s`), help parity, gofmt, vet, readonly build, diff, scope, and dependency gates passed. Parent base/head each passed 12 lifecycle operations, start rejected four adds, base-seeded start/head each passed eight inspect/remove operations, and eight base/head add/remove output pairs matched exactly. Implementation head: `9e87a007e4331d1afee7c66b4b079eb3694f3d8d`. No private output, service, dependency, checked-in docs/website delta, PR, or review.

## Compatibility correction

Planning completed before production edits. Strict RED covered 14 safety-valid short/double-hyphen names through config-only add/inspect/remove: all were rejected by private validation in `23.030s`, while raw internal-carrier rejection and invalid action/name ownership stayed green. The private validator is removed; privately carried names use ordinary credential identifier validation. Discovery and normalization are unchanged.

Focused GREEN (`56.416s`), repeated ×5 (`352.467s`), split race compatibility/adversarial (`457.137s`/`262.781s`), exact parent-base/start/head differential, full CLI (`333.259s`), help parity, gofmt, vet, build, diff, scope, and dependency gates passed. The aggregate race command timed out at 600 seconds without a failure before both exact partitions passed. Differential covered 14 names: base/head each passed 42 lifecycle operations, correction start rejected 42 operations, and 28 base/head add/remove outputs matched exactly. Implementation head: `199b802c4d2be7e62e335549a8c56dc0804c909b`. No private data output, service, dependency, checked-in docs/website delta, PR, or review.

## Final bounded correction

Strict RED first showed the leading-hyphen add compatibility case exiting 1 and both non-deduped overwrite final-open failures leaving raw temps. The router now privately owns every leading-hyphen first required-name token independent of later positionals, rejects raw carriers first, and validates that private token before action execution. Valid legacy `-legacy` names preserve later typed/global flags and ignored positionals; invalid flag-like first tokens cannot discover later names. App overwrite cleanup now starts immediately after raw-temp open, covering final-temp open failure.

Focused, repeated ×5, race CLI/app/localwrite/connectors, exact base/start/head differential, full relevant packages, help parity, gofmt, vet, build, diff, scope, and dependency checks passed. Differential exits were parent base `0`, correction start `1`, final head `0`, with byte-identical base/head output. Implementation head: `74e8cffe477ce713526963c6fd4cb37dcc973b84`. No private data display, service, dependency, checked-in docs/website delta, PR, or external review.

## Fourth bounded correction

Directory-only validation did not confine a final `os.OpenFile`: existing or dangling final JSONL symlinks could append, truncate, or create outside the selected local-write root. All 5 Warehouse/Outbox and all 6 app temp-only cases failed before production edits. `safety.LocalWriteFS` now holds a Go 1.25 `os.Root` and performs confined directory creation, final opens, app raw reads/cleanup, and raw/final renames beneath the selected root at effect time. Explicit external opt-in and nil-policy compatibility remain ordinary OS effects by design.

Focused ×5 (`42.12s`), race (`84.54s`), broader connectors/app/CLI (`350.01s`), full repository (`347.88s`), gofmt, vet (`3.22s`), Go 1.25 build (`1.81s`), and `make verify` (`374.34s`) passed. Lint reported 0 issues and connector validation 547/0. Modes, append/overwrite, nonexisting paths, in-root relative symlinks, safe rename replacement, and explicit external behavior are covered. Implementation head: `bc13b768d03f27f87f1f6bc262edf890925d58a7`. No private fixture display, service, dependency, checked-in CLI docs/website delta, PR, or external review.

## Third bounded correction

The hidden `--pm-internal-credentials-name` pflag was raw-user-addressable and could supersede the required first positional credential name. The 12-case add/inspect/test/remove × assigned/bare/spaced RED matrix failed before production edits in `11.651s`: assigned/spaced raw forms overrode every action and exited 0, while add/bare returned runtime code 3 rather than usage 2. The hidden pflag is now removed. A private command-context value carries normalized leading-hyphen names, while exact raw carrier spellings are rejected before Cobra parsing. Focused/adversarial (`34.099s`), repeated ×5 (`56.733s`), and focused race (`273.254s`) tests prove fail-closed behavior, unchanged records, no synthetic value output, and preservation of leading-hyphen names plus normal flags.

Seven unaffected exact-base cases matched exit/stdout/stderr byte-for-byte; every one of the 12 current raw carrier differential cases exited usage 2. Full CLI passed in `332.836s`; gofmt, vet, build, diff, scope, and dependency checks passed. Implementation head: `30875076c7cdb172727ffb506c10fb628dd3007c`. No private data display, service, dependency, checked-in docs/website delta, PR, or review.

## Second bounded correction

All three findings from `/tmp/pm-397-rereview-429.log` are closed. Plan/TDD/verification/run-state artifacts were reopened before production edits. Strict RED reproduced selected-root relative misses, post-resolution Warehouse/Outbox external effects, and leading-hyphen add failure; the state helper now requires the actual state file.

Relative local runtime paths now resolve beneath the selected root without changing persisted credential config. An optional non-secret runtime policy carries the selected root and explicit external opt-in; Warehouse/Outbox `Check` and `Write` plus app warehouse materialization validate it immediately before directory effects. Nil-policy direct connector calls retain compatibility. A hidden internal Cobra carrier preserves a safety-valid leading-hyphen first name while parsing later flags; suspicious later positional names remain fail-closed.

Focused, repeated, race, app, connectors, CLI, five-case exact-start preserved differential, full repository, gofmt, vet, build, and `make verify` passed. Full repository app/CLI/certify timings were `27.976s`/`285.504s`/`340.518s`; lint 0, connector validation 547/0, and built credentials help remained byte-identical. Implementation head: `ec7064a851e572feb8cffdde2c394917ad38662c`.

## Bounded review correction

From exact correction start `758b059bbeb54032dbcd1b9a2a540ca83058861b`, session `issue-429-bounded-security-compat-correction-pi-openai-codex-gpt-5.6-sol-high-20260718T155702Z` (`openai-codex/gpt-5.6-sol`, high) accepted all findings from `/tmp/pm-397-review-429.log`.

- Symlink-resolved local write paths now use nearest-existing-ancestor realpath containment and are revalidated immediately before resolved credentials reach connector effects. Warehouse/outbox tests prove no external directory appears without opt-in; `allow_external_path=true` remains effective.
- Safety-valid legacy credential names beginning `_`, `.`, or `-` remain inspectable/removable. Connector-name hardening remains.
- Long/short credentials namespace help ignores trailing unknown flags and byte-matches exact base `0f1ec1e8`; correction start exited 2.
- Strict RED preceded production edits. Focused/repeated/race/path/security/full CLI and repository, help/manual, gofmt/vet/build, and `make verify` pass. Implementation head: `7970896ca7f75a6976a2a6d2d3621c45bd3338f1`.
- No real credential, secret material, service, dependency, PR, or external review.

## Identity

- Session: `issue-429-pi-openai-codex-gpt-5.6-sol-high-20260718T143346Z`
- Model/thinking: `openai-codex/gpt-5.6-sol`, `high`
- Branch: `refactor/429-credentials-native-cobra`
- Exact start: `0f1ec1e89cdae761e9da06ab9906fcc641b38e0a`
- Parent: #397; umbrella: #407; draft parent PR #438

## Local security correction

A post-implementation local review found that Cobra could consume an invalid first name token after an exact add/remove action and discover a later name. Test-first correction reproduced eight bypasses. A required-name literal boundary now preserves the first token as the name, and credential/connector names must begin with an ASCII alphanumeric character. Focused, repeated, race, and golden correction gates pass; no secret source or external action was used.

## Delivered in focused GREEN

- Native Cobra ownership for credentials add/list/inspect/test/remove/help; only the credentials legacy parser call is removed.
- Typed repeated current flags with exact legacy bare/assigned/unknown/trailing-help/literal behavior.
- Controlled env/stdin-only secret intake through Cobra input; no interactive entry.
- Strict identifiers, pre-read source/config validation, existing path-containment behavior, output redaction, and fail-closed action discovery.
- Focused credentials/router and focused race tests pass; golden passes; 28/28 preserved differential cases match exact start behavior.

## TDD and verification

Initial RED was the missing native constructor before production edits. Initial focused GREEN passed. Local review then added a focused correction test that failed eight add/remove name-discovery cases before the boundary fix. Corrected focused (`40.299s`), repeated (`62.622s`), race (`248.367s`), golden (`5.602s`), and full CLI (`275.269s`) gates pass.

A 28-case start-vs-head differential matches exact exit/stdout/stderr for preserved help, list, add flag forms, unknown/extra inputs, tail help, literal separator, invalid namespace heads, and globals. Built help routes are byte-equal; temporary CLI docs generation matches `docs/cli`; connector docs validate; website generation writes 11 pages with no tracked diff. gofmt, vet, build, full repository tests, and final `make verify` pass (CLI `278.385s`, certify `342.715s`, lint 0, 547 connector definitions/0 findings).

## Workflow

GSD doctor/list/plan-phase prompt succeeded. The adapter has no `programming-loop` command, so the recorded manual universal-loop fallback enforced plan/TDD/verification. All six artifacts existed before test or production edits. Verify-work and code-review prompts generated (7161/6027 bytes) and ran inline; the boundary finding was fixed test-first and post-fix local review is clean. Execution remained `local_critical_path`; no subagent tool or external review was used.

## Safety

No real secret value was requested, read during agent-run checks, printed, summarized, stored, or logged. Opaque synthetic fixtures were confined to focused tests and plaintext scans. No interactive secret entry, credentialed external connector check, optional service, dependency, unrelated namespace, connector definition, checked-in docs/website/golden change, PR, external review, or merge occurred. Final `make verify` used its existing local temporary-root sample and retained plan → preview → approval → execute.

## Delivery

Pushed checkpoints:

- `cc1c13c5` — planning
- `eefbfdfa` — initial RED
- `36b2e388` — native implementation
- `3a5bdd25` — action-name discovery RED
- `92284dd2` — action-name boundary fix
- final verification artifact checkpoint

Verified implementation head: `92284dd2e55e250031389ce3673a9a6909253341`; verification ended `20260718T153350Z` UTC. No PR was created.
