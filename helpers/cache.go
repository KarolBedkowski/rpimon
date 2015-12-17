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
	mutex     sync.RWMutex
	timestamp time.Time
	value     Value
	maxAge    time.Duration
}

// NewSimpleCache create new cache structure
func NewSimpleCache(maxCacheAge int) *SimpleCache {
	return &SimpleCache{maxAge: time.Duration(maxCacheAge) * time.Second}
}

// Get value from cache; if cache is expired - call function and put result in cache
func (cache *SimpleCache) Get(f func() Value) Value {
	cache.mutex.RLock()
	now := time.Now()
	if cache.value != nil && now.Sub(cache.timestamp) < cache.maxAge {
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
	if cache.value != nil && time.Now().Sub(cache.timestamp) < cache.maxAge {
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

type SimpleCacheCB struct {
	*SimpleCache
	provider CacheValueProvider
}

type CacheValueProvider func() Value

// NewSimpleCacheCB create new cache structure
func NewSimpleCacheCB(maxCacheAge int, f CacheValueProvider) *SimpleCacheCB {
	return &SimpleCacheCB{
		SimpleCache: NewSimpleCache(maxCacheAge),
		provider:    f,
	}
}

// Get value from cache; if cache is expired - call function and put result in cache
func (cache *SimpleCacheCB) Get() Value {
	cache.mutex.RLock()
	now := time.Now()
	if cache.value != nil && now.Sub(cache.timestamp) < cache.maxAge {
		defer cache.mutex.RUnlock()
		return cache.value
	}
	cache.mutex.RUnlock()

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	value := cache.provider()
	cache.value = value
	cache.timestamp = now
	return value
}

type cacheItem struct {
	timestamp time.Time
	value     Value
}

// Cache structure
type Cache struct {
	mutex  sync.RWMutex
	values map[string]cacheItem
	maxAge time.Duration
}

// NewCache create new cache structure
func NewCache(maxCacheAge int) *Cache {
	return &Cache{maxAge: time.Duration(maxCacheAge) * time.Second,
		values: make(map[string]cacheItem)}
}

// Get value from cache; if cache is expired - call function and put result in cache
func (cache *Cache) Get(key string, f func(fkey string) Value) (value Value) {
	cache.mutex.RLock()
	now := time.Now()
	item, ok := cache.values[key]
	if ok && now.Sub(item.timestamp) < cache.maxAge {
		defer cache.mutex.RUnlock()
		return item.value
	}
	cache.mutex.RUnlock()

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	value = f(key)
	cache.values[key] = cacheItem{now, value}
	return value
}

// GetValue from cache
func (cache *Cache) GetValue(key string) (value Value, ok bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	now := time.Now()
	item, ok := cache.values[key]
	if ok && now.Sub(item.timestamp) < cache.maxAge {
		return item.value, true
	}
	return nil, false
}

// SetValue put value to cache
func (cache *Cache) SetValue(key string, value Value) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.values[key] = cacheItem{time.Now(), value}
}

// Clear cache
func (cache *Cache) Clear() {
	cache.values = make(map[string]cacheItem)
}
