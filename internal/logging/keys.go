package logging

import "sync"

var dynamicSensitiveKeys = struct {
	mu   sync.RWMutex
	keys map[string]struct{}
}{keys: map[string]struct{}{}}

// RegisterSensitiveKey marks key as sensitive for all redacting handlers.
func RegisterSensitiveKey(key string) {
	normalized := normalizeKey(key)
	if normalized == "" {
		return
	}
	dynamicSensitiveKeys.mu.Lock()
	defer dynamicSensitiveKeys.mu.Unlock()
	dynamicSensitiveKeys.keys[normalized] = struct{}{}
}

func registeredSensitiveKey(key string) bool {
	normalized := normalizeKey(key)
	dynamicSensitiveKeys.mu.RLock()
	defer dynamicSensitiveKeys.mu.RUnlock()
	_, ok := dynamicSensitiveKeys.keys[normalized]
	return ok
}
