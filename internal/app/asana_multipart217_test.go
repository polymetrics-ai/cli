package app_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestAsanaMultipartAliases217(t *testing.T) {
	for _, alias := range []string{"upload-attachment-file", "binary-upload-attachment"} {
		t.Run(alias, func(t *testing.T) {
			var mu sync.Mutex
			var requests []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, err := io.ReadAll(io.LimitReader(r.Body, 65537))
				if err != nil {
					t.Error(err)
				}
				mu.Lock()
				requests = append(requests, r.Method+" "+r.URL.RequestURI()+"\n"+string(raw))
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"gid":"attachment-local"}}`))
			}))
			defer server.Close()
			count := func() int { mu.Lock(); defer mu.Unlock(); return len(requests) }
			root := t.TempDir()
			if err := app.InitProject(root); err != nil {
				t.Fatal(err)
			}
			a, err := app.Open(root)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = a.Close() })
			ctx := t.Context()
			if _, err = a.AddCredential(ctx, app.AddCredentialRequest{Name: "asana-local", Connector: "asana", Config: map[string]string{"base_url": server.URL}, Secrets: map[string]string{"access_token": "synthetic-local-only"}}); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(root, "résumé +%.txt"), []byte("exact attachment payload217"), 0600); err != nil {
				t.Fatal(err)
			}
			plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{Connector: "asana", Credential: "asana-local", Path: []string{"attachments", alias}, Flags: map[string][]string{"parent": {"task-local"}, "file-path": {"résumé +%.txt"}}, Preview: true})
			if err != nil {
				t.Fatal(err)
			}
			if preview == nil || preview.Digest == "" || count() != 0 {
				t.Fatal("missing preview or unexpected request")
			}
			if _, err = a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID}); err == nil || count() != 0 {
				t.Fatal("missing approval reached provider")
			}
			run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, WithheldFlags: map[string][]string{"parent": {"task-local"}, "file-path": {"résumé +%.txt"}}})
			if err != nil {
				t.Fatal(err)
			}
			if run.Status != "completed" || run.RecordsSucceeded != 1 || count() != 1 {
				t.Fatalf("outcome status %s, sends %d", run.Status, count())
			}
			mu.Lock()
			wire := requests[0]
			mu.Unlock()
			for _, literal := range []string{"POST /attachments\n", `name="parent"`, "task-local", `name="file"; filename="r%C3%A9sum%C3%A9%20%2B%25.txt"`, "exact attachment payload217"} {
				if !strings.Contains(wire, literal) {
					t.Errorf("wire missing literal %q", literal)
				}
			}
			if strings.Contains(wire, "filename*=") {
				t.Error("unexpected filename* header")
			}
			_, _ = a.RunReverseETL(ctx, app.RunReverseETLRequest{PlanID: plan.ID, ApprovalToken: plan.ApprovalToken, WithheldFlags: map[string][]string{"parent": {"task-local"}, "file-path": {"résumé +%.txt"}}})
			if count() != 1 {
				t.Fatal("replayed approval sent")
			}
		})
	}
}
