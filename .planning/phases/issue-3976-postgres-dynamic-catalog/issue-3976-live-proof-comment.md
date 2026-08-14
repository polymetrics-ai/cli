## Live PostgreSQL catalog and source-read proof

The pinned PostgreSQL image was not cached. I bootstrapped it explicitly via
the same direct Colima socket used by the harness (the pinned BusyBox capacity
probe was already cached):

```sh
docker --host unix:///Users/karthiksivadas/.colima/default/docker.sock pull docker.io/library/postgres:16.10
```

Then I ran the live source/catalog proof:

```sh
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -run '^TestPostgresDynamicTypedCatalogUsesLiveMetadata$' -v ./internal/connectors/native/postgres
```

Verbatim output:

```text
=== RUN   TestPostgresDynamicTypedCatalogUsesLiveMetadata
    dynamic_catalog_integration_test.go:90: live PostgreSQL full read read_events: ids=1,2,3,4,5 labels=alpha,bravo,charlie,delta,echo
    dynamic_catalog_integration_test.go:90: live PostgreSQL cursor read read_events after=10: ids=3,4,5 labels=charlie,delta,echo
    dynamic_catalog_integration_test.go:90: live PostgreSQL cursor_field absent with stored cursor=12: ids=1,2,3,4,5
    dynamic_catalog_integration_test.go:90: live PostgreSQL nonexistent cursor column: read rejected
    dynamic_catalog_integration_test.go:90: live PostgreSQL nullable cursor rows after=1: ids=23; null cursor row omitted
    dynamic_catalog_integration_test.go:90: live PostgreSQL connection-level cursor_field=sequence: alternate_events rejected because it requires alternate_cursor
    dynamic_catalog_integration_test.go:76: PostgreSQL database test target image-store free bytes: before=100015849472 after=100015857664
--- PASS: TestPostgresDynamicTypedCatalogUsesLiveMetadata (5.41s)
PASS
ok  	polymetrics.ai/internal/connectors/native/postgres	6.146s
```

Observable proof: the real harness discovered the seeded catalog; full read
returned primary keys `1,2,3,4,5`; advancing the cursor to `10` returned only
`3,4,5`.

### Current `cursor_field` behavior observed live

- Without `cursor_field`, stored cursor `12` is ignored and the full set
  `1,2,3,4,5` is returned.
- A named column that does not exist is rejected by PostgreSQL.
- With nullable cursor data, the null row is omitted by the `>` predicate;
  after `1`, only ID `23` was returned.
- `cursor_field` is connection-level today: `sequence` works for
  `read_events` but the same connection config cannot read `alternate_events`,
  whose cursor is `alternate_cursor`.

The captain’s mandatory, user-supplied cursor-column change remains out of
scope for this proof and needs its separate issue; this comment records the
actual existing behavior rather than claiming that contract has changed.

### CDC check — explicitly not counted as proof

```sh
POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.colima/default/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -run '^TestHistoricalLogicalReplicationResumesAndCleansSlot$' -v ./internal/connectors/native/postgres
```

Verbatim output:

```text
=== RUN   TestHistoricalLogicalReplicationResumesAndCleansSlot
    cdc_integration_test.go:23: historical PostgreSQL CDC conformance is disabled while change capture is planned
--- SKIP: TestHistoricalLogicalReplicationResumesAndCleansSlot (0.00s)
PASS
ok  	polymetrics.ai/internal/connectors/native/postgres	0.743s
```

This is an intentional merged-base CDC containment fence, not a live CDC
success. I did not re-enable logical replication in this source-read/catalog
child.
