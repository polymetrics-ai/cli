package eventee

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Eventee API resource path (relative to
// base_url) it reads from, the dotted JSON path the records array lives under in
// the response, and the record mapper that flattens its objects.
//
// Most Eventee streams share the /content endpoint and select a nested array by
// the stream's own name (the response is one object holding lectures[],
// speakers[], days[], etc.). partners and participants have dedicated endpoints
// returning a top-level array (recordsPath "").
type streamEndpoint struct {
	// resource is the Eventee endpoint path segment (e.g. "content", "partners").
	resource string
	// recordsPath is the dotted path to the records array in the response body.
	// Empty means the body itself is the array.
	recordsPath string
	// mapRecord flattens a raw Eventee object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// eventeeStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in eventeeStreams; the read path
// is fully data-driven from this table.
var eventeeStreamEndpoints = map[string]streamEndpoint{
	"lectures":  {resource: "content", recordsPath: "lectures", mapRecord: eventeeLectureRecord},
	"speakers":  {resource: "content", recordsPath: "speakers", mapRecord: eventeeSpeakerRecord},
	"days":      {resource: "content", recordsPath: "days", mapRecord: eventeeDayRecord},
	"halls":     {resource: "content", recordsPath: "halls", mapRecord: eventeeHallRecord},
	"tracks":    {resource: "content", recordsPath: "tracks", mapRecord: eventeeTrackRecord},
	"workshops": {resource: "content", recordsPath: "workshops", mapRecord: eventeeLectureRecord},
	"partners":  {resource: "partners", recordsPath: "", mapRecord: eventeePartnerRecord},
}

// eventeeStreams returns the connector's published stream catalog. Eventee
// objects expose a numeric id (partners/participants use email as the key in the
// upstream schema, but the core set here keys on id) and string created_at /
// updated_at timestamps; the API offers only full-refresh syncs (no incremental
// cursor), so CursorFields is empty across the board.
func eventeeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "lectures",
			Description: "Eventee lectures (sessions) in the event agenda.",
			PrimaryKey:  []string{"id"},
			Fields:      eventeeLectureFields(),
		},
		{
			Name:        "speakers",
			Description: "Eventee speakers.",
			PrimaryKey:  []string{"id"},
			Fields:      eventeeSpeakerFields(),
		},
		{
			Name:        "days",
			Description: "Eventee event days.",
			PrimaryKey:  []string{"id"},
			Fields:      eventeeDayFields(),
		},
		{
			Name:        "halls",
			Description: "Eventee halls (rooms / stages).",
			PrimaryKey:  []string{"id"},
			Fields:      eventeeHallFields(),
		},
		{
			Name:        "tracks",
			Description: "Eventee tracks used to group sessions.",
			PrimaryKey:  []string{"id"},
			Fields:      eventeeTrackFields(),
		},
		{
			Name:        "workshops",
			Description: "Eventee workshops (bookable sessions).",
			PrimaryKey:  []string{"id"},
			Fields:      eventeeLectureFields(),
		},
		{
			Name:        "partners",
			Description: "Eventee partners (sponsors / exhibitors).",
			PrimaryKey:  []string{"id"},
			Fields:      eventeePartnerFields(),
		},
	}
}

func eventeeLectureFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "event_id", Type: "integer"},
		{Name: "event_day_id", Type: "integer"},
		{Name: "hall_id", Type: "integer"},
		{Name: "start", Type: "string"},
		{Name: "end", Type: "string"},
		{Name: "capacity", Type: "integer"},
		{Name: "available", Type: "boolean"},
		{Name: "booked", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func eventeeSpeakerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "position", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "country", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "bio", Type: "string"},
		{Name: "web", Type: "string"},
		{Name: "event_id", Type: "integer"},
		{Name: "order", Type: "integer"},
	}
}

func eventeeDayFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "date", Type: "string"},
		{Name: "event_id", Type: "integer"},
		{Name: "content_url", Type: "string"},
	}
}

func eventeeHallFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "event_id", Type: "integer"},
		{Name: "order", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func eventeeTrackFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "order", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func eventeePartnerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "company", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "address", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "web", Type: "string"},
		{Name: "sponsor", Type: "boolean"},
		{Name: "exhibitor", Type: "boolean"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
	}
}

func eventeeLectureRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"type":         item["type"],
		"code":         item["code"],
		"event_id":     item["event_id"],
		"event_day_id": item["event_day_id"],
		"hall_id":      item["hall_id"],
		"start":        item["start"],
		"end":          item["end"],
		"capacity":     item["capacity"],
		"available":    item["available"],
		"booked":       item["booked"],
		"created_at":   item["created_at"],
		"updated_at":   item["updated_at"],
	}
}

func eventeeSpeakerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":       item["id"],
		"name":     item["name"],
		"company":  item["company"],
		"position": item["position"],
		"email":    item["email"],
		"phone":    item["phone"],
		"country":  item["country"],
		"language": item["language"],
		"bio":      item["bio"],
		"web":      item["web"],
		"event_id": item["event_id"],
		"order":    item["order"],
	}
}

func eventeeDayRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"date":        item["date"],
		"event_id":    item["event_id"],
		"content_url": item["content_url"],
	}
}

func eventeeHallRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"event_id":   item["event_id"],
		"order":      item["order"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func eventeeTrackRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"color":      item["color"],
		"order":      item["order"],
		"created_at": item["created_at"],
		"updated_at": item["updated_at"],
	}
}

func eventeePartnerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"company":     item["company"],
		"description": item["description"],
		"address":     item["address"],
		"code":        item["code"],
		"email":       item["email"],
		"phone":       item["phone"],
		"web":         item["web"],
		"sponsor":     item["sponsor"],
		"exhibitor":   item["exhibitor"],
		"created_at":  item["created_at"],
		"updated_at":  item["updated_at"],
	}
}
