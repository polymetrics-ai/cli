#!/usr/bin/env python3
"""Apply the captain-ordered Recurly parity fixes to writes.json:

1. Give the five required-body writes real JSON bodies (body_type json +
   self-contained record_schema capturing the official request shape):
   update_account (AccountUpdate), create_billing_info (BillingInfoCreate),
   create_usage / update_usage (UsageCreate), update_subscription
   (SubscriptionUpdate).
2. Expand create_account's record_schema to allow billing_info, address,
   custom_fields, company, and the rest of AccountCreate/AccountUpdate.
3. Expand refund_invoice's record_schema to include amount/percentage/
   line_items/refund_method/credit_customer_notes/external_refund AND mark it
   confirm: destructive.

Only the dialect keywords the engine compiles are used:
type/required/properties/items/enum/pattern/minProperties/additionalProperties/
x-secret/x-primary-key/x-cursor-field plus annotations format/default/title/
description.

Rewrite writes.json to a byte-stable indent=2 dump so untouched actions keep
their formatting.
"""

import json

PATH = "internal/connectors/defs/recurly/writes.json"

ADDRESS = {
    "type": "object",
    "properties": {
        "phone": {"type": "string", "description": "Phone number."},
        "street1": {"type": "string", "description": "Street address line 1."},
        "street2": {"type": "string", "description": "Street address line 2."},
        "city": {"type": "string", "description": "City."},
        "region": {"type": "string", "description": "State or province."},
        "postal_code": {"type": "string", "description": "Zip or postal code."},
        "country": {"type": "string", "description": "Country, 2-letter ISO 3166-1 alpha-2 code."},
        "geo_code": {"type": "string", "description": "Geographic entity code (Vertex/Avalara)."},
    },
    "additionalProperties": False,
}

CUSTOM_FIELDS = {
    "type": "array",
    "description": "Custom fields; sending an empty array removes none, sending a name with null/empty removes that field.",
    "items": {
        "type": "object",
        "properties": {
            "name": {"type": "string", "description": "The custom field's name."},
            "value": {"type": "string", "description": "The custom field's value; null/empty removes an existing value."},
            "created_at": {"type": "string", "format": "date-time"},
            "updated_at": {"type": "string", "format": "date-time"},
        },
        "required": ["name"],
        "additionalProperties": False,
    },
}

BILLING_INFO_CREATE = {
    "type": "object",
    "properties": {
        "token_id": {"type": "string", "description": "Recurly token representing the billing info."},
        "first_name": {"type": "string"},
        "last_name": {"type": "string"},
        "company": {"type": "string"},
        "address": ADDRESS,
        "number": {"type": "string", "description": "Credit card number (write-only)."},
        "month": {"type": "string", "description": "Credit card expiration month."},
        "year": {"type": "string", "description": "Credit card expiration year."},
        "cvv": {"type": "string", "description": "Credit card verification value (write-only)."},
        "currency": {"type": "string", "description": "3-letter ISO 4217 currency code."},
        "vat_number": {"type": "string"},
        "ip_address": {"type": "string"},
        "gateway_token": {"type": "string"},
        "gateway_code": {"type": "string", "description": "Gateway-specific identifier."},
        "payment_gateway_references": {
            "type": "array",
            "items": {
                "type": "object",
                "properties": {
                    "reference_type": {"type": "string", "enum": ["stripe_confirmation_token", "upi_vpa"]},
                    "token": {"type": "string"},
                },
                "required": ["reference_type"],
                "additionalProperties": False,
            },
        },
        "gateway_attributes": {"type": "object", "additionalProperties": True},
        "amazon_billing_agreement_id": {"type": "string"},
        "paypal_billing_agreement_id": {"type": "string"},
        "roku_billing_agreement_id": {"type": "string"},
        "fraud_session_id": {"type": "string"},
        "adyen_risk_profile_reference_id": {"type": "string"},
        "transaction_type": {"type": "string"},
        "three_d_secure_action_result_token_id": {"type": "string"},
        "iban": {"type": "string"},
        "name_on_account": {"type": "string"},
        "account_number": {"type": "string"},
        "routing_number": {"type": "string"},
        "sort_code": {"type": "string"},
        "type": {"type": "string", "description": "Billing info type."},
        "account_type": {"type": "string"},
        "tax_identifier": {"type": "string"},
        "tax_identifier_type": {"type": "string"},
        "primary_payment_method": {"type": "boolean"},
        "backup_payment_method": {"type": "boolean"},
        "external_hpp_type": {"type": "string"},
        "online_banking_payment_type": {"type": "string"},
        "card_type": {"type": "string"},
        "card_network_preference": {"type": "string"},
        "return_url": {"type": "string"},
        "authentication_method": {"type": "string"},
    },
    "additionalProperties": False,
}

SUBSCRIPTION_SHIPPING_UPDATE = {
    "type": "object",
    "properties": {
        "object": {"type": "string"},
        "address": {
            "type": "object",
            "properties": {
                "id": {"type": "string"},
                "first_name": {"type": "string"},
                "last_name": {"type": "string"},
                "phone": {"type": "string"},
                "email": {"type": "string"},
                "address1": {"type": "string"},
                "address2": {"type": "string"},
                "city": {"type": "string"},
                "state": {"type": "string"},
                "zip": {"type": "string"},
                "country": {"type": "string"},
                "company": {"type": "string"},
            },
            "additionalProperties": False,
        },
        "address_id": {"type": "string", "description": "Assign a shipping address from the account's existing shipping addresses."},
    },
    "additionalProperties": False,
}

LINE_ITEM_REFUND = {
    "type": "object",
    "properties": {
        "id": {"type": "string", "description": "Line item ID."},
        "quantity": {"type": "integer", "description": "Line item quantity to refund."},
        "quantity_decimal": {"type": "string", "description": "Decimal quantity to refund."},
        "amount": {"type": "number", "description": "Specific amount to refund from the adjustment."},
        "percentage": {"type": "integer", "description": "Percentage of the adjustment's remaining balance to refund (1-100)."},
        "prorate": {"type": "boolean", "default": False},
    },
    "additionalProperties": False,
}

# --- AccountUpdate body (used for update_account and create_account) -------
ACCOUNT_UPDATE_BODY = {
    "username": {"type": "string"},
    "email": {"type": "string"},
    "preferred_locale": {"type": "string"},
    "preferred_time_zone": {"type": "string"},
    "cc_emails": {"type": "string"},
    "first_name": {"type": "string"},
    "last_name": {"type": "string"},
    "company": {"type": "string"},
    "vat_number": {"type": "string"},
    "tax_exempt": {"type": "boolean"},
    "exemption_certificate": {"type": "string"},
    "override_business_entity_id": {"type": "string"},
    "parent_account_code": {"type": "string"},
    "parent_account_id": {"type": "string"},
    "bill_to": {"type": "string"},
    "transaction_type": {"type": "string"},
    "dunning_campaign_id": {"type": "string"},
    "invoice_template_id": {"type": "string"},
    "address": ADDRESS,
    "billing_info": BILLING_INFO_CREATE,
    "custom_fields": CUSTOM_FIELDS,
    "entity_use_code": {"type": "string"},
    "bill_date": {"type": "string"},
}
ACCOUNT_UPDATE_BODY_ORDER = [
    "username", "email", "preferred_locale", "preferred_time_zone", "cc_emails",
    "first_name", "last_name", "company", "vat_number", "tax_exempt",
    "exemption_certificate", "override_business_entity_id", "parent_account_code",
    "parent_account_id", "bill_to", "transaction_type", "dunning_campaign_id",
    "invoice_template_id", "address", "billing_info", "custom_fields",
    "entity_use_code", "bill_date",
]

SUBSCRIPTION_UPDATE_BODY = {
    "collection_method": {"type": "string"},
    "custom_fields": CUSTOM_FIELDS,
    "remaining_billing_cycles": {"type": "integer"},
    "renewal_billing_cycles": {"type": "integer"},
    "auto_renew": {"type": "boolean"},
    "next_bill_date": {"type": "string"},
    "revenue_schedule_type": {"type": "string"},
    "terms_and_conditions": {"type": "string"},
    "customer_notes": {"type": "string"},
    "po_number": {"type": "string"},
    "price_segment_id": {"type": "string"},
    "net_terms": {"type": "integer"},
    "net_terms_type": {"type": "string"},
    "credit_application_policy": {"type": "string"},
    "gateway_code": {"type": "string"},
    "tax_inclusive": {"type": "boolean"},
    "shipping": SUBSCRIPTION_SHIPPING_UPDATE,
    "billing_info_id": {"type": "string"},
    "transaction_descriptor_suffix": {"type": "string"},
}
SUBSCRIPTION_UPDATE_BODY_ORDER = [
    "collection_method", "custom_fields", "remaining_billing_cycles",
    "renewal_billing_cycles", "auto_renew", "next_bill_date",
    "revenue_schedule_type", "terms_and_conditions", "customer_notes",
    "po_number", "price_segment_id", "net_terms", "net_terms_type",
    "credit_application_policy", "gateway_code", "tax_inclusive", "shipping",
    "billing_info_id", "transaction_descriptor_suffix",
]

BILLING_INFO_CREATE_ORDER = list(BILLING_INFO_CREATE["properties"].keys())
USAGE_CREATE_BODY = {
    "merchant_tag": {"type": "string", "description": "Custom field used to identify the usage record."},
    "amount": {"type": "number", "description": "Used to determine the unit amount of the usage to record."},
    "recording_timestamp": {"type": "string", "description": "When the usage was recorded."},
    "usage_timestamp": {"type": "string", "description": "When the usage actually occurred."},
}
USAGE_CREATE_ORDER = ["merchant_tag", "amount", "recording_timestamp", "usage_timestamp"]

INVOICE_REFUND_BODY = {
    "type": {
        "type": "string",
        "enum": ["amount", "percentage", "line_items"],
        "description": "The type of refund. Amount and line items cannot both be specified in the request.",
    },
    "amount": {"type": "number", "description": "The total amount to refund. Must be used when type is `amount`."},
    "percentage": {"type": "integer", "description": "The percentage of the invoice's remaining balance to refund. Must be used when type is `percentage`."},
    "line_items": {
        "type": "array",
        "description": "Line items to refund. Must be used when type is `line_items`.",
        "items": LINE_ITEM_REFUND,
    },
    "refund_method": {"type": "string", "description": "How to refund: `all_credit`, `all_transaction`, or `transaction_first`."},
    "credit_customer_notes": {"type": "string", "description": "Notes to append to the credit invoice."},
    "external_refund": {
        "type": "object",
        "description": "Indicates the refund was settled outside of Recurly and a manual transaction should be created to track it.",
        "properties": {
            "payment_method": {"type": "string"},
            "reference": {"type": "string"},
            "description": {"type": "string"},
            "recorded_at": {"type": "string", "format": "date-time"},
        },
        "additionalProperties": True,
    },
}
INVOICE_REFUND_ORDER = ["type", "amount", "percentage", "line_items", "refund_method", "credit_customer_notes", "external_refund"]


def make_schema(props, required, order):
    return {
        "type": "object",
        "properties": {k: props[k] for k in order},
        "required": required,
        "additionalProperties": False,
    }


def main():
    with open(PATH) as f:
        data = json.load(f)
    actions = {a["name"]: a for a in data["actions"]}

    # ---- update_account (AccountUpdate body) ----
    ua = actions["update_account"]
    ua["body_type"] = "json"
    ua["record_schema"] = make_schema(
        dict(ACCOUNT_UPDATE_BODY, **{"account_id": {"type": "string", "description": "Path parameter `account_id`."}}),
        ["account_id"], ["account_id"] + ACCOUNT_UPDATE_BODY_ORDER)
    ua["body_fields"] = list(ACCOUNT_UPDATE_BODY_ORDER)

    # ---- create_billing_info (BillingInfoCreate body) ----
    cbi = actions["create_billing_info"]
    cbi["body_type"] = "json"
    cbi["record_schema"] = make_schema(
        dict(BILLING_INFO_CREATE["properties"], **{"account_id": {"type": "string", "description": "Path parameter `account_id`."}}),
        ["account_id"], ["account_id"] + BILLING_INFO_CREATE_ORDER)
    cbi["body_fields"] = list(BILLING_INFO_CREATE_ORDER)

    # ---- create_usage / update_usage (UsageCreate body) ----
    cu = actions["create_usage"]
    cu["body_type"] = "json"
    cu["record_schema"] = make_schema(
        dict(USAGE_CREATE_BODY, **{
            "subscription_id": {"type": "string", "description": "Path parameter `subscription_id`."},
            "add_on_id": {"type": "string", "description": "Path parameter `add_on_id`."},
        }),
        ["subscription_id", "add_on_id"], ["subscription_id", "add_on_id"] + USAGE_CREATE_ORDER)
    cu["body_fields"] = list(USAGE_CREATE_ORDER)

    uu = actions["update_usage"]
    uu["body_type"] = "json"
    uu["record_schema"] = make_schema(
        dict(USAGE_CREATE_BODY, **{"usage_id": {"type": "string", "description": "Path parameter `usage_id`."}}),
        ["usage_id"], ["usage_id"] + USAGE_CREATE_ORDER)
    uu["body_fields"] = list(USAGE_CREATE_ORDER)

    # ---- update_subscription (SubscriptionUpdate body) ----
    us = actions["update_subscription"]
    us["body_type"] = "json"
    us["record_schema"] = make_schema(
        dict(SUBSCRIPTION_UPDATE_BODY, **{"subscription_id": {"type": "string", "description": "Path parameter `subscription_id`."}}),
        ["subscription_id"], ["subscription_id"] + SUBSCRIPTION_UPDATE_BODY_ORDER)
    us["body_fields"] = list(SUBSCRIPTION_UPDATE_BODY_ORDER)

    # ---- create_account (AccountCreate = code + AccountUpdate body) ----
    ca = actions["create_account"]
    ca["record_schema"] = make_schema(
        dict({"code": {"type": "string", "description": "The unique identifier of the account. This cannot be changed once the account is created."}},
             **ACCOUNT_UPDATE_BODY),
        ["code"], ["code"] + ACCOUNT_UPDATE_BODY_ORDER)
    ca["body_fields"] = ["code"] + list(ACCOUNT_UPDATE_BODY_ORDER)

    # ---- refund_invoice (InvoiceRefund body + destructive confirm) ----
    ri = actions["refund_invoice"]
    ri["record_schema"] = make_schema(
        dict(INVOICE_REFUND_BODY, **{"invoice_id": {"type": "string", "description": "Path parameter `invoice_id`."}}),
        ["invoice_id", "type"], ["invoice_id"] + INVOICE_REFUND_ORDER)
    ri["body_fields"] = list(INVOICE_REFUND_ORDER)
    ri["confirm"] = "destructive"
    # escalate risk for this money-moving action
    ri["risk"] = ("critical \u2014 refund_invoice moves money by refunding an invoice; requires destructive "
                  "confirmation and reverse ETL plan/preview/approval/execute. Recurly supports provider "
                  "idempotency for POST/PUT/DELETE through the Idempotency-Key header; do not reuse "
                  "idempotency keys across different records.")

    with open(PATH, "w") as f:
        json.dump(data, f, indent=2)
        f.write("\n")
    print("updated 7 actions in", PATH)


if __name__ == "__main__":
    main()
