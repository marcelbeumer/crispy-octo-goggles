package mutex

import "sync"

type Map[T any] struct {
	items map[string]T
	l     sync.RWMutex
}

func (u *Map[T]) Keys() []string {
	keys := make([]string, 0, len(u.items))
	for k := range u.items {
		keys = append(keys, k)
	}
	return keys
}

func (u *Map[T]) Values() []T {
	values := make([]T, 0, len(u.items))
	for _, v := range u.items {
		values = append(values, v)
	}
	return values
}

func (u *Map[T]) Get(key string) (T, bool) {
	u.l.RLock()
	i, ok := u.items[key]
	u.l.RUnlock()
	return i, ok
}

func (u *Map[T]) Set(key string, value T) {
	u.l.Lock()
	u.items[key] = value
	u.l.Unlock()
}

func (u *Map[T]) Remove(key string) bool {
	var ok bool
	u.l.Lock()
	if _, ok = u.items[key]; ok {
		delete(u.items, key)
	}
	u.l.Unlock()
	return ok
}

func NewMap[T any]() *Map[T] {
	return &Map[T]{
		items: make(map[string]T),
		l:     sync.RWMutex{},
	}
}
