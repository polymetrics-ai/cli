package app

import (
	"os"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
)

func TestPayloadIdentitiesBindFileSizeAndMTimeWithoutPathDisclosure(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/payload.csv"
	if err := os.WriteFile(path, []byte("one"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	records := []connectors.Record{{"data_file_path": "payload.csv"}}
	before, err := payloadIdentitiesForRecords(dir, records)
	if err != nil {
		t.Fatalf("payloadIdentitiesForRecords: %v", err)
	}
	if len(before) != 1 {
		t.Fatalf("identities = %+v, want one", before)
	}
	if before[0].Field != "data_file_path" || before[0].SizeBytes != 3 || before[0].PathHash == "" {
		t.Fatalf("identity = %+v, want field, size, and path hash", before[0])
	}
	if before[0].PathHash == path || before[0].PathHash == "payload.csv" {
		t.Fatalf("identity leaked raw path: %+v", before[0])
	}

	// Force both content size and mtime to differ so approval hashes bind the
	// local payload identity without hashing/logging file contents.
	time.Sleep(time.Millisecond)
	if err := os.WriteFile(path, []byte("one,two"), 0o600); err != nil {
		t.Fatalf("WriteFile second: %v", err)
	}
	after, err := payloadIdentitiesForRecords(dir, records)
	if err != nil {
		t.Fatalf("payloadIdentitiesForRecords after: %v", err)
	}
	if before[0].SizeBytes == after[0].SizeBytes && before[0].ModTimeUnixNano == after[0].ModTimeUnixNano {
		t.Fatalf("identity did not change after file rewrite: before=%+v after=%+v", before[0], after[0])
	}

	beforeHash, err := connectorCommandPlanHash("plan", "gong", "cred", nil, "crm upload-entities", []string{"crm", "upload-entities"}, "upload_crm_entities", records[0], before)
	if err != nil {
		t.Fatalf("connectorCommandPlanHash before: %v", err)
	}
	afterHash, err := connectorCommandPlanHash("plan", "gong", "cred", nil, "crm upload-entities", []string{"crm", "upload-entities"}, "upload_crm_entities", records[0], after)
	if err != nil {
		t.Fatalf("connectorCommandPlanHash after: %v", err)
	}
	if beforeHash == afterHash {
		t.Fatalf("plan hash did not change after payload identity changed")
	}
}

func TestPayloadIdentitiesRejectPathTraversal(t *testing.T) {
	_, err := payloadIdentitiesForRecords(t.TempDir(), []connectors.Record{{"media_file_path": "../secret.mp4"}})
	if err == nil {
		t.Fatal("payloadIdentitiesForRecords: want traversal rejection")
	}
}
