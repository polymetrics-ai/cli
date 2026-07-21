# TDD Ledger

Phase: `471-pi-agent-session-shepherd`

## RED: command parser and registration

- Status: expected failures captured before production code.
- `node --test .pi/extensions/shepherd/arguments.test.ts` exited 1 with
  `ERR_MODULE_NOT_FOUND` for `arguments.ts`.
- Offline Pi RPC command discovery exited 1 because `.pi/extensions/shepherd/index.ts` did not
  exist. No model, auth, or network call was made.

## RED: deterministic domain

- Status: expected failure captured before production code.
- `node --test .pi/extensions/shepherd/domain.test.ts` exited 1 with
  `ERR_MODULE_NOT_FOUND` for `domain.ts`.
- Covered contracts: geometric rating, low-score correction, hard-gate precedence, dependency
  readiness, bounded concurrency/one mutator, and crash reconciliation.

## RED: atomic state and controller

- Status: expected failures captured before production code.
- State-store test exited 1 with `ERR_MODULE_NOT_FOUND` for `state-store.ts`.
- Controller test exited 1 with `ERR_MODULE_NOT_FOUND` for `controller.ts`.
- Covered contracts: mode-0600 atomic state, summary redaction/bounds, duplicate ownership,
  concurrent independent lanes, stale evidence halt, fresh resume generation, stop, and shutdown.

## RED: Pi SDK adapter

- Status: expected failure captured before production code.
- `node --test .pi/extensions/shepherd/sdk-runner.test.ts` exited 1 with
  `ERR_MODULE_NOT_FOUND` for `sdk-runner.ts`.
- Covered contracts: resource isolation, exact model/thinking/tools, in-memory state, recursion and
  persistence rejection, run-bound abort, and unconditional cleanup.

## RED: exact target evidence

- Status: expected failure captured before production code.
- `node --test .pi/extensions/shepherd/target-evidence.test.ts` exited 1 with
  `ERR_MODULE_NOT_FOUND` for `target-evidence.ts`.
- Covered contracts: argv-only git/GitHub reads, clean tree, open PR, exact branch, and exact head.

## Gate result

The manual strict-TDD gate passes: all behavior tests and the registration smoke failed for the
expected missing-production-code reason. GREEN implementation may now begin. Refactor evidence will
be added only after the focused suite passes.
