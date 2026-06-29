package mantle

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Mantle API resource path (relative to
// base_url), the JSON selector under which its records array lives, and the
// record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Mantle list endpoint path segment (e.g. "v1/customers").
	resource string
	// selector is the top-level JSON field holding the records array
	// (e.g. "customers"), per the Mantle DpathExtractor field_path.
	selector string
	// mapRecord flattens a raw Mantle object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mantleStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in mantleStreams; the read path
// is fully data-driven from this table.
var mantleStreamEndpoints = map[string]streamEndpoint{
	"customers":     {resource: "v1/customers", selector: "customers", mapRecord: mantleCustomerRecord},
	"subscriptions": {resource: "v1/subscriptions", selector: "subscriptions", mapRecord: mantleSubscriptionRecord},
}

// mantleStreams returns the connector's published stream catalog. Mantle objects
// carry a string id and ISO-8601 timestamps; customers are incremental on
// updatedAt and subscriptions on createdAt, matching the upstream upstream
// manifest.
func mantleStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "Mantle customers (apps/merchants), including revenue and lifetime-value metrics.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       mantleCustomerFields(),
		},
		{
			Name:         "subscriptions",
			Description:  "Mantle subscriptions with plan, totals, trial, and billing period details.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       mantleSubscriptionFields(),
		},
	}
}

func mantleCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "industry", Type: "string"},
		{Name: "countryCode", Type: "string"},
		{Name: "shopifyDomain", Type: "string"},
		{Name: "shopifyShopId", Type: "string"},
		{Name: "test", Type: "boolean"},
		{Name: "tags", Type: "array"},
		{Name: "last30Revenue", Type: "number"},
		{Name: "lifetimeValue", Type: "number"},
		{Name: "averageMonthlyRevenue", Type: "number"},
		{Name: "firstInteractionAt", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func mantleSubscriptionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "total", Type: "number"},
		{Name: "subtotal", Type: "number"},
		{Name: "presentmentTotal", Type: "number"},
		{Name: "presentmentSubtotal", Type: "number"},
		{Name: "plan", Type: "object"},
		{Name: "lineItems", Type: "array"},
		{Name: "frozenAt", Type: "string"},
		{Name: "canceledAt", Type: "string"},
		{Name: "activatedAt", Type: "string"},
		{Name: "trialStartsAt", Type: "string"},
		{Name: "trialExpiresAt", Type: "string"},
		{Name: "currentPeriodStart", Type: "string"},
		{Name: "currentPeriodEnd", Type: "string"},
		{Name: "billingCycleAnchor", Type: "string"},
		{Name: "createdAt", Type: "string"},
	}
}

func mantleCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"name":                  item["name"],
		"email":                 item["email"],
		"domain":                item["domain"],
		"industry":              item["industry"],
		"countryCode":           item["countryCode"],
		"shopifyDomain":         item["shopifyDomain"],
		"shopifyShopId":         item["shopifyShopId"],
		"test":                  item["test"],
		"tags":                  item["tags"],
		"last30Revenue":         item["last30Revenue"],
		"lifetimeValue":         item["lifetimeValue"],
		"averageMonthlyRevenue": item["averageMonthlyRevenue"],
		"firstInteractionAt":    item["firstInteractionAt"],
		"createdAt":             item["createdAt"],
		"updatedAt":             item["updatedAt"],
	}
}

func mantleSubscriptionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"active":              item["active"],
		"total":               item["total"],
		"subtotal":            item["subtotal"],
		"presentmentTotal":    item["presentmentTotal"],
		"presentmentSubtotal": item["presentmentSubtotal"],
		"plan":                item["plan"],
		"lineItems":           item["lineItems"],
		"frozenAt":            item["frozenAt"],
		"canceledAt":          item["canceledAt"],
		"activatedAt":         item["activatedAt"],
		"trialStartsAt":       item["trialStartsAt"],
		"trialExpiresAt":      item["trialExpiresAt"],
		"currentPeriodStart":  item["currentPeriodStart"],
		"currentPeriodEnd":    item["currentPeriodEnd"],
		"billingCycleAnchor":  item["billingCycleAnchor"],
		"createdAt":           item["createdAt"],
	}
}
