# Issue #4193 discussion context

## Fixed decisions

- Issue: `Refs #4193 — fix(cli): leaf --help must render before project, credential, or required-flag resolution`.
- Base branch: `integration/4015-mvp-flat-r1` at `ff6a87101`, confirmed before production edits.
- Delivery: direct PR from `fm/cli-command-usability-scalable-r1` to the required integration base.
- Scope is the legacy Cobra-wrapper path in `internal/cli`; do not touch certification or transport-owned paths.
- Preserve approval-carrier validation before dispatch: malformed approval syntax remains a real safety error even when a help flag is present.
- Resolve ordinary leaf help generically before a wrapper invokes `withApp`, a credential resolver, a required-flag check, runtime doctor, documentation generation, or any other handler effect.

## Binary inventory and boundary

The initial binary built from `ff6a87101` reports 17 documented core namespaces and
37 dynamic connector namespaces. The connector catalog contains 556 names and the
declaration-owned provider surface contains 1,571 commands. Dynamic connector help
already checks `connectorHelpRequested` before `withApp`; this issue's regression is
the 21 legacy Cobra wrappers and their leaf paths.

The exhaustive core leaf inventory is recorded in `VERIFICATION.md`. It is obtained
from the compiled binary's root and namespace manuals, not from a hand-maintained
test list. The implementation test will derive the same paths from each wrapper's
manual synopsis/usage, so a new documented leaf is automatically covered.

## Deliberate implementation choice

Do not expand `legacyLeafManualTopic`: it is a two-entry per-command switch and is
the defect's cause. Completing a full Cobra leaf migration would be materially
larger and risky because existing handlers deliberately own argument parsing and
dynamic connector passthrough. Instead, put one shared `containsHelpFlag`-based
manual resolution at every legacy wrapper before its handler. Existing Cobra top-level
help remains intact. Missing manuals become a test failure for every registered
wrapper rather than a runtime surprise.

## Inline GSD fallback

This runner is non-interactive and has no compatible Pi subagent runtime. The
generated `scripts/gsd prompt` lifecycle is being executed inline by the single
assigned worker, as required by the canonical contract; no planner/reviewer/verifier
agents are spawned.
