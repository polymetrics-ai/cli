# Overview

Reads and writes Thinkific Admin API v1 courses, bundles, categories, coupons, enrollments, orders,
groups, instructors, products, promotions, reviews, users, and site scripts.

Readable streams: `courses`, `chapters`, `lessons`, `enrollments`, `bundle`, `bundle_courses`,
`bundle_enrollments`, `chapter`, `chapter_contents`, `collections`, `collection`,
`collection_products`, `coupons`, `coupon`, `content`, `course`, `course_chapters`,
`custom_profile_field_definitions`, `enrollment`, `groups`, `group`, `group_analysts`,
`instructors`, `instructor`, `orders`, `order`, `product_publish_requests`,
`product_publish_request`, `products`, `product`, `related_products`, `promotions`, `promotion`,
`promotion_by_coupon`, `course_reviews`, `course_review`, `users`, `user`, `user_authentication`,
`site_scripts`, `site_script`.

Write actions: `create_bundle_enrollment`, `update_bundle_enrollments`, `create_collection`,
`update_collection`, `delete_collection`, `add_product_to_collection`,
`delete_product_from_collection`, `update_coupon`, `delete_coupon`, `create_enrollment`,
`update_enrollment`, `create_external_order`, `refund_external_order_transaction`,
`purchase_external_order_transaction`, `create_group`, `delete_group`, `assign_group_analysts`,
`remove_group_analyst`, `add_groups_to_analyst`, `remove_group_from_analyst`, `create_instructor`,
`update_instructor`, `delete_instructor`, `create_group_users`, `approve_product_publish_request`,
`deny_product_publish_request`, `create_promotion`, `update_promotion`, `delete_promotion`,
`create_user`, `update_user`, `delete_user`, `create_site_script`, `update_site_script`,
`delete_site_script`.

Service API documentation: https://developers.thinkific.com/api/api-documentation/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Thinkific API key, sent as the X-Auth-API-Key header. Never
  logged.
- `base_url` (optional, string); default `https://api.thinkific.com/api/public/v1`; format `uri`;
  Thinkific public API base URL override for tests or proxies.
- `bundle_id` (optional, string); Bundle ID for bundle detail, courses, enrollments, and bundle
  enrollment writes.
- `chapter_id` (optional, string); Chapter ID for chapter detail and contents streams.
- `collection_id` (optional, string); Category/collection ID for category streams and writes.
- `content_id` (optional, string); Content ID for content detail streams.
- `coupon_code` (optional, string); Coupon code for the promotion_by_coupon stream.
- `coupon_id` (optional, string); Coupon ID for coupon detail/update/delete actions.
- `course_id` (optional, string); Course ID for course detail, course chapters, reviews, and
  course-scoped filters.
- `course_review_id` (optional, string); Course review ID for review detail streams.
- `enrollment_id` (optional, string); Enrollment ID for enrollment detail/update actions.
- `group_id` (optional, string); Group ID for group detail, analysts, and group membership actions.
- `instructor_id` (optional, string); Instructor ID for instructor detail/update/delete actions.
- `order_id` (optional, string); Order ID for order detail streams.
- `product_id` (optional, string); Product ID for product detail, related products, and promotion
  coupon lookup.
- `product_publish_request_id` (optional, string); Product publish request ID for detail and
  approve/deny actions.
- `promotion_id` (optional, string); Promotion ID for coupon filters and promotion
  detail/update/delete actions.
- `provider` (optional, string); default `thinkific`; Authentication provider key for the
  user_authentication stream.
- `site_script_id` (optional, string); Site script ID for site script detail/update/delete actions.
- `subdomain` (required, string); Thinkific account subdomain, sent as the X-Auth-Subdomain header.
- `user_id` (optional, string); User ID for user detail/authentication streams and user/group
  actions.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.thinkific.com/api/public/v1`,
`provider=thinkific`.

Authentication behavior:

- API key authentication in `X-Auth-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/courses` with query `limit`=`1`; `page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 100.

Pagination by stream: none: `bundle`, `chapter`, `collection`, `coupon`, `content`, `course`,
`enrollment`, `group`, `instructor`, `order`, `product_publish_request`, `product`, `promotion`,
`promotion_by_coupon`, `course_review`, `user`, `user_authentication`, `site_script`; page_number:
`courses`, `chapters`, `lessons`, `enrollments`, `bundle_courses`, `bundle_enrollments`,
`chapter_contents`, `collections`, `collection_products`, `coupons`, `course_chapters`,
`custom_profile_field_definitions`, `groups`, `group_analysts`, `instructors`, `orders`,
`product_publish_requests`, `products`, `related_products`, `promotions`, `course_reviews`, `users`,
`site_scripts`.

- `courses`: GET `/courses` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `chapters`: GET `/chapters` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `lessons`: GET `/lessons` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `enrollments`: GET `/enrollments` - records path `items`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `bundle`: GET `/bundles/{{ config.bundle_id }}` - single-object response; emits passthrough
  records.
- `bundle_courses`: GET `/bundles/{{ config.bundle_id }}/courses` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `bundle_enrollments`: GET `/bundles/{{ config.bundle_id }}/enrollments` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `chapter`: GET `/chapters/{{ config.chapter_id }}` - single-object response; emits passthrough
  records.
- `chapter_contents`: GET `/chapters/{{ config.chapter_id }}/contents` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `collections`: GET `/collections` - records path `items`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `collection`: GET `/collections/{{ config.collection_id }}` - single-object response; emits
  passthrough records.
- `collection_products`: GET `/collections/{{ config.collection_id }}/products` - records path
  `items`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; emits passthrough records.
- `coupons`: GET `/coupons` - records path `items`; query `promotion_id`=`{{ config.promotion_id
  }}`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size
  100; emits passthrough records.
- `coupon`: GET `/coupons/{{ config.coupon_id }}` - single-object response; emits passthrough
  records.
- `content`: GET `/contents/{{ config.content_id }}` - single-object response; emits passthrough
  records.
- `course`: GET `/courses/{{ config.course_id }}` - single-object response; emits passthrough
  records.
- `course_chapters`: GET `/courses/{{ config.course_id }}/chapters` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `custom_profile_field_definitions`: GET `/custom_profile_field_definitions` - records path
  `items`; page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page
  size 100; emits passthrough records.
- `enrollment`: GET `/enrollments/{{ config.enrollment_id }}` - single-object response; emits
  passthrough records.
- `groups`: GET `/groups` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `group`: GET `/groups/{{ config.group_id }}` - single-object response; emits passthrough records.
- `group_analysts`: GET `/groups/{{ config.group_id }}/analysts` - records path `items`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits
  passthrough records.
- `instructors`: GET `/instructors` - records path `items`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `instructor`: GET `/instructors/{{ config.instructor_id }}` - single-object response; emits
  passthrough records.
- `orders`: GET `/orders` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `order`: GET `/orders/{{ config.order_id }}` - single-object response; emits passthrough records.
- `product_publish_requests`: GET `/product_publish_requests` - records path `items`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100; emits
  passthrough records.
- `product_publish_request`: GET `/product_publish_requests/{{ config.product_publish_request_id }}`
  - single-object response; emits passthrough records.
- `products`: GET `/products` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `product`: GET `/products/{{ config.product_id }}` - single-object response; emits passthrough
  records.
- `related_products`: GET `/products/{{ config.product_id }}/related` - records path `items`;
  page-number pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 100;
  emits passthrough records.
- `promotions`: GET `/promotions` - records path `items`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `promotion`: GET `/promotions/{{ config.promotion_id }}` - single-object response; emits
  passthrough records.
- `promotion_by_coupon`: GET `/promotions/by_coupon` - single-object response; query
  `coupon_code`=`{{ config.coupon_code }}`; `product_id`=`{{ config.product_id }}`; emits
  passthrough records.
- `course_reviews`: GET `/course_reviews` - records path `items`; query `course_id`=`{{
  config.course_id }}`; page-number pagination; page parameter `page`; size parameter `limit`;
  starts at 1; page size 100; emits passthrough records.
- `course_review`: GET `/course_reviews/{{ config.course_review_id }}` - single-object response;
  emits passthrough records.
- `users`: GET `/users` - records path `items`; page-number pagination; page parameter `page`; size
  parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `user`: GET `/users/{{ config.user_id }}` - single-object response; emits passthrough records.
- `user_authentication`: GET `/users/{{ config.user_id }}/authentications/{{ config.provider }}` -
  single-object response; emits passthrough records.
- `site_scripts`: GET `/site_scripts` - records path `items`; page-number pagination; page parameter
  `page`; size parameter `limit`; starts at 1; page size 100; emits passthrough records.
- `site_script`: GET `/site_scripts/{{ config.site_script_id }}` - single-object response; emits
  passthrough records.

## Write actions & risks

Overall write risk: creates, updates, approves, denies, and deletes Thinkific enrollments,
categories, coupons, external orders, groups, instructors, promotions, users, and site scripts;
destructive deletes require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_bundle_enrollment`: POST `/bundles/{{ record.bundle_id }}/enrollments` - kind `create`;
  body type `json`; path fields `bundle_id`; required record fields `bundle_id`; accepted fields
  `bundle_id`; risk: creates enrollments in a bundle of courses.
- `update_bundle_enrollments`: PUT `/bundles/{{ record.bundle_id }}/enrollments` - kind `update`;
  body type `json`; path fields `bundle_id`; required record fields `bundle_id`; accepted fields
  `bundle_id`; risk: updates enrollments in a bundle.
- `create_collection`: POST `/collections` - kind `create`; body type `json`; risk: creates a
  category/collection.
- `update_collection`: PUT `/collections/{{ record.collection_id }}` - kind `update`; body type
  `json`; path fields `collection_id`; required record fields `collection_id`; accepted fields
  `collection_id`; risk: updates a category/collection.
- `delete_collection`: DELETE `/collections/{{ record.collection_id }}` - kind `delete`; body type
  `none`; path fields `collection_id`; required record fields `collection_id`; accepted fields
  `collection_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: deletes a category/collection.
- `add_product_to_collection`: POST `/collection_memberships/{{ record.collection_id }}` - kind
  `create`; body type `json`; path fields `collection_id`; required record fields `collection_id`;
  accepted fields `collection_id`; risk: adds products to a category/collection.
- `delete_product_from_collection`: DELETE `/collection_memberships/{{ record.collection_id }}` -
  kind `delete`; body type `json`; path fields `collection_id`; required record fields
  `collection_id`; accepted fields `collection_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: removes products from a category/collection.
- `update_coupon`: PUT `/coupons/{{ record.coupon_id }}` - kind `update`; body type `json`; path
  fields `coupon_id`; required record fields `coupon_id`; accepted fields `coupon_id`; risk: updates
  a coupon.
- `delete_coupon`: DELETE `/coupons/{{ record.coupon_id }}` - kind `delete`; body type `none`; path
  fields `coupon_id`; required record fields `coupon_id`; accepted fields `coupon_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: deletes a coupon.
- `create_enrollment`: POST `/enrollments` - kind `create`; body type `json`; risk: creates a course
  enrollment.
- `update_enrollment`: PUT `/enrollments/{{ record.enrollment_id }}` - kind `update`; body type
  `json`; path fields `enrollment_id`; required record fields `enrollment_id`; accepted fields
  `enrollment_id`; risk: updates a course enrollment.
- `create_external_order`: POST `/external_orders` - kind `create`; body type `json`; risk: creates
  an external order record.
- `refund_external_order_transaction`: POST `/external_orders/{{ record.order_id
  }}/transactions/refund` - kind `custom`; body type `json`; path fields `order_id`; required record
  fields `order_id`; accepted fields `order_id`; risk: records a refund transaction for an external
  order.
- `purchase_external_order_transaction`: POST `/external_orders/{{ record.order_id
  }}/transactions/purchase` - kind `custom`; body type `json`; path fields `order_id`; required
  record fields `order_id`; accepted fields `order_id`; risk: records a purchase transaction for an
  external order.
- `create_group`: POST `/groups` - kind `create`; body type `json`; risk: creates a group.
- `delete_group`: DELETE `/groups/{{ record.group_id }}` - kind `delete`; body type `none`; path
  fields `group_id`; required record fields `group_id`; accepted fields `group_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a group.
- `assign_group_analysts`: POST `/groups/{{ record.group_id }}/analysts` - kind `create`; body type
  `json`; path fields `group_id`; required record fields `group_id`; accepted fields `group_id`;
  risk: assigns analysts to a group.
- `remove_group_analyst`: DELETE `/groups/{{ record.group_id }}/analysts/{{ record.user_id }}` -
  kind `delete`; body type `none`; path fields `group_id`, `user_id`; required record fields
  `group_id`, `user_id`; accepted fields `group_id`, `user_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: removes an analyst from a group.
- `add_groups_to_analyst`: POST `/group_analysts/{{ record.user_id }}/groups` - kind `create`; body
  type `json`; path fields `user_id`; required record fields `user_id`; accepted fields `user_id`;
  risk: adds groups to an analyst.
- `remove_group_from_analyst`: DELETE `/group_analysts/{{ record.user_id }}/groups/{{
  record.group_id }}` - kind `delete`; body type `none`; path fields `user_id`, `group_id`; required
  record fields `user_id`, `group_id`; accepted fields `group_id`, `user_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: removes a group from an
  analyst.
- `create_instructor`: POST `/instructors` - kind `create`; body type `json`; risk: creates an
  instructor.
- `update_instructor`: PUT `/instructors/{{ record.instructor_id }}` - kind `update`; body type
  `json`; path fields `instructor_id`; required record fields `instructor_id`; accepted fields
  `instructor_id`; risk: updates an instructor.
- `delete_instructor`: DELETE `/instructors/{{ record.instructor_id }}` - kind `delete`; body type
  `none`; path fields `instructor_id`; required record fields `instructor_id`; accepted fields
  `instructor_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: deletes an instructor.
- `create_group_users`: POST `/group_users` - kind `create`; body type `json`; risk: adds users to
  existing groups.
- `approve_product_publish_request`: POST `/product_publish_requests/{{
  record.product_publish_request_id }}/approve` - kind `custom`; body type `json`; path fields
  `product_publish_request_id`; required record fields `product_publish_request_id`; accepted fields
  `product_publish_request_id`; risk: approves a product publish request.
- `deny_product_publish_request`: POST `/product_publish_requests/{{
  record.product_publish_request_id }}/deny` - kind `custom`; body type `json`; path fields
  `product_publish_request_id`; required record fields `product_publish_request_id`; accepted fields
  `product_publish_request_id`; risk: denies a product publish request.
- `create_promotion`: POST `/promotions` - kind `create`; body type `json`; risk: creates a
  promotion.
- `update_promotion`: PUT `/promotions/{{ record.promotion_id }}` - kind `update`; body type `json`;
  path fields `promotion_id`; required record fields `promotion_id`; accepted fields `promotion_id`;
  risk: updates a promotion.
- `delete_promotion`: DELETE `/promotions/{{ record.promotion_id }}` - kind `delete`; body type
  `none`; path fields `promotion_id`; required record fields `promotion_id`; accepted fields
  `promotion_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: deletes a promotion.
- `create_user`: POST `/users` - kind `create`; body type `json`; risk: creates a user.
- `update_user`: PUT `/users/{{ record.user_id }}` - kind `update`; body type `json`; path fields
  `user_id`; required record fields `user_id`; accepted fields `user_id`; risk: updates a user.
- `delete_user`: DELETE `/users/{{ record.user_id }}` - kind `delete`; body type `none`; path fields
  `user_id`; required record fields `user_id`; accepted fields `user_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes a user.
- `create_site_script`: POST `/site_scripts` - kind `create`; body type `json`; risk: creates a site
  script.
- `update_site_script`: PUT `/site_scripts/{{ record.site_script_id }}` - kind `update`; body type
  `json`; path fields `site_script_id`; required record fields `site_script_id`; accepted fields
  `site_script_id`; risk: updates a site script.
- `delete_site_script`: DELETE `/site_scripts/{{ record.site_script_id }}` - kind `delete`; body
  type `none`; path fields `site_script_id`; required record fields `site_script_id`; accepted
  fields `site_script_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a site script.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 41 stream-backed endpoint group(s), 35 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
