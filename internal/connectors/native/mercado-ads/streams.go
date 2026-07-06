package mercadoads

import "polymetrics.ai/internal/connectors"

// streamKind distinguishes the two read shapes this connector handles:
//   - advertisersKind: a single GET filtered by product_id, records at
//     body["advertisers"], no pagination.
//   - metricsKind: an offset/limit paginated GET against a per-campaign metrics
//     endpoint, records at body["metrics"].
type streamKind int

const (
	advertisersKind streamKind = iota
	metricsKind
)

// streamDef describes how to read a stream and how to flatten its records.
type streamDef struct {
	kind streamKind
	// productID is the product_id filter for advertiser streams (BADS, DISPLAY, PADS).
	productID string
	// recordsPath is the JSON dotted path to the records array.
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

// streamDefs is the per-stream routing table. The advertiser streams are the
// three top-level entry points of the Mercado Ads API (one per ad product); the
// *_campaigns_metrics streams demonstrate the offset-paginated metrics shape.
var streamDefs = map[string]streamDef{
	"brand_advertisers":   {kind: advertisersKind, productID: "BADS", recordsPath: "advertisers", mapRecord: advertiserRecord},
	"display_advertisers": {kind: advertisersKind, productID: "DISPLAY", recordsPath: "advertisers", mapRecord: advertiserRecord},
	"product_advertisers": {kind: advertisersKind, productID: "PADS", recordsPath: "advertisers", mapRecord: advertiserRecord},

	"brand_campaigns_metrics":   {kind: metricsKind, recordsPath: "metrics", mapRecord: campaignMetricRecord},
	"display_campaigns_metrics": {kind: metricsKind, recordsPath: "metrics", mapRecord: campaignMetricRecord},
	"product_campaigns_metrics": {kind: metricsKind, recordsPath: "metrics", mapRecord: campaignMetricRecord},
}

// metricsEndpointTemplate maps a metrics stream to the path template (formatted
// with advertiser_id then campaign_id) for its per-campaign metrics endpoint.
var metricsEndpointTemplate = map[string]string{
	"brand_campaigns_metrics":   "advertising/advertisers/%s/brand_ads/campaigns/%s/metrics",
	"display_campaigns_metrics": "advertising/advertisers/%s/display_ads/campaigns/%s/metrics",
	"product_campaigns_metrics": "advertising/advertisers/%s/product_ads/campaigns/%s/metrics",
}

// streams returns the connector's published stream catalog.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "brand_advertisers",
			Description: "Mercado Ads Brand Ads advertisers (product_id=BADS).",
			PrimaryKey:  []string{"advertiser_id"},
			Fields:      advertiserFields(),
		},
		{
			Name:        "display_advertisers",
			Description: "Mercado Ads Display Ads advertisers (product_id=DISPLAY).",
			PrimaryKey:  []string{"advertiser_id"},
			Fields:      advertiserFields(),
		},
		{
			Name:        "product_advertisers",
			Description: "Mercado Ads Product Ads advertisers (product_id=PADS).",
			PrimaryKey:  []string{"advertiser_id"},
			Fields:      advertiserFields(),
		},
		{
			Name:         "brand_campaigns_metrics",
			Description:  "Daily Brand Ads campaign metrics.",
			PrimaryKey:   []string{"date", "advertiser_id", "campaign_id"},
			CursorFields: []string{"date"},
			Fields:       campaignMetricFields(),
		},
		{
			Name:         "display_campaigns_metrics",
			Description:  "Daily Display Ads campaign metrics.",
			PrimaryKey:   []string{"date", "advertiser_id", "campaign_id"},
			CursorFields: []string{"date"},
			Fields:       campaignMetricFields(),
		},
		{
			Name:         "product_campaigns_metrics",
			Description:  "Daily Product Ads campaign metrics.",
			PrimaryKey:   []string{"date", "advertiser_id", "campaign_id"},
			CursorFields: []string{"date"},
			Fields:       campaignMetricFields(),
		},
	}
}

func advertiserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "advertiser_id", Type: "integer"},
		{Name: "advertiser_name", Type: "string"},
		{Name: "account_name", Type: "string"},
		{Name: "site_id", Type: "string"},
	}
}

func campaignMetricFields() []connectors.Field {
	return []connectors.Field{
		{Name: "date", Type: "string"},
		{Name: "advertiser_id", Type: "string"},
		{Name: "campaign_id", Type: "string"},
		{Name: "prints", Type: "number"},
		{Name: "clicks", Type: "number"},
		{Name: "cost", Type: "number"},
		{Name: "ctr", Type: "number"},
		{Name: "cpc", Type: "number"},
		{Name: "acos", Type: "number"},
		{Name: "total_amount", Type: "number"},
		{Name: "direct_amount", Type: "number"},
		{Name: "indirect_amount", Type: "number"},
		{Name: "units_quantity", Type: "number"},
	}
}

func advertiserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"advertiser_id":   item["advertiser_id"],
		"advertiser_name": item["advertiser_name"],
		"account_name":    item["account_name"],
		"site_id":         item["site_id"],
	}
}

func campaignMetricRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{}
	for _, k := range []string{
		"date", "advertiser_id", "campaign_id", "prints", "clicks", "cost",
		"ctr", "cpc", "acos", "total_amount", "direct_amount",
		"indirect_amount", "units_quantity",
	} {
		if v, ok := item[k]; ok {
			rec[k] = v
		}
	}
	return rec
}
