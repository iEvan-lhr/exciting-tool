package tools

import (
	"sync"
	"time"
)

type LockA struct {
	unix time.Time
	Name string `json:"name"`

	entry    *namedLockEntry
	released *sync.Once
}

type namedLockEntry struct {
	mutex sync.Mutex
	refs  int
}

var namedLockRegistry = struct {
	sync.Mutex
	entries map[string]*namedLockEntry
}{
	entries: make(map[string]*namedLockEntry),
}

// Lock acquires a mutex identified by a string. Call Unlock on the returned
// value when the protected operation is complete.
func Lock(value any) LockA {
	switch lock := value.(type) {
	case string:
		namedLockRegistry.Lock()
		entry := namedLockRegistry.entries[lock]
		if entry == nil {
			entry = &namedLockEntry{}
			namedLockRegistry.entries[lock] = entry
		}
		entry.refs++
		namedLockRegistry.Unlock()

		entry.mutex.Lock()
		return LockA{
			unix:     time.Now(),
			Name:     lock,
			entry:    entry,
			released: &sync.Once{},
		}
	case func():
		lock()
	case func(LockA):
		lock(LockA{unix: time.Now()})
	}
	return LockA{}
}

// Unlock releases a named lock. Repeated calls are safe.
func (l LockA) Unlock() {
	if l.entry == nil || l.released == nil {
		return
	}
	l.released.Do(func() {
		l.entry.mutex.Unlock()
		namedLockRegistry.Lock()
		l.entry.refs--
		if l.entry.refs == 0 {
			delete(namedLockRegistry.entries, l.Name)
		}
		namedLockRegistry.Unlock()
	})
}

// LockFunc executes f while holding the named lock.
func LockFunc(name string, f func()) {
	if f == nil {
		return
	}
	lock := Lock(name)
	defer lock.Unlock()
	f()
}
