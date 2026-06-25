package cloudbeds

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Cloudbeds API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Cloudbeds endpoint path segment (e.g. "getReservations").
	resource string
	// mapRecord flattens a raw Cloudbeds object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// cloudbedsStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in cloudbedsStreams; the read
// path is fully data-driven from this table. Paths and record paths come from the
// Cloudbeds API v1.2 (Bearer auth, pageNumber/pageSize pagination, records at
// "data").
var cloudbedsStreamEndpoints = map[string]streamEndpoint{
	"guests":       {resource: "getGuestList", mapRecord: cloudbedsGuestRecord},
	"hotels":       {resource: "getHotels", mapRecord: cloudbedsHotelRecord},
	"rooms":        {resource: "getRoomBlocks", mapRecord: cloudbedsRoomRecord},
	"reservations": {resource: "getReservations", mapRecord: cloudbedsReservationRecord},
	"transactions": {resource: "getTransactions", mapRecord: cloudbedsTransactionRecord},
}

// cloudbedsStreams returns the connector's published stream catalog. Cloudbeds is
// full-refresh only upstream, so no cursor fields are published; primary keys
// follow the Cloudbeds entity ids.
func cloudbedsStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "guests",
			Description: "Cloudbeds guests (getGuestList).",
			PrimaryKey:  []string{"guestID"},
			Fields:      cloudbedsGuestFields(),
		},
		{
			Name:        "hotels",
			Description: "Cloudbeds properties/hotels (getHotels).",
			PrimaryKey:  []string{"propertyID"},
			Fields:      cloudbedsHotelFields(),
		},
		{
			Name:        "rooms",
			Description: "Cloudbeds room blocks by property (getRoomBlocks).",
			PrimaryKey:  []string{"propertyID"},
			Fields:      cloudbedsRoomFields(),
		},
		{
			Name:        "reservations",
			Description: "Cloudbeds reservations (getReservations).",
			PrimaryKey:  []string{"reservationID"},
			Fields:      cloudbedsReservationFields(),
		},
		{
			Name:        "transactions",
			Description: "Cloudbeds folio transactions (getTransactions).",
			PrimaryKey:  []string{"transactionID"},
			Fields:      cloudbedsTransactionFields(),
		},
	}
}

func cloudbedsGuestFields() []connectors.Field {
	return []connectors.Field{
		{Name: "guestID", Type: "string"},
		{Name: "guestName", Type: "string"},
		{Name: "guestEmail", Type: "string"},
		{Name: "propertyID", Type: "string"},
		{Name: "reservationID", Type: "string"},
		{Name: "isMainGuest", Type: "boolean"},
		{Name: "isAnonymized", Type: "boolean"},
		{Name: "dateCreated", Type: "string"},
		{Name: "dateModified", Type: "string"},
	}
}

func cloudbedsHotelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "propertyID", Type: "string"},
		{Name: "organizationID", Type: "string"},
		{Name: "propertyName", Type: "string"},
		{Name: "propertyDescription", Type: "string"},
		{Name: "propertyCurrency", Type: "string"},
		{Name: "propertyTimezone", Type: "string"},
		{Name: "propertyImage", Type: "string"},
	}
}

func cloudbedsRoomFields() []connectors.Field {
	return []connectors.Field{
		{Name: "propertyID", Type: "string"},
		{Name: "roomBlocks", Type: "array"},
	}
}

func cloudbedsReservationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "reservationID", Type: "string"},
		{Name: "propertyID", Type: "string"},
		{Name: "guestID", Type: "string"},
		{Name: "guestName", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "startDate", Type: "string"},
		{Name: "endDate", Type: "string"},
		{Name: "adults", Type: "integer"},
		{Name: "children", Type: "integer"},
		{Name: "balance", Type: "number"},
		{Name: "sourceName", Type: "string"},
		{Name: "origin", Type: "string"},
		{Name: "dateCreated", Type: "string"},
		{Name: "dateModified", Type: "string"},
	}
}

func cloudbedsTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "transactionID", Type: "string"},
		{Name: "propertyID", Type: "string"},
		{Name: "reservationID", Type: "string"},
		{Name: "guestID", Type: "string"},
		{Name: "guestName", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "category", Type: "string"},
		{Name: "transactionCategory", Type: "string"},
		{Name: "transactionType", Type: "string"},
		{Name: "transactionCode", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "transactionDateTime", Type: "string"},
		{Name: "transactionDateTimeUTC", Type: "string"},
	}
}

func cloudbedsGuestRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"guestID":       item["guestID"],
		"guestName":     item["guestName"],
		"guestEmail":    item["guestEmail"],
		"propertyID":    item["propertyID"],
		"reservationID": item["reservationID"],
		"isMainGuest":   item["isMainGuest"],
		"isAnonymized":  item["isAnonymized"],
		"dateCreated":   item["dateCreated"],
		"dateModified":  item["dateModified"],
	}
}

func cloudbedsHotelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"propertyID":          item["propertyID"],
		"organizationID":      item["organizationID"],
		"propertyName":        item["propertyName"],
		"propertyDescription": item["propertyDescription"],
		"propertyCurrency":    item["propertyCurrency"],
		"propertyTimezone":    item["propertyTimezone"],
		"propertyImage":       item["propertyImage"],
	}
}

func cloudbedsRoomRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"propertyID": item["propertyID"],
		"roomBlocks": item["roomBlocks"],
	}
}

func cloudbedsReservationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"reservationID": item["reservationID"],
		"propertyID":    item["propertyID"],
		"guestID":       item["guestID"],
		"guestName":     item["guestName"],
		"status":        item["status"],
		"startDate":     item["startDate"],
		"endDate":       item["endDate"],
		"adults":        item["adults"],
		"children":      item["children"],
		"balance":       item["balance"],
		"sourceName":    item["sourceName"],
		"origin":        item["origin"],
		"dateCreated":   item["dateCreated"],
		"dateModified":  item["dateModified"],
	}
}

func cloudbedsTransactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"transactionID":          item["transactionID"],
		"propertyID":             item["propertyID"],
		"reservationID":          item["reservationID"],
		"guestID":                item["guestID"],
		"guestName":              item["guestName"],
		"amount":                 item["amount"],
		"currency":               item["currency"],
		"category":               item["category"],
		"transactionCategory":    item["transactionCategory"],
		"transactionType":        item["transactionType"],
		"transactionCode":        item["transactionCode"],
		"description":            item["description"],
		"transactionDateTime":    item["transactionDateTime"],
		"transactionDateTimeUTC": item["transactionDateTimeUTC"],
	}
}
