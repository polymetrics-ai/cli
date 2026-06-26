package navan

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Navan API resource path (relative to
// base_url) it reads from, optional fixed request parameters (e.g. the
// bookingType filter that splits the bookings endpoint), and the record mapper
// that flattens its objects.
type streamEndpoint struct {
	// resource is the Navan list endpoint path (e.g. "v1/bookings").
	resource string
	// params are fixed query parameters applied to every page request.
	params map[string]string
	// mapRecord flattens a raw Navan object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// navanStreamEndpoints is the per-stream routing table. The bookings endpoint is
// filtered by bookingType to expose flight, hotel, and car bookings as distinct
// streams; the read path is fully data-driven from this table.
var navanStreamEndpoints = map[string]streamEndpoint{
	"bookings":       {resource: "v1/bookings", params: map[string]string{"bookingType": "FLIGHT"}, mapRecord: navanBookingRecord},
	"hotel_bookings": {resource: "v1/bookings", params: map[string]string{"bookingType": "HOTEL"}, mapRecord: navanBookingRecord},
	"car_bookings":   {resource: "v1/bookings", params: map[string]string{"bookingType": "CAR"}, mapRecord: navanBookingRecord},
	"rail_bookings":  {resource: "v1/bookings", params: map[string]string{"bookingType": "RAIL"}, mapRecord: navanBookingRecord},
}

// navanStreams returns the connector's published stream catalog. Every Navan
// booking exposes a string uuid and a lastModified timestamp, so the primary key
// is ["uuid"] and the incremental cursor field is ["last_modified"].
func navanStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "bookings",
			Description:  "Navan flight bookings.",
			PrimaryKey:   []string{"uuid"},
			CursorFields: []string{"last_modified"},
			Fields:       navanBookingFields(),
		},
		{
			Name:         "hotel_bookings",
			Description:  "Navan hotel bookings.",
			PrimaryKey:   []string{"uuid"},
			CursorFields: []string{"last_modified"},
			Fields:       navanBookingFields(),
		},
		{
			Name:         "car_bookings",
			Description:  "Navan car rental bookings.",
			PrimaryKey:   []string{"uuid"},
			CursorFields: []string{"last_modified"},
			Fields:       navanBookingFields(),
		},
		{
			Name:         "rail_bookings",
			Description:  "Navan rail bookings.",
			PrimaryKey:   []string{"uuid"},
			CursorFields: []string{"last_modified"},
			Fields:       navanBookingFields(),
		},
	}
}

func navanBookingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "booking_id", Type: "string"},
		{Name: "booking_type", Type: "string"},
		{Name: "booking_status", Type: "string"},
		{Name: "booking_method", Type: "string"},
		{Name: "approval_status", Type: "string"},
		{Name: "confirmation_number", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "grand_total", Type: "number"},
		{Name: "base_price", Type: "number"},
		{Name: "booking_fee", Type: "number"},
		{Name: "destination", Type: "string"},
		{Name: "start_date", Type: "string"},
		{Name: "end_date", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "last_modified", Type: "string"},
		{Name: "cancelled_at", Type: "string"},
		{Name: "domestic", Type: "boolean"},
		{Name: "expensed", Type: "boolean"},
	}
}

// navanBookingRecord flattens a raw Navan booking object into a connectors.Record.
// The camelCase API fields are normalized to snake_case stable column names.
func navanBookingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":                item["uuid"],
		"booking_id":          item["bookingId"],
		"booking_type":        item["bookingType"],
		"booking_status":      item["bookingStatus"],
		"booking_method":      item["bookingMethod"],
		"approval_status":     item["approvalStatus"],
		"confirmation_number": item["confirmationNumber"],
		"currency":            item["currency"],
		"grand_total":         item["grandTotal"],
		"base_price":          item["basePrice"],
		"booking_fee":         item["bookingFee"],
		"destination":         item["destination"],
		"start_date":          item["startDate"],
		"end_date":            item["endDate"],
		"created":             item["created"],
		"last_modified":       item["lastModified"],
		"cancelled_at":        item["cancelledAt"],
		"domestic":            item["domestic"],
		"expensed":            item["expensed"],
	}
}

// MapBookingForTest exposes navanBookingRecord to the external test package so
// the record mapper can be unit-tested without a live server.
func MapBookingForTest(item map[string]any) connectors.Record {
	return navanBookingRecord(item)
}
