# pm connectors inspect amazon-seller-partner

```text
NAME
  pm connectors inspect amazon-seller-partner - Amazon Seller Partner connector manual

SYNOPSIS
  pm connectors inspect amazon-seller-partner
  pm connectors inspect amazon-seller-partner --json
  pm credentials add <name> --connector amazon-seller-partner [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Amazon Selling Partner API orders, inventory, finance, catalog, listings, fulfillment, reports, feeds, seller, shipping, vendor, and supporting JSON resources via Login with Amazon (LWA) authentication; exposes declarative writes for SP-API mutations that fit path/body JSON requests.

ICON
  asset: icons/amazonsellerpartner.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer-docs.amazon.com/sp-api/docs/sp-api-deprecations

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  access_point_types
  account_id
  asin
  base_amount
  base_url
  browse_node_id
  carrier_id
  collection_form_id
  content_reference_key
  content_update_preview_id
  country_code
  created_after
  created_before
  destination_country_code
  destination_currency_code
  destination_id
  event_group_id
  export_id
  feature_name
  feed_id
  inbound_plan_id
  invoice_id
  item_condition
  item_type
  keywords
  lwa_token_url
  marketplace_id
  max_pages
  metrics_granularity
  metrics_interval
  mskus
  notification_type
  order_id
  package_client_reference_id
  package_id
  package_number
  packing_group_id
  page_size
  postal_code
  product_type
  program
  purchase_order_number
  query_id
  query_type
  rate_id
  replication_start_date
  report_id
  report_schedule_id
  report_types
  request_token
  return_id
  seller_fulfillment_order_id
  seller_id
  seller_sku
  service_job_id
  ship_to_country_code
  shipment_id
  sku
  sort_by
  source_country_code
  source_currency_code
  status
  store_id
  subscription_id
  supply_source_id
  tracking_id
  transaction_id
  transfer_schedule_id
  vehicle_type
  lwa_app_id (secret)
  lwa_client_secret (secret)
  refresh_token (secret)

ETL STREAMS
  orders:
    primary key: AmazonOrderId
    cursor: LastUpdateDate
    fields: AmazonOrderId(), FulfillmentChannel(), IsBusinessOrder(), IsPrime(), LastUpdateDate(), MarketplaceId(), NumberOfItemsShipped(), NumberOfItemsUnshipped(), OrderStatus(), OrderTotal(), OrderType(), PurchaseDate(), SalesChannel(), SellerOrderId()
  inventory_summaries:
    primary key: sellerSku
    cursor: lastUpdatedTime
    fields: asin(), condition(), fnSku(), inventoryDetails(), lastUpdatedTime(), productName(), sellerSku(), totalQuantity()
  financial_event_groups:
    primary key: FinancialEventGroupId
    cursor: FinancialEventGroupEnd
    fields: AccountTail(), BeginningBalance(), ConvertedTotal(), FinancialEventGroupEnd(), FinancialEventGroupId(), FinancialEventGroupStart(), FundTransferDate(), FundTransferStatus(), OriginalTotal(), ProcessingStatus(), TraceId()
  list_content_document_asin_relations:
    primary key: asin
    fields: asin(), badgeSet(), contentReferenceKeySet(), content_reference_key(), imageUrl(), parent(), title()
  search_content_documents:
    primary key: contentReferenceKey
    fields: contentMetadata(), contentReferenceKey()
  search_content_publish_records:
    primary key: contentReferenceKey
    fields: asin(), contentReferenceKey(), contentSubType(), contentType(), locale(), marketplaceId()
  get_inbound:
    primary key: orderId
    fields: createdAt(), destinationDetails(), externalReferenceId(), orderId(), orderStatus(), order_id(), originAddress(), packagesToInbound(), preferences(), updatedAt()
  get_label_page_types:
    primary key: shipment_id
    fields: shipment_id()
  get_inbound_shipment:
    primary key: orderId
    fields: carrierCode(), createdAt(), destinationAddress(), destinationRegion(), externalReferenceId(), orderId(), originAddress(), receivedQuantity(), shipBy(), shipmentContainerQuantities(), shipmentId(), shipmentSkuQuantities(), shipmentStatus(), shipment_id(), trackingId(), updatedAt(), warehouseReferenceId()
  list_inbound_shipments:
    primary key: orderId
    fields: createdAt(), externalReferenceId(), orderId(), shipmentId(), shipmentStatus(), updatedAt()
  list_inventory:
    primary key: sku
    fields: expirationDetails(), inventoryDetails(), sku(), totalInboundQuantity(), totalOnhandQuantity()
  get_outbound:
    primary key: orderId
    fields: confirmedOn(), createdAt(), eligiblePackagesToOutbound(), eligibleProductsToOutbound(), executionErrors(), orderId(), orderPreferences(), orderStatus(), order_id(), outboundShipments(), packagesToOutbound(), productsToOutbound(), shippedOutboundPackages(), shippedOutboundProducts(), updatedAt()
  list_outbounds:
    primary key: orderId
    fields: confirmedOn(), createdAt(), eligiblePackagesToOutbound(), eligibleProductsToOutbound(), executionErrors(), orderId(), orderPreferences(), orderStatus(), outboundShipments(), packagesToOutbound(), productsToOutbound(), shippedOutboundPackages(), shippedOutboundProducts(), updatedAt()
  get_replenishment_order:
    primary key: orderId
    fields: confirmedOn(), createdAt(), distributionIneligibleReasons(), eligibleProducts(), orderId(), order_id(), outboundShipments(), products(), shippedProducts(), status(), updatedAt()
  list_replenishment_orders:
    primary key: orderId
    fields: confirmedOn(), createdAt(), distributionIneligibleReasons(), eligibleProducts(), orderId(), outboundShipments(), products(), shippedProducts(), status(), updatedAt()
  get_catalog_item:
    primary key: asin
    fields: asin(), attributes(), identifiers(), images(), productTypes(), salesRanks(), summaries(), variations(), vendorDetails()
  search_catalog_items:
    primary key: asin
    fields: asin(), attributes(), identifiers(), images(), productTypes(), salesRanks(), summaries(), variations(), vendorDetails()
  get_catalog_item_catalog_2022_04_01_items_asin:
    primary key: asin
    fields: asin(), attributes(), classifications(), dimensions(), identifiers(), images(), productTypes(), relationships(), salesRanks(), summaries(), vendorDetails()
  search_catalog_items_catalog_2022_04_01_items:
    primary key: asin
    fields: asin(), attributes(), classifications(), dimensions(), identifiers(), images(), productTypes(), relationships(), salesRanks(), summaries(), vendorDetails()
  get_vehicles:
    fields: bodyStyle(), driveType(), energy(), engineOutput(), identifiers(), lastProcessedDate(), make(), manufacturingStartDate(), manufacturingStopDate(), model(), status(), variantName()
  list_catalog_categories:
    primary key: ProductCategoryId
    fields: ProductCategoryId(), ProductCategoryName(), parent()
  get_browse_node_return_topics:
    primary key: browse_node_id
    fields: browseNodeMetrics(), browse_node_id(), topic()
  get_browse_node_return_trends:
    primary key: browse_node_id
    fields: browse_node_id(), topic(), trendMetrics()
  get_browse_node_review_topics:
    primary key: browse_node_id
    fields: browseNodeMetrics(), browse_node_id(), reviewSnippets(), subtopics(), topic()
  get_browse_node_review_trends:
    primary key: browse_node_id
    fields: browse_node_id(), topic(), trendMetrics()
  get_item_browse_node:
    primary key: asin
    fields: asin(), browseNodeId(), displayName()
  get_item_review_topics:
    primary key: asin
    fields: asin(), asinMetrics(), browseNodeMetrics(), childAsinMetrics(), parentAsinMetrics(), reviewSnippets(), subtopics(), topic()
  get_item_review_trends:
    primary key: asin
    fields: asin(), topic(), trendMetrics()
  get_query:
    primary key: queryId
    fields: createdTime(), dataDocumentId(), errorDocumentId(), pagination(), processingEndTime(), processingStartTime(), processingStatus(), query(), queryId(), query_id()
  get_queries:
    primary key: queryId
    fields: createdTime(), dataDocumentId(), errorDocumentId(), pagination(), processingEndTime(), processingStartTime(), processingStatus(), query(), queryId()
  get_definitions_product_type:
    primary key: product_type
    fields: displayName(), locale(), marketplaceIds(), metaSchema(), productType(), productTypeVersion(), product_type(), propertyGroups(), requirements(), requirementsEnforced(), schema()
  search_definitions_product_types:
    fields: displayName(), marketplaceIds(), name()
  get_scheduled_package:
    primary key: orderItemId
    fields: orderItemId(), orderItemSerialNumbers()
  get_return:
    primary key: id
    fields: creationDateTime(), fulfillmentLocationId(), id(), lastUpdatedDateTime(), marketplaceChannelDetails(), merchantSku(), numberOfUnits(), otpDetails(), packageDeliveryMode(), replanningDetails(), returnLocationId(), returnMetadata(), returnShippingInfo(), returnSubType(), returnType(), return_id(), status()
  list_returns:
    primary key: id
    fields: creationDateTime(), fulfillmentLocationId(), id(), lastUpdatedDateTime(), marketplaceChannelDetails(), merchantSku(), numberOfUnits(), otpDetails(), packageDeliveryMode(), replanningDetails(), returnLocationId(), returnMetadata(), returnShippingInfo(), returnSubType(), returnType(), status()
  retrieve_invoice:
    primary key: shipment_id
    fields: content(), format(), shipment_id()
  retrieve_shipping_options:
    primary key: shipment_id
    fields: carrierName(), handoverLocation(), pickupWindow(), shipBy(), shipment_id(), shippingOptionId(), timeSlot()
  get_shipment:
    primary key: id
    fields: charges(), creationDateTime(), earliestPackDateTime(), id(), invoiceInfo(), lastUpdatedDateTime(), lineItems(), locationId(), marketplaceAttributes(), packages(), partyInfoList(), reason(), shipmentInfo(), shipmentRequirements(), shipment_id(), shippingInfo(), status(), subStatus()
  get_shipments:
    primary key: id
    fields: charges(), creationDateTime(), earliestPackDateTime(), id(), invoiceInfo(), lastUpdatedDateTime(), lineItems(), locationId(), marketplaceAttributes(), packages(), partyInfoList(), reason(), shipmentInfo(), shipmentRequirements(), shippingInfo(), status(), subStatus()
  get_prep_instructions:
    primary key: ASIN
    fields: ASIN(), AmazonPrepFeesDetailsList(), BarcodeInstruction(), PrepGuidance(), PrepInstructionList(), SellerSKU()
  get_shipment_items:
    primary key: SellerSKU
    fields: FulfillmentNetworkSKU(), PrepDetailsList(), QuantityInCase(), QuantityReceived(), QuantityShipped(), ReleaseDate(), SellerSKU(), ShipmentId()
  get_shipment_items_by_shipment_id:
    primary key: shipment_id
    fields: FulfillmentNetworkSKU(), PrepDetailsList(), QuantityInCase(), QuantityReceived(), QuantityShipped(), ReleaseDate(), SellerSKU(), ShipmentId(), shipment_id()
  get_shipments_fba_inbound_v0_shipments:
    primary key: ShipmentId
    fields: AreCasesRequired(), BoxContentsSource(), ConfirmedNeedByDate(), DestinationFulfillmentCenterId(), EstimatedBoxContentsFee(), LabelPrepType(), ShipFromAddress(), ShipmentId(), ShipmentName(), ShipmentStatus()
  get_item_eligibility_preview:
  get_feature_sku:
    primary key: asin
    fields: asin(), featureName(), feature_name(), fnSku(), ineligibleReasons(), isEligible(), marketplaceId(), sellerSku(), seller_sku(), skuCount()
  get_feature_inventory:
    primary key: feature_name
    fields: featureName(), featureSkus(), feature_name(), marketplaceId(), nextToken()
  get_features:
    fields: featureDescription(), featureName(), sellerEligible()
  get_fulfillment_order:
    primary key: seller_fulfillment_order_id
    fields: fulfillmentOrder(), fulfillmentOrderItems(), fulfillmentShipments(), paymentInformation(), returnAuthorizations(), returnItems(), seller_fulfillment_order_id()
  list_all_fulfillment_orders:
    primary key: sellerFulfillmentOrderId
    fields: codSettings(), deliveryWindow(), destinationAddress(), displayableOrderComment(), displayableOrderDate(), displayableOrderId(), featureConstraints(), fulfillmentAction(), fulfillmentOrderStatus(), fulfillmentPolicy(), marketplaceId(), notificationEmails(), receivedDate(), sellerFulfillmentOrderId(), shippingSpeedCategory(), statusUpdatedDate()
  list_return_reason_codes:
    fields: description(), returnReasonCode(), translatedDescription()
  get_package_tracking_details:
    fields: eventAddress(), eventCode(), eventDate(), eventDescription()
  get_shipment_details:
    primary key: AmazonOrderId
    fields: AmazonOrderId(), AmazonShipmentId(), BuyerCounty(), BuyerName(), BuyerTaxInfo(), MarketplaceId(), MarketplaceTaxInfo(), PaymentMethodDetails(), Payments(), PurchaseDate(), SellerDisplayName(), SellerId(), ShipmentItems(), ShippingAddress(), WarehouseId(), shipment_id()
  get_feed:
    primary key: feedId
    fields: createdTime(), feedId(), feedType(), feed_id(), marketplaceIds(), processingEndTime(), processingStartTime(), processingStatus(), resultFeedDocumentId()
  get_feeds:
    primary key: feedId
    fields: createdTime(), feedId(), feedType(), marketplaceIds(), processingEndTime(), processingStartTime(), processingStatus(), resultFeedDocumentId()
  list_transactions:
    primary key: transactionId
    fields: breakdowns(), contexts(), description(), items(), marketplaceDetails(), postedDate(), relatedIdentifiers(), sellingPartnerMetadata(), totalAmount(), transactionId(), transactionStatus(), transactionType()
  get_payment_methods:
    primary key: paymentMethodId
    fields: accountHolderName(), assignmentType(), countryCode(), expiryDate(), paymentMethodId(), paymentMethodType(), tail()
  list_account_balances:
    primary key: accountId
    fields: accountId(), account_id(), balanceAmount(), balanceCurrency(), balanceType(), lastUpdateDate()
  get_account:
    primary key: accountId
    fields: accountCountryCode(), accountCurrency(), accountHolderName(), accountId(), account_id(), bankAccountHolderStatus(), bankAccountNumberFormat(), bankAccountNumberTail(), bankAccountOwnershipType(), bankName(), bankNumberFormat(), routingNumber()
  list_accounts:
    primary key: accountId
    fields: accountCountryCode(), accountCurrency(), accountHolderName(), accountId(), bankAccountHolderStatus(), bankAccountNumberFormat(), bankAccountNumberTail(), bankAccountOwnershipType(), bankName(), bankNumberFormat(), routingNumber()
  get_transaction:
    primary key: transaction_id
    fields: transaction_id()
  list_account_transactions:
    primary key: transactionId
    fields: accountId(), expectedCompletionDate(), lastUpdateDate(), requesterName(), transactionActualCompletionDate(), transactionDescription(), transactionDestinationAccount(), transactionFailureReason(), transactionFinalAmount(), transactionId(), transactionRequestAmount(), transactionRequestDate(), transactionRequesterSource(), transactionSourceAccount(), transactionStatus(), transactionType(), transferRateDetails()
  get_transfer_preview:
    primary key: feeId
    fields: feeAmount(), feeId(), feeRateValue(), feeType()
  get_transfer_schedule:
    primary key: transferScheduleId
    fields: paymentPreference(), transactionDestinationAccount(), transactionSourceAccount(), transactionType(), transferScheduleFailures(), transferScheduleId(), transferScheduleInformation(), transferScheduleStatus(), transfer_schedule_id()
  list_transfer_schedules:
    primary key: transferScheduleId
    fields: paymentPreference(), transactionDestinationAccount(), transactionSourceAccount(), transactionType(), transferScheduleFailures(), transferScheduleId(), transferScheduleInformation(), transferScheduleStatus()
  list_financial_events_by_group_id:
    primary key: AmazonOrderId
    fields: AmazonOrderId(), DirectPaymentList(), MarketplaceName(), OrderChargeAdjustmentList(), OrderChargeList(), OrderFeeAdjustmentList(), OrderFeeList(), PostedDate(), SellerOrderId(), ShipmentFeeAdjustmentList(), ShipmentFeeList(), ShipmentItemAdjustmentList(), ShipmentItemList(), StoreName(), event_group_id()
  list_financial_events:
    primary key: AmazonOrderId
    fields: AmazonOrderId(), DirectPaymentList(), MarketplaceName(), OrderChargeAdjustmentList(), OrderChargeList(), OrderFeeAdjustmentList(), OrderFeeList(), PostedDate(), SellerOrderId(), ShipmentFeeAdjustmentList(), ShipmentFeeList(), ShipmentItemAdjustmentList(), ShipmentItemList(), StoreName()
  list_financial_events_by_order_id:
    primary key: order_id
    fields: AmazonOrderId(), DirectPaymentList(), MarketplaceName(), OrderChargeAdjustmentList(), OrderChargeList(), OrderFeeAdjustmentList(), OrderFeeList(), PostedDate(), SellerOrderId(), ShipmentFeeAdjustmentList(), ShipmentFeeList(), ShipmentItemAdjustmentList(), ShipmentItemList(), StoreName(), order_id()
  list_inbound_plan_boxes:
    primary key: packageId
    fields: boxId(), contentInformationSource(), destinationRegion(), dimensions(), externalContainerIdentifier(), externalContainerIdentifierType(), inbound_plan_id(), items(), packageId(), quantity(), templateName(), weight()
  list_inbound_plan_items:
    primary key: asin
    fields: asin(), expiration(), fnsku(), inbound_plan_id(), labelOwner(), manufacturingLotCode(), msku(), prepInstructions(), quantity()
  list_packing_group_boxes:
    primary key: packageId
    fields: boxId(), contentInformationSource(), destinationRegion(), dimensions(), externalContainerIdentifier(), externalContainerIdentifierType(), inbound_plan_id(), items(), packageId(), packing_group_id(), quantity(), templateName(), weight()
  list_packing_group_items:
    primary key: asin
    fields: asin(), expiration(), fnsku(), inbound_plan_id(), labelOwner(), manufacturingLotCode(), msku(), packing_group_id(), prepInstructions(), quantity()
  list_packing_options:
    primary key: inbound_plan_id
    fields: discounts(), expiration(), fees(), inbound_plan_id(), packingGroups(), packingOptionId(), status(), supportedConfigurations(), supportedShippingConfigurations()
  list_inbound_plan_pallets:
    primary key: packageId
    fields: dimensions(), inbound_plan_id(), packageId(), quantity(), stackability(), weight()
  list_placement_options:
    primary key: inbound_plan_id
    fields: discounts(), expiration(), fees(), inbound_plan_id(), placementOptionId(), shipmentIds(), status()
  list_shipment_boxes:
    primary key: shipment_id
    fields: boxId(), contentInformationSource(), destinationRegion(), dimensions(), externalContainerIdentifier(), externalContainerIdentifierType(), inbound_plan_id(), items(), packageId(), quantity(), shipment_id(), templateName(), weight()
  get_shipment_content_update_preview:
    primary key: shipment_id
    fields: contentUpdatePreviewId(), content_update_preview_id(), expiration(), inbound_plan_id(), requestedUpdates(), shipment_id(), transportationOption()
  list_shipment_content_update_previews:
    primary key: shipment_id
    fields: contentUpdatePreviewId(), expiration(), inbound_plan_id(), requestedUpdates(), shipment_id(), transportationOption()
  list_delivery_window_options:
    primary key: shipment_id
    fields: availabilityType(), deliveryWindowOptionId(), endDate(), inbound_plan_id(), shipment_id(), startDate(), validUntil()
  list_shipment_items:
    primary key: shipment_id
    fields: asin(), expiration(), fnsku(), inbound_plan_id(), labelOwner(), manufacturingLotCode(), msku(), prepInstructions(), quantity(), shipment_id()
  list_shipment_pallets:
    primary key: shipment_id
    fields: dimensions(), inbound_plan_id(), packageId(), quantity(), shipment_id(), stackability(), weight()
  get_self_ship_appointment_slots:
    primary key: shipment_id
    fields: inbound_plan_id(), shipment_id(), slotId(), slotTime()
  get_shipment_inbound_fba_2024_03_20_inbound_plans_inbound_plan_id_shipments_shipment_id:
    primary key: shipmentId
    fields: amazonReferenceId(), contactInformation(), dates(), destination(), freightInformation(), inbound_plan_id(), name(), placementOptionId(), selectedDeliveryWindow(), selectedTransportationOptionId(), selfShipAppointmentDetails(), shipmentConfirmationId(), shipmentId(), shipment_id(), source(), status(), trackingDetails()
  list_transportation_options:
    primary key: shipmentId
    fields: carrier(), carrierAppointment(), inbound_plan_id(), preconditions(), quote(), shipmentId(), shippingMode(), shippingSolution(), transportationOptionId()
  get_inbound_plan:
    primary key: inboundPlanId
    fields: createdAt(), inboundPlanId(), inbound_plan_id(), lastUpdatedAt(), marketplaceIds(), name(), packingOptions(), placementOptions(), shipments(), sourceAddress(), status()
  list_inbound_plans:
    primary key: inboundPlanId
    fields: createdAt(), inboundPlanId(), lastUpdatedAt(), marketplaceIds(), name(), sourceAddress(), status()
  list_item_compliance_details:
    primary key: asin
    fields: asin(), fnsku(), msku(), taxDetails()
  list_prep_details:
    fields: allOwnersConstraint(), labelOwnerConstraint(), msku(), prepCategory(), prepOwnerConstraint(), prepTypes()
  get_listings_item:
    primary key: sku
    fields: attributes(), fulfillmentAvailability(), issues(), offers(), procurement(), productTypes(), relationships(), seller_id(), sku(), summaries()
  search_listings_items:
    primary key: sku
    fields: attributes(), fulfillmentAvailability(), issues(), offers(), procurement(), productTypes(), relationships(), seller_id(), sku(), summaries()
  get_listings_restrictions:
    primary key: marketplaceId
    fields: conditionType(), marketplaceId(), reasons()
  get_messaging_actions_for_order:
    primary key: order_id
    fields: _embedded(), _links(), errors(), order_id()
  get_shipment_mfn_v0_shipments_shipment_id:
    primary key: AmazonOrderId
    fields: AmazonOrderId(), CreatedDate(), Insurance(), ItemList(), Label(), LastUpdatedDate(), PackageDimensions(), SellerOrderId(), ShipFromAddress(), ShipToAddress(), ShipmentId(), ShippingService(), Status(), TrackingId(), Weight(), shipment_id()
  get_destination:
    primary key: destinationId
    fields: destinationId(), destination_id(), name(), resource()
  get_destinations:
    primary key: destinationId
    fields: destinationId(), name(), resource()
  get_subscription_by_id:
    primary key: destinationId
    fields: destinationId(), notification_type(), payloadVersion(), processingDirective(), subscriptionId(), subscription_id()
  get_subscription:
    primary key: destinationId
    fields: destinationId(), notification_type(), payloadVersion(), processingDirective(), subscriptionId()
  get_order:
    primary key: orderId
    fields: associatedOrders(), buyer(), createdTime(), fulfillment(), fulfillmentOrders(), lastUpdatedTime(), orderAliases(), orderId(), orderItems(), order_id(), packages(), payment(), proceeds(), programs(), recipient(), salesChannel(), tax()
  search_orders:
    primary key: orderId
    fields: associatedOrders(), buyer(), createdTime(), fulfillment(), fulfillmentOrders(), lastUpdatedTime(), orderAliases(), orderId(), orderItems(), packages(), payment(), proceeds(), programs(), recipient(), salesChannel(), tax()
  get_order_items:
    primary key: order_id
    fields: ASIN(), AmazonPrograms(), AssociatedItems(), BuyerInfo(), BuyerRequestedCancel(), CODFee(), CODFeeDiscount(), ConditionId(), ConditionNote(), ConditionSubtypeId(), DeemedResellerCategory(), IossNumber(), IsGift(), IsTransparency(), ItemPrice(), ItemTax(), Measurement(), OrderItemId(), PointsGranted(), PriceDesignation(), ProductInfo(), PromotionDiscount(), PromotionDiscountTax(), PromotionIds(), QuantityOrdered(), QuantityShipped(), ScheduledDeliveryEndDate(), ScheduledDeliveryStartDate(), SellerSKU(), SerialNumberRequired(), SerialNumbers(), ShippingConstraints(), ShippingDiscount(), ShippingDiscountTax(), ShippingPrice(), ShippingTax(), StoreChainStoreId(), SubstitutionPreferences(), TaxCollection(), Title(), order_id()
  get_order_orders_v0_orders_order_id:
    primary key: order_id
    fields: AmazonOrderId(), AutomatedShippingSettings(), BuyerInfo(), BuyerInvoicePreference(), BuyerTaxInformation(), CbaDisplayableShippingLabel(), DefaultShipFromLocationAddress(), EarliestDeliveryDate(), EarliestShipDate(), EasyShipShipmentStatus(), ElectronicInvoiceStatus(), FulfillmentChannel(), FulfillmentInstruction(), HasRegulatedItems(), IsAccessPointOrder(), IsBusinessOrder(), IsEstimatedShipDateSet(), IsGlobalExpressEnabled(), IsIBA(), IsISPU(), IsPremiumOrder(), IsPrime(), IsReplacementOrder(), IsSoldByAB(), LastUpdateDate(), LatestDeliveryDate(), LatestShipDate(), MarketplaceId(), MarketplaceTaxInfo(), NumberOfItemsShipped(), NumberOfItemsUnshipped(), OrderChannel(), OrderStatus(), OrderTotal(), OrderType(), PaymentExecutionDetail(), PaymentMethod(), PaymentMethodDetails(), PromiseResponseDueDate(), PurchaseDate(), ReplacedOrderId(), SalesChannel(), SellerDisplayName(), SellerOrderId(), ShipServiceLevel(), ShipmentServiceLevelCategory(), ShippingAddress(), order_id()
  get_competitive_pricing:
    primary key: ASIN
    fields: ASIN(), Product(), SellerSKU(), status()
  get_item_offers:
    primary key: asin
    fields: ConditionNotes(), IsBuyBoxWinner(), IsFeaturedMerchant(), IsFulfilledByAmazon(), ListingPrice(), MyOffer(), Points(), PrimeInformation(), SellerFeedbackRating(), SellerId(), Shipping(), ShippingTime(), ShipsFrom(), SubCondition(), asin(), offerType(), quantityDiscountPrices()
  get_listing_offers:
    primary key: seller_sku
    fields: ConditionNotes(), IsBuyBoxWinner(), IsFeaturedMerchant(), IsFulfilledByAmazon(), ListingPrice(), MyOffer(), Points(), PrimeInformation(), SellerFeedbackRating(), SellerId(), Shipping(), ShippingTime(), ShipsFrom(), SubCondition(), offerType(), quantityDiscountPrices(), seller_sku()
  get_pricing:
    primary key: ASIN
    fields: ASIN(), Product(), SellerSKU(), status()
  get_report:
    primary key: reportId
    fields: createdTime(), dataEndTime(), dataStartTime(), marketplaceIds(), processingEndTime(), processingStartTime(), processingStatus(), reportDocumentId(), reportId(), reportScheduleId(), reportType(), report_id()
  get_reports:
    primary key: reportId
    fields: createdTime(), dataEndTime(), dataStartTime(), marketplaceIds(), processingEndTime(), processingStartTime(), processingStatus(), reportDocumentId(), reportId(), reportScheduleId(), reportType()
  get_report_schedule:
    primary key: report_schedule_id
    fields: marketplaceIds(), nextReportCreationTime(), period(), reportOptions(), reportScheduleId(), reportType(), report_schedule_id()
  get_report_schedules:
    primary key: reportScheduleId
    fields: marketplaceIds(), nextReportCreationTime(), period(), reportOptions(), reportScheduleId(), reportType()
  get_order_metrics:
    fields: averageUnitPrice(), interval(), orderCount(), orderItemCount(), totalSales(), unitCount()
  get_account_sellers_v1_account:
    fields: marketplace(), participation(), storeName()
  get_marketplace_participations:
    fields: marketplace(), participation(), storeName()
  get_appointment_slots:
    fields: capacity(), endTime(), startTime()
  get_appointmment_slots_by_job_id:
    primary key: service_job_id
    fields: capacity(), endTime(), service_job_id(), startTime()
  get_service_job_by_service_job_id:
    primary key: serviceJobId
    fields: appointments(), associatedItems(), buyer(), createTime(), marketplaceId(), payments(), preferredAppointmentTimes(), productOrderIds(), scopeOfWork(), seller(), serviceJobId(), serviceJobProvider(), serviceJobStatus(), serviceLocation(), serviceOrderId(), service_job_id(), storeId(), trackingIds()
  get_service_jobs:
    primary key: serviceJobId
    fields: appointments(), associatedItems(), buyer(), createTime(), marketplaceId(), payments(), preferredAppointmentTimes(), productOrderIds(), scopeOfWork(), seller(), serviceJobId(), serviceJobProvider(), serviceJobStatus(), serviceLocation(), serviceOrderId(), storeId(), trackingIds()
  get_account_shipping_v1_account:
    primary key: accountId
    fields: accountId()
  get_shipment_shipping_v1_shipments_shipment_id:
    primary key: shipmentId
    fields: acceptedRate(), clientReferenceId(), containers(), shipFrom(), shipTo(), shipmentId(), shipment_id(), shipper()
  get_tracking_information:
    primary key: trackingId
    fields: eventHistory(), promisedDeliveryDate(), summary(), trackingId(), tracking_id()
  get_access_points:
    fields: accessPointsMap()
  get_carrier_account_form_inputs:
    primary key: carrierId
    fields: carrierId(), linkableAccountTypes()
  get_collection_form:
    primary key: collection_form_id
    fields: base64EncodedContent(), collection_form_id(), documentFormat()
  get_shipment_documents:
    primary key: shipment_id
    fields: contents(), format(), shipment_id(), type()
  get_additional_inputs:
  get_tracking:
  get_solicitation_actions_for_order:
    primary key: order_id
    fields: _embedded(), _links(), errors(), order_id()
  get_supply_source:
    primary key: supplySourceId
    fields: address(), alias(), capabilities(), configuration(), createdAt(), status(), supplySourceCode(), supplySourceId(), supply_source_id(), updatedAt()
  get_supply_sources:
    primary key: supplySourceId
    fields: address(), alias(), supplySourceCode(), supplySourceId()
  get_invoices_export:
    primary key: exportId
    fields: errorMessage(), exportId(), export_id(), generateExportFinishedAt(), generateExportStartedAt(), invoicesDocumentIds(), status()
  get_invoices_exports:
    primary key: exportId
    fields: errorMessage(), exportId(), generateExportFinishedAt(), generateExportStartedAt(), invoicesDocumentIds(), status()
  get_invoice:
    primary key: id
    fields: date(), errorCode(), externalInvoiceId(), govResponse(), id(), invoiceType(), invoice_id(), series(), status(), transactionIds(), transactionType()
  get_invoices:
    primary key: id
    fields: date(), errorCode(), externalInvoiceId(), govResponse(), id(), invoiceType(), series(), status(), transactionIds(), transactionType()
  get_order_vendor_direct_fulfillment_orders_2021_12_28_purchase_orders_purchase_order_numbe:
    primary key: purchaseOrderNumber
    fields: orderDetails(), purchaseOrderNumber(), purchase_order_number()
  get_orders:
    primary key: purchaseOrderNumber
    fields: orderDetails(), purchaseOrderNumber()
  get_order_vendor_direct_fulfillment_orders_v1_purchase_orders_purchase_order_number:
    primary key: purchaseOrderNumber
    fields: orderDetails(), purchaseOrderNumber(), purchase_order_number()
  get_orders_vendor_direct_fulfillment_orders_v1_purchase_orders:
    primary key: purchaseOrderNumber
    fields: orderDetails(), purchaseOrderNumber()
  get_order_scenarios:
    primary key: transactionId
    fields: status(), testCaseData(), transactionId(), transaction_id()
  get_customer_invoice:
    primary key: purchaseOrderNumber
    fields: content(), purchaseOrderNumber(), purchase_order_number()
  get_customer_invoices:
    primary key: purchaseOrderNumber
    fields: content(), purchaseOrderNumber()
  get_customer_invoice_vendor_direct_fulfillment_shipping_v1_customer_invoices_purchase_order_number:
    primary key: purchaseOrderNumber
    fields: content(), purchaseOrderNumber(), purchase_order_number()
  get_customer_invoices_vendor_direct_fulfillment_shipping_v1_customer_invoices:
    primary key: purchaseOrderNumber
    fields: content(), purchaseOrderNumber()
  get_transaction_status:
    primary key: transactionId
    fields: errors(), status(), transactionId(), transaction_id()
  get_transaction_status_vendor_direct_fulfillment_transactions_v1_transactions_transaction_id:
    primary key: transaction_id
    fields: transactionStatus(), transaction_id()
  get_purchase_order:
    primary key: purchaseOrderNumber
    fields: orderDetails(), purchaseOrderNumber(), purchaseOrderState(), purchase_order_number()
  get_purchase_orders:
    primary key: purchaseOrderNumber
    fields: orderDetails(), purchaseOrderNumber(), purchaseOrderState()
  get_purchase_orders_status:
    primary key: purchaseOrderNumber
    fields: itemStatus(), lastUpdatedDate(), purchaseOrderDate(), purchaseOrderNumber(), purchaseOrderStatus(), sellingParty(), shipToParty()
  get_shipment_details_vendor_shipping_v1_shipments:
    primary key: buyerReferenceNumber
    fields: buyerReferenceNumber(), collectFreightPickupDetails(), containers(), currentShipmentStatus(), currentshipmentStatusDate(), importDetails(), packageLabelCreateDate(), purchaseOrders(), sellingParty(), shipFromParty(), shipToParty(), shipmentConfirmDate(), shipmentCreateDate(), shipmentFreightTerm(), shipmentMeasurements(), shipmentStatusDetails(), transactionDate(), transactionType(), transportationDetails(), vendorShipmentIdentifier()
  get_transaction_vendor_transactions_v1_transactions_transaction_id:
    primary key: transaction_id
    fields: transactionStatus(), transaction_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  record_action_feedback:
    endpoint: POST /appIntegrations/2024-04-01/notifications/{{ record.notification_id }}/feedback
    required fields: notification_id
    risk: POST /appIntegrations/2024-04-01/notifications/{notificationId}/feedback against the live SP-API account
  delete_notifications:
    endpoint: POST /appIntegrations/2024-04-01/notifications/deletion
    risk: POST /appIntegrations/2024-04-01/notifications/deletion against the live SP-API account
  create_notification:
    endpoint: POST /appIntegrations/2024-04-01/notifications
    risk: POST /appIntegrations/2024-04-01/notifications against the live SP-API account
  cancel_inbound:
    endpoint: POST /awd/2024-05-09/inboundOrders/{{ record.order_id }}/cancellation
    required fields: order_id
    risk: POST /awd/2024-05-09/inboundOrders/{orderId}/cancellation against the live SP-API account
  confirm_inbound:
    endpoint: POST /awd/2024-05-09/inboundOrders/{{ record.order_id }}/confirmation
    required fields: order_id
    risk: POST /awd/2024-05-09/inboundOrders/{orderId}/confirmation against the live SP-API account
  update_inbound:
    endpoint: PUT /awd/2024-05-09/inboundOrders/{{ record.order_id }}
    required fields: order_id
    risk: PUT /awd/2024-05-09/inboundOrders/{orderId} against the live SP-API account
  create_inbound:
    endpoint: POST /awd/2024-05-09/inboundOrders
    risk: POST /awd/2024-05-09/inboundOrders against the live SP-API account
  update_inbound_shipment_transport_details:
    endpoint: PUT /awd/2024-05-09/inboundShipments/{{ record.shipment_id }}/transport
    required fields: shipment_id
    risk: PUT /awd/2024-05-09/inboundShipments/{shipmentId}/transport against the live SP-API account
  confirm_outbound:
    endpoint: POST /awd/2024-05-09/outboundOrders/{{ record.order_id }}/confirmation
    required fields: order_id
    risk: POST /awd/2024-05-09/outboundOrders/{orderId}/confirmation against the live SP-API account
  update_outbound:
    endpoint: PUT /awd/2024-05-09/outboundOrders/{{ record.order_id }}
    required fields: order_id
    risk: PUT /awd/2024-05-09/outboundOrders/{orderId} against the live SP-API account
  create_outbound:
    endpoint: POST /awd/2024-05-09/outboundOrders
    risk: POST /awd/2024-05-09/outboundOrders against the live SP-API account
  confirm_replenishment_order:
    endpoint: POST /awd/2024-05-09/replenishmentOrders/{{ record.order_id }}/confirmation
    required fields: order_id
    risk: POST /awd/2024-05-09/replenishmentOrders/{orderId}/confirmation against the live SP-API account
  create_replenishment_order:
    endpoint: POST /awd/2024-05-09/replenishmentOrders
    risk: POST /awd/2024-05-09/replenishmentOrders against the live SP-API account
  cancel_query:
    endpoint: DELETE /dataKiosk/2023-11-15/queries/{{ record.query_id }}
    required fields: query_id
    risk: DELETE /dataKiosk/2023-11-15/queries/{queryId} against the live SP-API account
  create_query:
    endpoint: POST /dataKiosk/2023-11-15/queries
    risk: POST /dataKiosk/2023-11-15/queries against the live SP-API account
  update_scheduled_packages:
    endpoint: PATCH /easyShip/2022-03-23/package
    risk: PATCH /easyShip/2022-03-23/package against the live SP-API account
  create_scheduled_package:
    endpoint: POST /easyShip/2022-03-23/package
    risk: POST /easyShip/2022-03-23/package against the live SP-API account
  create_scheduled_package_bulk:
    endpoint: POST /easyShip/2022-03-23/packages/bulk
    risk: POST /easyShip/2022-03-23/packages/bulk against the live SP-API account
  update_package_status:
    endpoint: PATCH /externalFulfillment/2024-09-11/shipments/{{ record.shipment_id }}/packages/{{ record.package_id }}
    required fields: shipment_id, package_id
    risk: PATCH /externalFulfillment/2024-09-11/shipments/{shipmentId}/packages/{packageId} against the live SP-API account
  update_package:
    endpoint: PUT /externalFulfillment/2024-09-11/shipments/{{ record.shipment_id }}/packages/{{ record.package_id }}
    required fields: shipment_id, package_id
    risk: PUT /externalFulfillment/2024-09-11/shipments/{shipmentId}/packages/{packageId} against the live SP-API account
  create_packages:
    endpoint: POST /externalFulfillment/2024-09-11/shipments/{{ record.shipment_id }}/packages
    required fields: shipment_id
    risk: POST /externalFulfillment/2024-09-11/shipments/{shipmentId}/packages against the live SP-API account
  batch_inventory:
    endpoint: POST /externalFulfillment/inventory/2024-09-11/inventories
    risk: POST /externalFulfillment/inventory/2024-09-11/inventories against the live SP-API account
  add_inventory:
    endpoint: POST /fba/inventory/v1/items/inventory
    risk: POST /fba/inventory/v1/items/inventory against the live SP-API account
  create_inventory_item:
    endpoint: POST /fba/inventory/v1/items
    risk: POST /fba/inventory/v1/items against the live SP-API account
  delivery_offers:
    endpoint: POST /fba/outbound/2020-07-01/deliveryOffers
    risk: POST /fba/outbound/2020-07-01/deliveryOffers against the live SP-API account
  cancel_fulfillment_order:
    endpoint: PUT /fba/outbound/2020-07-01/fulfillmentOrders/{{ record.seller_fulfillment_order_id }}/cancel
    required fields: seller_fulfillment_order_id
    risk: PUT /fba/outbound/2020-07-01/fulfillmentOrders/{sellerFulfillmentOrderId}/cancel against the live SP-API account
  create_fulfillment_return:
    endpoint: PUT /fba/outbound/2020-07-01/fulfillmentOrders/{{ record.seller_fulfillment_order_id }}/return
    required fields: seller_fulfillment_order_id
    risk: PUT /fba/outbound/2020-07-01/fulfillmentOrders/{sellerFulfillmentOrderId}/return against the live SP-API account
  submit_fulfillment_order_status_update:
    endpoint: PUT /fba/outbound/2020-07-01/fulfillmentOrders/{{ record.seller_fulfillment_order_id }}/status
    required fields: seller_fulfillment_order_id
    risk: PUT /fba/outbound/2020-07-01/fulfillmentOrders/{sellerFulfillmentOrderId}/status against the live SP-API account
  update_fulfillment_order:
    endpoint: PUT /fba/outbound/2020-07-01/fulfillmentOrders/{{ record.seller_fulfillment_order_id }}
    required fields: seller_fulfillment_order_id
    risk: PUT /fba/outbound/2020-07-01/fulfillmentOrders/{sellerFulfillmentOrderId} against the live SP-API account
  create_fulfillment_order:
    endpoint: POST /fba/outbound/2020-07-01/fulfillmentOrders
    risk: POST /fba/outbound/2020-07-01/fulfillmentOrders against the live SP-API account
  cancel_feed:
    endpoint: DELETE /feeds/2021-06-30/feeds/{{ record.feed_id }}
    required fields: feed_id
    risk: DELETE /feeds/2021-06-30/feeds/{feedId} against the live SP-API account
  create_feed:
    endpoint: POST /feeds/2021-06-30/feeds
    risk: POST /feeds/2021-06-30/feeds against the live SP-API account
  initiate_payout:
    endpoint: POST /finances/transfers/2024-06-01/payouts
    risk: POST /finances/transfers/2024-06-01/payouts against the live SP-API account
  cancel_inbound_plan:
    endpoint: PUT /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/cancellation
    required fields: inbound_plan_id
    risk: PUT /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/cancellation against the live SP-API account
  update_inbound_plan_name:
    endpoint: PUT /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/name
    required fields: inbound_plan_id
    risk: PUT /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/name against the live SP-API account
  set_packing_information:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/packingInformation
    required fields: inbound_plan_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/packingInformation against the live SP-API account
  confirm_packing_option:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/packingOptions/{{ record.packing_option_id }}/confirmation
    required fields: inbound_plan_id, packing_option_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/packingOptions/{packingOptionId}/confirmation against the live SP-API account
  generate_packing_options:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/packingOptions
    required fields: inbound_plan_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/packingOptions against the live SP-API account
  confirm_placement_option:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/placementOptions/{{ record.placement_option_id }}/confirmation
    required fields: inbound_plan_id, placement_option_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/placementOptions/{placementOptionId}/confirmation against the live SP-API account
  generate_placement_options:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/placementOptions
    required fields: inbound_plan_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/placementOptions against the live SP-API account
  confirm_delivery_window_options:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/deliveryWindowOptions/{{ record.delivery_window_option_id }}/confirmation
    required fields: inbound_plan_id, shipment_id, delivery_window_option_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/deliveryWindowOptions/{deliveryWindowOptionId}/confirmation against the live SP-API account
  generate_delivery_window_options:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/deliveryWindowOptions
    required fields: inbound_plan_id, shipment_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/deliveryWindowOptions against the live SP-API account
  update_shipment_name:
    endpoint: PUT /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/name
    required fields: inbound_plan_id, shipment_id
    risk: PUT /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/name against the live SP-API account
  cancel_self_ship_appointment:
    endpoint: PUT /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/selfShipAppointmentCancellation
    required fields: inbound_plan_id, shipment_id
    risk: PUT /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/selfShipAppointmentCancellation against the live SP-API account
  schedule_self_ship_appointment:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/selfShipAppointmentSlots/{{ record.slot_id }}/schedule
    required fields: inbound_plan_id, shipment_id, slot_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/selfShipAppointmentSlots/{slotId}/schedule against the live SP-API account
  update_shipment_source_address:
    endpoint: PUT /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/sourceAddress
    required fields: inbound_plan_id, shipment_id
    risk: PUT /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/sourceAddress against the live SP-API account
  update_shipment_tracking_details:
    endpoint: PUT /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/shipments/{{ record.shipment_id }}/trackingDetails
    required fields: inbound_plan_id, shipment_id
    risk: PUT /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/shipments/{shipmentId}/trackingDetails against the live SP-API account
  confirm_transportation_options:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/transportationOptions/confirmation
    required fields: inbound_plan_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/transportationOptions/confirmation against the live SP-API account
  generate_transportation_options:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans/{{ record.inbound_plan_id }}/transportationOptions
    required fields: inbound_plan_id
    risk: POST /inbound/fba/2024-03-20/inboundPlans/{inboundPlanId}/transportationOptions against the live SP-API account
  create_inbound_plan:
    endpoint: POST /inbound/fba/2024-03-20/inboundPlans
    risk: POST /inbound/fba/2024-03-20/inboundPlans against the live SP-API account
  set_prep_details:
    endpoint: POST /inbound/fba/2024-03-20/items/prepDetails
    risk: POST /inbound/fba/2024-03-20/items/prepDetails against the live SP-API account
  cancel_shipment:
    endpoint: DELETE /mfn/v0/shipments/{{ record.shipment_id }}
    required fields: shipment_id
    risk: DELETE /mfn/v0/shipments/{shipmentId} against the live SP-API account
  create_shipment:
    endpoint: POST /mfn/v0/shipments
    risk: POST /mfn/v0/shipments against the live SP-API account
  delete_destination:
    endpoint: DELETE /notifications/v1/destinations/{{ record.destination_id }}
    required fields: destination_id
    risk: DELETE /notifications/v1/destinations/{destinationId} against the live SP-API account
  create_destination:
    endpoint: POST /notifications/v1/destinations
    risk: POST /notifications/v1/destinations against the live SP-API account
  delete_subscription_by_id:
    endpoint: DELETE /notifications/v1/subscriptions/{{ record.notification_type }}/{{ record.subscription_id }}
    required fields: notification_type, subscription_id
    risk: DELETE /notifications/v1/subscriptions/{notificationType}/{subscriptionId} against the live SP-API account
  send_test_notification:
    endpoint: POST /notifications/v1/subscriptions/{{ record.notification_type }}/testNotification
    required fields: notification_type
    risk: POST /notifications/v1/subscriptions/{notificationType}/testNotification against the live SP-API account
  create_subscription:
    endpoint: POST /notifications/v1/subscriptions/{{ record.notification_type }}
    required fields: notification_type
    risk: POST /notifications/v1/subscriptions/{notificationType} against the live SP-API account
  update_verification_status:
    endpoint: PATCH /orders/v0/orders/{{ record.order_id }}/regulatedInfo
    required fields: order_id
    risk: PATCH /orders/v0/orders/{orderId}/regulatedInfo against the live SP-API account
  confirm_shipment:
    endpoint: POST /orders/v0/orders/{{ record.order_id }}/shipmentConfirmation
    required fields: order_id
    risk: POST /orders/v0/orders/{orderId}/shipmentConfirmation against the live SP-API account
  update_shipment_status:
    endpoint: POST /orders/v0/orders/{{ record.order_id }}/shipment
    required fields: order_id
    risk: POST /orders/v0/orders/{orderId}/shipment against the live SP-API account
  cancel_report:
    endpoint: DELETE /reports/2021-06-30/reports/{{ record.report_id }}
    required fields: report_id
    risk: DELETE /reports/2021-06-30/reports/{reportId} against the live SP-API account
  create_report:
    endpoint: POST /reports/2021-06-30/reports
    risk: POST /reports/2021-06-30/reports against the live SP-API account
  cancel_report_schedule:
    endpoint: DELETE /reports/2021-06-30/schedules/{{ record.report_schedule_id }}
    required fields: report_schedule_id
    risk: DELETE /reports/2021-06-30/schedules/{reportScheduleId} against the live SP-API account
  create_report_schedule:
    endpoint: POST /reports/2021-06-30/schedules
    risk: POST /reports/2021-06-30/schedules against the live SP-API account
  set_appointment_fulfillment_data:
    endpoint: PUT /service/v1/serviceJobs/{{ record.service_job_id }}/appointments/{{ record.appointment_id }}/fulfillment
    required fields: service_job_id, appointment_id
    risk: PUT /service/v1/serviceJobs/{serviceJobId}/appointments/{appointmentId}/fulfillment against the live SP-API account
  assign_appointment_resources:
    endpoint: PUT /service/v1/serviceJobs/{{ record.service_job_id }}/appointments/{{ record.appointment_id }}/resources
    required fields: service_job_id, appointment_id
    risk: PUT /service/v1/serviceJobs/{serviceJobId}/appointments/{appointmentId}/resources against the live SP-API account
  reschedule_appointment_for_service_job_by_service_job_id:
    endpoint: POST /service/v1/serviceJobs/{{ record.service_job_id }}/appointments/{{ record.appointment_id }}
    required fields: service_job_id, appointment_id
    risk: POST /service/v1/serviceJobs/{serviceJobId}/appointments/{appointmentId} against the live SP-API account
  add_appointment_for_service_job_by_service_job_id:
    endpoint: POST /service/v1/serviceJobs/{{ record.service_job_id }}/appointments
    required fields: service_job_id
    risk: POST /service/v1/serviceJobs/{serviceJobId}/appointments against the live SP-API account
  complete_service_job_by_service_job_id:
    endpoint: PUT /service/v1/serviceJobs/{{ record.service_job_id }}/completions
    required fields: service_job_id
    risk: PUT /service/v1/serviceJobs/{serviceJobId}/completions against the live SP-API account
  purchase_shipment:
    endpoint: POST /shipping/v1/purchaseShipment
    risk: POST /shipping/v1/purchaseShipment against the live SP-API account
  cancel_shipment_post_shipping_v1_shipments_shipment_id_cancel:
    endpoint: POST /shipping/v1/shipments/{{ record.shipment_id }}/cancel
    required fields: shipment_id
    risk: POST /shipping/v1/shipments/{shipmentId}/cancel against the live SP-API account
  create_shipment_post_shipping_v1_shipments:
    endpoint: POST /shipping/v1/shipments
    risk: POST /shipping/v1/shipments against the live SP-API account
  unlink_carrier_account:
    endpoint: PUT /shipping/v2/carrierAccounts/{{ record.carrier_id }}/unlink
    required fields: carrier_id
    risk: PUT /shipping/v2/carrierAccounts/{carrierId}/unlink against the live SP-API account
  link_carrier_account:
    endpoint: POST /shipping/v2/carrierAccounts/{{ record.carrier_id }}
    required fields: carrier_id
    risk: POST /shipping/v2/carrierAccounts/{carrierId} against the live SP-API account
  link_carrier_account_put_shipping_v2_carrier_accounts_carrier_id:
    endpoint: PUT /shipping/v2/carrierAccounts/{{ record.carrier_id }}
    required fields: carrier_id
    risk: PUT /shipping/v2/carrierAccounts/{carrierId} against the live SP-API account
  create_claim:
    endpoint: POST /shipping/v2/claims
    risk: POST /shipping/v2/claims against the live SP-API account
  generate_collection_form:
    endpoint: POST /shipping/v2/collectionForms
    risk: POST /shipping/v2/collectionForms against the live SP-API account
  submit_ndr_feedback:
    endpoint: POST /shipping/v2/ndrFeedback
    risk: POST /shipping/v2/ndrFeedback against the live SP-API account
  one_click_shipment:
    endpoint: POST /shipping/v2/oneClickShipment
    risk: POST /shipping/v2/oneClickShipment against the live SP-API account
  cancel_shipment_put_shipping_v2_shipments_shipment_id_cancel:
    endpoint: PUT /shipping/v2/shipments/{{ record.shipment_id }}/cancel
    required fields: shipment_id
    risk: PUT /shipping/v2/shipments/{shipmentId}/cancel against the live SP-API account
  direct_purchase_shipment:
    endpoint: POST /shipping/v2/shipments/directPurchase
    risk: POST /shipping/v2/shipments/directPurchase against the live SP-API account
  purchase_shipment_post_shipping_v2_shipments:
    endpoint: POST /shipping/v2/shipments
    risk: POST /shipping/v2/shipments against the live SP-API account
  update_supply_source_status:
    endpoint: PUT /supplySources/2020-07-01/supplySources/{{ record.supply_source_id }}/status
    required fields: supply_source_id
    risk: PUT /supplySources/2020-07-01/supplySources/{supplySourceId}/status against the live SP-API account
  archive_supply_source:
    endpoint: DELETE /supplySources/2020-07-01/supplySources/{{ record.supply_source_id }}
    required fields: supply_source_id
    risk: DELETE /supplySources/2020-07-01/supplySources/{supplySourceId} against the live SP-API account
  update_supply_source:
    endpoint: PUT /supplySources/2020-07-01/supplySources/{{ record.supply_source_id }}
    required fields: supply_source_id
    risk: PUT /supplySources/2020-07-01/supplySources/{supplySourceId} against the live SP-API account
  create_supply_source:
    endpoint: POST /supplySources/2020-07-01/supplySources
    risk: POST /supplySources/2020-07-01/supplySources against the live SP-API account
  submit_inventory_update:
    endpoint: POST /vendor/directFulfillment/inventory/v1/warehouses/{{ record.warehouse_id }}/items
    required fields: warehouse_id
    risk: POST /vendor/directFulfillment/inventory/v1/warehouses/{warehouseId}/items against the live SP-API account
  submit_acknowledgement:
    endpoint: POST /vendor/directFulfillment/orders/2021-12-28/acknowledgements
    risk: POST /vendor/directFulfillment/orders/2021-12-28/acknowledgements against the live SP-API account
  submit_acknowledgement_post_vendor_direct_fulfillment_orders_v1_acknowledgements:
    endpoint: POST /vendor/directFulfillment/orders/v1/acknowledgements
    risk: POST /vendor/directFulfillment/orders/v1/acknowledgements against the live SP-API account
  generate_order_scenarios:
    endpoint: POST /vendor/directFulfillment/sandbox/2021-10-28/orders
    risk: POST /vendor/directFulfillment/sandbox/2021-10-28/orders against the live SP-API account
  submit_shipment_confirmations:
    endpoint: POST /vendor/directFulfillment/shipping/2021-12-28/shipmentConfirmations
    risk: POST /vendor/directFulfillment/shipping/2021-12-28/shipmentConfirmations against the live SP-API account
  submit_shipment_status_updates:
    endpoint: POST /vendor/directFulfillment/shipping/2021-12-28/shipmentStatusUpdates
    risk: POST /vendor/directFulfillment/shipping/2021-12-28/shipmentStatusUpdates against the live SP-API account
  submit_shipment_confirmations_post_vendor_direct_fulfillment_shipping_v1_shipment_confirmations:
    endpoint: POST /vendor/directFulfillment/shipping/v1/shipmentConfirmations
    risk: POST /vendor/directFulfillment/shipping/v1/shipmentConfirmations against the live SP-API account
  submit_shipment_status_updates_post_vendor_direct_fulfillment_shipping_v1_shipment_status_updates:
    endpoint: POST /vendor/directFulfillment/shipping/v1/shipmentStatusUpdates
    risk: POST /vendor/directFulfillment/shipping/v1/shipmentStatusUpdates against the live SP-API account
  submit_acknowledgement_post_vendor_orders_v1_acknowledgements:
    endpoint: POST /vendor/orders/v1/acknowledgements
    risk: POST /vendor/orders/v1/acknowledgements against the live SP-API account
  submit_shipment_confirmations_post_vendor_shipping_v1_shipment_confirmations:
    endpoint: POST /vendor/shipping/v1/shipmentConfirmations
    risk: POST /vendor/shipping/v1/shipmentConfirmations against the live SP-API account
  submit_shipments:
    endpoint: POST /vendor/shipping/v1/shipments
    risk: POST /vendor/shipping/v1/shipments against the live SP-API account

SECURITY
  read risk: external Amazon Selling Partner API read of order, inventory, and financial data
  write risk: external Amazon Selling Partner API mutations such as feed/report creation, fulfillment, inventory, order, shipping, notification, and vendor workflow actions
  approval: reverse ETL writes require plan preview and approval token; destructive/cancel/delete actions are marked destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect amazon-seller-partner

  # Inspect as structured JSON
  pm connectors inspect amazon-seller-partner --json

AGENT WORKFLOW
  - Run pm connectors inspect amazon-seller-partner before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
