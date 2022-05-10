package chatbox

import "sync"

type SafeMap[T any] struct {
	items map[string]T
	l     sync.RWMutex
}

func (u *SafeMap[T]) Keys() []string {
	keys := make([]string, 0, len(u.items))
	for k := range u.items {
		keys = append(keys, k)
	}
	return keys
}

func (u *SafeMap[T]) Values() []T {
	values := make([]T, 0, len(u.items))
	for _, v := range u.items {
		values = append(values, v)
	}
	return values
}

func (u *SafeMap[T]) Get(key string) (T, bool) {
	u.l.RLock()
	i, ok := u.items[key]
	u.l.RUnlock()
	return i, ok
}

func (u *SafeMap[T]) Set(key string, value T) {
	u.l.Lock()
	u.items[key] = value
	u.l.Unlock()
}

func (u *SafeMap[T]) Remove(key string) bool {
	var ok bool
	u.l.Lock()
	if _, ok = u.items[key]; ok {
		delete(u.items, key)
	}
	u.l.Unlock()
	return ok
}

func NewSafeMap[T any]() *SafeMap[T] {
	return &SafeMap[T]{
		items: make(map[string]T),
		l:     sync.RWMutex{},
	}
}
