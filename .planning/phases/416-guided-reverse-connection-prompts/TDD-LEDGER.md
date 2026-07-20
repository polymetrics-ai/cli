# Phase 18 design-refinement TDD ledger

## RED

The pre-refinement contract combined reverse and setup in #416, did not assign setup a bounded
issue/PR, and lacked a single explicit rule for complete versus incomplete action invocations. The
user's manual CLI testing exposed three concrete UX failures the future implementation must cover:

- required GitHub `owner`/`repo` values surfaced first as unresolved interpolation or remote 404s;
- duplicate connection names surfaced as `internal_error` instead of correctable validation;
- long command construction was discoverable only through repeated trial and error.

The GSD UI checker supplied an additional design RED: generic `Cancel`, `Back`, `Retry`, and `Edit`
labels did not satisfy the requirement for exact consequential CTA copy.

## GREEN

| Slice | Evidence | Status |
|---|---|---|
| Interaction contract | `18-UI-SPEC.md` defines bare/help, complete/direct, incomplete/guided, invalid/error behavior | Complete |
| Security | Secret-source metadata only; placeholders validate locally; approval tokens remain hidden | Complete |
| Recovery | Duplicate names map to inspect, rename, or cancel with no overwrite | Complete |
| Automation | `--json --no-input`, optional stderr NDJSON progress, no new global agent-mode | Complete |
| Issue graph | #469 created under #416; #409/#462 block it; #417/#418 wait for it; #411→#463 drift repaired | Complete |
| GSD UI review | Revised CTA copy passed all six checker dimensions | Complete |

## REFACTOR and verification

- [x] Roadmap, issue backlog, Pi prompt, execution prompt, source plan, ADR, design docs, and skill
      use the same ownership and activation rules.
- [x] `git diff --check` passes.
- [x] repo-local skill validation passes in an isolated PyYAML validator environment.
- [x] `scripts/gsd doctor` passes with 69 commands.
- [x] `make docs-check` passes.
- [x] no production Go, dependency, generated CLI help, website, or connector definition changed.
- [ ] PR #468 is pushed and its GitHub body describes #469, the GSD UI verdict, verification, and
      remaining human review gate.

Production RED/GREEN evidence intentionally remains future work in #416 and #469; this PR changes
only documentation, planning, worker instructions, and the reusable design skill.
