package app

import (
	"fmt"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
)

const maximumTargetCopyWorkers = 8

// TargetCopyWorkersUnsupportedError is returned before a connection is saved
// when its destination has not declared a bounded immutable-COPY capacity.
type TargetCopyWorkersUnsupportedError struct{ Destination string }

func (e *TargetCopyWorkersUnsupportedError) Error() string {
	if e == nil {
		return "target COPY workers are not supported by the selected destination"
	}
	return fmt.Sprintf("target COPY workers are not supported by destination %q", e.Destination)
}

// TargetCopyWorkersRangeError reports the declaration-derived upper bound.
type TargetCopyWorkersRangeError struct {
	Requested int
	Maximum   int
}

func (e *TargetCopyWorkersRangeError) Error() string {
	if e == nil {
		return "target COPY workers are outside the declared destination limit"
	}
	return fmt.Sprintf("target COPY workers must be between 1 and %d (requested %d)", e.Maximum, e.Requested)
}

func targetCopyWorkerMaximum(destination connectors.Connector) int {
	descriptor, declared := connectors.DestinationTransportDescriptorOf(destination)
	if !declared || descriptor.CopyWorkerMaximum <= 0 {
		return 0
	}
	if descriptor.CopyWorkerMaximum < maximumTargetCopyWorkers {
		return descriptor.CopyWorkerMaximum
	}
	return maximumTargetCopyWorkers
}

func hasImmutableFullOverwriteStream(streams map[string]StreamConfig) bool {
	for _, stream := range streams {
		mode, err := ParseStreamSyncMode(stream)
		if err == nil && mode.ContractMode == synccontract.ModeFullOverwrite && stream.TransformPlan != "" && stream.TransformPlanHash != "" {
			return true
		}
	}
	return false
}

// resolveTargetCopyWorkers turns a destination declaration into stored
// connection policy. A value above one is reserved only for the immutable
// full-overwrite COPY contract; all other modes retain their ordered single
// lane and cannot claim this capacity.
func resolveTargetCopyWorkers(destination connectors.Connector, streams map[string]StreamConfig, requested int) (effective, maximum int, err error) {
	maximum = targetCopyWorkerMaximum(destination)
	if requested < 0 || requested > maximumTargetCopyWorkers {
		return 0, maximum, &TargetCopyWorkersRangeError{Requested: requested, Maximum: maximumTargetCopyWorkers}
	}
	if maximum == 0 || !hasImmutableFullOverwriteStream(streams) {
		if requested == 0 {
			return 0, maximum, nil
		}
		return 0, maximum, &TargetCopyWorkersUnsupportedError{Destination: destination.Name()}
	}
	if requested == 0 {
		requested = 2
		if requested > maximum {
			requested = maximum
		}
	}
	if requested < 1 || requested > maximum {
		return 0, maximum, &TargetCopyWorkersRangeError{Requested: requested, Maximum: maximum}
	}
	return requested, maximum, nil
}
