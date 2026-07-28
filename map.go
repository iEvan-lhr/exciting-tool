package tools

import (
	"fmt"
	"sync"
)

func Ok(value any, ok bool) any {
	if ok {
		return value
	}
	return nil
}

// Spider is a concurrent string-keyed map.
//
// Key, Values, and Next are retained for source compatibility with older
// releases. New code should use Add, Get, and Len.
//
// Deprecated: Use orderedmap.Map or sync.Map.
type Spider struct {
	Values any
	Key    []byte
	Next   [255]*Spider

	mu      sync.RWMutex
	entries map[string]any
}

func (s *Spider) Add(key any, value any) {
	if s == nil {
		return
	}
	normalized := spiderKey(key)
	s.mu.Lock()
	if s.entries == nil {
		s.entries = make(map[string]any)
	}
	s.entries[normalized] = value
	s.mu.Unlock()
}

func (s *Spider) Get(key any) any {
	if s == nil {
		return nil
	}
	normalized := spiderKey(key)
	s.mu.RLock()
	value := s.entries[normalized]
	s.mu.RUnlock()
	return value
}

func MakeSpider(key any, value any) *Spider {
	spider := &Spider{
		Key:     []byte(spiderKey(key)),
		entries: make(map[string]any),
	}
	spider.Add(key, value)
	return spider
}

func (s *Spider) Len() int {
	if s == nil {
		return 0
	}
	s.mu.RLock()
	length := len(s.entries)
	s.mu.RUnlock()
	return length
}

func spiderKey(key any) string {
	switch value := key.(type) {
	case *String:
		if value == nil {
			return "*tools.String:<nil>"
		}
		return "*tools.String:" + value.String()
	case String:
		return "tools.String:" + value.String()
	case string:
		return "string:" + value
	case []byte:
		return "[]byte:" + string(value)
	default:
		return fmt.Sprintf("%T:%v", key, key)
	}
}
