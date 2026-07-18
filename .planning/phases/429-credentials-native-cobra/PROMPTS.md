# Phase 429 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#429 as the ninth serialized Phase 9 unit under #407/#397 from exact parent HEAD `0f1ec1e89cdae761e9da06ab9906fcc641b38e0a`, using isolated branch `refactor/429-credentials-native-cobra`, with Sol/high explicit, no PR, and no external review.

Identity: session `issue-429-pi-openai-codex-gpt-5.6-sol-high-20260718T143346Z`; explicit model `openai-codex/gpt-5.6-sol`; explicit thinking `high`; start `20260718T143346Z` UTC.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 429 --skip-research
scripts/gsd prompt programming-loop init --phase 429 --dry-run
```

Doctor/list passed and the plan-phase prompt was generated for inline execution. Programming-loop is absent (`scripts/gsd: unknown GSD command: programming-loop`, exit 1), so the manual GSD universal-runtime-loop fallback enforces plan-before-production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with later units; this session has no subagent tool; user restricted the invocation to #429 and prohibited PR/external review.

Required skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-documentation`; `golang-spf13-cobra`.

Safety prompt: Never request, read, print, summarize, store, or log real secret values. Tests may use opaque synthetic redaction fixtures only and must not include fixture content in diagnostics. Do not add interactive secret entry, dependencies, credentialed checks, external services, or unrelated changes. Preserve env/stdin-only intake, path containment, error/output contracts, action boundaries, and legacy help/literal compatibility.

Downstream artifact: test-only initial and correction RED checkpoints, `internal/cli/credentials_cli.go`, credentials router/legacy-handler adaptation, focused security tests, and finalized six issue-local phase artifacts.

Verification result: pass. Initial focused GREEN passed. Local review then exposed a post-action name-discovery bypass; correction RED failed 8/10 cases before the fix, and corrected focused/repeated/race tests pass. Final full CLI (`275.269s`), preserved differential 28/28, built help/docs parity, website generation, gofmt, vet, build, full repository tests, and `make verify` pass. No real secret, credentialed external check, service, dependency, PR, or external review occurred.

## Normalization-order correction snapshot

Task: Accept the MEDIUM finding in `/tmp/pm-397-rereview7-429.log` from exact start `2013d361c949395f41bb6a30209d65cbec4a62c2`. Session `issue-429-normalization-order-correction-pi-openai-20260718T215811Z`.

GSD route: doctor/list passed; `programming-loop` remains unavailable, so the manual universal runtime loop is active. Execution decision: `local_critical_path` because this is one credentials normalization-order correction, no subagent tool is exposed, and the user prohibited services, dependencies, PR, and review.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`. CLI help/docs/website parity policy is active.

Downstream artifact: metadata-only RED coverage for single-hyphen spaced StringArray values after known-flag-shaped names; capture → normalize → filter correction; preserved parser/security protections; finalized phase evidence.

Verification result: pending strict RED, GREEN, focused/protection/repeated/race, exact parent-base/start/head differential, full CLI, gofmt, vet, build, and diff/scope/dependency gates.

## Targeted final parser-order correction snapshot

Task: Accept the MEDIUM finding in `/tmp/pm-397-rereview6-429.log` from exact start `4870f53b028613fbc3632a404e9a382419d87f8a`. Session `issue-429-targeted-parser-order-correction-pi-openai-20260718T212111Z`.

GSD route: doctor/list passed; `programming-loop` remains unavailable, so the manual universal runtime loop is active. Execution decision: `local_critical_path` because this is one credentials parser-order correction, this runtime exposes no subagent tool, and the user prohibited private output, services, dependencies, PR, and review.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-troubleshooting`, `golang-error-handling`, `golang-security`, and `golang-safety`. CLI help/docs/website parity policy is active.

Downstream artifact: strict four-name lifecycle RED coverage with later real flags and ignored positionals; preserved raw-carrier/action protections; required-name capture before tail normalization; finalized phase evidence.

Verification result: pass. Strict RED captured all four regressions while preservation guards stayed green; focused GREEN passed after first-name capture moved before tail normalization. Focused/adversarial (`79.397s`), repeated ×5 (`176.825s`), race (`392.438s`), parent-base/start/head differential, full CLI (`340.707s`), help parity, gofmt, vet, readonly build, diff/scope/dependency gates passed. No private output, services, dependencies, PR, or review.

## Bounded review correction snapshot

Task: Accept all findings in `/tmp/pm-397-review-429.log` from exact start `758b059bbeb54032dbcd1b9a2a540ca83058861b`. Session `issue-429-bounded-security-compat-correction-pi-openai-codex-gpt-5.6-sol-high-20260718T155702Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`.

Execution: `local_critical_path`; this Pi runtime exposed no subagent tool, the accepted findings shared credentials/router boundaries, and the user prohibited PR/external review. GSD doctor/list passed; the adapter's documented programming-loop command remained absent, so the existing manual universal-loop fallback was used. Skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-lint`.

Downstream artifact: test-first symlink-effect, legacy-name, and help-tail corrections in `internal/safety`, `internal/app`, and `internal/cli`, plus finalized issue-local phase evidence.

Verification result: pass. RED reproduced all findings in `6.546s`. Focused, repeated, race, security/path, exact base differential, full CLI/repository, built help, golden/manual, gofmt, vet, build, and `make verify` pass. No secret material, real credential/service, dependency, PR, or external review.

## Second bounded correction snapshot

Task: Accept all three findings in `/tmp/pm-397-rereview-429.log` from exact start `fae7d599668637bea345fe76877dd75e31dd2ad8`. Session `issue-429-second-bounded-correction-pi-openai-20260718T170705Z`; profile `Sol/high`.

GSD route: doctor/list passed; `scripts/gsd prompt plan-phase 429 --skip-research` generated 10692 bytes; `programming-loop` remains unavailable, so the manual universal runtime loop is active. Execution decision: `local_critical_path` because the correction shares app/connectors/router safety seams, this runtime exposes no subagent tool, and PR/external review are prohibited.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-troubleshooting`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint`. CLI help/docs/website parity policy is active.

Downstream artifact: strict RED tests, actual-state assertion, runtime-only local-path normalization and non-secret effect policy, Warehouse/Outbox/app effect validation, bounded Cobra name carrier, and finalized phase evidence.

Verification result: pass. Focused, repeated, race, all-credentials, app, connectors, CLI, exact-start differential, full repository, gofmt, vet, build, and `make verify` passed. Full repository app/CLI/certify timings were `27.976s`/`285.504s`/`340.518s`; lint 0 and connector validation 547/0. No real credential, private fixture output, external service, dependency, PR, or review.

## Compatibility correction snapshot

Task: Accept the MEDIUM finding in `/tmp/pm-397-rereview5-429.log` from exact start `7497483de2187b3117c32b9cafb3db54ebac792f`. Session `issue-429-compatibility-correction-pi-openai-20260718T202616Z`.

GSD route: doctor/list passed; `scripts/gsd prompt programming-loop init --phase issue-429 --dry-run` remains unavailable, so the manual universal runtime loop is active. Execution decision: `local_critical_path` because this is one credentials validation/router correction, this runtime exposes no subagent tool, and services, PR, and review are prohibited.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-troubleshooting`, `golang-error-handling`, `golang-security`, and `golang-safety`. CLI help/docs/website parity policy is active.

Downstream artifact: strict add/inspect/remove RED coverage for every requested class of safety-valid short leading-hyphen and double-hyphen names; preserved raw-carrier rejection and action/name ownership; ordinary credential validation for private carriage; finalized phase evidence.

Verification result: pass. Strict RED failed as required in `23.030s`; focused GREEN passed in `56.416s`; repeated ×5 passed in `352.467s`; split compatibility/adversarial race passed in `457.137s`/`262.781s` after the aggregate command timed out at 600 seconds without a failure. Parent-base/start/head differential covered 14 names and 42 operations per lifecycle, full CLI passed in `333.259s`, and help parity, gofmt, vet, build, diff, scope, and dependency gates passed. No private data output, service, dependency, PR, or review.

## Final bounded correction snapshot

Task: Accept the MEDIUM and LOW findings in `/tmp/pm-397-rereview4-429.log` from exact start `80246e42f508f685d281fecbcc3735eadcf271a9`. Session `issue-429-final-bounded-correction-pi-openai-20260718T194756Z`.

GSD route: doctor/list passed; `programming-loop` remains unavailable, so the manual universal runtime loop is active. Execution decision: `local_critical_path` because both corrections are on the credentials/app critical path, this runtime exposes no subagent tool, and services, PR, and external review are prohibited.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`. CLI help/docs/website parity policy is active.

Downstream artifact: strict leading-hyphen add/extra-positional and overwrite final-open cleanup RED tests; private first-token validation and ownership correction; immediate overwrite raw-temp cleanup registration; finalized phase evidence.

Verification result: pass. Strict RED reproduced both findings. Focused, repeated ×5, race CLI/app/localwrite/connectors, exact parent-base/start/head differential (`0/1/0`, exact base/head output), full relevant packages, help parity, gofmt, vet, build, diff/scope/dependency guards passed. Planning/RED/GREEN checkpoints are pushed and final evidence is prepared. No private data display, service, dependency, PR, or external review occurred.

## Fourth bounded correction snapshot

Task: Accept the HIGH finding in `/tmp/pm-397-rereview3-429.log` from exact start `0d70335f37456f42432b3c502860f7b43231ed98`. Session `issue-429-fourth-bounded-correction-pi-openai-20260718T185126Z`.

GSD route: doctor/list passed; `programming-loop` remains unavailable, so the manual universal runtime loop is active. Execution decision: `local_critical_path` because this is one shared local-write effect seam, this runtime exposes no subagent tool, and PR/external review are prohibited.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint`. CLI help/docs/website parity policy is active.

Downstream artifact: temp-only Warehouse/Outbox/app final-link RED tests; an `os.Root`-backed local-write effect helper; all relevant final opens/renames migrated to the helper; finalized phase evidence.

Verification result: pass. All 11 strict RED cases failed before production edits. Focused ×5 (`42.12s`), race (`84.54s`), broader connectors/app/CLI (`350.01s`), full repository (`347.88s`), gofmt, vet, Go 1.25 build, and `make verify` (`374.34s`) passed; lint 0 and connector validation 547/0. Planning/RED/GREEN checkpoints are pushed and final evidence is prepared. No private data display, service, dependency, PR, or external review.

## Third bounded correction snapshot

Task: Accept the MEDIUM finding in `/tmp/pm-397-rereview2-429.log` from exact start `6158cdc92d5df01cbaa577ceeb5a870ddcb8f685`. Session `issue-429-third-bounded-correction-pi-openai-20260718T180016Z`.

GSD route: doctor/list passed; `programming-loop` remains unavailable, so the manual universal runtime loop is active. Execution decision: `local_critical_path` because this is one credentials/router trust-boundary correction, this runtime exposes no subagent tool, and PR/external review are prohibited.

Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, and `golang-safety`. CLI help/docs/website parity policy is active.

Downstream artifact: strict 12-case RED matrix for raw assigned/bare/spaced internal carrier use across add/inspect/test/remove; inaccessible command-local leading-name carriage; raw-carrier fail-closed guard; finalized phase evidence.

Verification result: pass. RED failed 9/12 contract cases before production edits (`11.651s`). Focused/adversarial (`34.099s`), repeated ×5 (`56.733s`), race (`273.254s`), seven-case exact preserved differential, 12-case corrected raw differential, full CLI (`332.836s`), gofmt, vet, build, diff, scope, and dependency gates passed. No private data display, real credential, service, dependency, checked-in docs/website delta, PR, or review occurred.

## Verification and local review snapshot

```bash
scripts/gsd prompt verify-work 429 > /tmp/gsd-verify-work-429.prompt
scripts/gsd prompt code-review 429 > /tmp/gsd-code-review-429.prompt
```

Prompt generation passed (7161 and 6027 bytes). Both were executed inline under the manual universal loop. The first local security review produced the action-name boundary correction described in PLAN/TDD-LEDGER; the post-fix diff/error/security/safety review found no remaining actionable issue. External review was prohibited.
