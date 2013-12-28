/*
Geenric Cache functions

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

type Value interface{}
type SimpleCache struct {
	mutex       sync.RWMutex
	timestamp   time.Time
	value       Value
	MaxCacheAge time.Duration
}

func NewSimpleCache(maxCacheAge int) *SimpleCache {
	return &SimpleCache{MaxCacheAge: time.Duration(maxCacheAge) * time.Second}
}

type CacheValueFunc func() Value

func (cache *SimpleCache) Get(f CacheValueFunc) Value {
	cache.mutex.RLock()
	now := time.Now()
	if cache.value != nil && now.Sub(cache.timestamp) < cache.MaxCacheAge {
		cache.mutex.RUnlock()
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
