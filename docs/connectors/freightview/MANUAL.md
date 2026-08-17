# pm connectors inspect freightview

```text
NAME
  pm connectors inspect freightview - Freightview connector manual

SYNOPSIS
  pm connectors inspect freightview
  pm connectors inspect freightview --json
  pm credentials add <name> --connector freightview [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Freightview shipments, quotes, and tracking events through the Freightview v2.0 REST API using the client-credentials session-token flow. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  shipments:
    primary key: shipmentId
    fields: billTo(), bol(), bookedBy(), bookedDate(), createdDate(), direction(), documents(), equipment(), isArchived(), isLiveLoad(), items(), locations(), pickup(), pickupDate(), quotedBy(), refNums(), selectedQuote(), shipmentId(), status(), tracking()
  quotes:
    primary key: quoteId
    fields: amount(), carrierId(), createdDate(), currency(), equipmentType(), method(), mode(), paymentTerms(), pricingMethod(), pricingType(), providerCode(), providerName(), quoteId(), quoteNum(), serviceId(), source(), status()
  tracking:
    primary key: createdDate
    fields: createdDate(), eventDate(), eventTime(), eventType(), summary()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Freightview API reads performed by the legacy connector via a Tier-2 hook
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect freightview

  # Inspect as structured JSON
  pm connectors inspect freightview --json

AGENT WORKFLOW
  - Run pm connectors inspect freightview before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
