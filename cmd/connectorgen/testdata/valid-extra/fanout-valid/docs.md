# Overview

Fanout Valid is a control connectorgen validate corpus case (S4 engine mini-wave item 2): its
`streams.json` `tasks` stream declares a well-formed `fan_out` block (config_key + path_var +
stamp_field), including a `{{ fanout.id }}` reference in `path` — proving the new `fanout`
pseudo-namespace resolves cleanly through `connectorgen validate`'s static interpolation checks.

## Auth setup

None; public test fixture.

## Streams notes

`tasks` fans out over the `project_ids` config CSV, one request sequence per project id.

## Write actions & risks

None; read-only test fixture.

## Known limits

None; this is test fixture data.
