# Inline code review — Refs #4093

The final diff was reviewed inline because the non-Pi runner cannot host the
adapter's isolated review worker.

- Definition evidence is validated before adapter construction and retained as
  exact role/reference/evidence triples. Unknown evidence is refused.
- A shared factory is registered once per exact reference only after every
  descriptor has validated. The reusable source reads its request connector,
  and the typed destination reads its preflighted destination; neither
  production adapter retains the first connector encountered during registry
  iteration.
- The destination remains a closed issue-label action adapter. It uses the
  existing plan/preview/approval and independent read-back path; no generic
  HTTP, SQL, shell, or arbitrary action surface was added.
- New stage fields are connector interfaces passed only in-memory through the
  orchestrator. They contain no credentials, are not serialized, and do not
  affect receipt ownership or checkpoint persistence.
- Existing fail-closed loading, mode/source-binding preflight, post-commit
  reconciliation, and bounded owned-stage cleanup are preserved by regression
  tests.
- `change_capture` is rejected as a destination declaration and remains on its
  distinct source-only connection-warehouse execution path.
- The Gate A proof test permits a retained baseline only when it is exactly the
  named label owned by that private proof issue; an empty baseline remains
  valid and any unexpected label fails before the test can mutate GitHub. The
  proof logs each completed one-record run only after its durable checkpoint
  and independent post-exit read-back assertions pass.

No unresolved actionable finding remains. The broader ad-hoc staticcheck
findings are pre-existing and recorded in `VERIFICATION.md`; `make lint` is
green.
