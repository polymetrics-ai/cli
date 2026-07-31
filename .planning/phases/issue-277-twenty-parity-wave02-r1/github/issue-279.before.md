Part of #277 (S2). Write scope: `schemas/**`. Deps: #278.

- 28 per-object JSON schemas, 546 fields total; each with `x-primary-key: id` and `x-cursor-field: updatedAt`; common `id`/`createdAt`/`updatedAt` typed.

Acceptance: `jq .` clean; conformance `path_fields ⊆ record_schema` prerequisite satisfied.
Refs #277