package mixmax

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Mixmax API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Mixmax list endpoint path segment (e.g. "codesnippets").
	resource string
	// mapRecord projects a raw Mixmax object onto a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// mixmaxStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in mixmaxStreams; the read path
// is fully data-driven from this table. Every Mixmax list endpoint returns
// {results:[...], next:<token>, hasNext:<bool>} and keys objects on "_id".
var mixmaxStreamEndpoints = map[string]streamEndpoint{
	"codesnippets": {resource: "codesnippets", mapRecord: mixmaxCodeSnippetRecord},
	"messages":     {resource: "messages", mapRecord: mixmaxMessageRecord},
	"rules":        {resource: "rules", mapRecord: mixmaxRuleRecord},
	"sequences":    {resource: "sequences", mapRecord: mixmaxSequenceRecord},
	"meetingtypes": {resource: "meetingtypes", mapRecord: mixmaxMeetingTypeRecord},
}

// mixmaxStreams returns the connector's published stream catalog. Every Mixmax
// object exposes a string "_id" primary key. Most streams carry a "createdAt"
// timestamp used as the incremental cursor; messages use "updatedAt".
func mixmaxStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "codesnippets",
			Description:  "Mixmax code snippets.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"createdAt"},
			Fields:       mixmaxCodeSnippetFields(),
		},
		{
			Name:         "messages",
			Description:  "Mixmax tracked messages.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"updatedAt"},
			Fields:       mixmaxMessageFields(),
		},
		{
			Name:         "rules",
			Description:  "Mixmax automation rules.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"createdAt"},
			Fields:       mixmaxRuleFields(),
		},
		{
			Name:         "sequences",
			Description:  "Mixmax sequences.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"createdAt"},
			Fields:       mixmaxSequenceFields(),
		},
		{
			Name:         "meetingtypes",
			Description:  "Mixmax meeting types.",
			PrimaryKey:   []string{"_id"},
			CursorFields: []string{"createdAt"},
			Fields:       mixmaxMeetingTypeFields(),
		},
	}
}

func mixmaxCodeSnippetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "language", Type: "string"},
		{Name: "theme", Type: "string"},
		{Name: "html", Type: "string"},
		{Name: "background", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "createdAt", Type: "string"},
	}
}

func mixmaxMessageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "subject", Type: "string"},
		{Name: "from", Type: "string"},
		{Name: "to", Type: "string"},
		{Name: "cc", Type: "string"},
		{Name: "bcc", Type: "string"},
		{Name: "sequence", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "trackingEnabled", Type: "boolean"},
		{Name: "linkTrackingEnabled", Type: "boolean"},
		{Name: "fileTrackingEnabled", Type: "boolean"},
		{Name: "created", Type: "string"},
		{Name: "sent", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func mixmaxRuleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "trigger", Type: "string"},
		{Name: "isPaused", Type: "boolean"},
		{Name: "userId", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "modifiedAt", Type: "string"},
	}
}

func mixmaxSequenceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "linkTrackingEnabled", Type: "boolean"},
		{Name: "fileTrackingEnabled", Type: "boolean"},
		{Name: "notificationsEnabled", Type: "boolean"},
		{Name: "syncToOrg", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
	}
}

func mixmaxMeetingTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "link", Type: "string"},
		{Name: "durationMin", Type: "integer"},
		{Name: "userId", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func mixmaxCodeSnippetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":        item["_id"],
		"title":      item["title"],
		"language":   item["language"],
		"theme":      item["theme"],
		"html":       item["html"],
		"background": item["background"],
		"userId":     item["userId"],
		"createdAt":  item["createdAt"],
	}
}

func mixmaxMessageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":                 item["_id"],
		"subject":             item["subject"],
		"from":                item["from"],
		"to":                  item["to"],
		"cc":                  item["cc"],
		"bcc":                 item["bcc"],
		"sequence":            item["sequence"],
		"userId":              item["userId"],
		"trackingEnabled":     item["trackingEnabled"],
		"linkTrackingEnabled": item["linkTrackingEnabled"],
		"fileTrackingEnabled": item["fileTrackingEnabled"],
		"created":             item["created"],
		"sent":                item["sent"],
		"updatedAt":           item["updatedAt"],
	}
}

func mixmaxRuleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":        item["_id"],
		"name":       item["name"],
		"trigger":    item["trigger"],
		"isPaused":   item["isPaused"],
		"userId":     item["userId"],
		"createdAt":  item["createdAt"],
		"modifiedAt": item["modifiedAt"],
	}
}

func mixmaxSequenceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":                  item["_id"],
		"name":                 item["name"],
		"timezone":             item["timezone"],
		"userId":               item["userId"],
		"linkTrackingEnabled":  item["linkTrackingEnabled"],
		"fileTrackingEnabled":  item["fileTrackingEnabled"],
		"notificationsEnabled": item["notificationsEnabled"],
		"syncToOrg":            item["syncToOrg"],
		"createdAt":            item["createdAt"],
	}
}

func mixmaxMeetingTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"_id":         item["_id"],
		"name":        item["name"],
		"type":        item["type"],
		"link":        item["link"],
		"durationMin": item["durationMin"],
		"userId":      item["userId"],
		"createdAt":   item["createdAt"],
		"updatedAt":   item["updatedAt"],
	}
}
