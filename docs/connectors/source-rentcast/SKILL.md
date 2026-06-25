---
name: pm-source-rentcast
description: RentCast connector knowledge and safe action guide.
---

# pm-source-rentcast

## Purpose

RentCast catalog connector for https://docs.airbyte.com/integrations/sources/rentcast. Native implementation status: planned_native_port.

## Capabilities

- catalog_metadata=true
- connector type: source
- release stage: alpha
- support level: community

## Implementation Status

- implementation_status: planned_native_port
- runtime_kind: declarative_http_go
- notes: Catalog metadata is available; ETL is disabled until a native Go port passes conformance tests.
- upstream image reference: airbyte/source-rentcast:0.0.53 (metadata only; not executed)

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

- family: declarative_http_source
- priority_wave: 3
- etl_operations: catalog, check, read_snapshot
- reverse_etl_operations: none until native write conformance passes
- conformance: authenticator, catalog, check, docs_skill, pagination, rate_limit_retry, read_fixture, schema_mapping, secret_redaction, spec, state_checkpoint

## Official Application Documentation

- RentCast API documentation: https://developers.rentcast.io/
- Airbyte connector documentation: https://docs.airbyte.com/integrations/sources/rentcast

## Configuration

- address (string): The full address of the property, in the format of Street, City, State, Zip. Used to retrieve data for a specific property, or together with the radius parameter to search for l...
- api_key (string) required secret
- bath_rooms (integer): The number of bathrooms, used to search for listings matching this criteria. Supports fractions to indicate partial bathrooms
- bedrooms (number): The number of bedrooms, used to search for listings matching this criteria. Use 0 to indicate a studio layout
- city (string): The name of the city, used to search for listings in a specific city. This parameter is case-sensitive
- data_type_ (string): The type of aggregate market data to return. Defaults to "All" if not provided : All , Sale , Rental
- days_old (string): The maximum number of days since a property was listed on the market, with a minimum of 1 or The maximum number of days since a property was last sold, with a minimum of 1. Used...
- history_range (string): The time range for historical record entries, in months. Defaults to 12 if not provided
- latitude (string): The latitude of the search area. Use the latitude/longitude and radius parameters to search for listings in a specific area
- longitude (string): The longitude of the search area. Use the latitude/longitude and radius parameters to search for listings in a specific area
- property_type (string): The type of the property, used to search for listings matching this criteria : Single Family , Condo , Townhouse , Manufactured , Multi-Family , Apartment , Land ,
- radius (string): The radius of the search area in miles, with a maximum of 100. Use in combination with the latitude/longitude or address parameters to search for listings in a specific area
- state (string): The 2-character state abbreviation, used to search for listings in a specific state. This parameter is case-sensitive
- status (string): The current listing status, used to search for listings matching this criteria : Active or Inactive
- zipcode (string): The 5-digit zip code, used to search for listings in a specific zip code
- secret fields: api_key

## Sync Modes

- supported sync modes: full_refresh
- supports incremental: false

## Security

- Secret values are never rendered; only secret field names are shown.
- Upstream image references are metadata only and are not executed by pm.
- Catalog-only connectors cannot run ETL until a native Go implementation is enabled.

## Documentation

- https://docs.airbyte.com/integrations/sources/rentcast

## Commands

### Inspect catalog entry

```bash
pm connectors inspect source-rentcast
```

### Inspect as JSON

```bash
pm connectors inspect source-rentcast --json
```

## Agent Rules

- Read implementation_status before planning ETL or reverse ETL.
- If implementation_status is planned_native_port, do not create credentials or runs for this connector yet.
- Never ask for secret values in chat; use pm credentials with --from-env or --value-stdin after native support is enabled.

## References

- [RentCast documentation](https://docs.airbyte.com/integrations/sources/rentcast)
