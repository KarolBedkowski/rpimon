/*
Package helpers - generic cache for functions

Example:
var fsInfoCache = h.NewSimpleCache(5)

func GetFilesystemsInfo() *FilesystemsStruct {
	result := fsInfoCache.Get(func() h.Value {
		return gatherFSInfo()
	})
	return result.(*FilesystemsStruct)
}
*/
package helpers

import (
	"sync"
	"time"
)

// Value stored in cache
type Value interface{}

// SimpleCache structure holding all settings of cache
type SimpleCache struct {
	mutex       sync.RWMutex
	timestamp   time.Time
	value       Value
	MaxCacheAge time.Duration
}

// NewSimpleCache create new cache structure
func NewSimpleCache(maxCacheAge int) *SimpleCache {
	return &SimpleCache{MaxCacheAge: time.Duration(maxCacheAge) * time.Second}
}

// Get value from cache; if cache is expired - call function and put result in cache
func (cache *SimpleCache) Get(f func() Value) Value {
	cache.mutex.RLock()
	now := time.Now()
	if cache.value != nil && now.Sub(cache.timestamp) < cache.MaxCacheAge {
		defer cache.mutex.RUnlock()
		return cache.value
	}
	cache.mutex.RUnlock()

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	value := f()
	cache.value = value
	cache.timestamp = now
	return value
}
