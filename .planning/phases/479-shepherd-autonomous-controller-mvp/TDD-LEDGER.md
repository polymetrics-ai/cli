# TDD ledger

## RED

Behavior contract and compiling port/state scaffold committed before controller behavior. Command:

```bash
node --test --test-concurrency=1 .pi/extensions/shepherd/autonomous-controller.test.ts
```

Result: 1 executed, 0 passed, exactly 1 intended failure at
`AutonomousShepherdController.start`: `autonomous Shepherd MVP is not implemented`. The trajectory
covers parallel independent children, dependency waiting, execute/verify/review/integrate ordering,
durable state, stop/join/resume, the final human gate, and absence of a merge-main capability.

## GREEN

Pending.

## REFACTOR

Allowed only after the focused trajectory passes. No speculative hardening loop.
