package square

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Square API resource path (relative to
// base_url) it reads from, the JSON key under which the records array lives, the
// query parameter used for the incremental lower bound (empty if unsupported),
// and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Square list endpoint path segment (e.g. "payments").
	resource string
	// arrayKey is the top-level JSON field that holds the records array.
	arrayKey string
	// timeParam is the query param for the start of the time range; empty when
	// the endpoint does not support time filtering (e.g. customers, locations).
	timeParam string
	// mapRecord flattens a raw Square object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// squareStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in squareStreams; the read path
// is fully data-driven from this table.
var squareStreamEndpoints = map[string]streamEndpoint{
	"payments":  {resource: "payments", arrayKey: "payments", timeParam: "begin_time", mapRecord: squarePaymentRecord},
	"refunds":   {resource: "refunds", arrayKey: "refunds", timeParam: "begin_time", mapRecord: squareRefundRecord},
	"customers": {resource: "customers", arrayKey: "customers", timeParam: "", mapRecord: squareCustomerRecord},
	"locations": {resource: "locations", arrayKey: "locations", timeParam: "", mapRecord: squareLocationRecord},
}

// squareStreams returns the connector's published stream catalog. Square objects
// expose a string id; payments/refunds/customers carry created_at/updated_at
// timestamps, so the incremental cursor field is ["updated_at"] for those.
// locations is a small reference list with no incremental cursor.
func squareStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "payments",
			Description:  "Square payments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       squarePaymentFields(),
		},
		{
			Name:         "refunds",
			Description:  "Square payment refunds.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       squareRefundFields(),
		},
		{
			Name:         "customers",
			Description:  "Square customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       squareCustomerFields(),
		},
		{
			Name:        "locations",
			Description: "Square business locations.",
			PrimaryKey:  []string{"id"},
			Fields:      squareLocationFields(),
		},
	}
}

func squarePaymentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "source_type", Type: "string"},
		{Name: "location_id", Type: "string"},
		{Name: "order_id", Type: "string"},
		{Name: "receipt_number", Type: "string"},
		{Name: "amount_money", Type: "object"},
		{Name: "total_money", Type: "object"},
		{Name: "processing_fee", Type: "object"},
	}
}

func squareRefundFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "payment_id", Type: "string"},
		{Name: "order_id", Type: "string"},
		{Name: "location_id", Type: "string"},
		{Name: "reason", Type: "string"},
		{Name: "amount_money", Type: "object"},
		{Name: "processing_fee", Type: "object"},
	}
}

func squareCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "given_name", Type: "string"},
		{Name: "family_name", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "phone_number", Type: "string"},
		{Name: "company_name", Type: "string"},
		{Name: "reference_id", Type: "string"},
		{Name: "creation_source", Type: "string"},
	}
}

func squareLocationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "merchant_id", Type: "string"},
		{Name: "timezone", Type: "string"},
	}
}

func squarePaymentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
		"status":         item["status"],
		"source_type":    item["source_type"],
		"location_id":    item["location_id"],
		"order_id":       item["order_id"],
		"receipt_number": item["receipt_number"],
		"amount_money":   item["amount_money"],
		"total_money":    item["total_money"],
		"processing_fee": item["processing_fee"],
	}
}

func squareRefundRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"created_at":     item["created_at"],
		"updated_at":     item["updated_at"],
		"status":         item["status"],
		"payment_id":     item["payment_id"],
		"order_id":       item["order_id"],
		"location_id":    item["location_id"],
		"reason":         item["reason"],
		"amount_money":   item["amount_money"],
		"processing_fee": item["processing_fee"],
	}
}

func squareCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"created_at":      item["created_at"],
		"updated_at":      item["updated_at"],
		"given_name":      item["given_name"],
		"family_name":     item["family_name"],
		"email_address":   item["email_address"],
		"phone_number":    item["phone_number"],
		"company_name":    item["company_name"],
		"reference_id":    item["reference_id"],
		"creation_source": item["creation_source"],
	}
}

func squareLocationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"status":      item["status"],
		"created_at":  item["created_at"],
		"country":     item["country"],
		"currency":    item["currency"],
		"type":        item["type"],
		"merchant_id": item["merchant_id"],
		"timezone":    item["timezone"],
	}
}
