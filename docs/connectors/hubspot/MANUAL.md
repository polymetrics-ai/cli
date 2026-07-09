# pm connectors inspect hubspot

```text
NAME
  pm connectors inspect hubspot - HubSpot connector manual

SYNOPSIS
  pm connectors inspect hubspot
  pm connectors inspect hubspot --json
  pm credentials add <name> --connector hubspot [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  HubSpot public API command-surface metadata scaffold. Safe execution lanes are added by the HubSpot CLI parity sub-issues.

ICON
  asset: icons/hubspot.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.hubspot.com/docs/api/overview

CAPABILITIES
  check=true catalog=true read=false write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  private_app_token (secret)

SECURITY
  read risk: metadata-only scaffold; HubSpot read execution is added through typed stream and direct-read lanes
  write risk: metadata-only scaffold; HubSpot mutations must be modeled as named reverse ETL actions before execution
  approval: reverse ETL writes require plan, preview, approval, execute; destructive/admin/sensitive actions require typed confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with HubSpot CRM, marketing, commerce, files, and settings APIs from typed connector commands.
  Usage: pm hubspot <group> <resource> <action> [flags]
  Source CLI: HubSpot public OpenAPI collection (https://github.com/HubSpot/HubSpot-public-api-spec-collection)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved HubSpot connector credential.: maps_to=connection
    --batch-size (integer): Bound ETL reads with an explicit batch size.: maps_to=read.batch_size
  CRM Commands
    crm contacts list - List CRM contacts [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns executable stream coverage.; flags: --properties, --limit
    crm contacts view - View one CRM contact [intent=direct_read availability=planned]; notes: Planned bounded direct-read command; issue #138 owns output policy and redaction.; flags: --contact-id, --properties
    crm contacts search - Search CRM contacts with a fixed request schema [intent=direct_read availability=planned]; notes: Planned POST search direct-read; issue #139 owns fixed body schemas and rejects arbitrary JSON bodies.; flags: --filter, --sort
    crm contacts create - Create a CRM contact [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a contact record in the configured HubSpot portal.; notes: Planned named reverse ETL action; no direct write execution is exposed.; flags: --email, --firstname, --lastname
    crm contacts update - Update a CRM contact [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates properties on an existing HubSpot contact.; notes: Planned named reverse ETL action with a fixed record schema.; flags: --contact-id, --property
    crm contacts delete - Archive a CRM contact [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute plus typed confirmation for destructive actions.; risk: Archives a contact; destructive/admin policy must require typed confirmation before dispatch.; notes: Planned destructive reverse ETL action; issue #140 owns confirmation and redaction policy.; flags: --contact-id
    crm companies list - List CRM companies [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    crm deals list - List CRM deals [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    crm tickets list - List CRM tickets [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    crm line-items list - List CRM line items [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    crm products list - List CRM products [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    crm quotes list - List CRM quotes [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    crm properties list - List CRM object properties [intent=direct_read availability=planned]; notes: Planned bounded direct-read command; issue #138 owns output policy.; flags: --object-type
    crm pipelines list - List CRM pipelines [intent=direct_read availability=planned]; notes: Planned bounded direct-read command; issue #138 owns output policy.; flags: --object-type
    crm associations list - List CRM association labels [intent=direct_read availability=planned]; notes: Planned bounded direct-read command; issue #138 owns output policy.; flags: --from-object-type, --to-object-type
    crm owners list - List HubSpot owners [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
  Marketing Commands
    marketing forms list - List marketing forms [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    marketing lists list - List HubSpot lists [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    marketing emails list - List marketing emails [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
  Commerce Commands
    commerce payment-links list - List commerce payment links [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    commerce invoices list - List commerce invoices [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
    commerce discounts list - List commerce discounts [intent=etl availability=planned]; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
  Files Commands
    files list - List HubSpot file metadata [intent=etl availability=planned]; notes: Planned stream-backed file metadata command; issue #136 owns schemas and fixtures.
    files download - Download a HubSpot file with a bounded destination policy [intent=binary availability=planned]; approval: Binary file transfers require explicit destination approval and bounded max bytes.; risk: Downloads binary content; destination path, max bytes, no-overwrite, and redaction policy must be enforced before execution.; notes: Planned binary command; issue #139 owns bounded binary executor and destination policy.; flags: --file-id, --output
    files upload - Upload a file to HubSpot [intent=reverse_etl availability=planned]; approval: Reverse ETL/file writes require plan, preview, approval, execute and bounded file policy.; risk: Uploads local file content to HubSpot; source path and max bytes must be explicitly approved.; notes: Planned file-upload reverse ETL action; no raw upload execution is exposed.; flags: --file, --folder-id
  Settings And Admin Commands
    settings users list - List HubSpot users [intent=direct_read availability=planned]; risk: Lists account user metadata; output redaction policy must avoid exposing sensitive admin details.; notes: Planned bounded direct-read command; issue #138 owns output policy.
    settings users create - Provision a HubSpot user [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute plus typed admin confirmation.; risk: Admin action that can grant portal access; must require explicit scopes and typed confirmation.; notes: Planned admin reverse ETL action; issue #140 owns sensitive/admin policy.; flags: --email, --role-id
    settings business-units list - List business units [intent=direct_read availability=planned]; notes: Planned bounded direct-read command; issue #138 owns output policy.
    automation flows list - List automation flows [intent=etl availability=planned]; risk: Workflow metadata may reveal automation structure; output policy must avoid sensitive payloads.; notes: Planned stream-backed command; issue #136 owns schemas and fixtures.
  Help topics:
    hubspot-safety - HubSpot writes are reverse ETL only: plan, preview, approval, execute, with typed confirmation for destructive/admin/sensitive actions.
    hubspot-official-inventory - Full operation ledger is owned by issue #137 and must account for all 3,060 unique official method/path operations.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect hubspot

  # Inspect as structured JSON
  pm connectors inspect hubspot --json

AGENT WORKFLOW
  - Run pm connectors inspect hubspot before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
