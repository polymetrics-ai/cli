package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
)

func TestWriteConnectorCommandResultPreservesStatusCheckJSON(t *testing.T) {
	result := commandrunner.Result{
		Connector: "github-live-qualification",
		Command:   "dataset status",
		Stream:    "must-not-be-rendered",
		Count:     99,
		StatusCheck: &connectors.OperationStatusCheckResult{
			Connector: "github-live-qualification",
			Operation: "github_live.dataset.status",
			Method:    "HEAD",
			Path:      "/vega-datasets/data/seattle-weather.csv",
			Status:    200,
			BodyBytes: 0,
		},
	}
	rows := []connectors.Record{{"must_not": "be rendered as an ETL record"}}

	var stdout, stderr bytes.Buffer
	if err := writeConnectorCommandResult(&stdout, &stderr, true, result, rows); err != nil {
		t.Fatalf("writeConnectorCommandResult: %v", err)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("status JSON stderr = %q, want empty", got)
	}

	var got struct {
		APIVersion string `json:"api_version"`
		Kind       string `json:"kind"`
		Connector  string `json:"connector"`
		Command    string `json:"command"`
		Operation  string `json:"operation"`
		Method     string `json:"method"`
		Path       string `json:"path"`
		Status     int    `json:"status"`
		BodyBytes  int    `json:"body_bytes"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode status JSON: %v\n%s", err, stdout.String())
	}
	if want := (struct {
		APIVersion string
		Kind       string
		Connector  string
		Command    string
		Operation  string
		Method     string
		Path       string
		Status     int
		BodyBytes  int
	}{
		APIVersion: apiVersion,
		Kind:       "ConnectorCommandStatusCheck",
		Connector:  "github-live-qualification",
		Command:    "dataset status",
		Operation:  "github_live.dataset.status",
		Method:     "HEAD",
		Path:       "/vega-datasets/data/seattle-weather.csv",
		Status:     200,
		BodyBytes:  0,
	}); got.APIVersion != want.APIVersion || got.Kind != want.Kind || got.Connector != want.Connector || got.Command != want.Command || got.Operation != want.Operation || got.Method != want.Method || got.Path != want.Path || got.Status != want.Status || got.BodyBytes != want.BodyBytes {
		t.Fatalf("status JSON envelope = %+v, want %+v", got, want)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(stdout.Bytes(), &raw); err != nil {
		t.Fatalf("decode raw status JSON: %v", err)
	}
	for _, forbidden := range []string{"stream", "count", "records"} {
		if _, ok := raw[forbidden]; ok {
			t.Fatalf("status JSON retained ETL fallback field %q: %s", forbidden, stdout.String())
		}
	}
}

func TestWriteConnectorCommandResultPreservesStatusCheckHumanOutput(t *testing.T) {
	for _, tt := range []struct {
		name   string
		status int
	}{
		{name: "successful zero-byte HEAD", status: 200},
		{name: "non-200 status metadata remains visible", status: 503},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result := commandrunner.Result{
				Connector: "github-live-qualification",
				Command:   "dataset status",
				StatusCheck: &connectors.OperationStatusCheckResult{
					Connector: "github-live-qualification",
					Operation: "github_live.dataset.status",
					Method:    "HEAD",
					Path:      "/vega-datasets/data/seattle-weather.csv",
					Status:    tt.status,
					BodyBytes: 0,
				},
			}

			var stdout, stderr bytes.Buffer
			if err := writeConnectorCommandResult(&stdout, &stderr, false, result, nil); err != nil {
				t.Fatalf("writeConnectorCommandResult: %v", err)
			}
			want := "connector=github-live-qualification command=\"dataset status\" operation=github_live.dataset.status method=HEAD path=/vega-datasets/data/seattle-weather.csv status=" + strconv.Itoa(tt.status) + " body_bytes=0\n"
			if got := stdout.String(); got != want {
				t.Fatalf("status human output = %q, want %q", got, want)
			}
			if got := stderr.String(); got != "" {
				t.Fatalf("status human stderr = %q, want empty", got)
			}
		})
	}
}

func TestWriteConnectorCommandResultPreservesDeclaredStatusHeadersForHumanOutput(t *testing.T) {
	result := commandrunner.Result{
		Connector: "github-live-qualification",
		Command:   "dataset status",
		StatusCheck: &connectors.OperationStatusCheckResult{
			Connector: "github-live-qualification",
			Operation: "github_live.dataset.status",
			Method:    "HEAD",
			Path:      "/vega-datasets/data/seattle-weather.csv",
			Status:    404,
			BodyBytes: 0,
			Headers: map[string]connectors.OperationResponseHeader{
				"X-Provider-Status": {Values: []string{"not-found"}},
			},
		},
	}

	var stdout, stderr bytes.Buffer
	if err := writeConnectorCommandResult(&stdout, &stderr, false, result, nil); err != nil {
		t.Fatalf("writeConnectorCommandResult: %v", err)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("status header stderr = %q, want empty", got)
	}
	var got struct {
		Status  int `json:"status"`
		Headers map[string]struct {
			Values []string `json:"values"`
		} `json:"headers"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode human status headers: %v\n%s", err, stdout.String())
	}
	if got.Status != 404 || len(got.Headers["X-Provider-Status"].Values) != 1 || got.Headers["X-Provider-Status"].Values[0] != "not-found" {
		t.Fatalf("human status headers = %#v, want declared final 404 metadata", got)
	}
}

func TestWriteConnectorCommandResultPreservesBinaryDownloadEnvelope(t *testing.T) {
	result := commandrunner.Result{
		Connector: "github-live-qualification",
		Command:   "dataset export",
		BinaryDownload: &connectors.OperationBinaryDownloadResult{
			Operation: "github_live.dataset.export",
			Record: connectors.Record{
				"file_name":       "seattle-weather.csv",
				"file_size_bytes": 48219,
				"file_sha256":     "0845078a290b48e3149ab8639966824110a251db4e06fc144c06ebb534af23be",
				"truncated":       false,
			},
			Receipt: &connectors.ProviderResponseReceipt{
				ResponseReceived: true,
				Status:           200,
				BodyPresent:      true,
				BodyBytes:        48219,
				Body: connectors.Record{
					"file_size_bytes": 48219,
					"file_sha256":     "0845078a290b48e3149ab8639966824110a251db4e06fc144c06ebb534af23be",
				},
			},
		},
	}

	var stdout, stderr bytes.Buffer
	if err := writeConnectorCommandResult(&stdout, &stderr, true, result, nil); err != nil {
		t.Fatalf("writeConnectorCommandResult: %v", err)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("binary download stderr = %q, want empty", got)
	}

	var got struct {
		Kind      string `json:"kind"`
		Connector string `json:"connector"`
		Command   string `json:"command"`
		Operation string `json:"operation"`
		Record    struct {
			FileName      string `json:"file_name"`
			FileSizeBytes int    `json:"file_size_bytes"`
			FileSHA256    string `json:"file_sha256"`
			Truncated     bool   `json:"truncated"`
		} `json:"record"`
		Receipt struct {
			ResponseReceived bool `json:"response_received"`
			Status           int  `json:"status"`
			BodyPresent      bool `json:"body_present"`
			BodyBytes        int  `json:"body_bytes"`
			Body             struct {
				FileSHA256 string `json:"file_sha256"`
			} `json:"body"`
		} `json:"receipt"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode binary download JSON: %v\n%s", err, stdout.String())
	}
	if got.Kind != "ConnectorCommandBinaryDownload" || got.Connector != "github-live-qualification" || got.Command != "dataset export" || got.Operation != "github_live.dataset.export" || got.Record.FileName != "seattle-weather.csv" || got.Record.FileSizeBytes != 48219 || got.Record.FileSHA256 != "0845078a290b48e3149ab8639966824110a251db4e06fc144c06ebb534af23be" || got.Record.Truncated {
		t.Fatalf("binary download envelope = %+v", got)
	}
	if !got.Receipt.ResponseReceived || got.Receipt.Status != 200 || !got.Receipt.BodyPresent || got.Receipt.BodyBytes != 48219 || got.Receipt.Body.FileSHA256 != got.Record.FileSHA256 {
		t.Fatalf("binary download receipt = %+v", got.Receipt)
	}
}

func TestWriteConnectorCommandFailureResultPreservesBinaryDownloadErrorEnvelope(t *testing.T) {
	result := commandrunner.Result{
		Connector: "youtube-analytics",
		Command:   "reports download",
		BinaryDownload: &connectors.OperationBinaryDownloadResult{
			Operation: "download_report",
			Method:    http.MethodGet,
			Path:      "/v1/media/jobs/job_fixture/reports/report_fixture?alt=media",
			Status:    http.StatusOK,
			Receipt: &connectors.ProviderResponseReceipt{
				ResponseReceived: true,
				Status:           http.StatusOK,
				BodyPresent:      true,
				BodyBytes:        9,
			},
		},
	}

	var stdout, stderr bytes.Buffer
	err := writeConnectorCommandFailureResult(&stdout, &stderr, true, result, nil, errors.New("binary download response too large: exceeds limit 8 bytes"))
	if err != nil {
		t.Fatalf("writeConnectorCommandFailureResult: %v", err)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("binary download error stderr = %q, want empty", got)
	}

	var got struct {
		Kind    string `json:"kind"`
		Receipt struct {
			ResponseReceived bool  `json:"response_received"`
			Status           int   `json:"status"`
			BodyBytes        int64 `json:"body_bytes"`
		} `json:"receipt"`
		Error struct {
			Category string `json:"category"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode binary download failure JSON: %v\n%s", err, stdout.String())
	}
	if got.Kind != "ConnectorCommandBinaryDownload" || !got.Receipt.ResponseReceived || got.Receipt.Status != http.StatusOK || got.Receipt.BodyBytes != 9 {
		t.Fatalf("binary download failure receipt = %+v", got)
	}
	if got.Error.Category != "internal" || got.Error.Code != "internal_error" || !strings.Contains(got.Error.Message, "too large") {
		t.Fatalf("binary download failure error = %+v", got.Error)
	}
}

func TestWriteConnectorCommandFailureResultPreservesStatusReceipt(t *testing.T) {
	result := commandrunner.Result{
		Connector: "github-live-qualification",
		Command:   "dataset status",
		StatusCheck: &connectors.OperationStatusCheckResult{
			Operation: "github_live.dataset.status", Method: http.MethodHead, Path: "/locked.csv", Status: http.StatusNotFound,
			Receipt: &connectors.ProviderResponseReceipt{
				ResponseReceived: true, Status: http.StatusNotFound, BodyPresent: false, BodyBytes: 0,
				Headers: map[string]connectors.OperationResponseHeader{"X-Provider-Trace": {Values: []string{"first", "second"}}},
			},
		},
	}
	var stdout, stderr bytes.Buffer
	if err := writeConnectorCommandFailureResult(&stdout, &stderr, true, result, nil, errors.New("provider redirect refused")); err != nil {
		t.Fatalf("writeConnectorCommandFailureResult: %v", err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("status failure stderr = %q, want one stdout envelope", stderr.String())
	}
	var got struct {
		Kind    string `json:"kind"`
		Status  int    `json:"status"`
		Receipt struct {
			ResponseReceived bool `json:"response_received"`
			Status           int  `json:"status"`
		} `json:"receipt"`
		Error json.RawMessage `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode status failure JSON: %v\n%s", err, stdout.String())
	}
	if got.Kind != "ConnectorCommandStatusCheck" || got.Status != http.StatusNotFound || !got.Receipt.ResponseReceived || got.Receipt.Status != http.StatusNotFound || len(got.Error) == 0 {
		t.Fatalf("status failure envelope = %+v", got)
	}
}

func TestConnectorCommandFailureEnvelopeRecognizesPostProviderResultWithoutReceipt(t *testing.T) {
	if !connectorCommandHasPostProviderResult(commandrunner.Result{
		DirectRead: &connectors.DirectReadResult{
			Operation: "github.records.lookup", Method: http.MethodGet, Path: "/records/fixture", Status: http.StatusBadGateway,
			Body: map[string]any{"occurrence_id": "provider-occurrence-9007199254740993"},
		},
	}) {
		t.Fatal("post-provider direct result without receipt was not admitted to the CLI failure envelope")
	}
	if connectorCommandHasPostProviderResult(commandrunner.Result{}) {
		t.Fatal("empty runner result was admitted to the CLI failure envelope")
	}
}

func TestWriteConnectorCommandResultPreservesDirectReadReceipt(t *testing.T) {
	result := commandrunner.Result{
		Connector: "github-live-qualification",
		Command:   "records lookup",
		DirectRead: &connectors.DirectReadResult{
			Operation: "github_live.records.lookup", Method: "GET", Path: "/records/1", Status: http.StatusBadGateway,
			Receipt: &connectors.ProviderResponseReceipt{
				ResponseReceived: true, Status: http.StatusBadGateway, BodyPresent: true, BodyBytes: 62,
				BodyRaw: `{"occurrence_id":"read-occurrence-9007199254740993"}`,
				Body:    map[string]any{"occurrence_id": "read-occurrence-9007199254740993"},
			},
		},
	}

	var stdout, stderr bytes.Buffer
	if err := writeConnectorCommandResult(&stdout, &stderr, true, result, nil); err != nil {
		t.Fatalf("writeConnectorCommandResult: %v", err)
	}
	var got struct {
		Kind      string `json:"kind"`
		Operation string `json:"operation"`
		Receipt   struct {
			Status  int    `json:"status"`
			BodyRaw string `json:"body_raw"`
		} `json:"receipt"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode direct-read JSON: %v\n%s", err, stdout.String())
	}
	if got.Kind != "ConnectorCommandDirectRead" || got.Operation != "github_live.records.lookup" || got.Receipt.Status != http.StatusBadGateway || !strings.Contains(got.Receipt.BodyRaw, "9007199254740993") {
		t.Fatalf("direct-read receipt envelope = %+v", got)
	}
}
