# Overview

Reads and writes Thinkific courses, users, enrollments, products, orders, and site administration
resources through the Thinkific Admin API.

Readable streams: `courses`, `users`, `enrollments`, `collections`, `products`, `orders`, `coupons`,
`promotions`, `groups`, `instructors`, `course_reviews`, `custom_profile_field_definitions`,
`site_scripts`, `product_publish_requests`.

Write actions: `create_user`, `update_user`, `delete_user`, `create_enrollment`,
`update_enrollment`, `create_coupon`, `update_coupon`, `delete_coupon`, `create_collection`,
`update_collection`, `delete_collection`, `add_products_to_collection`,
`remove_products_from_collection`, `create_group`, `delete_group`, `add_user_to_groups`,
`create_instructor`, `update_instructor`, `delete_instructor`, `create_promotion`,
`update_promotion`, `delete_promotion`, `create_site_script`, `update_site_script`,
`delete_site_script`, `create_course_review`, `approve_product_publish_request`,
`deny_product_publish_request`.

Service API documentation: https://developers.thinkific.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Thinkific API key, sent as the X-Auth-API-Key header. Never
  logged.
- `base_url` (optional, string); default `https://api.thinkific.com`; format `uri`; Thinkific API
  base URL override for tests or proxies.
- `subdomain` (required, string); Thinkific site subdomain, sent as the X-Auth-Subdomain header
  (e.g. 'academy' for academy.thinkific.com).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.thinkific.com`.

Provide the secret fields listed above. Authentication is applied by the connector-specific
implementation for this service.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/public/v1/courses`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

- `courses`: GET `/api/public/v1/courses` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `users`: GET `/api/public/v1/users` - records path `items`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100.
- `enrollments`: GET `/api/public/v1/enrollments` - records path `items`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `collections`: GET `/api/public/v1/collections` - records path `items`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `products`: GET `/api/public/v1/products` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `orders`: GET `/api/public/v1/orders` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `coupons`: GET `/api/public/v1/coupons` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `promotions`: GET `/api/public/v1/promotions` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `groups`: GET `/api/public/v1/groups` - records path `items`; page-number pagination; page
  parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `instructors`: GET `/api/public/v1/instructors` - records path `items`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `course_reviews`: GET `/api/public/v1/course_reviews` - records path `items`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `custom_profile_field_definitions`: GET `/api/public/v1/custom_profile_field_definitions` -
  records path `items`; page-number pagination; page parameter `page`; size parameter `limit`;
  starts at 1; page size 100.
- `site_scripts`: GET `/api/public/v1/site_scripts` - records path `items`; page-number pagination;
  page parameter `page`; size parameter `limit`; starts at 1; page size 100.
- `product_publish_requests`: GET `/api/public/v1/product_publish_requests` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100.

## Write actions & risks

Overall write risk: external Thinkific Admin API mutations covering
user/enrollment/coupon/promotion/group/instructor/collection/course-review lifecycle and site-script
injection; site_scripts writes inject arbitrary HTML/JS site-wide and are marked
destructive/confirm-gated, delete_user is destructive/confirm-gated.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_user`: POST `/api/public/v1/users` - kind `create`; body type `json`; required record
  fields `email`, `first_name`, `last_name`; accepted fields `bio`, `company`, `email`,
  `external_id`, `first_name`, `headline`, `last_name`, `password`, `roles`, `send_welcome_email`;
  risk: creates a new Thinkific user account; low-risk additive external mutation, no approval
  required.
- `update_user`: PUT `/api/public/v1/users/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `bio`, `company`, `email`, `first_name`,
  `headline`, `id`, `last_name`, `password`, `roles`; risk: mutates an existing user's profile,
  email, password, or role assignment; a role change can grant or revoke site-admin/course-admin
  access.
- `delete_user`: DELETE `/api/public/v1/users/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: permanently deletes a Thinkific user
  account and revokes their access to every enrolled course; destructive, approval required.
- `create_enrollment`: POST `/api/public/v1/enrollments` - kind `create`; body type `json`; accepted
  fields `activated_at`, `course_id`, `expiry_date`, `user_id`; risk: grants a user access to a
  course; low-risk additive external mutation, no approval required.
- `update_enrollment`: PUT `/api/public/v1/enrollments/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `activated_at`,
  `expiry_date`, `id`; risk: changes an enrollment's activation or expiry date, which can extend or
  revoke a user's access window to a course.
- `create_coupon`: POST `/api/public/v1/coupons` - kind `create`; body type `json`; required record
  fields `code`; accepted fields `code`, `note`, `quantity`; risk: creates a discount coupon code
  redeemable at checkout; low-risk additive external mutation, no approval required.
- `update_coupon`: PUT `/api/public/v1/coupons/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `code`; accepted fields `code`, `id`, `note`,
  `quantity`, `quantity_used`; risk: mutates an existing coupon's code, quantity, or usage counter;
  can change a customer-facing discount code that has already been shared.
- `delete_coupon`: DELETE `/api/public/v1/coupons/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently deletes a coupon; any customer relying on
  the code at checkout will see it rejected.
- `create_collection`: POST `/api/public/v1/collections` - kind `create`; body type `json`; required
  record fields `name`, `description`, `slug`; accepted fields `description`, `name`, `slug`; risk:
  creates a new course category (Collection); low-risk additive external mutation, no approval
  required.
- `update_collection`: PUT `/api/public/v1/collections/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `description`, `id`,
  `name`, `slug`; risk: renames or re-slugs an existing category, which changes its public
  landing-page URL.
- `delete_collection`: DELETE `/api/public/v1/collections/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently deletes a course category; any public page
  linking to it will 404.
- `add_products_to_collection`: POST `/api/public/v1/collection_memberships/{{ record.collection_id
  }}` - kind `custom`; body type `json`; path fields `collection_id`; body fields `product_ids`;
  required record fields `collection_id`, `product_ids`; accepted fields `collection_id`,
  `product_ids`; risk: adds one or more products (courses/bundles) to a public category, changing
  what appears on that category's landing page.
- `remove_products_from_collection`: DELETE `/api/public/v1/collection_memberships/{{
  record.collection_id }}` - kind `delete`; body type `json`; path fields `collection_id`; body
  fields `product_ids`; required record fields `collection_id`, `product_ids`; accepted fields
  `collection_id`, `product_ids`; missing records treated as success for status `404`; risk: removes
  one or more products from a public category, which can hide previously-listed courses from that
  category's landing page.
- `create_group`: POST `/api/public/v1/groups` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `name`; risk: creates a new Group (used for bulk
  enrollment/organization management); low-risk additive external mutation, no approval required.
- `delete_group`: DELETE `/api/public/v1/groups/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; risk: permanently deletes a Group; members lose any group-scoped
  access/reporting association.
- `add_user_to_groups`: POST `/api/public/v1/group_users` - kind `custom`; body type `json`;
  required record fields `group_names`, `user_id`; accepted fields `group_names`, `user_id`; risk:
  adds a user to one or more existing Groups by name; low-risk additive external mutation, no
  approval required.
- `create_instructor`: POST `/api/public/v1/instructors` - kind `create`; body type `json`; required
  record fields `first_name`, `last_name`, `slug`; accepted fields `bio`, `email`, `first_name`,
  `last_name`, `slug`, `title`, `user_id`; risk: creates a new public Instructor profile; low-risk
  additive external mutation, no approval required.
- `update_instructor`: PUT `/api/public/v1/instructors/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `first_name`, `last_name`, `slug`; accepted
  fields `bio`, `email`, `first_name`, `id`, `last_name`, `slug`, `title`; risk: mutates a public
  Instructor profile's name, bio, or slug, changing what's shown on every course page that credits
  them.
- `delete_instructor`: DELETE `/api/public/v1/instructors/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently deletes an Instructor profile; any course
  crediting them loses that attribution.
- `create_promotion`: POST `/api/public/v1/promotions` - kind `create`; body type `json`; required
  record fields `name`, `discount_type`, `amount`; accepted fields `amount`, `coupon_ids`,
  `description`, `discount_type`, `duration`, `expires_at`, `name`, `product_ids`, `starts_at`;
  risk: creates a discount promotion applied automatically at checkout for the targeted products;
  low-risk additive external mutation, no approval required.
- `update_promotion`: PUT `/api/public/v1/promotions/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `name`, `discount_type`, `amount`; accepted
  fields `amount`, `coupon_ids`, `description`, `discount_type`, `duration`, `expires_at`, `id`,
  `name`, `product_ids`, `starts_at`; risk: mutates an active discount promotion's amount, type, or
  eligible products, directly changing checkout pricing.
- `delete_promotion`: DELETE `/api/public/v1/promotions/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: permanently deletes an active discount promotion;
  checkout pricing reverts to full price immediately.
- `create_site_script`: POST `/api/public/v1/site_scripts` - kind `create`; body type `json`;
  required record fields `name`, `description`, `page_scopes`, `category`; accepted fields
  `category`, `content`, `description`, `load_method`, `location`, `name`, `page_scopes`, `src`;
  confirmation `destructive`; risk: injects arbitrary third-party HTML/JavaScript into every scoped
  page of the public site; high-risk external mutation (site-wide script injection), approval
  required.
- `update_site_script`: PUT `/api/public/v1/site_scripts/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`, `name`, `description`, `page_scopes`,
  `category`; accepted fields `category`, `content`, `description`, `id`, `load_method`, `location`,
  `name`, `page_scopes`, `src`; confirmation `destructive`; risk: changes the injected third-party
  HTML/JavaScript payload site-wide; high-risk external mutation, approval required.
- `delete_site_script`: DELETE `/api/public/v1/site_scripts/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; risk: removes an injected site script from every scoped page
  immediately.
- `create_course_review`: POST `/api/public/v1/course_reviews` - kind `create`; body type `json`;
  required record fields `approved`, `rating`, `review_text`, `title`, `user_id`; accepted fields
  `approved`, `course_id`, `rating`, `review_text`, `title`, `user_id`; risk: creates a course
  review that, once approved, is publicly visible on the course landing page; low-risk additive
  external mutation, no approval required.
- `approve_product_publish_request`: POST `/api/public/v1/product_publish_requests/{{ record.id
  }}/approve` - kind `custom`; body type `json`; path fields `id`; required record fields `id`,
  `user_id`; accepted fields `id`, `notify_requester`, `response_text`, `user_id`; risk: approves a
  pending course-publish request, making the course publicly visible/purchasable; approval required.
- `deny_product_publish_request`: POST `/api/public/v1/product_publish_requests/{{ record.id
  }}/deny` - kind `custom`; body type `json`; path fields `id`; required record fields `id`,
  `user_id`; accepted fields `id`, `notify_requester`, `response_text`, `user_id`; risk: denies a
  pending course-publish request, blocking the course from going live.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 14 stream-backed endpoint group(s), 28 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=17, non_data_endpoint=1, out_of_scope=3, requires_elevated_scope=14.
