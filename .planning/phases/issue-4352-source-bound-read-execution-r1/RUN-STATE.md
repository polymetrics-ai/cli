# Run state — #4352

Current phase: final local verification and review complete; PR #4353 update,
external automated review, and human merge gate remain. Do not merge.

Base: `main` at `b33983927d863032dac8220949990506e812937d` after the required rebase.

Firstmate inbox dispositions:

- `001.msg`: do not merge. Before a merged certification/mapping foundation can integrate this work, require complete exact-head tests, an independent audit, and real connector-shaped evidence including Outreach where applicable.
- `002.msg`: after a foundation merges, connectors with exact mapping may merge with named deferred rows; release notes and PR handoff must distinguish enabled commands from honest `missing_foundation`/declaration-pending rows.
- `004.msg`: resume audit repair only on `fm/cli-source-bound-read-execution-r1`; repair F1 with the tracked Asana skill and root-help golden artifacts and repair F2 with a retained pinned Asana source artifact/lock plus hermetic generator-to-bundle divergence assertions. Do not merge, fabricate provider proof, or copy #4350/#4351 changes.
- `005.msg`: keep the repair bounded to F1/F2, remove temporary `PM_DEBUG` instrumentation, and record rather than absorb any broader schema-v3 foundation dependency. The retained Asana source-import check is clean without such a dependency.
- `006.msg`: disk headroom is near the floor; all remaining generation and validation is serialized and focused. No caches or other workers' work are deleted.
- `007.msg`: source-bound read projection had rewritten reverse-ETL actions. The repair is read-only: source import now has a bounded `--read-projection-only` lane and a regression that holds `writes.json` and reverse-ETL/delete command artifacts byte-identical.
- `008.msg`: Asana generated-artifact preservation is proven by the same retained-source regression. Only the three derived read files were restored to their committed baseline and regenerated through the corrected lane; source locks, source evidence, and write/delete artifacts were retained.

Audit-repair progress: F1 stale artifact failures were reproduced, then the tracked Asana skills/manual and nine root-help golden records were regenerated from the current bundle and revalidated. F2 retains the pinned Asana OpenAPI capture and v3 lock; its hermetic test imports those bytes and rejects source identity, method/path, required typed-input, and workspace-pagination drift. The final lane is explicitly read-only: it binds the nine already executable Asana reads (three direct reads and six streams) and leaves the remaining planned GETs, all write actions, and all delete controls intact. Temporary debug instrumentation has been removed. Remaining work is focused runtime/binary evidence, review, and exact-head PR validation.

Captain parity reconciliation (`011.msg`, resolved under the bounded `012.msg`
authorization): the source validator reported 90 Asana mutation gaps. Their
operation-granular dispositions are now retained beside the pinned source
artifact in `sources/asana-mutation-dispositions.json`, with the exact inventory
in `MUTATION-GAP-INVENTORY.md`:

- 21 source mutations with no command/action remain source-cited
  non-executable declarations.
- 65 existing implemented reverse-ETL commands retain their executable surface
  and record `cli-request-schema-foundation-r1` partial coverage because the
  authoritative provider request body is not yet represented by the typed CLI
  contract.
- 4 existing implemented delete commands retain their executable surface and
  record `source-path-parameter-alias-foundation-r1` partial coverage because
  their local `gid` flag is not yet an exact provider path-parameter mapping.

The new partial disposition is source identity/method/path bound, accepts only a
named known foundation, requires an existing but incomplete implemented command,
and rejects a fully source-covered action. It therefore cannot downgrade a
working operation, invent an executor, conceal an absent action, or label a
complete contract deferred. Source import, full source validation, and
surface-sync are clean. A concurrent source-lock audit owns the sole broad
`cmd/connectorgen` test; its completion is required before this branch runs its
own package-wide generator test. Do not push or recommend merge until that
serialized verification and the remaining runtime/binary checks are green.

Current repair status (`014.msg`/`015.msg`): the source-lock audit's broad
generator lane exited PASS, then this exact head passed
`go test -timeout 20m -count=1 ./cmd/connectorgen` in 150.671s;
`source-import asana --read-projection-only --check` (249 operations);
full 553-connector `validate` (0 findings); and full 553-connector
`surface-sync --check` (zero drift). Runtime, Asana preservation, generated
help/manual, docs validation, `go vet ./cmd/connectorgen`, the canonical
agent-contract check, and the built-binary credential-boundary preflight are
also green. The retained source-cited mutation inventory is 21 absent actions
under non-executable declarations, 65 implemented reverse-ETL request-schema
gaps, and 4 implemented delete path-parameter aliases. No provider I/O,
credential, temporary workspace, unrelated artifact, merge, or force push is
in scope. The next authorized action is to inspect, stage only this PR's
reviewed files, commit, and non-force push for a fresh audit/CI.

Delivery status (`017.msg`): the reviewed repair is locally committed on
`fm/cli-source-bound-read-execution-r1`, pending the approved non-force push to
PR #4353. The byte-pinned Asana artifact remains unchanged at
SHA-256 `cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56`
and 3,066,750 bytes; both the source lock and retained-artifacts manifest
match it, and `.gitattributes` exempts only that file from whitespace checks.

Captain correction status (`021.msg` and `023.msg`): the historical 9/100
partition is superseded. Source import now evaluates every non-mutating locked
GET by capability and is deterministic against historical status drift: a
complete exact REST contract becomes bounded `direct_read`; an exact stream
with records/schema/pagination becomes ETL; and only a concrete named gap may
remain deferred. The retained Asana result is 106 direct reads, 12 ETL streams,
and one deferred GET (`asana.rest.getMembership`,
`cli-openapi30-reference-sibling-foundation-r1`). The current worktree has not
changed the raw artifact, source lock, `.gitattributes`, other connectors, or
another worker's branch. Final validation is serialized with 12 GiB free disk.

## r4 direct-PR completion state

The immutable #4356 review set is frozen and locally green on
`fm/cli-source-bound-read-execution-r1-continuation`. The final repair adds
closed source paging, public-only origin preflight before `withApp`, direct
`Read` source-route proof, 21 existing-lane Asana action promotions, fixed-100
cohort isolation, and regenerated source-backed documentation/evidence. The
final built-binary Asana census is `212/212` implemented-or-partial commands at
the `missing --credential` boundary, with zero provider calls or credentials.

Do next: commit the reviewed repair and planning evidence, non-force push only
to PR #4356's existing branch, verify its remote SHA/base, then request a fresh
independent Codex audit. Do not merge. PR #4351 at
`fd400c501d99daa22210d42f736742706b4d8f1a` is related generic admission work,
not a blocking dependency for this source-bound preflight.

## Current-main integration delivery header — 2026-08-27

- **Issue:** Refs #4352 — source-bound read execution foundation (existing PR
  #4356 repair continuation).
- **Base branch:** `main`.
- **Merges into:**
  `fm/cli-source-bound-read-execution-r1-continuation` → `main`.
- **Delivery:** Existing PR #4356 remains open against `main`, includes
  `origin/main` at `1324c52bab0b224ed8958858af7676b8b8e191b4`, and has passing
  source-bound read/delete/ETL validation plus an Asana built-binary
  credential-boundary census. No PR merge is authorized.
- **Working branch:** `fm/cli-source-bound-read-execution-r1-continuation`.
- **Task:** Integrate the Captain-authorized #4351 current-main foundation by
  a normal merge only; resolve only real integration conflicts; preserve all
  other work; then prove the shared foundation and Asana source-bound surface
  still reach the credential boundary without a credential or provider I/O.
- **Verification:** `git merge --no-ff origin/main`, affected Go packages,
  source import/validation/surface-sync/operation-evidence/runtime-preflight/
  canon gates, docs validation, and a newly built isolated `pm` census of all
  implemented Asana commands.
- **GSD/manual fallback:** resolved `discuss-phase`, `plan-phase --tdd`,
  `execute-phase`, `verify-work`, and `code-review` through `scripts/gsd`.
  The compatible Pi role runtime is unavailable and this task is a bounded
  integration, so the approved single-worker inline fallback records its
  red/green and review evidence here instead of spawning roles.
- **Required skills:** `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-lint`.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| The approved #4351 foundation is in this existing PR branch | live | `git merge-base HEAD origin/main` equals `1324c52...`; before integration it was `b339839...`, so the new foundation was absent. |
| Existing source-bound behavior retains its credential boundary | live | A freshly built `pm` in an isolated project runs every implemented Asana command and each returns `missing --credential`, rather than a provider result or unknown-command failure. |
| Delete, ETL, and generated source contracts remain valid | live | Affected Go suites and source-import/validate/surface-sync/operation-evidence/runtime-preflight/canon checks pass after the merge. |
