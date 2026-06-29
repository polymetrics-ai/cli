package easypromos

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Easypromos API resource it reads from,
// the record mapper that flattens its objects, and whether it is a per-promotion
// sub-stream (path /{resource}/{promotion_id}) requiring the promotion_id config.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "promotions", "users").
	resource string
	// perPromotion marks streams whose path is /{resource}/{promotion_id}.
	perPromotion bool
	// mapRecord flattens a raw Easypromos object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// easypromosStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in easypromosStreams; the
// read path is fully data-driven from this table.
//
// promotions and organizing_brands are account-level lists. The remaining streams
// are scoped to a single promotion: their path embeds the promotion_id (taken
// from config), mirroring the upstream upstream connector's SubstreamPartitionRouter.
var easypromosStreamEndpoints = map[string]streamEndpoint{
	"promotions":        {resource: "promotions", mapRecord: easypromosPromotionRecord},
	"organizing_brands": {resource: "organizing_brands", mapRecord: easypromosOrganizingBrandRecord},
	"stages":            {resource: "stages", perPromotion: true, mapRecord: easypromosStageRecord},
	"users":             {resource: "users", perPromotion: true, mapRecord: easypromosUserRecord},
	"participations":    {resource: "participations", perPromotion: true, mapRecord: easypromosParticipationRecord},
	"prizes":            {resource: "prizes", perPromotion: true, mapRecord: easypromosPrizeRecord},
}

// easypromosStreams returns the connector's published stream catalog. Every
// Easypromos object exposes a string id, so the primary key is ["id"] across the
// board. The API only supports full_refresh, so no cursor fields are published.
func easypromosStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "promotions",
			Description: "Easypromos promotions (contests, giveaways, campaigns).",
			PrimaryKey:  []string{"id"},
			Fields:      easypromosPromotionFields(),
		},
		{
			Name:        "organizing_brands",
			Description: "Brands organizing the account's promotions.",
			PrimaryKey:  []string{"id"},
			Fields:      easypromosOrganizingBrandFields(),
		},
		{
			Name:        "stages",
			Description: "Stages within a promotion. Requires config promotion_id.",
			PrimaryKey:  []string{"id"},
			Fields:      easypromosStageFields(),
		},
		{
			Name:        "users",
			Description: "Users registered in a promotion. Requires config promotion_id.",
			PrimaryKey:  []string{"id"},
			Fields:      easypromosUserFields(),
		},
		{
			Name:        "participations",
			Description: "Participations submitted in a promotion. Requires config promotion_id.",
			PrimaryKey:  []string{"id"},
			Fields:      easypromosParticipationFields(),
		},
		{
			Name:        "prizes",
			Description: "Prizes awarded in a promotion. Requires config promotion_id.",
			PrimaryKey:  []string{"id"},
			Fields:      easypromosPrizeFields(),
		},
	}
}

func easypromosPromotionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "promotion_type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "default_language", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "start_date", Type: "string"},
		{Name: "end_date", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "organizing_brand_id", Type: "string"},
		{Name: "organizing_brand_name", Type: "string"},
	}
}

func easypromosOrganizingBrandFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func easypromosStageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "visible", Type: "boolean"},
		{Name: "start_date", Type: "string"},
		{Name: "end_date", Type: "string"},
	}
}

func easypromosUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "external_id", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "nickname", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "login_type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "promotion_id", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func easypromosParticipationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "promotion_id", Type: "string"},
		{Name: "stage_id", Type: "string"},
		{Name: "ip", Type: "string"},
		{Name: "user_agent", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func easypromosPrizeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "participation_id", Type: "string"},
		{Name: "stage_id", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "redeem_url", Type: "string"},
		{Name: "download_url", Type: "string"},
		{Name: "prize_type_id", Type: "string"},
		{Name: "prize_type_name", Type: "string"},
		{Name: "created", Type: "string"},
	}
}

func easypromosPromotionRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":               item["id"],
		"title":            item["title"],
		"description":      item["description"],
		"promotion_type":   item["promotion_type"],
		"status":           item["status"],
		"default_language": item["default_language"],
		"timezone":         item["timezone"],
		"url":              item["url"],
		"start_date":       item["start_date"],
		"end_date":         item["end_date"],
		"created":          item["created"],
	}
	if brand, ok := item["organizing_brand"].(map[string]any); ok {
		rec["organizing_brand_id"] = brand["id"]
		rec["organizing_brand_name"] = brand["name"]
	}
	return rec
}

func easypromosOrganizingBrandRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"name": item["name"],
	}
}

func easypromosStageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"type":       item["type"],
		"visible":    item["visible"],
		"start_date": item["start_date"],
		"end_date":   item["end_date"],
	}
}

func easypromosUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"external_id":  item["external_id"],
		"first_name":   item["first_name"],
		"last_name":    item["last_name"],
		"nickname":     item["nickname"],
		"email":        item["email"],
		"country":      item["country"],
		"language":     item["language"],
		"login_type":   item["login_type"],
		"status":       item["status"],
		"promotion_id": item["promotion_id"],
		"created":      item["created"],
	}
}

func easypromosParticipationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"user_id":      item["user_id"],
		"promotion_id": item["promotion_id"],
		"stage_id":     item["stage_id"],
		"ip":           item["ip"],
		"user_agent":   item["user_agent"],
		"created":      item["created"],
	}
}

func easypromosPrizeRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":               item["id"],
		"participation_id": item["participation_id"],
		"stage_id":         item["stage_id"],
		"code":             item["code"],
		"redeem_url":       item["redeem_url"],
		"download_url":     item["download_url"],
		"created":          item["created"],
	}
	if pt, ok := item["prize_type"].(map[string]any); ok {
		rec["prize_type_id"] = pt["id"]
		rec["prize_type_name"] = pt["name"]
	}
	return rec
}
