# TDD Ledger — durable parking resume claim race

## Red

- Pending reproduction of the existing real CLI-process `resume-race` failure.
- Pending separately named happy, bad, and cross-process interleaving coverage
  before the production correction. The test contract classifications are:
  happy = real due-checkpoint resume; bad = live claim typed refusal before
  I/O; edge = two concurrent reopeners over one durable record.

## Green

- Pending production correction and repeated process-test certification.

## Refactor

- Pending review after the causal state/lock transition is established.
