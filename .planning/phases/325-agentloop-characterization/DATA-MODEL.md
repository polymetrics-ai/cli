# Data Model: Synthetic Incident Fixture

Each fixture contains exactly:

- `schema_version`: `1.0`.
- `incident_id`: synthetic identity unique across the fixture directory.
- `summary`: short sanitized description.
- `binding`: complete run, generation, controller, turn, attempt, stage, ticket, evidence, and head
  identities; every value starts with `synthetic:`.
- `events`: one or more events containing a positive contiguous `sequence`, closed-vocabulary
  `kind`, synthetic `actor_id`, and a complete binding exactly equal to the fixture binding.
- `expected`: stable `violation_code`, legacy decision, required decision, and required exit class.

Unknown fields are forbidden at every level. No arbitrary payload map exists. That omission is a
security boundary: incident semantics are represented by typed event kinds rather than copied raw
logs, commands, prompts, or session records.
