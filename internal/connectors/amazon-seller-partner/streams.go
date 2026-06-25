package amazonsellerpartner

import (
	"net/url"
	"strconv"

	"polymetrics.ai/internal/connectors"
)

// streamEndpoint maps a stream name to the SP-API resource path it reads from,
// the JSON paths to the records array and the next-page token, the query param
// that carries that token back, plus the per-stream record mapper, base-query
// builder, and fixture-item generator. The read path is fully data-driven from
// this table (mirrors the stripe stream-endpoint table).
type streamEndpoint struct {
	// resource is the SP-API path relative to base_url (e.g. "orders/v0/orders").
	resource string
	// recordsPath is the dotted JSON path to the records array in the response
	// body (e.g. "payload.Orders").
	recordsPath string
	// tokenPath is the dotted JSON path to the next-page token (e.g.
	// "payload.NextToken" for orders, "pagination.nextToken" for inventory).
	tokenPath string
	// tokenParam is the query parameter the token is resupplied as on the next
	// request (e.g. "NextToken" or "nextToken").
	tokenParam string
	// mapRecord flattens a raw SP-API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// baseQuery builds the first-page query (filters, marketplace, page size).
	baseQuery func(req connectors.ReadRequest) (url.Values, error)
	// fixtureItem produces a deterministic record source for fixture mode.
	fixtureItem func(i int) map[string]any
}

// spStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in spStreams.
var spStreamEndpoints = map[string]streamEndpoint{
	"orders": {
		resource:    "orders/v0/orders",
		recordsPath: "payload.Orders",
		tokenPath:   "payload.NextToken",
		tokenParam:  "NextToken",
		mapRecord:   orderRecord,
		baseQuery:   ordersBaseQuery,
		fixtureItem: orderFixture,
	},
	"inventory_summaries": {
		resource:    "fba/inventory/v1/summaries",
		recordsPath: "payload.inventorySummaries",
		tokenPath:   "pagination.nextToken",
		tokenParam:  "nextToken",
		mapRecord:   inventoryRecord,
		baseQuery:   inventoryBaseQuery,
		fixtureItem: inventoryFixture,
	},
	"financial_event_groups": {
		resource:    "finances/v0/financialEventGroups",
		recordsPath: "payload.FinancialEventGroupList",
		tokenPath:   "payload.NextToken",
		tokenParam:  "NextToken",
		mapRecord:   financialEventGroupRecord,
		baseQuery:   financialEventGroupsBaseQuery,
		fixtureItem: financialEventGroupFixture,
	},
}

// spStreams returns the connector's published stream catalog.
func spStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "orders",
			Description:  "Amazon Selling Partner orders (Orders API v0).",
			PrimaryKey:   []string{"AmazonOrderId"},
			CursorFields: []string{"LastUpdateDate"},
			Fields:       orderFields(),
		},
		{
			Name:         "inventory_summaries",
			Description:  "FBA inventory summaries (FBA Inventory API v1).",
			PrimaryKey:   []string{"sellerSku"},
			CursorFields: []string{"lastUpdatedTime"},
			Fields:       inventoryFields(),
		},
		{
			Name:         "financial_event_groups",
			Description:  "Financial event groups / settlements (Finances API v0).",
			PrimaryKey:   []string{"FinancialEventGroupId"},
			CursorFields: []string{"FinancialEventGroupEnd"},
			Fields:       financialEventGroupFields(),
		},
	}
}

// --- orders ---

func ordersBaseQuery(req connectors.ReadRequest) (url.Values, error) {
	q := url.Values{}
	q.Set("MarketplaceIds", marketplaceID(req.Config))
	pageSize, err := spPageSize(req.Config)
	if err != nil {
		return nil, err
	}
	q.Set("MaxResultsPerPage", strconv.Itoa(pageSize))
	lower, err := spReplicationStart(req)
	if err != nil {
		return nil, err
	}
	if lower == "" {
		lower = spDefaultCreatedAfter()
	}
	// Use LastUpdatedAfter so the incremental cursor (LastUpdateDate) advances
	// the window on subsequent syncs.
	q.Set("LastUpdatedAfter", lower)
	return q, nil
}

func orderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "AmazonOrderId", Type: "string"},
		{Name: "SellerOrderId", Type: "string"},
		{Name: "PurchaseDate", Type: "string"},
		{Name: "LastUpdateDate", Type: "string"},
		{Name: "OrderStatus", Type: "string"},
		{Name: "FulfillmentChannel", Type: "string"},
		{Name: "SalesChannel", Type: "string"},
		{Name: "OrderType", Type: "string"},
		{Name: "NumberOfItemsShipped", Type: "integer"},
		{Name: "NumberOfItemsUnshipped", Type: "integer"},
		{Name: "MarketplaceId", Type: "string"},
		{Name: "OrderTotal", Type: "object"},
		{Name: "IsBusinessOrder", Type: "boolean"},
		{Name: "IsPrime", Type: "boolean"},
	}
}

func orderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"AmazonOrderId":          item["AmazonOrderId"],
		"SellerOrderId":          item["SellerOrderId"],
		"PurchaseDate":           item["PurchaseDate"],
		"LastUpdateDate":         item["LastUpdateDate"],
		"OrderStatus":            item["OrderStatus"],
		"FulfillmentChannel":     item["FulfillmentChannel"],
		"SalesChannel":           item["SalesChannel"],
		"OrderType":              item["OrderType"],
		"NumberOfItemsShipped":   item["NumberOfItemsShipped"],
		"NumberOfItemsUnshipped": item["NumberOfItemsUnshipped"],
		"MarketplaceId":          item["MarketplaceId"],
		"OrderTotal":             item["OrderTotal"],
		"IsBusinessOrder":        item["IsBusinessOrder"],
		"IsPrime":                item["IsPrime"],
	}
}

func orderFixture(i int) map[string]any {
	return map[string]any{
		"AmazonOrderId":          "111-fixture-" + strconv.Itoa(i),
		"SellerOrderId":          "so-fixture-" + strconv.Itoa(i),
		"PurchaseDate":           spFixtureUpdated,
		"LastUpdateDate":         spFixtureUpdated,
		"OrderStatus":            "Shipped",
		"FulfillmentChannel":     "AFN",
		"SalesChannel":           "Amazon.com",
		"OrderType":              "StandardOrder",
		"NumberOfItemsShipped":   int64(i),
		"NumberOfItemsUnshipped": int64(0),
		"MarketplaceId":          defaultMarketplaceID,
		"OrderTotal":             map[string]any{"CurrencyCode": "USD", "Amount": "10.00"},
		"IsBusinessOrder":        false,
		"IsPrime":                true,
	}
}

// --- inventory_summaries ---

func inventoryBaseQuery(req connectors.ReadRequest) (url.Values, error) {
	mp := marketplaceID(req.Config)
	q := url.Values{}
	q.Set("granularityType", "Marketplace")
	q.Set("granularityId", mp)
	q.Set("marketplaceIds", mp)
	q.Set("details", "true")
	lower, err := spReplicationStart(req)
	if err != nil {
		return nil, err
	}
	if lower != "" {
		q.Set("startDateTime", lower)
	}
	return q, nil
}

func inventoryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "sellerSku", Type: "string"},
		{Name: "fnSku", Type: "string"},
		{Name: "asin", Type: "string"},
		{Name: "condition", Type: "string"},
		{Name: "productName", Type: "string"},
		{Name: "totalQuantity", Type: "integer"},
		{Name: "lastUpdatedTime", Type: "string"},
		{Name: "inventoryDetails", Type: "object"},
	}
}

func inventoryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"sellerSku":        item["sellerSku"],
		"fnSku":            item["fnSku"],
		"asin":             item["asin"],
		"condition":        item["condition"],
		"productName":      item["productName"],
		"totalQuantity":    item["totalQuantity"],
		"lastUpdatedTime":  item["lastUpdatedTime"],
		"inventoryDetails": item["inventoryDetails"],
	}
}

func inventoryFixture(i int) map[string]any {
	return map[string]any{
		"sellerSku":       "SKU-fixture-" + strconv.Itoa(i),
		"fnSku":           "X00FIXTURE" + strconv.Itoa(i),
		"asin":            "B00FIXTURE" + strconv.Itoa(i),
		"condition":       "NewItem",
		"productName":     "Fixture Product " + strconv.Itoa(i),
		"totalQuantity":   int64(10 * i),
		"lastUpdatedTime": spFixtureUpdated,
		"inventoryDetails": map[string]any{
			"fulfillableQuantity": int64(10 * i),
		},
	}
}

// --- financial_event_groups ---

func financialEventGroupsBaseQuery(req connectors.ReadRequest) (url.Values, error) {
	q := url.Values{}
	pageSize, err := spPageSize(req.Config)
	if err != nil {
		return nil, err
	}
	q.Set("MaxResultsPerPage", strconv.Itoa(pageSize))
	lower, err := spReplicationStart(req)
	if err != nil {
		return nil, err
	}
	if lower == "" {
		lower = spDefaultCreatedAfter()
	}
	q.Set("FinancialEventGroupStartedAfter", lower)
	return q, nil
}

func financialEventGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "FinancialEventGroupId", Type: "string"},
		{Name: "ProcessingStatus", Type: "string"},
		{Name: "FundTransferStatus", Type: "string"},
		{Name: "OriginalTotal", Type: "object"},
		{Name: "ConvertedTotal", Type: "object"},
		{Name: "FundTransferDate", Type: "string"},
		{Name: "TraceId", Type: "string"},
		{Name: "AccountTail", Type: "string"},
		{Name: "BeginningBalance", Type: "object"},
		{Name: "FinancialEventGroupStart", Type: "string"},
		{Name: "FinancialEventGroupEnd", Type: "string"},
	}
}

func financialEventGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"FinancialEventGroupId":    item["FinancialEventGroupId"],
		"ProcessingStatus":         item["ProcessingStatus"],
		"FundTransferStatus":       item["FundTransferStatus"],
		"OriginalTotal":            item["OriginalTotal"],
		"ConvertedTotal":           item["ConvertedTotal"],
		"FundTransferDate":         item["FundTransferDate"],
		"TraceId":                  item["TraceId"],
		"AccountTail":              item["AccountTail"],
		"BeginningBalance":         item["BeginningBalance"],
		"FinancialEventGroupStart": item["FinancialEventGroupStart"],
		"FinancialEventGroupEnd":   item["FinancialEventGroupEnd"],
	}
}

func financialEventGroupFixture(i int) map[string]any {
	return map[string]any{
		"FinancialEventGroupId":    "feg-fixture-" + strconv.Itoa(i),
		"ProcessingStatus":         "Closed",
		"FundTransferStatus":       "Successful",
		"OriginalTotal":            map[string]any{"CurrencyCode": "USD", "CurrencyAmount": "100.00"},
		"ConvertedTotal":           map[string]any{"CurrencyCode": "USD", "CurrencyAmount": "100.00"},
		"FundTransferDate":         spFixtureUpdated,
		"TraceId":                  "trace-" + strconv.Itoa(i),
		"AccountTail":              "1234",
		"BeginningBalance":         map[string]any{"CurrencyCode": "USD", "CurrencyAmount": "0.00"},
		"FinancialEventGroupStart": spFixtureUpdated,
		"FinancialEventGroupEnd":   spFixtureUpdated,
	}
}
