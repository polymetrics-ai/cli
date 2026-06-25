package clazar

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Clazar API resource path (relative to
// base_url) it reads from, plus the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Clazar list endpoint path segment (e.g. "buyers").
	resource string
	// mapRecord flattens a raw Clazar object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// clazarStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in clazarStreams; the read path
// is fully data-driven from this table. Mirrors the Airbyte source-clazar
// manifest core streams (buyers, listings, contracts, opportunities,
// private_offers), all of which share the page-increment paginator, the
// `results` record selector, the `id` primary key, and the `last_modified_at`
// incremental cursor.
var clazarStreamEndpoints = map[string]streamEndpoint{
	"buyers":         {resource: "buyers", mapRecord: clazarBuyerRecord},
	"listings":       {resource: "listings", mapRecord: clazarListingRecord},
	"contracts":      {resource: "contracts", mapRecord: clazarContractRecord},
	"opportunities":  {resource: "opportunities", mapRecord: clazarOpportunityRecord},
	"private_offers": {resource: "private_offers", mapRecord: clazarPrivateOfferRecord},
}

// clazarStreams returns the connector's published stream catalog. Every Clazar
// object exposes a string id and an RFC3339(.us)Z last_modified_at timestamp, so
// the primary key is ["id"] and the incremental cursor field is
// ["last_modified_at"] across the board.
func clazarStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "buyers",
			Description:  "Clazar buyers (customers acquired through cloud marketplaces).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified_at"},
			Fields:       clazarBuyerFields(),
		},
		{
			Name:         "listings",
			Description:  "Clazar cloud marketplace listings.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified_at"},
			Fields:       clazarListingFields(),
		},
		{
			Name:         "contracts",
			Description:  "Clazar marketplace contracts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified_at"},
			Fields:       clazarContractFields(),
		},
		{
			Name:         "opportunities",
			Description:  "Clazar co-sell opportunities.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified_at"},
			Fields:       clazarOpportunityFields(),
		},
		{
			Name:         "private_offers",
			Description:  "Clazar private offers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_modified_at"},
			Fields:       clazarPrivateOfferFields(),
		},
	}
}

func clazarBuyerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "cloud", Type: "string"},
		{Name: "domain", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "listing_id", Type: "string"},
		{Name: "cloud_account_id", Type: "string"},
		{Name: "latest_contract_id", Type: "string"},
		{Name: "last_modified_at", Type: "timestamp"},
	}
}

func clazarListingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "cloud", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "cloud_id", Type: "string"},
		{Name: "cloud_url", Type: "string"},
		{Name: "eula_type", Type: "string"},
		{Name: "short_description", Type: "string"},
		{Name: "long_description", Type: "string"},
		{Name: "last_modified_at", Type: "timestamp"},
	}
}

func clazarContractFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "cloud", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "buyer_id", Type: "string"},
		{Name: "cloud_id", Type: "string"},
		{Name: "listing_id", Type: "string"},
		{Name: "offer_type", Type: "string"},
		{Name: "duration", Type: "string"},
		{Name: "start_at", Type: "timestamp"},
		{Name: "end_at", Type: "timestamp"},
		{Name: "accepted_at", Type: "timestamp"},
		{Name: "auto_renew", Type: "boolean"},
		{Name: "latest_offer_id", Type: "string"},
		{Name: "last_modified_at", Type: "timestamp"},
	}
}

func clazarOpportunityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "cloud", Type: "string"},
		{Name: "stage", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "cloud_id", Type: "string"},
		{Name: "accept_by", Type: "timestamp"},
		{Name: "customer_company", Type: "string"},
		{Name: "customer_website", Type: "string"},
		{Name: "target_close_date", Type: "timestamp"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "last_modified_at", Type: "timestamp"},
	}
}

func clazarPrivateOfferFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "cloud", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "archived", Type: "string"},
		{Name: "cloud_id", Type: "string"},
		{Name: "duration", Type: "string"},
		{Name: "eula_type", Type: "string"},
		{Name: "listing_id", Type: "string"},
		{Name: "offer_type", Type: "string"},
		{Name: "accepted_at", Type: "timestamp"},
		{Name: "published_at", Type: "timestamp"},
		{Name: "expiration_at", Type: "timestamp"},
		{Name: "last_modified_at", Type: "timestamp"},
	}
}

func clazarBuyerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"cloud":              item["cloud"],
		"domain":             item["domain"],
		"status":             item["status"],
		"listing_id":         item["listing_id"],
		"cloud_account_id":   item["cloud_account_id"],
		"latest_contract_id": item["latest_contract_id"],
		"last_modified_at":   item["last_modified_at"],
	}
}

func clazarListingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"cloud":             item["cloud"],
		"title":             item["title"],
		"status":            item["status"],
		"cloud_id":          item["cloud_id"],
		"cloud_url":         item["cloud_url"],
		"eula_type":         item["eula_type"],
		"short_description": item["short_description"],
		"long_description":  item["long_description"],
		"last_modified_at":  item["last_modified_at"],
	}
}

func clazarContractRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"cloud":            item["cloud"],
		"status":           item["status"],
		"buyer_id":         item["buyer_id"],
		"cloud_id":         item["cloud_id"],
		"listing_id":       item["listing_id"],
		"offer_type":       item["offer_type"],
		"duration":         item["duration"],
		"start_at":         item["start_at"],
		"end_at":           item["end_at"],
		"accepted_at":      item["accepted_at"],
		"auto_renew":       item["auto_renew"],
		"latest_offer_id":  item["latest_offer_id"],
		"last_modified_at": item["last_modified_at"],
	}
}

func clazarOpportunityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"cloud":             item["cloud"],
		"stage":             item["stage"],
		"title":             item["title"],
		"status":            item["status"],
		"cloud_id":          item["cloud_id"],
		"accept_by":         item["accept_by"],
		"customer_company":  item["customer_company"],
		"customer_website":  item["customer_website"],
		"target_close_date": item["target_close_date"],
		"created_at":        item["created_at"],
		"last_modified_at":  item["last_modified_at"],
	}
}

func clazarPrivateOfferRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"cloud":            item["cloud"],
		"status":           item["status"],
		"archived":         item["archived"],
		"cloud_id":         item["cloud_id"],
		"duration":         item["duration"],
		"eula_type":        item["eula_type"],
		"listing_id":       item["listing_id"],
		"offer_type":       item["offer_type"],
		"accepted_at":      item["accepted_at"],
		"published_at":     item["published_at"],
		"expiration_at":    item["expiration_at"],
		"last_modified_at": item["last_modified_at"],
	}
}
