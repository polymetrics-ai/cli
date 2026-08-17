# pm connectors inspect thinkific-courses

```text
NAME
  pm connectors inspect thinkific-courses - Thinkific Courses connector manual

SYNOPSIS
  pm connectors inspect thinkific-courses
  pm connectors inspect thinkific-courses --json
  pm credentials add <name> --connector thinkific-courses [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Thinkific Admin API v1 courses, bundles, categories, coupons, enrollments, orders, groups, instructors, products, promotions, reviews, users, and site scripts.

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
  bundle_id
  chapter_id
  collection_id
  content_id
  coupon_code
  coupon_id
  course_id
  course_review_id
  enrollment_id
  group_id
  instructor_id
  order_id
  product_id
  product_publish_request_id
  promotion_id
  provider
  site_script_id
  subdomain
  user_id
  api_key (secret)

ETL STREAMS
  courses:
    primary key: id
    fields: created_at(), description(), id(), name(), slug(), updated_at()
  chapters:
    primary key: id
    fields: course_id(), created_at(), id(), name(), position(), updated_at()
  lessons:
    primary key: id
    fields: chapter_id(), course_id(), created_at(), id(), name(), position(), updated_at()
  enrollments:
    primary key: id
    fields: activated_at(), completed_at(), course_id(), id(), percentage_completed(), updated_at(), user_id()
  bundle:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  bundle_courses:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  bundle_enrollments:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  chapter:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  chapter_contents:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  collections:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  collection:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  collection_products:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  coupons:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  coupon:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  content:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  course:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  course_chapters:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  custom_profile_field_definitions:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  enrollment:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  groups:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  group:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  group_analysts:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  instructors:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  instructor:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  orders:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  order:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  product_publish_requests:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  product_publish_request:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  products:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  product:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  related_products:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  promotions:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  promotion:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  promotion_by_coupon:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  course_reviews:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  course_review:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  users:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  user:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  user_authentication:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  site_scripts:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()
  site_script:
    primary key: id
    fields: course_id(), created_at(), description(), email(), id(), items(), meta(), name(), percentage_completed(), product_id(), slug(), title(), updated_at(), user_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_bundle_enrollment:
    endpoint: POST /bundles/{{ record.bundle_id }}/enrollments
    required fields: bundle_id
    risk: creates enrollments in a bundle of courses
  update_bundle_enrollments:
    endpoint: PUT /bundles/{{ record.bundle_id }}/enrollments
    required fields: bundle_id
    risk: updates enrollments in a bundle
  create_collection:
    endpoint: POST /collections
    risk: creates a category/collection
  update_collection:
    endpoint: PUT /collections/{{ record.collection_id }}
    required fields: collection_id
    risk: updates a category/collection
  delete_collection:
    endpoint: DELETE /collections/{{ record.collection_id }}
    required fields: collection_id
    risk: deletes a category/collection
  add_product_to_collection:
    endpoint: POST /collection_memberships/{{ record.collection_id }}
    required fields: collection_id
    risk: adds products to a category/collection
  delete_product_from_collection:
    endpoint: DELETE /collection_memberships/{{ record.collection_id }}
    required fields: collection_id
    risk: removes products from a category/collection
  update_coupon:
    endpoint: PUT /coupons/{{ record.coupon_id }}
    required fields: coupon_id
    risk: updates a coupon
  delete_coupon:
    endpoint: DELETE /coupons/{{ record.coupon_id }}
    required fields: coupon_id
    risk: deletes a coupon
  create_enrollment:
    endpoint: POST /enrollments
    risk: creates a course enrollment
  update_enrollment:
    endpoint: PUT /enrollments/{{ record.enrollment_id }}
    required fields: enrollment_id
    risk: updates a course enrollment
  create_external_order:
    endpoint: POST /external_orders
    risk: creates an external order record
  refund_external_order_transaction:
    endpoint: POST /external_orders/{{ record.order_id }}/transactions/refund
    required fields: order_id
    risk: records a refund transaction for an external order
  purchase_external_order_transaction:
    endpoint: POST /external_orders/{{ record.order_id }}/transactions/purchase
    required fields: order_id
    risk: records a purchase transaction for an external order
  create_group:
    endpoint: POST /groups
    risk: creates a group
  delete_group:
    endpoint: DELETE /groups/{{ record.group_id }}
    required fields: group_id
    risk: deletes a group
  assign_group_analysts:
    endpoint: POST /groups/{{ record.group_id }}/analysts
    required fields: group_id
    risk: assigns analysts to a group
  remove_group_analyst:
    endpoint: DELETE /groups/{{ record.group_id }}/analysts/{{ record.user_id }}
    required fields: group_id, user_id
    risk: removes an analyst from a group
  add_groups_to_analyst:
    endpoint: POST /group_analysts/{{ record.user_id }}/groups
    required fields: user_id
    risk: adds groups to an analyst
  remove_group_from_analyst:
    endpoint: DELETE /group_analysts/{{ record.user_id }}/groups/{{ record.group_id }}
    required fields: user_id, group_id
    risk: removes a group from an analyst
  create_instructor:
    endpoint: POST /instructors
    risk: creates an instructor
  update_instructor:
    endpoint: PUT /instructors/{{ record.instructor_id }}
    required fields: instructor_id
    risk: updates an instructor
  delete_instructor:
    endpoint: DELETE /instructors/{{ record.instructor_id }}
    required fields: instructor_id
    risk: deletes an instructor
  create_group_users:
    endpoint: POST /group_users
    risk: adds users to existing groups
  approve_product_publish_request:
    endpoint: POST /product_publish_requests/{{ record.product_publish_request_id }}/approve
    required fields: product_publish_request_id
    risk: approves a product publish request
  deny_product_publish_request:
    endpoint: POST /product_publish_requests/{{ record.product_publish_request_id }}/deny
    required fields: product_publish_request_id
    risk: denies a product publish request
  create_promotion:
    endpoint: POST /promotions
    risk: creates a promotion
  update_promotion:
    endpoint: PUT /promotions/{{ record.promotion_id }}
    required fields: promotion_id
    risk: updates a promotion
  delete_promotion:
    endpoint: DELETE /promotions/{{ record.promotion_id }}
    required fields: promotion_id
    risk: deletes a promotion
  create_user:
    endpoint: POST /users
    risk: creates a user
  update_user:
    endpoint: PUT /users/{{ record.user_id }}
    required fields: user_id
    risk: updates a user
  delete_user:
    endpoint: DELETE /users/{{ record.user_id }}
    required fields: user_id
    risk: deletes a user
  create_site_script:
    endpoint: POST /site_scripts
    risk: creates a site script
  update_site_script:
    endpoint: PUT /site_scripts/{{ record.site_script_id }}
    required fields: site_script_id
    risk: updates a site script
  delete_site_script:
    endpoint: DELETE /site_scripts/{{ record.site_script_id }}
    required fields: site_script_id
    risk: deletes a site script

SECURITY
  read risk: external Thinkific Admin API read of course catalog, enrollment, commerce, user, group, promotion, review, and site-script data
  write risk: creates, updates, approves, denies, and deletes Thinkific enrollments, categories, coupons, external orders, groups, instructors, promotions, users, and site scripts; destructive deletes require approval
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect thinkific-courses

  # Inspect as structured JSON
  pm connectors inspect thinkific-courses --json

AGENT WORKFLOW
  - Run pm connectors inspect thinkific-courses before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
