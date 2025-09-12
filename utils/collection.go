package utils

import (
	"strings"
	"sync"
)

func SubSlice[T any](slice []T, from, to uint) []T {
	if int(from) > len(slice) {
		from = 0
	}
	if int(to) > len(slice) {
		to = uint(len(slice))
	}
	return slice[from:int(to)]
}

func CloneStringMap(m map[string]string, keysToLower bool) map[string]string {
	m2 := make(map[string]string)
	for k, v := range m {
		if keysToLower {
			k = strings.ToLower(k)
		}
		m2[k] = v
	}
	return m2
}

// Concurrent hash map implementation

type ConcurrentMap[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

func NewConcurrentMap[K comparable, V any]() *ConcurrentMap[K, V] {
	return &ConcurrentMap[K, V]{
		items: make(map[K]V),
	}
}

func (m *ConcurrentMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[key] = value
}

func (m *ConcurrentMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.items[key]
	return value, ok
}

func (m *ConcurrentMap[K, V]) MustGet(key K) V {
	val, _ := m.Get(key)
	return val
}

func (m *ConcurrentMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, key)
}

func (m *ConcurrentMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.items)
}
