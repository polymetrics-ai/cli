package certify

import (
	"io"
	"net/http"
	"sync"
)

const defaultObservedBodyBytes = 1 << 20

// ObservedBody is a bounded, in-memory view of one HTTP body. Bytes contains
// at most the limit supplied to NewObservedTransport. Complete proves the
// caller consumed the body through EOF; a closed early body is incomplete even
// when its captured prefix fits the bound.
type ObservedBody struct {
	Bytes         []byte
	OriginalBytes int
	Truncated     bool
	Complete      bool
}

// ObservedHTTPRequest is the request half of one observed transport exchange.
// It retains exact, pre-serialization values only in process memory. The proof
// boundary must fingerprint them before an artifact can be written.
type ObservedHTTPRequest struct {
	Method  string
	Target  string
	Headers http.Header
	Body    ObservedBody
}

// ObservedHTTPResponse is the response half of one observed transport exchange.
type ObservedHTTPResponse struct {
	Status  int
	Headers http.Header
	Body    ObservedBody
}

// ObservedHTTPExchange is one request and its corresponding response observed
// by ObservedTransport. An exchange is appended only after the caller closes
// the response body, preventing a partial or abandoned read from looking like
// a complete proof.
type ObservedHTTPExchange struct {
	Request  ObservedHTTPRequest
	Response ObservedHTTPResponse
}

// ObservedTransport wraps an HTTP transport and collects exact exchanges in
// memory. It never sanitizes, logs, serializes, or writes captured values: the
// certification proof boundary owns that later fingerprint-first conversion.
// It is safe to use concurrently.
type ObservedTransport struct {
	underlying http.RoundTripper
	maxBody    int

	mu        sync.Mutex
	exchanges []ObservedHTTPExchange
}

// NewObservedTransport builds an in-memory observer around underlying. A nil
// transport uses http.DefaultTransport. Non-positive limits use the proof
// boundary's one-megabyte default rather than permitting unbounded capture.
func NewObservedTransport(underlying http.RoundTripper, maxBodyBytes int) *ObservedTransport {
	if underlying == nil {
		underlying = http.DefaultTransport
	}
	if maxBodyBytes <= 0 {
		maxBodyBytes = defaultObservedBodyBytes
	}
	return &ObservedTransport{underlying: underlying, maxBody: maxBodyBytes}
}

// RoundTrip observes the request after authentication has prepared it and the
// response as its consumer reads it. The wrapped bodies preserve the bytes the
// underlying transport and caller see; only the retained copy is bounded.
func (t *ObservedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	requestBody := newObservedBodyCapture(t.maxBody)
	if req.Body == nil {
		requestBody.complete = true
	} else {
		req.Body = &observedReadCloser{ReadCloser: req.Body, capture: requestBody}
	}

	request := ObservedHTTPRequest{
		Method:  req.Method,
		Target:  req.URL.String(),
		Headers: req.Header.Clone(),
	}
	resp, err := t.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	exchange := ObservedHTTPExchange{
		Request: request,
		Response: ObservedHTTPResponse{
			Status:  resp.StatusCode,
			Headers: resp.Header.Clone(),
		},
	}
	responseBody := newObservedBodyCapture(t.maxBody)
	if resp.Body == nil {
		responseBody.complete = true
		exchange.Request.Body = requestBody.snapshot()
		exchange.Response.Body = responseBody.snapshot()
		t.appendExchange(exchange)
		return resp, nil
	}
	resp.Body = &observedReadCloser{
		ReadCloser: resp.Body,
		capture:    responseBody,
		onClose: func() {
			exchange.Request.Body = requestBody.snapshot()
			exchange.Response.Body = responseBody.snapshot()
			t.appendExchange(exchange)
		},
	}
	return resp, nil
}

// Exchanges returns a defensive snapshot. Callers cannot mutate an observed
// exchange after it has been accepted into the in-memory transcript.
func (t *ObservedTransport) Exchanges() []ObservedHTTPExchange {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]ObservedHTTPExchange, len(t.exchanges))
	for i, exchange := range t.exchanges {
		out[i] = cloneObservedHTTPExchange(exchange)
	}
	return out
}

func (t *ObservedTransport) appendExchange(exchange ObservedHTTPExchange) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.exchanges = append(t.exchanges, cloneObservedHTTPExchange(exchange))
}

func cloneObservedHTTPExchange(exchange ObservedHTTPExchange) ObservedHTTPExchange {
	exchange.Request.Headers = exchange.Request.Headers.Clone()
	exchange.Response.Headers = exchange.Response.Headers.Clone()
	exchange.Request.Body.Bytes = append([]byte(nil), exchange.Request.Body.Bytes...)
	exchange.Response.Body.Bytes = append([]byte(nil), exchange.Response.Body.Bytes...)
	return exchange
}

type observedBodyCapture struct {
	mu       sync.Mutex
	bytes    []byte
	seen     int
	max      int
	complete bool
}

func newObservedBodyCapture(max int) *observedBodyCapture {
	return &observedBodyCapture{max: max}
}

func (c *observedBodyCapture) write(p []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.seen += len(p)
	remaining := c.max - len(c.bytes)
	if remaining <= 0 {
		return
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	c.bytes = append(c.bytes, p...)
}

func (c *observedBodyCapture) markComplete() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.complete = true
}

func (c *observedBodyCapture) snapshot() ObservedBody {
	c.mu.Lock()
	defer c.mu.Unlock()
	return ObservedBody{
		Bytes:         append([]byte(nil), c.bytes...),
		OriginalBytes: c.seen,
		Truncated:     c.seen > len(c.bytes),
		Complete:      c.complete,
	}
}

type observedReadCloser struct {
	io.ReadCloser
	capture *observedBodyCapture
	onClose func()
	once    sync.Once
}

func (r *observedReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if n > 0 {
		r.capture.write(p[:n])
	}
	if err == io.EOF {
		r.capture.markComplete()
	}
	return n, err
}

func (r *observedReadCloser) Close() error {
	err := r.ReadCloser.Close()
	r.once.Do(func() {
		if r.onClose != nil {
			r.onClose()
		}
	})
	return err
}
