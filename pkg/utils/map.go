package utils

import (
	"sync"
)

type Map[K comparable, V any] struct {
	data *sync.Map

	lock *sync.RWMutex
}

func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		data: new(sync.Map),
		lock: new(sync.RWMutex),
	}
}

func (g *Map[K, V]) Clear() {
	g.lock.RLock()
	defer g.lock.RUnlock()
	g.data.Clear()
}

func (g *Map[K, V]) Store(key K, value V) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	g.data.Store(key, value)
}

func (g *Map[K, V]) Load(key K) (value V, ok bool) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	load, o := g.data.Load(key)
	return load, o
}

func (g *Map[K, V]) LoadOrStore(data string, set V) (V, bool) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.data.LoadOrStore(data, set)
}

func (g *Map[K, V]) LoadAndDelete(key string) (value V, ok bool) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	value, loaded := g.data.LoadAndDelete(key)
	return value, loaded
}

func (g *Map[K, V]) RemoveIf(key string, f func(group V) bool) {
	g.lock.Lock()
	defer g.lock.Unlock()
	load, ok := g.data.Load(key)
	if !ok {
		return
	}
	if f(load.(V)) {
		g.data.Delete(key)
	}
}

func (g *Map[K, V]) Range(f func(key K, value V) bool) {
	g.data.Range(func(key, value interface{}) bool {
		return f(key.(K), value.(V))
	})
}
