---
name: pm-appfollow
description: Appfollow connector knowledge and safe action guide.
---

# pm-appfollow

## Purpose

Reads AppFollow account users, app collections, app lists, reviews, review summaries, ratings/ratings history, ASO keywords, rankings, and version/what's-new metadata through the AppFollow REST API v2 (config-list-driven fan-out per app/collection); writes review replies/tags/notes, ASO keyword edits, and account user/app/collection management actions.

## Icon

- asset: icons/appfollow.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://appfollow.docs.apiary.io/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- app_collection_ids
- base_url
- ext_ids
- report_country
- report_from
- report_store
- report_to
- api_secret (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: email(), id(), name(), role(), status(), updated()
- app_collections:
  - primary key: id
  - fields: count_apps(), countries(), created(), id(), languages(), title(), title_normalized()
- app_lists:
  - primary key: app_id
  - fields: app_collection_id(), app_id(), count_reviews(), count_whatsnew(), created(), ext_id(), is_favorite(), store(), watch_url()
- ratings:
  - primary key: ext_id, date, country
  - fields: country(), date(), ext_id(), rating(), stars1(), stars2(), stars3(), stars4(), stars5(), stars_total(), store(), version()
- reviews:
  - primary key: id
  - fields: app_id(), app_version(), author(), content(), created(), date(), dt(), ext_id(), id(), is_answer(), locale(), rating(), rating_prev(), review_id(), store(), time(), title(), updated(), user_id(), was_changed()
- reviews_summary:
  - primary key: ext_id, date, country
  - fields: avg_rating(), country(), date(), ext_id(), stars1(), stars2(), stars3(), stars4(), stars5(), total(), version()
- keywords:
  - primary key: ext_id, country, device, date
  - fields: country(), date(), device(), ext_id(), keyword(), no_pos(), page(), popularity(), pos(), store(), total()
- rankings:
  - primary key: ext_id, country, device, genre_id, date
  - fields: category(), country(), date(), device(), ext_id(), genre_id(), position()
- versions:
  - primary key: ext_id, version, country
  - fields: country(), ext_id(), release_date(), size(), version(), whats_new()
- versions_whatsnew:
  - primary key: ext_id, version, country
  - fields: country(), ext_id(), last_modified(), version(), whats_new()
- ratings_history:
  - primary key: ext_id, date, country, version
  - fields: avg_rating(), country(), date(), ext_id(), period(), stars(), stars1(), stars2(), stars3(), stars4(), stars5(), store(), total(), version()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- reply_to_review:
  - endpoint: POST /reviews/reply
  - risk: external mutation; posts a public reply to a live app-store review, cannot be unsent programmatically; approval required
- update_review_tags:
  - endpoint: POST /reviews/tags
  - risk: external mutation; overwrites a review's tag set; approval required
- update_review_notes:
  - endpoint: POST /reviews/notes
  - risk: external mutation; overwrites a review's internal note; approval required
- edit_keywords:
  - endpoint: POST /aso/keywords
  - risk: external mutation; replaces the tracked ASO keyword list for a country/device pair; approval required
- add_user:
  - endpoint: POST /account/users
  - risk: external mutation; grants AppFollow account access to a new user; approval required
- update_user:
  - endpoint: PATCH /account/users
  - risk: external mutation; changes an existing account user's role/access; approval required
- remove_user:
  - endpoint: DELETE /account/users
  - optional fields: id, email
  - risk: irreversible external mutation; revokes an AppFollow account user's access; approval required
- add_collection:
  - endpoint: POST /account/apps
  - risk: external mutation; creates a new billable app collection; approval required
- remove_collection:
  - endpoint: DELETE /account/apps
  - optional fields: apps_id
  - risk: irreversible external deletion; removes an app collection and every app tracked under it; approval required
- add_app:
  - endpoint: POST /account/apps/app
  - risk: external mutation; adds a tracked app to an existing collection; approval required
- remove_app:
  - endpoint: DELETE /account/apps/app
  - optional fields: store, ext_id, apps_id, user_id
  - risk: irreversible external deletion; stops tracking an app under a collection; approval required

## Security

- read risk: external AppFollow API read of account, app collection, review, rating, and ASO data
- write risk: external mutations: posts public review replies, edits review tags/notes/custom-status, replaces tracked ASO keyword sets, and adds/updates/removes account users, app collections, and tracked apps
- approval: required for every write action; each is flagged high-visibility (public review reply) or irreversible (collection/app/user removal) in writes.json's per-action risk field
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect appfollow
```

### Inspect as structured JSON

```bash
pm connectors inspect appfollow --json
```

## Agent Rules

- Run pm connectors inspect appfollow before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
