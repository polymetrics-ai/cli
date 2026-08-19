# Plan — Issue #3978: final PostgreSQL certification and publication

## Goal

Publish PostgreSQL truthfully: `write=false`, `cdc=true`, and `query=false`, while retaining the separately declared and live-proven closed `postgres_managed_target` destination transport. This is certification/publication repair, not new transport behavior.

## Gap closure plan

1. **Red:** demonstrate that complete managed-target evidence can incorrectly promote generic `Capabilities.Write` while `Connector.Write` remains unsupported.
2. **Green:** keep generic write unpublished and unimplemented; prove six exact `database_write_from_warehouse` sync-mode cells remain declared, implemented, and live-tested under `sync_transport.destination_transport`.
3. **Evidence discipline:** retract the aggregate `capability:write` binding and writer because the capability schema has no closed-destination form. Do not relabel or re-scope mode evidence. Keep the already-produced PostgreSQL 16 proof as support for the twelve exact mode records.
4. **CDC:** retain only the receipt-backed `capability:cdc` and `change_capture/database_read_into_warehouse` proof. CDC-to-API and destination change capture remain non-pass.
5. Regenerate the certification shard, connector docs, and website data. Verify consumer packages, generic-write composition guard, matrix drift, API certification gate, help/docs parity, and relevant repository gates.

## Publication boundary

- `Capabilities.Write=false` means no generic direct writer is advertised.
- `sync_transport.destination_transport` is the narrow live managed-target publication: exact executor, bounded source and destination contract, six named modes, fixed strategy per mode, and receipt-before-acknowledgement.
- `Capabilities.CDC=true` is source-only PostgreSQL 14+ pgoutput to the connection-owned warehouse.
- `Capabilities.Query=false` remains concrete non-support.
