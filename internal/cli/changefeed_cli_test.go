package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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

func TestInspectPostgresKeepsStaticPollingWatermarkPlannedWhileRuntimeBindsItPerStream(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "inspect", "postgres", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors inspect postgres --json) code = %d stderr = %s", code, stderr.String())
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("decode postgres inspect response: %v", err)
	}
	pollingRaw, ok := response["polling_watermark"]
	if !ok {
		t.Fatalf("postgres inspect omitted the polling-watermark status: %s", stdout.String())
	}
	var polling struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(pollingRaw, &polling); err != nil {
		t.Fatalf("decode postgres polling-watermark descriptor: %v", err)
	}
	if polling.Status != "planned" {
		t.Fatalf("postgres polling-watermark status = %q, want planned until a source/object/destination preflight can succeed", polling.Status)
	}
	if polling.Reason == "" {
		t.Fatalf("postgres planned polling-watermark status omitted its blocking reason: %s", pollingRaw)
	}
	if bytes.Contains(pollingRaw, []byte(`"implemented"`)) || bytes.Contains(pollingRaw, []byte("change_capture")) {
		t.Fatalf("postgres polling-watermark inspection fabricated executable CDC semantics: %s", pollingRaw)
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"connectors", "inspect", "postgres"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors inspect postgres) code = %d stderr = %s", code, stderr.String())
	}
	if bytes.Contains(stdout.Bytes(), []byte("No polling source ordering, checkpoint, snapshot, deletion, or rebootstrap behavior is implemented")) {
		t.Fatalf("postgres manual denied its dynamically-bound polling transport: %s", stdout.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("native_database/postgres_polling_watermark")) {
		t.Fatalf("postgres manual omitted the production polling executor: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"connectors", "catalog", "--capability", "cdc", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors catalog --capability cdc --json) code = %d stderr = %s", code, stderr.String())
	}
	var catalog struct {
		Connectors []struct {
			Name             string `json:"name"`
			PollingWatermark struct {
				Status string `json:"status"`
			} `json:"polling_watermark"`
		} `json:"connectors"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &catalog); err != nil {
		t.Fatalf("decode PostgreSQL CDC catalog: %v", err)
	}
	for _, connector := range catalog.Connectors {
		if connector.Name != "postgres" {
			continue
		}
		if connector.PollingWatermark.Status != "planned" {
			t.Fatalf("PostgreSQL catalog polling-watermark status = %q, want planned rather than an executable polling claim", connector.PollingWatermark.Status)
		}
		return
	}
	t.Fatalf("PostgreSQL missing from the CDC catalog: %s", stdout.String())
}

func TestPostgresNativeAPISurfaceHasNoFabricatedRESTEndpoints(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	apiSurfacePath := filepath.Join(filepath.Dir(thisFile), "..", "connectors", "defs", "postgres", "api_surface.json")
	raw, err := os.ReadFile(apiSurfacePath)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", apiSurfacePath, err)
	}
	var surface struct {
		API       string            `json:"api"`
		Endpoints []json.RawMessage `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("decode PostgreSQL api surface: %v", err)
	}
	if len(surface.Endpoints) != 0 {
		t.Fatalf("PostgreSQL native API surface fabricated %d REST endpoints: %s", len(surface.Endpoints), raw)
	}
	if surface.API == "" {
		t.Fatalf("PostgreSQL native API surface omitted protocol identity: %s", raw)
	}
}
