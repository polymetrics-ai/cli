# Inline manual code review

The repository delivery contract requires code review. The direct-PR task uses
the documented inline/manual GSD fallback, so the implementer reviewed the
final diff after all tests and `make verify` passed.

Reviewed paths:

- generic candidate projection and tests in `cmd/connectorgen/`;
- certification schema/validation and direct-read assertion execution;
- GitHub-owned cohort/default definition and regenerated candidate/sweep data;
- Makefile drift gate and phase evidence.

Result: no actionable finding. In particular, shared Go has no connector
identifier, generated candidates cannot become passes without a live stage,
the produced-value assertion is below `/response`, and manual candidates are
explicit exact-command overrides rather than an unbounded bypass.
