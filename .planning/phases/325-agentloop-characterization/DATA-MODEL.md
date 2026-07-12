# Data Model: Synthetic Incident Fixture

Each fixture contains exactly:

- `schema_version`: `1.0`.
- `incident_id`: synthetic identity unique across the fixture directory.
- `summary`: short sanitized description.
- `binding`: complete run, generation, controller, turn, attempt, stage, ticket, evidence, and head
  identities; every value starts with `synthetic:`.
- `events`: one or more neutral `transition` records containing a positive contiguous `sequence`,
  synthetic `actor_id`, complete binding, and a fact with closed-vocabulary `kind`, synthetic
  `resource_id`, synthetic `owner_id`, and typed `before`/`after` values.
- `expected`: stable `violation_code`; separately recorded observed decision/outcome and correctness
  booleans; required decision/outcome; and required exit class. Observed may equal required.

Unknown fields are forbidden at every level. No arbitrary payload map exists. That omission is a
security boundary: incident semantics are represented by typed event kinds rather than copied raw
logs, commands, prompts, or session records. Replay rules compare facts across resources and owners;
they do not dispatch on `incident_id`, summary, or a conclusion-shaped event name.
