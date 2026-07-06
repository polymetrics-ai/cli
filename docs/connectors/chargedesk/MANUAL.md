# pm connectors inspect chargedesk

```text
NAME
  pm connectors inspect chargedesk - ChargeDesk connector manual

SYNOPSIS
  pm connectors inspect chargedesk
  pm connectors inspect chargedesk --json
  pm credentials add <name> --connector chargedesk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads ChargeDesk charges, customers, subscriptions, and products through the ChargeDesk REST API.

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
  mode
  username
  password (secret)

ETL STREAMS
  charges:
    primary key: charge_id
    cursor: occurred
    fields: amount(), amount_refunded(), charge_id(), currency(), customer_email(), customer_id(), customer_name(), description(), object(), occurred(), payment_method(), product_id(), status(), subscription_id(), transaction_id()
  customers:
    primary key: customer_id
    cursor: occurred
    fields: country(), currency(), customer_id(), delinquent(), email(), name(), object(), occurred(), phone(), tax_number()
  subscriptions:
    primary key: subscription_id
    cursor: occurred
    fields: amount(), currency(), current_period_end(), current_period_start(), customer_id(), interval(), object(), occurred(), product_id(), status(), subscription_id()
  products:
    primary key: product_id
    cursor: occurred
    fields: amount(), currency(), interval(), name(), object(), occurred(), product_id(), status()
  log_activity:
    cursor: occurred
    fields: action_params(), action_reason(), action_type(), company(), context(), description(), event(), ip(), object_id(), object_type(), occurred(), params(), source(), sub_description()
  log_cancellations:
    cursor: occurred
    fields: action(), customer_id(), email(), ip(), method(), occurred(), reason(), subscription_id()
  webhook_notifications:
    primary key: notification
    fields: description(), name(), notification(), object()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers
    risk: external mutation creating a new ChargeDesk customer record; approval required
  update_customer:
    endpoint: POST /customers/{{ record.customer_id }}
    required fields: customer_id
    risk: external mutation updating an existing ChargeDesk customer record; approval required
  delete_customer:
    endpoint: DELETE /customers/{{ record.customer_id }}
    required fields: customer_id
    risk: irreversible deletion of a customer record (and, by ChargeDesk's own default, all associated charges/tickets); approval required
  update_charge:
    endpoint: POST /charges/{{ record.charge_id }}
    required fields: charge_id
    risk: external mutation updating an existing charge record's stored data; approval required
  delete_charge:
    endpoint: DELETE /charges/{{ record.charge_id }}
    required fields: charge_id
    risk: irreversible deletion of a charge record; approval required
  refund_charge:
    endpoint: POST /gateway/charges/{{ record.charge_id }}/refund
    required fields: charge_id
    risk: gateway method; irreversibly refunds a charge (full or partial) on the originating payment gateway as well as ChargeDesk; approval required
  capture_charge:
    endpoint: POST /gateway/charges/{{ record.charge_id }}/capture
    required fields: charge_id
    risk: gateway method; captures (settles) a previously authorized charge on the originating payment gateway; approval required
  void_charge:
    endpoint: POST /gateway/charges/{{ record.charge_id }}/void
    required fields: charge_id
    risk: gateway method; voids an authorized charge or cancels an outstanding payment request on the originating payment gateway; approval required
  cancel_subscription:
    endpoint: POST /gateway/subscriptions/{{ record.subscription_id }}/cancel
    required fields: subscription_id
    risk: gateway method; irreversibly cancels future recurring charges for a subscription on the originating payment gateway as well as ChargeDesk; approval required
  create_webhook:
    endpoint: POST /webhooks
    risk: external mutation creating a new outbound webhook subscription that will POST ChargeDesk event data to a third-party URL; approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: irreversible removal of an outbound webhook subscription; approval required
  create_agent:
    endpoint: POST /agents
    risk: external mutation inviting a new support agent (or updating an existing agent's role) with account access to ChargeDesk; approval required
  delete_agent:
    endpoint: DELETE /agents/{{ record.email }}
    required fields: email
    risk: irreversible removal of a support agent's ChargeDesk account access; approval required

SECURITY
  read risk: external ChargeDesk API read of billing/charge, customer, subscription, product, activity-log, and cancellation-log data
  write risk: external mutations creating/updating/deleting customers, charges, webhooks, and agents, plus live gateway methods (refund/capture/void a charge, cancel a subscription) that mutate the connected payment gateway; every write action requires approval
  approval: read: none; write: required for every action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect chargedesk

  # Inspect as structured JSON
  pm connectors inspect chargedesk --json

AGENT WORKFLOW
  - Run pm connectors inspect chargedesk before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
