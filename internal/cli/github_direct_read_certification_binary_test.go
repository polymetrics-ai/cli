package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// TestPMBinaryExecutesGitHubDirectReadCandidatesAgainstFixture is a local
// behavioral proof for the declaration-generated GitHub read surface. It
// drives a freshly built pm binary through the real certification runner,
// command parser, direct-read engine, and GitHub HTTP transport. The fixture
// deliberately returns a concrete JSON object for every declared candidate so
// the runner's output assertions execute rather than merely inspecting the
// candidate ledger.
func TestPMBinaryExecutesGitHubDirectReadCandidatesAgainstFixture(t *testing.T) {
	const token = "github-direct-read-fixture-token"
	const tokenEnv = "PM_GITHUB_DIRECT_READ_FIXTURE_TOKEN"
	t.Setenv(tokenEnv, token)

	var observed struct {
		sync.Mutex
		paths []string
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(w, "direct-read fixture received a non-GET request", http.StatusMethodNotAllowed)
			return
		}
		if strings.Contains(request.URL.EscapedPath(), "{") || strings.Contains(request.URL.EscapedPath(), "}") {
			http.Error(w, "direct-read fixture received an unresolved path parameter", http.StatusBadRequest)
			return
		}
		if request.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "direct-read fixture received no declared bearer authentication", http.StatusUnauthorized)
			return
		}
		observed.Lock()
		observed.paths = append(observed.paths, request.URL.EscapedPath())
		observed.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"fixture":"github-direct-read","request_path":"`+request.URL.EscapedPath()+`"}`)
	}))
	t.Cleanup(server.Close)

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	generatedStages := githubGeneratedDirectReadStageSelection(t)
	output, err := runTransportPM(binary, "",
		"connectors", "certify", "github",
		"--direct-read-only",
		"--config", "base_url="+server.URL,
		"--config", "auth_type=token",
		"--config", "rate_limit_account=fixture-account",
		"--config", "owner=Polymetrics-Cert",
		"--config", "repo=pm-cert-3993-20260810-wz0fru",
		"--config", "certification_stages="+generatedStages,
		"--from-env", "token="+tokenEnv,
		"--root", root,
		"--json",
	)

	var envelope struct {
		Kind   string         `json:"kind"`
		Report certify.Report `json:"report"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode GitHub direct-read fixture certification output: %v\n%s", err, redactTransportFailureOutput(output, token))
	}
	if envelope.Kind != "ConnectorCertification" {
		t.Fatalf("GitHub direct-read fixture certification kind = %q, want ConnectorCertification", envelope.Kind)
	}
	if envelope.Report.Capabilities.DirectRead == nil || envelope.Report.Capabilities.DirectRead.Result != "pass" || envelope.Report.Capabilities.DirectRead.StagesChecked != 97 {
		t.Fatalf("GitHub direct-read fixture capability = %+v (command error=%v), want 97 passing generated candidate executions", envelope.Report.Capabilities.DirectRead, err)
	}
	generatedStageCount := 0
	for _, stage := range envelope.Report.Stages {
		if strings.HasPrefix(stage.Name, "generated_direct_read_") {
			if !stage.Passed || stage.CLI.Kind != "ConnectorCommandDirectRead" || stage.CLI.ExitCode != 0 {
				t.Fatalf("generated direct-read fixture stage = %+v, want successful command output", stage)
			}
			generatedStageCount++
		}
	}
	if generatedStageCount != 97 {
		t.Fatalf("generated direct-read fixture stages = %d, want 97 including all 77 restored commands", generatedStageCount)
	}
	observed.Lock()
	requestCount := len(observed.paths)
	observed.Unlock()
	if requestCount < 98 { // 97 direct-read stages plus credential validation.
		t.Fatalf("GitHub direct-read fixture requests = %d, want at least 98 real connector HTTP requests", requestCount)
	}
}

func githubGeneratedDirectReadStageSelection(t *testing.T) string {
	t.Helper()
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub certification candidates: %v", err)
	}
	if bundle.Certification == nil {
		t.Fatal("GitHub certification definition is absent")
	}
	stages := make([]string, 0, len(bundle.Certification.DirectReadCandidates))
	for _, candidate := range bundle.Certification.DirectReadCandidates {
		if candidate.Generated {
			stages = append(stages, candidate.StageName)
		}
	}
	sort.Strings(stages)
	return strings.Join(stages, ",")
}
