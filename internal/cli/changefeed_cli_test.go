package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
	nativepostgres "polymetrics.ai/internal/connectors/native/postgres"
)

func TestCatalogedCDCRejectsPostgresUntilStreamedStagingExists(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "postgres")
	if err != nil {
		t.Fatalf("load postgres bundle: %v", err)
	}
	if bundle.Metadata.Capabilities.CDC {
		t.Fatal("postgres metadata must not claim CDC before streamed staging is implemented")
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
	stubErr := reader.ReadCDC(context.Background(), connectors.CDCReadRequest{}, func(connectors.CDCEvent) error {
		return nil
	})
	if !errors.Is(stubErr, connectors.ErrUnsupportedOperation) {
		t.Fatalf("postgres ReadCDC = %v, want fail-closed unsupported result", stubErr)
	}

	if postgresListed {
		t.Fatalf("CDC catalog advertised postgres before streamed staging is implemented: %v", stubErr)
	}
}

func TestInspectPostgresReportsPlannedChangefeed(t *testing.T) {
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
		Reason    string `json:"reason"`
		Source    struct {
			ArtifactURL string `json:"artifact_url"`
			Version     string `json:"artifact_version"`
			RetrievedAt string `json:"retrieved_at"`
		} `json:"source"`
	}
	if err := json.Unmarshal(changefeedRaw, &changefeed); err != nil {
		t.Fatalf("decode postgres changefeed descriptor: %v", err)
	}
	if changefeed.Status != "planned" {
		t.Fatalf("postgres changefeed status = %q, want planned", changefeed.Status)
	}
	if changefeed.Mechanism != "logical_replication" {
		t.Fatalf("postgres changefeed mechanism = %q, want logical_replication", changefeed.Mechanism)
	}
	if changefeed.Reason == "" {
		t.Fatal("planned postgres changefeed must explain why it is not executable")
	}
	if changefeed.Source.ArtifactURL == "" || changefeed.Source.Version == "" || changefeed.Source.RetrievedAt == "" {
		t.Fatalf("implemented postgres changefeed must retain source evidence, got %+v", changefeed.Source)
	}
}
