package breezyhr

import "polymetrics.ai/internal/connectors"

// breezyStreams returns the connector's published stream catalog. Breezy exposes
// positions for a company, the candidates on each position, and the company's
// pipelines. Every object is keyed by Breezy's "_id" (surfaced as id /
// position_id), and there is no incremental cursor in the public API, so streams
// run full-refresh.
func breezyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "positions",
			Description: "Breezy positions (job openings) for the configured company.",
			PrimaryKey:  []string{"position_id"},
			Fields:      breezyPositionFields(),
		},
		{
			Name:        "candidates",
			Description: "Breezy candidates across every position in the company.",
			PrimaryKey:  []string{"id"},
			Fields:      breezyCandidateFields(),
		},
		{
			Name:        "pipelines",
			Description: "Breezy hiring pipelines configured for the company.",
			PrimaryKey:  []string{"id"},
			Fields:      breezyPipelineFields(),
		},
	}
}

func breezyPositionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "position_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "department", Type: "string"},
		{Name: "org_type", Type: "string"},
		{Name: "pipeline_id", Type: "string"},
		{Name: "country_id", Type: "string"},
		{Name: "country_name", Type: "string"},
		{Name: "creation_date", Type: "string"},
		{Name: "updated_date", Type: "string"},
	}
}

func breezyCandidateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "position_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "phone_number", Type: "string"},
		{Name: "headline", Type: "string"},
		{Name: "origin", Type: "string"},
		{Name: "stage", Type: "string"},
		{Name: "creation_date", Type: "string"},
		{Name: "updated_date", Type: "string"},
	}
}

func breezyPipelineFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

// breezyPositionRecord flattens a raw Breezy position object. Breezy keys the
// object on "_id"; the manifest surfaces it as position_id, and nested
// type.name / location.country.* are flattened to scalar columns.
func breezyPositionRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"position_id":   firstString(item, "_id", "position_id"),
		"name":          item["name"],
		"type":          nestedName(item["type"]),
		"state":         item["state"],
		"department":    item["department"],
		"org_type":      item["org_type"],
		"pipeline_id":   item["pipeline_id"],
		"creation_date": item["creation_date"],
		"updated_date":  item["updated_date"],
	}
	if loc, ok := item["location"].(map[string]any); ok {
		if country, ok := loc["country"].(map[string]any); ok {
			rec["country_id"] = country["id"]
			rec["country_name"] = country["name"]
		}
	}
	return rec
}

// breezyCandidateRecord flattens a raw Breezy candidate object. positionID is the
// parent position the candidate was read under (the candidate endpoint is a
// substream of positions and the object itself does not always echo it).
func breezyCandidateRecord(item map[string]any, positionID string) connectors.Record {
	pid := positionID
	if pid == "" {
		pid = stringField(item, "position_id")
	}
	return connectors.Record{
		"id":            firstString(item, "_id", "id"),
		"position_id":   pid,
		"name":          item["name"],
		"email_address": item["email_address"],
		"phone_number":  item["phone_number"],
		"headline":      item["headline"],
		"origin":        item["origin"],
		"stage":         nestedName(item["stage"]),
		"creation_date": item["creation_date"],
		"updated_date":  item["updated_date"],
	}
}

// breezyPipelineRecord flattens a raw Breezy pipeline object to id/name.
func breezyPipelineRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   firstString(item, "_id", "id"),
		"name": item["name"],
	}
}

// nestedName flattens a {"name": ...} sub-object to its name string, leaving an
// already-scalar value untouched (Breezy returns type/stage either way across
// endpoints).
func nestedName(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return t["name"]
	default:
		return v
	}
}

// firstString returns the first present non-empty string-able field among keys.
func firstString(item map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := item[k]; ok && v != nil {
			return v
		}
	}
	return nil
}
