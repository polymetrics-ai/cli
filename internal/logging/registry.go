package logging

import (
	"crypto/rand"
	"crypto/sha256"
	"hash"
	"net/url"
	"sort"
	"sync"
)

const redactedValue = "[redacted]"
const defaultMaxRegistryEntries = 4096

// ValueRegistry stores fingerprints of values that must be redacted from log text.
// It never stores or returns the raw registered values.
type ValueRegistry struct {
	mu         sync.RWMutex
	salt       [32]byte
	entries    map[int]map[[32]byte]struct{}
	order      []registryEntry
	maxEntries int
}

type registryEntry struct {
	length int
	digest [32]byte
}

var defaultRegistry = NewValueRegistry()

// NewValueRegistry returns an empty bounded concurrency-safe redaction registry.
func NewValueRegistry() *ValueRegistry {
	return NewValueRegistryWithLimit(defaultMaxRegistryEntries)
}

// NewValueRegistryWithLimit returns an empty registry that retains at most maxEntries fingerprints.
func NewValueRegistryWithLimit(maxEntries int) *ValueRegistry {
	if maxEntries <= 0 {
		maxEntries = defaultMaxRegistryEntries
	}
	r := &ValueRegistry{entries: map[int]map[[32]byte]struct{}{}, maxEntries: maxEntries}
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
	variants := registryVariants(value)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.entries == nil {
		r.entries = map[int]map[[32]byte]struct{}{}
	}
	for _, variant := range variants {
		r.registerLocked(variant)
	}
}

func registryVariants(value string) []string {
	variants := []string{value}
	for _, encoded := range []string{url.QueryEscape(value), url.PathEscape(value)} {
		if encoded == "" || encoded == value {
			continue
		}
		seen := false
		for _, existing := range variants {
			if existing == encoded {
				seen = true
				break
			}
		}
		if !seen {
			variants = append(variants, encoded)
		}
	}
	return variants
}

func (r *ValueRegistry) registerLocked(value string) {
	if value == "" {
		return
	}
	digest := r.digest(value)
	length := len(value)
	if r.entries[length] == nil {
		r.entries[length] = map[[32]byte]struct{}{}
	}
	if _, exists := r.entries[length][digest]; exists {
		return
	}
	r.entries[length][digest] = struct{}{}
	r.order = append(r.order, registryEntry{length: length, digest: digest})
	r.pruneLocked()
}

func (r *ValueRegistry) pruneLocked() {
	for r.maxEntries > 0 && len(r.order) > r.maxEntries {
		oldest := r.order[0]
		copy(r.order, r.order[1:])
		r.order = r.order[:len(r.order)-1]
		if set := r.entries[oldest.length]; set != nil {
			delete(set, oldest.digest)
			if len(set) == 0 {
				delete(r.entries, oldest.length)
			}
		}
	}
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
