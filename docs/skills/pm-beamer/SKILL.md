---
name: pm-beamer
description: Beamer connector knowledge and safe action guide.
---

# pm-beamer

## Purpose

Reads and writes Beamer NPS survey responses, announcement posts, feature requests, comments, reactions, votes, and end users through the Beamer REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- api_key (secret) (required)

## ETL Streams

- nps:
  - primary key: id
  - cursor: date
  - fields: browser(string), city(string), country(string), date(string), feedback(string), filter(string), id(string), language(string), origin(string), os(string), refUrl(string), score(number), url(string), userEmail(string), userFirstName(string), userId(string), userLastName(string)
- posts:
  - primary key: id
  - cursor: date
  - fields: category(string), clicks(integer), content(string), date(string), feedbackEnabled(boolean), id(string), published(boolean), reactionsEnabled(boolean), title(string), url(string), views(integer)
- feature_requests:
  - primary key: id
  - cursor: date
  - fields: commentsCount(integer), content(string), date(string), id(string), status(string), title(string), url(string), userEmail(string), userId(string), votesCount(integer)
- comments:
  - primary key: id
  - cursor: date
  - fields: content(string), date(string), featureRequestId(string), id(string), postId(string), userEmail(string), userFirstName(string), userId(string), userLastName(string)
- post_reactions:
  - primary key: id
  - cursor: date
  - fields: date(string), id(string), postTitle(string), post_id(string), reaction(string), url(string), userEmail(string), userFirstName(string), userId(string), userLastName(string)
- feature_request_votes:
  - primary key: id
  - cursor: date
  - fields: date(string), featureRequestTitle(string), feature_request_id(string), id(string), url(string), userEmail(string), userFirstName(string), userId(string), userLastName(string)
- users:
  - primary key: beamerId
  - fields: beamerId(string), browser(string), city(string), country(string), filter(string), firstSeen(string), ip(string), language(string), lastSeen(string), latitude(string), longitude(string), os(string), userEmail(string), userFirstName(string), userId(string), userLastName(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_post:
  - endpoint: POST /posts
  - required fields: title, content
  - risk: external mutation; creates a new Beamer announcement post, optionally published immediately (visible to end users); approval required
- update_post:
  - endpoint: PUT /posts/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates a live announcement post visible to end users; approval required
- delete_post:
  - endpoint: DELETE /posts/{{ record.id }}
  - required fields: id
  - risk: permanently removes an announcement post; irreversible; approval required
- create_post_comment:
  - endpoint: POST /posts/{{ record.post_id }}/comments
  - required fields: post_id, text
  - risk: external mutation; adds a comment to a live announcement post on behalf of a user; approval required
- delete_post_comment:
  - endpoint: DELETE /posts/{{ record.post_id }}/comments/{{ record.id }}
  - required fields: post_id, id
  - risk: permanently removes a comment from a post; irreversible; approval required
- create_feature_request:
  - endpoint: POST /feature-requests
  - required fields: title, content
  - risk: external mutation; creates a new feature request, optionally visible immediately to end users; approval required
- update_feature_request:
  - endpoint: PUT /feature-requests/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates a feature request visible to end users (status changes are commonly user-facing); approval required
- delete_feature_request:
  - endpoint: DELETE /feature-requests/{{ record.id }}
  - required fields: id
  - risk: permanently removes a feature request; irreversible; approval required
- create_feature_request_comment:
  - endpoint: POST /feature-requests/{{ record.feature_request_id }}/comments
  - required fields: feature_request_id, text
  - risk: external mutation; adds a comment to a feature request on behalf of a user; approval required
- delete_feature_request_comment:
  - endpoint: DELETE /feature-requests/{{ record.feature_request_id }}/comments/{{ record.id }}
  - required fields: feature_request_id, id
  - risk: permanently removes a comment from a feature request; irreversible; approval required
- create_post_reaction:
  - endpoint: POST /posts/{{ record.post_id }}/reactions
  - required fields: post_id, reaction
  - risk: external mutation; records a reaction to a post on behalf of a user; approval required
- delete_post_reaction:
  - endpoint: DELETE /posts/{{ record.post_id }}/reactions/{{ record.id }}
  - required fields: post_id, id
  - risk: permanently removes a reaction from a post; irreversible; approval required
- create_feature_request_vote:
  - endpoint: POST /feature-requests/{{ record.feature_request_id }}/votes
  - required fields: feature_request_id
  - risk: external mutation; records a vote for a feature request on behalf of a user; approval required
- delete_feature_request_vote:
  - endpoint: DELETE /feature-requests/{{ record.feature_request_id }}/votes/{{ record.id }}
  - required fields: feature_request_id, id
  - risk: permanently removes a vote from a feature request; irreversible; approval required

## Security

- read risk: external Beamer API read of NPS feedback, posts, feature requests, comments, reactions, votes, and end users
- write risk: external mutation of Beamer posts, feature requests, comments, reactions, and votes; a published post/feature-request write is immediately end-user-visible in the customer-facing widget
- approval: required for every write action; see writes.json risk field per action
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect beamer
```

### Inspect as structured JSON

```bash
pm connectors inspect beamer --json
```

## Agent Rules

- Run pm connectors inspect beamer before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
