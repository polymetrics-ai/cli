# Vercel Track A discussion log

## Bound decisions

- Source denominator: the 400 `rest.operations` retained in the frozen Vercel
  schema-v2 source lock. A later importer, declaration, or certification result
  cannot remove one of those rows.
- Boundary: the crosswalk's 22 `surface_only` identities are retained as
  `not_source_row` boundary evidence. They are not folded into the 400-row
  denominator and are not silently discarded.
- Lane evidence is source-first. HTTP verbs and media identify candidates only;
  no cell is an executable claim in Track A.
- The selected source paging cohort is the 22 GET rows with a retained
  `pagination` response wrapper and either (a) `limit` plus one of
  `since`, `until`, `cursor`, or `from`, or (b) both `page` and `per_page`.
  Other retained pagination-shaped source facts remain visible in each row but
  are not silently promoted into this bounded cohort.
- `vercel.rest.createWebhook` retains required `url` and `events` request-body
  fields. The Atlas has generic transport orchestration but no Vercel inbound
  receiver / source executor / conformance intersection, so its sync cell is
  `missing_foundation` only. No shared implementation is authorized.
