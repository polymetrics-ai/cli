# pm connectors inspect freightview

```text
NAME
  pm connectors inspect freightview - Freightview connector manual

SYNOPSIS
  pm connectors inspect freightview
  pm connectors inspect freightview --json
  pm credentials add <name> --connector freightview [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Freightview shipments, quotes, and tracking events through fixed Freightview v2.0 REST routes using client-credentials authentication.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  client_id (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  shipments:
    primary key: shipmentId
    fields: billTo(object), bol(object), bookedBy(string), bookedDate(string), createdDate(string), direction(string), documents(array), equipment(object), isArchived(boolean), isLiveLoad(boolean), items(array), locations(array), pickup(object), pickupDate(string), quotedBy(string), refNums(array), selectedQuote(object), shipmentId(string), status(string), tracking(object)
  quotes:
    primary key: quoteId
    fields: amount(number), carrierId(string), createdDate(string), currency(string), equipmentType(string), method(string), mode(string), paymentTerms(string), pricingMethod(string), pricingType(string), providerCode(string), providerName(string), quoteId(string), quoteNum(string), serviceId(string), source(string), status(string)
  tracking:
    primary key: createdDate
    fields: createdDate(string), eventDate(string), eventTime(string), eventType(string), summary(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

SECURITY
  read risk: Bounded Freightview v2.0 reads use declared client-credentials authentication and fixed provider routes.
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
