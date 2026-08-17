# Verification — GitHub certification candidate generation

Status: complete — implementation, bounded live proof, and final-base local
repository verification passed on integration `17c43c75a`.

- [x] Candidate projection is deterministic and preserves named manual cases.
- [x] Candidate assertions select only produced response values.
- [x] Perishable classification totals 192 and sweep totals 1,571.
- [x] Provider refusal and product defect are separately recorded.
- [x] Generated artifacts are byte-stable: candidate and sweep generation ran
  twice, then both `--check` modes passed.
- [x] The real binary executed all 97 generated trial reads serially; 95 trial reverse-ETL commands are deferred to their lifecycle delivery.
- [x] `copilot configuration view` demonstrably failed when its own generated assertion was broken, then passed after restoration.
- [x] Targeted and consumer tests pass:
  `go test -timeout 20m ./cmd/connectorgen` (117.973s),
  `go test -timeout 20m ./internal/connectors/certify`, and
  `go test -timeout 20m ./internal/connectors/engine`.
- [x] After rebase to `17c43c75a`, candidate and sweep generators again ran
  twice byte-stably and their `--check` modes passed; the 97-member generated
  set remained unchanged from the rerun set.
- [x] Final `make verify` passed, including `internal/cli` (580.634s),
  candidate generation/sweep drift checks, connector validation/boundary, docs,
  smoke, lint, and release-target parity.
- [x] Manual inline code review found no actionable issue: generic Go contains
  no provider identifier, generated candidates assert `/response` rather than
  exit status, manual overrides remain explicit, and the 1,571 sweep total is
  asserted by the generator and its tests.

## Named boundary

The mutation-candidate/fixture lifecycle is not part of this direct-read task.
It must supply the ceremony data that the declared CLI surface cannot express:
fixture creation, plan, preview, approval, execution, independent read-back,
and cleanup. No claim here extends from the 97 executed reads to the remaining
95 reverse-ETL trial commands or to the full 1,571-command gate.
