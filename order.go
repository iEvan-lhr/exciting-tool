package tools

import (
	"sync"
)

// Map stores values in insertion order.
// Deprecated: Use orderedmap.Map for type-safe keys and values.
type Map struct {
	mu    sync.RWMutex
	key   map[string]int
	value []any
}

func (m *Map) Set(key, value any) {
	if m == nil {
		return
	}
	normalized := spiderKey(key)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.key == nil {
		m.key = make(map[string]int)
	}
	if index, ok := m.key[normalized]; ok {
		m.value[index] = value
		return
	}
	m.key[normalized] = len(m.value)
	m.value = append(m.value, value)
}

func (m *Map) Get(key any) (any, bool) {
	if m == nil {
		return nil, false
	}
	normalized := spiderKey(key)
	m.mu.RLock()
	defer m.mu.RUnlock()
	index, ok := m.key[normalized]
	if !ok {
		return nil, false
	}
	return m.value[index], true
}

func (m *Map) Len() int {
	if m == nil {
		return 0
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.value)
}
