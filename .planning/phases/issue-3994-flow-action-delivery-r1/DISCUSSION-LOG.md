# Discussion log — Issue #3994

Inline `discuss-phase` fallback: the issue fixes the decisions that would normally require interactive intake. The durable authorization scope from #4132 is authoritative; payload content is excluded from standing scope; a changed scope must fail before a provider request; approval tokens are single-use and are not stored in schedules. This phase must use a typed connector write path and cannot reintroduce a generic URL, HTTP, SQL, or raw-operation write surface.

The unresolved external dependency is live-provider access. No credential will be requested or recorded. If the captain runbook cannot be used safely after implementation, the PR/issue will name that deferred proof and the substitute hermetic observable tests.
