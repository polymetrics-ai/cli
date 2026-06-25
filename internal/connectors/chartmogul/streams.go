package chartmogul

import "polymetrics/internal/connectors"

// paginationKind describes how a ChartMogul endpoint paginates.
type paginationKind int

const (
	// pageCursor is ChartMogul's cursor/has_more pagination over entries[]
	// (customers, activities).
	pageCursor paginationKind = iota
	// pageSingle is a single response whose entries[] holds all rows
	// (metrics endpoints).
	pageSingle
	// pageObject is a single JSON object treated as one record (account).
	pageObject
)

// streamEndpoint maps a stream name to the ChartMogul API resource path
// (relative to base_url), its pagination style, and the record mapper that
// flattens its objects. Adding a stream means adding one entry here plus a
// Stream definition in chartmogulStreams; the read path is data-driven from
// this table.
type streamEndpoint struct {
	resource   string
	pagination paginationKind
	// metricsInterval, when set, is sent as the metrics `interval` query param.
	metricsInterval string
	mapRecord       func(map[string]any) connectors.Record
}

// chartmogulStreamEndpoints is the per-stream routing table.
var chartmogulStreamEndpoints = map[string]streamEndpoint{
	"customers":  {resource: "customers", pagination: pageCursor, mapRecord: chartmogulCustomerRecord},
	"activities": {resource: "activities", pagination: pageCursor, mapRecord: chartmogulActivityRecord},
	"customer_count": {
		resource:        "metrics/customer-count",
		pagination:      pageSingle,
		metricsInterval: "month",
		mapRecord:       chartmogulCustomerCountRecord,
	},
	"account": {resource: "account", pagination: pageObject, mapRecord: chartmogulAccountRecord},
}

// chartmogulStreams returns the connector's published stream catalog.
func chartmogulStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "ChartMogul customers and their billing/subscription metrics.",
			PrimaryKey:   []string{"uuid"},
			CursorFields: []string{"customer-since"},
			Fields:       chartmogulCustomerFields(),
		},
		{
			Name:         "activities",
			Description:  "ChartMogul subscription activity log (new_biz, churn, expansion, contraction, reactivation).",
			PrimaryKey:   []string{"uuid"},
			CursorFields: []string{"date"},
			Fields:       chartmogulActivityFields(),
		},
		{
			Name:         "customer_count",
			Description:  "ChartMogul monthly customer-count metric series.",
			PrimaryKey:   []string{"date"},
			CursorFields: []string{"date"},
			Fields:       chartmogulCustomerCountFields(),
		},
		{
			Name:        "account",
			Description: "ChartMogul account details.",
			PrimaryKey:  []string{"uuid"},
			Fields:      chartmogulAccountFields(),
		},
	}
}

func chartmogulCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "external_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "customer-since", Type: "timestamp"},
		{Name: "company", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "mrr", Type: "number"},
		{Name: "arr", Type: "number"},
		{Name: "billing-system-type", Type: "string"},
	}
}

func chartmogulActivityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "date", Type: "timestamp"},
		{Name: "currency", Type: "string"},
		{Name: "activity-mrr", Type: "number"},
		{Name: "activity-mrr-movement", Type: "number"},
		{Name: "activity-arr", Type: "number"},
		{Name: "description", Type: "string"},
		{Name: "customer-uuid", Type: "string"},
		{Name: "customer-name", Type: "string"},
		{Name: "customer-external-id", Type: "string"},
		{Name: "subscription-external-id", Type: "string"},
		{Name: "plan-external-id", Type: "string"},
	}
}

func chartmogulCustomerCountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "customers", Type: "integer"},
		{Name: "percentage-change", Type: "number"},
	}
}

func chartmogulAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "time_zone", Type: "string"},
		{Name: "week_start_on", Type: "string"},
	}
}

func chartmogulCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":                item["uuid"],
		"external_id":         item["external_id"],
		"name":                item["name"],
		"email":               item["email"],
		"status":              item["status"],
		"customer-since":      item["customer-since"],
		"company":             item["company"],
		"country":             item["country"],
		"city":                item["city"],
		"currency":            item["currency"],
		"mrr":                 item["mrr"],
		"arr":                 item["arr"],
		"billing-system-type": item["billing-system-type"],
	}
}

func chartmogulActivityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":                     item["uuid"],
		"type":                     item["type"],
		"date":                     item["date"],
		"currency":                 item["currency"],
		"activity-mrr":             item["activity-mrr"],
		"activity-mrr-movement":    item["activity-mrr-movement"],
		"activity-arr":             item["activity-arr"],
		"description":              item["description"],
		"customer-uuid":            item["customer-uuid"],
		"customer-name":            item["customer-name"],
		"customer-external-id":     item["customer-external-id"],
		"subscription-external-id": item["subscription-external-id"],
		"plan-external-id":         item["plan-external-id"],
	}
}

func chartmogulCustomerCountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"date":              item["date"],
		"customers":         item["customers"],
		"percentage-change": item["percentage-change"],
	}
}

func chartmogulAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":          item["uuid"],
		"name":          item["name"],
		"currency":      item["currency"],
		"time_zone":     item["time_zone"],
		"week_start_on": item["week_start_on"],
	}
}
