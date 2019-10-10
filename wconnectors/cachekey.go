package wconnectors

import (
	"strings"
)

var cacheKeys map[string]string

func RegisterCacheKeys(newCacheKeys map[string]string) {
	cacheKeys = newCacheKeys
}

// GetCacheKey replaces the placeholders in the config key with actual values
func GetCacheKey(cacheKey string, mapValues map[string]string) (string, error) {
	cacheKey, ok := cacheKeys[cacheKey]
	if !ok {
		return "", 
	}
	for mapKey, mapValue := range mapValues {
		cacheKey = strings.Replace(cacheKey, "["+mapKey+"]", mapValue, 1)
	}
	return cacheKey, nil
}
