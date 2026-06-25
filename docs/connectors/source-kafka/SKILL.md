---
name: pm-source-kafka
description: Kafka connector knowledge and safe action guide.
---

# pm-source-kafka

## Purpose

Kafka catalog connector for https://docs.airbyte.com/integrations/sources/kafka. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: database_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-kafka:0.4.2 (metadata only; not executed)

## Runtime Capabilities

- metadata=true
- check=false
- catalog=false
- read=false
- write=false
- query=false
- etl=false
- reverse_etl=false
- unsupported_reason: Native Go port is planned but not enabled; only catalog metadata is available.

## Native Port Plan

- family: database_source
- priority_wave: 3
- etl_operations: catalog, check, read_incremental, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: catalog, check, cursor_incremental, docs_skill, query_safety, read_fixture, secret_redaction, spec, state_checkpoint, type_mapping

## Official Application Documentation

- Apache Kafka documentation: https://kafka.apache.org/documentation/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/kafka

## Configuration

- MessageFormat (object): The serialization used based on this
- auto_commit_interval_ms (integer): The frequency in milliseconds that the consumer offsets are auto-committed to Kafka if enable.auto.commit is set to true.
- auto_offset_reset (string): What to do when there is no initial offset in Kafka or if the current offset does not exist any more on the server - earliest: automatically reset the offset to the earliest off...
- bootstrap_servers (string) required: A list of host/port pairs to use for establishing the initial connection to the Kafka cluster. The client will make use of all servers irrespective of which servers are specifie...
- client_dns_lookup (string): Controls how the client uses DNS lookups. If set to use_all_dns_ips, connect to each returned IP address in sequence until a successful connection is established. After a discon...
- client_id (string): An ID string to pass to the server when making requests. The purpose of this is to be able to track the source of requests beyond just ip/port by allowing a logical application ...
- enable_auto_commit (boolean): If true, the consumer's offset will be periodically committed in the background.
- group_id (string): The Group ID is how you distinguish different consumer groups.
- max_poll_records (integer): The maximum number of records returned in a single call to poll(). Note, that max_poll_records does not impact the underlying fetching behavior. The consumer will cache the reco...
- max_records_process (integer): The Maximum to be processed per execution
- polling_time (integer): Amount of time in milliseconds Kafka connector should try to poll for messages.
- protocol (object) required: The Protocol used to communicate with brokers.
- receive_buffer_bytes (integer): The size of the TCP receive buffer (SO_RCVBUF) to use when reading data. If the value is -1, the OS default will be used.
- repeated_calls (integer): The number of repeated calls to poll() if no messages were received.
- request_timeout_ms (integer): The configuration controls the maximum amount of time the client will wait for the response of a request. If the response is not received before the timeout elapses the client w...
- retry_backoff_ms (integer): The amount of time to wait before attempting to retry a failed request to a given topic partition. This avoids repeatedly sending requests in a tight loop under some failure sce...
- subscription (object) required: You can choose to manually assign a list of partitions, or subscribe to all topics matching specified pattern to get dynamically assigned partitions.
- test_topic (string): The Topic to test in case the Airbyte can consume messages.
- secret fields: MessageFormat.schema_registry_password, protocol.sasl_jaas_config

## Sync Modes

- supported sync modes: full_refresh, incremental
- supports incremental: true

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/kafka

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-kafka
```

### Inspect as JSON

```bash
pm connectors inspect source-kafka --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [Kafka documentation](https://docs.airbyte.com/integrations/sources/kafka)
