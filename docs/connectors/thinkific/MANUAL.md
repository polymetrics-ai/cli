# pm connectors inspect thinkific

```text
NAME
  pm connectors inspect thinkific - Thinkific connector manual

SYNOPSIS
  pm connectors inspect thinkific
  pm connectors inspect thinkific --json
  pm credentials add <name> --connector thinkific [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Thinkific courses, users, enrollments, products, orders, and site administration resources through the Thinkific Admin API.

ICON
  id: thinkific
  asset: icons/thinkific.svg
  source: official
  review_status: official_verified
  review_url: https://developers.thinkific.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  subdomain (required)
  api_key (secret) (required)

ETL STREAMS
  courses:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(integer), name(string), slug(string)
  users:
    primary key: id
    fields: company(string), created_at(string), email(string), first_name(string), full_name(string), id(integer), last_name(string), roles(array)
  enrollments:
    primary key: id
    fields: activated_at(string), completed(boolean), completed_at(string), course_id(integer), course_name(string), expired(boolean), expiry_date(string), id(integer), is_free_trial(boolean), percentage_completed(number), started_at(string), updated_at(string), user_email(string), user_id(integer), user_name(string)
  collections:
    primary key: id
    fields: created_at(string), default(boolean), description(string), id(integer), name(string), product_ids(array), slug(string)
  products:
    primary key: id
    fields: created_at(string), hidden(boolean), id(integer), name(string), price(number), private(boolean), productable_id(integer), productable_type(string), slug(string), status(string), subscription(boolean)
  orders:
    primary key: id
    fields: amount_cents(integer), amount_dollars(number), coupon_code(string), id(integer), product_id(integer), product_name(string), status(string), subscription(boolean), user_email(string), user_id(integer), user_name(string)
  coupons:
    primary key: id
    fields: code(string), created_at(string), id(integer), note(string), promotion_id(integer), quantity(integer), quantity_used(integer)
  promotions:
    primary key: id
    fields: amount(number), coupon_ids(array), description(string), discount_type(string), duration(number), expires_at(string), id(integer), name(string), starts_at(string)
  groups:
    primary key: id
    fields: created_at(string), id(integer), name(string), token(string)
  instructors:
    primary key: id
    fields: bio(string), created_at(string), email(string), first_name(string), id(integer), last_name(string), slug(string), title(string), user_id(integer)
  course_reviews:
    primary key: id
    fields: approved(boolean), course_id(integer), created_at(string), id(integer), rating(number), review_text(string), title(string), user_id(integer)
  custom_profile_field_definitions:
    primary key: id
    fields: field_type(string), id(integer), label(string), required(boolean)
  site_scripts:
    primary key: id
    fields: category(string), content(string), created_at(string), description(string), id(integer), load_method(string), location(string), name(string), src(string), updated_at(string)
  product_publish_requests:
    primary key: id
    fields: completed_at(string), created_at(string), id(integer), product_id(integer), requesting_user_id(integer), responding_user_id(integer), response_text(string), status(string), updated_at(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_user:
    endpoint: POST /api/public/v1/users
    required fields: email, first_name, last_name
    risk: creates a new Thinkific user account; low-risk additive external mutation, no approval required
  update_user:
    endpoint: PUT /api/public/v1/users/{{ record.id }}
    required fields: id
    risk: mutates an existing user's profile, email, password, or role assignment; a role change can grant or revoke site-admin/course-admin access
  delete_user:
    endpoint: DELETE /api/public/v1/users/{{ record.id }}
    required fields: id
    risk: permanently deletes a Thinkific user account and revokes their access to every enrolled course; destructive, approval required
  create_enrollment:
    endpoint: POST /api/public/v1/enrollments
    risk: grants a user access to a course; low-risk additive external mutation, no approval required
  update_enrollment:
    endpoint: PUT /api/public/v1/enrollments/{{ record.id }}
    required fields: id
    risk: changes an enrollment's activation or expiry date, which can extend or revoke a user's access window to a course
  create_coupon:
    endpoint: POST /api/public/v1/coupons
    required fields: code
    risk: creates a discount coupon code redeemable at checkout; low-risk additive external mutation, no approval required
  update_coupon:
    endpoint: PUT /api/public/v1/coupons/{{ record.id }}
    required fields: id, code
    risk: mutates an existing coupon's code, quantity, or usage counter; can change a customer-facing discount code that has already been shared
  delete_coupon:
    endpoint: DELETE /api/public/v1/coupons/{{ record.id }}
    required fields: id
    risk: permanently deletes a coupon; any customer relying on the code at checkout will see it rejected
  create_collection:
    endpoint: POST /api/public/v1/collections
    required fields: name, description, slug
    risk: creates a new course category (Collection); low-risk additive external mutation, no approval required
  update_collection:
    endpoint: PUT /api/public/v1/collections/{{ record.id }}
    required fields: id
    risk: renames or re-slugs an existing category, which changes its public landing-page URL
  delete_collection:
    endpoint: DELETE /api/public/v1/collections/{{ record.id }}
    required fields: id
    risk: permanently deletes a course category; any public page linking to it will 404
  add_products_to_collection:
    endpoint: POST /api/public/v1/collection_memberships/{{ record.collection_id }}
    required fields: collection_id, product_ids
    risk: adds one or more products (courses/bundles) to a public category, changing what appears on that category's landing page
  remove_products_from_collection:
    endpoint: DELETE /api/public/v1/collection_memberships/{{ record.collection_id }}
    required fields: collection_id, product_ids
    risk: removes one or more products from a public category, which can hide previously-listed courses from that category's landing page
  create_group:
    endpoint: POST /api/public/v1/groups
    required fields: name
    risk: creates a new Group (used for bulk enrollment/organization management); low-risk additive external mutation, no approval required
  delete_group:
    endpoint: DELETE /api/public/v1/groups/{{ record.id }}
    required fields: id
    risk: permanently deletes a Group; members lose any group-scoped access/reporting association
  add_user_to_groups:
    endpoint: POST /api/public/v1/group_users
    required fields: group_names, user_id
    risk: adds a user to one or more existing Groups by name; low-risk additive external mutation, no approval required
  create_instructor:
    endpoint: POST /api/public/v1/instructors
    required fields: first_name, last_name, slug
    risk: creates a new public Instructor profile; low-risk additive external mutation, no approval required
  update_instructor:
    endpoint: PUT /api/public/v1/instructors/{{ record.id }}
    required fields: id, first_name, last_name, slug
    risk: mutates a public Instructor profile's name, bio, or slug, changing what's shown on every course page that credits them
  delete_instructor:
    endpoint: DELETE /api/public/v1/instructors/{{ record.id }}
    required fields: id
    risk: permanently deletes an Instructor profile; any course crediting them loses that attribution
  create_promotion:
    endpoint: POST /api/public/v1/promotions
    required fields: name, discount_type, amount
    risk: creates a discount promotion applied automatically at checkout for the targeted products; low-risk additive external mutation, no approval required
  update_promotion:
    endpoint: PUT /api/public/v1/promotions/{{ record.id }}
    required fields: id, name, discount_type, amount
    risk: mutates an active discount promotion's amount, type, or eligible products, directly changing checkout pricing
  delete_promotion:
    endpoint: DELETE /api/public/v1/promotions/{{ record.id }}
    required fields: id
    risk: permanently deletes an active discount promotion; checkout pricing reverts to full price immediately
  create_site_script:
    endpoint: POST /api/public/v1/site_scripts
    required fields: name, description, page_scopes, category
    risk: injects arbitrary third-party HTML/JavaScript into every scoped page of the public site; high-risk external mutation (site-wide script injection), approval required
  update_site_script:
    endpoint: PUT /api/public/v1/site_scripts/{{ record.id }}
    required fields: id, name, description, page_scopes, category
    risk: changes the injected third-party HTML/JavaScript payload site-wide; high-risk external mutation, approval required
  delete_site_script:
    endpoint: DELETE /api/public/v1/site_scripts/{{ record.id }}
    required fields: id
    risk: removes an injected site script from every scoped page immediately
  create_course_review:
    endpoint: POST /api/public/v1/course_reviews
    required fields: approved, rating, review_text, title, user_id
    risk: creates a course review that, once approved, is publicly visible on the course landing page; low-risk additive external mutation, no approval required
  approve_product_publish_request:
    endpoint: POST /api/public/v1/product_publish_requests/{{ record.id }}/approve
    required fields: id, user_id
    risk: approves a pending course-publish request, making the course publicly visible/purchasable; approval required
  deny_product_publish_request:
    endpoint: POST /api/public/v1/product_publish_requests/{{ record.id }}/deny
    required fields: id, user_id
    risk: denies a pending course-publish request, blocking the course from going live

SECURITY
  read risk: external Thinkific Admin API read of course, user, enrollment, order, and site-administration data
  write risk: external Thinkific Admin API mutations covering user/enrollment/coupon/promotion/group/instructor/collection/course-review lifecycle and site-script injection; site_scripts writes inject arbitrary HTML/JS site-wide and are marked destructive/confirm-gated, delete_user is destructive/confirm-gated
  approval: required for site_scripts create/update and delete_user (confirm: destructive); other writes are low-risk additive/idempotent mutations, no approval required
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect thinkific

  # Inspect as structured JSON
  pm connectors inspect thinkific --json

AGENT WORKFLOW
  - Run pm connectors inspect thinkific before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
