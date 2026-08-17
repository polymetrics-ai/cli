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
  subdomain
  api_key (secret)

ETL STREAMS
  courses:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), name(), slug()
  users:
    primary key: id
    fields: company(), created_at(), email(), first_name(), full_name(), id(), last_name(), roles()
  enrollments:
    primary key: id
    fields: activated_at(), completed(), completed_at(), course_id(), course_name(), expired(), expiry_date(), id(), is_free_trial(), percentage_completed(), started_at(), updated_at(), user_email(), user_id(), user_name()
  collections:
    primary key: id
    fields: created_at(), default(), description(), id(), name(), product_ids(), slug()
  products:
    primary key: id
    fields: created_at(), hidden(), id(), name(), price(), private(), productable_id(), productable_type(), slug(), status(), subscription()
  orders:
    primary key: id
    fields: amount_cents(), amount_dollars(), coupon_code(), id(), product_id(), product_name(), status(), subscription(), user_email(), user_id(), user_name()
  coupons:
    primary key: id
    fields: code(), created_at(), id(), note(), promotion_id(), quantity(), quantity_used()
  promotions:
    primary key: id
    fields: amount(), coupon_ids(), description(), discount_type(), duration(), expires_at(), id(), name(), starts_at()
  groups:
    primary key: id
    fields: created_at(), id(), name(), token()
  instructors:
    primary key: id
    fields: bio(), created_at(), email(), first_name(), id(), last_name(), slug(), title(), user_id()
  course_reviews:
    primary key: id
    fields: approved(), course_id(), created_at(), id(), rating(), review_text(), title(), user_id()
  custom_profile_field_definitions:
    primary key: id
    fields: field_type(), id(), label(), required()
  site_scripts:
    primary key: id
    fields: category(), content(), created_at(), description(), id(), load_method(), location(), name(), src(), updated_at()
  product_publish_requests:
    primary key: id
    fields: completed_at(), created_at(), id(), product_id(), requesting_user_id(), responding_user_id(), response_text(), status(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_user:
    endpoint: POST /api/public/v1/users
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
    risk: creates a discount coupon code redeemable at checkout; low-risk additive external mutation, no approval required
  update_coupon:
    endpoint: PUT /api/public/v1/coupons/{{ record.id }}
    required fields: id
    risk: mutates an existing coupon's code, quantity, or usage counter; can change a customer-facing discount code that has already been shared
  delete_coupon:
    endpoint: DELETE /api/public/v1/coupons/{{ record.id }}
    required fields: id
    risk: permanently deletes a coupon; any customer relying on the code at checkout will see it rejected
  create_collection:
    endpoint: POST /api/public/v1/collections
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
    required fields: collection_id
    optional fields: product_ids
    risk: adds one or more products (courses/bundles) to a public category, changing what appears on that category's landing page
  remove_products_from_collection:
    endpoint: DELETE /api/public/v1/collection_memberships/{{ record.collection_id }}
    required fields: collection_id
    optional fields: product_ids
    risk: removes one or more products from a public category, which can hide previously-listed courses from that category's landing page
  create_group:
    endpoint: POST /api/public/v1/groups
    risk: creates a new Group (used for bulk enrollment/organization management); low-risk additive external mutation, no approval required
  delete_group:
    endpoint: DELETE /api/public/v1/groups/{{ record.id }}
    required fields: id
    risk: permanently deletes a Group; members lose any group-scoped access/reporting association
  add_user_to_groups:
    endpoint: POST /api/public/v1/group_users
    risk: adds a user to one or more existing Groups by name; low-risk additive external mutation, no approval required
  create_instructor:
    endpoint: POST /api/public/v1/instructors
    risk: creates a new public Instructor profile; low-risk additive external mutation, no approval required
  update_instructor:
    endpoint: PUT /api/public/v1/instructors/{{ record.id }}
    required fields: id
    risk: mutates a public Instructor profile's name, bio, or slug, changing what's shown on every course page that credits them
  delete_instructor:
    endpoint: DELETE /api/public/v1/instructors/{{ record.id }}
    required fields: id
    risk: permanently deletes an Instructor profile; any course crediting them loses that attribution
  create_promotion:
    endpoint: POST /api/public/v1/promotions
    risk: creates a discount promotion applied automatically at checkout for the targeted products; low-risk additive external mutation, no approval required
  update_promotion:
    endpoint: PUT /api/public/v1/promotions/{{ record.id }}
    required fields: id
    risk: mutates an active discount promotion's amount, type, or eligible products, directly changing checkout pricing
  delete_promotion:
    endpoint: DELETE /api/public/v1/promotions/{{ record.id }}
    required fields: id
    risk: permanently deletes an active discount promotion; checkout pricing reverts to full price immediately
  create_site_script:
    endpoint: POST /api/public/v1/site_scripts
    risk: injects arbitrary third-party HTML/JavaScript into every scoped page of the public site; high-risk external mutation (site-wide script injection), approval required
  update_site_script:
    endpoint: PUT /api/public/v1/site_scripts/{{ record.id }}
    required fields: id
    risk: changes the injected third-party HTML/JavaScript payload site-wide; high-risk external mutation, approval required
  delete_site_script:
    endpoint: DELETE /api/public/v1/site_scripts/{{ record.id }}
    required fields: id
    risk: removes an injected site script from every scoped page immediately
  create_course_review:
    endpoint: POST /api/public/v1/course_reviews
    risk: creates a course review that, once approved, is publicly visible on the course landing page; low-risk additive external mutation, no approval required
  approve_product_publish_request:
    endpoint: POST /api/public/v1/product_publish_requests/{{ record.id }}/approve
    required fields: id
    risk: approves a pending course-publish request, making the course publicly visible/purchasable; approval required
  deny_product_publish_request:
    endpoint: POST /api/public/v1/product_publish_requests/{{ record.id }}/deny
    required fields: id
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
