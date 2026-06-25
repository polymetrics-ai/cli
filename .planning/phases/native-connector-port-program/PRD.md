# Native Connector Port Program PRD

## Goal

Create an implementation program for native Go ports for all catalog connectors, including database CDC requirements and reverse ETL write boundaries, without marking unimplemented connectors as runnable.

## Requirements

- Every catalog connector has a native port plan.
- Plans classify each connector into a runtime family: native SaaS, declarative HTTP source, database source with CDC, file/object source, destination writer, or custom Go port.
- Database source plans describe CDC modes and prerequisites for Postgres logical replication, MySQL binlog/GTID, MongoDB change streams, SQL Server CDC, Oracle LogMiner/XStream, and warehouse-native incremental extraction where applicable.
- Plans list required conformance checks before a connector can move from `planned_native_port` to `enabled`.
- CLI exposes all plans in JSON and manual form for agents and humans.
- Generated docs include the implementation plan and CDC requirements.

## Non-Goals

- Do not implement 646 production data-plane ports in one patch.
- Do not add database drivers or CDC client dependencies in this phase.
- Do not run upstream connector images.
- Do not enable reverse ETL writes for connectors without native write conformance tests.
