# pm connectors inspect aws-cloudtrail

```text
NAME
  aws-cloudtrail - AWS CloudTrail

SYNOPSIS
  pm credentials add <name> --connector aws-cloudtrail --config aws_region_name=<region> --from-env aws_key_id=AWS_ACCESS_KEY_ID --from-env aws_secret_key=AWS_SECRET_ACCESS_KEY
  pm etl catalog --connector aws-cloudtrail --credential <name> --json
  pm etl read --connector aws-cloudtrail --credential <name> --stream <stream> --json

DESCRIPTION
  Reads AWS CloudTrail configuration lists through fixed AWS JSON-RPC streams that need no per-call resource identifiers. Provider query/direct-read, parameterized read, and write/admin actions remain planned until safe shared forwarding exists.

  Scope-corrected status: 60 official CloudTrail API actions remain inventoried, but only 9 read-stream actions are currently exposed as executable connector-local runtime behavior. The 10 provider query/direct-read actions, 10 parameterized read actions, and 31 write/admin actions are blocked/planned until shared promoted-native forwarding or a typed request-parameter boundary exposes them safely.

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
  get_event_configuration
  get_insight_selectors
  list_channels
  list_dashboards
  list_event_data_stores
  list_imports
  list_public_keys
  list_trails

BLOCKED / PLANNED OPERATIONS
  Parameterized read actions blocked: GetChannel, GetDashboard, GetEventDataStore, GetEventSelectors, GetImport, GetResourcePolicy, GetTrail, GetTrailStatus, ListImportFailures, ListTags.

  Provider query/direct-read actions blocked: CancelQuery, DescribeQuery, GenerateQuery, GetQueryResults, ListInsightsData, ListInsightsMetricData, ListQueries, LookupEvents, SearchSampleQueries, StartQuery.

  Write/admin actions blocked: AddTags, CreateChannel, CreateDashboard, CreateEventDataStore, CreateTrail, DeleteChannel, DeleteDashboard, DeleteEventDataStore, DeleteResourcePolicy, DeleteTrail, DeregisterOrganizationDelegatedAdmin, DisableFederation, EnableFederation, PutEventConfiguration, PutEventSelectors, PutInsightSelectors, PutResourcePolicy, RegisterOrganizationDelegatedAdmin, RemoveTags, RestoreEventDataStore, StartDashboardRefresh, StartEventDataStoreIngestion, StartImport, StartLogging, StopEventDataStoreIngestion, StopImport, StopLogging, UpdateChannel, UpdateDashboard, UpdateEventDataStore, UpdateTrail.

SECURITY
  Read streams use fixed AWS CloudTrail JSON-RPC action names and SigV4 authentication. No raw AWS action, path, header, body, shell, file, SQL, or generic HTTP escape hatch is exposed.

  No CloudTrail write action is executable in this corrective head. Future write enablement must preserve plan -> preview -> approval -> execute plus destructive confirmation metadata.

AGENT WORKFLOW
  1. Inspect metadata with pm connectors inspect aws-cloudtrail --json; this does not read credentials.
  2. Add credentials from environment variables or stdin only.
  3. Use pm etl catalog/read/run for the 9 implemented read streams.
  4. Treat parameterized reads, provider query/direct-read commands, and all write/admin actions as blocked/planned until a shared-runtime forwarding or typed request-parameter slice lands.
```
