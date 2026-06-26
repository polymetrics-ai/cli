package humanitix

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Humanitix API resource path and the
// DpathExtractor field that holds the records array, plus the record mapper that
// flattens its objects. substream marks event-scoped resources
// (events/{eventid}/orders, events/{eventid}/tickets) that require an event_id.
type streamEndpoint struct {
	// resource is the path template. For substreams "{eventid}" is replaced with
	// the configured event id at read time.
	resource string
	// recordsField is the JSON field holding the array (events, orders, ...).
	recordsField string
	// substream is true for event-scoped resources needing an event_id config.
	substream bool
	mapRecord func(map[string]any) connectors.Record
}

// humanitixStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in humanitixStreams; the read
// path is fully data-driven from this table.
var humanitixStreamEndpoints = map[string]streamEndpoint{
	"events":  {resource: "events", recordsField: "events", mapRecord: humanitixEventRecord},
	"tags":    {resource: "tags", recordsField: "tags", mapRecord: humanitixTagRecord},
	"orders":  {resource: "events/{eventid}/orders", recordsField: "orders", substream: true, mapRecord: humanitixOrderRecord},
	"tickets": {resource: "events/{eventid}/tickets", recordsField: "tickets", substream: true, mapRecord: humanitixTicketRecord},
}

// humanitixStreams returns the connector's published stream catalog. Every
// Humanitix object exposes a string _id and an updatedAt timestamp, so the
// primary key is ["_id"] and the incremental cursor field is ["updatedAt"].
func humanitixStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "events",
			Description:  "Humanitix events for the account.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"updatedAt"},
			Fields:       humanitixEventFields(),
		},
		{
			Name:         "orders",
			Description:  "Orders for a Humanitix event (requires event_id config).",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"updatedAt"},
			Fields:       humanitixOrderFields(),
		},
		{
			Name:         "tickets",
			Description:  "Tickets for a Humanitix event (requires event_id config).",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"updatedAt"},
			Fields:       humanitixTicketFields(),
		},
		{
			Name:         "tags",
			Description:  "Humanitix tags for the account.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"updatedAt"},
			Fields:       humanitixTagFields(),
		},
	}
}

func humanitixEventFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "slug", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "startDate", Type: "string"},
		{Name: "endDate", Type: "string"},
		{Name: "organiserId", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "public", Type: "boolean"},
		{Name: "published", Type: "boolean"},
		{Name: "markedAsSoldOut", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func humanitixOrderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "eventId", Type: "string"},
		{Name: "eventDateId", Type: "string"},
		{Name: "orderName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "mobile", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "financialStatus", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "total", Type: "number"},
		{Name: "manualOrder", Type: "boolean"},
		{Name: "completedAt", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func humanitixTicketFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "eventId", Type: "string"},
		{Name: "eventDateId", Type: "string"},
		{Name: "orderId", Type: "string"},
		{Name: "orderName", Type: "string"},
		{Name: "ticketTypeId", Type: "string"},
		{Name: "ticketTypeName", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "price", Type: "number"},
		{Name: "total", Type: "number"},
		{Name: "number", Type: "number"},
		{Name: "isDonation", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func humanitixTagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func humanitixEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":             item["_id"],
		"name":            item["name"],
		"slug":            item["slug"],
		"currency":        item["currency"],
		"location":        item["location"],
		"startDate":       item["startDate"],
		"endDate":         item["endDate"],
		"organiserId":     item["organiserId"],
		"userId":          item["userId"],
		"public":          item["public"],
		"published":       item["published"],
		"markedAsSoldOut": item["markedAsSoldOut"],
		"createdAt":       item["createdAt"],
		"updatedAt":       item["updatedAt"],
	}
}

func humanitixOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":             item["_id"],
		"eventId":         item["eventId"],
		"eventDateId":     item["eventDateId"],
		"orderName":       item["orderName"],
		"email":           item["email"],
		"firstName":       item["firstName"],
		"lastName":        item["lastName"],
		"mobile":          item["mobile"],
		"status":          item["status"],
		"financialStatus": item["financialStatus"],
		"currency":        item["currency"],
		"total":           item["total"],
		"manualOrder":     item["manualOrder"],
		"completedAt":     item["completedAt"],
		"createdAt":       item["createdAt"],
		"updatedAt":       item["updatedAt"],
	}
}

func humanitixTicketRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":            item["_id"],
		"eventId":        item["eventId"],
		"eventDateId":    item["eventDateId"],
		"orderId":        item["orderId"],
		"orderName":      item["orderName"],
		"ticketTypeId":   item["ticketTypeId"],
		"ticketTypeName": item["ticketTypeName"],
		"firstName":      item["firstName"],
		"lastName":       item["lastName"],
		"status":         item["status"],
		"currency":       item["currency"],
		"price":          item["price"],
		"total":          item["total"],
		"number":         item["number"],
		"isDonation":     item["isDonation"],
		"createdAt":      item["createdAt"],
		"updatedAt":      item["updatedAt"],
	}
}

func humanitixTagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":       item["_id"],
		"name":      item["name"],
		"location":  item["location"],
		"userId":    item["userId"],
		"createdAt": item["createdAt"],
		"updatedAt": item["updatedAt"],
	}
}
