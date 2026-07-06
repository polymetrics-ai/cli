# Overview

Reads Amazon Selling Partner API orders, inventory, finance, catalog, listings, fulfillment,
reports, feeds, seller, shipping, vendor, and supporting JSON resources via Login with Amazon (LWA)
authentication; exposes write actions for SP-API mutations that fit path/body JSON requests.

Readable streams: `orders`, `inventory_summaries`, `financial_event_groups`,
`list_content_document_asin_relations`, `search_content_documents`,
`search_content_publish_records`, `get_inbound`, `get_label_page_types`, `get_inbound_shipment`,
`list_inbound_shipments`, `list_inventory`, `get_outbound`, `list_outbounds`,
`get_replenishment_order`, `list_replenishment_orders`, `get_catalog_item`, `search_catalog_items`,
`get_catalog_item_catalog_2022_04_01_items_asin`, `search_catalog_items_catalog_2022_04_01_items`,
`get_vehicles`, `list_catalog_categories`, `get_browse_node_return_topics`,
`get_browse_node_return_trends`, `get_browse_node_review_topics`, `get_browse_node_review_trends`,
`get_item_browse_node`, `get_item_review_topics`, `get_item_review_trends`, `get_query`,
`get_queries`, `get_definitions_product_type`, `search_definitions_product_types`,
`get_scheduled_package`, `get_return`, `list_returns`, `retrieve_invoice`,
`retrieve_shipping_options`, `get_shipment`, `get_shipments`, `get_prep_instructions`,
`get_shipment_items`, `get_shipment_items_by_shipment_id`, `get_shipments_fba_inbound_v0_shipments`,
`get_item_eligibility_preview`, `get_feature_sku`, `get_feature_inventory`, `get_features`,
`get_fulfillment_order`, `list_all_fulfillment_orders`, `list_return_reason_codes`,
`get_package_tracking_details`, `get_shipment_details`, `get_feed`, `get_feeds`,
`list_transactions`, `get_payment_methods`, `list_account_balances`, `get_account`, `list_accounts`,
`get_transaction`, `list_account_transactions`, `get_transfer_preview`, `get_transfer_schedule`,
`list_transfer_schedules`, `list_financial_events_by_group_id`, `list_financial_events`,
`list_financial_events_by_order_id`, `list_inbound_plan_boxes`, `list_inbound_plan_items`,
`list_packing_group_boxes`, `list_packing_group_items`, `list_packing_options`,
`list_inbound_plan_pallets`, `list_placement_options`, `list_shipment_boxes`,
`get_shipment_content_update_preview`, `list_shipment_content_update_previews`,
`list_delivery_window_options`, `list_shipment_items`, `list_shipment_pallets`,
`get_self_ship_appointment_slots`,
`get_shipment_inbound_fba_2024_03_20_inbound_plans_inbound_plan_id_shipments_shipment_id`,
`list_transportation_options`, `get_inbound_plan`, `list_inbound_plans`,
`list_item_compliance_details`, `list_prep_details`, `get_listings_item`, `search_listings_items`,
`get_listings_restrictions`, `get_messaging_actions_for_order`,
`get_shipment_mfn_v0_shipments_shipment_id`, `get_destination`, `get_destinations`,
`get_subscription_by_id`, `get_subscription`, `get_order`, `search_orders`, `get_order_items`,
`get_order_orders_v0_orders_order_id`, `get_competitive_pricing`, `get_item_offers`,
`get_listing_offers`, `get_pricing`, `get_report`, `get_reports`, `get_report_schedule`,
`get_report_schedules`, `get_order_metrics`, `get_account_sellers_v1_account`,
`get_marketplace_participations`, `get_appointment_slots`, `get_appointmment_slots_by_job_id`,
`get_service_job_by_service_job_id`, `get_service_jobs`, `get_account_shipping_v1_account`,
`get_shipment_shipping_v1_shipments_shipment_id`, `get_tracking_information`, `get_access_points`,
`get_carrier_account_form_inputs`, `get_collection_form`, `get_shipment_documents`,
`get_additional_inputs`, `get_tracking`, `get_solicitation_actions_for_order`, `get_supply_source`,
`get_supply_sources`, `get_invoices_export`, `get_invoices_exports`, `get_invoice`, `get_invoices`,
`get_order_vendor_direct_fulfillment_orders_2021_12_28_purchase_orders_purchase_order_numbe`,
`get_orders`, `get_order_vendor_direct_fulfillment_orders_v1_purchase_orders_purchase_order_number`,
`get_orders_vendor_direct_fulfillment_orders_v1_purchase_orders`, `get_order_scenarios`,
`get_customer_invoice`, `get_customer_invoices`,
`get_customer_invoice_vendor_direct_fulfillment_shipping_v1_customer_invoices_purchase_order_number`,
`get_customer_invoices_vendor_direct_fulfillment_shipping_v1_customer_invoices`,
`get_transaction_status`,
`get_transaction_status_vendor_direct_fulfillment_transactions_v1_transactions_transaction_id`,
`get_purchase_order`, `get_purchase_orders`, `get_purchase_orders_status`,
`get_shipment_details_vendor_shipping_v1_shipments`,
`get_transaction_vendor_transactions_v1_transactions_transaction_id`.

Write actions: `record_action_feedback`, `delete_notifications`, `create_notification`,
`cancel_inbound`, `confirm_inbound`, `update_inbound`, `create_inbound`,
`update_inbound_shipment_transport_details`, `confirm_outbound`, `update_outbound`,
`create_outbound`, `confirm_replenishment_order`, `create_replenishment_order`, `cancel_query`,
`create_query`, `update_scheduled_packages`, `create_scheduled_package`,
`create_scheduled_package_bulk`, `update_package_status`, `update_package`, `create_packages`,
`batch_inventory`, `add_inventory`, `create_inventory_item`, `delivery_offers`,
`cancel_fulfillment_order`, `create_fulfillment_return`, `submit_fulfillment_order_status_update`,
`update_fulfillment_order`, `create_fulfillment_order`, `cancel_feed`, `create_feed`,
`initiate_payout`, `cancel_inbound_plan`, `update_inbound_plan_name`, `set_packing_information`,
`confirm_packing_option`, `generate_packing_options`, `confirm_placement_option`,
`generate_placement_options`, `confirm_delivery_window_options`, `generate_delivery_window_options`,
`update_shipment_name`, `cancel_self_ship_appointment`, `schedule_self_ship_appointment`,
`update_shipment_source_address`, `update_shipment_tracking_details`,
`confirm_transportation_options`, `generate_transportation_options`, `create_inbound_plan`,
`set_prep_details`, `cancel_shipment`, `create_shipment`, `delete_destination`,
`create_destination`, `delete_subscription_by_id`, `send_test_notification`, `create_subscription`,
`update_verification_status`, `confirm_shipment`, `update_shipment_status`, `cancel_report`,
`create_report`, `cancel_report_schedule`, `create_report_schedule`,
`set_appointment_fulfillment_data`, `assign_appointment_resources`,
`reschedule_appointment_for_service_job_by_service_job_id`,
`add_appointment_for_service_job_by_service_job_id`, `complete_service_job_by_service_job_id`,
`purchase_shipment`, `cancel_shipment_post_shipping_v1_shipments_shipment_id_cancel`,
`create_shipment_post_shipping_v1_shipments`, `unlink_carrier_account`, `link_carrier_account`,
`link_carrier_account_put_shipping_v2_carrier_accounts_carrier_id`, `create_claim`,
`generate_collection_form`, `submit_ndr_feedback`, `one_click_shipment`, and 18 more.

Service API documentation: https://developer-docs.amazon.com/sp-api/.

## Auth setup

Connection fields:

- `access_point_types` (optional, string).
- `account_id` (optional, string).
- `asin` (optional, string); ASIN used by item, catalog, pricing, eligibility, and service lookup
  streams.
- `base_amount` (optional, string).
- `base_url` (optional, string); default `https://sellingpartnerapi-na.amazon.com`; format `uri`;
  SP-API base URL. Set to https://sellingpartnerapi-eu.amazon.com or
  https://sellingpartnerapi-fe.amazon.com for EU/FE sellers.
- `browse_node_id` (optional, string).
- `carrier_id` (optional, string).
- `collection_form_id` (optional, string).
- `content_reference_key` (optional, string).
- `content_update_preview_id` (optional, string).
- `country_code` (optional, string).
- `created_after` (optional, string); RFC3339 lower bound used by streams with required createdAfter
  filters.
- `created_before` (optional, string); RFC3339 upper bound used by streams with required
  createdBefore filters.
- `destination_country_code` (optional, string).
- `destination_currency_code` (optional, string).
- `destination_id` (optional, string).
- `event_group_id` (optional, string).
- `export_id` (optional, string).
- `feature_name` (optional, string).
- `feed_id` (optional, string).
- `inbound_plan_id` (optional, string).
- `invoice_id` (optional, string).
- `item_condition` (optional, string).
- `item_type` (optional, string).
- `keywords` (optional, string).
- `lwa_app_id` (required, secret, string); Login with Amazon (LWA) application client ID, used as
  client_id in the refresh_token token exchange. Never logged.
- `lwa_client_secret` (required, secret, string); Login with Amazon (LWA) application client secret,
  used as client_secret in the refresh_token token exchange. Never logged.
- `lwa_token_url` (optional, string); default `https://api.amazon.com/auth/o2/token`; format `uri`.
- `marketplace_id` (optional, string); default `ATVPDKIKX0DER`; Amazon marketplace ID sent on
  Orders/FBA Inventory requests.
- `max_pages` (optional, string); default `0`.
- `metrics_granularity` (optional, string); Sales API granularity value for order metrics.
- `metrics_interval` (optional, string); Sales API interval value for order metrics.
- `mskus` (optional, string).
- `notification_type` (optional, string).
- `order_id` (optional, string); Amazon order identifier for order detail and sub-resource streams.
- `package_client_reference_id` (optional, string).
- `package_id` (optional, string).
- `package_number` (optional, string).
- `packing_group_id` (optional, string).
- `page_size` (optional, string); default `100`.
- `postal_code` (optional, string).
- `product_type` (optional, string).
- `program` (optional, string).
- `purchase_order_number` (optional, string).
- `query_id` (optional, string).
- `query_type` (optional, string).
- `rate_id` (optional, string).
- `refresh_token` (required, secret, string); Long-lived LWA refresh_token exchanged for a
  short-lived access_token on every Check/Read. Never logged.
- `replication_start_date` (required, string); format `date-time`; RFC3339 lower bound for
  time-windowed streams (orders, inventory_summaries, financial_event_groups) on a fresh sync with
  no incremental cursor yet, sent as LastUpdatedAfter/startDateTime/FinancialEventGroupStartedAfter
  respectively.
- `report_id` (optional, string).
- `report_schedule_id` (optional, string).
- `report_types` (optional, string).
- `request_token` (optional, string).
- `return_id` (optional, string).
- `seller_fulfillment_order_id` (optional, string).
- `seller_id` (optional, string).
- `seller_sku` (optional, string); Seller SKU used by SKU-specific streams.
- `service_job_id` (optional, string).
- `ship_to_country_code` (optional, string).
- `shipment_id` (optional, string).
- `sku` (optional, string).
- `sort_by` (optional, string).
- `source_country_code` (optional, string).
- `source_currency_code` (optional, string).
- `status` (optional, string).
- `store_id` (optional, string).
- `subscription_id` (optional, string).
- `supply_source_id` (optional, string).
- `tracking_id` (optional, string).
- `transaction_id` (optional, string).
- `transfer_schedule_id` (optional, string).
- `vehicle_type` (optional, string).

Secret fields are redacted in logs and write previews: `lwa_app_id`, `lwa_client_secret`,
`refresh_token`.

Default configuration values: `base_url=https://sellingpartnerapi-na.amazon.com`,
`lwa_token_url=https://api.amazon.com/auth/o2/token`, `marketplace_id=ATVPDKIKX0DER`, `max_pages=0`,
`page_size=100`.

Authentication behavior:

- Connector-specific authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/orders/v0/orders` with query `CreatedAfter`=`{{
config.replication_start_date }}`; `MarketplaceIds`=`{{ config.marketplace_id }}`;
`MaxResultsPerPage`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `orders`, `inventory_summaries`, `financial_event_groups`,
`list_content_document_asin_relations`, `search_content_documents`,
`search_content_publish_records`, `list_inbound_shipments`, `list_inventory`, `list_outbounds`,
`list_replenishment_orders`, `search_catalog_items`,
`search_catalog_items_catalog_2022_04_01_items`, `get_vehicles`, `get_queries`, `list_returns`,
`get_shipments`, `get_shipment_items`, `get_shipments_fba_inbound_v0_shipments`,
`list_all_fulfillment_orders`, `get_feeds`, `list_transactions`, `list_account_transactions`,
`list_transfer_schedules`, `list_financial_events_by_group_id`, `list_financial_events`,
`list_financial_events_by_order_id`, `list_inbound_plan_boxes`, `list_inbound_plan_items`,
`list_packing_group_boxes`, `list_packing_group_items`, `list_packing_options`,
`list_inbound_plan_pallets`, `list_placement_options`, `list_shipment_boxes`,
`list_shipment_content_update_previews`, `list_delivery_window_options`, `list_shipment_items`,
`list_shipment_pallets`, `get_self_ship_appointment_slots`, `list_transportation_options`,
`list_inbound_plans`, `search_listings_items`, `search_orders`, `get_order_items`, `get_reports`,
`get_service_jobs`, `get_supply_sources`, `get_invoices_exports`, `get_invoices`, `get_orders`,
`get_orders_vendor_direct_fulfillment_orders_v1_purchase_orders`, `get_customer_invoices`,
`get_customer_invoices_vendor_direct_fulfillment_shipping_v1_customer_invoices`,
`get_purchase_orders`, `get_purchase_orders_status`,
`get_shipment_details_vendor_shipping_v1_shipments`; none: `get_inbound`, `get_label_page_types`,
`get_inbound_shipment`, `get_outbound`, `get_replenishment_order`, `get_catalog_item`,
`get_catalog_item_catalog_2022_04_01_items_asin`, `list_catalog_categories`,
`get_browse_node_return_topics`, `get_browse_node_return_trends`, `get_browse_node_review_topics`,
`get_browse_node_review_trends`, `get_item_browse_node`, `get_item_review_topics`,
`get_item_review_trends`, `get_query`, `get_definitions_product_type`,
`search_definitions_product_types`, `get_scheduled_package`, `get_return`, `retrieve_invoice`,
`retrieve_shipping_options`, `get_shipment`, `get_prep_instructions`,
`get_shipment_items_by_shipment_id`, `get_item_eligibility_preview`, `get_feature_sku`,
`get_feature_inventory`, `get_features`, `get_fulfillment_order`, `list_return_reason_codes`,
`get_package_tracking_details`, `get_shipment_details`, `get_feed`, `get_payment_methods`,
`list_account_balances`, `get_account`, `list_accounts`, `get_transaction`, `get_transfer_preview`,
`get_transfer_schedule`, `get_shipment_content_update_preview`,
`get_shipment_inbound_fba_2024_03_20_inbound_plans_inbound_plan_id_shipments_shipment_id`,
`get_inbound_plan`, `list_item_compliance_details`, `list_prep_details`, `get_listings_item`,
`get_listings_restrictions`, `get_messaging_actions_for_order`,
`get_shipment_mfn_v0_shipments_shipment_id`, `get_destination`, `get_destinations`,
`get_subscription_by_id`, `get_subscription`, `get_order`, `get_order_orders_v0_orders_order_id`,
`get_competitive_pricing`, `get_item_offers`, `get_listing_offers`, `get_pricing`, and 31 more.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `orders`: GET `/orders/v0/orders` - records path `payload.Orders`; query `MarketplaceIds`=`{{
  config.marketplace_id }}`; `MaxResultsPerPage`=`{{ config.page_size }}`; cursor pagination; cursor
  parameter `NextToken`; next token from `payload.NextToken`; incremental cursor `LastUpdateDate`;
  sent as `LastUpdatedAfter`; formatted as `rfc3339`; initial lower bound from
  `replication_start_date`.
- `inventory_summaries`: GET `/fba/inventory/v1/summaries` - records path
  `payload.inventorySummaries`; query `details`=`true`; `granularityId`=`{{ config.marketplace_id
  }}`; `granularityType`=`Marketplace`; `marketplaceIds`=`{{ config.marketplace_id }}`; cursor
  pagination; cursor parameter `nextToken`; next token from `pagination.nextToken`; incremental
  cursor `lastUpdatedTime`; sent as `startDateTime`; formatted as `rfc3339`; initial lower bound
  from `replication_start_date`.
- `financial_event_groups`: GET `/finances/v0/financialEventGroups` - records path
  `payload.FinancialEventGroupList`; query `MaxResultsPerPage`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `NextToken`; next token from `payload.NextToken`; incremental cursor
  `FinancialEventGroupEnd`; sent as `FinancialEventGroupStartedAfter`; formatted as `rfc3339`;
  initial lower bound from `replication_start_date`.
- `list_content_document_asin_relations`: GET `/aplus/2020-11-01/contentDocuments/{{
  config.content_reference_key }}/asins` - records path `asinMetadataSet`; query `marketplaceId`=`{{
  config.marketplace_id }}`; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; computed output fields `content_reference_key`; emits passthrough records.
- `search_content_documents`: GET `/aplus/2020-11-01/contentDocuments` - records path
  `contentMetadataRecords`; query `marketplaceId`=`{{ config.marketplace_id }}`; cursor pagination;
  cursor parameter `pageToken`; next token from `nextPageToken`; emits passthrough records.
- `search_content_publish_records`: GET `/aplus/2020-11-01/contentPublishRecords` - records path
  `publishRecordList`; query `asin`=`{{ config.asin }}`; `marketplaceId`=`{{ config.marketplace_id
  }}`; cursor pagination; cursor parameter `pageToken`; next token from `nextPageToken`; emits
  passthrough records.
- `get_inbound`: GET `/awd/2024-05-09/inboundOrders/{{ config.order_id }}` - single-object response;
  records path `.`; computed output fields `order_id`; emits passthrough records.
- `get_label_page_types`: GET `/awd/2024-05-09/inboundShipments/{{ config.shipment_id
  }}/labelPageTypes` - records path `pageTypes`; computed output fields `shipment_id`; emits
  passthrough records.
- `get_inbound_shipment`: GET `/awd/2024-05-09/inboundShipments/{{ config.shipment_id }}` -
  single-object response; records path `.`; computed output fields `shipment_id`; emits passthrough
  records.
- `list_inbound_shipments`: GET `/awd/2024-05-09/inboundShipments` - records path `shipments`; query
  `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor parameter `nextToken`; next token
  from `nextToken`; emits passthrough records.
- `list_inventory`: GET `/awd/2024-05-09/inventory` - records path `inventory`; query
  `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor parameter `nextToken`; next token
  from `nextToken`; emits passthrough records.
- `get_outbound`: GET `/awd/2024-05-09/outboundOrders/{{ config.order_id }}` - single-object
  response; records path `.`; computed output fields `order_id`; emits passthrough records.
- `list_outbounds`: GET `/awd/2024-05-09/outboundOrders` - records path `outboundOrders`; query
  `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor parameter `nextToken`; next token
  from `nextToken`; emits passthrough records.
- `get_replenishment_order`: GET `/awd/2024-05-09/replenishmentOrders/{{ config.order_id }}` -
  single-object response; records path `.`; computed output fields `order_id`; emits passthrough
  records.
- `list_replenishment_orders`: GET `/awd/2024-05-09/replenishmentOrders` - records path `orders`;
  query `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor parameter `nextToken`; next
  token from `nextToken`; emits passthrough records.
- `get_catalog_item`: GET `/catalog/2020-12-01/items/{{ config.asin }}` - single-object response;
  records path `.`; query `marketplaceIds`=`{{ config.marketplace_id }}`; computed output fields
  `asin`; emits passthrough records.
- `search_catalog_items`: GET `/catalog/2020-12-01/items` - records path `items`; query
  `keywords`=`{{ config.keywords }}`; `marketplaceIds`=`{{ config.marketplace_id }}`; `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next token from
  `pagination.nextToken`; emits passthrough records.
- `get_catalog_item_catalog_2022_04_01_items_asin`: GET `/catalog/2022-04-01/items/{{ config.asin
  }}` - single-object response; records path `.`; query `marketplaceIds`=`{{ config.marketplace_id
  }}`; computed output fields `asin`; emits passthrough records.
- `search_catalog_items_catalog_2022_04_01_items`: GET `/catalog/2022-04-01/items` - records path
  `items`; query `marketplaceIds`=`{{ config.marketplace_id }}`; `pageSize`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `pageToken`; next token from `pagination.nextToken`;
  emits passthrough records.
- `get_vehicles`: GET `/catalog/2024-11-01/automotive/vehicles` - records path `vehicles`; query
  `marketplaceId`=`{{ config.marketplace_id }}`; `vehicleType`=`{{ config.vehicle_type }}`; cursor
  pagination; cursor parameter `pageToken`; next token from `pagination.nextToken`; emits
  passthrough records.
- `list_catalog_categories`: GET `/catalog/v0/categories` - records path `payload`; query
  `MarketplaceId`=`{{ config.marketplace_id }}`; emits passthrough records.
- `get_browse_node_return_topics`: GET `/customerFeedback/2024-06-01/browseNodes/{{
  config.browse_node_id }}/returns/topics` - records path `topics`; query `marketplaceId`=`{{
  config.marketplace_id }}`; computed output fields `browse_node_id`; emits passthrough records.
- `get_browse_node_return_trends`: GET `/customerFeedback/2024-06-01/browseNodes/{{
  config.browse_node_id }}/returns/trends` - records path `returnTrends`; query `marketplaceId`=`{{
  config.marketplace_id }}`; computed output fields `browse_node_id`; emits passthrough records.
- `get_browse_node_review_topics`: GET `/customerFeedback/2024-06-01/browseNodes/{{
  config.browse_node_id }}/reviews/topics` - records path `topics.positiveTopics`; query
  `marketplaceId`=`{{ config.marketplace_id }}`; `sortBy`=`{{ config.sort_by }}`; computed output
  fields `browse_node_id`; emits passthrough records.
- `get_browse_node_review_trends`: GET `/customerFeedback/2024-06-01/browseNodes/{{
  config.browse_node_id }}/reviews/trends` - records path `reviewTrends.positiveTopics`; query
  `marketplaceId`=`{{ config.marketplace_id }}`; computed output fields `browse_node_id`; emits
  passthrough records.
- `get_item_browse_node`: GET `/customerFeedback/2024-06-01/items/{{ config.asin }}/browseNode` -
  single-object response; records path `.`; query `marketplaceId`=`{{ config.marketplace_id }}`;
  computed output fields `asin`; emits passthrough records.
- `get_item_review_topics`: GET `/customerFeedback/2024-06-01/items/{{ config.asin
  }}/reviews/topics` - records path `topics.positiveTopics`; query `marketplaceId`=`{{
  config.marketplace_id }}`; `sortBy`=`{{ config.sort_by }}`; computed output fields `asin`; emits
  passthrough records.
- `get_item_review_trends`: GET `/customerFeedback/2024-06-01/items/{{ config.asin
  }}/reviews/trends` - records path `reviewTrends.positiveTopics`; query `marketplaceId`=`{{
  config.marketplace_id }}`; computed output fields `asin`; emits passthrough records.
- `get_query`: GET `/dataKiosk/2023-11-15/queries/{{ config.query_id }}` - single-object response;
  records path `.`; computed output fields `query_id`; emits passthrough records.
- `get_queries`: GET `/dataKiosk/2023-11-15/queries` - records path `queries`; query `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `paginationToken`; next token from
  `pagination.nextToken`; emits passthrough records.
- `get_definitions_product_type`: GET `/definitions/2020-09-01/productTypes/{{ config.product_type
  }}` - single-object response; records path `.`; query `marketplaceIds`=`{{ config.marketplace_id
  }}`; computed output fields `product_type`; emits passthrough records.
- `search_definitions_product_types`: GET `/definitions/2020-09-01/productTypes` - records path
  `productTypes`; query `marketplaceIds`=`{{ config.marketplace_id }}`; emits passthrough records.
- `get_scheduled_package`: GET `/easyShip/2022-03-23/package` - records path `packageItems`; query
  `amazonOrderId`=`{{ config.order_id }}`; `marketplaceId`=`{{ config.marketplace_id }}`; emits
  passthrough records.
- `get_return`: GET `/externalFulfillment/2024-09-11/returns/{{ config.return_id }}` - single-object
  response; records path `.`; computed output fields `return_id`; emits passthrough records.
- `list_returns`: GET `/externalFulfillment/2024-09-11/returns` - records path `returns`; query
  `maxResults`=`{{ config.page_size }}`; cursor pagination; cursor parameter `nextToken`; next token
  from `nextToken`; emits passthrough records.
- `retrieve_invoice`: GET `/externalFulfillment/2024-09-11/shipments/{{ config.shipment_id
  }}/invoice` - single-object response; records path `document`; computed output fields
  `shipment_id`; emits passthrough records.
- `retrieve_shipping_options`: GET `/externalFulfillment/2024-09-11/shipments/{{ config.shipment_id
  }}/shippingOptions` - records path `shippingOptions`; query `packageId`=`{{ config.package_id }}`;
  computed output fields `shipment_id`; emits passthrough records.
- `get_shipment`: GET `/externalFulfillment/2024-09-11/shipments/{{ config.shipment_id }}` -
  single-object response; records path `.`; computed output fields `shipment_id`; emits passthrough
  records.
- `get_shipments`: GET `/externalFulfillment/2024-09-11/shipments` - records path `shipments`; query
  `maxResults`=`{{ config.page_size }}`; `status`=`{{ config.status }}`; cursor pagination; cursor
  parameter `paginationToken`; next token from `pagination.nextToken`; emits passthrough records.
- `get_prep_instructions`: GET `/fba/inbound/v0/prepInstructions` - records path
  `payload.SKUPrepInstructionsList`; query `ShipToCountryCode`=`{{ config.ship_to_country_code }}`;
  emits passthrough records.
- `get_shipment_items`: GET `/fba/inbound/v0/shipmentItems` - records path `payload.ItemData`; query
  `MarketplaceId`=`{{ config.marketplace_id }}`; `QueryType`=`{{ config.query_type }}`; cursor
  pagination; cursor parameter `NextToken`; next token from `payload.NextToken`; emits passthrough
  records.
- `get_shipment_items_by_shipment_id`: GET `/fba/inbound/v0/shipments/{{ config.shipment_id
  }}/items` - records path `payload.ItemData`; computed output fields `shipment_id`; emits
  passthrough records.
- `get_shipments_fba_inbound_v0_shipments`: GET `/fba/inbound/v0/shipments` - records path
  `payload.ShipmentData`; query `MarketplaceId`=`{{ config.marketplace_id }}`; `QueryType`=`{{
  config.query_type }}`; cursor pagination; cursor parameter `NextToken`; next token from
  `payload.NextToken`; emits passthrough records.
- `get_item_eligibility_preview`: GET `/fba/inbound/v1/eligibility/itemPreview` - records path
  `payload.ineligibilityReasonList`; query `asin`=`{{ config.asin }}`; `program`=`{{ config.program
  }}`; emits passthrough records.
- `get_feature_sku`: GET `/fba/outbound/2020-07-01/features/inventory/{{ config.feature_name }}/{{
  config.seller_sku }}` - single-object response; records path `payload`; query `marketplaceId`=`{{
  config.marketplace_id }}`; computed output fields `feature_name`, `seller_sku`; emits passthrough
  records.
- `get_feature_inventory`: GET `/fba/outbound/2020-07-01/features/inventory/{{ config.feature_name
  }}` - single-object response; records path `payload`; query `marketplaceId`=`{{
  config.marketplace_id }}`; computed output fields `feature_name`; emits passthrough records.
- `get_features`: GET `/fba/outbound/2020-07-01/features` - records path `payload.features`; query
  `marketplaceId`=`{{ config.marketplace_id }}`; emits passthrough records.
- `get_fulfillment_order`: GET `/fba/outbound/2020-07-01/fulfillmentOrders/{{
  config.seller_fulfillment_order_id }}` - single-object response; records path `payload`; computed
  output fields `seller_fulfillment_order_id`; emits passthrough records.
- `list_all_fulfillment_orders`: GET `/fba/outbound/2020-07-01/fulfillmentOrders` - records path
  `payload.fulfillmentOrders`; cursor pagination; cursor parameter `nextToken`; next token from
  `payload.nextToken`; emits passthrough records.
- `list_return_reason_codes`: GET `/fba/outbound/2020-07-01/returnReasonCodes` - records path
  `payload.reasonCodeDetails`; query `sellerSku`=`{{ config.seller_sku }}`; emits passthrough
  records.
- `get_package_tracking_details`: GET `/fba/outbound/2020-07-01/tracking` - records path
  `payload.trackingEvents`; query `packageNumber`=`{{ config.package_number }}`; emits passthrough
  records.
- `get_shipment_details`: GET `/fba/outbound/brazil/v0/shipments/{{ config.shipment_id }}` -
  single-object response; records path `payload`; computed output fields `shipment_id`; emits
  passthrough records.
- `get_feed`: GET `/feeds/2021-06-30/feeds/{{ config.feed_id }}` - single-object response; records
  path `.`; computed output fields `feed_id`; emits passthrough records.
- `get_feeds`: GET `/feeds/2021-06-30/feeds` - records path `feeds`; query `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `nextToken`; next token from
  `nextToken`; emits passthrough records.
- `list_transactions`: GET `/finances/2024-06-19/transactions` - records path
  `payload.transactions`; cursor pagination; cursor parameter `nextToken`; next token from
  `payload.nextToken`; emits passthrough records.
- `get_payment_methods`: GET `/finances/transfers/2024-06-01/paymentMethods` - records path
  `paymentMethods`; query `marketplaceId`=`{{ config.marketplace_id }}`; emits passthrough records.
- `list_account_balances`: GET `/finances/transfers/wallet/2024-03-01/accounts/{{ config.account_id
  }}/balance` - records path `balances`; query `marketplaceId`=`{{ config.marketplace_id }}`;
  computed output fields `account_id`; emits passthrough records.
- `get_account`: GET `/finances/transfers/wallet/2024-03-01/accounts/{{ config.account_id }}` -
  single-object response; records path `.`; query `marketplaceId`=`{{ config.marketplace_id }}`;
  computed output fields `account_id`; emits passthrough records.
- `list_accounts`: GET `/finances/transfers/wallet/2024-03-01/accounts` - records path `accounts`;
  query `marketplaceId`=`{{ config.marketplace_id }}`; emits passthrough records.
- `get_transaction`: GET `/finances/transfers/wallet/2024-03-01/transactions/{{
  config.transaction_id }}` - single-object response; records path `transactionStatus`; query
  `marketplaceId`=`{{ config.marketplace_id }}`; computed output fields `transaction_id`; emits
  passthrough records.
- `list_account_transactions`: GET `/finances/transfers/wallet/2024-03-01/transactions` - records
  path `transactions`; query `accountId`=`{{ config.account_id }}`; `marketplaceId`=`{{
  config.marketplace_id }}`; cursor pagination; cursor parameter `nextPageToken`; next token from
  `nextPageToken`; emits passthrough records.
- `get_transfer_preview`: GET `/finances/transfers/wallet/2024-03-01/transferPreview` - records path
  `fees`; query `baseAmount`=`{{ config.base_amount }}`; `destinationCountryCode`=`{{
  config.destination_country_code }}`; `destinationCurrencyCode`=`{{
  config.destination_currency_code }}`; `marketplaceId`=`{{ config.marketplace_id }}`;
  `sourceCountryCode`=`{{ config.source_country_code }}`; `sourceCurrencyCode`=`{{
  config.source_currency_code }}`; emits passthrough records.
- `get_transfer_schedule`: GET `/finances/transfers/wallet/2024-03-01/transferSchedules/{{
  config.transfer_schedule_id }}` - single-object response; records path `.`; query
  `marketplaceId`=`{{ config.marketplace_id }}`; computed output fields `transfer_schedule_id`;
  emits passthrough records.
- `list_transfer_schedules`: GET `/finances/transfers/wallet/2024-03-01/transferSchedules` - records
  path `transferSchedules`; query `accountId`=`{{ config.account_id }}`; `marketplaceId`=`{{
  config.marketplace_id }}`; cursor pagination; cursor parameter `nextPageToken`; next token from
  `nextPageToken`; emits passthrough records.
- `list_financial_events_by_group_id`: GET `/finances/v0/financialEventGroups/{{
  config.event_group_id }}/financialEvents` - records path
  `payload.FinancialEvents.ShipmentEventList`; query `MaxResultsPerPage`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `NextToken`; next token from `payload.NextToken`; computed
  output fields `event_group_id`; emits passthrough records.
- `list_financial_events`: GET `/finances/v0/financialEvents` - records path
  `payload.FinancialEvents.ShipmentEventList`; query `MaxResultsPerPage`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `NextToken`; next token from `payload.NextToken`; emits
  passthrough records.
- `list_financial_events_by_order_id`: GET `/finances/v0/orders/{{ config.order_id
  }}/financialEvents` - records path `payload.FinancialEvents.ShipmentEventList`; query
  `MaxResultsPerPage`=`{{ config.page_size }}`; cursor pagination; cursor parameter `NextToken`;
  next token from `payload.NextToken`; computed output fields `order_id`; emits passthrough records.
- `list_inbound_plan_boxes`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/boxes` - records path `boxes`; query `pageSize`=`{{ config.page_size }}`; cursor pagination;
  cursor parameter `paginationToken`; next token from `pagination.nextToken`; computed output fields
  `inbound_plan_id`; emits passthrough records.
- `list_inbound_plan_items`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/items` - records path `items`; query `pageSize`=`{{ config.page_size }}`; cursor pagination;
  cursor parameter `paginationToken`; next token from `pagination.nextToken`; computed output fields
  `inbound_plan_id`; emits passthrough records.
- `list_packing_group_boxes`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/packingGroups/{{ config.packing_group_id }}/boxes` - records path `boxes`; query `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `paginationToken`; next token from
  `pagination.nextToken`; computed output fields `inbound_plan_id`, `packing_group_id`; emits
  passthrough records.
- `list_packing_group_items`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/packingGroups/{{ config.packing_group_id }}/items` - records path `items`; query `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `paginationToken`; next token from
  `pagination.nextToken`; computed output fields `inbound_plan_id`, `packing_group_id`; emits
  passthrough records.
- `list_packing_options`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/packingOptions` - records path `packingOptions`; query `pageSize`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `paginationToken`; next token from `pagination.nextToken`;
  computed output fields `inbound_plan_id`; emits passthrough records.
- `list_inbound_plan_pallets`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/pallets` - records path `pallets`; query `pageSize`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `paginationToken`; next token from `pagination.nextToken`; computed
  output fields `inbound_plan_id`; emits passthrough records.
- `list_placement_options`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/placementOptions` - records path `placementOptions`; query `pageSize`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `paginationToken`; next token from `pagination.nextToken`;
  computed output fields `inbound_plan_id`; emits passthrough records.
- `list_shipment_boxes`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/shipments/{{ config.shipment_id }}/boxes` - records path `boxes`; query `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `paginationToken`; next token from
  `pagination.nextToken`; computed output fields `inbound_plan_id`, `shipment_id`; emits passthrough
  records.
- `get_shipment_content_update_preview`: GET `/inbound/fba/2024-03-20/inboundPlans/{{
  config.inbound_plan_id }}/shipments/{{ config.shipment_id }}/contentUpdatePreviews/{{
  config.content_update_preview_id }}` - single-object response; records path `.`; computed output
  fields `content_update_preview_id`, `inbound_plan_id`, `shipment_id`; emits passthrough records.
- `list_shipment_content_update_previews`: GET `/inbound/fba/2024-03-20/inboundPlans/{{
  config.inbound_plan_id }}/shipments/{{ config.shipment_id }}/contentUpdatePreviews` - records path
  `contentUpdatePreviews`; query `pageSize`=`{{ config.page_size }}`; cursor pagination; cursor
  parameter `paginationToken`; next token from `pagination.nextToken`; computed output fields
  `inbound_plan_id`, `shipment_id`; emits passthrough records.
- `list_delivery_window_options`: GET `/inbound/fba/2024-03-20/inboundPlans/{{
  config.inbound_plan_id }}/shipments/{{ config.shipment_id }}/deliveryWindowOptions` - records path
  `deliveryWindowOptions`; query `pageSize`=`{{ config.page_size }}`; cursor pagination; cursor
  parameter `paginationToken`; next token from `pagination.nextToken`; computed output fields
  `inbound_plan_id`, `shipment_id`; emits passthrough records.
- `list_shipment_items`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/shipments/{{ config.shipment_id }}/items` - records path `items`; query `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `paginationToken`; next token from
  `pagination.nextToken`; computed output fields `inbound_plan_id`, `shipment_id`; emits passthrough
  records.
- `list_shipment_pallets`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/shipments/{{ config.shipment_id }}/pallets` - records path `pallets`; query `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `paginationToken`; next token from
  `pagination.nextToken`; computed output fields `inbound_plan_id`, `shipment_id`; emits passthrough
  records.
- `get_self_ship_appointment_slots`: GET `/inbound/fba/2024-03-20/inboundPlans/{{
  config.inbound_plan_id }}/shipments/{{ config.shipment_id }}/selfShipAppointmentSlots` - records
  path `selfShipAppointmentSlotsAvailability.slots`; query `pageSize`=`{{ config.page_size }}`;
  cursor pagination; cursor parameter `paginationToken`; next token from `pagination.nextToken`;
  computed output fields `inbound_plan_id`, `shipment_id`; emits passthrough records.
- `get_shipment_inbound_fba_2024_03_20_inbound_plans_inbound_plan_id_shipments_shipment_id`: GET
  `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id }}/shipments/{{ config.shipment_id
  }}` - single-object response; records path `.`; computed output fields `inbound_plan_id`,
  `shipment_id`; emits passthrough records.
- `list_transportation_options`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id
  }}/transportationOptions` - records path `transportationOptions`; query `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `paginationToken`; next token from
  `pagination.nextToken`; computed output fields `inbound_plan_id`; emits passthrough records.
- `get_inbound_plan`: GET `/inbound/fba/2024-03-20/inboundPlans/{{ config.inbound_plan_id }}` -
  single-object response; records path `.`; computed output fields `inbound_plan_id`; emits
  passthrough records.
- `list_inbound_plans`: GET `/inbound/fba/2024-03-20/inboundPlans` - records path `inboundPlans`;
  query `pageSize`=`{{ config.page_size }}`; cursor pagination; cursor parameter `paginationToken`;
  next token from `pagination.nextToken`; emits passthrough records.
- `list_item_compliance_details`: GET `/inbound/fba/2024-03-20/items/compliance` - records path
  `complianceDetails`; query `marketplaceId`=`{{ config.marketplace_id }}`; `mskus`=`{{ config.mskus
  }}`; emits passthrough records.
- `list_prep_details`: GET `/inbound/fba/2024-03-20/items/prepDetails` - records path
  `mskuPrepDetails`; query `marketplaceId`=`{{ config.marketplace_id }}`; `mskus`=`{{ config.mskus
  }}`; emits passthrough records.
- `get_listings_item`: GET `/listings/2021-08-01/items/{{ config.seller_id }}/{{ config.sku }}` -
  single-object response; records path `.`; query `marketplaceIds`=`{{ config.marketplace_id }}`;
  computed output fields `seller_id`, `sku`; emits passthrough records.
- `search_listings_items`: GET `/listings/2021-08-01/items/{{ config.seller_id }}` - records path
  `items`; query `marketplaceIds`=`{{ config.marketplace_id }}`; `pageSize`=`{{ config.page_size
  }}`; cursor pagination; cursor parameter `pageToken`; next token from `pagination.nextToken`;
  computed output fields `seller_id`; emits passthrough records.
- `get_listings_restrictions`: GET `/listings/2021-08-01/restrictions` - records path
  `restrictions`; query `asin`=`{{ config.asin }}`; `marketplaceIds`=`{{ config.marketplace_id }}`;
  `sellerId`=`{{ config.seller_id }}`; emits passthrough records.
- `get_messaging_actions_for_order`: GET `/messaging/v1/orders/{{ config.order_id }}` -
  single-object response; records path `.`; query `marketplaceIds`=`{{ config.marketplace_id }}`;
  computed output fields `order_id`; emits passthrough records.
- `get_shipment_mfn_v0_shipments_shipment_id`: GET `/mfn/v0/shipments/{{ config.shipment_id }}` -
  single-object response; records path `payload`; computed output fields `shipment_id`; emits
  passthrough records.
- `get_destination`: GET `/notifications/v1/destinations/{{ config.destination_id }}` -
  single-object response; records path `payload`; computed output fields `destination_id`; emits
  passthrough records.
- `get_destinations`: GET `/notifications/v1/destinations` - records path `payload`; emits
  passthrough records.
- `get_subscription_by_id`: GET `/notifications/v1/subscriptions/{{ config.notification_type }}/{{
  config.subscription_id }}` - single-object response; records path `payload`; computed output
  fields `notification_type`, `subscription_id`; emits passthrough records.
- `get_subscription`: GET `/notifications/v1/subscriptions/{{ config.notification_type }}` -
  single-object response; records path `payload`; computed output fields `notification_type`; emits
  passthrough records.
- `get_order`: GET `/orders/2026-01-01/orders/{{ config.order_id }}` - single-object response;
  records path `order`; computed output fields `order_id`; emits passthrough records.
- `search_orders`: GET `/orders/2026-01-01/orders` - records path `orders`; query
  `maxResultsPerPage`=`{{ config.page_size }}`; cursor pagination; cursor parameter
  `paginationToken`; next token from `pagination.nextToken`; emits passthrough records.
- `get_order_items`: GET `/orders/v0/orders/{{ config.order_id }}/orderItems` - records path
  `payload.OrderItems`; cursor pagination; cursor parameter `NextToken`; next token from
  `payload.NextToken`; computed output fields `order_id`; emits passthrough records.
- `get_order_orders_v0_orders_order_id`: GET `/orders/v0/orders/{{ config.order_id }}` -
  single-object response; records path `payload`; computed output fields `order_id`; emits
  passthrough records.
- `get_competitive_pricing`: GET `/products/pricing/v0/competitivePrice` - records path `payload`;
  query `ItemType`=`{{ config.item_type }}`; `MarketplaceId`=`{{ config.marketplace_id }}`; emits
  passthrough records.
- `get_item_offers`: GET `/products/pricing/v0/items/{{ config.asin }}/offers` - records path
  `payload.Offers`; query `ItemCondition`=`{{ config.item_condition }}`; `MarketplaceId`=`{{
  config.marketplace_id }}`; computed output fields `asin`; emits passthrough records.
- `get_listing_offers`: GET `/products/pricing/v0/listings/{{ config.seller_sku }}/offers` - records
  path `payload.Offers`; query `ItemCondition`=`{{ config.item_condition }}`; `MarketplaceId`=`{{
  config.marketplace_id }}`; computed output fields `seller_sku`; emits passthrough records.
- `get_pricing`: GET `/products/pricing/v0/price` - records path `payload`; query `ItemType`=`{{
  config.item_type }}`; `MarketplaceId`=`{{ config.marketplace_id }}`; emits passthrough records.
- `get_report`: GET `/reports/2021-06-30/reports/{{ config.report_id }}` - single-object response;
  records path `.`; computed output fields `report_id`; emits passthrough records.
- `get_reports`: GET `/reports/2021-06-30/reports` - records path `reports`; query `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `nextToken`; next token from
  `nextToken`; emits passthrough records.
- `get_report_schedule`: GET `/reports/2021-06-30/schedules/{{ config.report_schedule_id }}` -
  single-object response; records path `.`; computed output fields `report_schedule_id`; emits
  passthrough records.
- `get_report_schedules`: GET `/reports/2021-06-30/schedules` - records path `reportSchedules`;
  query `reportTypes`=`{{ config.report_types }}`; emits passthrough records.
- `get_order_metrics`: GET `/sales/v1/orderMetrics` - records path `payload`; query
  `granularity`=`{{ config.metrics_granularity }}`; `interval`=`{{ config.metrics_interval }}`;
  `marketplaceIds`=`{{ config.marketplace_id }}`; emits passthrough records.
- `get_account_sellers_v1_account`: GET `/sellers/v1/account` - records path
  `payload.marketplaceParticipationList`; emits passthrough records.
- `get_marketplace_participations`: GET `/sellers/v1/marketplaceParticipations` - records path
  `payload`; emits passthrough records.
- `get_appointment_slots`: GET `/service/v1/appointmentSlots` - records path
  `payload.appointmentSlots`; query `asin`=`{{ config.asin }}`; `marketplaceIds`=`{{
  config.marketplace_id }}`; `storeId`=`{{ config.store_id }}`; emits passthrough records.
- `get_appointmment_slots_by_job_id`: GET `/service/v1/serviceJobs/{{ config.service_job_id
  }}/appointmentSlots` - records path `payload.appointmentSlots`; query `marketplaceIds`=`{{
  config.marketplace_id }}`; computed output fields `service_job_id`; emits passthrough records.
- `get_service_job_by_service_job_id`: GET `/service/v1/serviceJobs/{{ config.service_job_id }}` -
  single-object response; records path `payload`; computed output fields `service_job_id`; emits
  passthrough records.
- `get_service_jobs`: GET `/service/v1/serviceJobs` - records path `payload.jobs`; query
  `marketplaceIds`=`{{ config.marketplace_id }}`; `pageSize`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `pageToken`; next token from `payload.nextPageToken`; emits
  passthrough records.
- `get_account_shipping_v1_account`: GET `/shipping/v1/account` - single-object response; records
  path `payload`; emits passthrough records.
- `get_shipment_shipping_v1_shipments_shipment_id`: GET `/shipping/v1/shipments/{{
  config.shipment_id }}` - single-object response; records path `payload`; computed output fields
  `shipment_id`; emits passthrough records.
- `get_tracking_information`: GET `/shipping/v1/tracking/{{ config.tracking_id }}` - single-object
  response; records path `payload`; computed output fields `tracking_id`; emits passthrough records.
- `get_access_points`: GET `/shipping/v2/accessPoints` - single-object response; records path
  `payload`; query `accessPointTypes`=`{{ config.access_point_types }}`; `countryCode`=`{{
  config.country_code }}`; `postalCode`=`{{ config.postal_code }}`; emits passthrough records.
- `get_carrier_account_form_inputs`: GET `/shipping/v2/carrierAccountFormInputs` - records path
  `linkableCarriersList`; emits passthrough records.
- `get_collection_form`: GET `/shipping/v2/collectionForms/{{ config.collection_form_id }}` -
  single-object response; records path `collectionsFormDocument`; computed output fields
  `collection_form_id`; emits passthrough records.
- `get_shipment_documents`: GET `/shipping/v2/shipments/{{ config.shipment_id }}/documents` -
  records path `payload.packageDocumentDetail.packageDocuments`; query
  `packageClientReferenceId`=`{{ config.package_client_reference_id }}`; computed output fields
  `shipment_id`; emits passthrough records.
- `get_additional_inputs`: GET `/shipping/v2/shipments/additionalInputs/schema` - single-object
  response; records path `payload`; query `rateId`=`{{ config.rate_id }}`; `requestToken`=`{{
  config.request_token }}`; emits passthrough records.
- `get_tracking`: GET `/shipping/v2/tracking` - records path
  `payload.summary.trackingDetailCodes.returns`; query `carrierId`=`{{ config.carrier_id }}`;
  `trackingId`=`{{ config.tracking_id }}`; emits passthrough records.
- `get_solicitation_actions_for_order`: GET `/solicitations/v1/orders/{{ config.order_id }}` -
  single-object response; records path `.`; query `marketplaceIds`=`{{ config.marketplace_id }}`;
  computed output fields `order_id`; emits passthrough records.
- `get_supply_source`: GET `/supplySources/2020-07-01/supplySources/{{ config.supply_source_id }}` -
  single-object response; records path `.`; computed output fields `supply_source_id`; emits
  passthrough records.
- `get_supply_sources`: GET `/supplySources/2020-07-01/supplySources` - records path
  `supplySources`; query `pageSize`=`{{ config.page_size }}`; cursor pagination; cursor parameter
  `nextPageToken`; next token from `nextPageToken`; emits passthrough records.
- `get_invoices_export`: GET `/tax/invoices/2024-06-19/exports/{{ config.export_id }}` -
  single-object response; records path `export`; computed output fields `export_id`; emits
  passthrough records.
- `get_invoices_exports`: GET `/tax/invoices/2024-06-19/exports` - records path `exports`; query
  `marketplaceId`=`{{ config.marketplace_id }}`; `pageSize`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `nextToken`; next token from `nextToken`; emits passthrough records.
- `get_invoice`: GET `/tax/invoices/2024-06-19/invoices/{{ config.invoice_id }}` - single-object
  response; records path `invoice`; query `marketplaceId`=`{{ config.marketplace_id }}`; computed
  output fields `invoice_id`; emits passthrough records.
- `get_invoices`: GET `/tax/invoices/2024-06-19/invoices` - records path `invoices`; query
  `marketplaceId`=`{{ config.marketplace_id }}`; `pageSize`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `nextToken`; next token from `nextToken`; emits passthrough records.
- `get_order_vendor_direct_fulfillment_orders_2021_12_28_purchase_orders_purchase_order_numbe`: GET
  `/vendor/directFulfillment/orders/2021-12-28/purchaseOrders/{{ config.purchase_order_number }}` -
  single-object response; records path `.`; computed output fields `purchase_order_number`; emits
  passthrough records.
- `get_orders`: GET `/vendor/directFulfillment/orders/2021-12-28/purchaseOrders` - records path
  `orders`; query `createdAfter`=`{{ config.created_after }}`; `createdBefore`=`{{
  config.created_before }}`; `limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter
  `nextToken`; next token from `pagination.nextToken`; emits passthrough records.
- `get_order_vendor_direct_fulfillment_orders_v1_purchase_orders_purchase_order_number`: GET
  `/vendor/directFulfillment/orders/v1/purchaseOrders/{{ config.purchase_order_number }}` -
  single-object response; records path `payload`; computed output fields `purchase_order_number`;
  emits passthrough records.
- `get_orders_vendor_direct_fulfillment_orders_v1_purchase_orders`: GET
  `/vendor/directFulfillment/orders/v1/purchaseOrders` - records path `payload.orders`; query
  `createdAfter`=`{{ config.created_after }}`; `createdBefore`=`{{ config.created_before }}`;
  `limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter `nextToken`; next token from
  `payload.pagination.nextToken`; emits passthrough records.
- `get_order_scenarios`: GET `/vendor/directFulfillment/sandbox/2021-10-28/transactions/{{
  config.transaction_id }}` - single-object response; records path `transactionStatus`; computed
  output fields `transaction_id`; emits passthrough records.
- `get_customer_invoice`: GET `/vendor/directFulfillment/shipping/2021-12-28/customerInvoices/{{
  config.purchase_order_number }}` - single-object response; records path `.`; computed output
  fields `purchase_order_number`; emits passthrough records.
- `get_customer_invoices`: GET `/vendor/directFulfillment/shipping/2021-12-28/customerInvoices` -
  records path `customerInvoices`; query `createdAfter`=`{{ config.created_after }}`;
  `createdBefore`=`{{ config.created_before }}`; `limit`=`{{ config.page_size }}`; cursor
  pagination; cursor parameter `nextToken`; next token from `pagination.nextToken`; emits
  passthrough records.
- `get_customer_invoice_vendor_direct_fulfillment_shipping_v1_customer_invoices_purchase_order_number`:
  GET `/vendor/directFulfillment/shipping/v1/customerInvoices/{{ config.purchase_order_number }}` -
  single-object response; records path `payload`; computed output fields `purchase_order_number`;
  emits passthrough records.
- `get_customer_invoices_vendor_direct_fulfillment_shipping_v1_customer_invoices`: GET
  `/vendor/directFulfillment/shipping/v1/customerInvoices` - records path
  `payload.customerInvoices`; query `createdAfter`=`{{ config.created_after }}`; `createdBefore`=`{{
  config.created_before }}`; `limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter
  `nextToken`; next token from `payload.pagination.nextToken`; emits passthrough records.
- `get_transaction_status`: GET `/vendor/directFulfillment/transactions/2021-12-28/transactions/{{
  config.transaction_id }}` - single-object response; records path `transactionStatus`; computed
  output fields `transaction_id`; emits passthrough records.
- `get_transaction_status_vendor_direct_fulfillment_transactions_v1_transactions_transaction_id`:
  GET `/vendor/directFulfillment/transactions/v1/transactions/{{ config.transaction_id }}` -
  single-object response; records path `payload`; computed output fields `transaction_id`; emits
  passthrough records.
- `get_purchase_order`: GET `/vendor/orders/v1/purchaseOrders/{{ config.purchase_order_number }}` -
  single-object response; records path `payload`; computed output fields `purchase_order_number`;
  emits passthrough records.
- `get_purchase_orders`: GET `/vendor/orders/v1/purchaseOrders` - records path `payload.orders`;
  query `limit`=`{{ config.page_size }}`; cursor pagination; cursor parameter `nextToken`; next
  token from `payload.pagination.nextToken`; emits passthrough records.
- `get_purchase_orders_status`: GET `/vendor/orders/v1/purchaseOrdersStatus` - records path
  `payload.ordersStatus`; query `limit`=`{{ config.page_size }}`; cursor pagination; cursor
  parameter `nextToken`; next token from `payload.pagination.nextToken`; emits passthrough records.
- `get_shipment_details_vendor_shipping_v1_shipments`: GET `/vendor/shipping/v1/shipments` - records
  path `payload.shipments`; query `limit`=`{{ config.page_size }}`; cursor pagination; cursor
  parameter `nextToken`; next token from `payload.pagination.nextToken`; emits passthrough records.
- `get_transaction_vendor_transactions_v1_transactions_transaction_id`: GET
  `/vendor/transactions/v1/transactions/{{ config.transaction_id }}` - single-object response;
  records path `payload`; computed output fields `transaction_id`; emits passthrough records.

## Write actions & risks

Overall write risk: external Amazon Selling Partner API mutations such as feed/report creation,
fulfillment, inventory, order, shipping, notification, and vendor workflow actions.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `record_action_feedback`: POST `/appIntegrations/2024-04-01/notifications/{{
  record.notification_id }}/feedback` - kind `custom`; body type `json`; path fields
  `notification_id`; required record fields `notification_id`; accepted fields `notification_id`;
  risk: POST /appIntegrations/2024-04-01/notifications/{notificationId}/feedback against the live
  SP-API account.
- `delete_notifications`: POST `/appIntegrations/2024-04-01/notifications/deletion` - kind `custom`;
  body type `json`; confirmation `destructive`; risk: POST
  /appIntegrations/2024-04-01/notifications/deletion against the live SP-API account.
- `create_notification`: POST `/appIntegrations/2024-04-01/notifications` - kind `create`; body type
  `json`; confirmation `destructive`; risk: POST /appIntegrations/2024-04-01/notifications against
  the live SP-API account.
- `cancel_inbound`: POST `/awd/2024-05-09/inboundOrders/{{ record.order_id }}/cancellation` - kind
  `custom`; body type `none`; path fields `order_id`; required record fields `order_id`; accepted
  fields `order_id`; confirmation `destructive`; risk: POST
  /awd/2024-05-09/inboundOrders/{orderId}/cancellation against the live SP-API account.
- `confirm_inbound`: POST `/awd/2024-05-09/inboundOrders/{{ record.order_id }}/confirmation` - kind
  `update`; body type `none`; path fields `order_id`; required record fields `order_id`; accepted
  fields `order_id`; confirmation `destructive`; risk: POST
  /awd/2024-05-09/inboundOrders/{orderId}/confirmation against the live SP-API account.
- `update_inbound`: PUT `/awd/2024-05-09/inboundOrders/{{ record.order_id }}` - kind `update`; body
  type `json`; path fields `order_id`; required record fields `order_id`; accepted fields
  `order_id`; confirmation `destructive`; risk: PUT /awd/2024-05-09/inboundOrders/{orderId} against
  the live SP-API account.
- `create_inbound`: POST `/awd/2024-05-09/inboundOrders` - kind `create`; body type `json`;
  confirmation `destructive`; risk: POST /awd/2024-05-09/inboundOrders against the live SP-API
  account.
- `update_inbound_shipment_transport_details`: PUT `/awd/2024-05-09/inboundShipments/{{
  record.shipment_id }}/transport` - kind `update`; body type `json`; path fields `shipment_id`;
  required record fields `shipment_id`; accepted fields `shipment_id`; confirmation `destructive`;
  risk: PUT /awd/2024-05-09/inboundShipments/{shipmentId}/transport against the live SP-API account.
- `confirm_outbound`: POST `/awd/2024-05-09/outboundOrders/{{ record.order_id }}/confirmation` -
  kind `update`; body type `none`; path fields `order_id`; required record fields `order_id`;
  accepted fields `order_id`; confirmation `destructive`; risk: POST
  /awd/2024-05-09/outboundOrders/{orderId}/confirmation against the live SP-API account.
- `update_outbound`: PUT `/awd/2024-05-09/outboundOrders/{{ record.order_id }}` - kind `update`;
  body type `json`; path fields `order_id`; required record fields `order_id`; accepted fields
  `order_id`; confirmation `destructive`; risk: PUT /awd/2024-05-09/outboundOrders/{orderId} against
  the live SP-API account.
- `create_outbound`: POST `/awd/2024-05-09/outboundOrders` - kind `create`; body type `json`;
  confirmation `destructive`; risk: POST /awd/2024-05-09/outboundOrders against the live SP-API
  account.
- `confirm_replenishment_order`: POST `/awd/2024-05-09/replenishmentOrders/{{ record.order_id
  }}/confirmation` - kind `update`; body type `none`; path fields `order_id`; required record fields
  `order_id`; accepted fields `order_id`; confirmation `destructive`; risk: POST
  /awd/2024-05-09/replenishmentOrders/{orderId}/confirmation against the live SP-API account.
- `create_replenishment_order`: POST `/awd/2024-05-09/replenishmentOrders` - kind `create`; body
  type `json`; confirmation `destructive`; risk: POST /awd/2024-05-09/replenishmentOrders against
  the live SP-API account.
- `cancel_query`: DELETE `/dataKiosk/2023-11-15/queries/{{ record.query_id }}` - kind `delete`; body
  type `none`; path fields `query_id`; required record fields `query_id`; accepted fields
  `query_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  DELETE /dataKiosk/2023-11-15/queries/{queryId} against the live SP-API account.
- `create_query`: POST `/dataKiosk/2023-11-15/queries` - kind `create`; body type `json`;
  confirmation `destructive`; risk: POST /dataKiosk/2023-11-15/queries against the live SP-API
  account.
- `update_scheduled_packages`: PATCH `/easyShip/2022-03-23/package` - kind `update`; body type
  `json`; confirmation `destructive`; risk: PATCH /easyShip/2022-03-23/package against the live
  SP-API account.
- `create_scheduled_package`: POST `/easyShip/2022-03-23/package` - kind `create`; body type `json`;
  confirmation `destructive`; risk: POST /easyShip/2022-03-23/package against the live SP-API
  account.
- `create_scheduled_package_bulk`: POST `/easyShip/2022-03-23/packages/bulk` - kind `create`; body
  type `json`; confirmation `destructive`; risk: POST /easyShip/2022-03-23/packages/bulk against the
  live SP-API account.
- `update_package_status`: PATCH `/externalFulfillment/2024-09-11/shipments/{{ record.shipment_id
  }}/packages/{{ record.package_id }}` - kind `update`; body type `json`; path fields `shipment_id`,
  `package_id`; required record fields `shipment_id`, `package_id`; accepted fields `package_id`,
  `shipment_id`; confirmation `destructive`; risk: PATCH
  /externalFulfillment/2024-09-11/shipments/{shipmentId}/packages/{packageId} against the live
  SP-API account.
- `update_package`: PUT `/externalFulfillment/2024-09-11/shipments/{{ record.shipment_id
  }}/packages/{{ record.package_id }}` - kind `update`; body type `json`; path fields `shipment_id`,
  `package_id`; required record fields `shipment_id`, `package_id`; accepted fields `package_id`,
  `shipment_id`; confirmation `destructive`; risk: PUT
  /externalFulfillment/2024-09-11/shipments/{shipmentId}/packages/{packageId} against the live
  SP-API account.
- `create_packages`: POST `/externalFulfillment/2024-09-11/shipments/{{ record.shipment_id
  }}/packages` - kind `create`; body type `json`; path fields `shipment_id`; required record fields
  `shipment_id`; accepted fields `shipment_id`; confirmation `destructive`; risk: POST
  /externalFulfillment/2024-09-11/shipments/{shipmentId}/packages against the live SP-API account.
- `batch_inventory`: POST `/externalFulfillment/inventory/2024-09-11/inventories` - kind `custom`;
  body type `json`; risk: POST /externalFulfillment/inventory/2024-09-11/inventories against the
  live SP-API account.
- `add_inventory`: POST `/fba/inventory/v1/items/inventory` - kind `custom`; body type `json`; risk:
  POST /fba/inventory/v1/items/inventory against the live SP-API account.
- `create_inventory_item`: POST `/fba/inventory/v1/items` - kind `create`; body type `json`;
  confirmation `destructive`; risk: POST /fba/inventory/v1/items against the live SP-API account.
- `delivery_offers`: POST `/fba/outbound/2020-07-01/deliveryOffers` - kind `custom`; body type
  `json`; risk: POST /fba/outbound/2020-07-01/deliveryOffers against the live SP-API account.
- `cancel_fulfillment_order`: PUT `/fba/outbound/2020-07-01/fulfillmentOrders/{{
  record.seller_fulfillment_order_id }}/cancel` - kind `custom`; body type `none`; path fields
  `seller_fulfillment_order_id`; required record fields `seller_fulfillment_order_id`; accepted
  fields `seller_fulfillment_order_id`; confirmation `destructive`; risk: PUT
  /fba/outbound/2020-07-01/fulfillmentOrders/{sellerFulfillmentOrderId}/cancel against the live
  SP-API account.
- `create_fulfillment_return`: PUT `/fba/outbound/2020-07-01/fulfillmentOrders/{{
  record.seller_fulfillment_order_id }}/return` - kind `create`; body type `json`; path fields
  `seller_fulfillment_order_id`; required record fields `seller_fulfillment_order_id`; accepted
  fields `seller_fulfillment_order_id`; confirmation `destructive`; risk: PUT
  /fba/outbound/2020-07-01/fulfillmentOrders/{sellerFulfillmentOrderId}/return against the live
  SP-API account.
- `submit_fulfillment_order_status_update`: PUT `/fba/outbound/2020-07-01/fulfillmentOrders/{{
  record.seller_fulfillment_order_id }}/status` - kind `update`; body type `json`; path fields
  `seller_fulfillment_order_id`; required record fields `seller_fulfillment_order_id`; accepted
  fields `seller_fulfillment_order_id`; confirmation `destructive`; risk: PUT
  /fba/outbound/2020-07-01/fulfillmentOrders/{sellerFulfillmentOrderId}/status against the live
  SP-API account.
- `update_fulfillment_order`: PUT `/fba/outbound/2020-07-01/fulfillmentOrders/{{
  record.seller_fulfillment_order_id }}` - kind `update`; body type `json`; path fields
  `seller_fulfillment_order_id`; required record fields `seller_fulfillment_order_id`; accepted
  fields `seller_fulfillment_order_id`; confirmation `destructive`; risk: PUT
  /fba/outbound/2020-07-01/fulfillmentOrders/{sellerFulfillmentOrderId} against the live SP-API
  account.
- `create_fulfillment_order`: POST `/fba/outbound/2020-07-01/fulfillmentOrders` - kind `create`;
  body type `json`; confirmation `destructive`; risk: POST
  /fba/outbound/2020-07-01/fulfillmentOrders against the live SP-API account.
- `cancel_feed`: DELETE `/feeds/2021-06-30/feeds/{{ record.feed_id }}` - kind `delete`; body type
  `none`; path fields `feed_id`; required record fields `feed_id`; accepted fields `feed_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /feeds/2021-06-30/feeds/{feedId} against the live SP-API account.
- `create_feed`: POST `/feeds/2021-06-30/feeds` - kind `create`; body type `json`; confirmation
  `destructive`; risk: POST /feeds/2021-06-30/feeds against the live SP-API account.
- `initiate_payout`: POST `/finances/transfers/2024-06-01/payouts` - kind `update`; body type
  `json`; confirmation `destructive`; risk: POST /finances/transfers/2024-06-01/payouts against the
  live SP-API account.
- `cancel_inbound_plan`: PUT `/inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id
  }}/cancellation` - kind `custom`; body type `none`; path fields `inbound_plan_id`; required record
  fields `inbound_plan_id`; accepted fields `inbound_plan_id`; confirmation `destructive`; risk: PUT
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/cancellation against the live SP-API account.
- `update_inbound_plan_name`: PUT `/inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id
  }}/name` - kind `update`; body type `json`; path fields `inbound_plan_id`; required record fields
  `inbound_plan_id`; accepted fields `inbound_plan_id`; confirmation `destructive`; risk: PUT
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/name against the live SP-API account.
- `set_packing_information`: POST `/inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id
  }}/packingInformation` - kind `update`; body type `json`; path fields `inbound_plan_id`; required
  record fields `inbound_plan_id`; accepted fields `inbound_plan_id`; risk: POST
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/packingInformation against the live SP-API
  account.
- `confirm_packing_option`: POST `/inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id
  }}/packingOptions/{{ record.packing_option_id }}/confirmation` - kind `update`; body type `none`;
  path fields `inbound_plan_id`, `packing_option_id`; required record fields `inbound_plan_id`,
  `packing_option_id`; accepted fields `inbound_plan_id`, `packing_option_id`; confirmation
  `destructive`; risk: POST
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/packingOptions/{packingOptionId}/confirmation
  against the live SP-API account.
- `generate_packing_options`: POST `/inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id
  }}/packingOptions` - kind `update`; body type `none`; path fields `inbound_plan_id`; required
  record fields `inbound_plan_id`; accepted fields `inbound_plan_id`; confirmation `destructive`;
  risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/packingOptions against the live
  SP-API account.
- `confirm_placement_option`: POST `/inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id
  }}/placementOptions/{{ record.placement_option_id }}/confirmation` - kind `update`; body type
  `none`; path fields `inbound_plan_id`, `placement_option_id`; required record fields
  `inbound_plan_id`, `placement_option_id`; accepted fields `inbound_plan_id`,
  `placement_option_id`; confirmation `destructive`; risk: POST
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/placementOptions/{placementOptionId}/confirmation
  against the live SP-API account.
- `generate_placement_options`: POST `/inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id
  }}/placementOptions` - kind `update`; body type `json`; path fields `inbound_plan_id`; required
  record fields `inbound_plan_id`; accepted fields `inbound_plan_id`; confirmation `destructive`;
  risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/placementOptions against the live
  SP-API account.
- `confirm_delivery_window_options`: POST `/inbound/fba/2024-03-20/inboundPlans/{{
  record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/deliveryWindowOptions/{{
  record.delivery_window_option_id }}/confirmation` - kind `update`; body type `none`; path fields
  `inbound_plan_id`, `shipment_id`, `delivery_window_option_id`; required record fields
  `inbound_plan_id`, `shipment_id`, `delivery_window_option_id`; accepted fields
  `delivery_window_option_id`, `inbound_plan_id`, `shipment_id`; confirmation `destructive`; risk:
  POST
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/deliveryWindowOptions/{deliveryWindowOptionId}/confirmation
  against the live SP-API account.
- `generate_delivery_window_options`: POST `/inbound/fba/2024-03-20/inboundPlans/{{
  record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/deliveryWindowOptions` - kind
  `update`; body type `none`; path fields `inbound_plan_id`, `shipment_id`; required record fields
  `inbound_plan_id`, `shipment_id`; accepted fields `inbound_plan_id`, `shipment_id`; confirmation
  `destructive`; risk: POST
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/deliveryWindowOptions
  against the live SP-API account.
- `update_shipment_name`: PUT `/inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id
  }}/shipments/{{ record.shipment_id }}/name` - kind `update`; body type `json`; path fields
  `inbound_plan_id`, `shipment_id`; required record fields `inbound_plan_id`, `shipment_id`;
  accepted fields `inbound_plan_id`, `shipment_id`; confirmation `destructive`; risk: PUT
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/name against the live
  SP-API account.
- `cancel_self_ship_appointment`: PUT `/inbound/fba/2024-03-20/inboundPlans/{{
  record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/selfShipAppointmentCancellation` -
  kind `custom`; body type `json`; path fields `inbound_plan_id`, `shipment_id`; required record
  fields `inbound_plan_id`, `shipment_id`; accepted fields `inbound_plan_id`, `shipment_id`;
  confirmation `destructive`; risk: PUT
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/selfShipAppointmentCancellation
  against the live SP-API account.
- `schedule_self_ship_appointment`: POST `/inbound/fba/2024-03-20/inboundPlans/{{
  record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/selfShipAppointmentSlots/{{
  record.slot_id }}/schedule` - kind `update`; body type `json`; path fields `inbound_plan_id`,
  `shipment_id`, `slot_id`; required record fields `inbound_plan_id`, `shipment_id`, `slot_id`;
  accepted fields `inbound_plan_id`, `shipment_id`, `slot_id`; confirmation `destructive`; risk:
  POST
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/selfShipAppointmentSlots/{slotId}/schedule
  against the live SP-API account.
- `update_shipment_source_address`: PUT `/inbound/fba/2024-03-20/inboundPlans/{{
  record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/sourceAddress` - kind `update`; body
  type `json`; path fields `inbound_plan_id`, `shipment_id`; required record fields
  `inbound_plan_id`, `shipment_id`; accepted fields `inbound_plan_id`, `shipment_id`; confirmation
  `destructive`; risk: PUT
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/sourceAddress against
  the live SP-API account.
- `update_shipment_tracking_details`: PUT `/inbound/fba/2024-03-20/inboundPlans/{{
  record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/trackingDetails` - kind `update`;
  body type `json`; path fields `inbound_plan_id`, `shipment_id`; required record fields
  `inbound_plan_id`, `shipment_id`; accepted fields `inbound_plan_id`, `shipment_id`; confirmation
  `destructive`; risk: PUT
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/trackingDetails
  against the live SP-API account.
- `confirm_transportation_options`: POST `/inbound/fba/2024-03-20/inboundPlans/{{
  record.inbound_plan_id }}/transportationOptions/confirmation` - kind `update`; body type `json`;
  path fields `inbound_plan_id`; required record fields `inbound_plan_id`; accepted fields
  `inbound_plan_id`; confirmation `destructive`; risk: POST
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/transportationOptions/confirmation against
  the live SP-API account.
- `generate_transportation_options`: POST `/inbound/fba/2024-03-20/inboundPlans/{{
  record.inbound_plan_id }}/transportationOptions` - kind `update`; body type `json`; path fields
  `inbound_plan_id`; required record fields `inbound_plan_id`; accepted fields `inbound_plan_id`;
  confirmation `destructive`; risk: POST
  /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/transportationOptions against the live SP-API
  account.
- `create_inbound_plan`: POST `/inbound/fba/2024-03-20/inboundPlans` - kind `create`; body type
  `json`; confirmation `destructive`; risk: POST /inbound/fba/2024-03-20/inboundPlans against the
  live SP-API account.
- `set_prep_details`: POST `/inbound/fba/2024-03-20/items/prepDetails` - kind `update`; body type
  `json`; risk: POST /inbound/fba/2024-03-20/items/prepDetails against the live SP-API account.
- `cancel_shipment`: DELETE `/mfn/v0/shipments/{{ record.shipment_id }}` - kind `delete`; body type
  `none`; path fields `shipment_id`; required record fields `shipment_id`; accepted fields
  `shipment_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: DELETE /mfn/v0/shipments/{shipmentId} against the live SP-API account.
- `create_shipment`: POST `/mfn/v0/shipments` - kind `create`; body type `json`; confirmation
  `destructive`; risk: POST /mfn/v0/shipments against the live SP-API account.
- `delete_destination`: DELETE `/notifications/v1/destinations/{{ record.destination_id }}` - kind
  `delete`; body type `none`; path fields `destination_id`; required record fields `destination_id`;
  accepted fields `destination_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: DELETE /notifications/v1/destinations/{destinationId} against
  the live SP-API account.
- `create_destination`: POST `/notifications/v1/destinations` - kind `create`; body type `json`;
  confirmation `destructive`; risk: POST /notifications/v1/destinations against the live SP-API
  account.
- `delete_subscription_by_id`: DELETE `/notifications/v1/subscriptions/{{ record.notification_type
  }}/{{ record.subscription_id }}` - kind `delete`; body type `none`; path fields
  `notification_type`, `subscription_id`; required record fields `notification_type`,
  `subscription_id`; accepted fields `notification_type`, `subscription_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: DELETE
  /notifications/v1/subscriptions/{notificationType}/{subscriptionId} against the live SP-API
  account.
- `send_test_notification`: POST `/notifications/v1/subscriptions/{{ record.notification_type
  }}/testNotification` - kind `custom`; body type `json`; path fields `notification_type`; required
  record fields `notification_type`; accepted fields `notification_type`; risk: POST
  /notifications/v1/subscriptions/{notificationType}/testNotification against the live SP-API
  account.
- `create_subscription`: POST `/notifications/v1/subscriptions/{{ record.notification_type }}` -
  kind `create`; body type `json`; path fields `notification_type`; required record fields
  `notification_type`; accepted fields `notification_type`; confirmation `destructive`; risk: POST
  /notifications/v1/subscriptions/{notificationType} against the live SP-API account.
- `update_verification_status`: PATCH `/orders/v0/orders/{{ record.order_id }}/regulatedInfo` - kind
  `update`; body type `json`; path fields `order_id`; required record fields `order_id`; accepted
  fields `order_id`; confirmation `destructive`; risk: PATCH
  /orders/v0/orders/{orderId}/regulatedInfo against the live SP-API account.
- `confirm_shipment`: POST `/orders/v0/orders/{{ record.order_id }}/shipmentConfirmation` - kind
  `update`; body type `json`; path fields `order_id`; required record fields `order_id`; accepted
  fields `order_id`; confirmation `destructive`; risk: POST
  /orders/v0/orders/{orderId}/shipmentConfirmation against the live SP-API account.
- `update_shipment_status`: POST `/orders/v0/orders/{{ record.order_id }}/shipment` - kind `update`;
  body type `json`; path fields `order_id`; required record fields `order_id`; accepted fields
  `order_id`; confirmation `destructive`; risk: POST /orders/v0/orders/{orderId}/shipment against
  the live SP-API account.
- `cancel_report`: DELETE `/reports/2021-06-30/reports/{{ record.report_id }}` - kind `delete`; body
  type `none`; path fields `report_id`; required record fields `report_id`; accepted fields
  `report_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: DELETE /reports/2021-06-30/reports/{reportId} against the live SP-API account.
- `create_report`: POST `/reports/2021-06-30/reports` - kind `create`; body type `json`;
  confirmation `destructive`; risk: POST /reports/2021-06-30/reports against the live SP-API
  account.
- `cancel_report_schedule`: DELETE `/reports/2021-06-30/schedules/{{ record.report_schedule_id }}` -
  kind `delete`; body type `none`; path fields `report_schedule_id`; required record fields
  `report_schedule_id`; accepted fields `report_schedule_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: DELETE
  /reports/2021-06-30/schedules/{reportScheduleId} against the live SP-API account.
- `create_report_schedule`: POST `/reports/2021-06-30/schedules` - kind `create`; body type `json`;
  confirmation `destructive`; risk: POST /reports/2021-06-30/schedules against the live SP-API
  account.
- `set_appointment_fulfillment_data`: PUT `/service/v1/serviceJobs/{{ record.service_job_id
  }}/appointments/{{ record.appointment_id }}/fulfillment` - kind `update`; body type `json`; path
  fields `service_job_id`, `appointment_id`; required record fields `service_job_id`,
  `appointment_id`; accepted fields `appointment_id`, `service_job_id`; risk: PUT
  /service/v1/serviceJobs/{serviceJobId}/appointments/{appointmentId}/fulfillment against the live
  SP-API account.
- `assign_appointment_resources`: PUT `/service/v1/serviceJobs/{{ record.service_job_id
  }}/appointments/{{ record.appointment_id }}/resources` - kind `update`; body type `json`; path
  fields `service_job_id`, `appointment_id`; required record fields `service_job_id`,
  `appointment_id`; accepted fields `appointment_id`, `service_job_id`; risk: PUT
  /service/v1/serviceJobs/{serviceJobId}/appointments/{appointmentId}/resources against the live
  SP-API account.
- `reschedule_appointment_for_service_job_by_service_job_id`: POST `/service/v1/serviceJobs/{{
  record.service_job_id }}/appointments/{{ record.appointment_id }}` - kind `update`; body type
  `json`; path fields `service_job_id`, `appointment_id`; required record fields `service_job_id`,
  `appointment_id`; accepted fields `appointment_id`, `service_job_id`; confirmation `destructive`;
  risk: POST /service/v1/serviceJobs/{serviceJobId}/appointments/{appointmentId} against the live
  SP-API account.
- `add_appointment_for_service_job_by_service_job_id`: POST `/service/v1/serviceJobs/{{
  record.service_job_id }}/appointments` - kind `custom`; body type `json`; path fields
  `service_job_id`; required record fields `service_job_id`; accepted fields `service_job_id`; risk:
  POST /service/v1/serviceJobs/{serviceJobId}/appointments against the live SP-API account.
- `complete_service_job_by_service_job_id`: PUT `/service/v1/serviceJobs/{{ record.service_job_id
  }}/completions` - kind `update`; body type `none`; path fields `service_job_id`; required record
  fields `service_job_id`; accepted fields `service_job_id`; confirmation `destructive`; risk: PUT
  /service/v1/serviceJobs/{serviceJobId}/completions against the live SP-API account.
- `purchase_shipment`: POST `/shipping/v1/purchaseShipment` - kind `update`; body type `json`;
  confirmation `destructive`; risk: POST /shipping/v1/purchaseShipment against the live SP-API
  account.
- `cancel_shipment_post_shipping_v1_shipments_shipment_id_cancel`: POST `/shipping/v1/shipments/{{
  record.shipment_id }}/cancel` - kind `custom`; body type `none`; path fields `shipment_id`;
  required record fields `shipment_id`; accepted fields `shipment_id`; confirmation `destructive`;
  risk: POST /shipping/v1/shipments/{shipmentId}/cancel against the live SP-API account.
- `create_shipment_post_shipping_v1_shipments`: POST `/shipping/v1/shipments` - kind `create`; body
  type `json`; confirmation `destructive`; risk: POST /shipping/v1/shipments against the live SP-API
  account.
- `unlink_carrier_account`: PUT `/shipping/v2/carrierAccounts/{{ record.carrier_id }}/unlink` - kind
  `update`; body type `json`; path fields `carrier_id`; required record fields `carrier_id`;
  accepted fields `carrier_id`; confirmation `destructive`; risk: PUT
  /shipping/v2/carrierAccounts/{carrierId}/unlink against the live SP-API account.
- `link_carrier_account`: POST `/shipping/v2/carrierAccounts/{{ record.carrier_id }}` - kind
  `update`; body type `json`; path fields `carrier_id`; required record fields `carrier_id`;
  accepted fields `carrier_id`; risk: POST /shipping/v2/carrierAccounts/{carrierId} against the live
  SP-API account.
- `link_carrier_account_put_shipping_v2_carrier_accounts_carrier_id`: PUT
  `/shipping/v2/carrierAccounts/{{ record.carrier_id }}` - kind `update`; body type `json`; path
  fields `carrier_id`; required record fields `carrier_id`; accepted fields `carrier_id`; risk: PUT
  /shipping/v2/carrierAccounts/{carrierId} against the live SP-API account.
- `create_claim`: POST `/shipping/v2/claims` - kind `create`; body type `json`; confirmation
  `destructive`; risk: POST /shipping/v2/claims against the live SP-API account.
- `generate_collection_form`: POST `/shipping/v2/collectionForms` - kind `update`; body type `json`;
  confirmation `destructive`; risk: POST /shipping/v2/collectionForms against the live SP-API
  account.
- `submit_ndr_feedback`: POST `/shipping/v2/ndrFeedback` - kind `update`; body type `json`;
  confirmation `destructive`; risk: POST /shipping/v2/ndrFeedback against the live SP-API account.
- `one_click_shipment`: POST `/shipping/v2/oneClickShipment` - kind `custom`; body type `json`;
  confirmation `destructive`; risk: POST /shipping/v2/oneClickShipment against the live SP-API
  account.
- `cancel_shipment_put_shipping_v2_shipments_shipment_id_cancel`: PUT `/shipping/v2/shipments/{{
  record.shipment_id }}/cancel` - kind `custom`; body type `none`; path fields `shipment_id`;
  required record fields `shipment_id`; accepted fields `shipment_id`; confirmation `destructive`;
  risk: PUT /shipping/v2/shipments/{shipmentId}/cancel against the live SP-API account.
- `direct_purchase_shipment`: POST `/shipping/v2/shipments/directPurchase` - kind `update`; body
  type `json`; confirmation `destructive`; risk: POST /shipping/v2/shipments/directPurchase against
  the live SP-API account.
- `purchase_shipment_post_shipping_v2_shipments`: POST `/shipping/v2/shipments` - kind `update`;
  body type `json`; confirmation `destructive`; risk: POST /shipping/v2/shipments against the live
  SP-API account.
- `update_supply_source_status`: PUT `/supplySources/2020-07-01/supplySources/{{
  record.supply_source_id }}/status` - kind `update`; body type `json`; path fields
  `supply_source_id`; required record fields `supply_source_id`; accepted fields `supply_source_id`;
  confirmation `destructive`; risk: PUT
  /supplySources/2020-07-01/supplySources/{supplySourceId}/status against the live SP-API account.
- `archive_supply_source`: DELETE `/supplySources/2020-07-01/supplySources/{{
  record.supply_source_id }}` - kind `delete`; body type `none`; path fields `supply_source_id`;
  required record fields `supply_source_id`; accepted fields `supply_source_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: DELETE
  /supplySources/2020-07-01/supplySources/{supplySourceId} against the live SP-API account.
- `update_supply_source`: PUT `/supplySources/2020-07-01/supplySources/{{ record.supply_source_id
  }}` - kind `update`; body type `json`; path fields `supply_source_id`; required record fields
  `supply_source_id`; accepted fields `supply_source_id`; confirmation `destructive`; risk: PUT
  /supplySources/2020-07-01/supplySources/{supplySourceId} against the live SP-API account.
- `create_supply_source`: POST `/supplySources/2020-07-01/supplySources` - kind `create`; body type
  `json`; confirmation `destructive`; risk: POST /supplySources/2020-07-01/supplySources against the
  live SP-API account.
- `submit_inventory_update`: POST `/vendor/directFulfillment/inventory/v1/warehouses/{{
  record.warehouse_id }}/items` - kind `update`; body type `json`; path fields `warehouse_id`;
  required record fields `warehouse_id`; accepted fields `warehouse_id`; confirmation `destructive`;
  risk: POST /vendor/directFulfillment/inventory/v1/warehouses/{warehouseId}/items against the live
  SP-API account.
- `submit_acknowledgement`: POST `/vendor/directFulfillment/orders/2021-12-28/acknowledgements` -
  kind `update`; body type `json`; confirmation `destructive`; risk: POST
  /vendor/directFulfillment/orders/2021-12-28/acknowledgements against the live SP-API account.
- `submit_acknowledgement_post_vendor_direct_fulfillment_orders_v1_acknowledgements`: POST
  `/vendor/directFulfillment/orders/v1/acknowledgements` - kind `update`; body type `json`;
  confirmation `destructive`; risk: POST /vendor/directFulfillment/orders/v1/acknowledgements
  against the live SP-API account.
- `generate_order_scenarios`: POST `/vendor/directFulfillment/sandbox/2021-10-28/orders` - kind
  `update`; body type `json`; confirmation `destructive`; risk: POST
  /vendor/directFulfillment/sandbox/2021-10-28/orders against the live SP-API account.
- `submit_shipment_confirmations`: POST
  `/vendor/directFulfillment/shipping/2021-12-28/shipmentConfirmations` - kind `update`; body type
  `json`; confirmation `destructive`; risk: POST
  /vendor/directFulfillment/shipping/2021-12-28/shipmentConfirmations against the live SP-API
  account.
- `submit_shipment_status_updates`: POST
  `/vendor/directFulfillment/shipping/2021-12-28/shipmentStatusUpdates` - kind `update`; body type
  `json`; confirmation `destructive`; risk: POST
  /vendor/directFulfillment/shipping/2021-12-28/shipmentStatusUpdates against the live SP-API
  account.
- `submit_shipment_confirmations_post_vendor_direct_fulfillment_shipping_v1_shipment_confirmations`:
  POST `/vendor/directFulfillment/shipping/v1/shipmentConfirmations` - kind `update`; body type
  `json`; confirmation `destructive`; risk: POST
  /vendor/directFulfillment/shipping/v1/shipmentConfirmations against the live SP-API account.
- `submit_shipment_status_updates_post_vendor_direct_fulfillment_shipping_v1_shipment_status_updates`:
  POST `/vendor/directFulfillment/shipping/v1/shipmentStatusUpdates` - kind `update`; body type
  `json`; confirmation `destructive`; risk: POST
  /vendor/directFulfillment/shipping/v1/shipmentStatusUpdates against the live SP-API account.
- `submit_acknowledgement_post_vendor_orders_v1_acknowledgements`: POST
  `/vendor/orders/v1/acknowledgements` - kind `update`; body type `json`; confirmation
  `destructive`; risk: POST /vendor/orders/v1/acknowledgements against the live SP-API account.
- `submit_shipment_confirmations_post_vendor_shipping_v1_shipment_confirmations`: POST
  `/vendor/shipping/v1/shipmentConfirmations` - kind `update`; body type `json`; confirmation
  `destructive`; risk: POST /vendor/shipping/v1/shipmentConfirmations against the live SP-API
  account.
- `submit_shipments`: POST `/vendor/shipping/v1/shipments` - kind `update`; body type `json`;
  confirmation `destructive`; risk: POST /vendor/shipping/v1/shipments against the live SP-API
  account.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 147 stream-backed endpoint group(s), 98 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=40, non_data_endpoint=4, out_of_scope=54, requires_elevated_scope=10.
