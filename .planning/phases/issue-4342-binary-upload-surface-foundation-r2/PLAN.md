# Plan — issue #4342 binary upload CLI and certification foundation

## GSD command path

- `scripts/gsd doctor`: passed.
- `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, `code-review`: resolved to the pinned official command sources.
- `scripts/gsd prompt discuss-phase issue-4342-binary-upload-surface-foundation-r2`: generated and executed inline as recorded in `CONTEXT.md`.
- `scripts/gsd prompt plan-phase issue-4342-binary-upload-surface-foundation-r2 --tdd`: generated and executed inline as recorded in `CONTEXT.md`.
- `scripts/gsd prompt execute-phase issue-4342-binary-upload-surface-foundation-r2`, `verify-work`, and `code-review`: generated and executed inline. The canonical worker contract forbids role spawning in this direct-PR lane and no compatible isolated Pi runtime was available, so the same worker recorded the red/green evidence, verification, and review below as the documented manual-GSD fallback.

## Implementation slices

1. **Red — closed command admission.** Complete. Commandrunner tests prove raw binary, base64, and multipart required-file actions plan only through their declaration; malformed mappings and JSON actions fail closed.
2. **Green — runtime and public surface.** Complete. The schema, engine preflight, commandrunner, static validator, GitHub declaration, help, manuals, and website surface route the command into the existing approval-bound lifecycle.
3. **Red — truthful certification.** Complete. The stage test drives the actual in-process harness and proves a rejected command records a non-passing `blocked` result. A plan without a live transfer is `not_live`, never `pass`.
4. **Green — candidate, stage, and sweep.** Complete. Download remains backward-compatible at `binary`; upload is independent at `binary_upload`, with separate candidate/sweep/evidence classes. `file_upload` remains declarable but non-executable.
5. **Refactor and generated projections.** Complete. Regenerated the GitHub manual/skills/website, certification candidates/sweep, operation evidence fixed-100 projection, and current certification subject.
6. **Verify/review.** Complete. No review finding remained after manual source/diff review; the only lint finding (a conditional that staticcheck required to be a tagged switch) was corrected and its full generator suite rerun.

## CLI help/manual/website parity

- [x] `pm github releases assets upload --help` lists the declared upload source flag and approval behavior.
- [x] The same command's help resolves without a credential or project when invoked with `--help`.
- [x] Existing bare connector namespace behavior remains unchanged (`pm github` renders contextual group help at exit 0).
- [x] `docs/connectors/github/{MANUAL,SKILL}.md` and website generated CLI surface use `binary_upload` and its closed safety rule.
- [x] No `docs/cli` generic upload command is added because no generic command exists.
- [x] Generated documentation and website drift checks are run.

## Verification plan

Use `GOFLAGS='-p=3'` and one heavy suite at a time. The targeted package set is `internal/connectors/engine`, `internal/connectors/commandrunner`, `internal/cli`, `internal/connectors/certify`, and `cmd/connectorgen`, plus docs/generator checks. The full test suite and `make verify` are not run as one process in this memory-bound worktree; every omitted local gate will be recorded with its exact reason before PR creation.
