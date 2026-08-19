# PLAN — Issue 3985: connector implementation canon and delivery gate

## GSD setup and inline fallback

- `scripts/gsd doctor` passed in the isolated worktree.
- Resolved `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
  `code-review` through `scripts/gsd sources`; `go run ./cmd/agentcontractgen check` passed.
- Generated `discuss-phase` and `plan-phase --tdd` prompts. The project canonical contract
  forbids spawning GSD roles, so the generated workflows are executed inline and recorded here.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-documentation`,
  `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `vercel-react-best-practices`, and `vercel-composition-patterns`. The routing-required
  `frontend-design` and `web-design-guidelines` skills are not installed in this runtime;
  this docs-only website change uses the available website skills and existing site patterns.

## Goal

Make one versioned, repository-tracked connector implementation procedure the current canon and
make a false `availability: implemented` claim visibly fail local and CI verification.

## Scope

- Import the accepted reports and captain rulings into `data/` with source hashes.
- Add `docs/connector-canon/`:
  - `INDEX.md` records current, superseded, and actively wrong material.
  - `IMPLEMENTATION-PROCEDURE.md` defines the end-to-end procedure and foundation check.
  - `REMOTE-REPRODUCIBILITY.md` names clean-machine requirements and unavailable proof.
  - `archive/` preserves retired local planning/GSD material and external report snapshots.
- Add `make connector-runtime-preflight`, invoking the real commandrunner sweep as a standalone
  target; `make verify` and `make verify-parallel` each cover the same sweep once through `test`.
- Add `make connector-canon-check`, validating the tracked source/archive/procedure contract.
- Update README, `docs/connectors/CANON.md`, generated connector documentation, and website
  documentation to link the canon and state zero accepted certifications plainly.
- Update the stale GSD/planning entry points with archive pointers while preserving their original
  content in `docs/connector-canon/archive/`.
- Captain addendum: audit the live #3972 PostgreSQL parity tree over REST; correct any direct-route
  wording, classify all seven `synccontract.Mode` outcomes, and add a child issue where the
  warehouse-flow/mode conformance foundation is absent. Do not alter active #3974.

## Explicit exclusions

- No connector bundle, provider operation, runtime executor, database write session, CDC behavior,
  GitHub parity implementation, PostgreSQL parity implementation, live credential, or reverse-ETL
  execution. The captain addendum permits issue-tree/documentation corrections only.
- No capability/certification status is promoted.
- No raw generic SQL, HTTP-write, or shell surface is added.

## TDD slices

1. **Red — visible executable-command gate.** Run `make connector-runtime-preflight`; it must
   fail because no named target exists. **Green:** add a target running the real
   `TestEveryImplementedCommandPassesRuntimePreflight`; aggregate verification covers the same
   sweep once through `test`. Rerun the focused target.
2. **Red — canon integrity gate.** Run `make connector-canon-check`; it must fail because no
   target/canon exists. **Green:** add the canon files and deterministic check script, then run it.
3. **Archive preservation.** Copy original superseded material before replacing entry-point files
   with archive pointers. Verify hashes and index references, including explicit ACTIVE-WRONG labels.
4. **Documentation parity.** Update README, connector source/generated docs, and the website;
   run the repository docs check and website test/type checks appropriate to the touched content.
5. **Review and verification.** Run focused targets, format/diff checks, scoped Go tests and
   non-suite `make verify` gates separately per AGENTS.md, then GSD verify/code-review.
6. **PostgreSQL tree correction.** Read #3972 and all children through REST, leave active #3974
   untouched unless it is wrong, add the missing warehouse-flow/mode owner, and record the changed
   dependency graph in a current r2 report plus the deterministic canon check.

## Foundation check required by the deliverable

For every connector capability or flow:

1. Name the prerequisite runtime capability and its owning issue.
2. Exercise it through the real entry point without credentials where possible.
3. Require command preflight, fixture/conformance, and live proof appropriate to the claim.
4. If any prerequisite is missing, mark the operation non-implemented with a named dependency,
   file the foundation issue, and stop the connector completion claim.

No declaration, file name, help route, or generated surface substitutes for an executable foundation.

## CLI/help/docs/website parity

No user-facing command/flag/output change is planned. Documentation explicitly states this; it does
not alter command help. Runtime `pm help` checks are therefore not applicable. README, connector
manual index, and website architecture docs must link the same canon and use the same honest
certification language.

## Commit checkpoints

1. Planning evidence.
2. Red/green gate + canonical imports/archive/docs.
3. Review/verification fixes, if any.
