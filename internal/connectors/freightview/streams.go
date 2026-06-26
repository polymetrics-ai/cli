package freightview

import "polymetrics.ai/internal/connectors"

// streamDef describes how a Freightview stream is read. shipments is the root
// stream paginated by continuationToken; quotes and tracking are substreams that
// fan out over each shipment id.
type streamDef struct {
	// name is the published stream name.
	name string
	// recordsPath is the dotted JSON path to the record array in a list
	// response. An empty string selects the response root (a top-level array).
	recordsPath string
	// substream marks streams that read per-shipment subresources at
	// /shipments/{shipmentId}/<sub>.
	substream bool
	// sub is the subresource path segment for a substream (e.g. "quotes").
	sub string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// freightviewStreamDefs is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in freightviewStreams.
var freightviewStreamDefs = map[string]streamDef{
	"shipments": {
		name:        "shipments",
		recordsPath: "shipments",
		mapRecord:   shipmentRecord,
	},
	"quotes": {
		name:        "quotes",
		recordsPath: "quotes",
		substream:   true,
		sub:         "quotes",
		mapRecord:   quoteRecord,
	},
	"tracking": {
		name:        "tracking",
		recordsPath: "", // tracking events are returned as a top-level array.
		substream:   true,
		sub:         "tracking",
		mapRecord:   trackingRecord,
	},
}

// freightviewStreams returns the connector's published stream catalog.
func freightviewStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "shipments",
			Description: "Freightview shipments (root stream, paginated by continuationToken).",
			PrimaryKey:  []string{"shipmentId"},
			Fields:      shipmentFields(),
		},
		{
			Name:        "quotes",
			Description: "Freightview quotes per shipment.",
			PrimaryKey:  []string{"quoteId"},
			Fields:      quoteFields(),
		},
		{
			Name:        "tracking",
			Description: "Freightview tracking events per shipment.",
			PrimaryKey:  []string{"createdDate"},
			Fields:      trackingFields(),
		},
	}
}

func shipmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "shipmentId", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "direction", Type: "string"},
		{Name: "bookedBy", Type: "string"},
		{Name: "quotedBy", Type: "string"},
		{Name: "bookedDate", Type: "string"},
		{Name: "pickupDate", Type: "string"},
		{Name: "createdDate", Type: "string"},
		{Name: "isArchived", Type: "boolean"},
		{Name: "isLiveLoad", Type: "boolean"},
		{Name: "items", Type: "array"},
		{Name: "refNums", Type: "array"},
		{Name: "locations", Type: "array"},
		{Name: "documents", Type: "array"},
		{Name: "equipment", Type: "object"},
		{Name: "pickup", Type: "object"},
		{Name: "billTo", Type: "object"},
		{Name: "bol", Type: "object"},
		{Name: "tracking", Type: "object"},
		{Name: "selectedQuote", Type: "object"},
	}
}

func quoteFields() []connectors.Field {
	return []connectors.Field{
		{Name: "quoteId", Type: "string"},
		{Name: "mode", Type: "string"},
		{Name: "method", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "currency", Type: "string"},
		{Name: "quoteNum", Type: "string"},
		{Name: "carrierId", Type: "string"},
		{Name: "serviceId", Type: "string"},
		{Name: "createdDate", Type: "string"},
		{Name: "pricingType", Type: "string"},
		{Name: "paymentTerms", Type: "string"},
		{Name: "providerCode", Type: "string"},
		{Name: "providerName", Type: "string"},
		{Name: "equipmentType", Type: "string"},
		{Name: "pricingMethod", Type: "string"},
	}
}

func trackingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "createdDate", Type: "string"},
		{Name: "summary", Type: "string"},
		{Name: "eventDate", Type: "string"},
		{Name: "eventTime", Type: "string"},
		{Name: "eventType", Type: "string"},
	}
}

func shipmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"shipmentId":    item["shipmentId"],
		"status":        item["status"],
		"direction":     item["direction"],
		"bookedBy":      item["bookedBy"],
		"quotedBy":      item["quotedBy"],
		"bookedDate":    item["bookedDate"],
		"pickupDate":    item["pickupDate"],
		"createdDate":   item["createdDate"],
		"isArchived":    item["isArchived"],
		"isLiveLoad":    item["isLiveLoad"],
		"items":         item["items"],
		"refNums":       item["refNums"],
		"locations":     item["locations"],
		"documents":     item["documents"],
		"equipment":     item["equipment"],
		"pickup":        item["pickup"],
		"billTo":        item["billTo"],
		"bol":           item["bol"],
		"tracking":      item["tracking"],
		"selectedQuote": item["selectedQuote"],
	}
}

func quoteRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"quoteId":       item["quoteId"],
		"mode":          item["mode"],
		"method":        item["method"],
		"source":        item["source"],
		"status":        item["status"],
		"amount":        item["amount"],
		"currency":      item["currency"],
		"quoteNum":      item["quoteNum"],
		"carrierId":     item["carrierId"],
		"serviceId":     item["serviceId"],
		"createdDate":   item["createdDate"],
		"pricingType":   item["pricingType"],
		"paymentTerms":  item["paymentTerms"],
		"providerCode":  item["providerCode"],
		"providerName":  item["providerName"],
		"equipmentType": item["equipmentType"],
		"pricingMethod": item["pricingMethod"],
	}
}

func trackingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"createdDate": item["createdDate"],
		"summary":     item["summary"],
		"eventDate":   item["eventDate"],
		"eventTime":   item["eventTime"],
		"eventType":   item["eventType"],
	}
}
