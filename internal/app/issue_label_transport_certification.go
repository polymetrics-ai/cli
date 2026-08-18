package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// DeclaredTransportCertificationProof is the content-free execution evidence
// returned to the connector-neutral certification stage. It contains no
// credential, provider payload, URL, or connector-specific configuration.
type DeclaredTransportCertificationProof struct {
	Declared             bool
	Applicable           bool
	SkipReason           string
	SourceReference      string
	DestinationReference string
	ProviderReads        int
	ProviderWrites       int
	RecordsRead          int
	RecordsLoaded        int
	CheckpointCommitted  bool
	WarehouseManifests   int
	WarehouseParquet     int
	Modes                []DeclaredTransportModeProof
}

// DeclaredTransportModeProof is the redaction-safe outcome for one declared
// database transport mode. The target address is opaque, definition-derived
// identity; it contains neither a credential nor a rendered connection value.
type DeclaredTransportModeProof struct {
	Mode                string
	ApplyStrategy       string
	RecordsRead         int
	RecordsLoaded       int
	CheckpointCommitted bool
	TargetNamespace     string
	TargetRelation      string
}

// ProbeDeclaredTransportForCertification executes the existing closed
// issue-label transport adapter through production Open/RunETL composition.
// It returns Applicable=false for every other connector. The HTTP server is a
// bounded provider fixture; all staging, reopening, approval, dispatch,
// readback, checkpointing, and persistence are production paths.
func ProbeDeclaredTransportForCertification(ctx context.Context, certificationRoot, connectorName string) (proof DeclaredTransportCertificationProof, resultErr error) {
	root, err := os.MkdirTemp(certificationRoot, "declared-transport-pair-")
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("create declared transport proof root: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(root); err != nil && resultErr == nil {
			resultErr = fmt.Errorf("clean declared transport proof root: %w", err)
		}
	}()
	if err := InitProject(root); err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("initialize declared transport proof project: %w", err)
	}
	application, err := Open(root)
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("open declared transport proof project: %w", err)
	}
	registered, ok := application.registry.Get(connectorName)
	if !ok {
		return DeclaredTransportCertificationProof{}, nil
	}
	applicable, err := declaredTransportCertificationApplicability(registered)
	if err != nil {
		return DeclaredTransportCertificationProof{}, err
	}
	if !applicable {
		return DeclaredTransportCertificationProof{}, nil
	}
	descriptor, ok := connectors.SyncTransportDescriptorOf(registered)
	if !ok || descriptor.Source == nil || descriptor.Destination == nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("certification transport adapter lost its source or destination declaration")
	}
	proof.Declared = true

	var mu sync.Mutex
	reads := 0
	writes := 0
	targetLabels := []string{"legacy"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			if request.URL.Path != "/repos/acme/widgets/issues" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			mu.Lock()
			reads++
			labels := append([]string(nil), targetLabels...)
			mu.Unlock()
			page := request.URL.Query().Get("page")
			if page != "" && page != "1" {
				_ = json.NewEncoder(w).Encode([]map[string]any{})
				return
			}
			_ = json.NewEncoder(w).Encode(certificationIssueLabelPage(labels))
		case http.MethodPost:
			if request.URL.Path != "/repos/acme/widgets/issues/200/labels" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			var body struct {
				Labels []string `json:"labels"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			mu.Lock()
			writes++
			for _, label := range body.Labels {
				if !containsCertificationLabel(targetLabels, label) {
					targetLabels = append(targetLabels, label)
				}
			}
			response := certificationLabelResponse(targetLabels)
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(response)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	if _, err := application.AddCredential(ctx, AddCredentialRequest{
		Name:      "transport-certification",
		Connector: connectorName,
		Config: map[string]string{
			"owner": "acme", "repo": "widgets", "public_access": "true", "base_url": server.URL,
		},
	}); err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("add declared transport proof credential: %w", err)
	}
	connection, err := application.CreateConnection(ctx, CreateConnectionRequest{
		Name: "declared_transport_certification",
		Source: EndpointConfig{Connector: connectorName, Credential: "transport-certification", Config: map[string]string{
			issueLabelTransportSourceIssueConfig: "100",
		}},
		Destination: EndpointConfig{Connector: connectorName, Credential: "transport-certification", Config: map[string]string{
			issueLabelTransportTargetIssueConfig: "200", issueLabelTransportLabelConfig: "transport-demo",
		}},
		Streams: map[string]StreamConfig{
			"issues": {SyncMode: string(synccontract.ModeFullAppend), DestinationTable: "issues"},
		},
	})
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("create declared transport proof connection: %w", err)
	}
	plan, err := application.PlanIssueLabelTransport(ctx, connection.ID)
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("plan declared transport proof: %w", err)
	}
	plan, preview, err := application.PreviewIssueLabelTransport(ctx, plan.ID)
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("preview declared transport proof: %w", err)
	}
	if preview.Digest == "" || plan.ApprovalToken == "" {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("declared transport proof did not produce an approval-bound preview")
	}
	run, err := application.RunETL(ctx, RunETLRequest{
		Connection: connection.Name, Stream: "issues", BatchSize: 1,
		DestinationApproval: synctransport.DestinationApproval{
			PlanID: plan.ID, ApprovalToken: plan.ApprovalToken,
			Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
		},
	})
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("execute declared transport pair: %w", err)
	}
	if run.Status != "completed" {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("declared transport run status = %q, want completed", run.Status)
	}
	reopened, err := Open(root)
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("reopen declared transport proof project: %w", err)
	}
	storedRun, err := reopened.GetRun(run.ID)
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("read declared transport run after reopen: %w", err)
	}
	manifests, parquet, err := countCertificationTransportStageArtifacts(root)
	if err != nil {
		return DeclaredTransportCertificationProof{}, fmt.Errorf("inspect declared transport warehouse after reopen: %w", err)
	}
	mu.Lock()
	proof = DeclaredTransportCertificationProof{
		Applicable: true, SourceReference: descriptor.Source.Executor.ID, DestinationReference: descriptor.Destination.Executor.ID,
		ProviderReads: reads, ProviderWrites: writes, RecordsRead: storedRun.RecordsRead, RecordsLoaded: storedRun.RecordsLoaded,
		CheckpointCommitted: len(storedRun.Checkpoint) > 0, WarehouseManifests: manifests, WarehouseParquet: parquet,
	}
	mu.Unlock()
	return proof, nil
}

func declaredTransportCertificationApplicability(connector connectors.Connector) (bool, error) {
	definition, ok := connectors.DefinitionOf(connector)
	if !ok || !definitionDeclaresIssueLabelTransport(definition) {
		return false, nil
	}
	if !isIssueLabelTransportConnector(connector) {
		return false, fmt.Errorf("declared issue-label transport definition %q is absent or incomplete", connector.Name())
	}
	return true, nil
}

func countCertificationTransportStageArtifacts(root string) (int, int, error) {
	manifests, parquet := 0, 0
	err := filepath.Walk(filepath.Join(root, ".polymetrics", "warehouse"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) == "transport" && strings.HasSuffix(info.Name(), ".json") {
			manifests++
		}
		if strings.HasPrefix(info.Name(), "transport-") && strings.HasSuffix(info.Name(), ".parquet") {
			parquet++
		}
		return nil
	})
	return manifests, parquet, err
}

func certificationIssueLabelPage(targetLabels []string) []map[string]any {
	records := []map[string]any{certificationIssueLabelRecord(100, nil), certificationIssueLabelRecord(200, targetLabels)}
	for number := 1; len(records) < 100; number++ {
		if number != 100 && number != 200 {
			records = append(records, certificationIssueLabelRecord(number, nil))
		}
	}
	return records
}

func certificationIssueLabelRecord(number int, labels []string) map[string]any {
	return map[string]any{
		"id": number, "node_id": "I_" + strconv.Itoa(number), "number": number,
		"title": "transport certification issue", "state": "open", "labels": certificationLabelResponse(labels),
	}
}

func certificationLabelResponse(labels []string) []map[string]any {
	response := make([]map[string]any, 0, len(labels))
	for _, label := range labels {
		response = append(response, map[string]any{"name": label})
	}
	return response
}

func containsCertificationLabel(labels []string, want string) bool {
	for _, label := range labels {
		if label == want {
			return true
		}
	}
	return false
}
