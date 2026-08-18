# Verification — PostgreSQL CDC Restart Recovery

## Acceptance checklist

- [x] Current failure reproduced live before production changes, with the mechanism hypothesis confirmed or refuted.
- [x] Focused restart regression observed red before green.
- [x] Accepted checkpoint resumes at the exact durable logical-replication position.
- [x] Source/system/timeline/slot/publication/relation/schema drift still fails closed with typed rebootstrap outcomes.
- [x] Receipt, checkpoint, and acknowledgement ordering remains durable under injected failure.
- [x] Independent target query records exact row counts before interruption, at interruption, and after restart.
- [x] The post-interruption key appears exactly once; no row is lost or duplicated.
- [x] PostgreSQL CDC capability evidence matches the observed truth.
- [ ] Focused tests, package tests, `internal/cli` where touched, vet, build, generated/connector/doc gates, and diff hygiene pass.
- [ ] CLI help/manual/website parity is confirmed not applicable, or completed if the user-facing contract changes.
- [ ] PR targets `integration/4015-mvp-flat-r1`, carries `Refs #4015`, and identifies release 0.2.1 in its title.

## Executed checks

1. Focused red/green and package results are recorded in `traces/focused-red.txt` and `traces/focused-green.txt`.
2. Live failure and repaired process-restart runs are recorded in `traces/live-red.txt` and `traces/live-green.txt`.
3. Exact CDC target counts: before interruption `1`, at interruption `1`, after restart `2`; resumed key multiplicity `1`. The independent control target stayed at `1,001` rows.
4. The immutable capability record remains `passed`: its explicit bounded-stage/receipt/readback/acknowledgement facts passed again with the fixed binary. It does not claim generic exactly-once delivery; PostgreSQL CDC remains declared at-least-once.
5. Broader repository commands remain pending below.
