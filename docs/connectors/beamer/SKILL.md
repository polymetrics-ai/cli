---
name: pm-beamer
description: Beamer connector knowledge and safe action guide.
---

# pm-beamer

## Purpose

Reads and writes Beamer NPS survey responses, announcement posts, feature requests, comments, reactions, votes, and end users through the Beamer REST API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- api_key (secret)

## ETL Streams

- nps:
  - primary key: id
  - cursor: date
  - fields: browser(), city(), country(), date(), feedback(), filter(), id(), language(), origin(), os(), refUrl(), score(), url(), userEmail(), userFirstName(), userId(), userLastName()
- posts:
  - primary key: id
  - cursor: date
  - fields: category(), clicks(), content(), date(), feedbackEnabled(), id(), published(), reactionsEnabled(), title(), url(), views()
- feature_requests:
  - primary key: id
  - cursor: date
  - fields: commentsCount(), content(), date(), id(), status(), title(), url(), userEmail(), userId(), votesCount()
- comments:
  - primary key: id
  - cursor: date
  - fields: content(), date(), featureRequestId(), id(), postId(), userEmail(), userFirstName(), userId(), userLastName()
- post_reactions:
  - primary key: id
  - cursor: date
  - fields: date(), id(), postTitle(), post_id(), reaction(), url(), userEmail(), userFirstName(), userId(), userLastName()
- feature_request_votes:
  - primary key: id
  - cursor: date
  - fields: date(), featureRequestTitle(), feature_request_id(), id(), url(), userEmail(), userFirstName(), userId(), userLastName()
- users:
  - primary key: beamerId
  - fields: beamerId(), browser(), city(), country(), filter(), firstSeen(), ip(), language(), lastSeen(), latitude(), longitude(), os(), userEmail(), userFirstName(), userId(), userLastName()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_post:
  - endpoint: POST /posts
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
  - required fields: post_id
  - risk: external mutation; adds a comment to a live announcement post on behalf of a user; approval required
- delete_post_comment:
  - endpoint: DELETE /posts/{{ record.post_id }}/comments/{{ record.id }}
  - required fields: post_id, id
  - risk: permanently removes a comment from a post; irreversible; approval required
- create_feature_request:
  - endpoint: POST /feature-requests
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
  - required fields: feature_request_id
  - risk: external mutation; adds a comment to a feature request on behalf of a user; approval required
- delete_feature_request_comment:
  - endpoint: DELETE /feature-requests/{{ record.feature_request_id }}/comments/{{ record.id }}
  - required fields: feature_request_id, id
  - risk: permanently removes a comment from a feature request; irreversible; approval required
- create_post_reaction:
  - endpoint: POST /posts/{{ record.post_id }}/reactions
  - required fields: post_id
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
