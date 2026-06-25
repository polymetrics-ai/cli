package castoredc

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Castor EDC API resource path
// (relative to base_url), the HAL collection key under "_embedded" where the
// records live, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Castor list endpoint path segment (e.g. "study").
	resource string
	// embeddedKey is the key under "_embedded" holding the record array.
	// Castor's HAL responses nest collections under different names than the
	// endpoint (e.g. /country -> _embedded.countries).
	embeddedKey string
	// mapRecord flattens a raw Castor object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// castorStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in castorStreams; the read
// path is fully data-driven from this table.
var castorStreamEndpoints = map[string]streamEndpoint{
	"study":      {resource: "study", embeddedKey: "study", mapRecord: castorStudyRecord},
	"user":       {resource: "user", embeddedKey: "user", mapRecord: castorUserRecord},
	"country":    {resource: "country", embeddedKey: "countries", mapRecord: castorCountryRecord},
	"study_user": {resource: "user", embeddedKey: "user", mapRecord: castorUserRecord},
	"audit_trail": {
		resource: "audit-trail", embeddedKey: "audit_trail", mapRecord: castorAuditTrailRecord,
	},
}

// castorStreams returns the connector's published stream catalog.
func castorStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "study",
			Description:  "Castor EDC studies accessible to the authenticated account.",
			PrimaryKey:   []string{"study_id"},
			CursorFields: []string{"updated_on"},
			Fields:       castorStudyFields(),
		},
		{
			Name:         "user",
			Description:  "Castor EDC users.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_login"},
			Fields:       castorUserFields(),
		},
		{
			Name:        "country",
			Description: "Reference list of countries supported by Castor EDC.",
			PrimaryKey:  []string{"id"},
			Fields:      castorCountryFields(),
		},
		{
			Name:         "audit_trail",
			Description:  "Castor EDC audit trail events.",
			PrimaryKey:   []string{"uuid"},
			CursorFields: []string{"datetime"},
			Fields:       castorAuditTrailFields(),
		},
	}
}

func castorStudyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "study_id", Type: "string"},
		{Name: "crf_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "live", Type: "boolean"},
		{Name: "randomization_enabled", Type: "boolean"},
		{Name: "gcp_enabled", Type: "boolean"},
		{Name: "surveys_enabled", Type: "boolean"},
		{Name: "premium_support_enabled", Type: "boolean"},
		{Name: "main_contact", Type: "string"},
		{Name: "institute_id", Type: "string"},
		{Name: "created_on", Type: "string"},
		{Name: "updated_on", Type: "string"},
		{Name: "duration", Type: "integer"},
	}
}

func castorUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "full_name", Type: "string"},
		{Name: "is_active", Type: "boolean"},
		{Name: "institute", Type: "string"},
		{Name: "last_login", Type: "string"},
	}
}

func castorCountryFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "country_id", Type: "string"},
		{Name: "country_name", Type: "string"},
		{Name: "country_tld", Type: "string"},
		{Name: "country_cca2", Type: "string"},
		{Name: "country_cca3", Type: "string"},
	}
}

func castorAuditTrailFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uuid", Type: "string"},
		{Name: "datetime", Type: "string"},
		{Name: "event_type", Type: "string"},
		{Name: "user_id", Type: "string"},
		{Name: "user_name", Type: "string"},
		{Name: "user_email", Type: "string"},
		{Name: "event_details", Type: "object"},
	}
}

func castorStudyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"study_id":                item["study_id"],
		"crf_id":                  item["crf_id"],
		"name":                    item["name"],
		"live":                    item["live"],
		"randomization_enabled":   item["randomization_enabled"],
		"gcp_enabled":             item["gcp_enabled"],
		"surveys_enabled":         item["surveys_enabled"],
		"premium_support_enabled": item["premium_support_enabled"],
		"main_contact":            item["main_contact"],
		"institute_id":            item["institute_id"],
		"created_on":              item["created_on"],
		"updated_on":              item["updated_on"],
		"duration":                item["duration"],
	}
}

func castorUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"user_id":       item["user_id"],
		"email_address": item["email_address"],
		"first_name":    item["first_name"],
		"last_name":     item["last_name"],
		"full_name":     item["full_name"],
		"is_active":     item["is_active"],
		"institute":     item["institute"],
		"last_login":    item["last_login"],
	}
}

func castorCountryRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"country_id":   item["country_id"],
		"country_name": item["country_name"],
		"country_tld":  item["country_tld"],
		"country_cca2": item["country_cca2"],
		"country_cca3": item["country_cca3"],
	}
}

func castorAuditTrailRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uuid":          item["uuid"],
		"datetime":      item["datetime"],
		"event_type":    item["event_type"],
		"user_id":       item["user_id"],
		"user_name":     item["user_name"],
		"user_email":    item["user_email"],
		"event_details": item["event_details"],
	}
}
