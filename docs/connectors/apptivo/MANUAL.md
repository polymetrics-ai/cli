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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  access_key (secret)
  api_key (secret)

ETL STREAMS
  customers:
    primary key: customerId
    fields: creationDate(), currencyCode(), customerId(), customerName(), customerNumber(), emailAddress(), lastUpdateDate(), phoneNumber(), statusName(), website()
  contacts:
    primary key: contactId
    fields: companyName(), contactId(), creationDate(), emailAddress(), firstName(), fullName(), lastName(), lastUpdateDate(), phoneNumber()
  leads:
    primary key: id
    fields: companyName(), creationDate(), emailAddress(), firstName(), id(), lastName(), leadId(), leadSource(), phoneNumber(), statusName()
  opportunities:
    primary key: opportunityId
    fields: closingDate(), creationDate(), currencyCode(), customerName(), lastUpdateDate(), opportunityAmount(), opportunityId(), opportunityName(), salesStageName()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

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
