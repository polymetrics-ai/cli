# Summary — Issue #151 Chatwoot stream runner

Status: merged to parent branch.

Completed slice:

- Added `fan_out.ids_from.request.pagination` so parent id-list requests can use a different pagination model than child fan-out streams.
- Updated Chatwoot `messages` to use the official `after` cursor over message `id` while the parent conversation sweep remains page-number based.
- Updated Chatwoot message cursor metadata, cursor-stop fixtures, generated docs/catalog/website data, and connector authoring conventions.
- Added engine coverage for different parent/child fan-out pagination and Chatwoot conformance coverage proving all seven streams replay non-empty fixtures.

Targeted checks and full handoff gates passed; see `VERIFICATION.md`. PR #246 remote checks passed and was squash-merged into parent commit `8a030090f2c505163a8df9b0bffe01b7fbf35c39`. CodeRabbit skipped the non-default-base sub-PR, then parent PR #223 manual CodeRabbit review replied `Review finished` with no inline findings returned by GitHub API.
