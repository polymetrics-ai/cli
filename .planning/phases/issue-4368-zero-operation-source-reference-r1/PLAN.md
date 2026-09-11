# PLAN — issue #4368 zero-operation source-reference foundation

Issue: #4368. Parent: #4292. Direct PR base: `main`. Frozen initial base:
`cf29d302c13f7fcd340d31ad6dc27872880ccf42`. Branch:
`fm/cli-zero-operation-source-reference-foundation-r1`.

## GSD lifecycle

- Inline `discuss-phase`: issue scope determines the discriminator, strict retained-source identity, non-goals, and the five exact cohort counts. No product decision remains open.
- Inline `plan-phase --tdd`: this plan and `TDD-LEDGER.md` are created before production edits.
- Inline `execute-phase`: each slice begins with its named failing focused test; no extra GSD role is spawned because the canonical delivery contract is single-worker and this runner has no compatible Pi worker runtime.
- Inline `verify-work`: record each acceptance assertion in `VERIFICATION.md`; gap-plan only if a required assertion remains red.
- Inline `code-review`: conduct a deep changed-path, source-integrity, security, and command-boundary review before PR publication.

## TDD slices

1. **Closed zero-operation representation.** Add RED fixtures proving that only an explicit rendered-reference coverage declaration with ordinary retained source evidence and manifest proof is accepted. Prove malformed/missing marker, missing publication/capture/manifest, mismatch, duplicates, accidental empty, and mixed valid/invalid documents fail with a document location. GREEN only the narrow validator and retained-artifact plumbing; do not alter OpenAPI/source-reference or non-empty rendered bytes.
2. **Importer/projection/evidence closure.** Add RED tests that a marked empty coverage document is read and integrity-checked yet produces no descriptor operation and cannot silently bypass a bad peer document. GREEN the source-import/projection/evidence path with deterministic document ordering and duplicate checks.
3. **Five-cohort evidence fan-in.** Import only real retained source bytes, locks, manifests, source descriptors and generated source-accounting artifacts required for Amplitude 187, Dremio 49, Ashby 193, Workable 84, and HiBob 207. Add exact 720 reconciliation tests that require each row's citation, provider operation ID, stable identity, six-lane classification, and exactly one named `missing_foundation` disposition.
4. **Runtime boundary.** Add RED commandrunner/registry checks for each declared deferred representative: structured `missing_foundation` before credential/provider work. Preserve a representative pre-existing runnable command's missing-credential boundary. GREEN only declared deferred source-target wiring; no generic transport or fabricated provider action.
5. **Refactor and review.** Keep parsing and marker validation local, run generated checks and scoped repository gates, compare non-empty lock fixture bytes/counts, and audit that no unrelated connector/runnable operation regresses.

## Verification plan

- `go test -timeout 20m ./cmd/connectorgen -run 'TestSourceImport.*(Zero|Rendered|SourceReference)|TestSourceRetain.*(Zero|Rendered)' -count=1` for RED/GREEN then the full `./cmd/connectorgen` package.
- Targeted `internal/connectors/commandrunner`, `internal/connectors/engine`, and `internal/cli` checks; build `./cmd/pm` and run metadata/help-only probes.
- `go run ./cmd/connectorgen source-import <each target> --check`, source projection/evidence, `surface-sync --check`, `validate`, and duplicate/identity invariant scripts for five cohorts.
- `gofmt`, `go vet ./...`, `go build ./cmd/pm`, `git diff --check`; individual make gates specified by AGENTS.md rather than a timeout-prone broad local suite.

## Checkpoints

1. Planning evidence committed.
2. RED test commit (or recorded RED immediately before a coherent green slice).
3. Green foundation and bounded cohort-evidence commit.
4. Review/verification correction commit, rebase to latest `origin/main`, final exact-head verification, push and direct PR.
