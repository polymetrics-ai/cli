package googleclassroom

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Google Classroom REST shape it reads:
// the path template (relative to base_url), the JSON array key in the response,
// whether it is nested under a course, and the record mapper.
type streamEndpoint struct {
	// pathTemplate is the resource path. For course-nested streams it contains a
	// "%s" placeholder filled with the course id at read time.
	pathTemplate string
	// recordsKey is the JSON object key holding the records array (e.g. "courses").
	recordsKey string
	// nested is true when the endpoint must be fetched once per course.
	nested bool
	// mapRecord flattens a raw API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// classroomStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in classroomStreams; the read
// path is fully data-driven from this table.
var classroomStreamEndpoints = map[string]streamEndpoint{
	"courses":       {pathTemplate: "v1/courses", recordsKey: "courses", mapRecord: courseRecord},
	"teachers":      {pathTemplate: "v1/courses/%s/teachers", recordsKey: "teachers", nested: true, mapRecord: teacherRecord},
	"students":      {pathTemplate: "v1/courses/%s/students", recordsKey: "students", nested: true, mapRecord: studentRecord},
	"courseWork":    {pathTemplate: "v1/courses/%s/courseWork", recordsKey: "courseWork", nested: true, mapRecord: courseWorkRecord},
	"announcements": {pathTemplate: "v1/courses/%s/announcements", recordsKey: "announcements", nested: true, mapRecord: announcementRecord},
}

// classroomStreams returns the connector's published stream catalog.
func classroomStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "courses",
			Description:  "Google Classroom courses.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updateTime"},
			Fields:       courseFields(),
		},
		{
			Name:         "teachers",
			Description:  "Teachers enrolled in each course.",
			PrimaryKey:   []string{"courseId", "userId"},
			CursorFields: nil,
			Fields:       memberFields(),
		},
		{
			Name:         "students",
			Description:  "Students enrolled in each course.",
			PrimaryKey:   []string{"courseId", "userId"},
			CursorFields: nil,
			Fields:       memberFields(),
		},
		{
			Name:         "courseWork",
			Description:  "Assignments and questions for each course.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updateTime"},
			Fields:       courseWorkFields(),
		},
		{
			Name:         "announcements",
			Description:  "Announcements posted to each course.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updateTime"},
			Fields:       announcementFields(),
		},
	}
}

func courseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "section", Type: "string"},
		{Name: "descriptionHeading", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "room", Type: "string"},
		{Name: "ownerId", Type: "string"},
		{Name: "creationTime", Type: "timestamp"},
		{Name: "updateTime", Type: "timestamp"},
		{Name: "enrollmentCode", Type: "string"},
		{Name: "courseState", Type: "string"},
		{Name: "alternateLink", Type: "string"},
		{Name: "teacherGroupEmail", Type: "string"},
		{Name: "courseGroupEmail", Type: "string"},
		{Name: "guardiansEnabled", Type: "boolean"},
		{Name: "calendarId", Type: "string"},
	}
}

func memberFields() []connectors.Field {
	return []connectors.Field{
		{Name: "courseId", Type: "string"},
		{Name: "userId", Type: "string"},
		{Name: "fullName", Type: "string"},
		{Name: "emailAddress", Type: "string"},
		{Name: "photoUrl", Type: "string"},
	}
}

func courseWorkFields() []connectors.Field {
	return []connectors.Field{
		{Name: "courseId", Type: "string"},
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "alternateLink", Type: "string"},
		{Name: "creationTime", Type: "timestamp"},
		{Name: "updateTime", Type: "timestamp"},
		{Name: "dueDate", Type: "string"},
		{Name: "workType", Type: "string"},
		{Name: "maxPoints", Type: "number"},
	}
}

func announcementFields() []connectors.Field {
	return []connectors.Field{
		{Name: "courseId", Type: "string"},
		{Name: "id", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "alternateLink", Type: "string"},
		{Name: "creationTime", Type: "timestamp"},
		{Name: "updateTime", Type: "timestamp"},
		{Name: "creatorUserId", Type: "string"},
	}
}

func courseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                 item["id"],
		"name":               item["name"],
		"section":            item["section"],
		"descriptionHeading": item["descriptionHeading"],
		"description":        item["description"],
		"room":               item["room"],
		"ownerId":            item["ownerId"],
		"creationTime":       item["creationTime"],
		"updateTime":         item["updateTime"],
		"enrollmentCode":     item["enrollmentCode"],
		"courseState":        item["courseState"],
		"alternateLink":      item["alternateLink"],
		"teacherGroupEmail":  item["teacherGroupEmail"],
		"courseGroupEmail":   item["courseGroupEmail"],
		"guardiansEnabled":   item["guardiansEnabled"],
		"calendarId":         item["calendarId"],
	}
}

// teacherRecord and studentRecord flatten the shared {courseId, userId, profile}
// shape returned by the teachers/students list endpoints.
func teacherRecord(item map[string]any) connectors.Record { return memberRecord(item) }
func studentRecord(item map[string]any) connectors.Record { return memberRecord(item) }

func memberRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"courseId": item["courseId"],
		"userId":   item["userId"],
	}
	if profile, ok := item["profile"].(map[string]any); ok {
		if name, ok := profile["name"].(map[string]any); ok {
			rec["fullName"] = name["fullName"]
		}
		rec["emailAddress"] = profile["emailAddress"]
		rec["photoUrl"] = profile["photoUrl"]
		if rec["userId"] == nil {
			rec["userId"] = profile["id"]
		}
	}
	return rec
}

func courseWorkRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"courseId":      item["courseId"],
		"id":            item["id"],
		"title":         item["title"],
		"description":   item["description"],
		"state":         item["state"],
		"alternateLink": item["alternateLink"],
		"creationTime":  item["creationTime"],
		"updateTime":    item["updateTime"],
		"dueDate":       item["dueDate"],
		"workType":      item["workType"],
		"maxPoints":     item["maxPoints"],
	}
}

func announcementRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"courseId":      item["courseId"],
		"id":            item["id"],
		"text":          item["text"],
		"state":         item["state"],
		"alternateLink": item["alternateLink"],
		"creationTime":  item["creationTime"],
		"updateTime":    item["updateTime"],
		"creatorUserId": item["creatorUserId"],
	}
}
