package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
)

func TestBatch1CP17AsanaUploadExactBytesAndApproval(t *testing.T) {
	for _, command := range []string{"upload-attachment-file", "binary-upload-attachment"} {
		t.Run(command, func(t *testing.T) {
			batch1CP17AsanaUpload(t, command, "payload.bin", "payload.bin")
		})
	}
}

func TestBatch1CP17AsanaUploadSourceFilename(t *testing.T) {
	for _, command := range []string{"upload-attachment-file", "binary-upload-attachment"} {
		t.Run(command, func(t *testing.T) {
			batch1CP17AsanaUpload(t, command, "résumé +%.bin", "r%C3%A9sum%C3%A9%20%2B%25.bin")
		})
	}
}

func batch1CP17AsanaUpload(t *testing.T, command, filename, expectedFilename string) {
	batch1CP17AsanaUploadSize(t, command, filename, expectedFilename, 9)
}

func TestBatch1CP17AsanaUploadExactLimit(t *testing.T) {
	for _, command := range []string{"upload-attachment-file", "binary-upload-attachment"} {
		t.Run(command, func(t *testing.T) {
			batch1CP17AsanaUploadSize(t, command, "limit.bin", "limit.bin", 104857600)
		})
	}
}

func TestBatch1CP17AsanaUploadBelowLimit(t *testing.T) {
	for _, command := range []string{"upload-attachment-file", "binary-upload-attachment"} {
		t.Run(command, func(t *testing.T) {
			batch1CP17AsanaUploadSize(t, command, "near-limit.bin", "near-limit.bin", 104856576)
		})
	}
}

func batch1CP17AsanaUploadSize(t *testing.T, command, filename, expectedFilename string, payloadSize int64) {
	t.Helper()
	payload := []byte{0, 255, 1, 128, '\r', '\n', 'a', 0, 'z'}
	expectedHash := sha256.Sum256(payload)
	if payloadSize != int64(len(payload)) {
		hash := sha256.New()
		chunk := make([]byte, 65536)
		for remaining := payloadSize; remaining > 0; {
			n := int64(len(chunk))
			if remaining < n {
				n = remaining
			}
			_, _ = hash.Write(chunk[:n])
			remaining -= n
		}
		copy(expectedHash[:], hash.Sum(nil))
	}
	var sends atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sends.Add(1)
		if r.Method != "POST" || r.URL.Path != "/attachments" || r.URL.RawQuery != "" {
			t.Errorf("unexpected method/path/query: %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Authorization") != "Bearer synthetic-cp17-upload" {
			t.Error("wrong fixture auth")
		}
		media, parameters, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "multipart/form-data" {
			t.Error("wrong multipart media")
		}
		capture := &batch1CP17WireCapture{ReadCloser: r.Body}
		r.Body = capture
		reader, err := r.MultipartReader()
		if err != nil {
			t.Error(err)
			http.Error(w, "multipart", 400)
			return
		}
		parts := map[string]int{}
		fileHeader := `Content-Disposition: form-data; name="file"; filename="` + expectedFilename + `"`
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Error(err)
				break
			}
			parts[part.FormName()]++
			switch part.FormName() {
			case "parent":
				body, err := io.ReadAll(io.LimitReader(part, 1024))
				if err != nil || string(body) != "task-upload" {
					t.Error("wrong parent")
				}
			case "file":
				hash := sha256.New()
				n, err := io.Copy(hash, io.LimitReader(part, payloadSize+1))
				bytesEqual := err == nil && n == payloadSize && bytes.Equal(hash.Sum(nil), expectedHash[:])
				if part.Header.Get("Content-Disposition") != strings.TrimPrefix(fileHeader, "Content-Disposition: ") {
					t.Error("raw file disposition differs from literal expected header")
				}
				if !bytesEqual || part.FileName() != expectedFilename || part.Header.Get("Content-Type") != "" {
					t.Errorf("upload bytes_equal=%t byte_count=%d filename=%q expected_filename=%q media=%q", bytesEqual, n, part.FileName(), expectedFilename, part.Header.Get("Content-Type"))
				}
			default:
				t.Errorf("undeclared part %q", part.FormName())
			}
			if err := part.Close(); err != nil {
				t.Error(err)
			}
		}
		if len(parts) != 2 || parts["parent"] != 1 || parts["file"] != 1 {
			t.Errorf("wrong part multiplicity: %v", parts)
		}
		_, _ = io.Copy(io.Discard, capture)
		boundary := parameters["boundary"]
		expectedMetadata := int64(len("--"+boundary+"\r\n"+fileHeader+"\r\n\r\n") + len("\r\n") + len("--"+boundary+"\r\nContent-Disposition: form-data; name=\"parent\"\r\n\r\ntask-upload\r\n") + len("--"+boundary+"--\r\n"))
		if capture.count != payloadSize+expectedMetadata || r.ContentLength != -1 || len(r.TransferEncoding) != 1 || r.TransferEncoding[0] != "chunked" || expectedMetadata > 65536 || capture.count > 104923136 {
			t.Errorf("wire count=%d content_length=%d expected=%d metadata=%d", capture.count, r.ContentLength, payloadSize+expectedMetadata, expectedMetadata)
		}
		if !bytes.Contains(capture.prefix.Bytes(), []byte(fileHeader+"\r\n")) || bytes.Contains(capture.prefix.Bytes(), []byte("filename*=")) {
			t.Error("literal raw header missing or filename-star present")
		}
		t.Logf("alias=%s file_bytes=%d sha256=%x wire_bytes=%d metadata_bytes=%d filename=%q physical_send=%d", command, payloadSize, expectedHash, capture.count, expectedMetadata, expectedFilename, sends.Load())
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":{"gid":"attachment-created","resource_type":"attachment"}}`)
	}))
	defer server.Close()
	root := t.TempDir()
	runCLIForReverseTest(t, []string{"init", "--root", root, "--json"})
	if payloadSize == int64(len(payload)) {
		if err := os.WriteFile(filepath.Join(root, filename), payload, 0600); err != nil {
			t.Fatal(err)
		}
	} else {
		file, err := os.Create(filepath.Join(root, filename))
		if err != nil {
			t.Fatal(err)
		}
		if err := file.Truncate(payloadSize); err != nil {
			_ = file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PM_CP17_UPLOAD", "synthetic-cp17-upload")
	runCLIForReverseTest(t, []string{"credentials", "add", "fixture", "--connector", "asana", "--config", "base_url=" + server.URL, "--from-env", "access_token=PM_CP17_UPLOAD", "--root", root, "--json"})
	base := []string{"asana", "attachments", command}
	args := append(append([]string{}, base...), "--parent", "task-upload", "--file-path", filename, "--credential", "fixture", "--root", root)
	var out, diag bytes.Buffer
	if cli.Run(args, &out, &diag) != 0 {
		t.Fatalf("plan upload: %s", diag.String())
	}
	plan := extractReverseField(t, out.String(), `Created connector command plan (\S+)`)
	token := ""
	// Saved direct writes issue the token at plan time; binary uploads
	// require payload preview before issuing theirs.
	if strings.Contains(out.String(), "Approval token:") {
		token = extractReverseField(t, out.String(), `Approval token: (\S+)`)
	}
	if sends.Load() != 0 {
		t.Fatal("plan sent a mutation")
	}
	continuation := append(append([]string{}, base...), "--plan", plan, "--root", root, "--json", "--parent", "task-upload", "--file-path", filename)
	out.Reset()
	diag.Reset()
	previewArgs := append(append([]string{}, base...), "--plan", plan, "--root", root, "--preview", "--parent", "task-upload", "--file-path", filename)
	if cli.Run(previewArgs, &out, &diag) != 0 {
		t.Fatalf("preview: %s", diag.String())
	}
	if sends.Load() != 0 {
		t.Fatal("preview sent a mutation")
	}
	if token == "" {
		token = extractReverseField(t, out.String(), `Approval token: (\S+)`)
	}
	approved := append(append([]string{}, continuation...), "--approval-token-stdin")
	out.Reset()
	diag.Reset()
	if runCLIWithApprovalStdin(t, approved, "incorrect\n", &out, &diag) == 0 || sends.Load() != 0 {
		t.Fatal("wrong approval crossed boundary")
	}
	out.Reset()
	diag.Reset()
	if code := runCLIWithApprovalStdin(t, approved, token+"\n", &out, &diag); code != 0 {
		var result struct {
			Run struct {
				Error string `json:"error"`
			} `json:"run"`
		}
		parsed := json.Unmarshal(out.Bytes(), &result) == nil
		t.Fatalf("approved upload exit=%d sends=%d bytes=%d run_json=%t multipart_bound=%t file_bound=%t", code, sends.Load(), payloadSize, parsed, strings.Contains(result.Run.Error, "multipart payload too large"), strings.Contains(result.Run.Error, "file too large"))
	}
	var envelope struct {
		Run struct {
			DestinationResult json.RawMessage `json:"destination_result"`
		} `json:"run"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	var receipt connectors.WriteResult
	if err := json.Unmarshal(envelope.Run.DestinationResult, &receipt); err != nil {
		t.Fatal(err)
	}
	expectedResponse := `{"data":{"gid":"attachment-created","resource_type":"attachment"}}`
	if len(receipt.ProviderResponses) != 1 {
		t.Fatal("missing exact provider receipt")
	}
	response := receipt.ProviderResponses[0]
	if response.Status != 200 || !response.BodyPresent || response.BodyBytes != len(expectedResponse) || response.BodyRaw != expectedResponse {
		t.Fatal("provider result bytes/status changed")
	}
	local, err := os.Open(filepath.Join(root, filename))
	if err != nil {
		t.Fatal(err)
	}
	localHash := sha256.New()
	localCount, localErr := io.Copy(localHash, local)
	_ = local.Close()
	if localErr != nil || localCount != payloadSize || !bytes.Equal(localHash.Sum(nil), expectedHash[:]) {
		t.Fatal("local source bytes/identity changed")
	}
	t.Logf("local artifact retained: name=%q bytes=%d digest=%s", filename, localCount, fmt.Sprintf("%x", localHash.Sum(nil)))
	if sends.Load() != 1 || !strings.Contains(out.String(), `"records_succeeded": 1`) {
		t.Fatal("upload did not reconcile one physical send")
	}
	out.Reset()
	diag.Reset()
	if runCLIWithApprovalStdin(t, approved, token+"\n", &out, &diag) == 0 || sends.Load() != 1 {
		t.Fatal("consumed upload approval replayed")
	}
	// Sparse oversized input tests the declared functional file bound
	// without allocating or sending 100 MiB.
	oversized, err := os.Create(filepath.Join(root, "too-large.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if err := oversized.Truncate(104857601); err != nil {
		t.Fatal(err)
	}
	if err := oversized.Close(); err != nil {
		t.Fatal(err)
	}
	bad := append(append([]string{}, base...), "--parent", "task-upload", "--file-path", "too-large.bin", "--credential", "fixture", "--root", root)
	out.Reset()
	diag.Reset()
	code := cli.Run(bad, &out, &diag)
	stage := "plan"
	if code == 0 {
		stage = "execute"
		// The direct-write alias defers file inspection until execution.
		badPlan := extractReverseField(t, out.String(), `Created connector command plan (\S+)`)
		badToken := extractReverseField(t, out.String(), `Approval token: (\S+)`)
		badApply := append(append([]string{}, base...), "--plan", badPlan, "--root", root, "--json", "--parent", "task-upload", "--file-path", "too-large.bin", "--approval-token-stdin")
		out.Reset()
		diag.Reset()
		code = runCLIWithApprovalStdin(t, badApply, badToken+"\n", &out, &diag)
	}
	t.Logf("alias=%s stage=%s exit=%d send_delta=%d stderr_bound=%t stdout_bound=%t stderr_size=%t stdout_size=%t failed_record=%t approval_error=%t", command, stage, code, sends.Load()-1, strings.Contains(diag.String(), "104857600"), strings.Contains(out.String(), "104857600"), strings.Contains(diag.String(), "too large"), strings.Contains(out.String(), "too large"), strings.Contains(out.String(), `"records_failed": 1`), strings.Contains(diag.String(), "approval"))
	if code == 0 {
		t.Error("oversized input returned exit 0; expected nonzero")
	}
	if sends.Load() != 1 {
		t.Errorf("oversized input physical send delta=%d; expected 0", sends.Load()-1)
	}
	boundError := diag.String()
	boundClassification := "payload file exceeds declared byte cap"
	if stage == "execute" {
		boundClassification = "file too large"
		// A started saved write reports its execution failure in the
		// ReverseRun JSON envelope, with a nonzero process result.
		var result struct {
			Kind string `json:"kind"`
			Run  struct {
				RecordsSucceeded int    `json:"records_succeeded"`
				RecordsFailed    int    `json:"records_failed"`
				Error            string `json:"error"`
			} `json:"run"`
		}
		if err := json.Unmarshal(out.Bytes(), &result); err != nil {
			t.Fatal("oversized execution did not return valid JSON")
		}
		if result.Kind != "ReverseRun" || result.Run.RecordsSucceeded != 0 || result.Run.RecordsFailed != 1 {
			t.Error("oversized execution did not reconcile exactly one failed record")
		}
		boundError = result.Run.Error
	}
	if !strings.Contains(boundError, "104857600") || !strings.Contains(boundError, boundClassification) {
		t.Error("oversized input lacks size-limit classification and expected bound 104857600")
	}
}

// Captures the literal wire header prefix while counting the entire streamed body.
// The fixed prefix bound does not buffer the large file fixture.
type batch1CP17WireCapture struct {
	io.ReadCloser
	count  int64
	prefix bytes.Buffer
}

func (c *batch1CP17WireCapture) Read(p []byte) (int, error) {
	n, err := c.ReadCloser.Read(p)
	c.count += int64(n)
	keep := 4096 - c.prefix.Len()
	if keep > n {
		keep = n
	}
	if keep > 0 {
		_, _ = c.prefix.Write(p[:keep])
	}
	return n, err
}
