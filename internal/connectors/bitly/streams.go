package bitly

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to its Bitly list endpoint shape: the
// resource path (relative to base_url), the JSON key under which the records
// array lives, whether the endpoint paginates via pagination.next, and the
// record mapper.
type streamEndpoint struct {
	// resource is the list endpoint path. For bitlinks it is a template that is
	// resolved with the group_guid at read time (see groupScoped).
	resource string
	// recordsKey is the top-level JSON key holding the array of objects.
	recordsKey string
	// paginated is true when the endpoint returns a pagination.next absolute URL
	// to walk subsequent pages (currently only bitlinks).
	paginated bool
	// groupScoped is true when the path needs a {group_guid} substituted in.
	groupScoped bool
	// mapRecord flattens a raw Bitly object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// bitlyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in bitlyStreams; the read path
// is fully data-driven from this table.
var bitlyStreamEndpoints = map[string]streamEndpoint{
	"organizations": {resource: "organizations", recordsKey: "organizations", mapRecord: bitlyOrganizationRecord},
	"groups":        {resource: "groups", recordsKey: "groups", mapRecord: bitlyGroupRecord},
	"campaigns":     {resource: "campaigns", recordsKey: "campaigns", mapRecord: bitlyCampaignRecord},
	"bitlinks":      {resource: "groups/{group_guid}/bitlinks", recordsKey: "links", paginated: true, groupScoped: true, mapRecord: bitlyBitlinkRecord},
}

// bitlyStreams returns the connector's published stream catalog. Bitly resources
// are keyed by a string guid (or id for bitlinks); none expose an incremental
// cursor field in these core list endpoints, so reads are full refresh.
func bitlyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "Bitly organizations the access token can see.",
			PrimaryKey:  []string{"guid"},
			Fields:      bitlyOrganizationFields(),
		},
		{
			Name:        "groups",
			Description: "Bitly groups across the accessible organizations.",
			PrimaryKey:  []string{"guid"},
			Fields:      bitlyGroupFields(),
		},
		{
			Name:        "campaigns",
			Description: "Bitly campaigns.",
			PrimaryKey:  []string{"guid"},
			Fields:      bitlyCampaignFields(),
		},
		{
			Name:        "bitlinks",
			Description: "Bitlinks (short links) for the configured group.",
			PrimaryKey:  []string{"id"},
			Fields:      bitlyBitlinkFields(),
		},
	}
}

func bitlyOrganizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "guid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "tier", Type: "string"},
		{Name: "tier_family", Type: "string"},
		{Name: "tier_display_name", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func bitlyGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "guid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "organization_guid", Type: "string"},
		{Name: "bsds", Type: "array"},
		{Name: "is_active", Type: "boolean"},
		{Name: "role", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func bitlyCampaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "guid", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "group_guid", Type: "string"},
		{Name: "channel_guids", Type: "array"},
		{Name: "created", Type: "string"},
		{Name: "modified", Type: "string"},
	}
}

func bitlyBitlinkFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "link", Type: "string"},
		{Name: "long_url", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "tags", Type: "array"},
		{Name: "deeplinks", Type: "array"},
		{Name: "references", Type: "object"},
		{Name: "created_at", Type: "string"},
	}
}

func bitlyOrganizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"guid":              item["guid"],
		"name":              item["name"],
		"is_active":         item["is_active"],
		"tier":              item["tier"],
		"tier_family":       item["tier_family"],
		"tier_display_name": item["tier_display_name"],
		"role":              item["role"],
		"created":           item["created"],
		"modified":          item["modified"],
	}
}

func bitlyGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"guid":              item["guid"],
		"name":              item["name"],
		"organization_guid": item["organization_guid"],
		"bsds":              item["bsds"],
		"is_active":         item["is_active"],
		"role":              item["role"],
		"created":           item["created"],
		"modified":          item["modified"],
	}
}

func bitlyCampaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"guid":          item["guid"],
		"name":          item["name"],
		"description":   item["description"],
		"group_guid":    item["group_guid"],
		"channel_guids": item["channel_guids"],
		"created":       item["created"],
		"modified":      item["modified"],
	}
}

func bitlyBitlinkRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"link":       item["link"],
		"long_url":   item["long_url"],
		"title":      item["title"],
		"archived":   item["archived"],
		"tags":       item["tags"],
		"deeplinks":  item["deeplinks"],
		"references": item["references"],
		"created_at": item["created_at"],
	}
}
