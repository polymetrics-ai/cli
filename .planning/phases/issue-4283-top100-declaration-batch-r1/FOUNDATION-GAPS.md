# Foundation gaps — increment 001

This declaration-only increment does not claim a generic API sync transport. The complete, operation-level foundation reasons are machine-readable in `FOUNDATION-GAP-REASONS.json` (140 distinct reasons across 1,359 disabled public operations); the entries retain their exact source operation and recovery condition.

## Shared generic declarative API transport registration

- State: open, tracked by [#4093](https://github.com/polymetrics-ai/cli/issues/4093).
- Affected connectors: Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana.
- Effect: no `sync_transport.json` is emitted for this cohort. Source ETL and reverse-ETL destination transport remain disabled rather than claiming a source or destination executor which the runtime cannot register.
- Evidence: `internal/connectors/sync_transport.go:34` requires every descriptor to name a concrete, registered executor. `internal/connectors/engine/bundle.go:1751` loads and validates a descriptor but does not register one. `internal/connectors/certify/stages_transport_internal_test.go:89` proves that an unregistered `declarative_stream_source` fails before any provider call.
- Recovery: deliver #4093 with registered, bounded declarative API source and destination adapters plus durable acknowledgement / apply-strategy evidence. Then derive each connector-local descriptor from its pinned source and rerun the non-live sweep.

## Schema-shaped operations that remain disabled

The rejection list classifies 1,602 operations as `schema-incompatible`. This increment intentionally did not create request, response, pagination, or body schemas where the pinned provider document lacked a bounded usable contract. Those entries become enabled only after the provider publishes a usable schema or the engine implements the documented shape.
