package engine

import (
	"testing"

	"polymetrics.ai/internal/connectors/defs"
)

func TestTwentyBundleReadStreams(t *testing.T) {
	b, err := Load(defs.FS, "twenty")
	if err != nil {
		t.Fatalf("Load twenty: %v", err)
	}

	expected := []struct {
		name          string
		objectsPlural string
	}{
		{name: "attachments", objectsPlural: "attachments"},
		{name: "blocklists", objectsPlural: "blocklists"},
		{name: "calendar_channel_event_associations", objectsPlural: "calendarChannelEventAssociations"},
		{name: "calendar_event_participants", objectsPlural: "calendarEventParticipants"},
		{name: "calendar_events", objectsPlural: "calendarEvents"},
		{name: "call_recordings", objectsPlural: "callRecordings"},
		{name: "companies", objectsPlural: "companies"},
		{name: "dashboards", objectsPlural: "dashboards"},
		{name: "message_campaigns", objectsPlural: "messageCampaigns"},
		{name: "message_channel_message_association_message_folders", objectsPlural: "messageChannelMessageAssociationMessageFolders"},
		{name: "message_channel_message_associations", objectsPlural: "messageChannelMessageAssociations"},
		{name: "message_list_members", objectsPlural: "messageListMembers"},
		{name: "message_lists", objectsPlural: "messageLists"},
		{name: "message_participants", objectsPlural: "messageParticipants"},
		{name: "message_threads", objectsPlural: "messageThreads"},
		{name: "messages", objectsPlural: "messages"},
		{name: "note_targets", objectsPlural: "noteTargets"},
		{name: "notes", objectsPlural: "notes"},
		{name: "opportunities", objectsPlural: "opportunities"},
		{name: "people", objectsPlural: "people"},
		{name: "task_targets", objectsPlural: "taskTargets"},
		{name: "tasks", objectsPlural: "tasks"},
		{name: "timeline_activities", objectsPlural: "timelineActivities"},
		{name: "workflow_automated_triggers", objectsPlural: "workflowAutomatedTriggers"},
		{name: "workflow_runs", objectsPlural: "workflowRuns"},
		{name: "workflow_versions", objectsPlural: "workflowVersions"},
		{name: "workflows", objectsPlural: "workflows"},
		{name: "workspace_members", objectsPlural: "workspaceMembers"},
	}

	if len(b.Streams) != len(expected) {
		t.Fatalf("twenty stream count = %d, want %d", len(b.Streams), len(expected))
	}

	streams := make(map[string]StreamSpec, len(b.Streams))
	for _, stream := range b.Streams {
		streams[stream.Name] = stream
	}

	for _, tt := range expected {
		t.Run(tt.name, func(t *testing.T) {
			stream, ok := streams[tt.name]
			if !ok {
				t.Fatalf("stream %q missing", tt.name)
			}

			wantPath := "/rest/" + tt.objectsPlural
			if stream.Path != wantPath {
				t.Fatalf("Path = %q, want %q", stream.Path, wantPath)
			}
			wantRecordsPath := "data." + tt.objectsPlural
			if stream.Records.Path != wantRecordsPath {
				t.Fatalf("Records.Path = %q, want %q", stream.Records.Path, wantRecordsPath)
			}
			wantSchema := "schemas/" + tt.objectsPlural + ".json"
			if stream.SchemaRef != wantSchema {
				t.Fatalf("SchemaRef = %q, want %q", stream.SchemaRef, wantSchema)
			}
			if stream.Projection != "" && stream.Projection != "schema" {
				t.Fatalf("Projection = %q, want default schema projection", stream.Projection)
			}

			pagination := stream.Pagination
			if pagination == nil {
				t.Fatalf("Pagination = nil, want cursor pagination")
			}
			if pagination.Type != "cursor" || pagination.CursorParam != "starting_after" || pagination.LimitParam != "limit" ||
				pagination.TokenPath != "pageInfo.endCursor" || pagination.StopPath != "pageInfo.hasNextPage" {
				t.Fatalf("Pagination = %+v, want cursor starting_after/limit pageInfo.endCursor pageInfo.hasNextPage", *pagination)
			}

			schema := b.Schemas[tt.name]
			if schema == nil || schema.Schema == nil {
				t.Fatalf("schema for %q did not resolve", tt.name)
			}
		})
	}
}
