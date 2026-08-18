# Verification — PostgreSQL CDC Restart Recovery

## Acceptance checklist

- [ ] Current failure reproduced live before production changes, with the mechanism hypothesis confirmed or refuted.
- [ ] Focused restart regression observed red before green.
- [ ] Accepted checkpoint resumes at the exact durable logical-replication position.
- [ ] Source/system/timeline/slot/publication/relation/schema drift still fails closed with typed rebootstrap outcomes.
- [ ] Receipt, checkpoint, and acknowledgement ordering remains durable under injected failure.
- [ ] Independent target query records exact row counts before interruption, at interruption, and after restart.
- [ ] The post-interruption key appears exactly once; no row is lost or duplicated.
- [ ] PostgreSQL CDC capability evidence matches the observed truth.
- [ ] Focused tests, package tests, `internal/cli` where touched, vet, build, generated/connector/doc gates, and diff hygiene pass.
- [ ] CLI help/manual/website parity is confirmed not applicable, or completed if the user-facing contract changes.
- [ ] PR targets `integration/4015-mvp-flat-r1`, carries `Refs #4015`, and identifies release 0.2.1 in its title.

## Executed checks

Pending execution. Exact commands, results, row counts, and trace paths will be recorded here before delivery.
