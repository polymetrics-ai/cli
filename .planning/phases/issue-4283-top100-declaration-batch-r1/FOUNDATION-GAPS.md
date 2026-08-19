# Foundation gaps — increment 001

This declaration-only increment does not claim a generic API sync transport. The complete, operation-level foundation reasons are machine-readable in `FOUNDATION-GAP-REASONS.json`; the entries retain their exact source operation and recovery condition. `TRANSPORT-GAP.md` records the explicit, recoverable transport decision for all ten connectors.

## Shared generic declarative API transport registration

- State: open, tracked by [#4093](https://github.com/polymetrics-ai/cli/issues/4093).
- Affected connectors: Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana.
- Effect: no `sync_transport.json` is emitted for this cohort. Source ETL and reverse-ETL destination transport are marked `foundation-gap`, `recoverable: true` rather than claiming an executor or a conformance run which does not apply to these bundles.
- Evidence: `internal/synctransport/definition_composition.go:145-168` requires an exact registered factory and accepted evidence. `internal/app/issue_label_warehouse_transport.go:54-103` admits the source only under the GitHub-specific evidence constant, while lines 322-368 restrict the destination to the issue-label action contract. `internal/connectors/certify/stages_transport_internal_test.go:89` proves failure before any provider call if registration is missing.
- Recovery: deliver #4093 with a connector-neutral source factory and per-bundle evidence, plus a closed typed destination adapter with explicit bindings, acknowledgement, and per-mode strategies. Then derive each connector-local descriptor from its pinned source and rerun the non-live sweep.

## Schema-shaped operations that remain disabled

The rejection list classifies 1,602 operations as `schema-incompatible`. This increment intentionally did not create request, response, pagination, or body schemas where the pinned provider document lacked a bounded usable contract. Those entries become enabled only after the provider publishes a usable schema or the engine implements the documented shape.
