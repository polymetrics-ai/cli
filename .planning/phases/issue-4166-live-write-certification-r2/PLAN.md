# Plan — Issue 4166 Live Write Certification

**Status:** repository-safe wave implemented and live-proven for 25 actions;
three commit-comment actions remain deliberately blocked pending an exact
fine-grained permission repair. The confirmed
disposable certification identity, Polymetrics-Cert organisation, and
in-progress Enterprise Cloud trial make the other 579 actions conditionally
live-testable once their named browser-only credentials/fixtures exist. Until
then they are concrete `not_live` rows and make a full-parity claim unavailable.

## Lifecycle and execution mode

- Generated and executed inline: `scripts/gsd prompt discuss-phase 4166`,
  `scripts/gsd prompt plan-phase 4166 --tdd`,
  `scripts/gsd prompt execute-phase 4166`,
  `scripts/gsd prompt verify-work 4166`, and
  `scripts/gsd prompt code-review 4166`.
- Inline/manual fallback: the canonical contract forbids spawning workflow
  roles; no role was spawned. The required code review used the documented
  manual fallback and is recorded in `REVIEW.md`.
- Required skills are recorded in `4166-LIVE-CONTEXT.md`.

## Slice 0 — scope gate (complete)

**Evidence:** `SCOPE.md` classifies every GitHub declared action, establishes
the resource boundary, estimates rate/runtime, and defines resume behavior.

**Gate resolved:** captain selected option 1. The browser-only prerequisites
are consolidated in `MANUAL-PROVISIONING.md`; implementation proceeds without
waiting for that larger-infrastructure work.

## TDD slices

1. **Report truthfulness (red/green).** Red tests prove a prepared-only or
   infrastructure-gated action cannot become `pass`, cannot make the batch
   write cell `pass`, and cannot support a full-parity artifact. Green adds a
   closed non-live outcome and report aggregation.
2. **Full-parity semantics (red/green).** Red tests invoke the exact external
   child path with `--full-parity` alone and show that it is not full/write.
   Green makes it enable full and write stages, rejects contradictory skip
   flags, and refuses an artifact when an applicable full/write stage skipped.
3. **Repository-safe scenarios (red/green; complete for 25/28).** Red tests require each scenario
   to use the production `runCertify` entry point, mutate its run-owned
   resource, independently read it back, and clean it. Green supplies
   definition-owned scenario metadata and durable resume ledger handling. The
   combined production run passed 25 named actions, including independent
   read-back and verified cleanup/restoration. The three commit-comment
   actions are `blocked` with the exact item-read refusal, and their historic
   tagged comments have independently verified collection-read cleanup.
4. **Post-schema fault proof (red/green; complete).** A test-scoped defect after schema
   compilation makes the same production certification entry point fail with
   the exact action name; the intact control passes only after observable
   mutation/read-back.
5. **Approved infrastructure waves (red/green).** After the captain completes
   `MANUAL-PROVISIONING.md`, add one of the 579 named prerequisite families per
   bounded slice; each wave has its own resource guard, cleanup proof, rate
   budget, and resumability test.
6. **Help/docs/website parity (complete).** Update runtime help, CLI manual, website
   reference, generated docs, and parity tests with the live boundary,
   non-live statuses, resumability, and full-parity preconditions.

## Verification evidence

- Focused certification and CLI tests with `-timeout 20m`.
- Built `pm` help: `pm help connectors`, `pm connectors`, and
  `pm connectors certify github --help`.
- Credential-gated private-GitHub acceptance run: the 25-action result is
  `.polymetrics/certifications/github.json` under the retained temporary root
  `pm-cert-4166-all-wave.YY4wos`; it has 25 `pass`, three `blocked`, no
  `leaks`, and exits 2 precisely because incomplete coverage is not success.
  The commit-comment-only blocked report is retained under
  `pm-cert-4166-commit-blocked.nFzeva`; its cleanup is `verified` but every
  action remains `blocked`.
- A scratch post-schema path defect for `update_issue` made the production
  `--write-only` run exit 2 and name `write_wave_update_issue`; the definition
  was immediately restored with no source change retained.
- `gofmt`, changed-package `go vet`, lint, `go run ./cmd/connectorgen validate`,
  `go run ./cmd/connectorgen surface-sync --check`, connector boundary,
  `pnpm --dir website run gen:docs` twice for byte stability,
  `go run ./cmd/agentcontractgen check`, and the repository GSD workflow check.
