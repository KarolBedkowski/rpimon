package helpers

import "fmt"
import "sync"

// RingBuffer is structure that holds no more than configured items
type RingBuffer struct {
	sync.RWMutex
	size    int
	maxSize int
	items   []interface{}
	first   int
	last    int
}

// NewRingBuffer create new RingBuffer structure with given max size
func NewRingBuffer(size int) *RingBuffer {
	rbuff := &RingBuffer{
		maxSize: size,
		items:   make([]interface{}, size, size),
	}
	return rbuff
}

// Size return buffer size (max number of items)
func (r *RingBuffer) Size() int {
	return r.maxSize
}

// Len return number of elements in buffer
func (r *RingBuffer) Len() int {
	return r.size
}

// Clear remove all items from buffer
func (r *RingBuffer) Clear() {
	r.Lock()
	defer r.Unlock()
	r.size = 0
	r.items = make([]interface{}, r.maxSize, r.maxSize)
	r.first = 0
	r.last = 0
}

// Put insert item to RingBuffer and return number of items in buffer
func (r *RingBuffer) Put(item interface{}) int {
	r.Lock()
	defer r.Unlock()
	r.items[r.last] = item
	r.last = (r.last + 1) % r.maxSize
	if r.size < r.maxSize {
		r.size++
	} else {
		r.first = (r.first + 1) % r.maxSize
	}
	return r.size
}

// Get return item from buffer
func (r *RingBuffer) Get(index int) (interface{}, bool) {
	r.RLock()
	defer r.RUnlock()
	if index >= r.size {
		return nil, false
	}
	return r.items[(r.first+index)%r.maxSize], true
}

func (r *RingBuffer) String() string {
	r.RLock()
	defer r.RUnlock()
	return fmt.Sprintf("RingBuffer[size=%d, maxSize=%d, first=%d, last=%d, items='%#v']",
		r.size, r.maxSize, r.first, r.last, r.items)
}

// ToSlice return new slice with all elements in buffer
func (r *RingBuffer) ToSlice() []interface{} {
	r.RLock()
	defer r.RUnlock()
	if r.size < r.maxSize {
		return r.items[:r.size]
	}
	items := r.items[r.first:]
	return append(items, r.items[:r.last]...)
}

// ToStringSlice return new slice with all elements in buffer
func (r *RingBuffer) ToStringSlice() []string {
	r.RLock()
	defer r.RUnlock()
	result := make([]string, r.size, r.size)
	if r.size < r.maxSize {
		for i := 0; i < r.size; i++ {
			result[i] = r.items[i].(string)
		}
	} else {
		for i, j := r.first, 0; i < r.maxSize; i++ {
			result[j] = r.items[i].(string)
			j++
		}
		for i, j := 0, (r.maxSize - r.first); i < r.first; i++ {
			result[j] = r.items[i].(string)
			j++
		}
	}
	return result
}

// ToInt64Slice return new slice with all elements in buffer
func (r *RingBuffer) ToUInt64Slice() []uint64 {
	r.RLock()
	defer r.RUnlock()
	result := make([]uint64, r.size, r.size)
	if r.size < r.maxSize {
		for i := 0; i < r.size; i++ {
			result[i] = r.items[i].(uint64)
		}
	} else {
		for i, j := r.first, 0; i < r.maxSize; i++ {
			result[j] = r.items[i].(uint64)
			j++
		}
		for i, j := 0, (r.maxSize - r.first); i < r.first; i++ {
			result[j] = r.items[i].(uint64)
			j++
		}
	}
	return result
}
