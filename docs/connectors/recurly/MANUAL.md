# pm connectors inspect recurly

```text
NAME
  pm connectors inspect recurly - Recurly connector manual

SYNOPSIS
  pm connectors inspect recurly
  pm connectors inspect recurly --json
  pm credentials add <name> --connector recurly [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Recurly accounts, subscriptions, invoices, transactions, catalog, usage, exports, preview resources, and related V3 API data; models typed reverse-ETL mutations for official POST/PUT/DELETE endpoints.

ICON
  id: recurly
  asset: icons/recurly.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.recurly.com/api/v2021-02-25/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  account_id
  account_note_id
  add_on_id
  base_url
  billing_info_id
  business_entity_id
  coupon_id
  coupon_redemption_id
  credit_payment_id
  custom_field_definition_id
  dunning_campaign_id
  export_date
  external_account_id
  external_invoice_id
  external_payment_phase_id
  external_product_id
  external_product_reference_id
  external_subscription_id
  general_ledger_account_id
  gift_card_id
  invoice_id
  invoice_template_id
  item_id
  line_item_id
  measured_unit_id
  performance_obligation_id
  plan_id
  price_segment_id
  redemption_code
  shipping_address_id
  shipping_method_id
  site_id
  subscription_id
  transaction_id
  unique_coupon_code_id
  usage_id
  api_key (secret) (required)

ETL STREAMS
  list_sites:
    primary key: id
    fields: address(object), created_at(string), deleted_at(string), features(array), id(string), mode(string), object(string), public_api_key(string), settings(object), subdomain(string), updated_at(string)
  get_site:
    primary key: id
    fields: address(object), created_at(string), deleted_at(string), features(array), id(string), mode(string), object(string), public_api_key(string), settings(object), subdomain(string), updated_at(string)
  accounts:
    primary key: id
    cursor: updated_at
    fields: code(string), created_at(string), email(string), id(string), state(string), updated_at(string)
  get_account:
    primary key: id
    fields: address(object), bill_date(string), bill_to(string), billing_info(object), cc_emails(string), code(string), company(string), created_at(string), custom_fields(array), deleted_at(string), dunning_campaign_id(string), email(string), entity_use_code(string), exemption_certificate(string), external_accounts(array), first_name(string), has_active_subscription(boolean), has_canceled_subscription(boolean), has_future_subscription(boolean), has_live_subscription(boolean), has_past_due_invoice(boolean), has_paused_subscription(boolean), hosted_login_token(string), id(string), invoice_template_id(string), last_name(string), object(string), override_business_entity_id(string), parent_account_id(string), preferred_locale(string), preferred_time_zone(string), shipping_addresses(array), state(string), tax_exempt(boolean), updated_at(string), username(string), vat_number(string)
  get_account_acquisition:
    primary key: id
    fields: account(object), acquired_at(string), campaign(string), channel(string), cost(object), created_at(string), id(string), object(string), subchannel(string), updated_at(string)
  get_account_balance:
    primary key: account_id
    fields: account(object), account_id(string), balances(array), object(string), past_due(boolean)
  get_billing_info:
    primary key: id
    fields: account_id(string), address(object), backup_payment_method(boolean), company(string), created_at(string), first_name(string), fraud(object), id(string), last_name(string), object(string), payment_gateway_references(array), payment_method(object), primary_payment_method(boolean), updated_at(string), updated_by(object), valid(boolean), vat_number(string)
  list_billing_infos:
    primary key: id
    fields: account_id(string), address(object), backup_payment_method(boolean), company(string), created_at(string), first_name(string), fraud(object), id(string), last_name(string), object(string), payment_gateway_references(array), payment_method(object), primary_payment_method(boolean), updated_at(string), updated_by(object), valid(boolean), vat_number(string)
  get_a_billing_info:
    primary key: id
    fields: account_id(string), address(object), backup_payment_method(boolean), company(string), created_at(string), first_name(string), fraud(object), id(string), last_name(string), object(string), payment_gateway_references(array), payment_method(object), primary_payment_method(boolean), updated_at(string), updated_by(object), valid(boolean), vat_number(string)
  list_account_coupon_redemptions:
    primary key: id
    fields: account(object), coupon(object), created_at(string), currency(string), discounted(number), id(string), object(string), remaining_duration(object), removed_at(string), state(string), subscription_id(string), updated_at(string), uuid(string)
  list_active_coupon_redemptions:
    primary key: id
    fields: account(object), coupon(object), created_at(string), currency(string), discounted(number), id(string), object(string), remaining_duration(object), removed_at(string), state(string), subscription_id(string), updated_at(string), uuid(string)
  get_coupon_redemption:
    primary key: id
    fields: account(object), coupon(object), created_at(string), currency(string), discounted(number), id(string), object(string), remaining_duration(object), removed_at(string), state(string), subscription_id(string), updated_at(string), uuid(string)
  list_account_credit_payments:
    primary key: id
    fields: account(object), action(string), amount(number), applied_to_invoice(object), created_at(string), currency(string), id(string), object(string), original_credit_payment_id(string), original_invoice(object), refund_transaction(object), updated_at(string), uuid(string), voided_at(string)
  list_account_external_account:
    primary key: id
    fields: created_at(string), external_account_code(string), external_connection_type(string), id(string), object(string), updated_at(string)
  get_account_external_account:
    primary key: id
    fields: created_at(string), external_account_code(string), external_connection_type(string), id(string), object(string), updated_at(string)
  list_account_external_invoices:
    primary key: id
    fields: account(object), created_at(string), currency(string), external_id(string), external_subscription(object), id(string), line_items(array), object(string), purchased_at(string), state(string), total(string), updated_at(string)
  list_account_invoices:
    primary key: id
    fields: account(object), address(object), balance(number), billing_info_id(string), business_entity_id(string), closed_at(string), collection_method(string), coupon_redemptions(array), created_at(string), credit_payments(array), currency(string), custom_fields(array), customer_notes(string), discount(number), due_at(string), dunning_campaign_id(string), dunning_events_sent(integer), final_dunning_event(boolean), has_more_line_items(boolean), id(string), line_items(array), net_terms(integer), net_terms_type(string), number(string), object(string), origin(string), paid(number), po_number(string), previous_invoice_id(string), reference_only_currency_conversion(object), refundable_amount(number), shipping_address(object), state(string), subscription_ids(array), subtotal(number), subtotal_after_discount(number), tax(number), tax_info(object), terms_and_conditions(string), total(number), transactions(array), type(string), updated_at(string), used_tax_service(boolean), uuid(string), vat_number(string), vat_reverse_charge_notes(string)
  list_account_line_items:
    primary key: id
    fields: account(object), accounting_code(string), add_on_code(string), add_on_id(string), amount(number), avalara_service_type(integer), avalara_transaction_type(integer), bill_for_account_id(string), created_at(string), credit_applied(number), credit_reason_code(string), currency(string), custom_fields(array), description(string), destination_tax_address_source(string), discount(number), discounts(array), end_date(string), external_sku(string), harmonized_system_code(string), id(string), invoice_id(string), invoice_number(string), item_code(string), item_id(string), legacy_category(string), liability_gl_account_code(string), object(string), origin(string), origin_tax_address_source(string), original_line_item_invoice_id(string), performance_obligation_id(string), plan_code(string), plan_id(string), previous_line_item_id(string), product_code(string), proration_rate(number), quantity(integer), quantity_decimal(string), refund(boolean), refunded_quantity(integer), refunded_quantity_decimal(string), revenue_gl_account_code(string), revenue_schedule_type(string), shipping_address(object), start_date(string), state(string), subscription_id(string), subtotal(number), tax(number), tax_code(string), tax_exempt(boolean), tax_inclusive(boolean), tax_info(object), taxable(boolean), type(string), unit_amount(number), unit_amount_decimal(string), updated_at(string), uuid(string), vertex_transaction_type(string)
  list_account_notes:
    primary key: id
    fields: account_id(string), created_at(string), id(string), message(string), object(string), user(object)
  get_account_note:
    primary key: id
    fields: account_id(string), created_at(string), id(string), message(string), object(string), user(object)
  list_shipping_addresses:
    primary key: id
    fields: account_id(string), city(string), company(string), country(string), created_at(string), email(string), first_name(string), geo_code(string), id(string), last_name(string), nickname(string), object(string), phone(string), postal_code(string), region(string), street1(string), street2(string), updated_at(string), vat_number(string)
  get_shipping_address:
    primary key: id
    fields: account_id(string), city(string), company(string), country(string), created_at(string), email(string), first_name(string), geo_code(string), id(string), last_name(string), nickname(string), object(string), phone(string), postal_code(string), region(string), street1(string), street2(string), updated_at(string), vat_number(string)
  list_account_subscriptions:
    primary key: id
    fields: account(object), action_result(object), activated_at(string), active_invoice_id(string), add_ons(array), add_ons_total(number), auto_renew(boolean), bank_account_authorized_at(string), billing_info_id(string), business_entity_id(string), canceled_at(string), collection_method(string), converted_at(string), coupon_redemptions(array), created_at(string), credit_application_policy(object), currency(string), current_period_ends_at(string), current_period_started_at(string), current_term_ends_at(string), current_term_started_at(string), custom_fields(array), customer_notes(string), expiration_reason(string), expires_at(string), gateway_code(string), id(string), net_terms(integer), net_terms_type(string), object(string), paused_at(string), pending_change(object), plan(object), po_number(string), price_segment_id(string), quantity(integer), ramp_intervals(array), remaining_billing_cycles(integer), remaining_pause_cycles(integer), renewal_billing_cycles(integer), resume_at(string), revenue_schedule_type(string), shipping(object), started_with_gift(boolean), state(string), subtotal(number), tax(number), tax_inclusive(boolean), tax_info(object), terms_and_conditions(string), total(number), total_billing_cycles(integer), trial_ends_at(string), trial_started_at(string), unit_amount(number), updated_at(string), uuid(string)
  list_account_transactions:
    primary key: id
    fields: account(object), action_result(object), amount(number), avs_check(string), backup_payment_method_used(boolean), billing_address(object), collected_at(string), collection_method(string), created_at(string), currency(string), customer_message(string), customer_message_locale(string), cvv_check(string), description(string), fraud_info(object), gateway_approval_code(string), gateway_message(string), gateway_reference(string), gateway_response_code(string), gateway_response_time(number), gateway_response_values(object), id(string), initiator(string), invoice(object), ip_address_country(string), ip_address_v4(string), merchant_reason_code(string), next_action(object), object(string), origin(string), original_transaction_id(string), payment_gateway(object), payment_method(object), refunded(boolean), status(string), status_code(string), status_message(string), subscription_ids(array), success(boolean), type(string), updated_at(string), uuid(string), vat_number(string), voided_at(string), voided_by_invoice(object)
  list_child_accounts:
    primary key: id
    fields: address(object), bill_date(string), bill_to(string), billing_info(object), cc_emails(string), code(string), company(string), created_at(string), custom_fields(array), deleted_at(string), dunning_campaign_id(string), email(string), entity_use_code(string), exemption_certificate(string), external_accounts(array), first_name(string), has_active_subscription(boolean), has_canceled_subscription(boolean), has_future_subscription(boolean), has_live_subscription(boolean), has_past_due_invoice(boolean), has_paused_subscription(boolean), hosted_login_token(string), id(string), invoice_template_id(string), last_name(string), object(string), override_business_entity_id(string), parent_account_id(string), preferred_locale(string), preferred_time_zone(string), shipping_addresses(array), state(string), tax_exempt(boolean), updated_at(string), username(string), vat_number(string)
  list_account_acquisition:
    primary key: id
    fields: account(object), acquired_at(string), campaign(string), channel(string), cost(object), created_at(string), id(string), object(string), subchannel(string), updated_at(string)
  list_coupons:
    primary key: id
    fields: applies_to_all_items(boolean), applies_to_all_plans(boolean), applies_to_non_plan_charges(boolean), code(string), coupon_type(string), created_at(string), discount(object), duration(string), expired_at(string), free_trial_amount(integer), free_trial_unit(string), hosted_page_description(string), id(string), invoice_description(string), items(array), max_redemptions(integer), max_redemptions_per_account(integer), name(string), object(string), plans(array), redeem_by(string), redemption_resource(string), state(string), temporal_amount(integer), temporal_unit(string), unique_code_template(string), unique_coupon_code(object), unique_coupon_codes_count(integer), updated_at(string)
  get_coupon:
    primary key: id
    fields: applies_to_all_items(boolean), applies_to_all_plans(boolean), applies_to_non_plan_charges(boolean), code(string), coupon_type(string), created_at(string), discount(object), duration(string), expired_at(string), free_trial_amount(integer), free_trial_unit(string), hosted_page_description(string), id(string), invoice_description(string), items(array), max_redemptions(integer), max_redemptions_per_account(integer), name(string), object(string), plans(array), redeem_by(string), redemption_resource(string), state(string), temporal_amount(integer), temporal_unit(string), unique_code_template(string), unique_coupon_code(object), unique_coupon_codes_count(integer), updated_at(string)
  list_unique_coupon_codes:
    primary key: id
    fields: bulk_coupon_code(string), bulk_coupon_id(string), code(string), created_at(string), expired_at(string), id(string), object(string), redeemed_at(string), state(string), updated_at(string)
  list_credit_payments:
    primary key: id
    fields: account(object), action(string), amount(number), applied_to_invoice(object), created_at(string), currency(string), id(string), object(string), original_credit_payment_id(string), original_invoice(object), refund_transaction(object), updated_at(string), uuid(string), voided_at(string)
  get_credit_payment:
    primary key: id
    fields: account(object), action(string), amount(number), applied_to_invoice(object), created_at(string), currency(string), id(string), object(string), original_credit_payment_id(string), original_invoice(object), refund_transaction(object), updated_at(string), uuid(string), voided_at(string)
  list_custom_field_definitions:
    primary key: id
    fields: created_at(string), deleted_at(string), display_name(string), id(string), name(string), object(string), related_type(string), tooltip(string), updated_at(string), user_access(string)
  get_custom_field_definition:
    primary key: id
    fields: created_at(string), deleted_at(string), display_name(string), id(string), name(string), object(string), related_type(string), tooltip(string), updated_at(string), user_access(string)
  list_general_ledger_accounts:
    primary key: id
    fields: account_type(string), code(string), created_at(string), description(string), id(string), object(string), updated_at(string)
  get_general_ledger_account:
    primary key: id
    fields: account_type(string), code(string), created_at(string), description(string), id(string), object(string), updated_at(string)
  get_performance_obligation:
    primary key: id
    fields: created_at(string), id(string), name(string), updated_at(string)
  get_performance_obligations:
    primary key: id
    fields: created_at(string), id(string), name(string), updated_at(string)
  list_invoice_template_accounts:
    primary key: id
    fields: address(object), bill_date(string), bill_to(string), billing_info(object), cc_emails(string), code(string), company(string), created_at(string), custom_fields(array), deleted_at(string), dunning_campaign_id(string), email(string), entity_use_code(string), exemption_certificate(string), external_accounts(array), first_name(string), has_active_subscription(boolean), has_canceled_subscription(boolean), has_future_subscription(boolean), has_live_subscription(boolean), has_past_due_invoice(boolean), has_paused_subscription(boolean), hosted_login_token(string), id(string), invoice_template_id(string), last_name(string), object(string), override_business_entity_id(string), parent_account_id(string), preferred_locale(string), preferred_time_zone(string), shipping_addresses(array), state(string), tax_exempt(boolean), updated_at(string), username(string), vat_number(string)
  list_items:
    primary key: id
    fields: accounting_code(string), avalara_service_type(integer), avalara_transaction_type(integer), code(string), created_at(string), currencies(array), custom_fields(array), deleted_at(string), description(string), external_sku(string), harmonized_system_code(string), id(string), liability_gl_account_id(string), name(string), object(string), performance_obligation_id(string), revenue_gl_account_id(string), revenue_schedule_type(string), state(string), tax_code(string), tax_exempt(boolean), updated_at(string)
  get_item:
    primary key: id
    fields: accounting_code(string), avalara_service_type(integer), avalara_transaction_type(integer), code(string), created_at(string), currencies(array), custom_fields(array), deleted_at(string), description(string), external_sku(string), harmonized_system_code(string), id(string), liability_gl_account_id(string), name(string), object(string), performance_obligation_id(string), revenue_gl_account_id(string), revenue_schedule_type(string), state(string), tax_code(string), tax_exempt(boolean), updated_at(string)
  list_measured_unit:
    primary key: id
    fields: created_at(string), deleted_at(string), description(string), display_name(string), id(string), name(string), object(string), state(string), updated_at(string)
  get_measured_unit:
    primary key: id
    fields: created_at(string), deleted_at(string), description(string), display_name(string), id(string), name(string), object(string), state(string), updated_at(string)
  list_external_products:
    primary key: id
    fields: created_at(string), external_product_references(array), id(string), name(string), object(string), plan(object), updated_at(string)
  get_external_product:
    primary key: id
    fields: created_at(string), external_product_references(array), id(string), name(string), object(string), plan(object), updated_at(string)
  list_external_product_external_product_references:
    primary key: id
    fields: created_at(string), external_connection_type(string), id(string), object(string), reference_code(string), updated_at(string)
  get_external_product_external_product_reference:
    primary key: id
    fields: created_at(string), external_connection_type(string), id(string), object(string), reference_code(string), updated_at(string)
  list_external_subscriptions:
    primary key: id
    fields: account(object), activated_at(string), app_identifier(string), auto_renew(boolean), canceled_at(string), created_at(string), expires_at(string), external_id(string), external_payment_phases(array), external_product_reference(object), id(string), imported(boolean), in_grace_period(boolean), last_purchased(string), object(string), quantity(integer), state(string), test(boolean), trial_ends_at(string), trial_started_at(string), updated_at(string), uuid(string)
  get_external_subscription:
    primary key: id
    fields: account(object), activated_at(string), app_identifier(string), auto_renew(boolean), canceled_at(string), created_at(string), expires_at(string), external_id(string), external_payment_phases(array), external_product_reference(object), id(string), imported(boolean), in_grace_period(boolean), last_purchased(string), object(string), quantity(integer), state(string), test(boolean), trial_ends_at(string), trial_started_at(string), updated_at(string), uuid(string)
  list_external_subscription_external_invoices:
    primary key: id
    fields: account(object), created_at(string), currency(string), external_id(string), external_subscription(object), id(string), line_items(array), object(string), purchased_at(string), state(string), total(string), updated_at(string)
  invoices:
    primary key: id
    cursor: created_at
    fields: account_id(string), created_at(string), id(string), state(string), total(number)
  get_invoice:
    primary key: id
    fields: account(object), address(object), balance(number), billing_info_id(string), business_entity_id(string), closed_at(string), collection_method(string), coupon_redemptions(array), created_at(string), credit_payments(array), currency(string), custom_fields(array), customer_notes(string), discount(number), due_at(string), dunning_campaign_id(string), dunning_events_sent(integer), final_dunning_event(boolean), has_more_line_items(boolean), id(string), line_items(array), net_terms(integer), net_terms_type(string), number(string), object(string), origin(string), paid(number), po_number(string), previous_invoice_id(string), reference_only_currency_conversion(object), refundable_amount(number), shipping_address(object), state(string), subscription_ids(array), subtotal(number), subtotal_after_discount(number), tax(number), tax_info(object), terms_and_conditions(string), total(number), transactions(array), type(string), updated_at(string), used_tax_service(boolean), uuid(string), vat_number(string), vat_reverse_charge_notes(string)
  list_invoice_line_items:
    primary key: id
    fields: account(object), accounting_code(string), add_on_code(string), add_on_id(string), amount(number), avalara_service_type(integer), avalara_transaction_type(integer), bill_for_account_id(string), created_at(string), credit_applied(number), credit_reason_code(string), currency(string), custom_fields(array), description(string), destination_tax_address_source(string), discount(number), discounts(array), end_date(string), external_sku(string), harmonized_system_code(string), id(string), invoice_id(string), invoice_number(string), item_code(string), item_id(string), legacy_category(string), liability_gl_account_code(string), object(string), origin(string), origin_tax_address_source(string), original_line_item_invoice_id(string), performance_obligation_id(string), plan_code(string), plan_id(string), previous_line_item_id(string), product_code(string), proration_rate(number), quantity(integer), quantity_decimal(string), refund(boolean), refunded_quantity(integer), refunded_quantity_decimal(string), revenue_gl_account_code(string), revenue_schedule_type(string), shipping_address(object), start_date(string), state(string), subscription_id(string), subtotal(number), tax(number), tax_code(string), tax_exempt(boolean), tax_inclusive(boolean), tax_info(object), taxable(boolean), type(string), unit_amount(number), unit_amount_decimal(string), updated_at(string), uuid(string), vertex_transaction_type(string)
  list_invoice_coupon_redemptions:
    primary key: id
    fields: account(object), coupon(object), created_at(string), currency(string), discounted(number), id(string), object(string), remaining_duration(object), removed_at(string), state(string), subscription_id(string), updated_at(string), uuid(string)
  list_related_invoices:
    primary key: id
    fields: account(object), address(object), balance(number), billing_info_id(string), business_entity_id(string), closed_at(string), collection_method(string), coupon_redemptions(array), created_at(string), credit_payments(array), currency(string), custom_fields(array), customer_notes(string), discount(number), due_at(string), dunning_campaign_id(string), dunning_events_sent(integer), final_dunning_event(boolean), has_more_line_items(boolean), id(string), line_items(array), net_terms(integer), net_terms_type(string), number(string), object(string), origin(string), paid(number), po_number(string), previous_invoice_id(string), reference_only_currency_conversion(object), refundable_amount(number), shipping_address(object), state(string), subscription_ids(array), subtotal(number), subtotal_after_discount(number), tax(number), tax_info(object), terms_and_conditions(string), total(number), transactions(array), type(string), updated_at(string), used_tax_service(boolean), uuid(string), vat_number(string), vat_reverse_charge_notes(string)
  list_line_items:
    primary key: id
    fields: account(object), accounting_code(string), add_on_code(string), add_on_id(string), amount(number), avalara_service_type(integer), avalara_transaction_type(integer), bill_for_account_id(string), created_at(string), credit_applied(number), credit_reason_code(string), currency(string), custom_fields(array), description(string), destination_tax_address_source(string), discount(number), discounts(array), end_date(string), external_sku(string), harmonized_system_code(string), id(string), invoice_id(string), invoice_number(string), item_code(string), item_id(string), legacy_category(string), liability_gl_account_code(string), object(string), origin(string), origin_tax_address_source(string), original_line_item_invoice_id(string), performance_obligation_id(string), plan_code(string), plan_id(string), previous_line_item_id(string), product_code(string), proration_rate(number), quantity(integer), quantity_decimal(string), refund(boolean), refunded_quantity(integer), refunded_quantity_decimal(string), revenue_gl_account_code(string), revenue_schedule_type(string), shipping_address(object), start_date(string), state(string), subscription_id(string), subtotal(number), tax(number), tax_code(string), tax_exempt(boolean), tax_inclusive(boolean), tax_info(object), taxable(boolean), type(string), unit_amount(number), unit_amount_decimal(string), updated_at(string), uuid(string), vertex_transaction_type(string)
  get_line_item:
    primary key: id
    fields: account(object), accounting_code(string), add_on_code(string), add_on_id(string), amount(number), avalara_service_type(integer), avalara_transaction_type(integer), bill_for_account_id(string), created_at(string), credit_applied(number), credit_reason_code(string), currency(string), custom_fields(array), description(string), destination_tax_address_source(string), discount(number), discounts(array), end_date(string), external_sku(string), harmonized_system_code(string), id(string), invoice_id(string), invoice_number(string), item_code(string), item_id(string), legacy_category(string), liability_gl_account_code(string), object(string), origin(string), origin_tax_address_source(string), original_line_item_invoice_id(string), performance_obligation_id(string), plan_code(string), plan_id(string), previous_line_item_id(string), product_code(string), proration_rate(number), quantity(integer), quantity_decimal(string), refund(boolean), refunded_quantity(integer), refunded_quantity_decimal(string), revenue_gl_account_code(string), revenue_schedule_type(string), shipping_address(object), start_date(string), state(string), subscription_id(string), subtotal(number), tax(number), tax_code(string), tax_exempt(boolean), tax_inclusive(boolean), tax_info(object), taxable(boolean), type(string), unit_amount(number), unit_amount_decimal(string), updated_at(string), uuid(string), vertex_transaction_type(string)
  plans:
    primary key: id
    cursor: updated_at
    fields: code(string), id(string), name(string), state(string), updated_at(string)
  get_plan:
    primary key: id
    fields: accounting_code(string), allow_any_item_on_subscriptions(boolean), auto_renew(boolean), avalara_service_type(integer), avalara_transaction_type(integer), code(string), created_at(string), currencies(array), custom_fields(array), deleted_at(string), description(string), dunning_campaign_id(string), harmonized_system_code(string), hosted_pages(object), id(string), interval_length(integer), interval_unit(string), liability_gl_account_id(string), name(string), object(string), performance_obligation_id(string), pricing_model(string), ramp_intervals(array), revenue_gl_account_id(string), revenue_schedule_type(string), setup_fee_accounting_code(string), setup_fee_liability_gl_account_id(string), setup_fee_performance_obligation_id(string), setup_fee_revenue_gl_account_id(string), setup_fee_revenue_schedule_type(string), setup_fees(array), state(string), tax_code(string), tax_exempt(boolean), total_billing_cycles(integer), trial_length(integer), trial_requires_billing_info(boolean), trial_unit(string), updated_at(string), vertex_transaction_type(string)
  list_plan_add_ons:
    primary key: id
    fields: accounting_code(string), add_on_type(string), avalara_service_type(integer), avalara_transaction_type(integer), code(string), created_at(string), currencies(array), default_quantity(integer), deleted_at(string), display_quantity(boolean), external_sku(string), harmonized_system_code(string), id(string), item(object), liability_gl_account_id(string), measured_unit_id(string), name(string), object(string), optional(boolean), percentage_tiers(array), performance_obligation_id(string), plan_id(string), revenue_gl_account_id(string), revenue_schedule_type(string), state(string), tax_code(string), tier_type(string), tiers(array), updated_at(string), usage_calculation_type(string), usage_percentage(number), usage_timeframe(string), usage_type(string)
  get_plan_add_on:
    primary key: id
    fields: accounting_code(string), add_on_type(string), avalara_service_type(integer), avalara_transaction_type(integer), code(string), created_at(string), currencies(array), default_quantity(integer), deleted_at(string), display_quantity(boolean), external_sku(string), harmonized_system_code(string), id(string), item(object), liability_gl_account_id(string), measured_unit_id(string), name(string), object(string), optional(boolean), percentage_tiers(array), performance_obligation_id(string), plan_id(string), revenue_gl_account_id(string), revenue_schedule_type(string), state(string), tax_code(string), tier_type(string), tiers(array), updated_at(string), usage_calculation_type(string), usage_percentage(number), usage_timeframe(string), usage_type(string)
  list_price_segments:
    primary key: id
    fields: code(string), id(string), object(string)
  get_price_segment:
    primary key: id
    fields: code(string), id(string), object(string)
  list_add_ons:
    primary key: id
    fields: accounting_code(string), add_on_type(string), avalara_service_type(integer), avalara_transaction_type(integer), code(string), created_at(string), currencies(array), default_quantity(integer), deleted_at(string), display_quantity(boolean), external_sku(string), harmonized_system_code(string), id(string), item(object), liability_gl_account_id(string), measured_unit_id(string), name(string), object(string), optional(boolean), percentage_tiers(array), performance_obligation_id(string), plan_id(string), revenue_gl_account_id(string), revenue_schedule_type(string), state(string), tax_code(string), tier_type(string), tiers(array), updated_at(string), usage_calculation_type(string), usage_percentage(number), usage_timeframe(string), usage_type(string)
  get_add_on:
    primary key: id
    fields: accounting_code(string), add_on_type(string), avalara_service_type(integer), avalara_transaction_type(integer), code(string), created_at(string), currencies(array), default_quantity(integer), deleted_at(string), display_quantity(boolean), external_sku(string), harmonized_system_code(string), id(string), item(object), liability_gl_account_id(string), measured_unit_id(string), name(string), object(string), optional(boolean), percentage_tiers(array), performance_obligation_id(string), plan_id(string), revenue_gl_account_id(string), revenue_schedule_type(string), state(string), tax_code(string), tier_type(string), tiers(array), updated_at(string), usage_calculation_type(string), usage_percentage(number), usage_timeframe(string), usage_type(string)
  list_shipping_methods:
    primary key: id
    fields: accounting_code(string), code(string), created_at(string), deleted_at(string), id(string), liability_gl_account_id(string), name(string), object(string), performance_obligation_id(string), revenue_gl_account_id(string), tax_code(string), updated_at(string)
  get_shipping_method:
    primary key: id
    fields: accounting_code(string), code(string), created_at(string), deleted_at(string), id(string), liability_gl_account_id(string), name(string), object(string), performance_obligation_id(string), revenue_gl_account_id(string), tax_code(string), updated_at(string)
  subscriptions:
    primary key: id
    cursor: updated_at
    fields: account_id(string), created_at(string), id(string), plan_id(string), state(string), updated_at(string)
  get_subscription:
    primary key: id
    fields: account(object), action_result(object), activated_at(string), active_invoice_id(string), add_ons(array), add_ons_total(number), auto_renew(boolean), bank_account_authorized_at(string), billing_info_id(string), business_entity_id(string), canceled_at(string), collection_method(string), converted_at(string), coupon_redemptions(array), created_at(string), credit_application_policy(object), currency(string), current_period_ends_at(string), current_period_started_at(string), current_term_ends_at(string), current_term_started_at(string), custom_fields(array), customer_notes(string), expiration_reason(string), expires_at(string), gateway_code(string), id(string), net_terms(integer), net_terms_type(string), object(string), paused_at(string), pending_change(object), plan(object), po_number(string), price_segment_id(string), quantity(integer), ramp_intervals(array), remaining_billing_cycles(integer), remaining_pause_cycles(integer), renewal_billing_cycles(integer), resume_at(string), revenue_schedule_type(string), shipping(object), started_with_gift(boolean), state(string), subtotal(number), tax(number), tax_inclusive(boolean), tax_info(object), terms_and_conditions(string), total(number), total_billing_cycles(integer), trial_ends_at(string), trial_started_at(string), unit_amount(number), updated_at(string), uuid(string)
  get_subscription_change:
    primary key: id
    fields: activate_at(string), activated(boolean), add_ons(array), billing_info(object), business_entity(object), created_at(string), custom_fields(array), deleted_at(string), id(string), invoice_collection(object), next_bill_date(string), object(string), plan(object), quantity(integer), ramp_intervals(array), revenue_schedule_type(string), shipping(object), subscription_id(string), tax_inclusive(boolean), unit_amount(number), updated_at(string)
  list_subscription_invoices:
    primary key: id
    fields: account(object), address(object), balance(number), billing_info_id(string), business_entity_id(string), closed_at(string), collection_method(string), coupon_redemptions(array), created_at(string), credit_payments(array), currency(string), custom_fields(array), customer_notes(string), discount(number), due_at(string), dunning_campaign_id(string), dunning_events_sent(integer), final_dunning_event(boolean), has_more_line_items(boolean), id(string), line_items(array), net_terms(integer), net_terms_type(string), number(string), object(string), origin(string), paid(number), po_number(string), previous_invoice_id(string), reference_only_currency_conversion(object), refundable_amount(number), shipping_address(object), state(string), subscription_ids(array), subtotal(number), subtotal_after_discount(number), tax(number), tax_info(object), terms_and_conditions(string), total(number), transactions(array), type(string), updated_at(string), used_tax_service(boolean), uuid(string), vat_number(string), vat_reverse_charge_notes(string)
  list_subscription_line_items:
    primary key: id
    fields: account(object), accounting_code(string), add_on_code(string), add_on_id(string), amount(number), avalara_service_type(integer), avalara_transaction_type(integer), bill_for_account_id(string), created_at(string), credit_applied(number), credit_reason_code(string), currency(string), custom_fields(array), description(string), destination_tax_address_source(string), discount(number), discounts(array), end_date(string), external_sku(string), harmonized_system_code(string), id(string), invoice_id(string), invoice_number(string), item_code(string), item_id(string), legacy_category(string), liability_gl_account_code(string), object(string), origin(string), origin_tax_address_source(string), original_line_item_invoice_id(string), performance_obligation_id(string), plan_code(string), plan_id(string), previous_line_item_id(string), product_code(string), proration_rate(number), quantity(integer), quantity_decimal(string), refund(boolean), refunded_quantity(integer), refunded_quantity_decimal(string), revenue_gl_account_code(string), revenue_schedule_type(string), shipping_address(object), start_date(string), state(string), subscription_id(string), subtotal(number), tax(number), tax_code(string), tax_exempt(boolean), tax_inclusive(boolean), tax_info(object), taxable(boolean), type(string), unit_amount(number), unit_amount_decimal(string), updated_at(string), uuid(string), vertex_transaction_type(string)
  list_subscription_coupon_redemptions:
    primary key: id
    fields: account(object), coupon(object), created_at(string), currency(string), discounted(number), id(string), object(string), remaining_duration(object), removed_at(string), state(string), subscription_id(string), updated_at(string), uuid(string)
  get_subscription_coupon_redemption:
    primary key: id
    fields: account(object), coupon(object), created_at(string), currency(string), discounted(number), id(string), object(string), remaining_duration(object), removed_at(string), state(string), subscription_id(string), updated_at(string), uuid(string)
  list_usage:
    primary key: id
    fields: amount(number), billed_at(string), created_at(string), id(string), measured_unit_id(string), merchant_tag(string), object(string), percentage_tiers(array), recording_timestamp(string), tier_type(string), tiers(array), unit_amount(number), unit_amount_decimal(string), updated_at(string), usage_percentage(number), usage_timestamp(string), usage_type(string)
  get_usage:
    primary key: id
    fields: amount(number), billed_at(string), created_at(string), id(string), measured_unit_id(string), merchant_tag(string), object(string), percentage_tiers(array), recording_timestamp(string), tier_type(string), tiers(array), unit_amount(number), unit_amount_decimal(string), updated_at(string), usage_percentage(number), usage_timestamp(string), usage_type(string)
  transactions:
    primary key: id
    cursor: created_at
    fields: account_id(string), amount(number), created_at(string), id(string), status(string)
  get_transaction:
    primary key: id
    fields: account(object), action_result(object), amount(number), avs_check(string), backup_payment_method_used(boolean), billing_address(object), collected_at(string), collection_method(string), created_at(string), currency(string), customer_message(string), customer_message_locale(string), cvv_check(string), description(string), fraud_info(object), gateway_approval_code(string), gateway_message(string), gateway_reference(string), gateway_response_code(string), gateway_response_time(number), gateway_response_values(object), id(string), initiator(string), invoice(object), ip_address_country(string), ip_address_v4(string), merchant_reason_code(string), next_action(object), object(string), origin(string), original_transaction_id(string), payment_gateway(object), payment_method(object), refunded(boolean), status(string), status_code(string), status_message(string), subscription_ids(array), success(boolean), type(string), updated_at(string), uuid(string), vat_number(string), voided_at(string), voided_by_invoice(object)
  get_unique_coupon_code:
    primary key: id
    fields: bulk_coupon_code(string), bulk_coupon_id(string), code(string), created_at(string), expired_at(string), id(string), object(string), redeemed_at(string), state(string), updated_at(string)
  list_dunning_campaigns:
    primary key: id
    fields: code(string), created_at(string), default_campaign(boolean), deleted_at(string), description(string), dunning_cycles(array), id(string), name(string), object(string), updated_at(string)
  get_dunning_campaign:
    primary key: id
    fields: code(string), created_at(string), default_campaign(boolean), deleted_at(string), description(string), dunning_cycles(array), id(string), name(string), object(string), updated_at(string)
  list_invoice_templates:
    primary key: id
    fields: code(string), created_at(string), description(string), id(string), name(string), updated_at(string)
  get_invoice_template:
    primary key: id
    fields: code(string), created_at(string), description(string), id(string), name(string), updated_at(string)
  list_external_invoices:
    primary key: id
    fields: account(object), created_at(string), currency(string), external_id(string), external_subscription(object), id(string), line_items(array), object(string), purchased_at(string), state(string), total(string), updated_at(string)
  show_external_invoice:
    primary key: id
    fields: account(object), created_at(string), currency(string), external_id(string), external_subscription(object), id(string), line_items(array), object(string), purchased_at(string), state(string), total(string), updated_at(string)
  list_external_subscription_external_payment_phases:
    primary key: id
    fields: amount(string), created_at(string), currency(string), ending_billing_period_index(integer), ends_at(string), id(string), object(string), offer_name(string), offer_type(string), period_count(integer), period_length(string), started_at(string), starting_billing_period_index(integer), updated_at(string)
  get_external_subscription_external_payment_phase:
    primary key: id
    fields: amount(string), created_at(string), currency(string), ending_billing_period_index(integer), ends_at(string), id(string), object(string), offer_name(string), offer_type(string), period_count(integer), period_length(string), started_at(string), starting_billing_period_index(integer), updated_at(string)
  list_entitlements:
    primary key: customer_permission_id
    fields: created_at(string), customer_permission(object), customer_permission_id(string), granted_by(array), object(string), updated_at(string)
  list_account_external_subscriptions:
    primary key: id
    fields: account(object), activated_at(string), app_identifier(string), auto_renew(boolean), canceled_at(string), created_at(string), expires_at(string), external_id(string), external_payment_phases(array), external_product_reference(object), id(string), imported(boolean), in_grace_period(boolean), last_purchased(string), object(string), quantity(integer), state(string), test(boolean), trial_ends_at(string), trial_started_at(string), updated_at(string), uuid(string)
  get_business_entity:
    primary key: id
    fields: code(string), created_at(string), default_liability_gl_account_id(string), default_registration_number(string), default_revenue_gl_account_id(string), default_vat_number(string), destination_tax_address_source(string), id(string), invoice_display_address(object), name(string), object(string), origin_tax_address_source(string), subscriber_location_countries(array), tax_address(object), updated_at(string)
  list_business_entities:
    primary key: id
    fields: code(string), created_at(string), default_liability_gl_account_id(string), default_registration_number(string), default_revenue_gl_account_id(string), default_vat_number(string), destination_tax_address_source(string), id(string), invoice_display_address(object), name(string), object(string), origin_tax_address_source(string), subscriber_location_countries(array), tax_address(object), updated_at(string)
  list_gift_cards:
    primary key: id
    fields: balance(number), canceled_at(string), created_at(string), currency(string), delivered_at(string), delivery(object), gifter_account_id(string), id(string), liability_gl_account_id(string), object(string), performance_obligation_id(string), product_code(string), purchase_invoice_id(string), recipient_account_id(string), redeemed_at(string), redemption_code(string), redemption_invoice_id(string), revenue_gl_account_id(string), unit_amount(number), updated_at(string)
  get_gift_card:
    primary key: id
    fields: balance(number), canceled_at(string), created_at(string), currency(string), delivered_at(string), delivery(object), gifter_account_id(string), id(string), liability_gl_account_id(string), object(string), performance_obligation_id(string), product_code(string), purchase_invoice_id(string), recipient_account_id(string), redeemed_at(string), redemption_code(string), redemption_invoice_id(string), revenue_gl_account_id(string), unit_amount(number), updated_at(string)
  list_business_entity_invoices:
    primary key: id
    fields: account(object), address(object), balance(number), billing_info_id(string), business_entity_id(string), closed_at(string), collection_method(string), coupon_redemptions(array), created_at(string), credit_payments(array), currency(string), custom_fields(array), customer_notes(string), discount(number), due_at(string), dunning_campaign_id(string), dunning_events_sent(integer), final_dunning_event(boolean), has_more_line_items(boolean), id(string), line_items(array), net_terms(integer), net_terms_type(string), number(string), object(string), origin(string), paid(number), po_number(string), previous_invoice_id(string), reference_only_currency_conversion(object), refundable_amount(number), shipping_address(object), state(string), subscription_ids(array), subtotal(number), subtotal_after_discount(number), tax(number), tax_info(object), terms_and_conditions(string), total(number), transactions(array), type(string), updated_at(string), used_tax_service(boolean), uuid(string), vat_number(string), vat_reverse_charge_notes(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_account:
    endpoint: POST /accounts
    required fields: code
    optional fields: acquisition, external_accounts, shipping_addresses, username, email, preferred_locale, preferred_time_zone, cc_emails, first_name, last_name, company, vat_number, tax_exempt, exemption_certificate, override_business_entity_id, parent_account_code, parent_account_id, bill_to, transaction_type, dunning_campaign_id, invoice_template_id, address, billing_info, custom_fields, entity_use_code, bill_date
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_account:
    endpoint: PUT /accounts/{{ record.account_id }}
    required fields: account_id
    optional fields: username, email, preferred_locale, preferred_time_zone, cc_emails, first_name, last_name, company, vat_number, tax_exempt, exemption_certificate, override_business_entity_id, parent_account_code, parent_account_id, bill_to, transaction_type, dunning_campaign_id, invoice_template_id, address, billing_info, custom_fields, entity_use_code, bill_date
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  deactivate_account:
    endpoint: DELETE /accounts/{{ record.account_id }}
    required fields: account_id, redact
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  redact_account:
    endpoint: PUT /accounts/{{ record.account_id }}/redact
    required fields: account_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_account_acquisition:
    endpoint: PUT /accounts/{{ record.account_id }}/acquisition
    required fields: account_id
    optional fields: cost, channel, subchannel, campaign, acquired_at
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_account_acquisition:
    endpoint: DELETE /accounts/{{ record.account_id }}/acquisition
    required fields: account_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  reactivate_account:
    endpoint: PUT /accounts/{{ record.account_id }}/reactivate
    required fields: account_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_billing_info:
    endpoint: PUT /accounts/{{ record.account_id }}/billing_info
    required fields: account_id
    optional fields: token_id, first_name, last_name, company, address, number, month, year, cvv, currency, vat_number, ip_address, gateway_token, gateway_code, payment_gateway_references, gateway_attributes, amazon_billing_agreement_id, paypal_billing_agreement_id, roku_billing_agreement_id, fraud_session_id, adyen_risk_profile_reference_id, transaction_type, three_d_secure_action_result_token_id, iban, name_on_account, account_number, routing_number, sort_code, type, account_type, tax_identifier, tax_identifier_type, primary_payment_method, backup_payment_method, external_hpp_type, online_banking_payment_type, card_type, card_network_preference, return_url, authentication_method
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_billing_info:
    endpoint: DELETE /accounts/{{ record.account_id }}/billing_info
    required fields: account_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  verify_billing_info:
    endpoint: POST /accounts/{{ record.account_id }}/billing_info/verify
    required fields: account_id
    optional fields: gateway_code, three_d_secure_action_result_token_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  verify_billing_info_cvv:
    endpoint: POST /accounts/{{ record.account_id }}/billing_info/verify_cvv
    required fields: account_id
    optional fields: verification_value, gateway_code, three_d_secure_action_result_token_id, token_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_billing_info:
    endpoint: POST /accounts/{{ record.account_id }}/billing_infos
    required fields: account_id
    optional fields: token_id, first_name, last_name, company, address, number, month, year, cvv, currency, vat_number, ip_address, gateway_token, gateway_code, payment_gateway_references, gateway_attributes, amazon_billing_agreement_id, paypal_billing_agreement_id, roku_billing_agreement_id, fraud_session_id, adyen_risk_profile_reference_id, transaction_type, three_d_secure_action_result_token_id, iban, name_on_account, account_number, routing_number, sort_code, type, account_type, tax_identifier, tax_identifier_type, primary_payment_method, backup_payment_method, external_hpp_type, online_banking_payment_type, card_type, card_network_preference, return_url, authentication_method
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_a_billing_info:
    endpoint: PUT /accounts/{{ record.account_id }}/billing_infos/{{ record.billing_info_id }}
    required fields: account_id, billing_info_id
    optional fields: token_id, first_name, last_name, company, address, number, month, year, cvv, currency, vat_number, ip_address, gateway_token, gateway_code, payment_gateway_references, gateway_attributes, amazon_billing_agreement_id, paypal_billing_agreement_id, roku_billing_agreement_id, fraud_session_id, adyen_risk_profile_reference_id, transaction_type, three_d_secure_action_result_token_id, iban, name_on_account, account_number, routing_number, sort_code, type, account_type, tax_identifier, tax_identifier_type, primary_payment_method, backup_payment_method, external_hpp_type, online_banking_payment_type, card_type, card_network_preference, return_url, authentication_method
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_a_billing_info:
    endpoint: DELETE /accounts/{{ record.account_id }}/billing_infos/{{ record.billing_info_id }}
    required fields: account_id, billing_info_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  verify_billing_infos:
    endpoint: POST /accounts/{{ record.account_id }}/billing_infos/{{ record.billing_info_id }}/verify
    required fields: account_id, billing_info_id
    optional fields: gateway_code, three_d_secure_action_result_token_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  verify_billing_infos_cvv:
    endpoint: POST /accounts/{{ record.account_id }}/billing_infos/{{ record.billing_info_id }}/verify_cvv
    required fields: account_id, billing_info_id
    optional fields: verification_value, gateway_code, three_d_secure_action_result_token_id, token_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_coupon_redemption:
    endpoint: POST /accounts/{{ record.account_id }}/coupon_redemptions/active
    required fields: account_id, coupon_id
    optional fields: currency, subscription_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_coupon_redemption:
    endpoint: DELETE /accounts/{{ record.account_id }}/coupon_redemptions/active
    required fields: account_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_coupon_redemption_by_id:
    endpoint: DELETE /accounts/{{ record.account_id }}/coupon_redemptions/{{ record.coupon_redemption_id }}
    required fields: account_id, coupon_redemption_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_account_external_account:
    endpoint: POST /accounts/{{ record.account_id }}/external_accounts
    required fields: account_id, external_account_code, external_connection_type
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_account_external_account:
    endpoint: PUT /accounts/{{ record.account_id }}/external_accounts/{{ record.external_account_id }}
    required fields: account_id, external_account_id
    optional fields: external_account_code, external_connection_type
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  delete_account_external_account:
    endpoint: DELETE /accounts/{{ record.account_id }}/external_accounts/{{ record.external_account_id }}
    required fields: account_id, external_account_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_invoice:
    endpoint: POST /accounts/{{ record.account_id }}/invoices
    required fields: account_id, currency
    optional fields: business_entity_id, business_entity_code, collection_method, charge_customer_notes, credit_customer_notes, net_terms, net_terms_type, credit_application_policy, po_number, terms_and_conditions, vat_reverse_charge_notes, vertex_transaction_type, transaction_descriptor_suffix
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_line_item:
    endpoint: POST /accounts/{{ record.account_id }}/line_items
    required fields: account_id, currency, unit_amount, type
    optional fields: tax_inclusive, quantity, description, item_code, item_id, revenue_schedule_type, credit_reason_code, accounting_code, liability_gl_account_id, revenue_gl_account_id, performance_obligation_id, tax_exempt, avalara_transaction_type, avalara_service_type, vertex_transaction_type, tax_code, harmonized_system_code, product_code, origin, custom_fields, start_date, end_date, origin_tax_address_source, destination_tax_address_source
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_account_note:
    endpoint: POST /accounts/{{ record.account_id }}/notes
    required fields: account_id, message
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_account_note:
    endpoint: DELETE /accounts/{{ record.account_id }}/notes/{{ record.account_note_id }}
    required fields: account_id, account_note_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_shipping_address:
    endpoint: POST /accounts/{{ record.account_id }}/shipping_addresses
    required fields: account_id, first_name, last_name, street1, city, postal_code, country
    optional fields: nickname, company, email, vat_number, phone, street2, region, geo_code
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_shipping_address:
    endpoint: PUT /accounts/{{ record.account_id }}/shipping_addresses/{{ record.shipping_address_id }}
    required fields: account_id, shipping_address_id
    optional fields: nickname, first_name, last_name, company, email, vat_number, phone, street1, street2, city, region, postal_code, country, geo_code
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_shipping_address:
    endpoint: DELETE /accounts/{{ record.account_id }}/shipping_addresses/{{ record.shipping_address_id }}
    required fields: account_id, shipping_address_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_coupon:
    endpoint: POST /coupons
    required fields: code, discount_type
    optional fields: name, max_redemptions, max_redemptions_per_account, hosted_description, invoice_description, redeem_by_date, discount_percent, free_trial_unit, free_trial_amount, currencies, applies_to_non_plan_charges, applies_to_all_plans, applies_to_all_items, plan_codes, item_codes, duration, temporal_amount, temporal_unit, coupon_type, unique_code_template, redemption_resource
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_coupon:
    endpoint: PUT /coupons/{{ record.coupon_id }}
    required fields: coupon_id
    optional fields: name, max_redemptions, max_redemptions_per_account, hosted_description, invoice_description, redeem_by_date
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  deactivate_coupon:
    endpoint: DELETE /coupons/{{ record.coupon_id }}
    required fields: coupon_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  generate_unique_coupon_codes:
    endpoint: POST /coupons/{{ record.coupon_id }}/generate
    required fields: coupon_id
    optional fields: number_of_unique_codes
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  generate_unique_coupon_codes_sync:
    endpoint: POST /coupons/{{ record.coupon_id }}/generate_sync
    required fields: coupon_id, number_of_unique_codes
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  restore_coupon:
    endpoint: PUT /coupons/{{ record.coupon_id }}/restore
    required fields: coupon_id
    optional fields: name, max_redemptions, max_redemptions_per_account, hosted_description, invoice_description, redeem_by_date
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_general_ledger_account:
    endpoint: POST /general_ledger_accounts
    optional fields: code, description, account_type
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_general_ledger_account:
    endpoint: PUT /general_ledger_accounts/{{ record.general_ledger_account_id }}
    required fields: general_ledger_account_id
    optional fields: code, description
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_item:
    endpoint: POST /items
    required fields: code, name
    optional fields: description, external_sku, accounting_code, revenue_schedule_type, performance_obligation_id, liability_gl_account_id, revenue_gl_account_id, avalara_transaction_type, avalara_service_type, tax_code, harmonized_system_code, tax_exempt, custom_fields, currencies
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_item:
    endpoint: PUT /items/{{ record.item_id }}
    required fields: item_id
    optional fields: code, name, description, external_sku, accounting_code, revenue_schedule_type, performance_obligation_id, liability_gl_account_id, revenue_gl_account_id, avalara_transaction_type, avalara_service_type, tax_code, harmonized_system_code, tax_exempt, custom_fields, currencies
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  deactivate_item:
    endpoint: DELETE /items/{{ record.item_id }}
    required fields: item_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  reactivate_item:
    endpoint: PUT /items/{{ record.item_id }}/reactivate
    required fields: item_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_measured_unit:
    endpoint: POST /measured_units
    required fields: name, display_name
    optional fields: description
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_measured_unit:
    endpoint: PUT /measured_units/{{ record.measured_unit_id }}
    required fields: measured_unit_id
    optional fields: name, display_name, description
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_measured_unit:
    endpoint: DELETE /measured_units/{{ record.measured_unit_id }}
    required fields: measured_unit_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_external_product:
    endpoint: POST /external_products
    required fields: name
    optional fields: plan_id, external_product_references
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_external_product:
    endpoint: PUT /external_products/{{ record.external_product_id }}
    required fields: external_product_id, plan_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  deactivate_external_products:
    endpoint: DELETE /external_products/{{ record.external_product_id }}
    required fields: external_product_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_external_product_external_product_reference:
    endpoint: POST /external_products/{{ record.external_product_id }}/external_product_references
    required fields: external_product_id
    optional fields: reference_code, external_connection_type
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  deactivate_external_product_external_product_reference:
    endpoint: DELETE /external_products/{{ record.external_product_id }}/external_product_references/{{ record.external_product_reference_id }}
    required fields: external_product_id, external_product_reference_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_external_subscription:
    endpoint: POST /external_subscriptions
    optional fields: account, external_product_reference, external_id, last_purchased, auto_renew, state, app_identifier, quantity, activated_at, expires_at, trial_started_at, trial_ends_at, imported
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  put_external_subscription:
    endpoint: PUT /external_subscriptions/{{ record.external_subscription_id }}
    required fields: external_subscription_id
    optional fields: external_product_reference, external_id, last_purchased, auto_renew, state, app_identifier, quantity, activated_at, expires_at, trial_started_at, trial_ends_at, imported
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_external_invoice:
    endpoint: POST /external_subscriptions/{{ record.external_subscription_id }}/external_invoices
    required fields: external_subscription_id, external_id, state, total, currency, purchased_at
    optional fields: line_items, external_payment_phase, external_payment_phase_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_invoice:
    endpoint: PUT /invoices/{{ record.invoice_id }}
    required fields: invoice_id
    optional fields: po_number, vat_reverse_charge_notes, terms_and_conditions, customer_notes, net_terms, address, gateway_code
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  apply_credit_balance:
    endpoint: PUT /invoices/{{ record.invoice_id }}/apply_credit_balance
    required fields: invoice_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  collect_invoice:
    endpoint: PUT /invoices/{{ record.invoice_id }}/collect
    required fields: invoice_id
    optional fields: three_d_secure_action_result_token_id, transaction_type, billing_info_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  mark_invoice_failed:
    endpoint: PUT /invoices/{{ record.invoice_id }}/mark_failed
    required fields: invoice_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  mark_invoice_successful:
    endpoint: PUT /invoices/{{ record.invoice_id }}/mark_successful
    required fields: invoice_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  reopen_invoice:
    endpoint: PUT /invoices/{{ record.invoice_id }}/reopen
    required fields: invoice_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  void_invoice:
    endpoint: PUT /invoices/{{ record.invoice_id }}/void
    required fields: invoice_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  record_external_transaction:
    endpoint: POST /invoices/{{ record.invoice_id }}/transactions
    required fields: invoice_id
    optional fields: payment_method, description, amount, collected_at
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  refund_invoice:
    endpoint: POST /invoices/{{ record.invoice_id }}/refund
    required fields: invoice_id, type
    optional fields: amount, percentage, line_items, refund_method, credit_customer_notes, external_refund
    risk: critical — refund_invoice moves money by refunding an invoice; requires destructive confirmation and reverse ETL plan/preview/approval/execute. Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_invoice_retry:
    endpoint: POST /invoices/recovery
    required fields: currency, due_at, external_recovery_eligible, account, line_items
    optional fields: po_number, transaction_descriptor_suffix
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_line_item:
    endpoint: DELETE /line_items/{{ record.line_item_id }}
    required fields: line_item_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_plan:
    endpoint: POST /plans
    required fields: code, name, currencies
    optional fields: pricing_model, ramp_intervals, setup_fees, add_ons, interval_unit, interval_length, description, accounting_code, revenue_schedule_type, liability_gl_account_id, revenue_gl_account_id, performance_obligation_id, setup_fee_accounting_code, setup_fee_revenue_schedule_type, setup_fee_liability_gl_account_id, setup_fee_revenue_gl_account_id, setup_fee_performance_obligation_id, trial_unit, trial_length, trial_requires_billing_info, total_billing_cycles, auto_renew, custom_fields, avalara_transaction_type, avalara_service_type, tax_code, harmonized_system_code, tax_exempt, vertex_transaction_type, hosted_pages, allow_any_item_on_subscriptions, dunning_campaign_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_plan:
    endpoint: PUT /plans/{{ record.plan_id }}
    required fields: plan_id
    optional fields: code, name, currencies, ramp_intervals, setup_fees, description, accounting_code, revenue_schedule_type, liability_gl_account_id, revenue_gl_account_id, performance_obligation_id, setup_fee_accounting_code, setup_fee_revenue_schedule_type, setup_fee_liability_gl_account_id, setup_fee_revenue_gl_account_id, setup_fee_performance_obligation_id, trial_unit, trial_length, trial_requires_billing_info, total_billing_cycles, auto_renew, custom_fields, avalara_transaction_type, avalara_service_type, tax_code, harmonized_system_code, tax_exempt, vertex_transaction_type, hosted_pages, allow_any_item_on_subscriptions, dunning_campaign_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_plan:
    endpoint: DELETE /plans/{{ record.plan_id }}
    required fields: plan_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_plan_add_on:
    endpoint: POST /plans/{{ record.plan_id }}/add_ons
    required fields: plan_id, code, name
    optional fields: item_code, item_id, add_on_type, usage_type, usage_calculation_type, usage_percentage, measured_unit_id, measured_unit_name, liability_gl_account_id, revenue_gl_account_id, performance_obligation_id, accounting_code, revenue_schedule_type, display_quantity, default_quantity, optional, avalara_transaction_type, avalara_service_type, tax_code, harmonized_system_code, currencies, tier_type, usage_timeframe, tiers, percentage_tiers
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_plan_add_on:
    endpoint: PUT /plans/{{ record.plan_id }}/add_ons/{{ record.add_on_id }}
    required fields: plan_id, add_on_id
    optional fields: code, name, usage_percentage, usage_calculation_type, measured_unit_id, measured_unit_name, accounting_code, liability_gl_account_id, revenue_gl_account_id, performance_obligation_id, revenue_schedule_type, avalara_transaction_type, avalara_service_type, tax_code, harmonized_system_code, display_quantity, default_quantity, optional, currencies, tiers, percentage_tiers
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_plan_add_on:
    endpoint: DELETE /plans/{{ record.plan_id }}/add_ons/{{ record.add_on_id }}
    required fields: plan_id, add_on_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_shipping_method:
    endpoint: POST /shipping_methods
    required fields: code, name
    optional fields: accounting_code, tax_code, liability_gl_account_id, revenue_gl_account_id, performance_obligation_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_shipping_method:
    endpoint: PUT /shipping_methods/{{ record.shipping_method_id }}
    required fields: shipping_method_id
    optional fields: code, name, accounting_code, tax_code, liability_gl_account_id, revenue_gl_account_id, performance_obligation_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  deactivate_shipping_method:
    endpoint: DELETE /shipping_methods/{{ record.shipping_method_id }}
    required fields: shipping_method_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_subscription:
    endpoint: POST /subscriptions
    required fields: plan_code, currency, account
    optional fields: plan_id, business_entity_id, business_entity_code, price_segment_id, billing_info_id, shipping, collection_method, unit_amount, tax_inclusive, quantity, add_ons, coupon_codes, custom_fields, trial_ends_at, starts_at, next_bill_date, total_billing_cycles, renewal_billing_cycles, auto_renew, ramp_intervals, revenue_schedule_type, terms_and_conditions, customer_notes, credit_customer_notes, po_number, net_terms, net_terms_type, credit_application_policy, gateway_code, transaction_type, gift_card_redemption_code, bulk, proration_settings, transaction_descriptor_suffix
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_subscription:
    endpoint: PUT /subscriptions/{{ record.subscription_id }}
    required fields: subscription_id
    optional fields: collection_method, custom_fields, remaining_billing_cycles, renewal_billing_cycles, auto_renew, next_bill_date, revenue_schedule_type, terms_and_conditions, customer_notes, po_number, price_segment_id, net_terms, net_terms_type, credit_application_policy, gateway_code, tax_inclusive, shipping, billing_info_id, transaction_descriptor_suffix
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  terminate_subscription:
    endpoint: DELETE /subscriptions/{{ record.subscription_id }}
    required fields: subscription_id, refund, charge
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  cancel_subscription:
    endpoint: PUT /subscriptions/{{ record.subscription_id }}/cancel
    required fields: subscription_id
    optional fields: timeframe
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  reactivate_subscription:
    endpoint: PUT /subscriptions/{{ record.subscription_id }}/reactivate
    required fields: subscription_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  pause_subscription:
    endpoint: PUT /subscriptions/{{ record.subscription_id }}/pause
    required fields: subscription_id, remaining_pause_cycles
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  resume_subscription:
    endpoint: PUT /subscriptions/{{ record.subscription_id }}/resume
    required fields: subscription_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  convert_trial:
    endpoint: PUT /subscriptions/{{ record.subscription_id }}/convert_trial
    required fields: subscription_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_subscription_change:
    endpoint: POST /subscriptions/{{ record.subscription_id }}/change
    required fields: subscription_id
    optional fields: timeframe, plan_id, plan_code, business_entity_id, business_entity_code, price_segment_id, unit_amount, tax_inclusive, quantity, shipping, coupon_codes, add_ons, collection_method, revenue_schedule_type, custom_fields, po_number, net_terms, net_terms_type, transaction_type, billing_info, ramp_intervals, proration_settings, next_bill_date
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_subscription_change:
    endpoint: DELETE /subscriptions/{{ record.subscription_id }}/change
    required fields: subscription_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_subscription_coupon_redemption:
    endpoint: DELETE /subscriptions/{{ record.subscription_id }}/coupon_redemptions/{{ record.coupon_redemption_id }}
    required fields: subscription_id, coupon_redemption_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_usage:
    endpoint: POST /subscriptions/{{ record.subscription_id }}/add_ons/{{ record.add_on_id }}/usage
    required fields: subscription_id, add_on_id
    optional fields: merchant_tag, amount, recording_timestamp, usage_timestamp
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  update_usage:
    endpoint: PUT /usage/{{ record.usage_id }}
    required fields: usage_id
    optional fields: merchant_tag, amount, recording_timestamp, usage_timestamp
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  remove_usage:
    endpoint: DELETE /usage/{{ record.usage_id }}
    required fields: usage_id
    risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  deactivate_unique_coupon_code:
    endpoint: DELETE /unique_coupon_codes/{{ record.unique_coupon_code_id }}
    required fields: unique_coupon_code_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  reactivate_unique_coupon_code:
    endpoint: PUT /unique_coupon_codes/{{ record.unique_coupon_code_id }}/restore
    required fields: unique_coupon_code_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_purchase:
    endpoint: POST /purchases
    required fields: currency, account
    optional fields: billing_info_id, business_entity_id, business_entity_code, collection_method, po_number, net_terms, net_terms_type, credit_application_policy_override, terms_and_conditions, transaction, customer_notes, vat_reverse_charge_notes, vertex_transaction_type, credit_customer_notes, gateway_code, shipping, line_items, subscriptions, coupon_codes, gift_card_redemption_code, transaction_type
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_pending_purchase:
    endpoint: POST /purchases/pending
    required fields: currency, account
    optional fields: billing_info_id, business_entity_id, business_entity_code, collection_method, po_number, net_terms, net_terms_type, credit_application_policy_override, terms_and_conditions, transaction, customer_notes, vat_reverse_charge_notes, vertex_transaction_type, credit_customer_notes, gateway_code, shipping, line_items, subscriptions, coupon_codes, gift_card_redemption_code, transaction_type
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_authorize_purchase:
    endpoint: POST /purchases/authorize
    required fields: currency, account
    optional fields: billing_info_id, business_entity_id, business_entity_code, collection_method, po_number, net_terms, net_terms_type, credit_application_policy_override, terms_and_conditions, transaction, customer_notes, vat_reverse_charge_notes, vertex_transaction_type, credit_customer_notes, gateway_code, shipping, line_items, subscriptions, coupon_codes, gift_card_redemption_code, transaction_type
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_capture_purchase:
    endpoint: POST /purchases/{{ record.transaction_id }}/capture
    required fields: transaction_id
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  cancel_purchase:
    endpoint: POST /purchases/{{ record.transaction_id }}/cancel/
    required fields: transaction_id
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  put_dunning_campaign_bulk_update:
    endpoint: PUT /dunning_campaigns/{{ record.dunning_campaign_id }}/bulk_update
    required fields: dunning_campaign_id
    optional fields: plan_codes, plan_ids
    risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  create_gift_card:
    endpoint: POST /gift_cards
    required fields: product_code, unit_amount, currency, delivery, gifter_account
    optional fields: tax_service_opt_out
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.
  redeem_gift_card:
    endpoint: POST /gift_cards/{{ record.redemption_code }}/redeem
    required fields: redemption_code, recipient_account
    risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.

SECURITY
  read risk: external Recurly API reads can expose account, billing, invoice, subscription, usage, and transaction data; direct preview reads are fixed-path and redacted
  write risk: typed Recurly reverse ETL mutations for accounts, billing, subscriptions, invoices, catalog, usage, coupons, exports, and related lifecycle resources
  approval: reverse ETL writes require plan, preview, explicit approval, execute; destructive lifecycle actions require destructive confirmation; each Recurly mutation receives a stable per-record Idempotency-Key across automatic retries
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Recurly V3 API connector for accounts, subscriptions, invoices, transactions, catalog, previews, exports, and typed reverse ETL.
  Usage: pm recurly <command> [flags] --json
  Source CLI: Recurly API (https://recurly.com/developers/api/spec/v2021-02-25.yaml)
  PM execution policy pm-request-contract-bounds-v1: each max N bytes qualifier is the effective PM request limit, not a provider schema assertion; path/query values are measured after exact wire encoding and rejected rather than truncated.
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Sites
  Site
  Accounts
  Account
  Billing
  Billing Infos
  Active
  Coupon
  Shipping
  Child
  Coupons
  Unique
  Credit
  Custom
  General
  Performance Obligations
  Invoice
  Items
  Item
  Measured
  External
  Invoices
  Related
  Line
  Plans
  Plan
  Price
  Add
  Subscriptions
  Subscription
  Usage
  Transactions
  Transaction
  Dunning
  Entitlements
  Business
  Gift
  A
  Unique Coupon Codes
  External Subscriptions
  Purchase
  Purchases
  Dunning Campaigns
  Export
  Other Commands
    sites list - List Recurly sites as ETL records. [intent=etl availability=implemented stream=list_sites]
    site get - Get Recurly site as an ETL record. [intent=etl availability=implemented stream=get_site]
    accounts list - List Recurly accounts as ETL records. [intent=etl availability=implemented stream=accounts]
    account get - Get Recurly account as an ETL record. [intent=etl availability=implemented stream=get_account]
    account acquisition get - Get Recurly account acquisition as an ETL record. [intent=etl availability=implemented stream=get_account_acquisition]
    account balance get - Get Recurly account balance as an ETL record. [intent=etl availability=implemented stream=get_account_balance]
    accounts billing-infos get - Get Recurly billing info as an ETL record. [intent=etl availability=implemented stream=get_billing_info]
    billing infos list - List Recurly billing infos as ETL records. [intent=etl availability=implemented stream=list_billing_infos]
    billing-infos get - Get Recurly a billing info as an ETL record. [intent=etl availability=implemented stream=get_a_billing_info]
    account coupon redemptions list - List Recurly account coupon redemptions as ETL records. [intent=etl availability=implemented stream=list_account_coupon_redemptions]
    active coupon redemptions list - List Recurly active coupon redemptions as ETL records. [intent=etl availability=implemented stream=list_active_coupon_redemptions]
    coupon redemption get - Get Recurly coupon redemption as an ETL record. [intent=etl availability=implemented stream=get_coupon_redemption]
    account credit payments list - List Recurly account credit payments as ETL records. [intent=etl availability=implemented stream=list_account_credit_payments]
    account external account list - List Recurly account external account as ETL records. [intent=etl availability=implemented stream=list_account_external_account]
    account external account get - Get Recurly account external account as an ETL record. [intent=etl availability=implemented stream=get_account_external_account]
    account external invoices list - List Recurly account external invoices as ETL records. [intent=etl availability=implemented stream=list_account_external_invoices]
    account invoices list - List Recurly account invoices as ETL records. [intent=etl availability=implemented stream=list_account_invoices]
    account line items list - List Recurly account line items as ETL records. [intent=etl availability=implemented stream=list_account_line_items]
    account notes list - List Recurly account notes as ETL records. [intent=etl availability=implemented stream=list_account_notes]
    account note get - Get Recurly account note as an ETL record. [intent=etl availability=implemented stream=get_account_note]
    shipping addresses list - List Recurly shipping addresses as ETL records. [intent=etl availability=implemented stream=list_shipping_addresses]
    shipping address get - Get Recurly shipping address as an ETL record. [intent=etl availability=implemented stream=get_shipping_address]
    account subscriptions list - List Recurly account subscriptions as ETL records. [intent=etl availability=implemented stream=list_account_subscriptions]
    account transactions list - List Recurly account transactions as ETL records. [intent=etl availability=implemented stream=list_account_transactions]
    child accounts list - List Recurly child accounts as ETL records. [intent=etl availability=implemented stream=list_child_accounts]
    account acquisition list - List Recurly account acquisition as ETL records. [intent=etl availability=implemented stream=list_account_acquisition]
    coupons list - List Recurly coupons as ETL records. [intent=etl availability=implemented stream=list_coupons]
    coupon get - Get Recurly coupon as an ETL record. [intent=etl availability=implemented stream=get_coupon]
    unique coupon codes list - List Recurly unique coupon codes as ETL records. [intent=etl availability=implemented stream=list_unique_coupon_codes]
    credit payments list - List Recurly credit payments as ETL records. [intent=etl availability=implemented stream=list_credit_payments]
    credit payment get - Get Recurly credit payment as an ETL record. [intent=etl availability=implemented stream=get_credit_payment]
    custom field definitions list - List Recurly custom field definitions as ETL records. [intent=etl availability=implemented stream=list_custom_field_definitions]
    custom field definition get - Get Recurly custom field definition as an ETL record. [intent=etl availability=implemented stream=get_custom_field_definition]
    general ledger accounts list - List Recurly general ledger accounts as ETL records. [intent=etl availability=implemented stream=list_general_ledger_accounts]
    general ledger account get - Get Recurly general ledger account as an ETL record. [intent=etl availability=implemented stream=get_general_ledger_account]
    performance-obligations get - Get Recurly performance obligation as an ETL record. [intent=etl availability=implemented stream=get_performance_obligation]
    performance-obligations get-all - Get Recurly performance obligations as an ETL record. [intent=etl availability=implemented stream=get_performance_obligations]
    invoice template accounts list - List Recurly invoice template accounts as ETL records. [intent=etl availability=implemented stream=list_invoice_template_accounts]
    items list - List Recurly items as ETL records. [intent=etl availability=implemented stream=list_items]
    item get - Get Recurly item as an ETL record. [intent=etl availability=implemented stream=get_item]
    measured unit list - List Recurly measured unit as ETL records. [intent=etl availability=implemented stream=list_measured_unit]
    measured unit get - Get Recurly measured unit as an ETL record. [intent=etl availability=implemented stream=get_measured_unit]
    external products list - List Recurly external products as ETL records. [intent=etl availability=implemented stream=list_external_products]
    external product get - Get Recurly external product as an ETL record. [intent=etl availability=implemented stream=get_external_product]
    external product external product references list - List Recurly external product external product references as ETL records. [intent=etl availability=implemented stream=list_external_product_external_product_references]
    external product external product reference get - Get Recurly external product external product reference as an ETL record. [intent=etl availability=implemented stream=get_external_product_external_product_reference]
    external subscriptions list - List Recurly external subscriptions as ETL records. [intent=etl availability=implemented stream=list_external_subscriptions]
    external subscription get - Get Recurly external subscription as an ETL record. [intent=etl availability=implemented stream=get_external_subscription]
    external subscription external invoices list - List Recurly external subscription external invoices as ETL records. [intent=etl availability=implemented stream=list_external_subscription_external_invoices]
    invoices list - List Recurly invoices as ETL records. [intent=etl availability=implemented stream=invoices]
    invoice get - Get Recurly invoice as an ETL record. [intent=etl availability=implemented stream=get_invoice]
    invoice line items list - List Recurly invoice line items as ETL records. [intent=etl availability=implemented stream=list_invoice_line_items]
    invoice coupon redemptions list - List Recurly invoice coupon redemptions as ETL records. [intent=etl availability=implemented stream=list_invoice_coupon_redemptions]
    related invoices list - List Recurly related invoices as ETL records. [intent=etl availability=implemented stream=list_related_invoices]
    line items list - List Recurly line items as ETL records. [intent=etl availability=implemented stream=list_line_items]
    line item get - Get Recurly line item as an ETL record. [intent=etl availability=implemented stream=get_line_item]
    plans list - List Recurly plans as ETL records. [intent=etl availability=implemented stream=plans]
    plan get - Get Recurly plan as an ETL record. [intent=etl availability=implemented stream=get_plan]
    plan add ons list - List Recurly plan add ons as ETL records. [intent=etl availability=implemented stream=list_plan_add_ons]
    plan add on get - Get Recurly plan add on as an ETL record. [intent=etl availability=implemented stream=get_plan_add_on]
    price segments list - List Recurly price segments as ETL records. [intent=etl availability=implemented stream=list_price_segments]
    price segment get - Get Recurly price segment as an ETL record. [intent=etl availability=implemented stream=get_price_segment]
    add ons list - List Recurly add ons as ETL records. [intent=etl availability=implemented stream=list_add_ons]
    add on get - Get Recurly add on as an ETL record. [intent=etl availability=implemented stream=get_add_on]
    shipping methods list - List Recurly shipping methods as ETL records. [intent=etl availability=implemented stream=list_shipping_methods]
    shipping method get - Get Recurly shipping method as an ETL record. [intent=etl availability=implemented stream=get_shipping_method]
    subscriptions list - List Recurly subscriptions as ETL records. [intent=etl availability=implemented stream=subscriptions]
    subscription get - Get Recurly subscription as an ETL record. [intent=etl availability=implemented stream=get_subscription]
    subscription change get - Get Recurly subscription change as an ETL record. [intent=etl availability=implemented stream=get_subscription_change]
    subscription invoices list - List Recurly subscription invoices as ETL records. [intent=etl availability=implemented stream=list_subscription_invoices]
    subscription line items list - List Recurly subscription line items as ETL records. [intent=etl availability=implemented stream=list_subscription_line_items]
    subscription coupon redemptions list - List Recurly subscription coupon redemptions as ETL records. [intent=etl availability=implemented stream=list_subscription_coupon_redemptions]
    subscription coupon redemption get - Get Recurly subscription coupon redemption as an ETL record. [intent=etl availability=implemented stream=get_subscription_coupon_redemption]
    usage list - List Recurly usage as ETL records. [intent=etl availability=implemented stream=list_usage]
    usage get - Get Recurly usage as an ETL record. [intent=etl availability=implemented stream=get_usage]
    transactions list - List Recurly transactions as ETL records. [intent=etl availability=implemented stream=transactions]
    transaction get - Get Recurly transaction as an ETL record. [intent=etl availability=implemented stream=get_transaction]
    unique coupon code get - Get Recurly unique coupon code as an ETL record. [intent=etl availability=implemented stream=get_unique_coupon_code]
    dunning campaigns list - List Recurly dunning campaigns as ETL records. [intent=etl availability=implemented stream=list_dunning_campaigns]
    dunning campaign get - Get Recurly dunning campaign as an ETL record. [intent=etl availability=implemented stream=get_dunning_campaign]
    invoice templates list - List Recurly invoice templates as ETL records. [intent=etl availability=implemented stream=list_invoice_templates]
    invoice template get - Get Recurly invoice template as an ETL record. [intent=etl availability=implemented stream=get_invoice_template]
    external invoices list - List Recurly external invoices as ETL records. [intent=etl availability=implemented stream=list_external_invoices]
    external invoice get - Get Recurly  external invoice as an ETL record. [intent=etl availability=implemented stream=show_external_invoice]
    external subscription external payment phases list - List Recurly external subscription external payment phases as ETL records. [intent=etl availability=implemented stream=list_external_subscription_external_payment_phases]
    external subscription external payment phase get - Get Recurly external subscription external payment phase as an ETL record. [intent=etl availability=implemented stream=get_external_subscription_external_payment_phase]
    entitlements list - List Recurly entitlements as ETL records. [intent=etl availability=implemented stream=list_entitlements]
    account external subscriptions list - List Recurly account external subscriptions as ETL records. [intent=etl availability=implemented stream=list_account_external_subscriptions]
    business entity get - Get Recurly business entity as an ETL record. [intent=etl availability=implemented stream=get_business_entity]
    business entities list - List Recurly business entities as ETL records. [intent=etl availability=implemented stream=list_business_entities]
    gift cards list - List Recurly gift cards as ETL records. [intent=etl availability=implemented stream=list_gift_cards]
    gift card get - Get Recurly gift card as an ETL record. [intent=etl availability=implemented stream=get_gift_card]
    business entity invoices list - List Recurly business entity invoices as ETL records. [intent=etl availability=implemented stream=list_business_entity_invoices]
    account create - Create Recurly create account via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --code (required)
    account update - Update Recurly update account via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --email
    account deactivate - Delete Recurly deactivate account via typed reverse ETL. [intent=reverse_etl availability=implemented write=deactivate_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --redact (required)
    account redact - Update Recurly redact account via typed reverse ETL. [intent=reverse_etl availability=implemented write=redact_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required)
    account acquisition update - Update Recurly update account acquisition via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_account_acquisition]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --cost-currency, --channel
    account acquisition remove - Delete Recurly remove account acquisition via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_account_acquisition]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required)
    account reactivate - Update Recurly reactivate account via typed reverse ETL. [intent=reverse_etl availability=implemented write=reactivate_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required)
    billing info update - Update Recurly update billing info via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_billing_info]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --first-name
    billing info remove - Delete Recurly remove billing info via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_billing_info]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required)
    billing info verify - Create Recurly verify billing info via typed reverse ETL. [intent=reverse_etl availability=implemented write=verify_billing_info]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required)
    billing info cvv verify - Create Recurly verify billing info cvv via typed reverse ETL. [intent=reverse_etl availability=implemented write=verify_billing_info_cvv]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required)
    billing info create - Create Recurly create billing info via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_billing_info]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --first-name
    a billing info update - Update Recurly update a billing info via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_a_billing_info]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --billing-info-id (required), --first-name
    a billing info remove - Delete Recurly remove a billing info via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_a_billing_info]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --billing-info-id (required)
    billing infos verify - Create Recurly verify billing infos via typed reverse ETL. [intent=reverse_etl availability=implemented write=verify_billing_infos]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --billing-info-id (required)
    billing infos cvv verify - Create Recurly verify billing infos cvv via typed reverse ETL. [intent=reverse_etl availability=implemented write=verify_billing_infos_cvv]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --billing-info-id (required)
    coupon redemption create - Create Recurly create coupon redemption via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_coupon_redemption]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --coupon-id (required)
    coupon redemption remove - Delete Recurly remove coupon redemption via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_coupon_redemption]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required)
    coupon redemption by id remove - Delete Recurly remove coupon redemption by id via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_coupon_redemption_by_id]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --coupon-redemption-id (required)
    account external account create - Create Recurly create account external account via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_account_external_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --external-account-code (required), --external-connection-type (required)
    account external account update - Update Recurly update account external account via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_account_external_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --external-account-id (required)
    account external account delete - Delete Recurly delete account external account via typed reverse ETL. [intent=reverse_etl availability=implemented write=delete_account_external_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --external-account-id (required)
    invoice create - Create Recurly create invoice via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_invoice]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --currency (required)
    line item create - Create Recurly create line item via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_line_item]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --currency (required), --type (required), --unit-amount (required)
    account note create - Create Recurly create account note via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_account_note]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --message (required)
    account note remove - Delete Recurly remove account note via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_account_note]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --account-note-id (required)
    shipping address create - Create Recurly create shipping address via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_shipping_address]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --city (required), --country (required), --first-name (required), --last-name (required), --postal-code (required), --street1 (required)
    shipping address update - Update Recurly update shipping address via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_shipping_address]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --shipping-address-id (required), --first-name
    shipping address remove - Delete Recurly remove shipping address via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_shipping_address]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-id (required), --shipping-address-id (required)
    coupon create - Create Recurly create coupon via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_coupon]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --code (required), --discount-type (required), --name
    coupon update - Update Recurly update coupon via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_coupon]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --coupon-id (required), --name
    coupon deactivate - Delete Recurly deactivate coupon via typed reverse ETL. [intent=reverse_etl availability=implemented write=deactivate_coupon]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --coupon-id (required)
    unique-coupon-codes generate - Create Recurly generate unique coupon codes via typed reverse ETL. [intent=reverse_etl availability=implemented write=generate_unique_coupon_codes]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --coupon-id (required), --number-of-unique-codes
    unique-coupon-codes generate-sync - Create Recurly generate unique coupon codes sync via typed reverse ETL. [intent=reverse_etl availability=implemented write=generate_unique_coupon_codes_sync]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --coupon-id (required), --number-of-unique-codes (required)
    coupon restore - Update Recurly restore coupon via typed reverse ETL. [intent=reverse_etl availability=implemented write=restore_coupon]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --coupon-id (required), --name
    general ledger account create - Create Recurly create general ledger account via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_general_ledger_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-type, --code
    general ledger account update - Update Recurly update general ledger account via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_general_ledger_account]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --general-ledger-account-id (required), --code
    item create - Create Recurly create item via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_item]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --code (required), --name (required)
    item update - Update Recurly update item via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_item]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --item-id (required), --code
    item deactivate - Delete Recurly deactivate item via typed reverse ETL. [intent=reverse_etl availability=implemented write=deactivate_item]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --item-id (required)
    item reactivate - Update Recurly reactivate item via typed reverse ETL. [intent=reverse_etl availability=implemented write=reactivate_item]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --item-id (required)
    measured unit create - Create Recurly create measured unit via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_measured_unit]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --display-name (required), --name (required)
    measured unit update - Update Recurly update measured unit via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_measured_unit]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --measured-unit-id (required), --name
    measured unit remove - Delete Recurly remove measured unit via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_measured_unit]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --measured-unit-id (required)
    external product create - Create Recurly create external product via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_external_product]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --name (required)
    external product update - Update Recurly update external product via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_external_product]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --external-product-id (required), --plan-id (required)
    external products deactivate - Delete Recurly deactivate external products via typed reverse ETL. [intent=reverse_etl availability=implemented write=deactivate_external_products]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --external-product-id (required)
    external product external product reference create - Create Recurly create external product external product reference via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_external_product_external_product_reference]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --external-connection-type, --external-product-id (required), --reference-code
    external product external product reference deactivate - Delete Recurly deactivate external product external product reference via typed reverse ETL. [intent=reverse_etl availability=implemented write=deactivate_external_product_external_product_reference]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --external-product-id (required), --external-product-reference-id (required)
    external subscription create - Create Recurly create external subscription via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_external_subscription]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --activated-at (format=date-time), --expires-at (format=date-time), --external-id, --quantity
    external-subscriptions put - Update Recurly put external subscription via typed reverse ETL. [intent=reverse_etl availability=implemented write=put_external_subscription]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --external-subscription-id (required)
    external invoice create - Create Recurly create external invoice via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_external_invoice]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --currency (required), --external-id (required), --external-subscription-id (required), --purchased-at (required, format=date-time), --state (required), --total (required)
    invoice update - Update Recurly update invoice via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_invoice]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --invoice-id (required), --address-first-name, --po-number
    invoices apply-credit-balance - Update Recurly apply credit balance via typed reverse ETL. [intent=reverse_etl availability=implemented write=apply_credit_balance]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --invoice-id (required)
    invoice collect - Update Recurly collect invoice via typed reverse ETL. [intent=reverse_etl availability=implemented write=collect_invoice]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --invoice-id (required)
    invoice failed mark - Update Recurly mark invoice failed via typed reverse ETL. [intent=reverse_etl availability=implemented write=mark_invoice_failed]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --invoice-id (required)
    invoice successful mark - Update Recurly mark invoice successful via typed reverse ETL. [intent=reverse_etl availability=implemented write=mark_invoice_successful]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --invoice-id (required)
    invoice reopen - Update Recurly reopen invoice via typed reverse ETL. [intent=reverse_etl availability=implemented write=reopen_invoice]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --invoice-id (required)
    invoice void - Update Recurly void invoice via typed reverse ETL. [intent=reverse_etl availability=implemented write=void_invoice]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --invoice-id (required)
    invoices record-external-transaction - Create Recurly record external transaction via typed reverse ETL. [intent=reverse_etl availability=implemented write=record_external_transaction]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --invoice-id (required), --amount
    invoice refund - Create Recurly refund invoice via typed reverse ETL. [intent=reverse_etl availability=implemented write=refund_invoice]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — refund_invoice moves money by refunding an invoice; requires destructive confirmation and reverse ETL plan/preview/approval/execute. Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --invoice-id (required), --type (required)
    invoices retries create - Create Recurly create invoice retry via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_invoice_retry]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-billing-infos-0-gateway-code (required), --account-billing-infos-0-payment-gateway-references-0-reference-type (required), --account-billing-infos-0-transactions-0-attempted-collection-date (required, format=date-time), --account-billing-infos-0-transactions-0-gateway-error-code (required), --account-code (required), --currency (required), --due-at (required, format=date-time), --external-recovery-eligible (required), --line-items-0-unit-amount (required)
    line item remove - Delete Recurly remove line item via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_line_item]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --line-item-id (required)
    plan create - Create Recurly create plan via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_plan]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --code (required), --currencies-0-currency (required), --name (required)
    plan update - Update Recurly update plan via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_plan]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --plan-id (required), --code
    plan remove - Delete Recurly remove plan via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_plan]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --plan-id (required)
    plan add on create - Create Recurly create plan add on via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_plan_add_on]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --code (required), --name (required), --plan-id (required)
    plan add on update - Update Recurly update plan add on via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_plan_add_on]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --add-on-id (required), --plan-id (required), --code
    plan add on remove - Delete Recurly remove plan add on via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_plan_add_on]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --add-on-id (required), --plan-id (required)
    shipping method create - Create Recurly create shipping method via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_shipping_method]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --code (required), --name (required)
    shipping method update - Update Recurly update shipping method via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_shipping_method]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --shipping-method-id (required), --code
    shipping method deactivate - Delete Recurly deactivate shipping method via typed reverse ETL. [intent=reverse_etl availability=implemented write=deactivate_shipping_method]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --shipping-method-id (required)
    subscription create - Create Recurly create subscription via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_subscription]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-code (required), --currency (required), --plan-code (required)
    subscription update - Update Recurly update subscription via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_subscription]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --subscription-id (required), --collection-method
    subscription terminate - Delete Recurly terminate subscription via typed reverse ETL. [intent=reverse_etl availability=implemented write=terminate_subscription]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --subscription-id (required), --refund (required), --charge (required)
    subscription cancel - Update Recurly cancel subscription via typed reverse ETL. [intent=reverse_etl availability=implemented write=cancel_subscription]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --subscription-id (required)
    subscription reactivate - Update Recurly reactivate subscription via typed reverse ETL. [intent=reverse_etl availability=implemented write=reactivate_subscription]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --subscription-id (required)
    subscription pause - Update Recurly pause subscription via typed reverse ETL. [intent=reverse_etl availability=implemented write=pause_subscription]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --remaining-pause-cycles (required), --subscription-id (required)
    subscription resume - Update Recurly resume subscription via typed reverse ETL. [intent=reverse_etl availability=implemented write=resume_subscription]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --subscription-id (required)
    subscriptions convert-trial - Update Recurly convert trial via typed reverse ETL. [intent=reverse_etl availability=implemented write=convert_trial]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --subscription-id (required)
    subscription change create - Create Recurly create subscription change via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_subscription_change]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --subscription-id (required), --plan-code
    subscription change remove - Delete Recurly remove subscription change via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_subscription_change]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --subscription-id (required)
    subscription coupon redemption remove - Delete Recurly remove subscription coupon redemption via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_subscription_coupon_redemption]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --coupon-redemption-id (required), --subscription-id (required)
    usage create - Create Recurly create usage via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_usage]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --add-on-id (required), --subscription-id (required), --amount
    usage update - Update Recurly update usage via typed reverse ETL. [intent=reverse_etl availability=implemented write=update_usage]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --usage-id (required), --amount
    usage remove - Delete Recurly remove usage via typed reverse ETL. [intent=reverse_etl availability=implemented write=remove_usage]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: critical — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --usage-id (required)
    unique coupon code deactivate - Delete Recurly deactivate unique coupon code via typed reverse ETL. [intent=reverse_etl availability=implemented write=deactivate_unique_coupon_code]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --unique-coupon-code-id (required)
    unique coupon code reactivate - Update Recurly reactivate unique coupon code via typed reverse ETL. [intent=reverse_etl availability=implemented write=reactivate_unique_coupon_code]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --unique-coupon-code-id (required)
    purchase create - Create Recurly create purchase via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_purchase]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-code (required), --currency (required)
    purchases create-pending - Create Recurly create pending purchase via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_pending_purchase]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-code (required), --currency (required)
    purchases authorize - Create Recurly create authorize purchase via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_authorize_purchase]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --account-code (required), --currency (required)
    purchases capture - Create Recurly create capture purchase via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_capture_purchase]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --transaction-id (required)
    purchases cancel - Create Recurly cancel purchase via typed reverse ETL. [intent=reverse_etl availability=implemented write=cancel_purchase]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --transaction-id (required)
    dunning-campaigns bulk-update - Update Recurly put dunning campaign bulk update via typed reverse ETL. [intent=reverse_etl availability=implemented write=put_dunning_campaign_bulk_update]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: high — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --dunning-campaign-id (required)
    gift card create - Create Recurly create gift card via typed reverse ETL. [intent=reverse_etl availability=implemented write=create_gift_card]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --currency (required), --product-code (required), --unit-amount (required), --delivery-method (required), --gifter-account-code (required)
    gift card redeem - Create Recurly redeem gift card via typed reverse ETL. [intent=reverse_etl availability=implemented write=redeem_gift_card]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive lifecycle actions require destructive confirmation.; risk: medium — Recurly supports provider idempotency for POST/PUT/DELETE through the Idempotency-Key header; keep reverse ETL in plan/preview/approve/execute and the runtime binds one fresh key to each approved record and reuses it only for that record's automatic retries.; flags: --recipient-account-code (required), --redemption-code (required)
    invoices preview - Preview new invoice for pending line items [intent=direct_read availability=implemented operation=preview_invoice]; notes: Bounded Recurly read/preview operation; fixed method and path, typed request fields, and redacted JSON output.; flags: --currency (required), --account-id (required, max 4096 bytes), --page, --page-cursor
    invoice pdf get - Fetch an invoice as a PDF [intent=binary_download availability=implemented operation=get_invoice_pdf]; notes: Bounded Recurly download; the provider response is written only beneath an explicit --dest-root.; flags: --invoice-id (required, max 4096 bytes), --dest-root (required), --file-name, --max-bytes
    subscriptions preview renewal - Fetch a preview of a subscription's renewal invoice(s) [intent=direct_read availability=implemented operation=get_preview_renewal]; notes: Bounded Recurly read/preview operation; fixed method and path, typed request fields, and redacted JSON output.; flags: --subscription-id (required, max 4096 bytes), --page, --page-cursor
    subscriptions preview change - Preview a new subscription change [intent=direct_read availability=implemented operation=preview_subscription_change]; notes: Bounded Recurly read/preview operation; fixed method and path, typed request fields, and redacted JSON output.; flags: --subscription-id (required, max 4096 bytes), --page, --page-cursor
    purchases preview - Preview a new purchase [intent=direct_read availability=implemented operation=preview_purchase]; notes: Bounded Recurly read/preview operation; fixed method and path, typed request fields, and redacted JSON output.; flags: --currency (required), --account-code (required), --page, --page-cursor
    export dates get - List the dates that have an available export to download. [intent=binary_download availability=implemented operation=get_export_dates]; notes: Bounded Recurly download; the provider response is written only beneath an explicit --dest-root.; flags: --dest-root (required), --file-name, --max-bytes
    export files get - List of the export files that are available to download. [intent=binary_download availability=implemented operation=get_export_files]; notes: Bounded Recurly download; the provider response is written only beneath an explicit --dest-root.; flags: --export-date (required, max 4096 bytes), --dest-root (required), --file-name, --max-bytes
    gift cards preview - Preview gift card [intent=direct_read availability=implemented operation=preview_gift_card]; notes: Bounded Recurly read/preview operation; fixed method and path, typed request fields, and redacted JSON output.; flags: --product-code (required), --unit-amount (required), --currency (required), --page, --page-cursor
  Help topics:
    safety - Reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.
    parity - Operation ledger partitions all 197 Recurly v2021-02-25 OpenAPI operations once.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect recurly

  # Inspect as structured JSON
  pm connectors inspect recurly --json

AGENT WORKFLOW
  - Run pm connectors inspect recurly before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
