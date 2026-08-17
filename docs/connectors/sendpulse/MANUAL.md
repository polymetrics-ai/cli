# pm connectors inspect sendpulse

```text
NAME
  pm connectors inspect sendpulse - SendPulse connector manual

SYNOPSIS
  pm connectors inspect sendpulse
  pm connectors inspect sendpulse --json
  pm credentials add <name> --connector sendpulse [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SendPulse address books, campaigns, senders, per-book emails, and the account blacklist, and writes address-book/sender/blacklist lifecycle mutations and campaign create/cancel actions through the SendPulse API.

ICON
  asset: icons/sendpulse.svg
  source: official
  review_status: official_verified
  review_url: https://sendpulse.com/integrations/api

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  addressbooks:
    primary key: id
    fields: all_email_qty(), id(), name()
  campaigns:
    primary key: id
    fields: id(), name(), status()
  senders:
    primary key: email
    fields: email(), name()
  blacklist:
    primary key: email
    fields: comment(), email()
  emails_in_book:
    primary key: email
    fields: book_id(), email(), status()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_addressbook:
    endpoint: POST /addressbooks
    risk: creates a new address book (mailing list); external mutation, approval required
  update_addressbook:
    endpoint: PUT /addressbooks/{{ record.id }}
    required fields: id
    risk: renames an existing address book
  delete_addressbook:
    endpoint: DELETE /addressbooks/{{ record.id }}
    required fields: id
    risk: permanently removes an address book and all its subscriber associations; irreversible
  add_emails_to_book:
    endpoint: POST /addressbooks/{{ record.id }}/emails
    required fields: id
    optional fields: emails
    risk: subscribes new email addresses to an address book; each add may trigger a double opt-in confirmation email depending on account settings
  remove_emails_from_book:
    endpoint: DELETE /addressbooks/{{ record.id }}/emails
    required fields: id
    optional fields: emails
    risk: unsubscribes the given email addresses from an address book; irreversible without re-adding them
  create_campaign:
    endpoint: POST /campaigns
    risk: creates a new email campaign against a real address book; depending on account settings this may schedule actual sending to real subscribers, the highest-impact action in this bundle, approval required
  cancel_campaign:
    endpoint: DELETE /campaigns/{{ record.id }}
    required fields: id
    risk: cancels a scheduled/in-progress campaign; stops further sends but does not un-send already-delivered emails
  add_sender:
    endpoint: POST /senders
    risk: registers a new sender email address, which SendPulse will send an activation email to; low-risk external mutation
  remove_sender:
    endpoint: DELETE /senders
    optional fields: email
    risk: removes a sender email address; any campaign still referencing it as its sender will fail to send
  add_to_blacklist:
    endpoint: POST /blacklist
    risk: permanently suppresses future sends to the given address(es) account-wide
  remove_from_blacklist:
    endpoint: DELETE /blacklist
    optional fields: emails
    risk: removes an address from the account-wide suppression list; future campaigns can reach it again

SECURITY
  read risk: external SendPulse API read of address book, campaign, sender, per-book-email, and blacklist data
  write risk: external SendPulse API mutation, including creating email campaigns that may send to real subscribers
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sendpulse

  # Inspect as structured JSON
  pm connectors inspect sendpulse --json

AGENT WORKFLOW
  - Run pm connectors inspect sendpulse before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
