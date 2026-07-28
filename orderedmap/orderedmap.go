// Package orderedmap provides a concurrency-safe generic map that preserves
// insertion order.
package orderedmap

import "sync"

type Map[K comparable, V any] struct {
	mu     sync.RWMutex
	keys   []K
	values map[K]V
}

// New creates an ordered map with optional initial capacity.
func New[K comparable, V any](capacity ...int) *Map[K, V] {
	size := 0
	if len(capacity) > 0 && capacity[0] > 0 {
		size = capacity[0]
	}
	return &Map[K, V]{
		keys:   make([]K, 0, size),
		values: make(map[K]V, size),
	}
}

// Set inserts or replaces a value. Replacing a value preserves its position.
func (m *Map[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.values == nil {
		m.values = make(map[K]V)
	}
	if _, exists := m.values[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.values[key] = value
}

func (m *Map[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value, ok := m.values[key]
	return value, ok
}

func (m *Map[K, V]) Delete(key K) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.values[key]; !exists {
		return false
	}
	delete(m.values, key)
	for index, candidate := range m.keys {
		if candidate == key {
			copy(m.keys[index:], m.keys[index+1:])
			var zero K
			m.keys[len(m.keys)-1] = zero
			m.keys = m.keys[:len(m.keys)-1]
			break
		}
	}
	return true
}

func (m *Map[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.values)
}

func (m *Map[K, V]) Keys() []K {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]K(nil), m.keys...)
}

func (m *Map[K, V]) Values() []V {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]V, 0, len(m.keys))
	for _, key := range m.keys {
		result = append(result, m.values[key])
	}
	return result
}

// At returns the entry at an insertion-order index.
func (m *Map[K, V]) At(index int) (K, V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if index < 0 || index >= len(m.keys) {
		var key K
		var value V
		return key, value, false
	}
	key := m.keys[index]
	return key, m.values[key], true
}

// Range visits a stable snapshot in insertion order until visit returns false.
func (m *Map[K, V]) Range(visit func(K, V) bool) {
	if visit == nil {
		return
	}
	m.mu.RLock()
	keys := append([]K(nil), m.keys...)
	values := make([]V, 0, len(keys))
	for _, key := range keys {
		values = append(values, m.values[key])
	}
	m.mu.RUnlock()
	for index, key := range keys {
		if !visit(key, values[index]) {
			return
		}
	}
}

func (m *Map[K, V]) Clear() {
	m.mu.Lock()
	m.keys = nil
	m.values = nil
	m.mu.Unlock()
}
