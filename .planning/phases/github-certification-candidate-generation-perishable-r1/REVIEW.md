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

Rebase review: the final integration target `eba2658c5` changes only
package-scoped `pm` test-binary reuse and its planning evidence. Candidates and
sweep were regenerated twice after that rebase and remained byte-stable; the
final full verification gate passed. No candidate-projection surface was
restructured, preserving the extension point required by the mutation
lifecycle follow-on.
