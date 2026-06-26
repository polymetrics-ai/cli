package omnisend

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Omnisend API resource path (relative
// to base_url), the JSON field path the record array lives under (Omnisend uses
// a per-resource key, e.g. "contacts", but "campaign" singular for campaigns),
// and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Omnisend list endpoint path segment (e.g. "contacts").
	resource string
	// recordsPath is the JSON key holding the array of records in the response.
	recordsPath string
	// primaryKey is the per-stream identifier field.
	primaryKey string
	// mapRecord flattens a raw Omnisend object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// omnisendStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in omnisendStreams; the read
// path is fully data-driven from this table.
var omnisendStreamEndpoints = map[string]streamEndpoint{
	"contacts":  {resource: "contacts", recordsPath: "contacts", primaryKey: "contactID", mapRecord: omnisendContactRecord},
	"campaigns": {resource: "campaigns", recordsPath: "campaign", primaryKey: "campaignID", mapRecord: omnisendCampaignRecord},
	"carts":     {resource: "carts", recordsPath: "carts", primaryKey: "cartID", mapRecord: omnisendCartRecord},
	"orders":    {resource: "orders", recordsPath: "orders", primaryKey: "orderID", mapRecord: omnisendOrderRecord},
	"products":  {resource: "products", recordsPath: "products", primaryKey: "productID", mapRecord: omnisendProductRecord},
}

// omnisendStreams returns the connector's published stream catalog. Omnisend
// only supports full-refresh syncs upstream, but each resource carries a
// createdAt/updatedAt timestamp that we surface as a cursor field for callers
// that want to track high-water marks.
func omnisendStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "Omnisend contacts (subscribers) with channel statuses and consent metadata.",
			PrimaryKey:   []string{"contactID"},
			CursorFields: []string{"createdAt"},
			Fields:       omnisendContactFields(),
		},
		{
			Name:         "campaigns",
			Description:  "Omnisend email/SMS campaigns with delivery and engagement metrics.",
			PrimaryKey:   []string{"campaignID"},
			CursorFields: []string{"createdAt"},
			Fields:       omnisendCampaignFields(),
		},
		{
			Name:         "carts",
			Description:  "Omnisend shopping carts captured for abandonment automation.",
			PrimaryKey:   []string{"cartID"},
			CursorFields: []string{"createdAt"},
			Fields:       omnisendCartFields(),
		},
		{
			Name:         "orders",
			Description:  "Omnisend ecommerce orders with line items and totals.",
			PrimaryKey:   []string{"orderID"},
			CursorFields: []string{"createdAt"},
			Fields:       omnisendOrderFields(),
		},
		{
			Name:         "products",
			Description:  "Omnisend product catalog entries with variants and images.",
			PrimaryKey:   []string{"productID"},
			CursorFields: []string{"createdAt"},
			Fields:       omnisendProductFields(),
		},
	}
}

func omnisendContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "contactID", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "countryCode", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "segments", Type: "array"},
		{Name: "tags", Type: "array"},
	}
}

func omnisendCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "campaignID", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "fromName", Type: "string"},
		{Name: "sent", Type: "number"},
		{Name: "opened", Type: "number"},
		{Name: "clicked", Type: "number"},
		{Name: "bounced", Type: "number"},
		{Name: "unsubscribed", Type: "number"},
		{Name: "startDate", Type: "string"},
		{Name: "endDate", Type: "string"},
		{Name: "createdAt", Type: "string"},
	}
}

func omnisendCartFields() []connectors.Field {
	return []connectors.Field{
		{Name: "cartID", Type: "string"},
		{Name: "contactID", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "cartSum", Type: "number"},
		{Name: "cartRecoveryUrl", Type: "string"},
		{Name: "products", Type: "array"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func omnisendOrderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "orderID", Type: "string"},
		{Name: "orderNumber", Type: "number"},
		{Name: "contactID", Type: "string"},
		{Name: "cartID", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "orderSum", Type: "number"},
		{Name: "subTotalSum", Type: "number"},
		{Name: "taxSum", Type: "number"},
		{Name: "shippingSum", Type: "number"},
		{Name: "discountSum", Type: "number"},
		{Name: "paymentStatus", Type: "string"},
		{Name: "fulfillmentStatus", Type: "string"},
		{Name: "products", Type: "array"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func omnisendProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "productID", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "vendor", Type: "string"},
		{Name: "productUrl", Type: "string"},
		{Name: "categoryIDs", Type: "array"},
		{Name: "variants", Type: "array"},
		{Name: "images", Type: "array"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func omnisendContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"contactID":   item["contactID"],
		"email":       item["email"],
		"firstName":   item["firstName"],
		"lastName":    item["lastName"],
		"status":      item["status"],
		"country":     item["country"],
		"countryCode": item["countryCode"],
		"city":        item["city"],
		"state":       item["state"],
		"createdAt":   item["createdAt"],
		"segments":    item["segments"],
		"tags":        item["tags"],
	}
}

func omnisendCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"campaignID":   item["campaignID"],
		"name":         item["name"],
		"subject":      item["subject"],
		"type":         item["type"],
		"status":       item["status"],
		"fromName":     item["fromName"],
		"sent":         item["sent"],
		"opened":       item["opened"],
		"clicked":      item["clicked"],
		"bounced":      item["bounced"],
		"unsubscribed": item["unsubscribed"],
		"startDate":    item["startDate"],
		"endDate":      item["endDate"],
		"createdAt":    item["createdAt"],
	}
}

func omnisendCartRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"cartID":          item["cartID"],
		"contactID":       item["contactID"],
		"email":           item["email"],
		"phone":           item["phone"],
		"currency":        item["currency"],
		"cartSum":         item["cartSum"],
		"cartRecoveryUrl": item["cartRecoveryUrl"],
		"products":        item["products"],
		"createdAt":       item["createdAt"],
		"updatedAt":       item["updatedAt"],
	}
}

func omnisendOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"orderID":           item["orderID"],
		"orderNumber":       item["orderNumber"],
		"contactID":         item["contactID"],
		"cartID":            item["cartID"],
		"email":             item["email"],
		"currency":          item["currency"],
		"orderSum":          item["orderSum"],
		"subTotalSum":       item["subTotalSum"],
		"taxSum":            item["taxSum"],
		"shippingSum":       item["shippingSum"],
		"discountSum":       item["discountSum"],
		"paymentStatus":     item["paymentStatus"],
		"fulfillmentStatus": item["fulfillmentStatus"],
		"products":          item["products"],
		"createdAt":         item["createdAt"],
		"updatedAt":         item["updatedAt"],
	}
}

func omnisendProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"productID":   item["productID"],
		"title":       item["title"],
		"description": item["description"],
		"type":        item["type"],
		"status":      item["status"],
		"currency":    item["currency"],
		"vendor":      item["vendor"],
		"productUrl":  item["productUrl"],
		"categoryIDs": item["categoryIDs"],
		"variants":    item["variants"],
		"images":      item["images"],
		"createdAt":   item["createdAt"],
		"updatedAt":   item["updatedAt"],
	}
}
