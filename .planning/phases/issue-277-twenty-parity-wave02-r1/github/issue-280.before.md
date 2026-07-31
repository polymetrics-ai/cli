Part of #277 (S3). Write scope: `streams.json`. Deps: #278, #279.

- 28 list streams (`stream_read`, cursor pagination on `pageInfo`/`endCursor`, `starting_after`+`limit`, `order_by`, `filter`, `depth`) + 28 `direct_read` get-by-id ops. Each references its S2 schema.

Acceptance: pagination/cursor fields set; focused stream tests.
Refs #277