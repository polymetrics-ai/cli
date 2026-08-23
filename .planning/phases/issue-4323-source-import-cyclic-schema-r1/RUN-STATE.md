# Run state — issue 4323

State: #4326 red-test-recorded

- Issue opened: #4323.
- Base ancestry confirmed: `cf493b834` is an ancestor of the branch start.
- GSD adapter and canonical contract checks passed.
- Red/green source-import contract recorded for direct, mutual, deeply nested,
  finite, and unused-schema behavior.
- Real Grafana source import emitted 52 explicit recursive-schema gaps while
  retaining 314 operations; its temporary v2 lock and output have been removed.
- Focused recursive-schema behavioral tests, `go vet ./...`, `go build ./cmd/pm`,
  `git diff --check`, frozen GitHub artifact measurements, and full `make verify`
  are green. The shared 20-minute test timeout remains unchanged.
- Inline GSD code review found no unresolved actionable findings.
- The widened #4326 test is red: OpenAPI 3.0 response references with
  `description` and `summary` both fail grammar preflight as ambiguous siblings.
- Next gate: permit those non-semantic fields only, retain semantic-sibling
  rejection, run real provider imports, and re-run the full verification gate.
