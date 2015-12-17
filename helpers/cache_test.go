package helpers

import (
	h "k.prv/rpimon/helpers"
	"testing"
	"time"
)

func testDataGenerator() h.Value {
	return time.Now()
}

func TestGetCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
		return
	}
	cache := h.NewSimpleCache(5)
	val1 := cache.Get(testDataGenerator)
	time.Sleep(1 * time.Second)
	val2 := cache.Get(testDataGenerator)
	if val1 != val2 {
		t.Errorf("Value in cache changed after 1 sec")
	}
	time.Sleep(1 * time.Second)
	val3 := cache.Get(testDataGenerator)
	if val1 != val3 {
		t.Errorf("Value in cache changed after 2 sec")
	}
	time.Sleep(4 * time.Second)
	val4 := cache.Get(testDataGenerator)
	if val1 == val4 {
		t.Errorf("Value in cache don't change after 6 sec")
	}
}

func TestGetCacheNoCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
		return
	}
	cache := h.NewSimpleCache(0)
	val1 := cache.Get(testDataGenerator)
	time.Sleep(1 * time.Second)
	val2 := cache.Get(testDataGenerator)
	if val1 == val2 {
		t.Errorf("Value in cache don't change after 1 sec")
	}
	time.Sleep(1 * time.Second)
	val3 := cache.Get(testDataGenerator)
	if val1 == val3 {
		t.Errorf("Value in cache don't change after 2 sec")
	}
}

func TestGetCacheCB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
		return
	}
	cache := h.NewSimpleCacheCB(5, testDataGenerator)
	val1 := cache.Get()
	time.Sleep(1 * time.Second)
	val2 := cache.Get()
	if val1 != val2 {
		t.Errorf("Value in cache changed after 1 sec")
	}
	time.Sleep(1 * time.Second)
	val3 := cache.Get()
	if val1 != val3 {
		t.Errorf("Value in cache changed after 2 sec")
	}
	time.Sleep(4 * time.Second)
	val4 := cache.Get()
	if val1 == val4 {
		t.Errorf("Value in cache don't change after 6 sec")
	}
}

func TestGetCacheNoCacheCB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
		return
	}
	cache := h.NewSimpleCacheCB(0, testDataGenerator)
	val1 := cache.Get()
	time.Sleep(1 * time.Second)
	val2 := cache.Get()
	if val1 == val2 {
		t.Errorf("Value in cache don't change after 1 sec")
	}
	time.Sleep(1 * time.Second)
	val3 := cache.Get()
	if val1 == val3 {
		t.Errorf("Value in cache don't change after 2 sec")
	}
}
