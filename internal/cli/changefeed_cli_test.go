package cli_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	nativepostgres "polymetrics.ai/internal/connectors/native/postgres"
)

func TestCatalogedCDCAdvertisesProvenPostgresChangefeed(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "postgres")
	if err != nil {
		t.Fatalf("load postgres bundle: %v", err)
	}
	if !bundle.Metadata.Capabilities.CDC {
		t.Fatal("postgres metadata must claim the proven pgoutput v2 CDC capability")
	}

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "catalog", "--capability", "cdc", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors catalog --capability cdc --json) code = %d stderr = %s", code, stderr.String())
	}

	var catalog struct {
		Connectors []struct {
			Name string `json:"name"`
		} `json:"connectors"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &catalog); err != nil {
		t.Fatalf("decode CDC catalog: %v", err)
	}

	postgresListed := false
	for _, connector := range catalog.Connectors {
		if connector.Name == "postgres" {
			postgresListed = true
			break
		}
	}

	reader, ok := any(nativepostgres.New()).(connectors.ChangefeedExecutor)
	if !ok {
		t.Fatal("postgres must expose a matching logical-replication ChangefeedExecutor")
	}
	if descriptor := reader.ChangefeedExecutorDescriptor(); descriptor.Status != connectors.ChangefeedStatusImplemented {
		t.Fatalf("postgres runtime changefeed descriptor = %#v, want implemented", descriptor)
	}
	if !postgresListed {
		t.Fatal("CDC catalog omitted PostgreSQL despite the matching executable changefeed")
	}
}

func TestInspectPostgresReportsImplementedChangefeed(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "inspect", "postgres", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors inspect postgres --json) code = %d stderr = %s", code, stderr.String())
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("decode postgres inspect response: %v", err)
	}
	changefeedRaw, ok := response["changefeed"]
	if !ok {
		t.Fatalf("postgres inspect omitted the changefeed descriptor: %s", stdout.String())
	}

	var changefeed struct {
		Status    string `json:"status"`
		Mechanism string `json:"mechanism"`
		Executor  struct {
			Kind string `json:"kind"`
			ID   string `json:"id"`
		} `json:"executor"`
		Checkpoint struct {
			Kind        string   `json:"kind"`
			Keys        []string `json:"keys"`
			CommitAfter string   `json:"commit_after"`
			OnInvalid   string   `json:"on_invalid"`
		} `json:"checkpoint"`
		Source struct {
			ArtifactURL string `json:"artifact_url"`
			Version     string `json:"artifact_version"`
			RetrievedAt string `json:"retrieved_at"`
		} `json:"source"`
	}
	if err := json.Unmarshal(changefeedRaw, &changefeed); err != nil {
		t.Fatalf("decode postgres changefeed descriptor: %v", err)
	}
	if changefeed.Status != "implemented" {
		t.Fatalf("postgres changefeed status = %q, want implemented", changefeed.Status)
	}
	if changefeed.Mechanism != "logical_replication" {
		t.Fatalf("postgres changefeed mechanism = %q, want logical_replication", changefeed.Mechanism)
	}
	if changefeed.Executor.Kind != "native" || changefeed.Executor.ID != "postgres_logical_replication" {
		t.Fatalf("postgres changefeed executor = %#v, want native/postgres_logical_replication", changefeed.Executor)
	}
	if changefeed.Checkpoint.Kind != "lsn" || len(changefeed.Checkpoint.Keys) != 1 || changefeed.Checkpoint.Keys[0] != "lsn" || changefeed.Checkpoint.CommitAfter != "downstream_ack" || changefeed.Checkpoint.OnInvalid != "resnapshot_required" {
		t.Fatalf("postgres changefeed checkpoint = %#v, want executable LSN receipt contract", changefeed.Checkpoint)
	}
	if changefeed.Source.ArtifactURL == "" || changefeed.Source.Version == "" || changefeed.Source.RetrievedAt == "" {
		t.Fatalf("implemented postgres changefeed must retain source evidence, got %+v", changefeed.Source)
	}
}
