package synctransport

import (
	"fmt"
	"reflect"
	"sync"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

// Registry owns exact source and destination executor registrations. It does
// not register a provider by capability bit or infer one from a core
// Connector's Read/Write methods.
type Registry struct {
	mu           sync.RWMutex
	sources      map[connectors.TransportExecutorReference]SourceExecutor
	destinations map[connectors.TransportExecutorReference]DestinationExecutor
	verifier     ConformanceVerifier
}

func NewRegistry(verifier ConformanceVerifier) *Registry {
	if isNilInterface(verifier) {
		verifier = unavailableConformanceVerifier{}
	}
	return &Registry{
		sources:      make(map[connectors.TransportExecutorReference]SourceExecutor),
		destinations: make(map[connectors.TransportExecutorReference]DestinationExecutor),
		verifier:     verifier,
	}
}

func (r *Registry) RegisterSource(executor SourceExecutor) error {
	if r == nil {
		return fmt.Errorf("transport registry is required")
	}
	if isNilInterface(executor) {
		return fmt.Errorf("source transport executor is required")
	}
	reference := executor.TransportExecutorReference()
	if err := reference.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.sources[reference]; exists {
		return fmt.Errorf("source transport executor %q is already registered", reference.ID)
	}
	r.sources[reference] = executor
	return nil
}

func (r *Registry) RegisterDestination(executor DestinationExecutor) error {
	if r == nil {
		return fmt.Errorf("transport registry is required")
	}
	if isNilInterface(executor) {
		return fmt.Errorf("destination transport executor is required")
	}
	reference := executor.TransportExecutorReference()
	if err := reference.Validate(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.destinations[reference]; exists {
		return fmt.Errorf("destination transport executor %q is already registered", reference.ID)
	}
	r.destinations[reference] = executor
	return nil
}

// ResolvedTransport is the immutable result of a successful runtime
// preflight. It contains no provider records and no self-reported evidence.
type ResolvedTransport struct {
	Source                SourceExecutor
	Destination           DestinationExecutor
	SourceDescriptor      connectors.SourceTransportDescriptor
	DestinationDescriptor connectors.DestinationTransportDescriptor
	ApplyStrategy         connectors.DestinationApplyStrategy
}

// Preflight proves that both endpoint descriptors, closed families, exact
// registrations, mode, strategy, acknowledgement policy, and independent
// conformance verifier agree before a source executor can read.
func (r *Registry) Preflight(request PreflightRequest) (ResolvedTransport, error) {
	if r == nil {
		return ResolvedTransport{}, fmt.Errorf("transport registry is required")
	}
	if isNilConnector(request.Source) || isNilConnector(request.Destination) {
		return ResolvedTransport{}, fmt.Errorf("source and destination connectors are required")
	}
	if err := validateTransportName("stream", request.Stream); err != nil {
		return ResolvedTransport{}, err
	}
	if err := request.Mode.Validate(); err != nil {
		return ResolvedTransport{}, err
	}

	sourceTransport, ok := connectors.SyncTransportDescriptorOf(request.Source)
	if !ok || sourceTransport.Source == nil {
		return ResolvedTransport{}, fmt.Errorf("source connector %q has no declared source transport", request.Source.Name())
	}
	if err := sourceTransport.Validate(); err != nil {
		return ResolvedTransport{}, fmt.Errorf("source transport descriptor: %w", err)
	}
	sourceDescriptor := sourceTransport.Source
	destinationTransport, ok := connectors.SyncTransportDescriptorOf(request.Destination)
	if !ok || destinationTransport.Destination == nil {
		return ResolvedTransport{}, fmt.Errorf("destination connector %q has no declared destination transport", request.Destination.Name())
	}
	if err := destinationTransport.Validate(); err != nil {
		return ResolvedTransport{}, fmt.Errorf("destination transport descriptor: %w", err)
	}
	destinationDescriptor := destinationTransport.Destination
	if err := connectors.ValidateTransportExecutorFamily(request.Source.Metadata().IntegrationType, sourceDescriptor.Executor); err != nil {
		return ResolvedTransport{}, err
	}
	if err := connectors.ValidateTransportExecutorFamily(request.Destination.Metadata().IntegrationType, destinationDescriptor.Executor); err != nil {
		return ResolvedTransport{}, err
	}
	if !containsMode(sourceDescriptor.Modes, request.Mode) {
		return ResolvedTransport{}, fmt.Errorf("source transport does not support sync mode %q", request.Mode)
	}
	if !containsMode(destinationDescriptor.Modes, request.Mode) {
		return ResolvedTransport{}, fmt.Errorf("destination transport does not support sync mode %q", request.Mode)
	}
	if !containsName(sourceDescriptor.EligibleStreams, request.Stream) {
		return ResolvedTransport{}, fmt.Errorf("source transport does not support stream %q", request.Stream)
	}
	if destinationDescriptor.Acknowledgement != connectors.TransportAcknowledgementDurableWarehouse {
		return ResolvedTransport{}, fmt.Errorf("destination transport requires durable warehouse acknowledgement")
	}
	strategy, err := destinationDescriptor.ApplyStrategyFor(request.Mode)
	if err != nil {
		return ResolvedTransport{}, err
	}

	r.mu.RLock()
	source, sourceRegistered := r.sources[sourceDescriptor.Executor]
	destination, destinationRegistered := r.destinations[destinationDescriptor.Executor]
	verifier := r.verifier
	r.mu.RUnlock()
	if !sourceRegistered || isNilInterface(source) {
		return ResolvedTransport{}, fmt.Errorf("source transport executor %q is not registered", sourceDescriptor.Executor.ID)
	}
	if !destinationRegistered || isNilInterface(destination) {
		return ResolvedTransport{}, fmt.Errorf("destination transport executor %q is not registered", destinationDescriptor.Executor.ID)
	}
	if source.TransportExecutorReference() != sourceDescriptor.Executor {
		return ResolvedTransport{}, fmt.Errorf("registered source transport executor does not match source descriptor")
	}
	if destination.TransportExecutorReference() != destinationDescriptor.Executor {
		return ResolvedTransport{}, fmt.Errorf("registered destination transport executor does not match destination descriptor")
	}
	if isNilInterface(verifier) {
		return ResolvedTransport{}, fmt.Errorf("external transport conformance verification is unavailable")
	}
	if err := verifier.VerifyTransportConformance(ConformanceVerification{
		Role:          connectors.TransportRoleSource,
		ConnectorName: request.Source.Name(),
		Executor:      sourceDescriptor.Executor,
		Evidence:      sourceDescriptor.Conformance,
	}); err != nil {
		return ResolvedTransport{}, fmt.Errorf("source transport conformance: %w", err)
	}
	if err := verifier.VerifyTransportConformance(ConformanceVerification{
		Role:          connectors.TransportRoleDestination,
		ConnectorName: request.Destination.Name(),
		Executor:      destinationDescriptor.Executor,
		Evidence:      destinationDescriptor.Conformance,
	}); err != nil {
		return ResolvedTransport{}, fmt.Errorf("destination transport conformance: %w", err)
	}

	return ResolvedTransport{
		Source:                source,
		Destination:           destination,
		SourceDescriptor:      *sourceDescriptor,
		DestinationDescriptor: *destinationDescriptor,
		ApplyStrategy:         strategy,
	}, nil
}

func containsMode(modes []synccontract.Mode, want synccontract.Mode) bool {
	for _, mode := range modes {
		if mode == want {
			return true
		}
	}
	return false
}

func containsName(values []string, want string) bool {
	for _, value := range values {
		if value == want || value == "*" {
			return true
		}
	}
	return false
}

func isNilConnector(connector connectors.Connector) bool {
	if connector == nil {
		return true
	}
	return isNilInterface(connector)
}

func isNilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
