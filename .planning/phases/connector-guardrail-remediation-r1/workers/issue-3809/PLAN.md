# PLAN — issue 3809 curated icon-collapse authority

## Scope and ownership

- Issue: #3809 — `fix(iconregistrygen): let curated rows resolve upstream collapse conflicts so icon generation can complete`
- Branch: `fm/cli-found-icon-registry-generator-r1`
- Parent/sub-issue evidence: `gh-axi issue subissue list 3809` reported zero children. No parent relationship is exposed by the issue wrapper, so this is a single-issue branch and PR.
- Primary production ownership: `cmd/iconregistrygen/main.go`, `cmd/iconregistrygen/main_test.go`, and generated `internal/connectors/icon_data.json` only.
- Explicit non-goals: do not alter `internal/connectors/icons.go` or weaken `MustValidateIconCoverage`; do not add Shopify data; do not alter Simple Icons fetch/lockfile behavior; do not make provider or credentialed calls; do not touch command-runner or connector engine paths.

## GSD lifecycle and fallback

- Checked `scripts/gsd doctor`; resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` through `scripts/gsd sources`; `go run ./cmd/agentcontractgen check` passed.
- Generated the discussion trace with `scripts/gsd prompt discuss-phase 3809 --auto` and ran `gsd-sdk query init.phase-op 3809`. The adapter reports `phase_found: false`, because #3809 has no roadmap phase.
- The normal numbered-phase workflow therefore cannot create its artifacts. Use this established icon-registry worker directory as the documented inline/manual fallback. It preserves the same order: discussion/contract review → this plan and TDD ledger → red test → implementation/regeneration → verification → code review. No roles or subagents are spawned.
- Before each later lifecycle stage, generate its corresponding `scripts/gsd prompt ... 3809` trace and record the phase-not-found fallback plus actual evidence in `VERIFICATION.md` and the PR body.

## Required skills and safety constraints

- Loaded: `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `no-mistakes`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review`.
- Load `golang-lint` before the code-review stage and `javascript-testing-patterns` before interpreting the website script test results.
- Preserve the single-source policy in `docs/migration/icon-registry-single-source.md`: the canonical bare-key registry is authored state; raw source/destination prefix collapse applies only to upstream records; the shared-asset-path/source-URL conflict remains fatal.
- Never log or request secrets. Use public fixtures or the documented public registry artifact only. No live provider calls and no new dependencies.

## Locked implementation decisions

1. Index every supplied curated registry entry by its bare connector key before processing upstream records, rejecting duplicate keys and orphan entries exactly as today.
2. Index every supplied curated registry row before processing upstream records and make that authored row authoritative for its bare connector. Skip raw upstream records for any existing curated key, then add the curated rows to the final registry. This applies even when its provenance is `upstream_registry`/`upstream_seeded`: provenance describes source, not whether the row is authored registry state. This lets the same authored authority settle both a prefix-collapse disagreement and the current shared-asset disagreement without selecting either conflicting upstream URL.
3. Keep an upstream-only collision fatal. Preserve the existing ID/path conflict rule and enhance a conflicting-URL error to name both upstream record identifiers, both URLs, and the bare curated key an operator can author to resolve it. Never select a URL by order.
4. Keep final shared-asset validation unchanged: different connectors sharing one path still require identical non-empty source URLs.
5. Regenerate `internal/connectors/icon_data.json` by invoking `make icons-generate` with the public upstream registry source; no hand edits to that file are permitted.

## TDD slices and checkpoints

1. **Red test checkpoint.** Add a focused test with conflicting `source-demo`/`destination-demo` URLs plus a pre-existing curated bare `demo` row. Assert that the run returns the curated row and does not expose either conflicting URL. Run only that test before production code and record its failure.
2. **Diagnostic-preservation checkpoint.** Strengthen the existing no-curated collision test so it asserts the error contains both raw record identifiers, both URLs, and `demo` as the curated key. This is a deliberate change to an old assertion because its previous diagnostic omitted required operator context; it does not relax refusal.
3. **Green implementation checkpoint.** Refactor `buildIconEntries`/collapse provenance minimally so curated authority is known before upstream merging and unresolved diagnostics retain raw record identity. Run all `./cmd/iconregistrygen` tests.
4. **Generation checkpoint.** Run the documented `make icons-generate` command using the current public upstream registry source, inspect the generated-only diff, and keep only generator-produced `internal/connectors/icon_data.json` changes.
5. **Verification/review checkpoint.** Run focused Go, icon-coverage, and website script checks; then the scoped `make verify` gates listed in `VERIFICATION.md`. Commit only coherent green slices. Do not start no-mistakes, push, or open a PR until firstmate directs that delivery stage.

## Expected changed paths

- `cmd/iconregistrygen/main.go`
- `cmd/iconregistrygen/main_test.go`
- `internal/connectors/icon_data.json` (only generator output)
- `.planning/phases/connector-guardrail-remediation-r1/workers/issue-3809/{PLAN.md,TDD-LEDGER.md,VERIFICATION.md}`

## Acceptance checklist

- A curated `customer-io`-shaped row resolves an otherwise conflicting upstream source/destination collapse.
- An uncurated URL conflict still fails with both record names, both URLs, and a named bare curated key.
- Existing bare-key, curated-attribution, orphan-curated-entry, fallback-row, and shared-path conflict tests remain green.
- `make icons-generate` completes, emits coverage for every compiled connector, and changes `icon_data.json` only through the generator.
- `pnpm run test:scripts` continues to enforce exact registry-to-lockfile coverage.
