package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synctransport"
)

// TestGitLabSourceTransportMaterializesConnectionOwnedDuckDB is a local
// provider-double proof for the narrow, source-cited GitLab full snapshot
// lane. Its fixture inputs come from the immutable source lock and the
// enabled-connector contract; the double proves our request routing and
// lifecycle only, not any undeclared GitLab cursor, tombstone, or order
// semantics.
func TestGitLabSourceTransportMaterializesConnectionOwnedDuckDB(t *testing.T) {
	fixture := newGitLabTransportFixture(t, false)
	root := fixture.setupProject(t)

	runCLI(t, []string{
		"etl", "run", "--connection", "gitlab_groups_to_warehouse", "--stream", fixture.stream,
		"--root", root, "--json",
	})
	fixture.assertSourceReads(t, 2)
	fixture.assertDeletes(t, nil)

	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("reopen project after GitLab stage: %v", err)
	}
	rows, err := application.QueryTable(context.Background(), app.QueryTableRequest{Table: "gitlab_groups", Limit: 10})
	if err != nil {
		t.Fatalf("query connection-owned GitLab DuckDB table: %v", err)
	}
	if len(rows) != 2 || rows[0]["path"] == nil || rows[1]["path"] == nil {
		t.Fatalf("GitLab staged rows = %#v, want both source-lock fixture pages", rows)
	}

	planStdout, _ := runCLI(t, []string{
		"reverse", "plan", "gitlab_group_delete",
		"--source-table", "gitlab_groups",
		"--connection", "gitlab_groups_to_warehouse",
		"--destination", "gitlab:gitlab-local",
		"--action", fixture.deleteAction,
		"--map", "path:id",
		"--root", root, "--json",
	})
	planID := decodeGitLabReversePlanID(t, planStdout)
	fixture.assertDeletes(t, nil)

	previewStdout, _ := runCLI(t, []string{"reverse", "preview", planID, "--root", root, "--json"})
	var preview struct {
		Kind string `json:"kind"`
		Plan struct {
			RecordCount int `json:"record_count"`
		} `json:"plan"`
	}
	if err := json.Unmarshal([]byte(previewStdout), &preview); err != nil {
		t.Fatalf("decode GitLab reverse preview: %v\n%s", err, previewStdout)
	}
	if preview.Kind != "ReversePlanPreview" || preview.Plan.RecordCount != 2 {
		t.Fatalf("GitLab reverse preview = %+v, want a two-record DuckDB plan", preview)
	}
	fixture.assertDeletes(t, nil)

	humanPreview, _ := runCLI(t, []string{"reverse", "preview", planID, "--root", root})
	approvalToken := extractReverseField(t, humanPreview, `Approval token: (\S+)`)
	fixture.assertDeletes(t, nil)

	var runStdout, runStderr bytes.Buffer
	code := runCLIWithApprovalStdin(t, []string{
		"reverse", "run", planID,
		"--approval-token-stdin",
		"--confirm", "destructive",
		"--root", root, "--json",
	}, approvalToken+"\n", &runStdout, &runStderr)
	if code != 0 {
		t.Fatalf("approved GitLab reverse run code=%d stderr=%s stdout=%s", code, runStderr.String(), runStdout.String())
	}
	fixture.assertDeletes(t, []string{"group-one", "group-two"})
}

func TestGitLabSourceTransportFixtureStopsOnDeclaredNonSuccess(t *testing.T) {
	fixture := newGitLabTransportFixture(t, true)
	root := fixture.setupProject(t)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"etl", "run", "--connection", "gitlab_groups_to_warehouse", "--stream", fixture.stream,
		"--root", root, "--json",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("GitLab source 404 unexpectedly succeeded: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String()+stderr.String(), "404") {
		t.Fatalf("GitLab source 404 error did not retain status: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	fixture.assertSourceReads(t, 1)
	fixture.assertDeletes(t, nil)

	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("reopen project after failed GitLab source read: %v", err)
	}
	if rows, err := application.QueryTable(context.Background(), app.QueryTableRequest{Table: "gitlab_groups", Limit: 10}); err == nil || len(rows) != 0 {
		t.Fatalf("failed GitLab source read materialized rows=%#v err=%v", rows, err)
	}
}

// TestGitLabManagedDestinationDuckDBLifecycle is the fake-server contract for
// GitLab's declaration-owned single-attempt destination. It first materializes
// one group through the real DuckDB stage, then verifies the closed
// plan/preview/approval route emits exactly one destination request. Before
// this contract was admitted, CreateConnection refused GitLab as a destination
// before either source or target request reached the double; the retained error
// branch makes that regression explicit. No test adapter is installed and no
// production admission path is bypassed.
func TestGitLabManagedDestinationDuckDBLifecycle(t *testing.T) {
	for _, testCase := range []struct {
		name           string
		deleteStatus   int
		ambiguousWrite bool
		wantSuccess    bool
	}{
		{name: "acknowledged 202 checkpoints one DuckDB record", deleteStatus: http.StatusAccepted, wantSuccess: true},
		{name: "provider 4xx has no automatic retry", deleteStatus: http.StatusBadRequest},
		{name: "provider 5xx has no automatic retry", deleteStatus: http.StatusInternalServerError},
		{name: "timeout ambiguity has no automatic replay", ambiguousWrite: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fixture := newGitLabTransportFixture(t, false)
			fixture.setSourceGroups([]gitLabTransportGroup{{ID: 1, Path: "group-one"}})
			fixture.deleteStatus = testCase.deleteStatus
			fixture.blockDeleteUntilCancelled = testCase.ambiguousWrite
			root := fixture.setupProject(t)

			// Materialize the provider row first. The managed route below must
			// independently stage its selected source page through the same
			// DuckDB boundary before it can address GitLab as a destination.
			runCLI(t, []string{
				"etl", "run", "--connection", "gitlab_groups_to_warehouse", "--stream", fixture.stream,
				"--root", root, "--json",
			})
			fixture.assertSourceReads(t, 1)
			fixture.assertDeletes(t, nil)
			fixture.clearCapturedRequests()

			application, err := app.Open(root)
			if err != nil {
				t.Fatalf("open GitLab DuckDB project: %v", err)
			}
			connection, err := application.CreateConnection(context.Background(), app.CreateConnectionRequest{
				Name:        "gitlab_groups_to_gitlab",
				Source:      app.EndpointConfig{Connector: "gitlab", Credential: "gitlab-local"},
				Destination: app.EndpointConfig{Connector: "gitlab", Credential: "gitlab-local"},
				Streams: map[string]app.StreamConfig{fixture.stream: {
					SyncMode: "full_append", DestinationAction: fixture.deleteAction,
				}},
			})
			if err != nil {
				fixture.assertNoCapturedRequests(t)
				if !strings.Contains(err.Error(), "not a declarative typed destination") {
					t.Fatalf("GitLab managed destination preflight = %v, want the current closed destination-admission refusal before HTTP", err)
				}
				t.Fatalf("RED: GitLab managed destination is refused before plan/preview/approval and before HTTP (0 captured requests): %v; add only a source-backed production destination contract before making this lifecycle green", err)
			}

			plan, err := application.PlanDeclarativeTypedDestinationTransport(context.Background(), connection.Name, fixture.stream)
			if err != nil {
				t.Fatalf("plan GitLab managed destination: %v", err)
			}
			previewed, _, err := application.PreviewDeclarativeTypedDestinationTransport(context.Background(), plan.ID)
			if err != nil {
				t.Fatalf("preview GitLab managed destination: %v", err)
			}

			runContext := context.Background()
			cancel := func() {}
			if testCase.ambiguousWrite {
				runContext, cancel = context.WithCancel(context.Background())
			}
			defer cancel()
			type runResult struct {
				run app.Run
				err error
			}
			completed := make(chan runResult, 1)
			go func() {
				run, runErr := application.RunETL(runContext, app.RunETLRequest{
					Connection: connection.Name,
					Stream:     fixture.stream,
					BatchSize:  1,
					DestinationApproval: synctransport.DestinationApproval{
						PlanID: plan.ID, ApprovalToken: previewed.ApprovalToken,
						Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
					},
				})
				completed <- runResult{run: run, err: runErr}
			}()
			if testCase.ambiguousWrite {
				select {
				case <-fixture.deleteStarted:
					cancel()
				case <-time.After(2 * time.Second):
					t.Fatal("GitLab timeout fixture never observed its one destination request")
				}
			}
			result := <-completed

			fixture.assertSourceReads(t, 1)
			fixture.assertDeleteRequests(t, []gitLabTransportRequest{{Method: fixture.deleteMethod, Path: "/groups/group-one", Body: ""}})
			if testCase.wantSuccess {
				if result.err != nil {
					t.Fatalf("acknowledged GitLab managed destination run: %v", result.err)
				}
				if result.run.Status != "completed" || result.run.RecordsRead != 1 || result.run.RecordsLoaded != 1 || len(result.run.Checkpoint) == 0 {
					t.Fatalf("acknowledged GitLab managed destination run = %#v, want one staged/applied record and an acknowledged checkpoint", result.run)
				}
				return
			}
			if result.err == nil {
				t.Fatalf("GitLab destination %q unexpectedly completed after one terminal/ambiguous request: %#v", testCase.name, result.run)
			}
			if len(result.run.Checkpoint) != 0 {
				t.Fatalf("GitLab destination %q checkpoint = %#v, want no advancement after terminal or ambiguous delivery", testCase.name, result.run.Checkpoint)
			}

			// The consumed single-attempt authorization is revoked after any
			// failed destination delivery. A later scheduler/user retry of the
			// same saved plan must stop before a source or GitLab request; another
			// provider attempt requires a new plan and approval.
			fixture.clearCapturedRequests()
			_, retryErr := application.RunETL(context.Background(), app.RunETLRequest{
				Connection: connection.Name,
				Stream:     fixture.stream,
				BatchSize:  1,
				DestinationApproval: synctransport.DestinationApproval{
					PlanID: plan.ID, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
				},
			})
			if retryErr == nil || !strings.Contains(retryErr.Error(), "revoked") {
				t.Fatalf("GitLab destination %q replay error = %v, want revoked single-attempt authorization before I/O", testCase.name, retryErr)
			}
			fixture.assertNoCapturedRequests(t)
		})
	}
}

type gitLabTransportFixture struct {
	t                         *testing.T
	stream                    string
	sourceOperation           string
	sourceMethod              string
	sourcePath                string
	deleteAction              string
	deleteMethod              string
	deletePathTemplate        string
	token                     string
	badRead                   bool
	sourcePages               [][]gitLabTransportGroup
	deleteStatus              int
	blockDeleteUntilCancelled bool
	deleteStarted             chan struct{}
	server                    *httptest.Server
	mu                        sync.Mutex
	readPaths                 []string
	deletedGroupIDs           []string
	deleteRequests            []gitLabTransportRequest
}

type gitLabTransportGroup struct {
	ID   int    `json:"id"`
	Path string `json:"path"`
}

type gitLabTransportRequest struct {
	Method string
	Path   string
	Body   string
}

func newGitLabTransportFixture(t *testing.T, badRead bool) *gitLabTransportFixture {
	t.Helper()
	const defsRoot = "../connectors/defs"
	const deleteSourceOperation = "deleteApiV4GroupsId"

	bundle, err := engine.Load(os.DirFS(defsRoot), "gitlab")
	if err != nil {
		t.Fatalf("load GitLab definition bundle: %v", err)
	}
	if bundle.EnabledContract == nil {
		t.Fatal("GitLab enabled connector contract is absent")
	}

	fixture := &gitLabTransportFixture{
		t: t, stream: "groups", token: "gitlab-local-test-token", badRead: badRead,
		sourcePages: [][]gitLabTransportGroup{
			{{ID: 1, Path: "group-one"}},
			{{ID: 2, Path: "group-two"}},
		},
		deleteStatus:  http.StatusAccepted,
		deleteStarted: make(chan struct{}, 1),
	}
	for _, lane := range bundle.EnabledContract.Lanes {
		if lane.Name != "sync_transport" || lane.State != connectors.EnabledLaneImplemented || lane.Transport == nil {
			continue
		}
		modes := append([]string(nil), lane.Transport.Modes...)
		sort.Strings(modes)
		if got := strings.Join(modes, ","); got != "full_append,full_overwrite" {
			t.Fatalf("GitLab runtime full-snapshot modes = %q, want the closed local execution policy", got)
		}
		for _, evidence := range lane.Transport.Streams {
			if evidence.Stream == fixture.stream {
				fixture.sourceOperation = evidence.SourceOperation
			}
		}
	}
	if fixture.sourceOperation == "" {
		t.Fatalf("GitLab enabled source transport does not declare stream %q", fixture.stream)
	}

	for _, action := range bundle.Writes {
		if action.Method == http.MethodDelete && action.Path == "/groups/{{ record.id }}" && len(action.SuccessStatuses) == 1 && action.SuccessStatuses[0] == http.StatusAccepted {
			fixture.deleteAction = action.Name
			fixture.deleteMethod = action.Method
			fixture.deletePathTemplate = action.Path
			break
		}
	}
	if fixture.deleteAction == "" {
		t.Fatal("GitLab source-bound group deletion action is absent")
	}

	lockRaw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", "sources", "gitlab-operation-source-lock.json"))
	if err != nil {
		t.Fatalf("read GitLab immutable source lock: %v", err)
	}
	var lock gitLabTransportSourceLock
	if err := json.Unmarshal(lockRaw, &lock); err != nil {
		t.Fatalf("decode GitLab immutable source lock: %v", err)
	}
	for _, operation := range lock.REST.Operations {
		switch operation.ID {
		case fixture.sourceOperation:
			fixture.sourceMethod = operation.Method
			fixture.sourcePath = strings.TrimPrefix(operation.Path, "/api/v4")
			if fixture.sourceMethod != http.MethodGet || fixture.sourcePath != "/groups" || !operation.hasJSONSuccess(http.StatusOK) {
				t.Fatalf("GitLab groups source lock operation = %+v, want cited GET /api/v4/groups JSON 200", operation)
			}
		case deleteSourceOperation:
			if operation.Method != fixture.deleteMethod || operation.Path != "/api/v4/groups/{id}" || !operation.hasStatus(http.StatusAccepted) {
				t.Fatalf("GitLab group deletion source lock operation = %+v, want cited DELETE /api/v4/groups/{id} 202", operation)
			}
		}
	}
	if fixture.sourceMethod == "" {
		t.Fatalf("GitLab source lock does not retain %q", fixture.sourceOperation)
	}

	fixture.server = httptest.NewServer(http.HandlerFunc(fixture.serveHTTP))
	t.Cleanup(fixture.server.Close)
	return fixture
}

func (f *gitLabTransportFixture) setupProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("PM_TEST_GITLAB_TRANSPORT_TOKEN", f.token)
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "gitlab-local",
		"--connector", "gitlab",
		"--from-env", "access_token=PM_TEST_GITLAB_TRANSPORT_TOKEN",
		"--config", "base_url=" + f.server.URL,
		"--root", root, "--json",
	})
	runCLI(t, []string{
		"credentials", "add", "warehouse-local",
		"--connector", "warehouse",
		"--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"),
		"--root", root, "--json",
	})
	runCLI(t, []string{
		"connections", "create", "gitlab_groups_to_warehouse",
		"--source", "gitlab:gitlab-local",
		"--destination", "warehouse:warehouse-local",
		"--stream", f.stream,
		"--sync-mode", "full_append",
		"--table", "gitlab_groups",
		"--root", root, "--json",
	})
	return root
}

func (f *gitLabTransportFixture) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if got := r.Header.Get("Authorization"); got != "Bearer "+f.token {
		f.failf("authorization = %q, want bearer fixture token", got)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch {
	case r.Method == f.sourceMethod && r.URL.Path == f.sourcePath:
		f.serveSourcePage(w, r)
	case r.Method == f.deleteMethod && strings.HasPrefix(r.URL.Path, "/groups/"):
		f.serveDelete(w, r)
	default:
		f.failf("unexpected fixture request %s %s", r.Method, r.URL.String())
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *gitLabTransportFixture) serveSourcePage(w http.ResponseWriter, r *http.Request) {
	if got := r.URL.Query().Get("per_page"); got != "50" {
		f.failf("groups per_page = %q, want declared stream value 50", got)
	}
	page := r.URL.Query().Get("page")
	pageIndex := 0
	if page != "" {
		if page != "2" {
			f.failf("groups page = %q, want source first page or cited next page 2", page)
		}
		pageIndex = 1
	}
	f.mu.Lock()
	f.readPaths = append(f.readPaths, r.URL.RequestURI())
	f.mu.Unlock()
	if f.badRead {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"fixture source not found"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if pageIndex >= len(f.sourcePages) {
		f.failf("groups page index = %d, want a configured source fixture page", pageIndex)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if pageIndex+1 < len(f.sourcePages) {
		w.Header().Set("Link", fmt.Sprintf("<http://%s%s?page=2&per_page=50>; rel=\"next\"", r.Host, f.sourcePath))
	}
	if err := json.NewEncoder(w).Encode(f.sourcePages[pageIndex]); err != nil {
		f.failf("encode groups source fixture page: %v", err)
	}
}

func (f *gitLabTransportFixture) serveDelete(w http.ResponseWriter, r *http.Request) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		f.failf("read delete body: %v", err)
	}
	if len(raw) != 0 {
		f.failf("group delete body = %q, want no body for the source-cited DELETE", raw)
	}
	groupID := strings.TrimPrefix(r.URL.Path, "/groups/")
	if groupID == "" || strings.Contains(groupID, "/") || r.URL.RawQuery != "" {
		f.failf("group delete path/query = %s?%s, want one closed group id and no query", r.URL.Path, r.URL.RawQuery)
	}
	f.mu.Lock()
	f.deletedGroupIDs = append(f.deletedGroupIDs, groupID)
	f.deleteRequests = append(f.deleteRequests, gitLabTransportRequest{Method: r.Method, Path: r.URL.Path, Body: string(raw)})
	f.mu.Unlock()
	if f.blockDeleteUntilCancelled {
		select {
		case f.deleteStarted <- struct{}{}:
		default:
		}
		<-r.Context().Done()
		return
	}
	w.WriteHeader(f.deleteStatus)
}

func (f *gitLabTransportFixture) setSourceGroups(groups []gitLabTransportGroup) {
	f.sourcePages = [][]gitLabTransportGroup{append([]gitLabTransportGroup(nil), groups...)}
}

func (f *gitLabTransportFixture) clearCapturedRequests() {
	f.mu.Lock()
	f.readPaths = nil
	f.deletedGroupIDs = nil
	f.deleteRequests = nil
	f.mu.Unlock()
}

func (f *gitLabTransportFixture) assertNoCapturedRequests(t *testing.T) {
	t.Helper()
	f.mu.Lock()
	reads := append([]string(nil), f.readPaths...)
	writes := append([]gitLabTransportRequest(nil), f.deleteRequests...)
	f.mu.Unlock()
	if len(reads) != 0 || len(writes) != 0 {
		t.Fatalf("GitLab destination admission reached HTTP reads=%v writes=%v, want no provider request before preflight accepts", reads, writes)
	}
}

func (f *gitLabTransportFixture) assertSourceReads(t *testing.T, want int) {
	t.Helper()
	f.mu.Lock()
	got := append([]string(nil), f.readPaths...)
	f.mu.Unlock()
	if len(got) != want {
		t.Fatalf("GitLab source requests = %v, want %d requests", got, want)
	}
	if want == 2 && (got[0] != f.sourcePath+"?per_page=50" || got[1] != f.sourcePath+"?page=2&per_page=50") {
		t.Fatalf("GitLab source request paths = %v, want exact first and link-derived second pages", got)
	}
}

func (f *gitLabTransportFixture) assertDeletes(t *testing.T, want []string) {
	t.Helper()
	f.mu.Lock()
	got := append([]string(nil), f.deletedGroupIDs...)
	f.mu.Unlock()
	sort.Strings(got)
	want = append([]string(nil), want...)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("GitLab delete group ids = %v, want %v", got, want)
	}
}

func (f *gitLabTransportFixture) assertDeleteRequests(t *testing.T, want []gitLabTransportRequest) {
	t.Helper()
	f.mu.Lock()
	got := append([]gitLabTransportRequest(nil), f.deleteRequests...)
	f.mu.Unlock()
	if len(got) != len(want) {
		t.Fatalf("GitLab destination request count = %d (%#v), want %d (%#v)", len(got), got, len(want), want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("GitLab destination request %d = %#v, want %#v", index, got[index], want[index])
		}
	}
}

func (f *gitLabTransportFixture) failf(format string, args ...any) {
	// httptest handlers cannot safely call Fatalf. Errorf marks the test while
	// the response still closes the unexpected request path.
	f.t.Errorf(format, args...)
}

func decodeGitLabReversePlanID(t *testing.T, raw string) string {
	t.Helper()
	var envelope struct {
		Kind string `json:"kind"`
		Plan struct {
			ID string `json:"id"`
		} `json:"plan"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		t.Fatalf("decode GitLab reverse plan: %v\n%s", err, raw)
	}
	if envelope.Kind != "ReversePlan" || envelope.Plan.ID == "" {
		t.Fatalf("GitLab reverse plan = %+v, want persisted plan id", envelope)
	}
	return envelope.Plan.ID
}

type gitLabTransportSourceLock struct {
	REST struct {
		Operations []gitLabTransportSourceOperation `json:"operations"`
	} `json:"rest"`
}

type gitLabTransportSourceOperation struct {
	ID              string `json:"id"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	SourceOperation struct {
		Responses map[string]json.RawMessage `json:"responses"`
	} `json:"source_operation"`
}

func (o gitLabTransportSourceOperation) hasStatus(status int) bool {
	_, ok := o.SourceOperation.Responses[fmt.Sprintf("%d", status)]
	return ok
}

func (o gitLabTransportSourceOperation) hasJSONSuccess(status int) bool {
	raw, ok := o.SourceOperation.Responses[fmt.Sprintf("%d", status)]
	return ok && bytes.Contains(raw, []byte(`"application/json"`))
}
