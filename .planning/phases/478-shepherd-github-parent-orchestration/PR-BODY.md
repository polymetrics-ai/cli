## Summary

- add a bounded, typed GitHub parent/child orchestration transport with reconcile-before-mutate
  issue, stacked-PR, roster, integration, and parent-ready operations
- validate authoritative CI, review-thread, requested-change, finding-disposition, and exact-range
  independent Codex review evidence
- reuse dependency scheduling, autonomy reconciliation, workspace handoffs, and the existing human
  decision broker while keeping review execution and parent merge outside this slice

## GSD / TDD

- mode: `manual_gsd_fallback` because the healthy repo adapter does not expose
  `programming-loop`
- required skills: `gsd-programming-loop`, `github-issue-first-delivery`, `gsd-workstreams`,
  `architecture-patterns`, `javascript-testing-patterns`
- initial test-only RED: 0 pass / 3 absent-module failures
- minimal GREEN: 21/21 focused tests
- adversarial test-only RED: 17 pass / 10 expected failures against unchanged production
- corrected GREEN: 27/27 focused tests at `40ce66d4b5010b92089895a05709687143d15a05`

## Verification

- focused #478 tests: 27/27 pass
- serialized Shepherd tests: 290 pass, 0 fail, 1 intentional sandbox skip
- strict TypeScript 5.9.3 over owned tests/modules and all 20 production modules against cached Pi
  0.80.6: pass
- pinned Pi 0.80.6 offline RPC: `pm-shepherd` discovered from `extension`
- immutable-base ancestry, full-range diff check, and owned-path scope: pass
- Go, connector, certification, runtime-service, and `make` gates: not run by parent policy

## Review and safety

- fake orchestration transports only; no live issue/comment/ready/merge transport ran
- no reviewer was started; the parent owns the stable-head `codex_independent`
  `openai-codex/gpt-5.6-sol:xhigh` campaign
- no Claude, Copilot, parent merge, or default-branch mutation was requested

Refs #478

Refs #471
