package app_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestBatch1CP17AsanaMultipartSavedFrontiers237(t *testing.T) {
	for _, alias := range []string{"upload-attachment-file", "binary-upload-attachment"} {
		for _, frontier := range []string{"healthy", "invalid_filename", "changed_parent"} {
			t.Run(alias+"/"+frontier, func(t *testing.T) {
				var sends atomic.Int32
				payload := []byte{0, 255, 1, 128, '\r', '\n', 'a', 0, 'z'}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					sends.Add(1)
					if r.Method != "POST" || r.URL.RequestURI() != "/attachments" || r.Header.Get("Authorization") != "Bearer synthetic237" {
						t.Error("wrong route/auth")
					}
					reader, err := r.MultipartReader()
					if err != nil {
						t.Error(err)
						return
					}
					seen := map[string]int{}
					for {
						part, err := reader.NextPart()
						if err == io.EOF {
							break
						}
						if err != nil {
							t.Error(err)
							return
						}
						seen[part.FormName()]++
						body, err := io.ReadAll(io.LimitReader(part, 1024))
						if err != nil {
							t.Error(err)
						}
						switch part.FormName() {
						case "parent":
							if string(body) != "task237" {
								t.Error("parent drift reached wire")
							}
						case "file":
							if !bytes.Equal(body, payload) || part.Header.Get("Content-Disposition") != `form-data; name="file"; filename="r%C3%A9sum%C3%A9%20%2B%25.bin"` {
								t.Error("file bytes/header changed")
							}
						default:
							t.Error("unexpected part")
						}
						_ = part.Close()
					}
					if len(seen) != 2 || seen["parent"] != 1 || seen["file"] != 1 {
						t.Error("part multiplicity")
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = io.WriteString(w, `{"data":{"gid":"attachment237"}}`)
				}))
				defer server.Close()
				root := t.TempDir()
				if err := app.InitProject(root); err != nil {
					t.Fatal(err)
				}
				a, err := app.Open(root)
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = a.Close() })
				if _, err := a.AddCredential(t.Context(), app.AddCredentialRequest{Name: "fixture237", Connector: "asana", Config: map[string]string{"base_url": server.URL}, Secrets: map[string]string{"access_token": "synthetic237"}}); err != nil {
					t.Fatal(err)
				}
				filename := "résumé +%.bin"
				if frontier == "invalid_filename" {
					filename = "invalid\nname.bin"
				}
				if err := os.WriteFile(filepath.Join(root, filename), payload, 0600); err != nil {
					t.Fatal(err)
				}
				flags := map[string][]string{"parent": {"task237"}, "file-path": {filename}}
				plan, preview, err := a.PlanConnectorCommand(t.Context(), app.PlanConnectorCommandRequest{Connector: "asana", Credential: "fixture237", Path: []string{"attachments", alias}, Flags: flags, Preview: true})
				if frontier == "invalid_filename" {
					if err == nil || sends.Load() != 0 {
						t.Fatal("invalid filename did not refuse before send")
					}
					t.Log("invalid existing filename refused; physical sends=0")
					return
				}
				if err != nil || preview == nil || preview.Digest == "" || sends.Load() != 0 {
					t.Fatalf("healthy preview: %v sends=%d", err, sends.Load())
				}
				if plan.Action != "upload_attachment_file" {
					t.Fatalf("alias selected action %q", plan.Action)
				}
				if frontier == "changed_parent" {
					flags["parent"] = []string{"changed237"}
				}
				run, err := a.RunReverseETL(t.Context(), app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, WithheldFlags: flags})
				if frontier == "changed_parent" {
					if err == nil || sends.Load() != 0 {
						t.Fatalf("changed parent did not refuse: err=%v sends=%d", err, sends.Load())
					}
					t.Log("approved parent drift refused; physical sends=0")
					return
				}
				if err != nil || run.Status != "completed" || run.RecordsSucceeded != 1 || sends.Load() != 1 {
					t.Fatalf("saved route: err=%v sends=%d", err, sends.Load())
				}
				local, err := os.ReadFile(filepath.Join(root, filename))
				if err != nil || !bytes.Equal(local, payload) {
					t.Fatal("local artifact changed")
				}
				t.Log("saved actual App completed: exact9 file bytes, parent task237, one send, local bytes unchanged")
			})
		}
	}
}
