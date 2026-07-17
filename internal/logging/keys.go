package logging

import "sync"

const maxDynamicSensitiveKeys = 4096

var dynamicSensitiveKeys = struct {
	mu    sync.RWMutex
	keys  map[string]struct{}
	order []string
}{keys: map[string]struct{}{}}

// RegisterSensitiveKey marks key as sensitive for all redacting handlers.
func RegisterSensitiveKey(key string) {
	normalized := normalizeKey(key)
	if normalized == "" {
		return
	}
	dynamicSensitiveKeys.mu.Lock()
	defer dynamicSensitiveKeys.mu.Unlock()
	if _, exists := dynamicSensitiveKeys.keys[normalized]; exists {
		return
	}
	dynamicSensitiveKeys.keys[normalized] = struct{}{}
	dynamicSensitiveKeys.order = append(dynamicSensitiveKeys.order, normalized)
	for len(dynamicSensitiveKeys.order) > maxDynamicSensitiveKeys {
		oldest := dynamicSensitiveKeys.order[0]
		copy(dynamicSensitiveKeys.order, dynamicSensitiveKeys.order[1:])
		dynamicSensitiveKeys.order = dynamicSensitiveKeys.order[:len(dynamicSensitiveKeys.order)-1]
		delete(dynamicSensitiveKeys.keys, oldest)
	}
}

func registeredSensitiveKey(key string) bool {
	normalized := normalizeKey(key)
	dynamicSensitiveKeys.mu.RLock()
	defer dynamicSensitiveKeys.mu.RUnlock()
	_, ok := dynamicSensitiveKeys.keys[normalized]
	return ok
}
