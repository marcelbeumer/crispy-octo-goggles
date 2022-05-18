package safe

import "sync"

type Map[T any] struct {
	items map[string]T
	l     sync.RWMutex
}

func (u *Map[T]) Keys() []string {
	u.l.RLock()
	defer u.l.RUnlock()
	keys := make([]string, 0, len(u.items))
	for k := range u.items {
		keys = append(keys, k)
	}
	return keys
}

func (u *Map[T]) Values() []T {
	u.l.RLock()
	defer u.l.RUnlock()
	values := make([]T, 0, len(u.items))
	for _, v := range u.items {
		values = append(values, v)
	}
	return values
}

func (u *Map[T]) Get(key string) (T, bool) {
	u.l.RLock()
	defer u.l.RUnlock()
	i, ok := u.items[key]
	return i, ok
}

func (u *Map[T]) Set(key string, value T) {
	u.l.Lock()
	defer u.l.Unlock()
	u.items[key] = value
}

func (u *Map[T]) Delete(key string) bool {
	u.l.Lock()
	defer u.l.Unlock()
	var ok bool
	if _, ok = u.items[key]; ok {
		delete(u.items, key)
	}
	return ok
}

func NewMap[T any]() *Map[T] {
	return &Map[T]{
		items: make(map[string]T),
		l:     sync.RWMutex{},
	}
}
