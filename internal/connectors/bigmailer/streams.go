package bigmailer

import "polymetrics/internal/connectors"

// streamEndpoint describes how a BigMailer stream is read.
//
// BigMailer has two shapes of stream:
//   - top-level collections (brands, users) read directly from a fixed path;
//   - brand-substreams (contacts, lists, fields, ...) that live under
//     /brands/{brand_id}/<resource> and must be read once per brand.
//
// For substreams, brandSub is true and resource is the trailing path segment;
// the read path lists brands first, then paginates each brand's resource and
// stamps brand_id onto every record.
type streamEndpoint struct {
	// path is the full path for top-level streams (e.g. "brands").
	path string
	// resource is the trailing segment for brand-substreams (e.g. "contacts").
	resource string
	// brandSub marks a stream that must be read per brand.
	brandSub bool
	// mapRecord flattens a raw BigMailer object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// bigmailerStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in bigmailerStreams; the read
// path is fully data-driven from this table.
var bigmailerStreamEndpoints = map[string]streamEndpoint{
	"brands":   {path: "brands", mapRecord: brandRecord},
	"users":    {path: "users", mapRecord: userRecord},
	"contacts": {resource: "contacts", brandSub: true, mapRecord: contactRecord},
	"lists":    {resource: "lists", brandSub: true, mapRecord: listRecord},
	"fields":   {resource: "fields", brandSub: true, mapRecord: fieldRecord},
}

// bigmailerStreams returns the connector's published stream catalog. BigMailer
// objects expose a string id (primary key) and a unix `created` timestamp, but
// the API only supports full-refresh (no incremental cursor), so CursorFields is
// empty across the board.
func bigmailerStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "brands",
			Description: "BigMailer brands (sending identities) in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      brandFields(),
		},
		{
			Name:        "users",
			Description: "BigMailer account users.",
			PrimaryKey:  []string{"id"},
			Fields:      userFields(),
		},
		{
			Name:        "contacts",
			Description: "Contacts across every brand in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      contactFields(),
		},
		{
			Name:        "lists",
			Description: "Contact lists across every brand in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      listFields(),
		},
		{
			Name:        "fields",
			Description: "Custom contact fields across every brand in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      fieldFields(),
		},
	}
}

func brandFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "from_name", Type: "string"},
		{Name: "from_email", Type: "string"},
		{Name: "connection_id", Type: "string"},
		{Name: "num_contacts", Type: "integer"},
		{Name: "contact_limit", Type: "integer"},
		{Name: "created", Type: "integer"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "created", Type: "integer"},
	}
}

func contactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "brand_id", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "unsubscribe_all", Type: "boolean"},
		{Name: "num_soft_bounces", Type: "integer"},
		{Name: "num_hard_bounces", Type: "integer"},
		{Name: "num_complaints", Type: "integer"},
		{Name: "created", Type: "integer"},
	}
}

func listFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "brand_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "num_contacts", Type: "integer"},
		{Name: "created", Type: "integer"},
	}
}

func fieldFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "brand_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "tag", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "created", Type: "integer"},
	}
}

func brandRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"from_name":     item["from_name"],
		"from_email":    item["from_email"],
		"connection_id": item["connection_id"],
		"num_contacts":  item["num_contacts"],
		"contact_limit": item["contact_limit"],
		"created":       item["created"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"email":   item["email"],
		"name":    item["name"],
		"role":    item["role"],
		"created": item["created"],
	}
}

func contactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"brand_id":         item["brand_id"],
		"email":            item["email"],
		"unsubscribe_all":  item["unsubscribe_all"],
		"num_soft_bounces": item["num_soft_bounces"],
		"num_hard_bounces": item["num_hard_bounces"],
		"num_complaints":   item["num_complaints"],
		"created":          item["created"],
	}
}

func listRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"brand_id":     item["brand_id"],
		"name":         item["name"],
		"num_contacts": item["num_contacts"],
		"created":      item["created"],
	}
}

func fieldRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"brand_id": item["brand_id"],
		"name":     item["name"],
		"tag":      item["tag"],
		"type":     item["type"],
		"created":  item["created"],
	}
}
