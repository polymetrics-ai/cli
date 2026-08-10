# pm connectors inspect countercyclical

```text
NAME
  pm connectors inspect countercyclical - Countercyclical connector manual

SYNOPSIS
  pm connectors inspect countercyclical
  pm connectors inspect countercyclical --json
  pm credentials add <name> --connector countercyclical [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Countercyclical investments, valuations, research memos, teams, assumptions, and pipelines, and creates investments, through the Countercyclical REST API.

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
  mode
  api_key (secret)

ETL STREAMS
  investments:
    primary key: id
    fields: cik(string), country(string), createdAt(string), description(string), editedName(string), employees(integer), exchange(string), figi(string), financingType(string), id(string), industry(string), isArchived(boolean), isFavorite(boolean), isLocked(boolean), lei(string), marketType(string), name(string), sector(string), tickerSymbol(string), updatedAt(string), visibility(string), website(string)
  valuations:
    primary key: id
    fields: createdAt(string), delineation(string), description(string), discountRate(number), endingQuarter(integer), endingYear(integer), growthMetric(string), growthRate(number), id(string), isFavorite(boolean), name(string), shareToken(string), startingQuarter(integer), startingYear(integer), status(string), terminalPeriod(string), terminalRate(number), updatedAt(string)
  memos:
    primary key: id
    fields: archived(boolean), backgroundColor(string), bannerVisible(boolean), body(string), createdAt(string), documentType(string), emoji(string), favorited(boolean), foregroundColor(string), id(string), locked(boolean), publiclyVisible(boolean), sourcesVisible(boolean), title(string), tocVisible(boolean), updatedAt(string), views(integer)
  teams:
    primary key: id
    fields: id(string), title(string)
  assumptions:
    primary key: id
    fields: discountRate(string), id(string), name(string)
  pipelines:
    primary key: id
    fields: id(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_investment:
    endpoint: POST /integrations/make/actions/investments
    required fields: tickerSymbol
    risk: creates a new Investment in the caller's Countercyclical workspace via the Make-integration action endpoint (the only documented general-purpose creation endpoint; the functionally-identical Zapier-integration endpoint is not separately exposed, see api_surface.json); external mutation, no approval required

SECURITY
  read risk: external Countercyclical API read of investment and valuation data
  write risk: external mutation: creates a new Investment record in the caller's workspace; no update/delete actions are exposed
  approval: required for the create_investment write action; read-only otherwise
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Countercyclical's declared streams and reverse-ETL actions.
  Usage: pm countercyclical <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api get assumptions - Documented GET /assumptions (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.assumptions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get integrations make auth - Documented GET /integrations/make/auth (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.integrations-make-auth]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get integrations make triggers investments - Documented GET /integrations/make/triggers/investments (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.integrations-make-triggers-investments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get integrations zapier auth - Documented GET /integrations/zapier/auth (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.integrations-zapier-auth]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get integrations zapier triggers investments - Documented GET /integrations/zapier/triggers/investments (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.integrations-zapier-triggers-investments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get investments - Documented GET /investments (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.investments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get memos - Documented GET /memos (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.memos]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get pipelines - Documented GET /pipelines (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.pipelines]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get teams - Documented GET /teams (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.teams]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 integrations make auth - Documented GET /v1/integrations/make/auth (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.v1-integrations-make-auth]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 integrations make triggers investments - Documented GET /v1/integrations/make/triggers/investments (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.v1-integrations-make-triggers-investments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 integrations zapier auth - Documented GET /v1/integrations/zapier/auth (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.v1-integrations-zapier-auth]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 integrations zapier triggers investments - Documented GET /v1/integrations/zapier/triggers/investments (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.v1-integrations-zapier-triggers-investments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get valuations - Documented GET /valuations (not implemented) [intent=direct_read availability=not_implemented operation=countercyclical.get.valuations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post integrations make actions investments - Documented POST /integrations/make/actions/investments (not implemented) [intent=direct_write availability=not_implemented operation=countercyclical.post.integrations-make-actions-investments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post integrations zapier actions investments - Documented POST /integrations/zapier/actions/investments (not implemented) [intent=direct_write availability=not_implemented operation=countercyclical.post.integrations-zapier-actions-investments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 integrations zapier actions investments - Documented POST /v1/integrations/zapier/actions/investments (not implemented) [intent=direct_write availability=not_implemented operation=countercyclical.post.v1-integrations-zapier-actions-investments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 webhooks - Documented POST /v1/webhooks (not implemented) [intent=direct_write availability=not_implemented operation=countercyclical.post.v1-webhooks]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    assumptions list - Run the assumptions ETL stream [intent=etl availability=implemented stream=assumptions]; notes: discrepancy=present-in-surface-absent-from-artifact
    create investment apply - Plan and execute the create investment reverse-ETL action [intent=reverse_etl availability=implemented write=create_investment]; approval: requires plan, preview, approval, and execute; risk: creates a new Investment in the caller's Countercyclical workspace via the Make-integration action endpoint (the only documented general-purpose creation endpoint; the functionally-identical Zapier-integration endpoint is not separately exposed, see api_surface.json); external mutation, no approval required; flags: --tickerSymbol (required)
    investments list - Run the investments ETL stream [intent=etl availability=implemented stream=investments]; notes: discrepancy=present-in-surface-absent-from-artifact
    memos list - Run the memos ETL stream [intent=etl availability=implemented stream=memos]; notes: discrepancy=present-in-surface-absent-from-artifact
    pipelines list - Run the pipelines ETL stream [intent=etl availability=implemented stream=pipelines]; notes: discrepancy=present-in-surface-absent-from-artifact
    teams list - Run the teams ETL stream [intent=etl availability=implemented stream=teams]; notes: discrepancy=present-in-surface-absent-from-artifact
    valuations list - Run the valuations ETL stream [intent=etl availability=implemented stream=valuations]; notes: discrepancy=present-in-surface-absent-from-artifact

EXAMPLES
  # Inspect as a manual
  pm connectors inspect countercyclical

  # Inspect as structured JSON
  pm connectors inspect countercyclical --json

AGENT WORKFLOW
  - Run pm connectors inspect countercyclical before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
