# Phase 425 Verification

Session `issue-425-pi-openai-codex-gpt-5.6-sol-high-20260718T095316Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact start `479a62f930e7c8a9a51ba0b3deb088bf3aad3ecc`.

## Independent-review correction checklist

Correction session `issue-425-review-fix-pi-openai-codex-gpt-5.6-sol-high-20260718T102328Z`; model `openai-codex/gpt-5.6-sol`; thinking `high`; exact correction start `975cb21b55a32574ef754b8a0a0f0635125fb0e0`.

- [x] Focused tests added before production edits for `version --json=true`, `version --json=false`, and JSON help assignment syntax.
- [x] Exact focused RED captured at correction start (`0.568s`; all three assignment/help subtests failed as expected).
- [x] No accepted-but-disconnected `--json=<bool>` behavior remains.
- [x] Ordinary `--json`, help, unknown flags, malformed boolean assignments, deterministic output, and other namespaces do not regress.
- [x] Focused version/router/golden (`8.829s`) and global flag (`2.574s`) tests pass.
- [x] Broader `internal/cli` passed (`196.764s`).
- [x] gofmt, full vet, build, `git diff --check`, scope/dependency checks pass.
- [x] Implementation commit `fe2451cf` pushed to the existing branch; no PR or external review requested.

Correction result: focused GREEN passed in `0.562s`; deterministic/help/unknown regression focus passed in `0.579s`. Built-binary checks proved ordinary `--json` and `--json=true` byte parity, configured JSON overridden by `--json=false`, assigned JSON help, and assigned JSON on the `runtime` namespace. `--json=maybe` returned validation exit 3; an unknown version flag remained usage exit 2. Temp CLI docs generation and validation passed with no checked-in docs/website/golden delta.

## TDD and behavior

- [x] All six planning artifacts existed before production edits.
- [x] Exact RED captured (`0.612s`) before production edits.
- [x] Native `version` Cobra leaf; removed from legacy wrapper registry.
- [x] Removed obsolete version handler residual-argument parser.
- [x] Deterministic `pm version` plain output (`pm dev`, `none`, `unknown`).
- [x] Deterministic `Version` JSON envelope.
- [x] `pm help version`, `pm version --help`, `pm version -h`, and `pm version help` canonical parity.
- [x] Flag and positional JSON help return `CommandManual/version`.
- [x] Unknown flag and invalid action remain usage exit 2 and never render a manual.
- [x] No local repeated/bare-boolean version flags exist; ADR flag conventions are N/A. Global `--json` remains accepted before/after the command.
- [x] `cli.Run` signature, stdout/stderr discipline, JSON envelope, fresh-tree behavior, and golden contract preserved.

## Gates

- [x] Focused version/router: `0.553s`.
- [x] Focused version/router/golden: `7.814s`.
- [x] Full `internal/cli`: `195.315s`.
- [x] `gofmt -w cmd internal`.
- [x] `go vet ./...` (no output).
- [x] `go test -timeout 20m ./...` (pass; CLI `203.747s`, certify `355.702s`).
- [x] `go build ./cmd/pm` (no output).
- [x] `make verify` (pass; lint `0 issues`; 547 connector definitions, 0 findings).
- [x] `git diff --check`.

## CLI help/manual/website parity

- [x] Runtime help: all four text routes byte-identical; `help_bytes=350`.
- [x] Bare leaf: deterministic metadata, `plain_bytes=35`, exit 0.
- [x] JSON operation: `Version/dev`, exact fields, exit 0.
- [x] JSON manual: flag and positional forms return `CommandManual/version`.
- [x] Invalid action and unknown flag: usage exit 2; no help masking.
- [x] `docs/cli/version.md`: N/A/no update; temp `pm docs generate` + `diff -ru docs/cli` produced no diff.
- [x] Connector docs validation passed.
- [x] `website/**`: N/A/no update; `npm --prefix website run gen:docs` wrote 11 pages and `git diff --exit-code -- website` passed.
- [x] Golden/generated help: N/A/no update; focused Golden test passed and golden/docs help files have no diff.
- [x] Completion/discovery: command name unchanged; native no-file completion seam tested; Phase 15 completion additions are out of scope.

## Safety/scope

- [x] No `go.mod`/`go.sum` delta; no dependency additions.
- [x] No connector definition delta.
- [x] No unrelated namespace production files changed.
- [x] No secrets requested, read, printed, summarized, or stored.
- [x] No runtime services, credentialed connector checks, destructive/admin actions, or production deploys.
- [x] Required `make verify` local temp smoke followed plan → preview → approval → run and used only sample fixtures.
- [x] Coherent planning, RED, implementation, and verification checkpoints committed and pushed to `origin/refactor/425-version-native-cobra`.
- [x] No external review request and no PR created.
- [x] `scripts/gsd prompt verify-work ...` and local code-review prompt generated; manual review found no actionable issue in the scoped diff.

Result: full declared verification passed. Parser nativization intentionally allows Cobra/pflag diagnostics for invalid arguments while preserving usage category, JSON error shape, stdout/stderr placement, and exit 2.
