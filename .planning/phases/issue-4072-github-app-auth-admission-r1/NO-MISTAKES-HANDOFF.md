# Prepared no-mistakes handoff — Issue #4072

**Status:** prepared locally; execution is deliberately held.

## Preserved lineage

- Recovered base: `da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`
- Production GREEN: `3f83bf3afc6efa0ebc323e385e4345f588a41db1`
- Focused verification checkpoint: `72c573bca90f3803ccfe09e914a6bb411c903430`
- Generated-only acceptance checkpoint: `f52745f269fdf642a3315646b5c5ee798e959135`
- Fresh child correction budget: **1/5**; correction 1 is the resolved
  configured-linter RED. The inherited generated-only capability-matrix
  synchronization does not consume correction 2/5.

## Generated-artifact closure

The pre-generation `certification-matrix --check` RED failed identically at
recovered base and clean #4072 source. Canonical regeneration plus `--check`
passed with matrix SHA-256
`e63b906cb640b8fb4fc8fd46c1076b77b7dbced7889919d60527f9b4335d520a` and a
six-entry `discovery_source`-only diff. The stripped semantic SHA-256 remains
`bc5d14758c26755d83a9dc4dcbb715da31d95f67de38e352bc652b752c0819bc`.
#4026/#4034 is cited as generator precedent only; its ancestry is not used.

## Release preconditions

Do not start the command below until all are freshly true:

1. #3856 is no longer in a heavy validation stage; no other shared
   no-mistakes decision-point run is active.
2. The #4072 worktree is clean and contains both preserved commits above.
3. `no-mistakes axi` / `no-mistakes axi status` show no #4072 run to recover
   or reattach and no structured custody action that changes the branch.
4. Broad local acceptance recorded by `GAP-PLAN.md` is green (including the
   canonical certification-matrix GREEN and released bounded matrix at
   `f52745f…`).

If a run returns a gate, drive only that fresh #4072 run. Never send a command
to the parked #3754 run `01KZPZ3VJS7NQJDBNDJVREVPCM`; never edit around an
active pipeline. Ask-user findings become a `needs-decision:` status entry.

## Exact future local-only vector

```bash
no-mistakes axi run \
  --intent 'Deliver existing issue #4072 as the preserved child of #3754: make the GitHub App installation-token POST use engine-owned declared-route rate admission so missing or lost require-shared coordination refuses before transport, a grant produces exactly one Decide/send/Finish lifecycle, process-local and existing GitHub REST/write behavior remain intact, and no credential material enters coordination evidence. Preserve the existing branch, GSD phase, RED/GREEN lineage, and exclude UDS availability, GraphQL policy, provider calls, credentials, CLI, generic transport, PR merge, and parent completion.' \
  --skip=push,pr,ci
```

Run it without `--yes`. After terminal local success, use only the fresh
structured `branch_sync` action offered by no-mistakes. A final child head is
not assigned until supported custody is returned and the worktree is clean.

## External delivery block

Even after local no-mistakes success, do not push or create a child PR until a
captain/owner establishes the exact #3754 parent route: remote
`feat/3754-shared-rate-coordinator` must contain `da8a8ff…`, and exactly one
open draft parent PR must target
`docs/4015-connector-release-certification`. The prepared child PR base is
only `feat/3754-shared-rate-coordinator`; `main`, #4019, and the certification
base are not child bases.
