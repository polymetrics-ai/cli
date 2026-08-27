# Plan — issue 4364 deferred visibility/admission bridge

**Mode:** TDD. **Execution:** inline/manual GSD fallback; the available Codex runtime has no compatible Pi isolated-role executor and Firstmate's assigned-worker instruction forbids agent dispatch.

See [CONTEXT.md](CONTEXT.md) for the completed task delivery header, evidence table, required skills, source/base, and safety decisions.

## Scope boundary

This shared foundation consumes the reviewed `batch1-source-operation-mapping-manifest.json` and produces source-cited deferred visibility/admission artifacts. It deliberately does not carry over historical connector executor implementations, fabricate typed lane records, access credentials, invoke a provider, or change source locks. The preserved branch is read only to import the authoritative manifest; the branch base remains `origin/main` at `cf29d302c`.

## Slices

1. **Manifest contract and red reconciliation tests.** Add a closed schema/loader for the manifest and tests that currently fail because `main` has no 4,341-record input or 1,910 generated deferred declarations. Cover positive totals and every invalid form: duplicate identity, missing HTTPS citation/location, unknown/policy-only foundation, generic URL/HTTP target, fake lane binding, and multiple primary gaps.
2. **Generator and generated declaration bridge.** Materialize per-provider blocked `api_surface.json` records, deferred `cli_surface.json` commands, declaration-admission inventory/source/declaration entries, exact source target ledger rows, and operation-evidence rows. Use stable manifest paths; prevent writes to absent typed executor artifacts. Regenerate only the named source/canonical/generated artifacts.
3. **Runtime admission/preflight.** Extend bundle/admission validation and deferred command projection so every generated deferred command has a compact embedded exact target and returns its named `missing_foundation` terminal before transport setup or credential resolution. Keep existing implemented preflight as a stale-declaration guard.
4. **Parity and documentation.** Update declaration-admission and operation-evidence canon to explain manifest-derived deferred visibility. Add reconciliation and public CLI tests for count/lane/component rollups, exact errors, no-I/O, and retained source classes. Regenerate relevant manual/website catalog outputs only if the shared command discovery generator identifies them as affected.
5. **Verification/review.** Run the exact generator, validation, docs, build, targeted runtime, commandrunner, connector-boundary, and CI-equivalent non-full-suite gates. Rebase/integrate current `origin/main` before opening the PR, request automatic review, obtain a fresh independent exact-head Codex audit, disposition findings, and verify the PR API base is `main`.

## TDD plan

| Slice | Red | Green | Refactor / regression |
| --- | --- | --- | --- |
| Manifest closure | Add named table tests for the 4,341/1,910 invariant and malformed records; they fail because no manifest loader/materialization exists. | Implement a closed manifest decoder/validator that emits exact diagnostic paths and checks once-only source identity, citation, lane, foundation, and path. | Sort deterministically, share endpoint/citation validation with existing admission code, and retain fixture isolation. |
| Generated bridge | Assert one representative per provider/lane is missing a blocked API row, deferred command, admission rows, and evidence rollup; fail on current `main`. | Generate the named JSON artifacts from the manifest with a source-cited exact target and one foundation; leave typed runtime artifacts unchanged. | Compare snapshots by semantic identity and ensure `--check` detects all output drift. |
| Preflight ordering | Assert an admitted generated command reaches `missing_foundation:<foundation>` with zero request/credential calls; fail on the absent catalog. | Use the existing typed deferred preflight seam with a manifest-produced target and exact foundation identity. | Add stale/duplicate/swapped/implemented target regressions and preserve the existing missing-credential path for real commands. |
| Accounting | Assert exact provider/lane/component totals and source-class visibility; fail against the empty/generated-current catalog. | Project every manifest row into operation evidence and declaration admission, including delete/reverse/binary/ETL/unsafe/mutation rows. | Keep counts calculated, never hard-coded outside the manifest invariants; document the visibility versus runnable delta. |

## Verification plan

- `go test -timeout 20m ./cmd/connectorgen -run '^(TestBatch1Deferred|TestDeclarationAdmission|TestOperationEvidence)'`
- `go test -timeout 20m ./internal/connectors/engine -run '^(Test.*Deferred|Test.*Declaration)'`
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^(Test.*Deferred|Test.*Preflight)'`
- `go test -timeout 20m ./internal/app -run '.*Deferred.*'` and the affected `internal/cli` tests separately.
- `go run ./cmd/connectorgen declaration-admission --json`, `operation-evidence --check`, `surface-sync --check`, `validate internal/connectors/defs`, and `surface-reconcile --check` when compatible with the generated output.
- Individually run `make` gates named by repository policy: tidy-check, lint, docs-check, smoke-no-build, agent-contract-check, connectorgen-validate, connectorgen-surface-sync, connector-boundary, release-workflow-check; build `./cmd/pm`.
- Verify `pm help <topic>`, affected `pm <connector>`, and `pm <connector> <command> --help`, plus documentation/catalog generator output or a recorded not-applicable determination.
- `git diff --check`, `git status --short`, and PR API base read-back after opening.

