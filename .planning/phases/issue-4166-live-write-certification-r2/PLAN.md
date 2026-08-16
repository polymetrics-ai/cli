# Plan — Issue 4166 Live Write Certification

**Status:** blocked on the scope decision recorded in `SCOPE.md`; no provider
write harness or production edit has started.

## Lifecycle and execution mode

- Generated and executed inline: `scripts/gsd prompt discuss-phase 4166`,
  `scripts/gsd prompt plan-phase 4166 --tdd`,
  `scripts/gsd prompt execute-phase 4166`,
  `scripts/gsd prompt verify-work 4166`, and
  `scripts/gsd prompt code-review 4166`.
- Inline/manual fallback: the canonical contract forbids spawning workflow
  roles; no role was spawned.
- Required skills are recorded in `4166-LIVE-CONTEXT.md`.

## Slice 0 — scope gate (complete)

**Evidence:** `SCOPE.md` classifies every GitHub declared action, establishes
the resource boundary, estimates rate/runtime, and defines resume behavior.

**Gate:** Do not construct a live runner until the open decision selects the
allowed infrastructure scope.

## Subsequent TDD slices (pending decision)

1. **Report truthfulness (red/green).** Red tests prove a prepared-only or
   infrastructure-gated action cannot become `pass`, cannot make the batch
   write cell `pass`, and cannot support a full-parity artifact. Green adds a
   closed non-live outcome and report aggregation.
2. **Full-parity semantics (red/green).** Red tests invoke the exact external
   child path with `--full-parity` alone and show that it is not full/write.
   Green makes it enable full and write stages, rejects contradictory skip
   flags, and refuses an artifact when an applicable full/write stage skipped.
3. **Repository-safe scenarios (red/green).** Red tests require each scenario
   to use the production `runCertify` entry point, mutate its run-owned
   resource, independently read it back, and clean it. Green supplies
   definition-owned scenario metadata and durable resume ledger handling.
4. **Post-schema fault proof (red/green).** A test-scoped defect after schema
   compilation makes the same production certification entry point fail with
   the exact action name; the intact control passes only after observable
   mutation/read-back.
5. **Approved infrastructure waves (red/green).** Only after explicit
   provisioning authority, add one infrastructure family per bounded slice;
   each wave has its own resource guard, cleanup proof, rate budget, and
   resumability test.
6. **Help/docs/website parity.** Update runtime help, CLI manual, website
   reference, generated docs, and parity tests with the live boundary,
   non-live statuses, resumability, and full-parity preconditions.

## Verification pending implementation

- Focused certification and CLI tests with `-timeout 20m`.
- Built `pm` help: `pm help connectors`, `pm connectors`, and
  `pm connectors certify github --help`.
- Credential-gated private-GitHub acceptance run only after the selected
  boundary is authorized; no credential value may reach argv or artifacts.
- `gofmt`, changed-package `go vet`, lint, `go run ./cmd/connectorgen validate`,
  `go run ./cmd/connectorgen surface-sync --check`, connector boundary,
  `pnpm --dir website run gen:docs` twice for byte stability,
  `go run ./cmd/agentcontractgen check`, and the repository GSD workflow check.
