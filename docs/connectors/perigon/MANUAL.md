# pm connectors inspect perigon

```text
NAME
  pm connectors inspect perigon - Perigon connector manual

SYNOPSIS
  pm connectors inspect perigon
  pm connectors inspect perigon --json
  pm credentials add <name> --connector perigon [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Perigon news articles, story clusters, journalists, sources, companies, people, and topics through the Perigon REST API.

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
  base_url
  max_pages
  mode
  page_size
  query
  start_date
  api_key (secret)

ETL STREAMS
  articles:
    primary key: article_id
    cursor: pub_date
    fields: article_id(string), pub_date(string), source(object), title(string), url(string)
  stories:
    primary key: id
    fields: createdAt(string), id(string), name(string), updatedAt(string)
  journalists:
    primary key: id
    fields: avgMonthlyPosts(integer), description(string), facebookUrl(string), fullName(string), headline(string), id(string), imageUrl(string), instagramUrl(string), linkedinUrl(string), locations(array), name(string), title(string), topCategories(array), topCountries(array), topLabels(array), topSources(array), topTopics(array), twitterBio(string), twitterHandle(string), updatedAt(string), websiteUrl(string)
  sources:
    primary key: id
    fields: adFontesBiasRating(string), allSidesBiasRating(string), altNames(array), avgBiasRating(string), avgMonthlyPosts(integer), description(string), domain(string), globalRank(integer), id(string), location(object), mbfcBiasRating(string), monthlyVisits(integer), name(string), paywall(boolean), topCategories(array), topCountries(array), topLabels(array), topTopics(array), updatedAt(string)
  companies:
    primary key: id
    fields: address(string), altNames(array), ceo(string), city(string), country(string), description(string), domains(array), fullTimeEmployees(integer), globalRank(integer), id(string), industry(string), isActivelyTrading(boolean), isAdr(boolean), isEtf(boolean), isFund(boolean), monthlyVisits(integer), name(string), revenue(string), sector(string), state(string), symbols(array), updatedAt(string), webResources(object), yearFounded(integer), zip(string)
  people:
    primary key: wikidataId
    fields: abstract(string), aliases(array), createdAt(string), dateOfBirth(object), dateOfDeath(object), description(string), gender(object), image(object), name(string), occupation(array), politicalParty(array), position(array), updatedAt(string), wikidataId(string)
  topics:
    primary key: name
    fields: category(string), labels(object), name(string), subcategory(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Perigon API read of public news article, story, journalist, source, company, people, and topic data
  approval: none; read-only public news API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Perigon's declared streams and reverse-ETL actions.
  Usage: pm perigon <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete v1 api contactpoints uuid - Documented DELETE /v1/api/contactPoints/{uuid} (not implemented) [intent=direct_write availability=not_implemented operation=perigon.delete.v1-api-contactpoints-uuid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 api monitors uuid - Documented DELETE /v1/api/monitors/{uuid} (not implemented) [intent=direct_write availability=not_implemented operation=perigon.delete.v1-api-monitors-uuid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 api sourcegroups id - Documented DELETE /v1/api/sourceGroups/{id} (not implemented) [intent=direct_write availability=not_implemented operation=perigon.delete.v1-api-sourcegroups-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 api watchlists id - Documented DELETE /v1/api/watchlists/{id} (not implemented) [intent=direct_write availability=not_implemented operation=perigon.delete.v1-api-watchlists-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get v1 api contactpoints - Documented GET /v1/api/contactPoints (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-contactpoints]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api contactpoints uuid - Documented GET /v1/api/contactPoints/{uuid} (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-contactpoints-uuid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api monitors - Documented GET /v1/api/monitors (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-monitors]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api monitors uuid - Documented GET /v1/api/monitors/{uuid} (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-monitors-uuid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api monitors uuid events - Documented GET /v1/api/monitors/{uuid}/events (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-monitors-uuid-events]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api monitors uuid newsletters - Documented GET /v1/api/monitors/{uuid}/newsletters (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-monitors-uuid-newsletters]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api monitors uuid summary - Documented GET /v1/api/monitors/{uuid}/summary (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-monitors-uuid-summary]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api sourcegroups - Documented GET /v1/api/sourceGroups (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-sourcegroups]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api sourcegroups id - Documented GET /v1/api/sourceGroups/{id} (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-sourcegroups-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api sourcegroups resolve - Documented GET /v1/api/sourceGroups/resolve (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-sourcegroups-resolve]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api watchlists - Documented GET /v1/api/watchlists (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-watchlists]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api watchlists id - Documented GET /v1/api/watchlists/{id} (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-watchlists-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 api watchlists resolve - Documented GET /v1/api/watchlists/resolve (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-api-watchlists-resolve]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 articles refresh jobs id - Documented GET /v1/articles/refresh/jobs/{id} (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-articles-refresh-jobs-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 journalists id - Documented GET /v1/journalists/{id} (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-journalists-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 limits - Documented GET /v1/limits (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-limits]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 stories history - Documented GET /v1/stories/history (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-stories-history]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 stories stats - Documented GET /v1/stories/stats (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-stories-stats]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 wikipedia all - Documented GET /v1/wikipedia/all (not implemented) [intent=direct_read availability=not_implemented operation=perigon.get.v1-wikipedia-all]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch v1 api contactpoints uuid - Documented PATCH /v1/api/contactPoints/{uuid} (not implemented) [intent=direct_write availability=not_implemented operation=perigon.patch.v1-api-contactpoints-uuid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch v1 api monitors uuid - Documented PATCH /v1/api/monitors/{uuid} (not implemented) [intent=direct_write availability=not_implemented operation=perigon.patch.v1-api-monitors-uuid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch v1 api sourcegroups id - Documented PATCH /v1/api/sourceGroups/{id} (not implemented) [intent=direct_write availability=not_implemented operation=perigon.patch.v1-api-sourcegroups-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch v1 api watchlists id - Documented PATCH /v1/api/watchlists/{id} (not implemented) [intent=direct_write availability=not_implemented operation=perigon.patch.v1-api-watchlists-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 api contactpoints - Documented POST /v1/api/contactPoints (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-api-contactpoints]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 api monitors - Documented POST /v1/api/monitors (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-api-monitors]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 api monitors uuid activate - Documented POST /v1/api/monitors/{uuid}/activate (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-api-monitors-uuid-activate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 api monitors uuid pause - Documented POST /v1/api/monitors/{uuid}/pause (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-api-monitors-uuid-pause]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 api sourcegroups - Documented POST /v1/api/sourceGroups (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-api-sourcegroups]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 api watchlists - Documented POST /v1/api/watchlists (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-api-watchlists]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 articles refresh jobs - Documented POST /v1/articles/refresh/jobs (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-articles-refresh-jobs]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 articles refresh peek - Documented POST /v1/articles/refresh/peek (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-articles-refresh-peek]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 summarize - Documented POST /v1/summarize (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-summarize]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 vector news all - Documented POST /v1/vector/news/all (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-vector-news-all]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 vector wikipedia all - Documented POST /v1/vector/wikipedia/all (not implemented) [intent=direct_write availability=not_implemented operation=perigon.post.v1-vector-wikipedia-all]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    articles list - Run the articles ETL stream [intent=etl availability=implemented stream=articles]
    companies list - Run the companies ETL stream [intent=etl availability=implemented stream=companies]
    journalists list - Run the journalists ETL stream [intent=etl availability=implemented stream=journalists]
    people list - Run the people ETL stream [intent=etl availability=implemented stream=people]
    sources list - Run the sources ETL stream [intent=etl availability=implemented stream=sources]
    stories list - Run the stories ETL stream [intent=etl availability=implemented stream=stories]
    topics list - Run the topics ETL stream [intent=etl availability=implemented stream=topics]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect perigon

  # Inspect as structured JSON
  pm connectors inspect perigon --json

AGENT WORKFLOW
  - Run pm connectors inspect perigon before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
