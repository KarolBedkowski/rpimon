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

// GetValue from cache
func (cache *SimpleCache) GetValue() (value Value, ok bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	if cache.value != nil && time.Now().Sub(cache.timestamp) < cache.MaxCacheAge {
		return cache.value, true
	}
	return
}

// SetValue put value to cache
func (cache *SimpleCache) SetValue(value Value) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.value = value
	cache.timestamp = time.Now()
}

// Clear cache
func (cache *SimpleCache) Clear() {
	cache.value = nil
}

type keycacheitem struct {
	timestamp time.Time
	value     Value
}

// KeyCache structure
type KeyCache struct {
	mutex       sync.RWMutex
	values      map[string]keycacheitem
	MaxCacheAge time.Duration
}

// NewKeyCache create new cache structure
func NewKeyCache(maxCacheAge int) *KeyCache {
	return &KeyCache{MaxCacheAge: time.Duration(maxCacheAge) * time.Second,
		values: make(map[string]keycacheitem)}
}

// Get value from cache; if cache is expired - call function and put result in cache
func (cache *KeyCache) Get(key string, f func(fkey string) Value) (value Value) {
	cache.mutex.RLock()
	now := time.Now()
	cacheItem, ok := cache.values[key]
	if ok && now.Sub(cacheItem.timestamp) < cache.MaxCacheAge {
		defer cache.mutex.RUnlock()
		return cacheItem.value
	}
	cache.mutex.RUnlock()

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	value = f(key)
	cache.values[key] = keycacheitem{now, value}
	return value
}

// GetValue from cache
func (cache *KeyCache) GetValue(key string) (value Value, ok bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	now := time.Now()
	cacheItem, ok := cache.values[key]
	if ok && now.Sub(cacheItem.timestamp) < cache.MaxCacheAge {
		return cacheItem.value, true
	}
	return nil, false
}

// SetValue put value to cache
func (cache *KeyCache) SetValue(key string, value Value) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.values[key] = keycacheitem{time.Now(), value}
}

// Clear cache
func (cache *KeyCache) Clear() {
	cache.values = make(map[string]keycacheitem)
}
