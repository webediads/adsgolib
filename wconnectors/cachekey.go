package wconnectors

import (
	"strings"
)

var cacheKeys map[string]string

// RegisterCacheKeys register the cache keys
func RegisterCacheKeys(newCacheKeys map[string]string) {
	cacheKeys = newCacheKeys
}

// GetCacheKey replaces the placeholders in the config key with actual values
func GetCacheKey(cacheKey string, mapValues map[string]string) (string, bool) {
	cacheKey, ok := cacheKeys[cacheKey]
	if !ok {
		return "", false
	}
	for mapKey, mapValue := range mapValues {
		cacheKey = strings.Replace(cacheKey, "["+mapKey+"]", mapValue, 1)
	}
	return cacheKey, true
}
