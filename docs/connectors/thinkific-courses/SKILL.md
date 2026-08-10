---
name: pm-thinkific-courses
description: Thinkific Courses connector knowledge and safe action guide.
---

# pm-thinkific-courses

## Purpose

Reads and writes Thinkific Admin API v1 courses, bundles, categories, coupons, enrollments, orders, groups, instructors, products, promotions, reviews, users, and site scripts.

## Icon

- id: thinkific
- asset: icons/thinkific.svg
- source: official
- review_status: official_verified
- review_url: https://developers.thinkific.com/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- bundle_id
- chapter_id
- collection_id
- content_id
- coupon_code
- coupon_id
- course_id
- course_review_id
- enrollment_id
- group_id
- instructor_id
- order_id
- product_id
- product_publish_request_id
- promotion_id
- provider
- site_script_id
- subdomain
- user_id
- api_key (secret)

## ETL Streams

- courses:
  - primary key: id
  - fields: created_at(string), description(string), id(integer), name(string), slug(string), updated_at(string)
- chapters:
  - primary key: id
  - fields: course_id(integer), created_at(string), id(integer), name(string), position(integer), updated_at(string)
- lessons:
  - primary key: id
  - fields: chapter_id(integer), course_id(integer), created_at(string), id(integer), name(string), position(integer), updated_at(string)
- enrollments:
  - primary key: id
  - fields: activated_at(string), completed_at(string), course_id(integer), id(integer), percentage_completed(number), updated_at(string), user_id(integer)
- bundle:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- bundle_courses:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- bundle_enrollments:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- chapter:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- chapter_contents:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- collections:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- collection:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- collection_products:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- coupons:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- coupon:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- content:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- course:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- course_chapters:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- custom_profile_field_definitions:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- enrollment:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- groups:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- group:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- group_analysts:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- instructors:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- instructor:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- orders:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- order:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- product_publish_requests:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- product_publish_request:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- products:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- product:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- related_products:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- promotions:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- promotion:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- promotion_by_coupon:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- course_reviews:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- course_review:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- users:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- user:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- user_authentication:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- site_scripts:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)
- site_script:
  - primary key: id
  - fields: course_id(string), created_at(string), description(string), email(string), id(string), items(array), meta(object), name(string), percentage_completed(number), product_id(string), slug(string), title(string), updated_at(string), user_id(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_bundle_enrollment:
  - endpoint: POST /bundles/{{ record.bundle_id }}/enrollments
  - required fields: bundle_id
  - risk: creates enrollments in a bundle of courses
- update_bundle_enrollments:
  - endpoint: PUT /bundles/{{ record.bundle_id }}/enrollments
  - required fields: bundle_id
  - risk: updates enrollments in a bundle
- create_collection:
  - endpoint: POST /collections
  - risk: creates a category/collection
- update_collection:
  - endpoint: PUT /collections/{{ record.collection_id }}
  - required fields: collection_id
  - risk: updates a category/collection
- delete_collection:
  - endpoint: DELETE /collections/{{ record.collection_id }}
  - required fields: collection_id
  - risk: deletes a category/collection
- add_product_to_collection:
  - endpoint: POST /collection_memberships/{{ record.collection_id }}
  - required fields: collection_id
  - risk: adds products to a category/collection
- delete_product_from_collection:
  - endpoint: DELETE /collection_memberships/{{ record.collection_id }}
  - required fields: collection_id
  - risk: removes products from a category/collection
- update_coupon:
  - endpoint: PUT /coupons/{{ record.coupon_id }}
  - required fields: coupon_id
  - risk: updates a coupon
- delete_coupon:
  - endpoint: DELETE /coupons/{{ record.coupon_id }}
  - required fields: coupon_id
  - risk: deletes a coupon
- create_enrollment:
  - endpoint: POST /enrollments
  - risk: creates a course enrollment
- update_enrollment:
  - endpoint: PUT /enrollments/{{ record.enrollment_id }}
  - required fields: enrollment_id
  - risk: updates a course enrollment
- create_external_order:
  - endpoint: POST /external_orders
  - risk: creates an external order record
- refund_external_order_transaction:
  - endpoint: POST /external_orders/{{ record.order_id }}/transactions/refund
  - required fields: order_id
  - risk: records a refund transaction for an external order
- purchase_external_order_transaction:
  - endpoint: POST /external_orders/{{ record.order_id }}/transactions/purchase
  - required fields: order_id
  - risk: records a purchase transaction for an external order
- create_group:
  - endpoint: POST /groups
  - risk: creates a group
- delete_group:
  - endpoint: DELETE /groups/{{ record.group_id }}
  - required fields: group_id
  - risk: deletes a group
- assign_group_analysts:
  - endpoint: POST /groups/{{ record.group_id }}/analysts
  - required fields: group_id
  - risk: assigns analysts to a group
- remove_group_analyst:
  - endpoint: DELETE /groups/{{ record.group_id }}/analysts/{{ record.user_id }}
  - required fields: group_id, user_id
  - risk: removes an analyst from a group
- add_groups_to_analyst:
  - endpoint: POST /group_analysts/{{ record.user_id }}/groups
  - required fields: user_id
  - risk: adds groups to an analyst
- remove_group_from_analyst:
  - endpoint: DELETE /group_analysts/{{ record.user_id }}/groups/{{ record.group_id }}
  - required fields: user_id, group_id
  - risk: removes a group from an analyst
- create_instructor:
  - endpoint: POST /instructors
  - risk: creates an instructor
- update_instructor:
  - endpoint: PUT /instructors/{{ record.instructor_id }}
  - required fields: instructor_id
  - risk: updates an instructor
- delete_instructor:
  - endpoint: DELETE /instructors/{{ record.instructor_id }}
  - required fields: instructor_id
  - risk: deletes an instructor
- create_group_users:
  - endpoint: POST /group_users
  - risk: adds users to existing groups
- approve_product_publish_request:
  - endpoint: POST /product_publish_requests/{{ record.product_publish_request_id }}/approve
  - required fields: product_publish_request_id
  - risk: approves a product publish request
- deny_product_publish_request:
  - endpoint: POST /product_publish_requests/{{ record.product_publish_request_id }}/deny
  - required fields: product_publish_request_id
  - risk: denies a product publish request
- create_promotion:
  - endpoint: POST /promotions
  - risk: creates a promotion
- update_promotion:
  - endpoint: PUT /promotions/{{ record.promotion_id }}
  - required fields: promotion_id
  - risk: updates a promotion
- delete_promotion:
  - endpoint: DELETE /promotions/{{ record.promotion_id }}
  - required fields: promotion_id
  - risk: deletes a promotion
- create_user:
  - endpoint: POST /users
  - risk: creates a user
- update_user:
  - endpoint: PUT /users/{{ record.user_id }}
  - required fields: user_id
  - risk: updates a user
- delete_user:
  - endpoint: DELETE /users/{{ record.user_id }}
  - required fields: user_id
  - risk: deletes a user
- create_site_script:
  - endpoint: POST /site_scripts
  - risk: creates a site script
- update_site_script:
  - endpoint: PUT /site_scripts/{{ record.site_script_id }}
  - required fields: site_script_id
  - risk: updates a site script
- delete_site_script:
  - endpoint: DELETE /site_scripts/{{ record.site_script_id }}
  - required fields: site_script_id
  - risk: deletes a site script

## Security

- read risk: external Thinkific Admin API read of course catalog, enrollment, commerce, user, group, promotion, review, and site-script data
- write risk: creates, updates, approves, denies, and deletes Thinkific enrollments, categories, coupons, external orders, groups, instructors, promotions, users, and site scripts; destructive deletes require approval
- approval: reverse ETL writes require plan preview and approval token
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Thinkific Courses's declared streams and reverse-ETL actions.
- Usage: pm thinkific-courses <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - add groups to analyst apply - Plan and execute the add groups to analyst reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_groups_to_analyst]; approval: requires plan, preview, approval, and execute; risk: adds groups to an analyst; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - add product to collection apply - Plan and execute the add product to collection reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_product_to_collection]; approval: requires plan, preview, approval, and execute; risk: adds products to a category/collection; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - api get api v2 webhooks - Documented GET /api/v2/webhooks* (not implemented) [intent=direct_read availability=not_implemented operation=thinkific-courses.get.api-v2-webhooks]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get oauth operation - Documented GET /oauth/* (not implemented) [intent=direct_read availability=not_implemented operation=thinkific-courses.get.oauth]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post coupons - Documented POST /coupons (not implemented) [intent=direct_write availability=not_implemented operation=thinkific-courses.post.coupons]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post coupons bulk-create - Documented POST /coupons/bulk_create (not implemented) [intent=direct_write availability=not_implemented operation=thinkific-courses.post.coupons-bulk-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post course-reviews - Documented POST /course_reviews (not implemented) [intent=direct_write availability=not_implemented operation=thinkific-courses.post.course-reviews]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - approve product publish request apply - Plan and execute the approve product publish request reverse-ETL action [intent=reverse_etl availability=not_implemented write=approve_product_publish_request]; approval: requires plan, preview, approval, and execute; risk: approves a product publish request; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - assign group analysts apply - Plan and execute the assign group analysts reverse-ETL action [intent=reverse_etl availability=not_implemented write=assign_group_analysts]; approval: requires plan, preview, approval, and execute; risk: assigns analysts to a group; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - bundle courses list - Run the bundle courses ETL stream [intent=etl availability=implemented stream=bundle_courses]
  - bundle enrollments list - Run the bundle enrollments ETL stream [intent=etl availability=implemented stream=bundle_enrollments]
  - bundle list - Run the bundle ETL stream [intent=etl availability=implemented stream=bundle]
  - chapter contents list - Run the chapter contents ETL stream [intent=etl availability=implemented stream=chapter_contents]
  - chapter list - Run the chapter ETL stream [intent=etl availability=implemented stream=chapter]
  - chapters list - Run the chapters ETL stream [intent=etl availability=implemented stream=chapters]; notes: discrepancy=present-in-surface-absent-from-artifact
  - collection list - Run the collection ETL stream [intent=etl availability=implemented stream=collection]
  - collection products list - Run the collection products ETL stream [intent=etl availability=implemented stream=collection_products]
  - collections list - Run the collections ETL stream [intent=etl availability=implemented stream=collections]
  - content list - Run the content ETL stream [intent=etl availability=implemented stream=content]
  - coupon list - Run the coupon ETL stream [intent=etl availability=implemented stream=coupon]
  - coupons list - Run the coupons ETL stream [intent=etl availability=implemented stream=coupons]
  - course chapters list - Run the course chapters ETL stream [intent=etl availability=implemented stream=course_chapters]
  - course list - Run the course ETL stream [intent=etl availability=implemented stream=course]
  - course review list - Run the course review ETL stream [intent=etl availability=implemented stream=course_review]
  - course reviews list - Run the course reviews ETL stream [intent=etl availability=implemented stream=course_reviews]
  - courses list - Run the courses ETL stream [intent=etl availability=implemented stream=courses]
  - create bundle enrollment apply - Plan and execute the create bundle enrollment reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_bundle_enrollment]; approval: requires plan, preview, approval, and execute; risk: creates enrollments in a bundle of courses; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - create collection apply - Plan and execute the create collection reverse-ETL action [intent=reverse_etl availability=implemented write=create_collection]; approval: requires plan, preview, approval, and execute; risk: creates a category/collection
  - create enrollment apply - Plan and execute the create enrollment reverse-ETL action [intent=reverse_etl availability=implemented write=create_enrollment]; approval: requires plan, preview, approval, and execute; risk: creates a course enrollment
  - create external order apply - Plan and execute the create external order reverse-ETL action [intent=reverse_etl availability=implemented write=create_external_order]; approval: requires plan, preview, approval, and execute; risk: creates an external order record
  - create group apply - Plan and execute the create group reverse-ETL action [intent=reverse_etl availability=implemented write=create_group]; approval: requires plan, preview, approval, and execute; risk: creates a group
  - create group users apply - Plan and execute the create group users reverse-ETL action [intent=reverse_etl availability=implemented write=create_group_users]; approval: requires plan, preview, approval, and execute; risk: adds users to existing groups
  - create instructor apply - Plan and execute the create instructor reverse-ETL action [intent=reverse_etl availability=implemented write=create_instructor]; approval: requires plan, preview, approval, and execute; risk: creates an instructor
  - create promotion apply - Plan and execute the create promotion reverse-ETL action [intent=reverse_etl availability=implemented write=create_promotion]; approval: requires plan, preview, approval, and execute; risk: creates a promotion
  - create site script apply - Plan and execute the create site script reverse-ETL action [intent=reverse_etl availability=implemented write=create_site_script]; approval: requires plan, preview, approval, and execute; risk: creates a site script
  - create user apply - Plan and execute the create user reverse-ETL action [intent=reverse_etl availability=implemented write=create_user]; approval: requires plan, preview, approval, and execute; risk: creates a user
  - custom profile field definitions list - Run the custom profile field definitions ETL stream [intent=etl availability=implemented stream=custom_profile_field_definitions]
  - delete collection apply - Plan and execute the delete collection reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_collection]; approval: requires plan, preview, approval, and execute; risk: deletes a category/collection; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete coupon apply - Plan and execute the delete coupon reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_coupon]; approval: requires plan, preview, approval, and execute; risk: deletes a coupon; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete group apply - Plan and execute the delete group reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_group]; approval: requires plan, preview, approval, and execute; risk: deletes a group; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete instructor apply - Plan and execute the delete instructor reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_instructor]; approval: requires plan, preview, approval, and execute; risk: deletes an instructor; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete product from collection apply - Plan and execute the delete product from collection reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_product_from_collection]; approval: requires plan, preview, approval, and execute; risk: removes products from a category/collection; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete promotion apply - Plan and execute the delete promotion reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_promotion]; approval: requires plan, preview, approval, and execute; risk: deletes a promotion; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete site script apply - Plan and execute the delete site script reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_site_script]; approval: requires plan, preview, approval, and execute; risk: deletes a site script; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - delete user apply - Plan and execute the delete user reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: deletes a user; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - deny product publish request apply - Plan and execute the deny product publish request reverse-ETL action [intent=reverse_etl availability=not_implemented write=deny_product_publish_request]; approval: requires plan, preview, approval, and execute; risk: denies a product publish request; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - enrollment list - Run the enrollment ETL stream [intent=etl availability=implemented stream=enrollment]
  - enrollments list - Run the enrollments ETL stream [intent=etl availability=implemented stream=enrollments]
  - group analysts list - Run the group analysts ETL stream [intent=etl availability=implemented stream=group_analysts]
  - group list - Run the group ETL stream [intent=etl availability=implemented stream=group]
  - groups list - Run the groups ETL stream [intent=etl availability=implemented stream=groups]
  - instructor list - Run the instructor ETL stream [intent=etl availability=implemented stream=instructor]
  - instructors list - Run the instructors ETL stream [intent=etl availability=implemented stream=instructors]
  - lessons list - Run the lessons ETL stream [intent=etl availability=implemented stream=lessons]; notes: discrepancy=present-in-surface-absent-from-artifact
  - order list - Run the order ETL stream [intent=etl availability=implemented stream=order]
  - orders list - Run the orders ETL stream [intent=etl availability=implemented stream=orders]
  - product list - Run the product ETL stream [intent=etl availability=implemented stream=product]
  - product publish request list - Run the product publish request ETL stream [intent=etl availability=implemented stream=product_publish_request]
  - product publish requests list - Run the product publish requests ETL stream [intent=etl availability=implemented stream=product_publish_requests]
  - products list - Run the products ETL stream [intent=etl availability=implemented stream=products]
  - promotion by coupon list - Run the promotion by coupon ETL stream [intent=etl availability=implemented stream=promotion_by_coupon]
  - promotion list - Run the promotion ETL stream [intent=etl availability=implemented stream=promotion]
  - promotions list - Run the promotions ETL stream [intent=etl availability=implemented stream=promotions]
  - purchase external order transaction apply - Plan and execute the purchase external order transaction reverse-ETL action [intent=reverse_etl availability=not_implemented write=purchase_external_order_transaction]; approval: requires plan, preview, approval, and execute; risk: records a purchase transaction for an external order; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - refund external order transaction apply - Plan and execute the refund external order transaction reverse-ETL action [intent=reverse_etl availability=not_implemented write=refund_external_order_transaction]; approval: requires plan, preview, approval, and execute; risk: records a refund transaction for an external order; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - related products list - Run the related products ETL stream [intent=etl availability=implemented stream=related_products]
  - remove group analyst apply - Plan and execute the remove group analyst reverse-ETL action [intent=reverse_etl availability=not_implemented write=remove_group_analyst]; approval: requires plan, preview, approval, and execute; risk: removes an analyst from a group; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - remove group from analyst apply - Plan and execute the remove group from analyst reverse-ETL action [intent=reverse_etl availability=not_implemented write=remove_group_from_analyst]; approval: requires plan, preview, approval, and execute; risk: removes a group from an analyst; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - site script list - Run the site script ETL stream [intent=etl availability=implemented stream=site_script]
  - site scripts list - Run the site scripts ETL stream [intent=etl availability=implemented stream=site_scripts]
  - update bundle enrollments apply - Plan and execute the update bundle enrollments reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_bundle_enrollments]; approval: requires plan, preview, approval, and execute; risk: updates enrollments in a bundle; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update collection apply - Plan and execute the update collection reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_collection]; approval: requires plan, preview, approval, and execute; risk: updates a category/collection; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update coupon apply - Plan and execute the update coupon reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_coupon]; approval: requires plan, preview, approval, and execute; risk: updates a coupon; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update enrollment apply - Plan and execute the update enrollment reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_enrollment]; approval: requires plan, preview, approval, and execute; risk: updates a course enrollment; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update instructor apply - Plan and execute the update instructor reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_instructor]; approval: requires plan, preview, approval, and execute; risk: updates an instructor; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update promotion apply - Plan and execute the update promotion reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_promotion]; approval: requires plan, preview, approval, and execute; risk: updates a promotion; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update site script apply - Plan and execute the update site script reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_site_script]; approval: requires plan, preview, approval, and execute; risk: updates a site script; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - update user apply - Plan and execute the update user reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: updates a user; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
  - user authentication list - Run the user authentication ETL stream [intent=etl availability=implemented stream=user_authentication]
  - user list - Run the user ETL stream [intent=etl availability=implemented stream=user]
  - users list - Run the users ETL stream [intent=etl availability=implemented stream=users]

## Commands

### Inspect as a manual

```bash
pm connectors inspect thinkific-courses
```

### Inspect as structured JSON

```bash
pm connectors inspect thinkific-courses --json
```

## Agent Rules

- Run pm connectors inspect thinkific-courses before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
