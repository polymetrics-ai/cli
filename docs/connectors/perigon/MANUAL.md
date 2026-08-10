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
  api_key (secret) (required)

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
