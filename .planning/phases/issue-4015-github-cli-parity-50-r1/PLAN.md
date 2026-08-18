# Plan — issue #4015 GitHub declared-command parity

## Invocation and delivery mode

- Lifecycle: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`.
- Adapter invocation: `scripts/gsd prompt plan-phase issue-4015-github-cli-parity-50 --tdd`.
- Execution is inline/manual because the canonical single-worker connector lane forbids spawning
  planner, executor, verifier, reviewer, or orchestrator roles.
- Delivery header and acceptance evidence are in `RESEARCH.md`.

## Required skills

- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-graphql`
- `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`

## Dependency-ordered slices

### Slice 1 — inventory contract and fixed aliases

1. Add a table-driven regression test over the supplied 50 names (copied as names only into test
   code) that requires a non-empty surface for promoted commands and exact evidence notes for
   retained commands.
2. Promote `issue pin`, `issue unpin`, and `pr ready` as aliases of their existing fixed GraphQL
   mutations, using the established typed `--input` JSON flag.
3. Promote existing fixed REST reads: cache, secrets, and variables list/get.
4. Run the focused connector/commandrunner tests and record Red/Green evidence.

### Slice 2 — missing safe read declarations

1. Add fixed REST operations for license templates, gitignore templates, authenticated-user orgs,
   authenticated-user gists, codespaces, GPG keys, SSH keys, and agent tasks.
2. Add a fixed paginated GraphQL `RepositoryOwner.repositories` operation for `repo list`.
3. Promote the corresponding commands and derive flags/api surfaces via `surface-sync`.
4. Add operation/schema/pagination tests and run focused gates.

### Slice 3 — typed writes and bounded download

1. Complete provider schemas for autolink create/delete, workflow enable/disable, variable delete,
   and repository codespace creation, then expose source-CLI aliases through the plan lifecycle.
2. Promote `release download` and `run download` to existing bounded single-object downloaders.
3. Extend the bounded binary contract with one declaration-owned `Accept` media type and implement
   `pr diff` without admitting caller-supplied headers.
4. Add preflight and plan-generation tests that assert required flags, fixed endpoints, risk, and
   approval; exercise no-credential binary reachability.
5. Live-certify only operations admitted by the launch brief and current credential scopes. Reads
   assert values/counts; writes assert state, cleanup, and independent absence.

### Slice 4 — truthful retained commands and parity artifacts

1. Replace generic unsupported notes for all 25 retained commands with the exact provider/runtime
   evidence from `RESEARCH.md`; never delete a declared command.
2. Generate/synchronize connector surface, manuals, CLI docs, website references, and certification
   artifacts using repository entry points.
3. Produce `SUMMARY.md` with 50 numbered verdicts, provider observations, cleanup evidence, and the
   final implemented/retained sum.

### Slice 5 — verification, review, and delivery

1. Run `verify-work`; if gaps remain, use `plan-phase --gaps` and `execute-phase --gaps-only`.
2. Run focused tests plus the repository's non-monolithic verification gates, each under the
   documented timeout discipline.
3. Run `code-review`, disposition findings, check for credential material, commit coherent slices,
   push only `fm/cli-parity-implement-50`, and open a PR against `integration/4015-mvp-flat-r1`.
4. Read the PR base back from the GitHub API and require exact equality before reporting delivery.

### Gap closure — fixed GraphQL source roots versus supplemental documents

1. Reproduce PR #4236's `surface_inventory` failure locally and preserve its output as the Red
   evidence in `TDD-LEDGER.md`.
2. Keep `github-operation-source-lock.json` pinned to the provider-derived 31 Query roots and 274
   Mutation roots. `github.repo.list` is a legitimate supplemental fixed document over
   `RepositoryOwner.repositories`; it is not a new provider root and must not change that lock.
3. Make the certification test classify the shared `/graphql` transport bindings into
   source-generated roots and supplemental fixed documents. Require all 305 locked roots, require
   the supplemental set to be exactly `github.repo.list`, and still require the inventory stage to
   count all 306 executable operation bindings.
4. Run the focused Red/Green test, the full `certify` package, the relevant generated-surface gates,
   and the repository workflow-evidence check before pushing the correction to the existing PR.

Gap workflow: `scripts/gsd prompt plan-phase issue-4015-github-cli-parity-50-r1 --gaps` then
`scripts/gsd prompt execute-phase issue-4015-github-cli-parity-50-r1 --gaps-only`, executed through
the inline/manual fallback because this connector lane's canonical contract forbids role spawning.

## Guardrails

- No generic HTTP, caller-supplied GraphQL, generic shell, raw secret output, browser, or subprocess
  capability.
- No hand-authored opaque pagination flags. `surface-sync` owns derived flags and endpoint metadata.
- No writes outside `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`; no ambient gh credential.
- Provider-side live resources use `pm-cert-` identifiers and are deleted/reverted before the
  command family is considered green.
- A provider capability without an existing safe runtime executor is evidence for a retained
  declaration, not permission to smuggle a new shared foundation into the GitHub bundle.
