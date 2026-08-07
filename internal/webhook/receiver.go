package webhook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	// ErrReceiptBackpressure rejects a new event when durable receipt capacity
	// is exhausted. Existing durable duplicates remain successful.
	ErrReceiptBackpressure      = errors.New("webhook receipt capacity is exhausted")
	ErrReceiptHandoffInProgress = errors.New("webhook receipt handoff is in progress")
	ErrReceiptNotFound          = errors.New("webhook receipt was not found")
	// ErrInvalidEventIdentity rejects a verifier result without a documented,
	// stable provider event identity. The receiver never invents a body hash.
	ErrInvalidEventIdentity = errors.New("verified event identity is invalid")
)

// VerifiedEvent contains only provider-declared receipt identity. Provider
// lanes define how their verifier extracts it after validating the raw body.
type VerifiedEvent struct {
	ID string
}

// Receipt is handed to durable storage before a webhook acknowledgement. Raw
// bytes must be encrypted/minimized by the store and must never be logged.
type Receipt struct {
	Event      VerifiedEvent
	RawBody    []byte
	ReceivedAt time.Time
}

// ReceiptInsertResult distinguishes new, already-durable, and rejected work.
type ReceiptInsertResult string

const (
	ReceiptInsertNew       ReceiptInsertResult = "new"
	ReceiptInsertDuplicate ReceiptInsertResult = "duplicate"
	ReceiptInsertRejected  ReceiptInsertResult = "rejected"
)

type ReceiptConsumeResult string

const (
	ReceiptConsumeCompleted ReceiptConsumeResult = "completed"
	ReceiptConsumeDuplicate ReceiptConsumeResult = "duplicate"
	ReceiptConsumeRejected  ReceiptConsumeResult = "rejected"
)

// Verifier validates the original request bytes before any decoding or
// persistence. It must return the provider's stable event identity.
type Verifier interface {
	Verify(ctx context.Context, rawBody []byte, headers http.Header) (VerifiedEvent, error)
}

// ReceiptStore atomically records a receipt and deduplicates by the verified
// provider event identity within a subscription. It owns durable queue bounds.
type ReceiptStore interface {
	Insert(ctx context.Context, receipt Receipt) (ReceiptInsertResult, error)
}

type DurableReceiptHandoff func(context.Context, Receipt) error

type DurableReceiptStore interface {
	ReceiptStore
	Consume(context.Context, string, DurableReceiptHandoff) (ReceiptConsumeResult, error)
	ConsumeNext(context.Context, DurableReceiptHandoff) (ReceiptConsumeResult, error)
}

// ReceiverConfig requires explicit limits. There are no process-wide magic
// limits because providers must declare their own body and timeout evidence.
type ReceiverConfig struct {
	Method         string
	Path           string
	MaxBodyBytes   int64
	MaxInFlight    int
	RequestTimeout time.Duration
	Verifier       Verifier
	Store          ReceiptStore
}

// Receiver is a bounded HTTP ingress handler. It has no provider-specific
// parser, signer, registration call, tunnel client, or background poller.
type Receiver struct {
	config   ReceiverConfig
	inFlight chan struct{}
}

// LoopbackServer is a running external-tunnel receiver. It exposes only the
// listener address; callback URLs remain operator-supplied and never enter it.
type LoopbackServer struct {
	server   *http.Server
	listener net.Listener
	done     chan struct{}
	serveErr error
}

// NewReceiver validates the provider-declared route and explicit bounds.
func NewReceiver(config ReceiverConfig) (*Receiver, error) {
	if strings.TrimSpace(config.Method) == "" || strings.TrimSpace(config.Path) == "" || !strings.HasPrefix(config.Path, "/") {
		return nil, errors.New("receiver method and absolute path are required")
	}
	if config.MaxBodyBytes <= 0 || config.MaxInFlight <= 0 || config.RequestTimeout <= 0 {
		return nil, errors.New("receiver body, in-flight, and timeout bounds are required")
	}
	if config.Verifier == nil || config.Store == nil {
		return nil, errors.New("receiver verifier and receipt store are required")
	}
	return &Receiver{config: config, inFlight: make(chan struct{}, config.MaxInFlight)}, nil
}

// StartLoopback starts a receiver only for external_tunnel exposure. It does
// not invoke Tailscale or another binary: the operator's named tunnel connects
// to the returned 127.0.0.1 listener independently.
func (r *Receiver) StartLoopback(exposure Exposure, port int) (*LoopbackServer, error) {
	if r == nil || exposure.Mode != ExposureModeExternalTunnel || exposure.ListenerScope != ListenerScopeLoopback {
		return nil, errors.New("only external tunnel exposure can start a loopback receiver")
	}
	if port < 0 || port > 65535 {
		return nil, errors.New("loopback receiver port is invalid")
	}
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return nil, fmt.Errorf("listen on loopback receiver: %w", err)
	}
	server := &http.Server{
		Handler:           r,
		ReadHeaderTimeout: r.config.RequestTimeout,
		ReadTimeout:       r.config.RequestTimeout,
		WriteTimeout:      r.config.RequestTimeout,
		IdleTimeout:       r.config.RequestTimeout,
	}
	running := &LoopbackServer{server: server, listener: listener, done: make(chan struct{})}
	go func() {
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		running.serveErr = err
		close(running.done)
	}()
	return running, nil
}

// Address returns the loopback listener address for an already-running
// receiver. It is never a public callback URL.
func (s *LoopbackServer) Address() string {
	if s == nil || s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

// Shutdown waits for the owned serving goroutine to exit.
func (s *LoopbackServer) Shutdown(ctx context.Context) error {
	if s == nil || s.server == nil {
		return nil
	}
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	<-s.done
	return s.serveErr
}

// ServeHTTP verifies bounded raw bytes, stores a receipt durably, and only
// then acknowledges the provider. It intentionally makes no ordering claim.
func (r *Receiver) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.Method != r.config.Method || request.URL.Path != r.config.Path {
		reject(w, http.StatusNotFound)
		return
	}
	select {
	case r.inFlight <- struct{}{}:
		defer func() { <-r.inFlight }()
	default:
		reject(w, http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(request.Context(), r.config.RequestTimeout)
	defer cancel()
	body, err := io.ReadAll(http.MaxBytesReader(w, request.Body, r.config.MaxBodyBytes))
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			reject(w, http.StatusRequestEntityTooLarge)
			return
		}
		reject(w, http.StatusBadRequest)
		return
	}

	event, err := r.config.Verifier.Verify(ctx, body, request.Header)
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		reject(w, http.StatusServiceUnavailable)
		return
	}
	if err != nil || !ValidEventIdentity(event.ID) {
		reject(w, http.StatusUnauthorized)
		return
	}
	result, err := r.config.Store.Insert(ctx, Receipt{
		Event:      event,
		RawBody:    append([]byte(nil), body...),
		ReceivedAt: time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, ErrReceiptBackpressure) {
			reject(w, http.StatusServiceUnavailable)
			return
		}
		reject(w, http.StatusServiceUnavailable)
		return
	}
	if result != ReceiptInsertNew && result != ReceiptInsertDuplicate {
		reject(w, http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func reject(w http.ResponseWriter, status int) {
	http.Error(w, "webhook rejected", status)
}

func ValidEventIdentity(value string) bool {
	if strings.TrimSpace(value) == "" || len(value) > 512 {
		return false
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return false
		}
	}
	return true
}

var _ http.Handler = (*Receiver)(nil)
var _ = fmt.Sprintf
