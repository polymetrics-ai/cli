# pm connectors inspect apptivo

```text
NAME
  pm connectors inspect apptivo - Apptivo connector manual

SYNOPSIS
  pm connectors inspect apptivo
  pm connectors inspect apptivo --json
  pm credentials add <name> --connector apptivo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Apptivo CRM customers, contacts, leads, and opportunities through the Apptivo REST DAO API (full refresh); deletes CRM customer records via the documented deleteCustomer DAO action.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  access_key (secret) (required)
  api_key (secret) (required)

ETL STREAMS
  customers:
    primary key: customerId
    fields: creationDate(string), currencyCode(string), customerId(string), customerName(string), customerNumber(string), emailAddress(string), lastUpdateDate(string), phoneNumber(string), statusName(string), website(string)
  contacts:
    primary key: contactId
    fields: companyName(string), contactId(string), creationDate(string), emailAddress(string), firstName(string), fullName(string), lastName(string), lastUpdateDate(string), phoneNumber(string)
  leads:
    primary key: id
    fields: companyName(string), creationDate(string), emailAddress(string), firstName(string), id(string), lastName(string), leadId(string), leadSource(string), phoneNumber(string), statusName(string)
  opportunities:
    primary key: opportunityId
    fields: closingDate(string), creationDate(string), currencyCode(string), customerName(string), lastUpdateDate(string), opportunityAmount(string), opportunityId(string), opportunityName(string), salesStageName(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  remove_customer:
    endpoint: GET /app/dao/v6/customers?a=delete&customerId={{ record.id }}&apiKey={{ secrets.api_key }}&accessKey={{ secrets.access_key }}
    required fields: id
    risk: irreversible external deletion of a CRM customer record; approval required

SECURITY
  read risk: external Apptivo API read of CRM customer, contact, lead, and opportunity data
  write risk: external mutation: irreversibly deletes a CRM customer record
  approval: required; remove_customer is a destructive, irreversible external deletion
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect apptivo

  # Inspect as structured JSON
  pm connectors inspect apptivo --json

AGENT WORKFLOW
  - Run pm connectors inspect apptivo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
