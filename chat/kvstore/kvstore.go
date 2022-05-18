package kvstore

import "sync"

type KVStore[K comparable, T any] struct {
	items map[K]T
	l     sync.RWMutex
}

func (u *KVStore[K, T]) Keys() []K {
	u.l.RLock()
	defer u.l.RUnlock()
	keys := make([]K, 0, len(u.items))
	for k := range u.items {
		keys = append(keys, k)
	}
	return keys
}

func (u *KVStore[K, T]) Values() []T {
	u.l.RLock()
	defer u.l.RUnlock()
	values := make([]T, 0, len(u.items))
	for _, v := range u.items {
		values = append(values, v)
	}
	return values
}

func (u *KVStore[K, T]) Map() map[K]T {
	u.l.RLock()
	defer u.l.RUnlock()
	m := make(map[K]T, len(u.items))
	for k, v := range u.items {
		m[k] = v
	}
	return m
}

func (u *KVStore[K, T]) Get(key K) (T, bool) {
	u.l.RLock()
	defer u.l.RUnlock()
	i, ok := u.items[key]
	return i, ok
}

func (u *KVStore[K, T]) Set(key K, value T) {
	u.l.Lock()
	defer u.l.Unlock()
	u.items[key] = value
}

func (u *KVStore[K, T]) Delete(key K) bool {
	u.l.Lock()
	defer u.l.Unlock()
	var ok bool
	if _, ok = u.items[key]; ok {
		delete(u.items, key)
	}
	return ok
}

func NewKVStore[K comparable, T any]() *KVStore[K, T] {
	return &KVStore[K, T]{
		items: make(map[K]T),
		l:     sync.RWMutex{},
	}
}
