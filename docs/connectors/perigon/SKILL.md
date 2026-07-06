---
name: pm-perigon
description: Perigon connector knowledge and safe action guide.
---

# pm-perigon

## Purpose

Reads Perigon news articles, story clusters, journalists, sources, companies, people, and topics through the Perigon REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- query
- start_date
- api_key (secret)

## ETL Streams

- articles:
  - primary key: article_id
  - cursor: pub_date
  - fields: article_id(), pub_date(), source(), title(), url()
- stories:
  - primary key: id
  - fields: createdAt(), id(), name(), updatedAt()
- journalists:
  - primary key: id
  - fields: avgMonthlyPosts(), description(), facebookUrl(), fullName(), headline(), id(), imageUrl(), instagramUrl(), linkedinUrl(), locations(), name(), title(), topCategories(), topCountries(), topLabels(), topSources(), topTopics(), twitterBio(), twitterHandle(), updatedAt(), websiteUrl()
- sources:
  - primary key: id
  - fields: adFontesBiasRating(), allSidesBiasRating(), altNames(), avgBiasRating(), avgMonthlyPosts(), description(), domain(), globalRank(), id(), location(), mbfcBiasRating(), monthlyVisits(), name(), paywall(), topCategories(), topCountries(), topLabels(), topTopics(), updatedAt()
- companies:
  - primary key: id
  - fields: address(), altNames(), ceo(), city(), country(), description(), domains(), fullTimeEmployees(), globalRank(), id(), industry(), isActivelyTrading(), isAdr(), isEtf(), isFund(), monthlyVisits(), name(), revenue(), sector(), state(), symbols(), updatedAt(), webResources(), yearFounded(), zip()
- people:
  - primary key: wikidataId
  - fields: abstract(), aliases(), createdAt(), dateOfBirth(), dateOfDeath(), description(), gender(), image(), name(), occupation(), politicalParty(), position(), updatedAt(), wikidataId()
- topics:
  - primary key: name
  - fields: category(), labels(), name(), subcategory()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Perigon API read of public news article, story, journalist, source, company, people, and topic data
- approval: none; read-only public news API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect perigon
```

### Inspect as structured JSON

```bash
pm connectors inspect perigon --json
```

## Agent Rules

- Run pm connectors inspect perigon before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
