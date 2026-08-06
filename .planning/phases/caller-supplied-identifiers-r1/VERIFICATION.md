# VERIFICATION — caller-supplied identifier sets

Status: planned; populated after implementation and review.

## Acceptance checklist

- [ ] Closed operation declaration defines a name, shape, wire, minimum, and mandatory maximum.
- [ ] Test-only bundle reaches its operation command and none of the commands is an ETL stream.
- [ ] All four declared encodings have end-to-end wire assertions.
- [ ] Bounds and malformed input fail before a server receives a request and never echo a supplied identifier.
- [ ] Explicit blank and absent list flags have distinct tested outcomes.
- [ ] Provenance/coverage remains evidence-only; output-policy schema/runtime drift guard still passes.
- [ ] Migration conventions document authorship and the nested-batch decision.
