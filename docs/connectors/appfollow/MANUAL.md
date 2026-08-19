# pm connectors inspect appfollow

```text
NAME
  pm connectors inspect appfollow - Appfollow connector manual

SYNOPSIS
  pm connectors inspect appfollow
  pm connectors inspect appfollow --json
  pm credentials add <name> --connector appfollow [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads AppFollow account users, app collections, app lists, reviews, review summaries, ratings/ratings history, ASO keywords, rankings, and version/what's-new metadata through the AppFollow REST API v2 (config-list-driven fan-out per app/collection); writes review replies/tags/notes, ASO keyword edits, and account user/app/collection management actions.

ICON
  id: appfollow
  asset: icons/appfollow.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://appfollow.docs.apiary.io/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  app_collection_ids
  base_url
  ext_ids
  report_country
  report_from
  report_store
  report_to
  api_secret (secret) (required)

ETL STREAMS
  users:
    primary key: id
    fields: email(string), id(integer), name(string), role(string), status(string), updated(string)
  app_collections:
    primary key: id
    fields: count_apps(integer), countries(string), created(string), id(integer), languages(string), title(string), title_normalized(string)
  app_lists:
    primary key: app_id
    fields: app_collection_id(string), app_id(integer), count_reviews(integer), count_whatsnew(integer), created(string), ext_id(string), is_favorite(integer), store(string), watch_url(string)
  ratings:
    primary key: ext_id, date, country
    fields: country(string), date(string), ext_id(string), rating(number), stars1(integer), stars2(integer), stars3(integer), stars4(integer), stars5(integer), stars_total(integer), store(string), version(string)
  reviews:
    primary key: id
    fields: app_id(integer), app_version(string), author(string), content(string), created(string), date(string), dt(string), ext_id(string), id(integer), is_answer(boolean), locale(string), rating(integer), rating_prev(integer), review_id(string), store(string), time(string), title(string), updated(string), user_id(string), was_changed(boolean)
  reviews_summary:
    primary key: ext_id, date, country
    fields: avg_rating(number), country(string), date(string), ext_id(string), stars1(integer), stars2(integer), stars3(integer), stars4(integer), stars5(integer), total(integer), version(string)
  keywords:
    primary key: ext_id, country, device, date
    fields: country(string), date(string), device(string), ext_id(string), keyword(string), no_pos(boolean), page(integer), popularity(integer), pos(integer), store(string), total(integer)
  rankings:
    primary key: ext_id, country, device, genre_id, date
    fields: category(string), country(string), date(string), device(string), ext_id(string), genre_id(string), position(integer)
  versions:
    primary key: ext_id, version, country
    fields: country(string), ext_id(string), release_date(string), size(integer), version(string), whats_new(string)
  versions_whatsnew:
    primary key: ext_id, version, country
    fields: country(string), ext_id(string), last_modified(string), version(string), whats_new(string)
  ratings_history:
    primary key: ext_id, date, country, version
    fields: avg_rating(number), country(string), date(string), ext_id(string), period(string), stars(integer), stars1(integer), stars2(integer), stars3(integer), stars4(integer), stars5(integer), store(string), total(integer), version(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  reply_to_review:
    endpoint: POST /reviews/reply
    required fields: ext_id, review_id, answer_text
    risk: external mutation; posts a public reply to a live app-store review, cannot be unsent programmatically; approval required
  update_review_tags:
    endpoint: POST /reviews/tags
    required fields: ext_id, review_id, tags
    risk: external mutation; overwrites a review's tag set; approval required
  update_review_notes:
    endpoint: POST /reviews/notes
    required fields: ext_id, review_id, content
    risk: external mutation; overwrites a review's internal note; approval required
  edit_keywords:
    endpoint: POST /aso/keywords
    required fields: country, device, keywords
    risk: external mutation; replaces the tracked ASO keyword list for a country/device pair; approval required
  add_user:
    endpoint: POST /account/users
    required fields: name, role, email
    risk: external mutation; grants AppFollow account access to a new user; approval required
  update_user:
    endpoint: PATCH /account/users
    required fields: id, name, role, email
    risk: external mutation; changes an existing account user's role/access; approval required
  remove_user:
    endpoint: DELETE /account/users
    required fields: id
    optional fields: email
    risk: irreversible external mutation; revokes an AppFollow account user's access; approval required
  add_collection:
    endpoint: POST /account/apps
    required fields: title, countries
    risk: external mutation; creates a new billable app collection; approval required
  remove_collection:
    endpoint: DELETE /account/apps
    required fields: apps_id
    risk: irreversible external deletion; removes an app collection and every app tracked under it; approval required
  add_app:
    endpoint: POST /account/apps/app
    required fields: store, ext_id, apps_id, locale
    risk: external mutation; adds a tracked app to an existing collection; approval required
  remove_app:
    endpoint: DELETE /account/apps/app
    required fields: store, ext_id, apps_id
    optional fields: user_id
    risk: irreversible external deletion; stops tracking an app under a collection; approval required

SECURITY
  read risk: external AppFollow API read of account, app collection, review, rating, and ASO data
  write risk: external mutations: posts public review replies, edits review tags/notes/custom-status, replaces tracked ASO keyword sets, and adds/updates/removes account users, app collections, and tracked apps
  approval: required for every write action; each is flagged high-visibility (public review reply) or irreversible (collection/app/user removal) in writes.json's per-action risk field
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect appfollow

  # Inspect as structured JSON
  pm connectors inspect appfollow --json

AGENT WORKFLOW
  - Run pm connectors inspect appfollow before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
