package gmail

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Gmail API resource path (relative to
// base_url, with a %s userId slot) it reads from, the JSON path to the record
// array in the list response, whether the endpoint paginates, and the record
// mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the list endpoint path template; %s is filled with the userId
	// (e.g. "users/%s/messages").
	resource string
	// recordsPath is the dotted JSON path to the records array in the response.
	recordsPath string
	// paginated is true when the endpoint returns nextPageToken cursor pages.
	paginated bool
	// mapRecord flattens a raw Gmail object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in gmailStreams; the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"messages": {resource: "users/%s/messages", recordsPath: "messages", paginated: true, mapRecord: messageRecord},
	"threads":  {resource: "users/%s/threads", recordsPath: "threads", paginated: true, mapRecord: threadRecord},
	"drafts":   {resource: "users/%s/drafts", recordsPath: "drafts", paginated: true, mapRecord: draftRecord},
	"labels":   {resource: "users/%s/labels", recordsPath: "labels", paginated: false, mapRecord: labelRecord},
}

// gmailStreams returns the connector's published stream catalog. The Gmail list
// endpoints return lightweight id-bearing resources; the primary key is ["id"]
// across the board. The upstream upstream source only supports full_refresh, so
// no incremental cursor field is published.
func gmailStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "messages",
			Description: "Gmail messages (id and threadId from the messages.list endpoint).",
			PrimaryKey:  []string{"id"},
			Fields:      messageFields(),
		},
		{
			Name:        "threads",
			Description: "Gmail threads (id, snippet, historyId from the threads.list endpoint).",
			PrimaryKey:  []string{"id"},
			Fields:      threadFields(),
		},
		{
			Name:        "drafts",
			Description: "Gmail drafts (id and the associated message id/threadId from the drafts.list endpoint).",
			PrimaryKey:  []string{"id"},
			Fields:      draftFields(),
		},
		{
			Name:        "labels",
			Description: "Gmail labels (system and user labels from the labels.list endpoint).",
			PrimaryKey:  []string{"id"},
			Fields:      labelFields(),
		},
	}
}

func messageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "thread_id", Type: "string"},
	}
}

func threadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "snippet", Type: "string"},
		{Name: "history_id", Type: "string"},
	}
}

func draftFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "message_id", Type: "string"},
		{Name: "thread_id", Type: "string"},
	}
}

func labelFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "message_list_visibility", Type: "string"},
		{Name: "label_list_visibility", Type: "string"},
		{Name: "messages_total", Type: "integer"},
		{Name: "messages_unread", Type: "integer"},
		{Name: "threads_total", Type: "integer"},
		{Name: "threads_unread", Type: "integer"},
	}
}

func messageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"thread_id": item["threadId"],
	}
}

func threadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"snippet":    item["snippet"],
		"history_id": item["historyId"],
	}
}

func draftRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":         item["id"],
		"message_id": nil,
		"thread_id":  nil,
	}
	if msg, ok := item["message"].(map[string]any); ok {
		rec["message_id"] = msg["id"]
		rec["thread_id"] = msg["threadId"]
	}
	return rec
}

func labelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"name":                    item["name"],
		"type":                    item["type"],
		"message_list_visibility": item["messageListVisibility"],
		"label_list_visibility":   item["labelListVisibility"],
		"messages_total":          item["messagesTotal"],
		"messages_unread":         item["messagesUnread"],
		"threads_total":           item["threadsTotal"],
		"threads_unread":          item["threadsUnread"],
	}
}
