package logging

import (
	"crypto/rand"
	"crypto/sha256"
	"hash"
	"sort"
	"sync"
)

const redactedValue = "[redacted]"

// ValueRegistry stores fingerprints of values that must be redacted from log text.
// It never stores or returns the raw registered values.
type ValueRegistry struct {
	mu      sync.RWMutex
	salt    [32]byte
	entries map[int]map[[32]byte]struct{}
}

var defaultRegistry = NewValueRegistry()

// NewValueRegistry returns an empty concurrency-safe redaction registry.
func NewValueRegistry() *ValueRegistry {
	r := &ValueRegistry{entries: map[int]map[[32]byte]struct{}{}}
	_, _ = rand.Read(r.salt[:])
	return r
}

// RegisterValue adds value to the package-level redaction registry.
func RegisterValue(value string) {
	defaultRegistry.Register(value)
}

// Register adds value to the registry without retaining the raw value.
func (r *ValueRegistry) Register(value string) {
	if r == nil || value == "" {
		return
	}
	digest := r.digest(value)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.entries == nil {
		r.entries = map[int]map[[32]byte]struct{}{}
	}
	length := len(value)
	if r.entries[length] == nil {
		r.entries[length] = map[[32]byte]struct{}{}
	}
	r.entries[length][digest] = struct{}{}
}

func (r *ValueRegistry) redactString(value string) string {
	if r == nil || value == "" {
		return value
	}
	lengths, entries := r.snapshot()
	if len(lengths) == 0 {
		return value
	}

	changed := false
	var out []byte
	start := 0
	for i := 0; i < len(value); {
		matched := 0
		for _, length := range lengths {
			if length == 0 || i+length > len(value) {
				continue
			}
			if _, ok := entries[length][r.digest(value[i:i+length])]; ok {
				matched = length
				break
			}
		}
		if matched == 0 {
			i++
			continue
		}
		if !changed {
			out = make([]byte, 0, len(value))
			changed = true
		}
		out = append(out, value[start:i]...)
		out = append(out, redactedValue...)
		i += matched
		start = i
	}
	if !changed {
		return value
	}
	out = append(out, value[start:]...)
	return string(out)
}

func (r *ValueRegistry) snapshot() ([]int, map[int]map[[32]byte]struct{}) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lengths := make([]int, 0, len(r.entries))
	entries := make(map[int]map[[32]byte]struct{}, len(r.entries))
	for length, set := range r.entries {
		lengths = append(lengths, length)
		copied := make(map[[32]byte]struct{}, len(set))
		for digest := range set {
			copied[digest] = struct{}{}
		}
		entries[length] = copied
	}
	sort.Sort(sort.Reverse(sort.IntSlice(lengths)))
	return lengths, entries
}

func (r *ValueRegistry) digest(value string) [32]byte {
	h := sha256.New()
	writeHash(h, r.salt[:])
	writeHash(h, []byte(value))
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

func writeHash(h hash.Hash, data []byte) {
	_, _ = h.Write(data)
}
