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

## GREEN: deterministic core and SDK adapter

- Core worker commit `38dcb435745f333787fff3e3b4ea3dd0d585db1c` passed 14/14 focused
  tests and was integrated as `ee2285b1`.
- SDK worker commit `3583b584c14365b14257cbb27f81ea16d4a08340` passed 3/3 focused
  tests and strict TypeScript, and was integrated as `107e74fb`.
- Subsequent no-tools hardening commit `2ccd6f4c` expanded the SDK suite to 7/7: embedded sessions
  receive no built-in or custom tools, fail closed on malformed/oversized evidence and event
  limits, time out through owned abort, and always clean up.
- State persistence hardening commit `19e6dcaf` passed 6/6 focused tests: disk DTOs are explicitly
  allowlisted, unknown disk fields fail closed, summaries are redacted/bounded/single-line, and
  runtime-only fields are stripped before serialization.

## RED to GREEN: independent review corrections

- Reviewer RED: controller enum transitions could not round-trip through the validating store.
  GREEN: canonical `succeeded|failed|halted|stopped` lane states; real-store regression passes.
- Reviewer RED: persisted `running` state could not resume after a host crash. GREEN: resume first
  reconciles it to `interrupted`, then creates a fresh generation/head/nonce; regression passes.
- Reviewer RED: an in-flight lane could overwrite `stop`. GREEN: controller-owned cancellation is
  checked before and after final persistence; deferred-runner regression leaves disk `stopped`.
- Reviewer RED: child read tools were not repository-confined. GREEN: child sessions receive
  `tools: []`, `customTools: []`, and `noTools: "all"`; requested tools fail before creation.
- Reviewer RED: model-authored success could finish without recapturing the target. GREEN: host
  code recaptures exact local, PR, and check evidence after both lanes and hard-halts on change.
- Extension RED: two issues could launch concurrently and shutdown produced late UI output.
  GREEN: one process-wide active run, shutdown cleanup/waiting, and notification suppression.

## Current focused gate

- `node --test .pi/extensions/shepherd/*.test.ts`: 38/38 pass.
- Strict TypeScript no-emit over all production files including `index.ts`: pass.
- Offline Pi 0.80.6 RPC command registration: pass (`pm-shepherd`, source `extension`).
