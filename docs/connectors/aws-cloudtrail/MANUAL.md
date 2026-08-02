# pm connectors inspect aws-cloudtrail

```text
NAME
  aws-cloudtrail - AWS CloudTrail

SYNOPSIS
  pm credentials add <name> --connector aws-cloudtrail --config aws_region_name=<region> --from-env aws_key_id=AWS_ACCESS_KEY_ID --from-env aws_secret_key=AWS_SECRET_ACCESS_KEY
  pm etl catalog --connector aws-cloudtrail --credential <name> --json
  pm etl read --connector aws-cloudtrail --credential <name> --stream <stream> --json

DESCRIPTION
  Reads AWS CloudTrail configuration and resource metadata through fixed AWS JSON-RPC read streams.

  Current status: 60 official CloudTrail API actions remain inventoried, 19 read-stream actions are exposed as executable connector-local runtime behavior, and the 10 provider query/direct-read actions plus 31 write/admin actions are blocked/planned until typed operation/write metadata and shared promoted-native command surfaces, validation, dry-run previews, and operation-direct reads are available safely.

ICON
  asset: icons/aws-cloudtrail.svg
  source: upstream_registry
  review_status: upstream_seeded

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Store AWS credentials with pm credentials. Use --from-env or --value-stdin only; never put AWS secret values in chat, shell history, docs, or issue comments.

CONFIGURATION
  aws_region_name (required config)
  base_url (optional fixture/local endpoint)
  page_size (optional)
  max_pages (optional)
  mode=fixture (optional credential-free local tests)
  aws_key_id (required secret)
  aws_secret_key (required secret)

ETL STREAMS
  describe_trails
  get_channel
  get_dashboard
  get_event_configuration
  get_event_data_store
  get_event_selectors
  get_import
  get_insight_selectors
  get_resource_policy
  get_trail
  get_trail_status
  list_channels
  list_dashboards
  list_event_data_stores
  list_import_failures
  list_imports
  list_public_keys
  list_tags
  list_trails

BLOCKED / PLANNED OPERATIONS
  Provider query/direct-read actions blocked: CancelQuery, DescribeQuery, GenerateQuery, GetQueryResults, ListInsightsData, ListInsightsMetricData, ListQueries, LookupEvents, SearchSampleQueries, StartQuery.

  Write/admin actions blocked: AddTags, CreateChannel, CreateDashboard, CreateEventDataStore, CreateTrail, DeleteChannel, DeleteDashboard, DeleteEventDataStore, DeleteResourcePolicy, DeleteTrail, DeregisterOrganizationDelegatedAdmin, DisableFederation, EnableFederation, PutEventConfiguration, PutEventSelectors, PutInsightSelectors, PutResourcePolicy, RegisterOrganizationDelegatedAdmin, RemoveTags, RestoreEventDataStore, StartDashboardRefresh, StartEventDataStoreIngestion, StartImport, StartLogging, StopEventDataStoreIngestion, StopImport, StopLogging, UpdateChannel, UpdateDashboard, UpdateEventDataStore, UpdateTrail.

SECURITY
  Read streams use fixed AWS CloudTrail JSON-RPC action names and SigV4 authentication. Resource-detail streams derive identifiers through connector-local discovery/fan-out; no raw AWS action, path, header, body, shell, file, SQL, or generic HTTP escape hatch is exposed.

  No CloudTrail write action is executable in the current connector surface. Future write enablement must preserve plan -> preview -> approval -> execute plus destructive confirmation metadata.

AGENT WORKFLOW
  1. Inspect metadata with pm connectors inspect aws-cloudtrail --json; this does not read credentials.
  2. Add credentials from environment variables or stdin only.
  3. Use pm etl catalog/read/run for the 19 implemented read streams.
  4. Treat provider query/direct-read commands and all write/admin actions as blocked/planned until a shared-runtime forwarding slice lands.
```
