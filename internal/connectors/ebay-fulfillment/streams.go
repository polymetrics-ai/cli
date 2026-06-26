package ebayfulfillment

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the eBay Fulfillment API resource path
// (relative to api_host) it reads from, the JSON path to the records array in
// the response body, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "sell/fulfillment/v1/order").
	resource string
	// recordsPath is the dotted JSON path to the records array (e.g. "orders").
	recordsPath string
	// mapRecord flattens a raw API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams(); the read path is fully
// data-driven from this table.
//
// order_line_items is a sub-projection of the orders endpoint: each order's
// lineItems array is exploded into individual records by the read loop, so it
// shares the orders resource but carries a distinct mapper.
var streamEndpoints = map[string]streamEndpoint{
	"orders":                {resource: "sell/fulfillment/v1/order", recordsPath: "orders", mapRecord: orderRecord},
	"order_line_items":      {resource: "sell/fulfillment/v1/order", recordsPath: "orders", mapRecord: orderRecord},
	"shipping_fulfillments": {resource: "sell/fulfillment/v1/order", recordsPath: "orders", mapRecord: orderRecord},
	"payment_disputes":      {resource: "sell/fulfillment/v1/payment_dispute", recordsPath: "paymentDisputeSummaries", mapRecord: paymentDisputeRecord},
}

// streams returns the connector's published stream catalog. eBay only supports
// full_refresh for these resources, but creationDate / openDate give a natural
// incremental cursor we expose for downstream incremental syncs.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "orders",
			Description:  "eBay seller orders from the Fulfillment API (getOrders).",
			PrimaryKey:   []string{"order_id"},
			CursorFields: []string{"creation_date"},
			Fields:       orderFields(),
		},
		{
			Name:         "order_line_items",
			Description:  "Individual line items exploded from each eBay order.",
			PrimaryKey:   []string{"line_item_id"},
			CursorFields: []string{"creation_date"},
			Fields:       orderLineItemFields(),
		},
		{
			Name:         "shipping_fulfillments",
			Description:  "Shipment / fulfillment status derived from eBay orders.",
			PrimaryKey:   []string{"order_id"},
			CursorFields: []string{"creation_date"},
			Fields:       shippingFulfillmentFields(),
		},
		{
			Name:         "payment_disputes",
			Description:  "eBay payment dispute summaries from the Fulfillment API.",
			PrimaryKey:   []string{"payment_dispute_id"},
			CursorFields: []string{"open_date"},
			Fields:       paymentDisputeFields(),
		},
	}
}

func orderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "order_id", Type: "string"},
		{Name: "legacy_order_id", Type: "string"},
		{Name: "creation_date", Type: "timestamp"},
		{Name: "last_modified_date", Type: "timestamp"},
		{Name: "order_fulfillment_status", Type: "string"},
		{Name: "order_payment_status", Type: "string"},
		{Name: "seller_id", Type: "string"},
		{Name: "buyer_username", Type: "string"},
		{Name: "sales_record_reference", Type: "string"},
		{Name: "total_value", Type: "string"},
		{Name: "total_currency", Type: "string"},
		{Name: "line_item_count", Type: "integer"},
	}
}

func orderLineItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "line_item_id", Type: "string"},
		{Name: "order_id", Type: "string"},
		{Name: "legacy_item_id", Type: "string"},
		{Name: "sku", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "quantity", Type: "integer"},
		{Name: "line_item_fulfillment_status", Type: "string"},
		{Name: "total_value", Type: "string"},
		{Name: "total_currency", Type: "string"},
		{Name: "creation_date", Type: "timestamp"},
	}
}

func shippingFulfillmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "order_id", Type: "string"},
		{Name: "legacy_order_id", Type: "string"},
		{Name: "order_fulfillment_status", Type: "string"},
		{Name: "shipping_step", Type: "string"},
		{Name: "ship_to_name", Type: "string"},
		{Name: "ship_to_city", Type: "string"},
		{Name: "ship_to_state_or_province", Type: "string"},
		{Name: "ship_to_postal_code", Type: "string"},
		{Name: "ship_to_country_code", Type: "string"},
		{Name: "creation_date", Type: "timestamp"},
	}
}

func paymentDisputeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "payment_dispute_id", Type: "string"},
		{Name: "order_id", Type: "string"},
		{Name: "dispute_state", Type: "string"},
		{Name: "dispute_status", Type: "string"},
		{Name: "reason", Type: "string"},
		{Name: "open_date", Type: "timestamp"},
		{Name: "amount_value", Type: "string"},
		{Name: "amount_currency", Type: "string"},
		{Name: "buyer_username", Type: "string"},
	}
}

// orderRecord flattens a raw eBay order object into the connector's order shape.
func orderRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"order_id":                 item["orderId"],
		"legacy_order_id":          item["legacyOrderId"],
		"creation_date":            item["creationDate"],
		"last_modified_date":       item["lastModifiedDate"],
		"order_fulfillment_status": item["orderFulfillmentStatus"],
		"order_payment_status":     item["orderPaymentStatus"],
		"seller_id":                item["sellerId"],
		"sales_record_reference":   item["salesRecordReference"],
	}
	if buyer, ok := item["buyer"].(map[string]any); ok {
		rec["buyer_username"] = buyer["username"]
	}
	if pricing, ok := item["pricingSummary"].(map[string]any); ok {
		if total, ok := pricing["total"].(map[string]any); ok {
			rec["total_value"] = total["value"]
			rec["total_currency"] = total["currency"]
		}
	}
	if lineItems, ok := item["lineItems"].([]any); ok {
		rec["line_item_count"] = len(lineItems)
	}
	return rec
}

// lineItemRecord flattens a single lineItem (nested in an order) into the
// order_line_items shape, carrying the parent order's identity and date.
func lineItemRecord(order, lineItem map[string]any) connectors.Record {
	rec := connectors.Record{
		"line_item_id":                 lineItem["lineItemId"],
		"order_id":                     order["orderId"],
		"legacy_item_id":               lineItem["legacyItemId"],
		"sku":                          lineItem["sku"],
		"title":                        lineItem["title"],
		"quantity":                     lineItem["quantity"],
		"line_item_fulfillment_status": lineItem["lineItemFulfillmentStatus"],
		"creation_date":                order["creationDate"],
	}
	if total, ok := lineItem["total"].(map[string]any); ok {
		rec["total_value"] = total["value"]
		rec["total_currency"] = total["currency"]
	}
	return rec
}

// shippingFulfillmentRecord projects an order into a shipment-oriented row using
// the order's fulfillmentStartInstructions / shippingStep.
func shippingFulfillmentRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"order_id":                 item["orderId"],
		"legacy_order_id":          item["legacyOrderId"],
		"order_fulfillment_status": item["orderFulfillmentStatus"],
		"creation_date":            item["creationDate"],
	}
	if fsi, ok := item["fulfillmentStartInstructions"].([]any); ok && len(fsi) > 0 {
		if first, ok := fsi[0].(map[string]any); ok {
			rec["shipping_step"] = first["fulfillmentInstructionsType"]
			if ship, ok := first["shippingStep"].(map[string]any); ok {
				if to, ok := ship["shipTo"].(map[string]any); ok {
					if name, ok := to["fullName"]; ok {
						rec["ship_to_name"] = name
					}
					if addr, ok := to["contactAddress"].(map[string]any); ok {
						rec["ship_to_city"] = addr["city"]
						rec["ship_to_state_or_province"] = addr["stateOrProvince"]
						rec["ship_to_postal_code"] = addr["postalCode"]
						rec["ship_to_country_code"] = addr["countryCode"]
					}
				}
			}
		}
	}
	return rec
}

// paymentDisputeRecord flattens a raw eBay payment dispute summary.
func paymentDisputeRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"payment_dispute_id": item["paymentDisputeId"],
		"order_id":           item["orderId"],
		"dispute_state":      item["paymentDisputeStatus"],
		"dispute_status":     item["paymentDisputeStatus"],
		"reason":             item["reason"],
		"open_date":          item["openDate"],
		"buyer_username":     item["buyerUsername"],
	}
	if amount, ok := item["amount"].(map[string]any); ok {
		rec["amount_value"] = amount["value"]
		rec["amount_currency"] = amount["currency"]
	}
	return rec
}
