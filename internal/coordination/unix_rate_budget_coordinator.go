package coordination

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	unixRateBudgetProtocolVersion = 1
	maxUnixRateBudgetMessage      = 32 << 10
	unixRateBudgetRequestTimeout  = 5 * time.Second
)

// UnixRateBudgetCoordinatorOptions configures a single run-local owner.
type UnixRateBudgetCoordinatorOptions struct {
	MaxInFlight int
	LeaseTTL    time.Duration
}

// UnixRateBudgetCoordinatorOwner owns one short-lived private UDS endpoint.
// It removes only its known socket and private run directory on Close.
type UnixRateBudgetCoordinatorOwner struct {
	mu          sync.Mutex
	listener    *net.UnixListener
	coordinator *RateBudgetCoordinator
	runDir      string
	socketPath  string
	epoch       string
	closed      bool
	connections map[net.Conn]struct{}
	wg          sync.WaitGroup
}

// UnixRateBudgetCoordinatorClient is an in-memory capability for the owner
// created in the same run; endpoint and epoch never enter operator output.
type UnixRateBudgetCoordinatorClient struct {
	socketPath string
	epoch      string
}

var _ connsdk.BudgetCoordinator = (*UnixRateBudgetCoordinatorClient)(nil)

type sharedCoordinatorErrorReason string

const (
	sharedCoordinatorUnavailable   sharedCoordinatorErrorReason = "unavailable"
	sharedCoordinatorEpochMismatch sharedCoordinatorErrorReason = "epoch_mismatch"
)

// SharedCoordinatorError deliberately exposes only a closed safe class.
type SharedCoordinatorError struct{ reason sharedCoordinatorErrorReason }

func (e *SharedCoordinatorError) Error() string {
	if e != nil && e.reason == sharedCoordinatorEpochMismatch {
		return "shared rate-budget coordinator epoch mismatch"
	}
	return "shared rate-budget coordinator unavailable"
}

func IsSharedCoordinatorUnavailable(err error) bool {
	var coordinatorErr *SharedCoordinatorError
	return errors.As(err, &coordinatorErr) && coordinatorErr.reason == sharedCoordinatorUnavailable
}

func IsSharedCoordinatorEpochMismatch(err error) bool {
	var coordinatorErr *SharedCoordinatorError
	return errors.As(err, &coordinatorErr) && coordinatorErr.reason == sharedCoordinatorEpochMismatch
}

// StartUnixRateBudgetCoordinator creates a 0700 run directory and 0600 socket,
// then validates the capability with a versioned readiness exchange.
func StartUnixRateBudgetCoordinator(ctx context.Context, options UnixRateBudgetCoordinatorOptions) (*UnixRateBudgetCoordinatorOwner, *UnixRateBudgetCoordinatorClient, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if options.MaxInFlight <= 0 {
		options.MaxInFlight = 1
	}
	if options.LeaseTTL <= 0 {
		options.LeaseTTL = 15 * time.Second
	}
	runDir, err := os.MkdirTemp("/tmp", "pmrb-")
	if err != nil {
		return nil, nil, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	cleanupRunDir := true
	defer func() {
		if cleanupRunDir {
			_ = os.Remove(runDir)
		}
	}()
	if err := os.Chmod(runDir, 0o700); err != nil {
		return nil, nil, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	socketPath := filepath.Join(runDir, "s")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socketPath, Net: "unix"})
	if err != nil {
		return nil, nil, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		return nil, nil, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	epoch, err := newUnixRateBudgetEpoch()
	if err != nil {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		return nil, nil, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	owner := &UnixRateBudgetCoordinatorOwner{
		listener:    listener,
		coordinator: NewRateBudgetCoordinator(nil, RateBudgetCoordinatorOptions(options)),
		runDir:      runDir,
		socketPath:  socketPath,
		epoch:       epoch,
		connections: make(map[net.Conn]struct{}),
	}
	owner.wg.Add(1)
	go owner.accept()
	client := &UnixRateBudgetCoordinatorClient{socketPath: socketPath, epoch: epoch}
	if err := client.Ready(ctx); err != nil {
		_ = owner.Close()
		return nil, nil, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	cleanupRunDir = false
	return owner, client, nil
}

func (o *UnixRateBudgetCoordinatorOwner) Close() error {
	if o == nil {
		return nil
	}
	o.mu.Lock()
	if o.closed {
		o.mu.Unlock()
		return nil
	}
	o.closed = true
	listener := o.listener
	connections := make([]net.Conn, 0, len(o.connections))
	for conn := range o.connections {
		connections = append(connections, conn)
	}
	o.mu.Unlock()
	if listener != nil {
		_ = listener.Close()
	}
	for _, conn := range connections {
		_ = conn.Close()
	}
	o.wg.Wait()
	var cleanupFailed bool
	if err := os.Remove(o.socketPath); err != nil && !os.IsNotExist(err) {
		cleanupFailed = true
	}
	if err := os.Remove(o.runDir); err != nil && !os.IsNotExist(err) {
		cleanupFailed = true
	}
	if cleanupFailed {
		return errors.New("shared rate-budget coordinator cleanup failed")
	}
	return nil
}

func (o *UnixRateBudgetCoordinatorOwner) accept() {
	defer o.wg.Done()
	for {
		conn, err := o.listener.AcceptUnix()
		if err != nil {
			o.mu.Lock()
			closed := o.closed
			o.mu.Unlock()
			if closed {
				return
			}
			continue
		}
		o.mu.Lock()
		if o.closed {
			o.mu.Unlock()
			_ = conn.Close()
			return
		}
		o.connections[conn] = struct{}{}
		o.wg.Add(1)
		o.mu.Unlock()
		go o.handle(conn)
	}
}

func (o *UnixRateBudgetCoordinatorOwner) handle(conn *net.UnixConn) {
	defer o.wg.Done()
	defer func() { _ = conn.Close() }()
	defer func() {
		o.mu.Lock()
		delete(o.connections, conn)
		o.mu.Unlock()
	}()
	_ = conn.SetDeadline(time.Now().Add(unixRateBudgetRequestTimeout))
	request, err := readUnixRateBudgetRequest(conn)
	if err != nil {
		return
	}
	_ = writeUnixRateBudgetResponse(conn, o.handleRequest(request))
}

func (o *UnixRateBudgetCoordinatorOwner) handleRequest(request unixRateBudgetRequest) unixRateBudgetResponse {
	response := unixRateBudgetResponse{Version: unixRateBudgetProtocolVersion, Kind: request.Kind}
	if request.Version != unixRateBudgetProtocolVersion {
		response.Error = unixRateBudgetWireError{Code: unixRateBudgetErrorUnavailable}
		return response
	}
	if request.Epoch != o.epoch {
		response.Error = unixRateBudgetWireError{Code: unixRateBudgetErrorEpochMismatch}
		return response
	}
	switch request.Kind {
	case unixRateBudgetReady:
		response.Ready = true
	case unixRateBudgetDecide:
		decision, err := o.coordinator.Decide(context.Background(), request.Batch)
		if err != nil {
			response.Error = unixRateBudgetWireError{Code: unixRateBudgetErrorRefused}
			return response
		}
		response.Decision = decision
	case unixRateBudgetFinish:
		if err := o.coordinator.Finish(context.Background(), request.Lease, request.Observation); err != nil {
			response.Error = unixRateBudgetWireError{Code: unixRateBudgetErrorUnavailable}
		}
	default:
		response.Error = unixRateBudgetWireError{Code: unixRateBudgetErrorRefused}
	}
	return response
}

func (c *UnixRateBudgetCoordinatorClient) Ready(ctx context.Context) error {
	if c == nil || c.socketPath == "" || c.epoch == "" {
		return &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	response, err := c.exchange(ctx, unixRateBudgetRequest{Version: unixRateBudgetProtocolVersion, Kind: unixRateBudgetReady, Epoch: c.epoch})
	if err != nil {
		return err
	}
	if !response.Ready {
		return &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	return nil
}

func (c *UnixRateBudgetCoordinatorClient) Decide(ctx context.Context, batch connsdk.ReservationBatch) (connsdk.AdmissionDecision, error) {
	if c == nil || c.socketPath == "" || c.epoch == "" {
		return connsdk.AdmissionDecision{}, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	response, err := c.exchange(ctx, unixRateBudgetRequest{Version: unixRateBudgetProtocolVersion, Kind: unixRateBudgetDecide, Epoch: c.epoch, Batch: batch})
	if err != nil {
		return connsdk.AdmissionDecision{}, err
	}
	return response.Decision, nil
}

func (c *UnixRateBudgetCoordinatorClient) Finish(ctx context.Context, lease connsdk.RateBudgetLease, observation connsdk.CompletionObservation) error {
	if c == nil || c.socketPath == "" || c.epoch == "" {
		return &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	_, err := c.exchange(ctx, unixRateBudgetRequest{Version: unixRateBudgetProtocolVersion, Kind: unixRateBudgetFinish, Epoch: c.epoch, Lease: lease, Observation: observation})
	return err
}

// exchange connects all I/O to caller cancellation. The after-function wakes a
// blocked read or write; the final context check makes cancellation win even if
// a syntactically valid response arrives in the same scheduling window.
func (c *UnixRateBudgetCoordinatorClient) exchange(ctx context.Context, request unixRateBudgetRequest) (unixRateBudgetResponse, error) {
	if err := ctx.Err(); err != nil {
		return unixRateBudgetResponse{}, err
	}
	if c == nil || c.socketPath == "" || c.epoch == "" {
		return unixRateBudgetResponse{}, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		if ctx.Err() != nil {
			return unixRateBudgetResponse{}, ctx.Err()
		}
		return unixRateBudgetResponse{}, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	defer func() { _ = conn.Close() }()
	deadline := time.Now().Add(unixRateBudgetRequestTimeout)
	if callerDeadline, ok := ctx.Deadline(); ok && callerDeadline.Before(deadline) {
		deadline = callerDeadline
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return unixRateBudgetResponse{}, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	stop := context.AfterFunc(ctx, func() { _ = conn.SetDeadline(time.Now()) })
	defer stop()
	if err := ctx.Err(); err != nil {
		return unixRateBudgetResponse{}, err
	}
	if err := writeUnixRateBudgetRequest(conn, request); err != nil {
		if ctx.Err() != nil {
			return unixRateBudgetResponse{}, ctx.Err()
		}
		return unixRateBudgetResponse{}, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	response, err := readUnixRateBudgetResponse(conn)
	if err != nil {
		if ctx.Err() != nil {
			return unixRateBudgetResponse{}, ctx.Err()
		}
		return unixRateBudgetResponse{}, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	if err := ctx.Err(); err != nil {
		return unixRateBudgetResponse{}, err
	}
	if response.Version != unixRateBudgetProtocolVersion || response.Kind != request.Kind {
		return unixRateBudgetResponse{}, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
	switch response.Error.Code {
	case unixRateBudgetErrorNone:
		return response, nil
	case unixRateBudgetErrorEpochMismatch:
		return unixRateBudgetResponse{}, &SharedCoordinatorError{reason: sharedCoordinatorEpochMismatch}
	default:
		return unixRateBudgetResponse{}, &SharedCoordinatorError{reason: sharedCoordinatorUnavailable}
	}
}

func newUnixRateBudgetEpoch() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes[:]), nil
}

type unixRateBudgetMessageKind uint8

const (
	unixRateBudgetReady unixRateBudgetMessageKind = iota + 1
	unixRateBudgetDecide
	unixRateBudgetFinish
)

type unixRateBudgetErrorCode uint8

const (
	unixRateBudgetErrorNone unixRateBudgetErrorCode = iota
	unixRateBudgetErrorUnavailable
	unixRateBudgetErrorEpochMismatch
	unixRateBudgetErrorRefused
)

type unixRateBudgetWireError struct {
	Code unixRateBudgetErrorCode `json:"code"`
}

type unixRateBudgetRequest struct {
	Version     int                           `json:"version"`
	Kind        unixRateBudgetMessageKind     `json:"kind"`
	Epoch       string                        `json:"epoch"`
	Batch       connsdk.ReservationBatch      `json:"batch,omitempty"`
	Lease       connsdk.RateBudgetLease       `json:"lease,omitempty"`
	Observation connsdk.CompletionObservation `json:"observation,omitempty"`
}

type unixRateBudgetResponse struct {
	Version  int                       `json:"version"`
	Kind     unixRateBudgetMessageKind `json:"kind"`
	Ready    bool                      `json:"ready,omitempty"`
	Decision connsdk.AdmissionDecision `json:"decision,omitempty"`
	Error    unixRateBudgetWireError   `json:"error,omitempty"`
}

func writeUnixRateBudgetRequest(writer io.Writer, request unixRateBudgetRequest) error {
	return writeUnixRateBudgetMessage(writer, request)
}

func readUnixRateBudgetRequest(reader io.Reader) (unixRateBudgetRequest, error) {
	var request unixRateBudgetRequest
	err := readUnixRateBudgetMessage(reader, &request)
	return request, err
}

func writeUnixRateBudgetResponse(writer io.Writer, response unixRateBudgetResponse) error {
	return writeUnixRateBudgetMessage(writer, response)
}

func readUnixRateBudgetResponse(reader io.Reader) (unixRateBudgetResponse, error) {
	var response unixRateBudgetResponse
	err := readUnixRateBudgetMessage(reader, &response)
	return response, err
}

func writeUnixRateBudgetMessage(writer io.Writer, message any) error {
	payload, err := json.Marshal(message)
	if err != nil || len(payload) == 0 || len(payload) > maxUnixRateBudgetMessage {
		return errors.New("shared rate-budget coordinator message is invalid")
	}
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(payload)))
	if err := writeUnixRateBudgetBytes(writer, size[:]); err != nil {
		return err
	}
	return writeUnixRateBudgetBytes(writer, payload)
}

func writeUnixRateBudgetBytes(writer io.Writer, payload []byte) error {
	for len(payload) > 0 {
		written, err := writer.Write(payload)
		if err != nil {
			return err
		}
		if written <= 0 {
			return io.ErrShortWrite
		}
		payload = payload[written:]
	}
	return nil
}

func readUnixRateBudgetMessage(reader io.Reader, target any) error {
	var size [4]byte
	if _, err := io.ReadFull(reader, size[:]); err != nil {
		return err
	}
	length := binary.BigEndian.Uint32(size[:])
	if length == 0 || length > maxUnixRateBudgetMessage {
		return errors.New("shared rate-budget coordinator message is invalid")
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(reader, payload); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return errors.New("shared rate-budget coordinator message is invalid")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("shared rate-budget coordinator message is invalid")
	}
	return nil
}
