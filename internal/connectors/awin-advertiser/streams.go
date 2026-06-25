package awinadvertiser

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Awin Advertiser API resource it reads
// from (relative to base_url, with the {advertiserId} placeholder substituted at
// read time) and the record mapper that flattens its objects. dateType is the
// query param value Awin uses to bound the date window for that resource; an empty
// dateType means the resource takes no date window.
type streamEndpoint struct {
	// resource is the Awin path template under /advertisers/{advertiserId}.
	resource string
	// dateType is Awin's dateType query value (e.g. "transaction") for date-windowed
	// resources. Empty for resources that take no startDate/endDate.
	dateType string
	// mapRecord flattens a raw Awin object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// awinStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in awinStreams; the read path is
// fully data-driven from this table.
var awinStreamEndpoints = map[string]streamEndpoint{
	"transactions":         {resource: "transactions/", dateType: "transaction", mapRecord: awinTransactionRecord},
	"campaign_performance": {resource: "reports/aggregated/publisher", dateType: "", mapRecord: awinReportRecord},
	"publishers":           {resource: "publishers/", dateType: "", mapRecord: awinPublisherRecord},
}

// awinStreams returns the connector's published stream catalog.
func awinStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "transactions",
			Description:  "Awin advertiser commission transactions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"transactionDate"},
			Fields:       awinTransactionFields(),
		},
		{
			Name:         "campaign_performance",
			Description:  "Awin advertiser performance aggregated by publisher.",
			PrimaryKey:   []string{"publisherId"},
			CursorFields: nil,
			Fields:       awinReportFields(),
		},
		{
			Name:         "publishers",
			Description:  "Publishers the advertiser has a relationship with.",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       awinPublisherFields(),
		},
	}
}

func awinTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "url", Type: "string"},
		{Name: "advertiserId", Type: "integer"},
		{Name: "publisherId", Type: "integer"},
		{Name: "commissionSharingPublisherId", Type: "integer"},
		{Name: "siteName", Type: "string"},
		{Name: "transactionDate", Type: "string"},
		{Name: "validationDate", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "transactionStatus", Type: "string"},
		{Name: "saleAmount", Type: "object"},
		{Name: "commissionAmount", Type: "object"},
		{Name: "clickDate", Type: "string"},
		{Name: "clickRefs", Type: "object"},
		{Name: "customParameters", Type: "object"},
	}
}

func awinReportFields() []connectors.Field {
	return []connectors.Field{
		{Name: "advertiserId", Type: "integer"},
		{Name: "publisherId", Type: "integer"},
		{Name: "publisherName", Type: "string"},
		{Name: "region", Type: "string"},
		{Name: "currency", Type: "string"},
		{Name: "impressions", Type: "integer"},
		{Name: "clicks", Type: "integer"},
		{Name: "pendingNo", Type: "integer"},
		{Name: "confirmedNo", Type: "integer"},
		{Name: "declinedNo", Type: "integer"},
		{Name: "totalNo", Type: "integer"},
		{Name: "totalSaleAmount", Type: "object"},
		{Name: "totalComm", Type: "number"},
	}
}

func awinPublisherFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "displayUrl", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func awinTransactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                           item["id"],
		"url":                          item["url"],
		"advertiserId":                 item["advertiserId"],
		"publisherId":                  item["publisherId"],
		"commissionSharingPublisherId": item["commissionSharingPublisherId"],
		"siteName":                     item["siteName"],
		"transactionDate":              item["transactionDate"],
		"validationDate":               item["validationDate"],
		"type":                         item["type"],
		"transactionStatus":            item["transactionStatus"],
		"saleAmount":                   item["saleAmount"],
		"commissionAmount":             item["commissionAmount"],
		"clickDate":                    item["clickDate"],
		"clickRefs":                    item["clickRefs"],
		"customParameters":             item["customParameters"],
	}
}

func awinReportRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"advertiserId":    item["advertiserId"],
		"publisherId":     item["publisherId"],
		"publisherName":   item["publisherName"],
		"region":          item["region"],
		"currency":        item["currency"],
		"impressions":     item["impressions"],
		"clicks":          item["clicks"],
		"pendingNo":       item["pendingNo"],
		"confirmedNo":     item["confirmedNo"],
		"declinedNo":      item["declinedNo"],
		"totalNo":         item["totalNo"],
		"totalSaleAmount": item["totalSaleAmount"],
		"totalComm":       item["totalComm"],
	}
}

func awinPublisherRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"displayUrl": item["displayUrl"],
		"kind":       item["kind"],
		"status":     item["status"],
	}
}
