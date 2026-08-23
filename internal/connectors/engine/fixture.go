package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

type readFixturePage struct {
	file      string
	Request   readFixtureRequest  `json:"request"`
	ReadQuery map[string]string   `json:"read_query,omitempty"`
	Response  readFixtureResponse `json:"response"`
}

type readFixtureRequest struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Query  map[string]string `json:"query"`
}

type readFixtureResponse struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

type readFixtureTransport struct {
	mu    sync.Mutex
	pages []readFixturePage
	next  int
}

func (t *readFixtureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.next >= len(t.pages) {
		return fixtureHTTPResponse(req, http.StatusNotFound, []byte(`{"error":"fixture pages exhausted"}`)), nil
	}
	page := t.pages[t.next]
	if !readFixtureRequestMatches(req, page.Request) {
		body, err := json.Marshal(map[string]string{
			"error": fmt.Sprintf("request does not match fixture %s", page.file),
		})
		if err != nil {
			return nil, fmt.Errorf("encode fixture mismatch: %w", err)
		}
		return fixtureHTTPResponse(req, http.StatusBadRequest, body), nil
	}
	t.next++

	status := page.Response.Status
	if status == 0 {
		status = http.StatusOK
	}
	body := page.Response.Body
	if len(body) == 0 {
		body = json.RawMessage(`{}`)
	}
	return fixtureHTTPResponse(req, status, body), nil
}

func (t *readFixtureTransport) remaining() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.pages) - t.next
}

func fixtureHTTPResponse(req *http.Request, status int, body []byte) *http.Response {
	return &http.Response{
		Status:        fmt.Sprintf("%d %s", status, http.StatusText(status)),
		StatusCode:    status,
		Header:        http.Header{"Content-Type": []string{"application/json"}},
		Body:          io.NopCloser(bytes.NewReader(body)),
		ContentLength: int64(len(body)),
		Request:       req,
	}
}

func readFixtureRequestMatches(req *http.Request, want readFixtureRequest) bool {
	method := want.Method
	if method == "" {
		method = http.MethodGet
	}
	if !strings.EqualFold(req.Method, method) {
		return false
	}
	if want.Path != "" && req.URL.Path != want.Path {
		return false
	}
	query := req.URL.Query()
	for key, value := range want.Query {
		if query.Get(key) != value {
			return false
		}
	}
	return true
}

func loadReadFixturePages(fixtures fs.FS, stream string) ([]readFixturePage, error) {
	if fixtures == nil {
		return nil, fmt.Errorf("engine: stream %q has no fixture filesystem", stream)
	}
	dir := path.Join("streams", stream)
	entries, err := fs.ReadDir(fixtures, dir)
	if err != nil {
		return nil, fmt.Errorf("engine: read fixture directory %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	pages := make([]readFixturePage, 0, len(names))
	for _, name := range names {
		fixturePath := path.Join(dir, name)
		raw, err := fs.ReadFile(fixtures, fixturePath)
		if err != nil {
			return nil, fmt.Errorf("engine: read fixture %s: %w", fixturePath, err)
		}
		var page readFixturePage
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, fmt.Errorf("engine: parse fixture %s: %w", fixturePath, err)
		}
		page.file = fixturePath
		pages = append(pages, page)
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("engine: stream %q has no fixture pages", stream)
	}
	return pages, nil
}

// ReadFixture replays one stream from its bundle pages without external I/O.
func (c *Connector) ReadFixture(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream, err := findStream(c.bundle, req.Stream)
	if err != nil {
		return err
	}
	pages, err := loadReadFixturePages(c.bundle.Fixtures, stream.Name)
	if err != nil {
		return err
	}

	first := pages[0]
	if first.Request.Path != "" {
		stream.Path = first.Request.Path
	}
	if first.Request.Method != "" {
		stream.Method = first.Request.Method
	}
	stream.Query = nil
	stream.Incremental = nil
	req.Query = make(map[string]string, len(first.Request.Query)+len(first.ReadQuery))
	for key, value := range first.Request.Query {
		req.Query[key] = value
	}
	for key, value := range first.ReadQuery {
		req.Query[key] = value
	}

	transport := &readFixtureTransport{pages: pages}
	runtime := &Runtime{
		Requester: &connsdk.Requester{
			BaseURL:       "https://fixture.invalid",
			Client:        &http.Client{Transport: transport},
			RetryStatuses: map[int]bool{},
			Sleep: func(ctx context.Context, _ time.Duration) error {
				return ctx.Err()
			},
		},
		Bundle: &c.bundle,
		Config: req.Config,
	}
	if err := readDeclarative(ctx, c.bundle, stream, req, runtime, c.hooks, emit, false); err != nil {
		return err
	}
	if remaining := transport.remaining(); remaining != 0 {
		return fmt.Errorf("engine: stream %q left %d fixture page(s) unread", stream.Name, remaining)
	}
	return nil
}
